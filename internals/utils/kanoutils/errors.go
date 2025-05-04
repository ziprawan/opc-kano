package kanoutils

import "errors"

var (
	ErrVideoLengthIsZero = errors.New("kano_utils: Video byte length is 0")
	ErrVideoNoStreams    = errors.New("kano_utils: Video doesn't have any streams")
)
