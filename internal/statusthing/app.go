package statusthing

import (
	"net/http"
	"sync"

	"github.com/lusis/apithings/internal/statusthing/handlers"
	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/storers"

	"golang.ngrok.com/ngrok"

	"golang.org/x/exp/slog"
)

const (
	// DefaultHTTPBasePath is the default path for the http handler to serve requests
	DefaultHTTPBasePath = handlers.DefaultBasePath
)

// App is a application for statusthings
type App struct {
	config             *AppConfig
	statusThingHandler *handlers.StatusThingHandler
}

// New returns a new [App] with the provided options
// at a minimum either WithProvider or WithStorer must be provided
func New(opts ...AppOption) (*App, error) {
	cfg, err := parseOpts(opts...)
	if err != nil {
		return nil, err
	}
	// for now we'll use the api path until we get the handler logic updated
	handlerOpts := []handlers.HandlerOption{}
	if cfg.apiKey != "" {
		handlerOpts = append(handlerOpts, handlers.WithAPIKey(cfg.apiKey))
	}
	if cfg.basePath != "" {
		handlerOpts = append(handlerOpts, handlers.WithBasePath(cfg.basePath))
	}
	stHandler, err := handlers.NewStatusThingHandler(cfg.provider, handlerOpts...)
	if err != nil {
		return nil, err
	}
	return &App{config: cfg, statusThingHandler: stHandler}, nil
}

// AppConfig is the config for [App]
type AppConfig struct {
	lock        *sync.RWMutex
	provider    providers.Provider
	store       storers.StatusThingStorer
	basePath    string
	logger      *slog.Logger
	logHandler  slog.Handler
	apiKey      string
	httpServer  *http.Server
	listenAddr  string
	ngrokTunnel ngrok.Tunnel
}

func appRequestLogger(logger *slog.Logger, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("handling request", "http.path", r.URL.Path, "http.method", r.Method)
		next.ServeHTTP(w, r)
	}
}

// Start starts the app
func (a *App) Start() error {
	http.HandleFunc("/", appRequestLogger(a.config.logger, a.statusThingHandler))

	if a.config.ngrokTunnel != nil {
		go func() {
			if err := http.Serve(a.config.ngrokTunnel, appRequestLogger(a.config.logger, a.statusThingHandler)); err != nil {
				a.config.logger.Error("unable to start ngrok tunnel", "err", err)
			}
		}()
	}
	if err := a.config.httpServer.ListenAndServe(); err != nil && err.Error() != http.ErrServerClosed.Error() {
		return err
	}
	return nil
}

// Stop stops the app
func (a *App) Stop() error {
	var wg sync.WaitGroup

	if a.config.ngrokTunnel != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.config.ngrokTunnel.Close(); err != nil {
				slog.Error("unable to stop ngrok tunnel", "err", err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.config.httpServer.Close(); err != nil {
			slog.Error("unable to stop http server", "err", err)
		}
	}()
	wg.Wait()
	return nil
}
