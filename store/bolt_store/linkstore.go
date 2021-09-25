package bolt_store

import (
	"context"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/utils/strutils"
	bolt "go.etcd.io/bbolt"
	"time"
)

var _ app.LinkStore = &boltLinkStore{}

type boltLinkStore struct {
	db *storm.DB
}

func NewBoltLinkStore(ctx context.Context, cfg config.BoltStoreConfig) (app.LinkStore, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Init"))
	db, err := storm.Open(cfg.FilePath, storm.BoltOptions(0660, &bolt.Options{Timeout: cfg.Timeout}))
	if err != nil {
		return nil, errs.E(ctx, errs.KindStore, fmt.Errorf("opening bolt db [%s] failed: %w", cfg.FilePath, err))
	}
	return &boltLinkStore{db: db}, nil
}

func (b *boltLinkStore) Create(ctx context.Context, targetUrl string, expiredAt *time.Time) (int, bool, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Create"), errs.SetDefaultErrsKind(errs.KindStore))
	var ie error
	id, added := -1, false
	tx, ie := b.db.Begin(true)
	if ie == nil {
		defer func() {
			_ = tx.Rollback()
		}()
		link := Link{}
		if ie = tx.One("TargetUrl", targetUrl, &link); ie == nil {
			ie = tx.UpdateField(&Link{Id: link.Id}, "ExpiredAt", expiredAt)
		} else if ie == storm.ErrNotFound {
			link.TargetUrl = targetUrl
			link.CreatedAt = time.Now().UTC()
			link.ExpiredAt = expiredAt
			ie = tx.Save(&link)
			added = true
		}
		if ie == nil {
			if ie = tx.Commit(); ie == nil {
				return link.Id, added, nil
			}
		}
	}
	return id, added, errs.E(ctx, fmt.Errorf("adding link [%s] failed: %w", strutils.Truncate(targetUrl, 24, "..."), ie))
}

func (b *boltLinkStore) Get(ctx context.Context, id int) (*app.Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Get"), errs.SetDefaultErrsKind(errs.KindStore))
	link := Link{}
	if err := b.db.One("Id", id, &link); err != nil {
		if err == storm.ErrNotFound {
			return nil, errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return nil, errs.E(ctx, fmt.Errorf("getting link with id [%d] failed: %w", id, err))
	}
	return &app.Link{
		Id:        link.Id,
		TargetUrl: link.TargetUrl,
		CreatedAt: link.CreatedAt,
		ExpiredAt: link.ExpiredAt,
		DeletedAt: link.DeletedAt,
		Hits:      link.Hits,
	}, nil
}

func (b *boltLinkStore) Hit(ctx context.Context, id int) (*app.Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Hit"), errs.SetDefaultErrsKind(errs.KindStore))
	var ie error
	tx, ie := b.db.Begin(true)
	if ie == nil {
		defer func() {
			_ = tx.Rollback()
		}()
		link := Link{}
		if ie = tx.One("Id", id, &link); ie == nil {
			link.Hits++
			if ie = tx.UpdateField(&Link{Id: id}, "Hits", link.Hits); ie == nil {
				if ie = tx.Commit(); ie == nil {
					return &app.Link{
						Id:        link.Id,
						TargetUrl: link.TargetUrl,
						CreatedAt: link.CreatedAt,
						ExpiredAt: link.ExpiredAt,
						DeletedAt: link.DeletedAt,
						Hits:      link.Hits,
					}, nil
				}
			}
		}
		if ie == storm.ErrNotFound {
			return nil, errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
	}
	return nil, errs.E(ctx, fmt.Errorf("hitting link with id [%d] failed: %w", id, ie))
}

func (b *boltLinkStore) SetDeleted(ctx context.Context, id int) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.SetDel"), errs.SetDefaultErrsKind(errs.KindStore))
	deletedAt := time.Now().UTC()
	if err := b.db.UpdateField(&Link{Id: id}, "DeletedAt", &deletedAt); err != nil {
		if err == storm.ErrNotFound {
			return errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return errs.E(ctx, fmt.Errorf("setting deleted link with id [%d] failed: %w", id, err))
	}
	return nil
}

func (b *boltLinkStore) Delete(ctx context.Context, id int) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Delete"), errs.SetDefaultErrsKind(errs.KindStore))
	if err := b.db.DeleteStruct(&Link{Id: id}); err != nil {
		if err == storm.ErrNotFound {
			return errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return errs.E(ctx, fmt.Errorf("setting deleted link with id [%d] failed: %w", id, err))
	}
	return nil
}

func (b *boltLinkStore) Close(ctx context.Context) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("bolt.Fin"), errs.SetDefaultErrsKind(errs.KindStore))
	if err := b.db.Close(); err != nil {
		return errs.E(ctx, errs.SeverityCritical, fmt.Errorf("closing bolt db failed: %w", err))
	}
	return nil
}
