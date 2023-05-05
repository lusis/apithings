package handlers

import (
	"fmt"
)

// HandlerOption is a functional option type
type HandlerOption func(*StatusThingHandler) error

// WithUIPath provides a custom path to serve the ui elements
func WithUIPath(path string) HandlerOption {
	return func(sth *StatusThingHandler) error {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}
		sth.uiPath = path
		return nil
	}
}

// WithAPIPath provides a custom path to serve the api
func WithAPIPath(path string) HandlerOption {
	return func(sth *StatusThingHandler) error {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}
		sth.apiPath = path
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
