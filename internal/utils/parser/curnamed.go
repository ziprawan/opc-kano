package parser

import "strings"

type key string

func validKeyChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

func (k key) valid() bool {
	for _, r := range k {
		if !validKeyChar(r) {
			return false
		}
	}
	return true
}

// Just string to lower
func (k *key) normalize() {
	s := string(*k)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ToLower(s)

	*k = key(s)
}

func (k key) len() int {
	return len(k)
}

func (k *key) set(s string) {
	*k = key(s)
	k.normalize()
}

func (k key) val() string {
	return string(k)
}
