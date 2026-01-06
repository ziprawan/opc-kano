package parser

const (
	_ uint8 = iota
	space
	quote
	equal
)

var specialCase = [256]uint8{
	'\t': space, '\n': space, '\v': space, '\f': space, '\r': space, ' ': space, // Spaces
	'\'': quote, '"': quote, '`': quote, // Quotes
	'=': equal, // Equals sign
}

type Parser struct {
	prefixes []string
}

type Position struct {
	Start, End int
}

type Content struct {
	Data string
	Position
}

type Command struct {
	UsedPrefix string
	Name       Content
	Raw        Content
}

type Argument struct {
	Position // Including the quote character if it is used or the name of argument if the argument is named
	Content  Content

	InsideQuote bool // Must be true if it is quoted argument
	UsedQuote   rune // word.IsQuote() must return true if it is quoted argument
}

type NamedArgument map[string][]Argument

type ParseResult struct {
	isCommandParsed bool
	isArgParsed     bool

	Text    string
	Command Command

	Args      []Argument
	NamedArgs NamedArgument
	RawArg    Argument

	// Tagged  []string
}

type state uint8

const (
	// Nothing
	_ state = iota // 000
	// "Normal" argument. Support spaces if it started with quote right after a space.
	normal // 001
	// Named argument, where "key" is located at the left side of equals sign.
	// If the key started with quote, don't mark this as named argument
	namedKey // 010
	// Named argument, where "value" is located at the right side of equals sign.
	// Support spaces if it started by quote, right after the equlas sign.
	namedValue // 011
)
