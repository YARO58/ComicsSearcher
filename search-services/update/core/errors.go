package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrNotFound = errors.New("resource is not found")

var ErrComicsCount = errors.New("failed to get total count comics")
var ErrGetDBStats = errors.New("failed to get db stats")
var ErrGetDownloadedComics = errors.New("failed to get downloaded comics")
var ErrReadComic = errors.New("failed to read comic")
var ErrTruncateTable = errors.New("failed to truncate table")
