package dbfilters

import (
	"fmt"

	"github.com/lusis/apithings/internal/statusthing/types"
)

// WithStatus is a filter option to set the status of a thing
func WithStatus(status types.Status) Option {
	return func(f *Filters) error {
		if status == types.StatusUnknown {
			return fmt.Errorf("a valid status must be provided")
		}
		f.thingStatus = status
		return nil
	}
}

// Status gets the value of the [WithStatus] option
func (f *Filters) Status() types.Status {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.thingStatus
}
