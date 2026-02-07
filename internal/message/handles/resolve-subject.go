package handles

import (
	"encoding/json"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/six/schedules"
	"kano/internal/utils/word"
	"os"
	"slices"
	"strconv"
	"strings"
)

type IDCode struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
}

func ResolveSubject(c *messageutil.MessageContext) error {
	args := c.Parser.NamedArgs
	if len(args) == 0 {
		c.QuoteReply("Give argument (e.g. .resolve-subject ET1201=12345 ET1202=23455)")
		return nil
	}

	fBytes, err := os.ReadFile("dumps/six/subject-id_map.json")
	if err != nil {
		c.QuoteReply("Something is wrong: %s", err)
		return err
	}

	var rep []IDCode
	err = json.Unmarshal(fBytes, &rep)
	if err != nil {
		c.QuoteReply("Something is wrong: %s", err)
		return err
	}

	for key, val := range args {
		key = strings.ToUpper(key)
		if len(val) != 1 {
			c.QuoteReply("Expected value length is 1, got %d", len(val))
			return nil
		}
		id, err := strconv.ParseUint(val[0].Content.Data, 10, 0)
		if err != nil {
			c.QuoteReply("Unable to parse %q as uint: %s", val[0].Content.Data, err)
			return nil
		}
		code := key
		if len(code) != 6 {
			c.QuoteReply("Expected code length is 6, got %d", len(code))
			return nil
		}
		validCode := word.IsCharUpper(code[0]) &&
			word.IsCharUpper(code[1]) &&
			word.IsCharNumber(code[2]) &&
			word.IsCharNumber(code[3]) &&
			word.IsCharNumber(code[4]) &&
			word.IsCharNumber(code[5])
		if !validCode {
			c.QuoteReply("Invalid code %s", code)
			return nil
		}

		for _, e := range rep {
			if e.ID == uint(id) && e.Code != code {
				c.QuoteReply("Subject id %d is already taken by %s, please check again the id for %s", id, e.Code, code)
				return nil
			}
		}

		r := IDCode{ID: uint(id), Code: code}
		idx := slices.IndexFunc(rep, func(a IDCode) bool { return a.Code == code })
		if idx != -1 {
			rep[idx] = r
		} else {
			rep = append(rep, r)
		}
	}
	slices.SortFunc(rep, func(a, b IDCode) int { return int(a.ID - b.ID) })

	mar, err := json.MarshalIndent(rep, "", "\t")
	if err != nil {
		c.QuoteReply("Something is wrong: %s", err)
		return err
	}
	err = os.WriteFile("dumps/six/subject-id_map.json", mar, 0644)
	if err != nil {
		c.QuoteReply("Something is wrong: %s", err)
		return err
	}

	schedules.UpdateSubjects()
	c.QuoteReply("Done")

	return nil
}
