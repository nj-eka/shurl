package errs

import (
	"context"
	cu "github.com/nj-eka/shurl/internal/contexts"
)

func SetDefaultErrsSeverity(severity Severity) cu.PartialContextFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, cu.DefaultErrsSeverityKey, severity)
	}
}

func GetDefaultErrsSeverity(ctx context.Context) Severity {
	if severity, ok := ctx.Value(cu.DefaultErrsSeverityKey).(Severity); ok {
		return severity
	}
	return SeverityError
}
