package mem_store

import (
	"context"
	"fmt"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/utils/strutils"
	"time"
)

func NewMemStore(ctx context.Context, cfg config.MemStoreConfig) (app.LinkStore, errs.Error){
	done := make(chan struct{})
	mlm, err := newMapManager(done, cfg.FilePath);
	if err != nil{
		return nil, errs.E(ctx, errs.KindStore, fmt.Errorf("init mem store with path [%s] failed: %w", cfg.FilePath, err))
	}
	return &memLinkStore{ done: done, mlm: mlm}, nil
}

type memLinkStore struct{
	done chan struct{}
	mlm  *mapLinkManager
}

func (mls *memLinkStore) Close(ctx context.Context) errs.Error {
	close(mls.done)
	<-mls.mlm.Done()
	if err := mls.mlm.Err(); err != nil{
		return errs.E(ctx, errs.SeverityCritical, fmt.Errorf("closing mem store failed: %w", err))
	}
	return nil
}

func (mls *memLinkStore) Create(ctx context.Context, targetUrl string, expiredAt *time.Time) (int, bool, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("mem.Create"), errs.SetDefaultErrsKind(errs.KindStore))
	id, added, err := mls.mlm.addLink(targetUrl, expiredAt)
	if err != nil{
		if err == ErrNotFound{
			return id, added, errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return id, added, errs.E(ctx, fmt.Errorf("adding link [%s] failed: %w", strutils.Truncate(targetUrl, 24, "..."), err))
	}
	return id, added, nil
}

func (mls *memLinkStore) Get(ctx context.Context, id int) (*app.Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("mem.Get"), errs.SetDefaultErrsKind(errs.KindStore))
	if link, err := mls.mlm.getLink(id); err != nil{
		if err == ErrNotFound {
			return nil, errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return nil, errs.E(ctx, fmt.Errorf("getting link with id [%d] failed: %w", id, err))
	} else{
		return 	&app.Link{
			Id:        link.Id,
			TargetUrl: link.TargetUrl,
			CreatedAt: link.CreatedAt,
			ExpiredAt: link.ExpiredAt,
			DeletedAt: link.DeletedAt,
			Hits:      link.Hits,
		}, nil
	}
}

func (mls *memLinkStore) Hit(ctx context.Context, id int) (*app.Link, errs.Error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("mem.Hit"), errs.SetDefaultErrsKind(errs.KindStore))
	if link, err := mls.mlm.hitLink(id); err != nil{
		if err == ErrNotFound {
			return nil, errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return nil, errs.E(ctx, fmt.Errorf("hitting link with id [%d] failed: %w", id, err))
	} else{
		return 	&app.Link{
			Id:        link.Id,
			TargetUrl: link.TargetUrl,
			CreatedAt: link.CreatedAt,
			ExpiredAt: link.ExpiredAt,
			DeletedAt: link.DeletedAt,
			Hits:      link.Hits,
		}, nil
	}
}

func (mls *memLinkStore) SetDeleted(ctx context.Context, id int) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("mem.SetDeleted"), errs.SetDefaultErrsKind(errs.KindStore))
	if err := mls.mlm.setLinkDeleted(id); err != nil{
		if err == ErrNotFound {
			return errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return errs.E(ctx, fmt.Errorf("setting deleted link with id [%d] failed: %w", id, err))
	}
	return nil
}

func (mls *memLinkStore) Delete(ctx context.Context, id int) errs.Error {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("mem.Delete"), errs.SetDefaultErrsKind(errs.KindStore))
	if err := mls.mlm.removeLink(id); err != nil{
		if err == ErrNotFound {
			return errs.E(ctx, errs.SeverityWarning, app.ErrNotFound)
		}
		return errs.E(ctx, fmt.Errorf("setting deleted link with id [%d] failed: %w", id, err))
	}
	return nil
}
