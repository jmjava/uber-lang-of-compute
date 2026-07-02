package store

import "errors"

var (
	errMissingPath     = errors.New("store path is required for sqlite backend")
	errMissingEndpoint = errors.New("store endpoint is required for tsdb backend")
)
