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
	Text    string
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

	argsText := strings.TrimSpace(strings.Replace(text, command.Raw, "", 1)) // Ngilangin command dari text asli
	argsStartIdx := len(text) - len(argsText)                                // Di index ke berapa args dimulai (not trimmed)
	rawArg = Argument{
		Content:     argsText,
		Start:       argsStartIdx,
		End:         len(text) - 1,
		InsideQuote: false,
		UsedQuote:   0,
	}

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
	res.Text = text

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

// Return n-th until m-th argument and join them using single space.
// This function might ignore the quote character
func (r ParseResult) GetJoinedArg(n, m int) string {
	if n >= len(r.Args) {
		return ""
	}
	if m >= len(r.Args) {
		m = len(r.Args) - 1
	}
	if m < n {
		m = n
	}

	allArgsStr := []string{}
	for i := n; i <= m; i++ {
		allArgsStr = append(allArgsStr, r.Args[i].Content)
	}

	return strings.Join(allArgsStr, " ")
}

func (r ParseResult) GetAllJoinedArg() string {
	return r.GetJoinedArg(0, len(r.Args)-1)
}

// Return n-th until m-th argument and join them using original space
func (r ParseResult) GetOriginalArg(n, m int) string {
	if n >= len(r.Args) {
		return ""
	}
	if m >= len(r.Args) {
		m = len(r.Args) - 1
	}
	if m < n {
		m = n
	}

	first := r.Args[n]
	last := r.Args[m]
	firstIdx := first.Start
	lastIdx := last.End

	if first.InsideQuote {
		firstIdx--
	}
	if last.InsideQuote {
		lastIdx++
	}

	return r.Text[firstIdx : lastIdx+1]
}
