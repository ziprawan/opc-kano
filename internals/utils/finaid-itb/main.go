package finaiditb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const BASE_URL = "https://cms-finaid.itb.ac.id/api/v1"

func FetchPath(path []string, query map[string]string) ([]byte, error) {
	pathname := strings.Join(path, "/")
	parsedUrl, err := url.Parse(fmt.Sprintf("%s%s", BASE_URL, pathname))
	if err != nil {
		return nil, err
	}

	q := parsedUrl.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	parsedUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Del("User-Agent")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func FetchScholarships(size int64) (*FinaidScholarshipsResponse, error) {
	includes := []string{
		"scholarshipRequirements",
		"partner", "program",
		"strataPendidikans",
		"benefitScholarships.benefit",
		"angkatans",
	}
	query := map[string]string{
		"sort":         "-id",
		"page[size]":   strconv.FormatInt(size, 10),
		"page[number]": "1",
		"include":      strings.Join(includes, ","),
		// "filter[program_id]": "2", // Filter pekerjaan only
	}

	resp, err := FetchPath([]string{"/scholarships"}, query)
	if err != nil {
		return nil, err
	}

	var data FinaidScholarshipsResponse
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
