package core

import "errors"

var (
	ErrBadArguments = errors.New("bad arguments")
	ErrNotFound     = errors.New("not found")
)
