package core

import "errors"

var (
	ErrBadArguments       = errors.New("arguments are not acceptable")
	ErrAlreadyExists      = errors.New("resource or task already exists")
	ErrMessageTooLarge    = errors.New("message is too large")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)
