package storers

import (
	"context"

	"github.com/lusis/apithings/internal/statusthing/storers/dbfilters"
	"github.com/lusis/apithings/internal/statusthing/types"
)

// StatusThingStorer ...
type StatusThingStorer interface {
	// Get gets a statusthing
	Get(ctx context.Context, id string) (*types.StatusThing, error)
	// GetAll gets all statusthings
	GetAll(ctx context.Context) ([]*types.StatusThing, error)
	// Insert adds a statusthing
	Insert(ctx context.Context, thing *types.StatusThing) (*types.StatusThing, error)
	// Update updates a statusthing
	Update(ctx context.Context, id string, opts ...dbfilters.Option) (*types.StatusThing, error)
	// Delete deletes a statusthing
	Delete(ctx context.Context, id string) error
}

// UnimplementedStorer is a [StatusThingStorer] implementation for testing and backwards compatibility
type UnimplementedStorer struct{}

// ensure we always satisfy
var _ StatusThingStorer = (*UnimplementedStorer)(nil)

// Get gets a statusthing
func (us *UnimplementedStorer) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	panic("not implemented")
}

// GetAll gets all statusthings
func (us *UnimplementedStorer) GetAll(ctx context.Context) ([]*types.StatusThing, error) {
	panic("not implemented")
}

// Insert adds a statusthing
func (us *UnimplementedStorer) Insert(ctx context.Context, thing *types.StatusThing) (*types.StatusThing, error) {
	panic("not implemented")
}

// Update updates a statusthing
func (us *UnimplementedStorer) Update(ctx context.Context, id string, opts ...dbfilters.Option) (*types.StatusThing, error) {
	panic("not implemented")
}

// Delete deletes a statusthing
func (us *UnimplementedStorer) Delete(ctx context.Context, id string) error {
	panic("not implemented")
}
