package parser

import "errors"

var (
	ErrEmptyPrefixes = errors.New("atleast 1 prefixes is defined")
	ErrEmptyText     = errors.New("given text is an empty string")
)
