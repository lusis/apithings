package providers

import (
	"context"
	"fmt"
	"testing"

	"github.com/lusis/apithings/internal/statusthing/storers"
	"github.com/lusis/apithings/internal/statusthing/storers/dbfilters"
	"github.com/lusis/apithings/internal/statusthing/types"
	"github.com/stretchr/testify/require"
)

func TestImplements(t *testing.T) {
	t.Parallel()
	require.Implements(t, (*Provider)(nil), &StatusThingProvider{})
}

func TestHappyPath(t *testing.T) {
	thing := &types.StatusThing{
		ID:          t.Name() + "_id",
		Description: t.Name() + "_description",
		Status:      types.StatusGreen,
	}
	newStatus := types.StatusYellow
	ts := &testStorer{
		getFunc: func() (*types.StatusThing, error) { return thing, nil },
		allFunc: func() ([]*types.StatusThing, error) { return []*types.StatusThing{thing}, nil },
		insertFunc: func(thing *types.StatusThing) (*types.StatusThing, error) {
			if thing.ID == "" {
				return nil, fmt.Errorf("id was not provided")
			}
			return thing, nil
		},
		deleteFunc: func(id string) error {
			if id == "" {
				return fmt.Errorf("id was not provided")
			}
			return nil
		},
		updateFunc: func(id string, opts *dbfilters.Filters) (*types.StatusThing, error) {
			if opts.Status() == types.StatusUnknown {
				return nil, fmt.Errorf("status was not provided")
			}
			if opts.Status() != newStatus {
				return nil, fmt.Errorf("expected status not provided")
			}
			// we don't care what we return here
			return &types.StatusThing{}, nil
		},
	}
	p, err := NewStatusThingProvider(ts)
	require.NoError(t, err, "should create a provider")
	require.NotNil(t, p, "provider should not be nil")

	res, err := p.Get(context.Background(), t.Name())
	require.NoError(t, err, "get should not error")
	require.NotNil(t, res, "result should not be nil")

	allres, err := p.All(context.Background())
	require.NoError(t, err, "all should not return error")
	require.Len(t, allres, 1, "should have one result")

	insertres, err := p.Add(context.Background(), Params{Name: "othername", Description: "otherdesc", Status: types.StatusYellow})
	require.NoError(t, err, "insert should not error")
	require.NotNil(t, insertres, "result should not be nil")
	require.NotEmpty(t, insertres.ID, "id should have been generated")

	require.NoError(t, p.SetStatus(context.Background(), "fakeid", newStatus), "set status should work")
	require.NoError(t, p.Remove(context.Background(), "fakeid"), "delete should work")
}

func TestConstructor(t *testing.T) {
	p, err := NewStatusThingProvider(nil)
	require.Error(t, err, "should error")
	require.Nil(t, p, "should be nil")
}
func TestMissingParams(t *testing.T) {
	ts := &testStorer{}

	p, err := NewStatusThingProvider(ts)
	require.NoError(t, err, "should not error")
	require.NotNil(t, p, "provider should not be nil")

	testCases := map[string]struct {
		thing Params
		err   error
	}{
		"invalid-status":      {Params{Name: t.Name(), Description: t.Name()}, types.ErrRequiredValueMissing},
		"missing-name":        {Params{Status: types.StatusGreen, Description: t.Name()}, types.ErrRequiredValueMissing},
		"missing-description": {Params{Status: types.StatusGreen, Name: t.Name()}, types.ErrRequiredValueMissing},
	}
	t.Parallel()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res, err := p.Add(context.Background(), tc.thing)
			require.Nil(t, res, "should not return a result")
			require.ErrorIs(t, err, tc.err, "should be expected error type")
		})
	}
}

type testStorer struct {
	storers.UnimplementedStorer
	getFunc    func() (*types.StatusThing, error)
	allFunc    func() ([]*types.StatusThing, error)
	insertFunc func(*types.StatusThing) (*types.StatusThing, error)
	updateFunc func(id string, opts *dbfilters.Filters) (*types.StatusThing, error)
	deleteFunc func(id string) error
}

func (ts *testStorer) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	return ts.getFunc()
}

func (ts *testStorer) GetAll(ctx context.Context) ([]*types.StatusThing, error) {
	return ts.allFunc()
}

func (ts *testStorer) Insert(ctx context.Context, thing *types.StatusThing) (*types.StatusThing, error) {
	return ts.insertFunc(thing)
}

func (ts *testStorer) Delete(ctx context.Context, id string) error {
	return ts.deleteFunc(id)
}

func (ts *testStorer) Update(ctx context.Context, id string, opts ...dbfilters.Option) (*types.StatusThing, error) {
	dbopts, err := dbfilters.New(opts...)
	if err != nil {
		return nil, err
	}
	return ts.updateFunc(id, dbopts)
}
