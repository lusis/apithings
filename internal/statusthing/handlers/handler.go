package handlers

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"

	"github.com/lusis/apithings/internal/static"
	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/ui/templates"

	"golang.org/x/exp/slog"

	chi "github.com/go-chi/chi/v5"
)

const (
	// DefaultBasePath is the default base path
	DefaultBasePath = "/statusthings"
)

var siteTemplates = []string{
	"index.htmx",
	"card.htmx",
}

// StatusThingHandler is a struct that provides an http access for statusthings
type StatusThingHandler struct {
	mux      *chi.Mux
	apiMux   *chi.Mux
	provider providers.Provider

	// basePath is the path where the handler is mounted
	basePath string

	apikey string

	templates map[string]*template.Template
}

type httpRepresentation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

const applicationJSON = "application/json"
const textHTML = "text/html"
const contentTypeHeader = "content-type"

// NewStatusThingHandler returns a new statusthing handler
func NewStatusThingHandler(provider providers.Provider, opts ...HandlerOption) (*StatusThingHandler, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}
	mux := chi.NewRouter()

	sth := &StatusThingHandler{
		basePath:  DefaultBasePath,
		provider:  provider,
		templates: make(map[string]*template.Template),
		mux:       mux,
	}

	for _, opt := range opts {
		if err := opt(sth); err != nil {
			return nil, err
		}
	}

	// parse our templates
	tmplFs := templates.UITemplateFS
	for _, fname := range siteTemplates {
		slog.Debug("parsing template file " + fname)

		tmpl, err := template.New(fname).ParseFS(tmplFs, fname)
		if err != nil {
			return nil, err
		}

		if tmpl == nil {
			return nil, fmt.Errorf("got a nil template for " + fname)
		}
		sth.templates[fname] = tmpl
	}

	// static files
	staticfs := static.StaticFS
	sfs, err := fs.Sub(staticfs, "files")
	if err != nil {
		return nil, err
	}
	staticPath := path.Join(sth.basePath, "/static/")
	mux.Route(staticPath, func(r chi.Router) {
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			fserv := http.StripPrefix(staticPath, http.FileServer(http.FS(sfs)))
			fserv.ServeHTTP(w, r)
		})
	})

	// ui
	mux.Route(path.Join(sth.basePath, "/"), func(r chi.Router) {
		sth.addUIRoutes(r)
	})

	// api
	mux.Route(path.Join(sth.basePath, "/api/"), func(r chi.Router) {
		sth.addAPIRoutes(r)
	})

	return sth, nil
}

func (h *StatusThingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
