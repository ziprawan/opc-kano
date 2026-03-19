package wordle

import (
	"os"
	"strings"
	"testing"
)

func TestTiles(t *testing.T) {
	tests := map[string][]struct {
		word   string
		output []tilecolor
	}{}

	// Get them from:
	// https://github.com/yukosgiti/wordle-tests/blob/main/data/tests.txt
	testBytes, err := os.ReadFile("tests.txt")
	if err != nil {
		t.Errorf("Failed to read tests.txt")
		return
	}

	for _, line := range strings.Split(string(testBytes), "\n") {
		if line == "" {
			continue
		}

		splits := strings.Split(line, ",")
		target := splits[0]
		guess := splits[1]
		output := splits[2]

		theOutput := make([]tilecolor, len(output))
		for i := range len(output) {
			if output[i] == 'c' {
				theOutput[i] = green
			} else if output[i] == 'm' {
				theOutput[i] = yellow
			} else if output[i] == 'w' {
				theOutput[i] = gray
			}
		}

		tests[target] = append(tests[target], struct {
			word   string
			output []tilecolor
		}{
			word:   guess,
			output: theOutput,
		})
	}

	for target, guesses := range tests {
		t.Run(target, func(t *testing.T) {
			for _, guess := range guesses {
				tileResult, ok := generateTiles(target, guess.word)
				if !ok {
					t.Errorf("%s vs %s: generateTiles returned not ok", target, guess.word)
					t.SkipNow()
				}

				for i := range tileResult.Tiles {
					if tileResult.Tiles[i] != guess.output[i] {
						t.Errorf("%s vs %s: tiles #%d: expected tile %d, got %d", target, guess.word, i, tileResult.Tiles[i], guess.output[i])
					}
				}
			}
		})
	}
}
