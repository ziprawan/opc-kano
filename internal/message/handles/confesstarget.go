package handles

import (
	"context"
	"database/sql"
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"slices"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ConfessTargetHandler(c *messageutil.MessageContext) error {
	if c.Group != nil {
		return nil
	}

	part, err := gorm.G[models.Participant](db).
		Joins(clause.InnerJoin.Association("Group"), models.NoopJoin).
		Joins(clause.InnerJoin.Association("GroupSettings"), models.NoopJoin).
		Where(`"Group".is_announcement != TRUE`).
		Where(`"GroupSettings".is_confess_allowed = TRUE`).
		Where("contact_id = ?", c.Contact.ID).
		Where("role != ?", models.ParticipantRoleLeft).
		Find(context.Background())

	if err != nil {
		c.QuoteReply("Internal error.\nDebug: %s", err)
		return err
	}
	if len(part) == 0 {
		c.QuoteReply("You are not joining groups or not all group allowing confess.")
		return nil
	}

	settingsIdx := -1
	if c.Contact.ConfessTarget.Valid {
		settingsIdx = slices.IndexFunc(part, func(p models.Participant) bool { return p.GroupID == uint(c.Contact.ConfessTarget.Int32) })
		if settingsIdx == -1 {
			c.Contact.ConfessTarget = sql.NullInt32{} // Invalidate the settings
			c.Contact.Save()
		}
	}

	args := c.Parser.Args
	id := (uint64)(0)
	if len(args) > 0 {
		id, _ = strconv.ParseUint(args[0].Content.Data, 10, 0)
	}

	idx := slices.IndexFunc(part, func(p models.Participant) bool { return p.GroupID == uint(id) })

	if idx == -1 {
		var msg strings.Builder
		msg.WriteString("Use `.confesstarget group_id` to set the confess target.")
		if c.Contact.ConfessTarget.Valid {
			fmt.Fprintf(&msg, "\nYour current selection: %d - %q", part[settingsIdx].GroupID, part[settingsIdx].Group.Name)
		}

		msg.WriteString("\n\nAvailable groups:")
		for _, p := range part {
			fmt.Fprintf(&msg, "\n%d - %q", p.GroupID, p.Group.Name)
		}

		c.QuoteReply("%s", msg.String())
	} else {
		c.Contact.ConfessTarget.Valid = true
		c.Contact.ConfessTarget.Int32 = int32(part[idx].GroupID)
		err := c.Contact.Save()
		if err != nil {
			c.QuoteReply("Failed to save contact info: %s", err)
			return err
		}
		c.QuoteReply("Succesfully set the confess target into %d - %q", part[idx].GroupID, part[idx].Group.Name)
	}

	return nil
}

var ConfessTargetMan = CommandMan{
	Name:           "",
	Synopsis:       []string{},
	Description:    []string{},
	SourceFilename: "confesstarget.go",
	SeeAlso:        []SeeAlso{},
}
