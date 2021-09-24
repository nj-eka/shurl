package logging

import (
	"context"
	"fmt"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/sirupsen/logrus"
	"time"
)

func Msg(args ...interface{}) *logrus.Entry {
	msgFields := logrus.Fields{
		"mts":  time.Now().Format(DefaultTimeFormat),
		"rec":  "msg",
		"type": "string",
	}
	if len(args) == 1 {  // only these use cases there were in stack for now
		switch arg := args[0].(type) {
		case cu.Operation, cu.Operations:
			return logrus.WithFields(msgFields).WithField("ops", fmt.Sprintf("%s", arg))
		case context.Context:
			return logrus.WithFields(msgFields).WithContext(arg)
		case *logrus.Entry:
			arg = arg.WithFields(msgFields)
			return arg
		}
	}
	return logrus.WithFields(msgFields)
}
