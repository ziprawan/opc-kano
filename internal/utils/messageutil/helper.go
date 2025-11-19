package messageutil

import (
	"errors"

	"go.mau.fi/whatsmeow"
)

var (
	ErrMissingDirectPath    = errors.New("missing direct path and url field")
	ErrMissingFileEncSHA256 = errors.New("missing file enc sha 256 field")
	ErrMissingFileSHA256    = errors.New("missing file sha 256 field")
	ErrMissingMediaKey      = errors.New("missing media key field")
)

func (c MessageContext) ValidateDownloadableMessage(m whatsmeow.DownloadableMessage) error {
	if len(m.GetDirectPath()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetFileEncSHA256()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetFileSHA256()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetMediaKey()) == 0 {
		return ErrMissingDirectPath
	}

	return nil
}
