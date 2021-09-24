package contexts

import (
	"context"
	"github.com/go-chi/chi/middleware"
)

func GetRequestID(ctx context.Context) string{
	return middleware.GetReqID(ctx)
}
