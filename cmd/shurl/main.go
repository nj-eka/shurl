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
	appName = fp.Base(os.Args[0])
	appDir  = fp.Join(fp.Dir(appName), "config.yaml")

	appCfg = config.AppConfig{
		Logging: &config.LoggingConfig{
			FilePath: fp.Join(appDir, appName+".log"),
			Level:    logging.DefaultLevel.String(),
			Format:   logging.DefaultFormat,
		},
		Server: &config.ServerConfig{
			Address: ":8443",
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

	usr *user.User
	a   *app.App
)

// app init, exit on error
func init() {
	// todo: print app version
	var (
		err                               error
		envPrefix, dotEnvPath, configPath string
	)
	flag.StringVar(&envPrefix, "env-prefix", appName, "env prefix")
	flag.StringVar(&dotEnvPath, "env", ".env", "path to .env file")
	flag.StringVar(&configPath, "config", appDir, "path to config file")
	flag.Parse()
	_ = godotenv.Load(dotEnvPath)
	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	_ = viper.BindEnv("logging.level") // no case of "missing key to bind to"
	_ = viper.BindEnv("logging.format")
	_ = viper.BindEnv("logging.path")
	_ = viper.BindEnv("server.addr")
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
	usr, err = fsutils.GetCurrentUser()
	if err != nil {
		log.Fatalln("Unknown user: ", err)
	}
	if err := logging.Initialize(nil, appCfg.Logging, usr); err != nil {
		log.Fatalln("Logging init failed: ", err)
	}
	// logging is initialized -> start logging
	ctx := cu.BuildContext(context.Background(), cu.SetContextOperation("00.init"), errs.SetDefaultErrsSeverity(errs.SeverityCritical))
	if appCfg.Server == nil {
		logging.LogError(ctx, errs.KindServer, "no server config")
		log.Exit(1)
	}
	if appCfg.Router == nil{
		logging.LogError(ctx, errs.KindRouter, "no router config")
		log.Exit(1)
	}
	rd, err := fsutils.ResolvePath(appCfg.Router.WebPath, usr)
	if err != nil{
		logging.LogError(ctx, errs.KindRouter, "invalid router web path")
		log.Exit(1)
	}
	appCfg.Router.WebPath = rd
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
		if rd, err = fsutils.SafeParentResolvePath(appCfg.Store.Bolt.FilePath, usr, 0700); err == nil{
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
	a = app.NewApp(store, tokenizer)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt) //, syscall.SIGINT, syscall.SIGQUIT)
	ctx = cu.BuildContext(ctx, cu.SetContextOperation("0.main"))
	defer func() {
		cancel()
		if err := a.Close(ctx); err != nil {
			logging.LogError(ctx, errs.SeverityCritical, errs.KindStore, fmt.Errorf("closing store failed: %w", err))
		}
		logging.Finalize()
	}()
	art, err := router.NewAppRouter(ctx, a, appCfg.Router)
	if err != nil{
		logging.LogError(err)
		return
	}
	server := &http.Server{
		Addr:         appCfg.Server.Address,
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
