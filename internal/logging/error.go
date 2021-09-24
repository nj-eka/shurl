package logging

import (
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/sirupsen/logrus"
)

func LogError(args ...interface{}) {
	err := errs.E(args...)
	lvl := Severity2LogLevel[err.Severity()]
	logrus.WithFields(LoggerErrorFields(err)).Log(lvl, err)
}

func LoggerError(l *logrus.Logger, args ...interface{}) {
	err := errs.E(args...)
	lvl := Severity2LogLevel[err.Severity()]
	l.WithFields(LoggerErrorFields(err)).Log(lvl, err)
}

func LoggerErrorFields(err errs.Error) logrus.Fields {
	fields := make(map[string]interface{}, 8)
	fields["rec"] = "error"
	fields["type"] = "errs.Error" //fmt.Sprintf("%T", err) = *errs.errorData
	fields["severity"] = err.Severity().String()
	if err.Kind() != errs.KindOther {
		fields["kind"] = err.Kind().String()
	}
	if !err.OperationPath().Empty() {
		fields["ops"] = err.OperationPath().String()
	}
	if err.RequestID() != "" {
		fields["rid"] = err.RequestID()
	}
	if len(err.StackTrace()) > 0 {
		fields["frames"] = err.StackTrace()
	}
	fields["sts"] = err.TimeStamp().UTC().Format(DefaultTimeFormat)
	return fields
}

var Severity2LogLevel = map[errs.Severity]logrus.Level{
	errs.SeverityWarning:  logrus.WarnLevel,
	errs.SeverityError:    logrus.ErrorLevel,
	errs.SeverityCritical: logrus.FatalLevel,
}

//func GetSeveritiesFilter4CurrentLogLevel() (result []errs.Severity) {
//	for severity, logLevel := range Severity2LogLevel {
//		if logrus.IsLevelEnabled(logLevel) {
//			result = append(result, severity)
//		}
//	}
//	return
//}
