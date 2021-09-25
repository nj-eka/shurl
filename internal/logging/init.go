package logging

import (
	"context"
	"fmt"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/utils/fsutils"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/user"
	"strings"
)

var logFile *os.File

func Initialize(ctx context.Context, logCfg *config.LoggingConfig, usr *user.User) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("log_init"), errs.SetDefaultErrsSeverity(errs.SeverityCritical))
	logrus.SetOutput(os.Stdout)
	logrus.RegisterExitHandler(Finalize)
	if logCfg.FilePath != "" {
		var file *os.File
		var err error
		if logCfg.FilePath, err = fsutils.SafeParentResolvePath(logCfg.FilePath, usr, 0700); err != nil {
			return errs.E(ctx, errs.KindInvalidValue, fmt.Errorf("invalid log file name <%s>: %w", logCfg.FilePath, err))
		}
		file, err = os.OpenFile(logCfg.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			return errs.E(ctx, errs.KindIO, fmt.Errorf("open file <%s> for logging failed: %w", logCfg.FilePath, err))
		} else {
			logrus.SetOutput(file)
			logFile = file
			fmt.Println("logging to ", file.Name())
		}
	} else {
		fmt.Println("logging to standard output")
	}
	fieldMap := logrus.FieldMap{
		logrus.FieldKeyTime:  "ts",
		logrus.FieldKeyLevel: "lvl",
		logrus.FieldKeyMsg:   "msg"}
	switch strings.ToUpper(logCfg.Format) {
	case "JSON":
		logrus.SetFormatter(
			&ContextFormatter{
				&logrus.JSONFormatter{
					FieldMap: fieldMap,
				},
			})
	case "TEXT":
		logrus.SetFormatter(
			&ContextFormatter{
				&logrus.TextFormatter{
					ForceQuote:       false,
					DisableTimestamp: false,
					FullTimestamp:    true,
					TimestampFormat:  DefaultTimeFormat,
					QuoteEmptyFields: true,
					FieldMap:         fieldMap,
				},
			})
	default:
		return errs.E(ctx, errs.KindInvalidValue, fmt.Errorf("invalid log format [%s]. supported formats: json, text", logCfg.Format))
	}
	lvl, err := logrus.ParseLevel(logCfg.Level)
	if err != nil {
		return errs.E(ctx, errs.KindInvalidValue, fmt.Errorf("parsing log level from config failed: %w", err))
	}
	logrus.SetLevel(lvl)
	Msg(ctx).Debugf("Logging started with level <%s>", logCfg.Level) // first record in log file
	logrus.SetReportCaller(lvl > logrus.InfoLevel)
	errs.WithFrames(lvl > logrus.InfoLevel)
	log.SetOutput(logrus.StandardLogger().Writer()) // to use with standard log pkg
	return nil
}

func Finalize() { // force sync on exit (optional)
	op := cu.Operation("log_finalize")
	if nil != logFile {
		Msg(op).Debug("Logging stopped") // last record in log file
		if err := logFile.Sync(); err != nil {
			Msg(op).Errorf("sync log buffer with file [%s] - failed: %v", logFile.Name(), err)
		}
		if err := logFile.Close(); err != nil {
			Msg(op).Errorf("closing log file [%s] - failed: %v\n", logFile.Name(), err)
		}
	}
}
