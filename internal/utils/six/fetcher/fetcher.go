package fetcher

import (
	"bytes"
	"fmt"
	"io"
	"kano/internal/config"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const BASE_URL = "https://six.itb.ac.id"

var appContext = ""

func BuildUrl(path []string, queries map[string][]string) *url.URL {
	u, _ := url.Parse(BASE_URL)
	u = u.JoinPath(path...)

	q := u.Query()
	for key, vals := range queries {
		for _, val := range vals {
			q.Add(key, val)
		}
	}
	u.RawQuery = q.Encode()

	return u
}

func processFetch(resp *http.Response) (*url.URL, *goquery.Selection, error) {
	p, _ := io.ReadAll(resp.Body)

	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == "khongguan" {
			fmt.Println("Server sent a new credential, saving")
			curCookie.Set(c.Value)
		}
	}

	loc := resp.Request.URL
	if resp.StatusCode > 299 {
		return loc, nil, fmt.Errorf("server responded with status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(p))
	if err != nil {
		return loc, nil, err
	}

	homeHref, ok := doc.Find(`[title="Home"]`).Attr("href")
	if ok {
		u, err := url.Parse(BASE_URL + homeHref)
		if err != nil {
			appContext = ""
		} else {
			q := u.Query()
			appContext = q.Get("context")
		}
	} else {
		appContext = ""
	}

	return loc, doc.Selection, err
}

func GetPage(paths []string, queries map[string][]string) (*url.URL, *goquery.Selection, error) {
	retries := 5
	sleepTime := 5 * time.Second
	cookie := ReadCookie()

	for retries > 0 {
		client := http.Client{}
		req, err := http.NewRequest("GET", BuildUrl(paths, queries).String(), nil)
		if err != nil {
			return nil, nil, err
		}
		req.AddCookie(&http.Cookie{Name: "khongguan", Value: cookie.Get()})
		req.Header.Set("User-Agent", config.USER_AGENT)

		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			time.Sleep(sleepTime)
			retries--
			continue
		}

		loc, sel, err := processFetch(resp)

		if appContext == "" {
			return loc, sel, ErrInvalidCredential
		}

		return loc, sel, err
	}

	return nil, nil, fmt.Errorf("request failed after 5 retries")
}

// Build /app/<appContext>/<paths...>
func BuildAppPath(paths []string) ([]string, error) {
	if appContext == "" {
		_, _, err := GetPage(nil, nil)
		if err != nil {
			return nil, err
		}
	}

	res := make([]string, len(paths)+2)
	res[0] = "app"
	res[1] = appContext

	for i, path := range paths {
		res[i+2] = path
	}

	return res, nil
}

// Build /app/<appContext>+<sems>/<paths...>
func BuildAppPathWithSems(paths []string, sems string) ([]string, error) {
	if appContext == "" {
		_, _, err := GetPage(nil, nil)
		if err != nil {
			return nil, err
		}
	}

	res := make([]string, len(paths)+2)
	res[0] = "app"
	res[1] = fmt.Sprintf("%s+%s", appContext, sems)

	for i, path := range paths {
		res[i+2] = path
	}

	return res, nil
}
