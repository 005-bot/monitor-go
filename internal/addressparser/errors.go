package addressparser

import "errors"

var (
	ErrEmptyInput = errors.New("empty input")
	ErrNoMatch    = errors.New("no match")
)
