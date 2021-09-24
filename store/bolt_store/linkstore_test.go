package bolt_store

import (
	"context"
	"errors"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	"github.com/nj-eka/shurl/internal/errs"
	"log"
	"os"
	"testing"
	"time"
)

var store app.LinkStore
var expiredAt = time.Now()

func Init(){
	ctx := context.Background()
	_ = os.Remove("links.db")
	var err error
	store, err =  NewBoltLinkStore(ctx, config.BoltStoreConfig{FilePath: "links.db", Timeout: 10 * time.Second})
	if err != nil{
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	Init()
	code := m.Run()
	_ = store.Close(nil)
	_ = os.Remove("links.db")
	os.Exit(code)
}

func Test_boltLinkStore_Create(t *testing.T) {
	ctx := context.Background()
	type args struct {
		targetUrl string
		expiredAt *time.Time
	}
	tests := []struct {
		name      string
		args      args
		wantKey   int
		wantAdded bool
		wantErr   error
	}{
		{"add invalid url", args{"https//stackoverflow.com", nil}, 1, true, nil},
		{"add first url", args{"https://stackoverflow.com", nil}, 2, true, nil},
		{"add first url update", args{"https://stackoverflow.com", &expiredAt}, 2, false, nil},
		{"add second url", args{"https://stackoverflow.com/questions", &expiredAt}, 3, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotAdded, gotErr := store.Create(ctx, tt.args.targetUrl, tt.args.expiredAt)
			if gotKey != tt.wantKey {
				t.Errorf("Create() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if gotAdded != tt.wantAdded {
				t.Errorf("Create() gotAdded = %v, want %v", gotAdded, tt.wantAdded)
			}
			if tt.wantErr != nil || gotErr != nil{
				if ee, ok := tt.wantErr.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErr){
					t.Errorf("CreateToken() gotErr = %v, want %v", gotErr, tt.wantErr)
				}
			}
		})
	}
}

func Test_boltLinkStore_Get(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key int
	}
	tests := []struct {
		name   string
		args   args
		wantLink   *app.Link
		wantErrIs  error
	}{
		{"get not existed", args{68734}, nil, app.ErrNotFound},
		{"get first link", args{2}, &app.Link{
			Id:        2,
			TargetUrl: "https://stackoverflow.com",
			ExpiredAt: &expiredAt,
			Hits:      0,
		}, nil},
		{"get second link", args{3}, &app.Link{
			Id:        3,
			TargetUrl: "https://stackoverflow.com/questions",
			ExpiredAt: &expiredAt,
			Hits:      0,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, gotErr := store.Get(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil{
				if ee, ok := tt.wantErrIs.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs){
					t.Errorf("Get() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
			if gotLink == tt.wantLink {
				return
			}
			if (gotLink != nil && tt.wantLink == nil) || (gotLink == nil && tt.wantLink != nil){
				t.Errorf("Get() got = %v, want %v", gotLink, tt.wantLink)
				return
			}
			if gotLink.Id != tt.wantLink.Id ||
				gotLink.TargetUrl != tt.wantLink.TargetUrl ||
				gotLink.ExpiredAt.UnixNano() != tt.wantLink.ExpiredAt.UnixNano() ||
				gotLink.Hits != gotLink.Hits {
				t.Errorf("Get() got = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}

func Test_boltLinkStore_Hit(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key int
	}
	tests := []struct {
		name   string
		args   args
		wantLink   *app.Link
		wantErrIs  error
	}{
		{"hit not existed", args{1454987}, nil, app.ErrNotFound},
		{"hit first link - 1", args{2}, &app.Link{
			Id:        2,
			Hits:      1,
		}, nil},
		{"hit first link - 2", args{2}, &app.Link{
			Id:        2,
			Hits:      2,
		}, nil},
		{"hit second link", args{3}, &app.Link{
			Id:        3,
			TargetUrl: "https://stackoverflow.com/questions",
			ExpiredAt: &expiredAt,
			Hits:      1,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, gotErr := store.Hit(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil{
				if ee, ok := tt.wantErrIs.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs){
					t.Errorf("HitLink() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
			if gotLink == tt.wantLink {
				return
			}
			if (gotLink != nil && tt.wantLink == nil) || (gotLink == nil && tt.wantLink != nil){
				t.Errorf("HitLink() got = %v, want %v", gotLink, tt.wantLink)
				return
			}
			if gotLink.Id != tt.wantLink.Id || gotLink.Hits != gotLink.Hits {
				t.Errorf("HitLink() got = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}

func Test_boltLinkStore_SetDeleted(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key int
	}
	tests := []struct {
		name   string
		args   args
		wantErrIs  error
	}{
		{"delete not existed", args{5465}, app.ErrNotFound},
		{"delete second link - 1", args{3}, nil},
		{"delete second link - 2", args{3}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := store.SetDeleted(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil{
				if ee, ok := tt.wantErrIs.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs){
					t.Errorf("SetDeleted() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
		})
	}
}

func Test_boltLinkStore_Delete(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key int
	}
	tests := []struct {
		name      string
		args      args
		wantErrIs error
	}{
		{"delete not existed", args{5465}, app.ErrNotFound},
		{"delete second link - 1", args{3}, nil},
		{"delete second link - 2", args{3}, app.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := store.Delete(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil {
				if ee, ok := tt.wantErrIs.(errs.Error); ok {
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs) {
					t.Errorf("Delete() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
		})
	}
}
