package platform

import (
	"crypto/rand"
	"encoding/base32"
)

// Must is a helper function that panics if the error is not nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// ToPtr converts a value to a pointer.
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtr converts a pointer to a value.
// If the pointer is nil, it returns the zero value.
func FromPtr[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}

var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// RandString generates a random string encoded in base32.
func RandString() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	return base32Encoding.EncodeToString(b)
}
