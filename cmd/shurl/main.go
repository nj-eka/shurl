package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nj-eka/shurl/api/router"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/app/hashid_tokenizer"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/internal/logging"
	"github.com/nj-eka/shurl/store/bolt_store"
	"github.com/nj-eka/shurl/utils/fsutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	fp "path/filepath"
	"strings"
)

var (
	appName, appDir      = fp.Base(os.Args[0]), fp.Dir(os.Args[0])
	defaultAppConfigPath = fp.Join(appDir, "config.yaml")
	// default config settings
	appCfg = config.AppConfig{
		Logging: &config.LoggingConfig{
			FilePath: fp.Join(appDir, appName+".log"),
			Level:    logging.DefaultLevel.String(),
			Format:   logging.DefaultFormat,
		},
		Server: &config.ServerConfig{
			Host: "",
			Port: 8443,
			// Address: ":8443",
			Timeout: 0,
		},
		Store: &config.StoreConfig{
			Bolt: &config.BoltStoreConfig{
				FilePath: fp.Join(appDir, appName+".db"),
				Timeout:  0,
			},
		},
		Tokenizer: &config.TokenizerConfig{
			Hashid: &config.HashidTokenizerConfig{
				Salt:      "If you don't understand ladders then don't play Go.",
				MinLength: 0,
				Alphabet:  "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			},
		},
	}
	currConfigSaveToPath string
	usr                  *user.User
	a                    *app.App
)

func prepareConfig() {
	var (
		err                               error
		envPrefix, dotEnvPath, configPath string
		currConfigSaveToPath              string
	)
	flag.StringVar(&envPrefix, "env-prefix", appName, "env prefix")
	flag.StringVar(&dotEnvPath, "env", ".env", "path to .env file") // added to simplify deployment procedure
	flag.StringVar(&configPath, "config", defaultAppConfigPath, "path to config file")
	flag.StringVar(&currConfigSaveToPath, "save-config", "", "path to save current resolved config")
	flag.Parse()

	_ = godotenv.Load(dotEnvPath)
	viper.SetConfigFile(configPath)
	// list of environment variables that replace config values
	// (read from config file with default values specified in app: appCfg)
	// in format: APPNAME_LEVELS_WITH_UNDERLINE_IN_UPPERCASE (see .env)
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	_ = viper.BindEnv("logging.level") // there is no case of "missing key to bind to"
	_ = viper.BindEnv("logging.path")
	//_ = viper.BindEnv("server.addr")
	_ = viper.BindEnv("server.host")
	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("store.bolt.path")
	_ = viper.BindEnv("tokenizer.salt")
	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalln("Config file not found: ", configPath)
		} else {
			log.Fatalln("Invalid config: ", err)
		}
	} else {
		if err = viper.Unmarshal(&appCfg); err != nil {
			log.Fatalln("Invalid config: ", err)
		}
	}
	//usr, err = fsutils.GetCurrentUser()
	//if err != nil {
	//	log.Fatalln("Unknown user: ", err)
	//}
	if err := logging.Initialize(context.TODO(), appCfg.Logging, usr); err != nil {
		log.Fatalln("Logging init failed: ", err)
	}
	logging.Msg().Infof("app version %s built from %s on %s\n", app.Version, app.Commit, app.BuildTime)
}

// app init, exit on error
func init() {
	fmt.Printf("short url generator has version %s built from %s on %s\n", app.Version, app.Commit, app.BuildTime)
	prepareConfig()
	// logging is initialized
	ctx := cu.BuildContext(context.Background(), cu.SetContextOperation("00.init"), errs.SetDefaultErrsSeverity(errs.SeverityCritical))
	logging.Msg(ctx).Infof("pid:%d user:%d(%d) group:%d(%d)", os.Getpid(), os.Getuid(), os.Geteuid(), os.Getgid(), os.Getegid())
	if appCfg.Server == nil {
		logging.LogError(ctx, errs.KindServer, "no server config")
		log.Exit(1)
	}
	if appCfg.Router == nil {
		logging.LogError(ctx, errs.KindRouter, "no router config")
		log.Exit(1)
	}
	var err error
	appCfg.Router.WebPath, err = fsutils.ResolvePath(appCfg.Router.WebPath, usr)
	if err != nil {
		logging.LogError(ctx, errs.KindRouter, "invalid router web path")
		log.Exit(1)
	}
	if currConfigSaveToPath != "" {
		if currConfigSaveToPath, err = fsutils.SafeParentResolvePath(currConfigSaveToPath, usr, 0700); err != nil {
			logging.LogError(ctx, errs.KindInvalidValue, fmt.Errorf("invalid path [%s] to save config: %w", currConfigSaveToPath, err))
			log.Exit(1)
		}
	}
	var tokenizer app.Tokenizer
	if appCfg.Tokenizer != nil && appCfg.Tokenizer.Hashid != nil {
		tokenizer, err = hashid_tokenizer.NewHashidTokenizer(appCfg.Tokenizer.Hashid)
		if err != nil {
			logging.LogError(ctx, errs.KindTokenizer, fmt.Errorf("init tokenizer failed: %w", err))
			log.Exit(1)
		}
	} else {
		logging.LogError(ctx, errs.KindTokenizer, "no tokenizer config")
		log.Exit(1)
	}
	var store app.LinkStore
	if appCfg.Store != nil && appCfg.Store.Bolt != nil {
		if appCfg.Store.Bolt.FilePath, err = fsutils.SafeParentResolvePath(appCfg.Store.Bolt.FilePath, usr, 0700); err == nil {
			store, err = bolt_store.NewBoltLinkStore(ctx, *appCfg.Store.Bolt)
		}
		if err != nil {
			logging.LogError(ctx, errs.KindStore, fmt.Errorf("init bolt store failed: %w", err))
			log.Exit(1)
		}
	} else {
		logging.LogError(ctx, errs.KindStore, "no store config")
		log.Exit(1)
	}
	if currConfigSaveToPath != "" {
		if err = viper.WriteConfigAs(currConfigSaveToPath); err != nil {
			logging.LogError(ctx, errs.KindIO, fmt.Errorf("saving config to file [%s] failed: %w", currConfigSaveToPath, err))
			// no exit (or close store)
		}
	}

	a = app.NewApp(store, tokenizer)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx = cu.BuildContext(ctx, cu.SetContextOperation("0.main"))
	defer func() {
		cancel()
		if err := a.Close(ctx); err != nil {
			logging.LogError(ctx, errs.SeverityCritical, errs.KindStore, fmt.Errorf("closing store failed: %w", err))
		}
		logging.Finalize()
	}()
	art, err := router.NewAppRouter(ctx, a, appCfg.Router)
	if err != nil {
		logging.LogError(err)
		return
	}
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", appCfg.Server.Host, appCfg.Server.Port),
		Handler:      art,
		ReadTimeout:  appCfg.Server.Timeout,
		WriteTimeout: appCfg.Server.Timeout,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, stop := context.WithTimeout(context.Background(), appCfg.ShutdownTimeout)
		defer stop()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logging.LogError(ctx, errs.KindServer, fmt.Errorf("shutdown server timeout exceeded: %w", shutdownCtx.Err()))
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logging.LogError(ctx, errs.KindServer, fmt.Errorf("shutdown http server failed: %w", shutdownCtx.Err()))
		}
	}()
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logging.LogError(ctx, errs.KindServer, fmt.Errorf("http server failed: %w", err))
	}
}
