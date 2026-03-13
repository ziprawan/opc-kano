package handles

import (
	"errors"
	"fmt"
	"kano/internal/database/models"
	"kano/internal/message/handles/wordle"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/word"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

func filterString(str string) string {
	var res strings.Builder
	for i := range len(str) {
		if word.IsCharUpper(str[i]) || word.IsCharLower(str[i]) {
			res.WriteByte(str[i])
		}
	}
	return res.String()
}

func WordleHandler(c *messageutil.MessageContext) error {
	if c.Group != nil && !c.Group.GroupSettings.IsGameAllowed {
		c.Logger.Debugf("Game is not allowed in %s", c.Group.JID)
		return nil
	}

	now := time.Now().UTC()
	nowStr := now.Format("02-01-2006")

	foundUserWordle := models.UserWordle{DateStr: nowStr, UserId: c.Contact.ID}
	tx := db.
		Preload("Target").
		Where(&foundUserWordle).
		// Order(clause.OrderBy{
		// 	Columns: []clause.OrderByColumn{
		// 		{Column: clause.Column{Name: "id"}, Desc: true},
		// 	},
		// }).
		First(&foundUserWordle)

	if err := tx.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.QuoteReply("Failed to get user wordle: %s", err)
			return err
		}

		theWordle, err := wordle.RandomSelectWordle()
		if err != nil {
			c.QuoteReply("%s", err)
			return err
		}

		foundUserWordle.TargetId = theWordle.ID
		foundUserWordle.Guesses = []string{}
		tx = db.Preload("Target").Create(&foundUserWordle)
		err = tx.Error
		if err != nil {
			c.QuoteReply("Failed to set user wordle: %s", err)
			return err
		}
		foundUserWordle.Target = theWordle
	}

	caption := ""

	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	diff := tomorrow.Sub(now).Abs()
	hour := int(math.Floor(diff.Hours())) % 24
	minute := int(math.Floor(diff.Minutes())) % 60

	waitFor := fmt.Sprintf("%02dh%02dm", hour, minute)

	lg := len(foundUserWordle.Guesses)
	target := strings.ToUpper(foundUserWordle.Target.Word)
	foundUserWordle.Target = models.Wordle{} // Reset, so it won't overwrite "wordle" table at insert query

	if lg > 0 && strings.ToUpper(foundUserWordle.Guesses[lg-1]) == target {
		caption = fmt.Sprintf("Your wordle is correct for today (%s), please wait for %s", target, waitFor)
	} else if lg >= 6 {
		caption = fmt.Sprintf("Your attempts is over, the answer is %s", target)
	} else {
		args := c.Parser.Args
		if len(args) > 0 {
			guess := strings.ToUpper(filterString(c.Parser.Args[0].Content.Data))
			if len(guess) > 5 {
				guess = guess[:5]
			}

			if len(guess) < 5 {
				c.QuoteReply("Word length is too short")
				return nil
			} else {
				if !wordle.IsWordExists(guess) {
					c.QuoteReply("Word %q doesn't exists", guess)
					return nil
				}

				lg++
				foundUserWordle.Guesses = append(foundUserWordle.Guesses, guess)
				tx := db.Model(&models.UserWordle{}).Where("id = ?", foundUserWordle.ID).Update("guesses", foundUserWordle.Guesses)
				if tx.Error != nil {
					c.QuoteReply("failed to save user wordle guesses: %s", tx.Error)
					return tx.Error
				}

				if guess == target {
					switch lg {
					case 1:
						caption = "Impressive, done in just one guess"
					case 6:
						caption = "Your guess was just right, luckily it was right"
					default:
						caption = "Yeay, your guess was correct!"
					}
				} else {
					switch lg {
					case 6:
						caption = fmt.Sprintf("The guess is over, the answer is %s", target)
					default:
						caption = fmt.Sprintf("Yah, your guess is wrong, try again (%d/6)", lg)
					}
				}
			}
		} else {
			if len(foundUserWordle.Guesses) == 0 {
				caption = "Send .wordle [YOUR_GUESS] to guess the word for today!"
			} else {
				caption = "Send .wordle [YOUR_GUESS] to continue your guess!"
			}
		}
	}

	imgBytes, err := wordle.GenerateWordleImage(target, foundUserWordle.Guesses)
	if err != nil {
		c.QuoteReply("failed to generate wordle image: %s", err)
		return fmt.Errorf("failed to generate wordle image: %s", err)
	}

	c.ReplyImage(imgBytes, true, caption)

	return nil
}
