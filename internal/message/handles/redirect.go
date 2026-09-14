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

var RedirectMan = CommandMan{
	Name:     "redirect - get url redirect",
	Synopsis: []string{"*redirect* _url_"},
	Description: []string{
		"Retrieves the redirect from the specified URL. A redirect, also known as URL forwarding, is typically indicated by an HTTP 3xx response code and the presence of a `Location` header.",
		"At this time, the bot does not support client-side redirection, such as manual redirects using `window.location.href` or similar mechanisms.",
		"*url*" +
			"\n{SPACE}The HTTP URL from which the redirect result should be retrieved. If no redirection is found, the bot will return the original URL.",
		"*Planned improvement:*" +
			"\nIf the Location header returned by the server contains only a path (i.e., does not include a protocol and host, for example: /the-path?query=value), the bot should automatically prepend the protocol and host based on the provided URL.",
	},
	SourceFilename: "redirect.go",
	SeeAlso: []SeeAlso{
		{"https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Redirections", SeeAlsoTypeExternalLink},
	},
}
