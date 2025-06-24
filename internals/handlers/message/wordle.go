package message

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"kano/internals/database"
	"kano/internals/utils/kanoutils"
	"kano/internals/utils/messageutils"
	"kano/internals/utils/saveutils"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var (
	BORDER_UNIFORM = image.Uniform{color.RGBA{211, 214, 218, 255}}
	GRAY_UNIFORM   = image.Uniform{color.RGBA{120, 124, 126, 255}}
	YELLOW_UNIFORM = image.Uniform{color.RGBA{201, 180, 88, 255}}
	GREEN_UNIFORM  = image.Uniform{color.RGBA{106, 170, 100, 255}}
)

type Wordle struct {
	Word   string
	Points int
}

type WordlePoint struct {
	// Point for the guesses
	Guess int
	// Point for the correct guessed word
	Word int
	// Bonus point if the guess was correct
	Correct int
	// Bonus point for wordle streaks
	Streak int
	// Total of all the points
	Total int
}

func generateWordleImage(target string, guesses []string) ([]byte, error) {
	fontBytes, err := os.ReadFile("assets/fonts/ComicRelief-Bold.ttf")
	if err != nil {
		return nil, err
	}
	fontData, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, 1960, 2300))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fontData)
	c.SetFontSize(200)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.White))
	c.SetHinting(font.HintingNone)

	face := truetype.NewFace(fontData, &truetype.Options{
		Size: 200,
		DPI:  72,
	})
	metrics := face.Metrics()
	drawer := &font.Drawer{
		Face: face,
	}

	curPos := image.Point{150, 150} // For positioning purpose
	borderWidth := 4                // Square border length
	off := borderWidth / 2          // Just for position offset
	sLen := 300                     // Square length
	gap := 40                       // gap between the squares

	for _, text := range guesses {
		targetMaps := map[rune][]int{}
		for i, r := range target {
			l := targetMaps[r]
			if l == nil {
				targetMaps[r] = []int{i}
			} else {
				targetMaps[r] = append(l, i)
			}
		}

		for thisIdx, chr := range text {
			var uniform image.Uniform
			foundIdx := targetMaps[chr]
			if foundIdx == nil {
				uniform = GRAY_UNIFORM
			} else {
				sIdx := slices.Index(foundIdx, thisIdx)
				if sIdx == -1 {
					uniform = YELLOW_UNIFORM
					deleted := slices.Delete(foundIdx, 0, 1)
					if len(deleted) == 0 {
						delete(targetMaps, chr)
					} else {
						targetMaps[chr] = deleted
					}
				} else {
					uniform = GREEN_UNIFORM
					deleted := slices.Delete(foundIdx, sIdx, sIdx+1)
					if len(deleted) == 0 {
						delete(targetMaps, chr)
					} else {
						targetMaps[chr] = deleted
					}
				}
			}

			// Background
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+sLen+off), &uniform, image.Point{}, draw.Src)

			// Put a char
			x := drawer.MeasureString(string(chr)) >> 6
			baseline := curPos.Y + sLen - int(metrics.Descent>>6)
			c.DrawString(string(chr), freetype.Pt(curPos.X+(sLen-int(x))/2, baseline))

			// Shift the x coordinate to the right
			curPos.X += sLen + gap
		}

		curPos.Y += sLen + gap
		curPos.X = 150
	}

	for range 6 - len(guesses) {
		for range 5 {
			// Atas
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Bawah
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y+sLen-off, curPos.X+sLen+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Kiri
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Kanan
			draw.Draw(img, image.Rect(curPos.X+sLen-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)

			// Shift the x coordinate to the right
			curPos.X += sLen + gap
		}

		curPos.Y += sLen + gap
		curPos.X = 150
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func randomSelectWordle() (*Wordle, error) {
	db := database.GetDB()
	var ids []int
	rows, err := db.Query("SELECT id FROM wordle WHERE length = 5 AND lang = 'en' AND is_wordle = true")
	if err != nil {
		return nil, fmt.Errorf("randomizer: Something went wrong when retrieving dictionary status: %s", err.Error())
	}
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}
	selectedIdx := rand.Intn(len(ids))
	selectedWorldeID := ids[selectedIdx]
	var wordle Wordle
	err = db.QueryRow("SELECT word, points FROM wordle WHERE id = $1", selectedWorldeID).Scan(&wordle.Word, &wordle.Points)
	if err != nil {
		return nil, fmt.Errorf("randomizer: Something went wrong when retrieving a word: %s", err.Error())
	}

	return &wordle, nil
}

func countWordPoint(word string) int {
	point := 0
	pointMaps := map[rune]int{
		'A': 1,
		'B': 3,
		'C': 3,
		'D': 2,
		'E': 1,
		'F': 4,
		'G': 2,
		'H': 4,
		'I': 1,
		'J': 8,
		'K': 5,
		'L': 1,
		'M': 3,
		'N': 1,
		'O': 1,
		'P': 3,
		'Q': 10,
		'R': 1,
		'S': 1,
		'T': 1,
		'U': 1,
		'V': 4,
		'W': 4,
		'X': 8,
		'Y': 4,
		'Z': 10,
	}

	for _, rn := range word {
		score := pointMaps[rn]
		point += score
	}
	return point
}

func isWordExists(word string) bool {
	var id int
	db := database.GetDB()
	err := db.QueryRow("SELECT id FROM wordle WHERE word = $1", strings.ToLower(word)).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dict, _ := kanoutils.FindDefinition(word)
			if dict != nil && len(dict.Results) != 0 {
				db.Exec("INSERT INTO wordle VALUES (DEFAULT, $1, $2, 'en', 5, false)", strings.ToLower(word), countWordPoint(word))
				return true
			}
		}

		return false
	} else {
		return true
	}
}

func deleteIndex(slice []int, index int) []int {
	if index < 0 || index >= len(slice) {
		return slice // atau panic, tergantung kebutuhanmu
	}
	return append(slice[:index], slice[index+1:]...)
}

func determineWordlePoints(target string, settings *saveutils.ContactSettings) WordlePoint {
	var point WordlePoint

	if settings == nil {
		return point
	}
	if len(settings.WordleGuesses) == 0 {
		return point
	}
	// 0 for gray, 1 for yellow, 2 for green
	wordState := make([]int, len(target))
	targetMaps := map[rune][]int{}
	for i, r := range target {
		targetMaps[r] = append(targetMaps[r], i)
	}

	// Calculate for the guess point
	for idx, guess := range settings.WordleGuesses {
		localTarget := map[rune][]int{}
		for k, v := range targetMaps {
			copy(localTarget[k], v)
		}

		// First pass: Search for the green
		for runeIdx := range guess {
			chr := rune(guess[runeIdx])
			// Ignore if the current idx at guess is not same as target
			if chr != rune(target[runeIdx]) {
				continue
			}
			// Check if previous guess already correct at this index
			if wordState[runeIdx] == 2 {
				continue
			}

			wordState[runeIdx] = 2
			point.Guess += 2 * (5 - idx)
			if idx == 5 {
				// An exception for the last guess
				// It has additional 1 point
				point.Guess += 1
			}

			// Remove the index from the targetMaps
			targetMaps[chr] = deleteIndex(targetMaps[chr], slices.Index(targetMaps[chr], runeIdx))
			if len(targetMaps[chr]) == 0 {
				delete(targetMaps, chr)
			}
		}

		// Second pass: Do for the yellow and gray parts
		for runeIdx := range guess {
			// Already green, skip
			if wordState[runeIdx] == 2 {
				continue
			}

			chr := rune(guess[runeIdx])
			if idxs, ok := targetMaps[chr]; ok && len(idxs) > 0 {
				// It is yellow
				point.Guess += 5 - runeIdx

				// Remove 1 item
				targetMaps[chr] = idxs[1:]
				if len(targetMaps[chr]) == 0 {
					delete(targetMaps, chr)
				}
			}
		}
	}

	// Check for the last guess
	// I think it is guaranteed that guesses atleast has 1 string
	foundIdx := slices.Index(settings.WordleGuesses, target)
	if foundIdx != -1 {
		// We will add for the word point too for the correct answer :D
		point.Correct = 10 * (5 - foundIdx)
		point.Word = countWordPoint(target)
	}

	// Streaks point
	// Streaks point only will be calculated if the streaks is more than 2
	if settings.WordleStreaks > 2 {
		point.Streak = 5 * settings.WordleStreaks
	}

	point.Total = point.Correct + point.Guess + point.Streak + point.Word

	return point
}

func wordleMainLogic(guess, target string, guesses []string, settings *saveutils.ContactSettings) (imgBytes []byte, caption string, err error) {
	guessesCount := len(guesses)
	guessAlreadyCorrect := guessesCount > 0 && guesses[guessesCount-1] == target
	if guess == "" {
		imgBytes, err = generateWordleImage(target, guesses)
		if guessAlreadyCorrect {
			caption = "Tebakanmu hari ini udah bener :D"
			return
		} else {
			if guessesCount == 6 {
				caption = fmt.Sprintf("Kesempatan menebakmu hari ini udah habis, coba lagi besok yaw ~\nJawabannya: %s", target)
				return
			}
		}
		if guessesCount == 0 {
			caption = "Kirim .wordle KATA untuk mulai menebak"
		} else {
			caption = "Kirim .wordle KATA untuk lanjut menebak"
		}
		return
	}

	if guessAlreadyCorrect {
		caption = "Tebakan kamu hari ini udah bener :D"
		imgBytes, err = generateWordleImage(target, guesses)
		return
	}

	if guessesCount >= 6 {
		caption = fmt.Sprintf("Kesempatan menebakmu hari ini udah habis, coba lagi besok ya ~\nJawabannya: %s", target)
		imgBytes, err = generateWordleImage(target, guesses)
		return
	}
	if !isWordExists(guess) {
		caption = fmt.Sprintf("Kata %s tidak ada di kamus gweh", guess)
		return
	}

	guesses = append(guesses, guess)
	guessesCount = len(guesses)
	settings.WordleGuesses = guesses
	imgBytes, err = generateWordleImage(target, guesses)
	if guess == target {
		switch guessesCount {
		case 1:
			caption = "Weh keren, sekali nebak doang euy"
		case 6:
			caption = "Mantap, pas pasan banget"
		default:
			caption = "Sip, tebakanmu udah bener!"
		}

		settings.WordleStreaks++
		point := determineWordlePoints(target, settings)

		caption += fmt.Sprintf("\n\n*===== Point stats =====*\nPoint kata: %d\nPoin nebak: %d\nPoin bonus benar: %d\nPoin streak: %d\n*=======================*\nPoin total: %d", point.Word, point.Guess, point.Correct, point.Streak, point.Total)
		settings.GamePoints += point.Total

		return
	} else {
		if guessesCount == 6 {
			caption = fmt.Sprintf("Yah, kesempatanmu udah habis :(\nJawabannya: %s", target)

			point := determineWordlePoints(target, settings)
			caption += fmt.Sprintf("\n\n*===== Point stats =====*\nPoint kata: %d\nPoin nebak: %d\nPoin bonus benar: %d\nPoin streak: %d\n*=======================*\nPoin total: %d", point.Word, point.Guess, point.Correct, point.Streak, point.Total)

			settings.GamePoints += point.Total
		} else {
			caption = fmt.Sprintf("Tebakanmu masih salah nih, baru tebakan ke %d dari 6 kali", guessesCount)
		}
		return
	}
}

func WorldeHandler(ctx *MessageContext) {
	if ctx.Instance.Contact == nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengambil data kontak Anda", true)
		return
	}

	now := time.Now()
	var settings saveutils.ContactSettings
	if ctx.Instance.Contact.Settings != nil {
		settings = *ctx.Instance.Contact.Settings
	} else {
		settings = saveutils.ContactSettings{}
	}

	needUpdate := false
	if settings.WordleGeneratedAt.Valid {
		lastGenerated := settings.WordleGeneratedAt.Time
		generationExpireTime := time.Date(
			lastGenerated.Year(),
			lastGenerated.Month(),
			lastGenerated.Day()+1,
			0, 0, 0, 0,
			time.UTC,
		)
		if now.Unix() >= generationExpireTime.Unix() || !settings.CurrentWordle.Valid || len(settings.WordleGuesses) > 6 {
			needUpdate = true
		}
	} else {
		needUpdate = true
	}

	var target string
	if needUpdate {
		wordle, err := randomSelectWordle()
		if err != nil {
			ctx.Instance.Reply(err.Error(), true)
			return
		}
		settings.CurrentWordle = sql.NullString{String: wordle.Word, Valid: true}
		settings.WordleGeneratedAt = sql.NullTime{Time: now, Valid: true}
		settings.WordleGuesses = []string{}
		target = wordle.Word
	} else {
		target = settings.CurrentWordle.String
	}
	target = strings.ToUpper(target)

	var guess string
	args := ctx.Parser.GetArgs()
	if len(args) > 0 {
		arg := strings.ReplaceAll(filterString(ctx.Parser.Text[args[0].Start:]), " ", "")
		if len(arg) >= 5 {
			guess = strings.ToUpper(arg)[:5]
		}
	}

	guesses := settings.WordleGuesses
	imgBytes, caption, mainErr := wordleMainLogic(guess, target, guesses, &settings)

	ctx.Instance.Contact.Settings = &settings
	err := ctx.Instance.Contact.SaveContactSettings()
	if err != nil {
		ctx.Instance.Reply(err.Error(), true)
		return
	}

	if mainErr != nil {
		ctx.Instance.Reply(mainErr.Error(), true)
	} else {
		if len(imgBytes) == 0 {
			ctx.Instance.Reply(caption, true)
		} else {
			ctx.Instance.ReplyImage(imgBytes, messageutils.ReplyImageOptions{
				Caption: caption,
				Quoted:  true,
			})
		}
	}
}
