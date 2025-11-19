package parser

import (
	"kano/internal/utils/word"
	"strings"
	"unicode"
)

type Parser struct {
	prefixes []string
}

type Command struct {
	Name       string
	UsedPrefix string
	Raw        string
}

type Argument struct {
	Content     string
	Start       int
	End         int
	InsideQuote bool
	UsedQuote   rune
}

type ParseResult struct {
	Command Command
	Args    []Argument
	RawArg  Argument
	// Tagged  []string
}

func Init(prefixes []string) *Parser {
	return &Parser{prefixes: prefixes}
}

func (p Parser) getCommand(text string) (res Command) {
	for _, prefix := range p.prefixes {
		if strings.HasPrefix(text, prefix) {
			noPrefix := strings.Replace(text, prefix, "", 1)
			rawCommand := strings.Fields(noPrefix)[0]
			command := strings.ToLower(rawCommand)

			res = Command{Raw: prefix + rawCommand, Name: command, UsedPrefix: prefix}
		}
	}

	return
}

func (p Parser) getArgs(text string, command Command) (rawArg Argument, args []Argument) {
	args = []Argument{}

	if command.Name == "" {
		return
	}

	argsText := strings.Replace(text, command.Raw, "", 1) // Ngilangin command dari text asli
	argsStartIdx := len(command.Raw)                      // Di index ke berapa args dimulai (not trimmed)

	var currentArgContent string = ""         // Menampung pembacaan arg
	var currentArgStartIdx int = argsStartIdx // Buat naro arg dimulai dari index keberapa
	var isInsideQuote bool = false            // Apakah arg ada di dalam quote
	var usedQuote rune = 0

	for idx, char := range argsText + " " {
		// Ada spasi! Pembacaan satu arg berhenti
		if unicode.IsSpace(char) && !isInsideQuote {
			if len(currentArgContent) == 0 {
				currentArgStartIdx = argsStartIdx + idx + 1 // +1 because current index is space
				continue
			}

			newArg := Argument{
				Content: currentArgContent,
				Start:   currentArgStartIdx,
				End:     currentArgStartIdx + len(currentArgContent) - 1,
			}

			args = append(args, newArg)

			// Reset the states
			currentArgContent = ""
			currentArgStartIdx = argsStartIdx + idx + 1 // +1 because current index is space

			continue
		}

		// idk man, this feels so bad coded
		if word.IsQuote(char) {
			isTwin := word.IsTwinQuote(usedQuote, char)
			isInsideQuote = !(isInsideQuote && isTwin)

			if isInsideQuote {
				if usedQuote == 0 { // First time registered
					usedQuote = char
					currentArgStartIdx = argsStartIdx + idx + 1
				} else { // Already registered
					currentArgContent += string(char) // Append it instead
				}
			} else {
				newArg := Argument{
					Content:     currentArgContent,
					Start:       currentArgStartIdx,
					End:         currentArgStartIdx + len(currentArgContent) - 1,
					InsideQuote: true,
					UsedQuote:   usedQuote,
				}

				args = append(args, newArg)

				// Reset states
				currentArgContent = ""
				currentArgStartIdx = argsStartIdx + idx + 1
				usedQuote = 0
			}

			continue
		}

		currentArgContent += string(char)
	}

	return
}

func (p Parser) Parse(text string) (res ParseResult) {
	text = strings.TrimSpace(text)

	res.Command = p.getCommand(text)
	if len(res.Command.Name) != 0 {
		res.RawArg, res.Args = p.getArgs(text, res.Command)
	}

	return
}

// Get argument content start and end position index
//
// It remain the same even if a.InsideQuote is true
func (a Argument) GetPosition() (int, int) {
	return a.Start, a.End
}

// Get absolute argument position (start and end index)
//
// The index counted from the quote if it exists
func (a Argument) GetAbsolutePosition() (int, int) {
	if a.InsideQuote {
		return a.Start - 1, a.End + 1
	}

	return a.Start, a.End
}
