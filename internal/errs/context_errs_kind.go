package errs

import (
	"context"
	cu "github.com/nj-eka/shurl/internal/contexts"
)

func SetDefaultErrsKind(kind Kind) cu.PartialContextFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, cu.DefaultErrsKindKey, kind)
	}
}

func GetDefaultErrsKind(ctx context.Context) Kind {
	if kind, ok := ctx.Value(cu.DefaultErrsKindKey).(Kind); ok{
		return kind
	}
	return KindOther
}
