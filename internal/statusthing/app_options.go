package statusthing

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/storers"

	"golang.ngrok.com/ngrok"
	"golang.org/x/exp/slog"
)

// AppOption is a functional option for customizing an [AppConfig]
type AppOption func(*AppConfig) error

// WithProvider provides a custom [providers.Provider] implementation
func WithProvider(p providers.Provider) AppOption {
	return func(ac *AppConfig) error {
		if p == nil {
			return fmt.Errorf("provider cannot be nil")
		}
		if ac.store != nil {
			return fmt.Errorf("cannot provide both storer and provider options")
		}
		ac.provider = p
		return nil
	}
}

// WithStorer provides a custom [storers.StatusThingStorer] implementation
func WithStorer(s storers.StatusThingStorer) AppOption {
	return func(ac *AppConfig) error {
		if s == nil {
			return fmt.Errorf("storer cannot be nil")
		}
		if ac.provider != nil {
			return fmt.Errorf("cannot provide both storer and provider options")
		}
		ac.store = s
		return nil
	}
}

// WithLogger provides a custom [slog.Logger] to use
func WithLogger(l *slog.Logger) AppOption {
	return func(ac *AppConfig) error {
		if l == nil {
			return fmt.Errorf("logger cannot be nil")
		}
		if ac.logHandler != nil {
			return fmt.Errorf("cannot provide both logger and loghandler options")
		}
		ac.logger = l
		return nil
	}
}

// WithLogHandler provides a custom [slog.Handler] to configure the logger with
func WithLogHandler(lh slog.Handler) AppOption {
	return func(ac *AppConfig) error {
		if lh == nil {
			return fmt.Errorf("log handler cannot be nil")
		}
		if ac.logger != nil {
			return fmt.Errorf("cannot provide both logger and loghandler options")
		}
		ac.logHandler = lh
		return nil
	}
}

// WithUIPath sets a custom path to serve any UI elements
func WithUIPath(p string) AppOption {
	return func(ac *AppConfig) error {
		if p == "" {
			return fmt.Errorf("ui path cannot be empty")
		}
		ac.uiPath = p
		return nil
	}
}

// WithAPIPath sets a custom path to serve the api
func WithAPIPath(p string) AppOption {
	return func(ac *AppConfig) error {
		if p == "" {
			return fmt.Errorf("api path cannot be empty")
		}
		ac.apiPath = p
		return nil
	}
}

// WithAPIKey sets the optional key to protect the api
func WithAPIKey(p string) AppOption {
	return func(ac *AppConfig) error {
		if p == "" {
			return fmt.Errorf("apikey cannot be empty")
		}
		ac.apiKey = p
		return nil
	}
}

// WithNgrok serves the app from the provided ngrok tunnel as well
func WithNgrok(tun ngrok.Tunnel) AppOption {
	return func(ac *AppConfig) error {
		if tun == nil {
			return fmt.Errorf("tunnel cannot be nil")
		}
		ac.ngrokTunnel = tun
		return nil
	}
}

// parseOpts parses options and returns a config
func parseOpts(opts ...AppOption) (*AppConfig, error) {
	ac := &AppConfig{
		lock:     &sync.RWMutex{},
		provider: nil,
		store:    nil,
		uiPath:   "/statusthing/",
		apiPath:  "/statusthing/api/",
		logger:   nil,
		// default to json + info level + source metadata
		logHandler: nil,
		// by default no api key
		apiKey: "",
		// all interfaces port 9000
		listenAddr: ":9000",
		httpServer: nil,
	}

	ac.lock.Lock()
	for _, o := range opts {
		if err := o(ac); err != nil {
			ac.lock.Unlock()
			return nil, fmt.Errorf("unable to build config: %w", err)
		}
	}

	// logger first since we need to pass it around
	if ac.logger == nil {
		// build logger from handler
		if ac.logHandler == nil {
			ac.logHandler = slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}.NewJSONHandler(os.Stdout)
		}
		l := slog.New(ac.logHandler)
		slog.SetDefault(l)
		ac.logger = l
	}
	if ac.httpServer == nil {
		ac.httpServer = &http.Server{Addr: ac.listenAddr}
	}

	if ac.provider == nil && ac.store == nil {
		ac.lock.Unlock()
		return nil, fmt.Errorf("either an implementation of providers.Provider or an implementation of storers.StatusThingStorer must be provided")
	}

	if ac.provider == nil {
		// build default provider from store
		p, err := providers.NewStatusThingProvider(ac.store)
		if err != nil {
			ac.lock.Unlock()
			return nil, fmt.Errorf("unable to create provider from provided store: %w", err)
		}
		ac.provider = p
	}

	ac.lock.Unlock()
	return ac, nil
}
