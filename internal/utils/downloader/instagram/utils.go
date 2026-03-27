package instagram

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"net/http"
)

const instagram_encoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

var instagram_table = map[uint8]int{}

func instagramIdToPk(id string) uint64 {
	if len(instagram_table) == 0 {
		for i := range len(instagram_encoding) {
			instagram_table[instagram_encoding[i]] = i
		}
	}

	result, base := uint64(0), uint64(len(instagram_table))
	for i := range len(id) {
		result = (result * base) + uint64(instagram_table[id[i]])
	}

	return result
}

func instagramApiCheck(id string) (apiCheck instagramRuling, csrftoken string) {
	// Create request
	cli := http.Client{}
	checkReq, err := instagramCreateReq(
		fmt.Sprintf(
			"https://i.instagram.com/api/v1/web/get_ruling_for_content/?content_type=MEDIA&target_id=%d",
			instagramIdToPk(id),
		),
	)
	if err != nil {
		return
	}

	// Do the request and parse it
	// If it is returned errors, just return empty struct
	checkResp, err := cli.Do(checkReq)
	if err == nil {
		// Parse the response
		json.NewDecoder(checkResp.Body).Decode(&apiCheck)
		checkResp.Body.Close()

		// Get csrf token from Set-Cookie
		setCookies := checkResp.Cookies()
		for _, cookie := range setCookies {
			if cookie == nil {
				continue
			}

			if cookie.Name == "csrftoken" {
				csrftoken = cookie.Value
			}
		}
	}

	return
}

func instagramCreateReq(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return req, err
	}

	req.Header.Set("X-IG-App-ID", "936619743392459")
	req.Header.Set("X-ASBD-ID", "198387")
	req.Header.Set("X-IG-WWW-Claim", "0")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", config.USER_AGENT)

	return req, nil
}
