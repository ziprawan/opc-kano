package handles

import (
	"kano/internal/utils/messageutil"
	"net/http"
	"net/url"
)

func Redirect(c *messageutil.MessageContext) error {
	queryUrl := c.Parser.RawArg.Content.Data

	u, err := url.Parse(queryUrl)
	if err != nil {
		c.QuoteReply("Given string is not parsable into a URL")
		return nil
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		c.QuoteReply("Something went wrong\nDebug: %s", err.Error())
		return err
	}

	// req.Header.Set("Host", u.Host)
	// req.Header.Set("User-Agent", "curl/8.17.0")
	// req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		c.QuoteReply("Request error\nDebug: %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("location")
	if _, err := url.Parse(loc); err == nil && loc != "" {
		c.QuoteReply("%s", loc)
	} else {
		c.QuoteReply("%s", u.String())
	}

	return nil
}
