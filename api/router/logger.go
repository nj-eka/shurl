// guided by github.com/go-chi/chi/_examples/logging
package router

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/internal/logging"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

type StructuredLogger struct {
	Logger *logrus.Logger
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("request"), errs.SetDefaultErrsKind(errs.KindRouter))
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	entry.Logger = entry.Logger.(*logrus.Entry).WithContext(ctx)

	logFields := logrus.Fields{}
	logFields["rts"] = time.Now().UTC().Format(logging.DefaultTimeFormat)
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method
	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()
	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	logging.Msg(entry.Logger).Infoln("request started")

	return entry
}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})
	logging.Msg(l.Logger).Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
	logging.LoggerError(l.Logger.(*logrus.Logger), errs.SeverityCritical, v) // todo: choose one of from frames or stack ...
}
