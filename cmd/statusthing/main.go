// Package main ...
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/lusis/apithings/internal/statusthing/handlers"
	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/storers/sqlite3"

	"golang.ngrok.com/ngrok"
	ngrokconfig "golang.ngrok.com/ngrok/config"

	"golang.org/x/exp/slog"
)

const (
	envPrefix = "STATUSTHING"
)

var (
	basePathEnvKey   = fmt.Sprintf("%s_BASEPATH", envPrefix)
	addrEnvKey       = fmt.Sprintf("%s_ADDR", envPrefix)
	apiKeyEnvKey     = fmt.Sprintf("%s_APIKEY", envPrefix)
	dbFileNameEnvKey = fmt.Sprintf("%s_DBFILE", envPrefix)
	debugEnvKey      = fmt.Sprintf("%s_DEBUG", envPrefix)
)

type config struct {
	basepath          string
	addr              string
	apikey            string
	dbfile            string
	debug             bool
	enableNgrok       bool
	ngrokEndpointName string
}

func configFromEnv() (*config, error) { // nolint: unparam
	cfg := &config{
		basepath:          "/api/statusthing/",
		addr:              "127.0.0.1:9000",
		dbfile:            "statusthing.db",
		debug:             false,
		enableNgrok:       false,
		ngrokEndpointName: "",
	}
	if os.Getenv(debugEnvKey) != "" {
		cfg.debug = true
	}
	if os.Getenv(basePathEnvKey) != "" {
		cfg.basepath = os.Getenv(basePathEnvKey)
	}
	if os.Getenv(addrEnvKey) != "" {
		cfg.addr = os.Getenv(addrEnvKey)
	}
	if os.Getenv(dbFileNameEnvKey) != "" {
		cfg.dbfile = os.Getenv(dbFileNameEnvKey)
	}
	// We support the native ngrok env var here
	// if you set it, we map it
	if os.Getenv("NGROK_AUTHTOKEN") != "" {
		cfg.enableNgrok = true
		if os.Getenv("NGROK_ENDPOINT") != "" {
			cfg.ngrokEndpointName = os.Getenv("NGROK_ENDPOINT")
		}
	}
	return cfg, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout))

	cfg, err := configFromEnv()
	if err != nil {
		logger.Error("unable to build configuration", "err", err)
		os.Exit(1)
	}
	if cfg.debug {
		h := slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}.NewJSONHandler(os.Stdout)
		logger = slog.New(h)
	}

	logger.Debug("starting up")
	db, err := sql.Open("sqlite", cfg.dbfile)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	store, err := sqlite3.New(db, true)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	provider, err := providers.NewStatusThingProvider(store)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	handler, err := handlers.NewStatusThingHandler(provider, cfg.basepath, logger, cfg.apikey)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	server := &http.Server{
		Addr: cfg.addr,
	}
	http.Handle(cfg.basepath, handler)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var ngrokTunnel ngrok.Tunnel
	if cfg.enableNgrok {
		logger.Debug("starting ngrok tunnel")
		opts := []ngrokconfig.HTTPEndpointOption{
			// listen on ssl
			ngrokconfig.WithScheme(ngrokconfig.SchemeHTTPS),
		}
		if cfg.ngrokEndpointName != "" {
			opts = append(opts, ngrokconfig.WithDomain(cfg.ngrokEndpointName))
		}
		ngrokTunnel, err := ngrok.Listen(context.Background(), ngrokconfig.HTTPEndpoint(opts...), ngrok.WithAuthtokenFromEnv())
		if err != nil {
			logger.Error("unable to create ngrok tunnel", "err", err)
			return
		}
		logger.Debug("ngrok tunnel created", "ngrok.endpoint", ngrokTunnel.URL())
		go func() {
			if err := http.Serve(ngrokTunnel, handler); err != nil {
				logger.Error("unable to start ngrok tunnel", "err", err)
			}
		}()
	}
	go func() {
		<-sigs
		ctx := context.TODO()
		var wg sync.WaitGroup
		logger.Debug("shutting down")
		if ngrokTunnel != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := ngrokTunnel.CloseWithContext(ctx); err != nil {
					slog.Error("unable to stop ngrok tunnel", "err", err)
				}
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Close(); err != nil {
				logger.Error("unable to shut down http server", "err", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := db.Close(); err != nil {
				logger.Error("unable to close db", "err", err)
			}
		}()
		wg.Wait()
	}()

	if err := server.ListenAndServe(); err != nil && err.Error() != http.ErrServerClosed.Error() {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
