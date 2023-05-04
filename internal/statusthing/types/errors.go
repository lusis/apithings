package types

import "fmt"

var (
	// ErrAlreadyExists is the error when a record already exists
	ErrAlreadyExists = fmt.Errorf("record already exists")
	// ErrNotFound is the error when a record is not found in a storer
	ErrNotFound = fmt.Errorf("record not found")
	// ErrRequiredValueMissing is the error when a required param is missing or invalid
	ErrRequiredValueMissing = fmt.Errorf("invalid value provided")
)
