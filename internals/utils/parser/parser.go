package parser

import (
	"strings"
	"unicode"
)

type Parser struct {
	Prefixes []string
	Text     string
}

type Command struct {
	Command     string
	UsedPrefix  string
	FullCommand string
}

type Args struct {
	Content string
	Start   int
	End     int
}

func InitParser(prefixes []string, text string) (Parser, error) {
	if strings.TrimSpace(text) == "" {
		return Parser{}, ErrEmptyText
	}

	if len(prefixes) < 1 {
		return Parser{}, ErrEmptyPrefixes
	}

	return Parser{Prefixes: prefixes, Text: text}, nil
}

func (p Parser) GetCommand() Command {
	for _, prefix := range p.Prefixes {
		if strings.HasPrefix(p.Text, prefix) {
			firstSplit := strings.Fields(p.Text)[0]
			command := strings.Replace(firstSplit, prefix, "", 1)

			return Command{Command: command, UsedPrefix: prefix, FullCommand: firstSplit}
		}
	}

	return Command{}
}

func (p Parser) GetArgs() []Args {
	command := p.GetCommand()

	if command.Command == "" {
		return []Args{}
	}

	argsText := strings.TrimSpace(strings.Replace(p.Text, command.FullCommand, "", 1))
	args := []Args{}

	var inQuotes bool
	var currentArg string
	var quoteChar rune
	var argStart, start int

	// Initialize index pointing
	argStart = strings.Index(p.Text, argsText)
	start = argStart

	for idx, char := range argsText {
		if inQuotes {
			if char == quoteChar {
				end := argStart + idx - 1
				inQuotes = false

				args = append(args, Args{Content: currentArg, Start: start, End: end})
				currentArg = ""
			} else {
				if currentArg == "" {
					start = argStart + idx
				}

				currentArg += string(char)
			}
		} else {
			if char == '\'' || char == '"' {
				inQuotes = true
				quoteChar = char
			} else if unicode.IsSpace(char) {
				if currentArg != "" {
					end := argStart + idx - 1

					args = append(args, Args{Content: currentArg, Start: start, End: end})
					currentArg = ""
				}
			} else {
				if currentArg == "" {
					start = argStart + idx
				}

				currentArg += string(char)
			}
		}
	}

	if currentArg != "" {
		end := len(argsText) + argStart - 1
		args = append(args, Args{Content: currentArg, Start: start, End: end})
	}

	return args
}

func (p Parser) Tagged() (res []string) {
	var tags []string
	var currentTag string
	var inTag bool

	for _, c := range p.Text {
		if inTag {
			if unicode.IsSpace(c) {
				if currentTag != "" {
					tags = append(tags, strings.ToLower(currentTag))
				}
				inTag = false
				currentTag = ""
			} else {
				currentTag += string(c)
			}
		} else {
			if c == '@' {
				inTag = true
			}
		}
	}

	if currentTag != "" {
		tags = append(tags, strings.ToLower(currentTag))
	}

	allKeys := make(map[string]bool)
	for _, item := range tags {
		if _, val := allKeys[item]; !val {
			allKeys[item] = true
			res = append(res, item)
		}
	}

	return
}
