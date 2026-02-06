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

	var curstate state = default_state
	// Used to check the closer of the quote
	var endquote uint8 = 0
	// Used to store temporary key name for value to insert
	var curnamed key = ""

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
		// fmt.Printf("idx=%d, char=%s, skipSpace=%d, argType=%d, recordSpace=%d, argstart=%d, endquote=%s, curnamed=%s\n",
		// 	i, string(r), curstate.doSkipAllSpace(), curstate.getArgtype(), curstate.doRecordSpace(),
		// 	argstart, string(endquote), curnamed)
		if i < argstart {
			continue
		}
		if !utf8.ValidRune(r) {
			// How does one even achieved this???
			err = fmt.Errorf("invalid character")
			return
		}
		// Skip spaces in between arguments
		if curstate.doSkipAllSpace() == 1 {
			if r > 255 && unicode.IsSpace(r) {
				continue
			}
			if r <= 255 && specialCase[uint8(r)] == 1 {
				continue
			}

			curstate.changeSkipSpace(0)
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
		} else {
			b = uint8(r)
		}

		switch specialCase[b] {
		case space:
			if curstate.doRecordSpace() == 1 { // ignore it, let the space included in the argument
				continue
			}
		case quote:
			if curstate.doRecordSpace() == 0 {
				curstate.changeRecordSpace(1)
				endquote = b
				argstart = i + 1
				continue
			} else if b != endquote { // Different closer
				continue
			} // The remaining case should be recordSpace == 1 && b == endquote
		case equal:
			switch curstate.getArgtype() {
			case normal:
				if curstate.doRecordSpace() == 0 {
					curstate.changeArgtype(namedKey)
				} else {
					continue
				}
			case namedKey:
				err = fmt.Errorf("internal error: equal: invalid argtype namedKey")
				return
			case namedValue:
				continue
			default:
				err = fmt.Errorf("internal error: equal: out of range argtype: %d", curstate.getArgtype())
				return
			}
		default:
			continue
		}

		// I assume everything works as intended
		// And now just taking the string and make the argument object
		content := p.Text[argstart:i]
		recordSpace := curstate.doRecordSpace()
		arg := Argument{
			Position:    Position{Start: argstart - int(recordSpace), End: i - 1 + int(recordSpace)},
			Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
			InsideQuote: recordSpace == 1,
			UsedQuote:   rune(endquote),
		}
		if len(content) == 0 {
			arg.Position.End += 1 - int(recordSpace)
			arg.Content.Position.Start = 0
			arg.Content.Position.End = 0
		}

		switch curstate.getArgtype() {
		case normal:
			args = append(args, arg)
			curstate.changeArgtype(normal)
			curstate.changeSkipSpace(1)
		case namedKey:
			if len(args) > 0 {
				err = fmt.Errorf("named argument is not allowed when normal argument is already given")
				return
			}

			curnamed.set(content)
			namedArgs[curnamed.val()] = append(namedArgs[curnamed.val()], Argument{})

			curstate.changeArgtype(namedValue)
			argstart = i + 1
		case namedValue:
			arg.Position.Start -= curnamed.len() + 1
			if len(content) == 0 {
				arg.Position.End -= 1
			}

			idx := len(namedArgs[curnamed.val()]) - 1
			namedArgs[curnamed.val()][idx] = arg

			curstate.changeArgtype(normal)
			curstate.changeSkipSpace(1)
			curnamed = ""
		}

		curstate.changeRecordSpace(0)
		endquote = 0
	}

	// There is unrecorded argument in the EOF
	argtype := curstate.getArgtype()
	recordSpace := curstate.doRecordSpace()

	if recordSpace == 1 {
		// It recording space, but it is already EOF
		// I assume the input has unclosed quote
		err = fmt.Errorf("unclosed quote detected, perhaps you forgot (%c)?", endquote)
		return
	}

	if argstart < len(p.Text) && specialCase[p.Text[len(p.Text)-1]] != quote {
		content := p.Text[argstart:]
		i := len(p.Text)
		arg := Argument{
			Position:    Position{Start: argstart, End: i - 1},
			Content:     Content{Data: content, Position: Position{Start: argstart, End: i - 1}},
			InsideQuote: false, UsedQuote: 0,
		}

		switch argtype {
		case normal:
			args = append(args, arg)
		case namedKey:
			err = fmt.Errorf("invalid argtype namedKey at EOF")
			return
		case namedValue:
			idx := len(namedArgs[curnamed.val()]) - 1
			namedArgs[curnamed.val()][idx] = arg
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
