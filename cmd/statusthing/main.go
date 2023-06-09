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

	"github.com/lusis/apithings/internal/statusthing"
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
	enableDashEnvKey = fmt.Sprintf("%s_ENABLE_DASH", envPrefix)
)

type config struct {
	basepath          string
	addr              string
	apikey            string
	dbfile            string
	debug             bool
	enableNgrok       bool
	ngrokEndpointName string
	enableDash        bool
}

func configFromEnv() (*config, error) { // nolint: unparam
	cfg := &config{
		basepath:          statusthing.DefaultHTTPBasePath,
		addr:              ":9000",
		dbfile:            "statusthing.db",
		debug:             false,
		enableNgrok:       false,
		ngrokEndpointName: "",
		enableDash:        false,
		apikey:            "",
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
	if os.Getenv(enableDashEnvKey) != "" {
		cfg.enableDash = true
	}
	if os.Getenv(apiKeyEnvKey) != "" {
		cfg.apikey = os.Getenv(apiKeyEnvKey)
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
	h := slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}.NewJSONHandler(os.Stdout)
	logger := slog.New(h)

	cfg, err := configFromEnv()
	if err != nil {
		logger.Error("unable to build configuration", "err", err)
		os.Exit(1)
	}
	if cfg.debug {
		h := slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}.NewJSONHandler(os.Stdout)
		logger = slog.New(h)
	}

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
	appOptions := []statusthing.AppOption{
		statusthing.WithStorer(store),
	}
	if cfg.basepath != "" {
		appOptions = append(appOptions, statusthing.WithBasePath(cfg.basepath))
	}
	if cfg.apikey != "" {
		appOptions = append(appOptions, statusthing.WithAPIKey(cfg.apikey))
	}
	if cfg.enableNgrok {
		logger.Debug("creating ngrok tunnel")
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
		logger = logger.With("ngrok.endpoint", ngrokTunnel.URL())
		logger.Debug("ngrok tunnel created")
		appOptions = append(appOptions, statusthing.WithNgrok(ngrokTunnel))
	}
	// since we decorate the logger with metadata we need to add this after all is added
	appOptions = append(appOptions, statusthing.WithLogger(logger))
	app, err := statusthing.New(appOptions...)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		var wg sync.WaitGroup
		logger.Debug("shutting down")

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.Stop(); err != nil {
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
	logger.Info("starting application", "addr", cfg.addr, "basepath", cfg.basepath)
	if err := app.Start(); err != nil && err.Error() != http.ErrServerClosed.Error() {
		logger.Error("error attempting to start application: %s", err.Error())
		os.Exit(1)
	}
	logger.Info("stopped application")
}
