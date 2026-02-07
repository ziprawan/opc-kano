package fetcher

import (
	"os"
)

type Cookie string

var curCookie Cookie

func ReadCookie() Cookie {
	if curCookie != "" {
		return curCookie
	}

	f, err := os.ReadFile("secrets/khongguan")
	if err != nil {
		return curCookie
	}

	n := (Cookie)(string(f))
	curCookie = n

	return curCookie
}

func (c *Cookie) Get() string {
	if c == nil {
		return ""
	}

	return string(*c)
}

func (c *Cookie) Set(newStr string) {
	cn := (Cookie)(newStr)
	*c = cn
	curCookie = cn

	os.WriteFile("secrets/khongguan", []byte(c.Get()), 0644)
}
