package image

import "errors"

var (
	ErrNotRIFF            = errors.New("given bytes is not a RIFF format")
	ErrNotWebP            = errors.New("given bytes is not a WebP format")
	ErrInvalidRIFFLength  = errors.New("length of RIFF Payload is not equal to the specified RIFF Size")
	ErrInvalidChunkName   = errors.New("chunk name has non-printable ASCII character")
	ErrInvalidChunkLength = errors.New("length of Chunk Payload is not equal to the specified Chunk Size")

	ErrInvalidChunkNameLength = errors.New("chunk name length is not 4")
	ErrNoImageData            = errors.New("given Chunks doesn't have any image data")
	ErrDoubleImageData        = errors.New("given chunks has 2 or more image data")
)

var (
	ErrNoIFDProvided      = errors.New("no IFD provided")
	ErrMissmatchIFDSize   = errors.New("number of entries of IFD missmatch with the actual directory entries count")
	ErrMissmatchEntrySize = errors.New("given size missmatch with the actual entry size")
	// ErrInvalidEntryByteLength = errors.New("given value has invalid byte length (only accepts 1, 2, and 4)")
)
