package providers

import (
	"context"

	"github.com/lusis/apithings/internal/statusthing/types"
)

// Provider defines something that can provide [types.StatusThing]
type Provider interface {
	// All gets all [types.StatusThing]
	All(ctx context.Context) ([]*types.StatusThing, error)
	// Get gets a [types.StatusThing] by its id
	Get(ctx context.Context, id string) (*types.StatusThing, error)
	// Add adds a [types.StatusThing]
	Add(ctx context.Context, newThing Params) (*types.StatusThing, error)
	// Remove removes a [types.StatusThing] by its id
	Remove(ctx context.Context, id string) error
	// SetStatus sets the status of a [types.StatusThing] by its id
	SetStatus(ctx context.Context, id string, status types.Status) error
}

// Params are params that can be passed to a [Provider]
type Params struct {
	Name        string
	Description string
	Status      types.Status
}

// UnimplementedProvider is an implementation of Provider for testing and backwards compatibility
type UnimplementedProvider struct{}

// ensure we always satisfy
var _ Provider = (*UnimplementedProvider)(nil)

// All gets all [types.StatusThing]
func (up *UnimplementedProvider) All(ctx context.Context) ([]*types.StatusThing, error) {
	panic("not implemented")
}

// Get gets a [types.StatusThing] by its id
func (up *UnimplementedProvider) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	panic("not implemented")
}

// Add adds a [types.StatusThing]
func (up *UnimplementedProvider) Add(ctx context.Context, newThing Params) (*types.StatusThing, error) {
	panic("not implemented")
}

// Remove removes a [types.StatusThing] by its id
func (up *UnimplementedProvider) Remove(ctx context.Context, id string) error {
	panic("not implemented")
}

// SetStatus sets the status of a [types.StatusThing] by its id
func (up *UnimplementedProvider) SetStatus(ctx context.Context, id string, status types.Status) error {
	panic("not implemented")
}
