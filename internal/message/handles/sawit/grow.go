package sawit

import (
	"errors"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"math"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

const GROW_PROB = 0.9

var db = database.GetInstance()

func Grow(c *messageutil.MessageContext) error {
	participant, err := c.Group.GetParticipantByContactId(c.Contact.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			grpInfo, err := c.Client.GetGroupInfo(c.GetChat())
			if err != nil {
				c.QuoteReply("Failed to get group participants: %s", err)
				return err
			}
			err = c.Group.UpdateParticipantList(grpInfo)
			if err != nil {
				c.QuoteReply("Failed to update group participants: %s", err)
				return err
			}

			participant, err = c.Group.GetParticipantByContactId(c.Contact.ID)
			if err != nil {
				c.QuoteReply("Failed to get participant info: %s", err)
				return err
			}
		} else {
			c.QuoteReply("Failed to get participant info: %s", err)
			return err
		}
	}
	partId := participant.ID

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	now := time.Now().UTC()
	nowDateStr := now.Format("02-01-2006")
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	diff := tomorrow.Sub(now).Abs()
	hour := int(math.Floor(diff.Hours())) % 24
	minute := int(math.Floor(diff.Minutes())) % 60

	foundSawit := models.Sawit{ParticipantId: partId}
	tx := db.
		Where(&foundSawit).
		FirstOrCreate(&foundSawit)
	if err := tx.Error; err != nil {
		c.QuoteReply("Failed to get participant's sawit: %s", err)
		return err
	}

	if nowDateStr == foundSawit.LastGrowDate {
		c.QuoteReply("You already growed your sawit today.\nWait for *%dh %dm*", hour, minute)
		return nil
	}

	size := r.Intn(18) + 2
	status := "grown"

	isGrow := r.Float32() < GROW_PROB
	if !isGrow {
		status = "shrunk"
		size = -size
	}

	foundSawit.Height += size
	foundSawit.LastGrowDate = nowDateStr
	tx = db.Save(&foundSawit)
	if err := tx.Error; err != nil {
		c.QuoteReply("Failed to save participant's sawit: %s", err)
		return err
	}

	position, err := GetParticipantPosition(c.Group.ID, partId)
	if err != nil {
		c.QuoteReply("Failed to get participant sawit's rank position: %s", err)
		return err
	}

	c.QuoteReply(
		"Your sawit has %s by *%d cm* and now it is has *%d cm* height.\nYour position in the top is %d.\n\nNext grower in *%dh %dm*",
		status, int(math.Abs(float64(size))), foundSawit.Height, position, hour, minute,
	)
	return nil
}
