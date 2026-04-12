package handles

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"kano/internal/utils/messageutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RgsiId(c *messageutil.MessageContext) error {
	args := c.Parser.RawArg.Content.Data
	if len(args) == 0 {
		c.QuoteReply("Give the game id (positive number)")
		return nil
	}

	id, err := strconv.ParseUint(args, 10, 0)
	if err != nil {
		c.QuoteReply("Given ID is not a number")
	}

	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", rgsiApi()+fmt.Sprintf("/%d", id), nil)
		if err != nil {
			c.QuoteReply("Internal error while creating request object")
			return err
		}

		req.Header.Set("User-Agent", config.USER_AGENT)

		cli := http.Client{}

		type fresult struct {
			resp *http.Response
			err  error
		}

		resCh := make(chan fresult)

		go func() {
			resp, err := cli.Do(req)
			resCh <- fresult{resp: resp, err: err}
		}()

		var resp *http.Response

		select {
		case res := <-resCh:
			resp = res.resp
			err = res.err
		case <-time.After(10 * time.Second):
			c.QuoteReply("It seems the request is taking longer than expected, please wait")
			res := <-resCh
			resp = res.resp
			err = res.err
		}

		// ---- check result ----
		if err != nil {
			if attempt == maxRetries {
				c.QuoteReply("Request failed after %d attempts: %v", attempt, err)
				return err
			}
			if resp != nil {
				resp.Body.Close()
			}
			c.QuoteReply("Request failed, retrying... (%d/5)", attempt)
			continue // retry
		}

		if resp.StatusCode == 504 {
			if attempt == maxRetries {
				c.QuoteReply("API keeps returning 504 after %d attempts", attempt)
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
			c.QuoteReply("Request failed, retrying... (%d/5)", attempt)
			continue // retry
		}

		if resp.StatusCode > 299 {
			c.QuoteReply("API returned %s", resp.Status)
			return nil
		}

		// success
		var jsResp rgsiGameInfo
		err = json.NewDecoder(resp.Body).Decode(&jsResp)
		if err != nil {
			c.QuoteReply("Internal error while reading the response with error %s", err)
			return err
		}

		game := jsResp
		var builder strings.Builder

		fmt.Fprintf(&builder, "*%s - %s*\n", game.Name, game.Publisher)
		fmt.Fprintf(&builder, "> %s\n", strings.ReplaceAll(game.Description, "\n", ""))
		fmt.Fprintf(&builder, "Platforms: %s\n", strings.Join(game.Platforms, ", "))

		if len(game.Ratings) > 0 {
			ratings := make([]string, len(game.Ratings))
			for j, rating := range game.Ratings {
				if !rating.Enabled {
					ratings[j] = fmt.Sprintf("~%s~", rating.Name)
				} else {
					ratings[j] = rating.Name
				}
			}
			fmt.Fprintf(&builder, "*Ratings: %s*\n", strings.Join(ratings, ", "))
		}

		if len(game.Descriptors) > 0 {
			descriptors := make([]string, len(game.Descriptors))
			for j, desc := range game.Descriptors {
				if !desc.Enabled {
					descriptors[j] = fmt.Sprintf("- ~%s~", desc.Name)
				} else {
					descriptors[j] = fmt.Sprintf("- %s", desc.Name)
				}
			}
			fmt.Fprintf(&builder, "*Descriptors:*\n%s\n", strings.Join(descriptors, "\n"))
		}

		fmt.Fprintf(&builder, "Related URLs: %s - %s\n", game.InGameUrl, game.VideoUrl)

		c.QuoteReply("%s", builder.String())
		break
	}

	return nil
}
