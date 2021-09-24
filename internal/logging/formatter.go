package logging

import (
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/sirupsen/logrus"
)

type ContextFormatter struct {
	BaseFormatter logrus.Formatter
}

func (f *ContextFormatter) Format(e *logrus.Entry) ([]byte, error) {
	if ctx := e.Context; nil != ctx {
		if ops := cu.GetContextOperations(ctx).String(); ops != "" {
			e.Data["ops"] = ops
		}
		if rid := cu.GetRequestID(ctx); rid != "" {
			e.Data["rid"] = rid
		}
	}
	return f.BaseFormatter.Format(e)
}
