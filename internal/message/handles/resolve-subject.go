package handles

import (
	"encoding/json"
	"kano/internal/config"
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
	if !c.IsSenderSame(config.GetConfig().OwnerJID) {
		return nil
	}

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

var ResolveSubjectMan = CommandMan{
	Name: "resolve-subject - resolves a class subject ID (SIX)",
	Synopsis: []string{
		"*resolve-subject* _subject_code_*=*_subject_id_",
	},
	Description: []string{
		"Assigns a subject ID based on the provided subject code. Both the subject ID and subject code can be obtained from the SiX curriculum page. This command is restricted to the bot owner and is intended for troubleshooting issues related to refreshing SiX class schedules, which may occasionally fail when newly added classes are not yet recognized by the bot. It is unclear why the system requires the original subject ID instead of relying solely on an incrementing ID.",
		"_subject_code_*=*_subject_id_" +
			"\n{SPACE}`subject_code`: The code of the subject to which the ID will be assigned. It typically consists of 6 characters, where the first 2 characters are letters (A–Z) and the remaining characters are digits." +
			"\n{SPACE}`subject_id`: The ID of the subject. This value must be a positive integer.",
		"_Note: This command will likely be moved or merged as a subcommand under the .six command in the future._",
	},
	SourceFilename: "resolve-subject.go",
	SeeAlso:        []SeeAlso{},
}
