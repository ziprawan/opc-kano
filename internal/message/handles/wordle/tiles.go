package wordle

type tilecolor uint

const (
	_ tilecolor = iota
	gray
	yellow
	green
)

type tiles struct {
	Tiles []tilecolor
	Point uint // TODO: Add point for each green, yellow, and/or gray tiles
}

func generateTiles(target, guess string) (tiles, bool) {
	t := tiles{}
	if len(target) != len(guess) {
		return t, false
	}

	n := len(target)
	t.Tiles = make([]tilecolor, n)

	// targetMaps contains how many that character used in the word
	targetMaps := map[uint8]int{}
	for i := range n {
		targetMaps[target[i]]++
	}

	// Search for the green tiles first
	for i := range n {
		if target[i] == guess[i] {
			targetMaps[target[i]]--
			t.Tiles[i] = green
		}
	}

	// Search for the yellow and green
	for i := range n {
		if t.Tiles[i] == green { // Skip the green one, because it cannot be changed tho. Green is green
			continue
		}

		b := guess[i]
		if targetMaps[b] > 0 {
			// Char still exists but in the wrong position
			// I am thinking that it is not possible the char has the right position
			// since we already check it in the previous loop for finding the green tiles
			targetMaps[b]--
			t.Tiles[i] = yellow
		} else {
			// Otherwise, the target word doesn't contain this character
			t.Tiles[i] = gray
		}
	}

	return t, true
}
