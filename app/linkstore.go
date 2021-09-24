package app

import (
	"context"
	"errors"
	"github.com/nj-eka/shurl/internal/errs"
	"time"
)

var ErrNotFound = errors.New("not found")

type LinkStore interface{
	Create(ctx context.Context, targetUrl string, expiredAt *time.Time) (int, bool, errs.Error)
	Get(ctx context.Context, id int) (*Link, errs.Error)
	Hit(ctx context.Context, id int) (*Link, errs.Error)
	SetDeleted(ctx context.Context, id int) errs.Error
	Delete(ctx context.Context, id int) errs.Error
	Close(ctx context.Context) errs.Error
}
