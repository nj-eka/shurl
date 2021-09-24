package app_test

import (
	"context"
	"errors"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/app/hashid_tokenizer"
	"github.com/nj-eka/shurl/config"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/store/bolt_store"
	"log"
	"os"
	"testing"
	"time"
)

var ap *app.App
var id2key = map[int]string{
	0: "EZead",
	1: "EdGed",
	2: "XBa80",
}
var expiredAt = time.Now().AddDate(0,0,1)

func appInit(){
	ctx := context.Background()
	tokenizer, err := hashid_tokenizer.NewHashidTokenizer(&config.TokenizerConfig{"ecafbaf0-1bcc-11ec-9621-0242ac130002", 5, "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"})
	if err != nil{
		log.Fatal(err)
	}
	_ = os.Remove("links.db")
	store, err := bolt_store.NewBoltLinkStore(ctx, config.BoltStoreConfig{FilePath: "links.db", Timeout: 10 * time.Second})
	if err != nil{
		log.Fatal(err)
	}
	ap = app.NewApp(store, tokenizer)
}

func TestMain(m *testing.M) {
	appInit()
	code := m.Run()
	_ = ap.Close(nil)
	_ = os.Remove("links.db")
	os.Exit(code)
}

func TestApp_CreateToken(t *testing.T) {
	ctx := context.Background()
	type args struct {
		targetUrl string
		expiredAt *time.Time
	}
		tests := []struct {
		name      string
		args      args
		wantKey   string
		wantAdded bool
		wantErr   error
	}{
		{"add invalid url", args{"https//stackoverflow.com", nil}, "", false, app.ErrInvalidUrl},
		{"add first url", args{"https://stackoverflow.com", nil}, id2key[1], true, nil},
		{"add first url update", args{"https://stackoverflow.com", &expiredAt}, id2key[1], false, nil},
		{"add second url", args{"https://stackoverflow.com/questions", &expiredAt}, id2key[2], true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotAdded, gotErr := ap.CreateToken(ctx, tt.args.targetUrl, tt.args.expiredAt)
			if gotKey != tt.wantKey {
				t.Errorf("CreateToken() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
			if gotAdded != tt.wantAdded {
				t.Errorf("CreateToken() gotAdded = %v, want %v", gotAdded, tt.wantAdded)
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

func TestApp_GetLink(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		args   args
		wantLink   *app.Link
		wantErrIs  error
	}{
		{"get not existed", args{"DApEj4wbneowA"}, nil, app.ErrNotFound},
		{"get first link", args{id2key[1]}, &app.Link{
			Id:        1,
			TargetUrl: "https://stackoverflow.com",
			ExpiredAt: &expiredAt,
			Hits:      0,
		}, nil},
		{"get second link", args{id2key[2]}, &app.Link{
			Id:        2,
			TargetUrl: "https://stackoverflow.com/questions",
			ExpiredAt: &expiredAt,
			Hits:      0,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, gotErr := ap.GetLink(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil{
				if ee, ok := tt.wantErrIs.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs){
					t.Errorf("GetLink() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
			if (gotLink == tt.wantLink){
				return
			}
			if (gotLink != nil && tt.wantLink == nil) || (gotLink == nil && tt.wantLink != nil){
				t.Errorf("GetLink() got = %v, want %v", gotLink, tt.wantLink)
				return
			}
			if gotLink.Id != tt.wantLink.Id ||
				gotLink.TargetUrl != tt.wantLink.TargetUrl ||
				gotLink.ExpiredAt.UnixNano() != tt.wantLink.ExpiredAt.UnixNano() ||
				gotLink.Hits != gotLink.Hits {
				t.Errorf("GetLink() got = %v, want %v", gotLink, tt.wantLink)
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetLink() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
func TestApp_HitLink(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		args   args
		wantLink   *app.Link
		wantErrIs  error
	}{
		{"hit not existed", args{"DApEj4wbneowA"}, nil, app.ErrNotFound},
		{"hit first link - 1", args{id2key[1]}, &app.Link{
			Id:        1,
			Hits:      1,
		}, nil},
		{"hit first link - 2", args{id2key[1]}, &app.Link{
			Id:        1,
			Hits:      2,
		}, nil},
		{"hit second link", args{id2key[2]}, &app.Link{
			Id:        2,
			TargetUrl: "https://stackoverflow.com/questions",
			ExpiredAt: &expiredAt,
			Hits:      1,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, gotErr := ap.HitLink(ctx, tt.args.key)
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
			if (gotLink == tt.wantLink){
				return
			}
			if (gotLink != nil && tt.wantLink == nil) || (gotLink == nil && tt.wantLink != nil){
				t.Errorf("HitLink() got = %v, want %v", gotLink, tt.wantLink)
				return
			}
			if gotLink.Id != tt.wantLink.Id || gotLink.Hits != gotLink.Hits {
				t.Errorf("HitLink() got = %v, want %v", gotLink, tt.wantLink)
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetLink() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestApp_DeleteLink(t *testing.T) {
	ctx := context.Background()
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		args   args
		wantErrIs  error
	}{
		{"delete not existed", args{"DApEj4wbneowA"}, app.ErrNotFound},
		{"delete second link - 1", args{id2key[1]}, nil},
		{"delete second link - 2", args{id2key[2]}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := ap.DeleteLink(ctx, tt.args.key)
			if tt.wantErrIs != nil || gotErr != nil{
				if ee, ok := tt.wantErrIs.(errs.Error); ok{
					if ee == gotErr {
						return
					}
				}
				if !errors.Is(gotErr, tt.wantErrIs){
					t.Errorf("DeleteLink() gotErr = %v, want %v", gotErr, tt.wantErrIs)
					return
				}
			}
		})
	}
}
