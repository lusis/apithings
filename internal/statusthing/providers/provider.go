package providers

import (
	"context"
	"fmt"

	"github.com/lusis/apithings/internal/statusthing/storers"
	"github.com/lusis/apithings/internal/statusthing/storers/dbfilters"
	"github.com/lusis/apithings/internal/statusthing/types"

	"github.com/segmentio/ksuid"
)

// StatusThingProvider is an implementation of the [Provider] interface
type StatusThingProvider struct {
	store  storers.StatusThingStorer
	idFunc func() string
}

// NewStatusThingProvider returns a new StatusThingProvider backed by the provided store using ksuid for id generation
func NewStatusThingProvider(store storers.StatusThingStorer) (*StatusThingProvider, error) {
	if store == nil {
		return nil, fmt.Errorf("store cannot be nil")
	}
	return &StatusThingProvider{
		store: store,
		idFunc: func() string {
			return ksuid.New().String()
		},
	}, nil
}

// Get gets a [types.StatusThing] by its id
func (stp *StatusThingProvider) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	return stp.store.Get(ctx, id)
}

// All gets all [types.StatusThing]
func (stp *StatusThingProvider) All(ctx context.Context) ([]*types.StatusThing, error) {
	return stp.store.GetAll(ctx)
}

// Add adds a [types.StatusThing]
func (stp *StatusThingProvider) Add(ctx context.Context, newThing Params) (*types.StatusThing, error) {
	if newThing.Status == types.StatusUnknown {
		return nil, fmt.Errorf("a valid status must be provided: %w", types.ErrRequiredValueMissing)
	}
	if newThing.Name == "" {
		return nil, fmt.Errorf("name cannot be empty: %w", types.ErrRequiredValueMissing)
	}
	if newThing.Description == "" {
		return nil, fmt.Errorf("description cannot be empty: %w", types.ErrRequiredValueMissing)
	}
	return stp.store.Insert(ctx, &types.StatusThing{
		ID:          stp.idFunc(),
		Name:        newThing.Name,
		Description: newThing.Description,
		Status:      newThing.Status,
	})
}

// Remove removes a [types.StatusThing] by its id
func (stp *StatusThingProvider) Remove(ctx context.Context, id string) error {
	return stp.store.Delete(ctx, id)
}

// SetStatus sets the status of a [types.StatusThing] by its id
func (stp *StatusThingProvider) SetStatus(ctx context.Context, id string, status types.Status) error {
	_, err := stp.store.Update(ctx, id, dbfilters.WithStatus(status))
	return err
}
