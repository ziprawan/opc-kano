package message

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"kano/internals/database"
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
	rows, err := db.Query("SELECT id FROM wordle WHERE length = 5 AND lang = 'en'")
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

func isWordExists(word string) bool {
	var id int
	db := database.GetDB()
	err := db.QueryRow("SELECT id FROM wordle WHERE word = $1", strings.ToLower(word)).Scan(&id)
	if err != nil {
		return false
	} else {
		return true
	}
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
		if guessesCount == 1 {
			caption = "Weh keren, sekali nebak doang euy"
		} else if guessesCount == 6 {
			caption = "Mantap, pas pasan banget"
		} else {
			caption = "Sip, tebakanmu udah bener!"
		}
		return
	} else {
		if guessesCount == 6 {
			caption = fmt.Sprintf("Yah, kesempatanmu udah habis :(\nJawabannya: %s", target)
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
