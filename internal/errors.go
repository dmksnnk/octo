package internal

import (
	"errors"
)

var (
	// ErrNotFound is returned by database when a resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrNotAvailable is returned when a product is not available for booking.
	ErrNotAvailable = errors.New("not available")
)
