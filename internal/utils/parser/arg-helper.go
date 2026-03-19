package parser

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
