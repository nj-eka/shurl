package app

import (
	"context"
	"fmt"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"net/url"
	"time"
)

type App struct {
	store     LinkStore
	tokenizer Tokenizer
}

func NewApp(store LinkStore, tokenizer Tokenizer) *App {
	return &App{
		store:     store,
		tokenizer: tokenizer,
	}
}

func (a App) CreateToken(ctx context.Context, targetUrl string, expiredAt *time.Time) (key string, added bool, err errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("app.Create"))
	if _, ie := url.ParseRequestURI(targetUrl); ie != nil {
		return "", false, errs.E(
			ctx,
			errs.KindInvalidValue,
			ErrInvalidUrl,
		)
	}
	if id, added, err := a.store.Create(ctx, targetUrl, expiredAt); err != nil {
		return "", added, err
	} else if key, err := a.tokenizer.Encode(id); err != nil {
		if err2 := a.store.Delete(ctx, id); err2 != nil {
			return "", true, errs.E(
				ctx,
				errs.SeverityCritical,
				errs.KindInternal,
				fmt.Errorf("encoding id [%d] err [%w] occurred while adding / deleted failed: %v", id, err, err2),
			)
		}
		return "", false, errs.E(
			ctx,
			errs.SeverityCritical,
			errs.KindTokenizer,
			fmt.Errorf("encoding id [%d] err [%w] occurred while adding / deleted - ok", id, err),
		)
	} else {
		return key, added, nil
	}
}

func (a App) GetLink(ctx context.Context, key string) (*Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("app.Get"))
	id, err := a.tokenizer.Decode(key)
	if err != nil {
		return nil, errs.E(ctx, errs.SeverityCritical, errs.KindTokenizer, fmt.Errorf("decoding key [%s] failed: %w", key, err))
	}
	return a.store.Get(ctx, id) // return keyless obj, it is known
}

func (a App) HitLink(ctx context.Context, key string) (*Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("app.Hit"))
	id, err := a.tokenizer.Decode(key)
	if err != nil {
		return nil, errs.E(ctx, errs.SeverityCritical, errs.KindTokenizer, fmt.Errorf("decoding key [%s] failed: %w", key, err))
	}
	now := time.Now().UTC()
	if link, err := a.store.Get(ctx, id); err != nil {
		return nil, err
	} else {
		if link.DeletedAt != nil && now.After(link.DeletedAt.UTC()) {
			return nil, errs.E(ctx, errs.SeverityWarning, fmt.Errorf("hit deleted link with id[%d]: %w", id, ErrNotFound))
		}
		if link.ExpiredAt != nil && now.After(link.ExpiredAt.UTC()) {
			return nil, errs.E(ctx, errs.SeverityWarning, fmt.Errorf("hit expired link with id[%d]: %w", id, ErrNotFound))
		}
		return a.store.Hit(ctx, id) // return keyless obj, it is known
	}
}

func (a App) DeleteLink(ctx context.Context, key string) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("app.Delete"))
	id, err := a.tokenizer.Decode(key)
	if err != nil {
		return errs.E(ctx, errs.SeverityCritical, errs.KindTokenizer, fmt.Errorf("decoding key [%s] failed: %w", key, err))
	}
	return a.store.SetDeleted(ctx, id)
}

func (a App) Close(ctx context.Context) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("app.Close"))
	if a.store != nil {
		return a.store.Close(ctx)
	}
	return nil
}
