package handlers

import (
	"fmt"
)

// HandlerOption is a functional option type
type HandlerOption func(*StatusThingHandler) error

// WithBasePath sets a custom basepath where the handler is mounted
func WithBasePath(path string) HandlerOption {
	return func(sth *StatusThingHandler) error {
		if path == "" {
			return fmt.Errorf("path must be provided to use this option")
		}
		sth.basePath = path
		return nil
	}
}

// WithAPIKey sets the api key
func WithAPIKey(key string) HandlerOption {
	return func(sth *StatusThingHandler) error {
		if key == "" {
			return fmt.Errorf("a non-empty key must be provided")
		}
		sth.apikey = key
		return nil
	}
}
