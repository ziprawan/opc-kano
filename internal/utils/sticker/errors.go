package sticker

import "errors"

var (
	ErrUnsupportedFile = errors.New("given image type is unsupported (currently only supports JPEG, PNG, and WebP)")
	ErrNoStream        = errors.New("given file has no stream")
	ErrDurationTooLong = errors.New("given video has duration more than 10 seconds")
	ErrNoResult        = errors.New("conversion returned null")
)
