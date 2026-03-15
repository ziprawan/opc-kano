package sawit

import (
	"kano/internal/utils/messageutil"
	"math"
	"math/rand"
	"time"
)

const GROW_PROB = 0.9

func Grow(c *messageutil.MessageContext) error {
	partId, err := c.GetParticipantID()
	if err != nil {
		c.QuoteReply("%s", err)
		return err
	}

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	now := time.Now().UTC()
	nowDateStr := now.Format("02-01-2006")
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	diff := tomorrow.Sub(now).Abs()
	hour := int(math.Floor(diff.Hours())) % 24
	minute := int(math.Floor(diff.Minutes())) % 60

	foundSawit, err := GetParticipantSawit(partId)
	if err != nil {
		c.QuoteReply("Failed to get participant's sawit: %s", err)
		return err
	}

	if nowDateStr == foundSawit.LastGrowDate {
		c.QuoteReply("You already grew your sawit today.\nWait for *%dh %dm*", hour, minute)
		return nil
	}

	size := r.Intn(19) + 2
	status := "grown"

	isGrow := r.Float32() < GROW_PROB
	if !isGrow {
		status = "shrunk"
		size = -size
	}

	isForced := false

	if foundSawit.Height < 0 {
		isForced = true
		isGrow = true   // Force grow
		status = "grow" // Following the isGrow
		if size < -100 {
			size = 100
		} else {
			size = -foundSawit.Height // Let the height to be 0
		}
	}

	foundSawit.AddHeight(size)
	foundSawit.ChangeGrowDate(nowDateStr)
	err = foundSawit.Save()
	if err != nil {
		c.QuoteReply("Failed to save participant's sawit: %s", err)
		return err
	}

	position, err := GetParticipantPosition(c.Group.ID, partId)
	if err != nil {
		c.QuoteReply("Failed to get participant sawit's rank position: %s", err)
		return err
	}

	if isForced {
		if size == 100 {
			c.QuoteReply(
				"Dawg, why are your sawit is so deep, I can only give you 100 cm for now. Your sawit now is *%d cm* height\nYour position in the top is %d.\n\nNext grower in *%dh %dm*",
				foundSawit.Height, position, hour, minute,
			)
		} else {
			c.QuoteReply(
				"Your sawit height is negative, huh? I think I will just make it *0 cm* height.\nYour position in the top is %d.\n\nNext grower in *%dh %dm*",
				position, hour, minute,
			)
		}
	} else {
		c.QuoteReply(
			"Your sawit has %s by *%d cm* and now it is has *%d cm* height.\nYour position in the top is %d.\n\nNext grower in *%dh %dm*",
			status, int(math.Abs(float64(size))), foundSawit.Height, position, hour, minute,
		)
	}
	return nil
}
