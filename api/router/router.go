//go:generate oapi-codegen -package=app_openapi -generate types -o ../app_openapi/types.gen.go ../app_openapi/openapi.yaml
//go:generate oapi-codegen -package=app_openapi -generate chi-server -o ../app_openapi/chi_server.gen.go ../app_openapi/openapi.yaml
//go:generate oapi-codegen -package=app_openapi -generate spec -o ../app_openapi/spec.gen.go ../app_openapi/openapi.yaml
// guided by oapi-codegen/examples/petstore-expanded/chi/petstore.go
package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	chi_middleware "github.com/go-chi/chi/middleware"
	api "github.com/nj-eka/shurl/api/app_openapi"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	cu "github.com/nj-eka/shurl/internal/contexts"
	"github.com/nj-eka/shurl/internal/errs"
	"github.com/nj-eka/shurl/internal/logging"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var _ api.ServerInterface = (*AppRouter)(nil)

type AppRouter struct {
	http.Handler
	a *app.App
	cfg *config.RouterConfig
}

func NewAppRouter(ctx context.Context, a *app.App, cfg *config.RouterConfig) (*AppRouter, error) {
	ctx = cu.BuildContext(ctx, cu.AddContextOperation("router.init"), errs.SetDefaultErrsKind(errs.KindRouter), errs.SetDefaultErrsSeverity(errs.SeverityCritical))
	art := &AppRouter{a: a, cfg: cfg}
	r := chi.NewRouter()

	// middlewares
	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(NewStructuredLogger(logrus.StandardLogger()))
	r.Use(chi_middleware.Recoverer)
	//r.Use(chi_middleware.URLFormat)
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, errs.E(ctx, fmt.Errorf("loading swagger spec failed: %w", err))
	}
	//swagger.Servers = nil
	//r.Use(oapi_middleware.OapiRequestValidator(swagger)) // to check all requests against the OpenAPI schema + cut off frontend)

	// add static web server support
	logging.Msg(ctx).Debug("file server starts on dir: ", filepath.Join(cfg.WebPath, "static"))
	fileServer := http.FileServer(http.Dir(filepath.Join(cfg.WebPath, "static")))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		logging.Msg(cu.Operation("file_server")).Debug(r.RequestURI)
		//fileServer.ServeHTTP(w, r)
		http.StripPrefix("/static", fileServer).ServeHTTP(w, r)
	})
	// add frontend index page
	r.Get("/", art.GetMainPage)
	// add openapi (swagger) ui
	r.Get("/openapi", art.GetOpenAPI)
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(swagger)
	})

	////register AppRouter as handler for api.ServerInterface
	////api.HandlerFromMux(art, r)
	r.Mount("/", api.Handler(art))
	art.Handler = r
	return art, nil
}

func (art *AppRouter) GetMainPage(w http.ResponseWriter, r *http.Request){
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("get_main"), errs.SetDefaultErrsKind(errs.KindRouter))
	ts, err := template.ParseFiles(filepath.Join(art.cfg.WebPath, "templates/index.html"))
	if err != nil {
		logging.LogError(ctx, errs.SeverityCritical, fmt.Errorf("parsing main page template failed: %w", err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
	if err != ts.Execute(w, nil) {
		logging.LogError(ctx, errs.SeverityCritical, fmt.Errorf("executing main page template failed: %w", err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func (art *AppRouter) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("get_openapi"), errs.SetDefaultErrsKind(errs.KindRouter))
	ts, err := template.ParseFiles(filepath.Join(art.cfg.WebPath, "templates/openapi_index.html"))
	if err != nil {
		logging.LogError(ctx, errs.SeverityCritical, fmt.Errorf("parsing openapi template failed: %w", err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
	if err != ts.Execute(w, nil) {
		logging.LogError(ctx, errs.SeverityCritical, fmt.Errorf("executing openapi template failed: %w", err))
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func (art *AppRouter) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("create_shurl"), errs.SetDefaultErrsKind(errs.KindRouter))
	defer func() {
		_ = r.Body.Close()
	}()
	var requestShurl api.RequestShortUrl
	if err := json.NewDecoder(r.Body).Decode(&requestShurl); err != nil {
		logging.LogError(ctx, fmt.Errorf("Invalid request format: %w", err))
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	var expiredAt *time.Time
	if requestShurl.ExpiredInDays != nil {
		t := time.Now().UTC().AddDate(0,0, int(*requestShurl.ExpiredInDays))
		expiredAt = &t
	}
	token, added, err := art.a.CreateToken(ctx, requestShurl.TargetUrl, expiredAt)
	if err!= nil{
		logging.LogError(ctx, err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if added {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	shurlInfo := fmt.Sprintf("/%s/info", token)
	result := &api.ResponseShortUrl{
		ShortUrl:     fmt.Sprintf("/%s", token), // todo: use entry point from cfg
		ShortUrlInfo: &shurlInfo,
	}
	if err := json.NewEncoder(w).Encode(result); err != nil{
		logging.LogError(ctx, fmt.Errorf("encoding [%v] to json failed: %w", result, err))
	}
}

func (art *AppRouter) HitShortUrl(w http.ResponseWriter, r *http.Request, token string) {
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("hit_shurl"), errs.SetDefaultErrsKind(errs.KindRouter))
	link, err := art.a.HitLink(ctx, token)
	if err != nil{
		if errors.Is(err, app.ErrNotFound){
			http.Error(w, "", http.StatusNotFound)
			return
		}
		logging.LogError(ctx, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, link.TargetUrl, http.StatusSeeOther)
}

func (art *AppRouter) GetShortUrlInfo(w http.ResponseWriter, r *http.Request, token string) {
	ctx := cu.BuildContext(r.Context(), cu.AddContextOperation("hit_shurl"), errs.SetDefaultErrsKind(errs.KindRouter))
	link, err := art.a.GetLink(ctx, token)
	if err != nil{
		if errors.Is(err, app.ErrNotFound){
			http.Error(w, "", http.StatusNotFound)
			return
		}
		logging.LogError(ctx, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	result := api.Link{
		CreatedAt: link.CreatedAt,
		DeletedAt: link.DeletedAt,
		ExpiredAt: link.ExpiredAt,
		Hits:      int32(link.Hits),
		TargetUrl: link.TargetUrl,
		Token:     link.Key,
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil{
		logging.LogError(ctx, fmt.Errorf("encoding [%v] to json failed: %w", result, err))
	}
}
