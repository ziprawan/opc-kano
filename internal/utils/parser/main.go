package parser

import (
	"fmt"
	"kano/internal/utils/word"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Most of this code are inspired (or perhaps copied)
// from golang's string package

func Init(prefixes []string) *Parser {
	return &Parser{prefixes: prefixes}
}

func (p *ParseResult) parseCommand(prefixes []string) {
	if p == nil {
		return
	}
	if p.isCommandParsed {
		return
	}
	if p.Text == "" {
		return
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(p.Text, prefix) {
			// Find the start and end index of command name
			name := Content{}
			n := len(prefix)
			// Skip spaces after prefix and before command
			for i, r := range p.Text {
				if i < n {
					continue
				}
				if !unicode.IsSpace(r) {
					n = i
					break
				}
			}
			name.Start = n
			for i, r := range p.Text {
				if i <= n {
					continue
				}
				if unicode.IsSpace(r) {
					break
				} else {
					n = i
				}
			}
			name.End = n
			name.Data = p.Text[name.Start : name.End+1]

			raw := Content{}
			raw.Data = p.Text[:name.End+1]
			raw.Start = 0
			raw.End = name.End

			p.Command = Command{UsedPrefix: prefix, Name: name, Raw: raw}
			break
		}
	}

	p.isCommandParsed = true
}

// Weird ahh code
func (p *ParseResult) parseArgs() (err error) {
	if p == nil {
		return
	}
	if !p.isCommandParsed {
		return
	}
	if p.isCommandParsed && p.Command.Name.Data == "" {
		return
	}
	if p.isArgParsed {
		return
	}

	// State is 3 bit data (00000xxy),
	// where the rightmost bit (y) is indicating if it should record space or nah.
	// The remaining bits (xx) are indicating the state (as defined with type "state")
	var curstate uint8 = (uint8(normal) << 1)
	// Used to check the closer of the quote
	var endquote uint8 = 0
	// Used to store temporary key name for value to insert
	var curnamed string = ""

	skipSpace := false
	args := []Argument{}
	namedArgs := NamedArgument{}
	argstart := 0
	i := p.Command.Raw.End + 1

	// skip after command space
	for i < len(p.Text) && specialCase[p.Text[i]] == space {
		i++
	}
	if i == len(p.Text) {
		return
	}
	p.RawArg = Argument{
		Position:    Position{Start: i, End: len(p.Text) - 1},
		Content:     Content{Data: p.Text[i:], Position: Position{Start: i, End: len(p.Text) - 1}},
		InsideQuote: false, UsedQuote: 0,
	}

	argstart = i
	var r rune
	for i, r = range p.Text {
		if i < argstart {
			continue
		}
		if !utf8.ValidRune(r) {
			// How does one even achieved this???
			err = fmt.Errorf("invalid character")
			return
		}
		// Skip spaces in between arguments
		if skipSpace {
			if r > 255 && unicode.IsSpace(r) {
				continue
			}
			if r <= 255 && specialCase[uint8(r)] == 1 {
				continue
			}

			skipSpace = false
			argstart = i
		}

		b := uint8(0)
		if r > 255 {
			if unicode.IsSpace(r) {
				b = 0x20 // Treat them as regular space
			} else {
				// Yea, ts not ASCII dawg, it is also outside of my special table
				continue
			}
		}

		b = uint8(r)
		recordSpace := curstate & 0b001
		argtype := state(curstate&0b110) >> 1

		switch specialCase[b] {
		case space:
			if recordSpace == 1 { // ignore it, let the space included in the argument
				continue
			}
		case quote:
			if recordSpace == 0 {
				if argstart != i {
					// There is an argument right before the quote and there is not space between them
					arg := Argument{
						Position:    Position{Start: argstart, End: i - 1},
						Content:     Content{Data: p.Text[argstart:i], Position: Position{Start: argstart, End: i - 1}},
						InsideQuote: false, UsedQuote: 0, // I assume it is guaranteed not an inside quote argument
					}
					switch argtype {
					case normal:
						args = append(args, arg)
					case namedValue:
						namedArgs[curnamed] = append(namedArgs[curnamed], arg)
						curstate = (uint8(normal) << 1) | recordSpace
						curnamed = ""
					}
				}

				curstate |= 0b001
				endquote = b
				argstart = i + 1
				continue
			} else if b != endquote { // Different closer
				continue
			} // The remaining case should be recordSpace == 1 && b == endquote
		case equal:
			if argtype == normal {
				if recordSpace == 0 {
					argtype = namedKey
				} else {
					continue
				}
			} else if argtype == namedKey {
				err = fmt.Errorf("internal error: equal: invalid argtype namedKey")
				return
			} else if argtype == namedValue {
				continue
			} else {
				err = fmt.Errorf("internal error: equal: out of range argtype: %d", argtype)
			}
		default:
			continue
		}

		// I assume everything works as intended
		// And now just taking the string and make the argument object
		content := p.Text[argstart:i]

		switch argtype {
		case normal:
			args = append(args, Argument{
				Position:    Position{Start: argstart - int(recordSpace), End: i - 1 + int(recordSpace)},
				Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
				InsideQuote: recordSpace == 1,
				UsedQuote:   rune(endquote),
			})
			argtype = normal
		case namedKey:
			_, ok := namedArgs[content]
			if !ok {
				namedArgs[content] = []Argument{}
			}

			if i+1 < len(p.Text) && specialCase[p.Text[i+1]] == space {
				argtype = normal
				curnamed = ""
			} else {
				argtype = namedValue
				curnamed = content
			}

		case namedValue:
			namedArgs[curnamed] = append(namedArgs[curnamed], Argument{
				Position:    Position{Start: argstart - int(recordSpace), End: i - 1 + int(recordSpace)},
				Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
				InsideQuote: recordSpace == 1,
				UsedQuote:   rune(endquote),
			})

			argtype = normal
			curnamed = ""
		}

		endquote = 0
		curstate = (uint8(argtype) << 1)
		skipSpace = true
	}

	// There is unrecorded argument in the EOF
	argtype := state(curstate&0b110) >> 1
	recordSpace := curstate & 0b001

	if recordSpace == 1 {
		// It recording space, but it is already EOF
		// I assume the input has unclosed quote
		err = fmt.Errorf("unclosed quote detected, perhaps you forgot %c?", endquote)
		return
	}

	if argstart < len(p.Text) && specialCase[p.Text[len(p.Text)-1]] != quote {
		content := p.Text[argstart:]
		i := len(p.Text)

		switch argtype {
		case normal:
			args = append(args, Argument{
				Position:    Position{Start: argstart, End: i - 1},
				Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
				InsideQuote: false, UsedQuote: 0,
			})
		case namedKey:
			err = fmt.Errorf("invalid argtype namedKey at EOF")
			return
		case namedValue:
			namedArgs[curnamed] = append(namedArgs[curnamed], Argument{
				Position:    Position{Start: argstart, End: i - 1},
				Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
				InsideQuote: false, UsedQuote: 0,
			})
		}
	}

	p.Args = args
	p.NamedArgs = namedArgs
	p.isArgParsed = true

	return
}

func (p Parser) Parse(text string) (res ParseResult, err error) {
	res.Text = word.NormalizeQuote(strings.TrimSpace(text))
	res.parseCommand(p.prefixes)
	err = res.parseArgs()
	if err != nil {
		return
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
		allArgsStr = append(allArgsStr, r.Args[i].Content.Data)
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

func (r ParseResult) GetAllOriginalArg() string {
	return r.GetOriginalArg(0, len(r.Args)-1)
}
