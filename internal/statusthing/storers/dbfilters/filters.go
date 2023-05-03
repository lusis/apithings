package dbfilters

import (
	"sync"

	"github.com/lusis/apithings/internal/statusthing/types"
)

// Filters contains all the populated filters
// Filters should *ALWAYS* be created via [New]
type Filters struct {
	lock sync.RWMutex
	// thingStatus is the placeholder for a single service status
	thingStatus types.Status
}

// Option is a functional option for [Filters]
type Option func(*Filters) error

// New returns a new Filters from the provided options
func New(opts ...Option) (*Filters, error) {
	f := makeNewFilters()
	for _, o := range opts {
		f.lock.Lock()
		if err := o(f); err != nil {
			f.lock.Unlock()
			return nil, err
		}
		f.lock.Unlock()
	}
	return f, nil
}

// makeNewFilters initializes a [Filters] safely
func makeNewFilters() *Filters {
	return &Filters{
		thingStatus: types.StatusUnknown,
	}
}
