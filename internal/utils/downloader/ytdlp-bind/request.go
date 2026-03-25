package ytdlpbind

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

var ignoreKeys = []string{
	"domain", "path", "secure", "expires",
}

func Request(format YtDlpFormat) (*http.Response, error) {
	if format.Url == "" {
		return nil, fmt.Errorf("empty url")
	}

	req, err := http.NewRequest("GET", format.Url, nil)
	if err != nil {
		return nil, err
	}

	if len(format.HttpHeaders) > 0 {
		for key, value := range format.HttpHeaders {
			req.Header.Set(key, value)
		}
		fmt.Printf("%+v\n", req.Header)
	}

	if len(format.Cookies) > 0 {
		splits := strings.SplitSeq(format.Cookies, "; ")
		for split := range splits {
			key, val, ok := strings.Cut(split, "=")
			if !ok {
				continue
			}

			if slices.Contains(ignoreKeys, strings.ToLower(key)) {
				continue
			}

			c := &http.Cookie{}
			c.Name = key
			c.Value = val

			req.AddCookie(c)
		}
		fmt.Printf("%+v\n", req.Cookies())
	}

	cli := http.Client{}
	return cli.Do(req)
}
