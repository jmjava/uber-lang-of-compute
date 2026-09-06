package store

import "errors"

var (
	errMissingPath     = errors.New("store path is required for sqlite backend")
	errMissingEndpoint = errors.New("store endpoint is required for tsdb backend")
	// ErrSealedOverwrite is returned when a sealed snapshot would be mutated.
	ErrSealedOverwrite = errors.New("store: refusing to overwrite sealed snapshot")
	// ErrMemoConflict is returned when a memo key already holds a different output.
	ErrMemoConflict = errors.New("store: memo key already holds a different output")
)
