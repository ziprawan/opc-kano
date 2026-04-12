package handles

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/word"
	"net/http"
	"strings"
	"time"
)

type rgsiGameRating struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type rgsiRatingDescriptors struct {
	Name    string `json:"nameEn"`
	Enabled bool   `json:"enabled"`
}

type rgsiGameInfo struct {
	Id          int                     `json:"id"`
	Name        string                  `json:"name"`
	Release     int                     `json:"releaseYear"`
	Publisher   string                  `json:"publisherName"`
	Platforms   []string                `json:"platformsName"`
	Description string                  `json:"description"`
	VideoUrl    string                  `json:"videoUrl"`
	InGameUrl   string                  `json:"inGameUrl"`
	Ratings     []rgsiGameRating        `json:"ratings"`
	Descriptors []rgsiRatingDescriptors `json:"descriptors"`
}

type rgsiGameList struct {
	Result []rgsiGameInfo `json:"publicGameResultList"`
}

type rgsiResponse struct {
	Embedded rgsiGameList `json:"_embedded"`
}

func rgsiApi() string {
	return "https://" + word.Reverse(word.FromBase64("c2VtYWcvY2lsYnVwL2RpLnNyZ2kuaXBh"))
}

func Rgsi(c *messageutil.MessageContext) error {
	args := c.Parser.RawArg.Content.Data
	if len(args) == 0 {
		c.QuoteReply("Give the game name")
		return nil
	}

	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", rgsiApi(), nil)
		if err != nil {
			c.QuoteReply("Internal error while creating request object")
			return err
		}

		req.Header.Set("User-Agent", config.USER_AGENT)

		q := req.URL.Query()
		q.Set("nameLike", args)
		q.Set("page", "0")
		q.Set("size", "1000")
		req.URL.RawQuery = q.Encode()

		cli := http.Client{}

		done := make(chan struct{})
		var resp *http.Response

		go func() {
			resp, err = cli.Do(req)
			close(done)
		}()

		select {
		case <-done:
			// finished (success or error)
		case <-time.After(10 * time.Second):
			c.QuoteReply("It seems the request is taking longer than expected, please wait")
			<-done
		}

		// ---- check result ----
		if err != nil {
			if attempt == maxRetries {
				c.QuoteReply("Request failed after %d attempts: %v", attempt, err)
				return err
			}
			continue // retry
		}

		if resp != nil {
			resp.Body.Close()
		}

		if resp.StatusCode == 504 {
			if attempt == maxRetries {
				c.QuoteReply("API keeps returning 504 after %d attempts", attempt)
				return nil
			}
			continue // retry
		}

		if resp.StatusCode > 299 {
			c.QuoteReply("API returned %s", resp.Status)
			return nil
		}

		// success
		var jsResp rgsiResponse
		err = json.NewDecoder(resp.Body).Decode(&jsResp)
		if err != nil {
			c.QuoteReply("Internal error while reading the response with error %s", err)
			return err
		}

		result := jsResp.Embedded.Result
		if len(result) == 0 {
			c.QuoteReply("Game not found")
			return nil
		}

		infoStrs := make([]string, len(result))
		for i, game := range result {
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

			infoStrs[i] = builder.String()
		}

		c.QuoteReply("%s", strings.Join(infoStrs, "\n====\n\n"))
		break
	}

	return nil
}
