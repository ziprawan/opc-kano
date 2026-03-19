package word

import "encoding/base64"

func ToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func FromBase64(enc string) string {
	dec, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return ""
	} else {
		return string(dec)
	}
}
