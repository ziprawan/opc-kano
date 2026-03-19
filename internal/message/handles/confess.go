package handles

import (
	"context"
	"database/sql"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"slices"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Confess(c *messageutil.MessageContext) error {
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

	grpJid := types.EmptyJID
	if c.Contact.ConfessTarget.Valid {
		idx := slices.IndexFunc(part, func(p models.Participant) bool { return p.GroupID == uint(c.Contact.ConfessTarget.Int32) })
		if idx == -1 {
			c.Contact.ConfessTarget = sql.NullInt32{} // Invalidate the settings
			c.Contact.Save()
		} else {
			grpJid = part[idx].Group.JID
		}
	}
	if grpJid == types.EmptyJID && len(part) == 1 {
		grpJid = part[0].Group.JID
	}

	if grpJid == types.EmptyJID {
		c.QuoteReply("I seem to see you in multiple groups, please set it up first using `.confesstarget`")
		return nil
	}

	args := c.Parser.RawArg.Content.Data

	// Goosebump
	pl := &waE2E.Message{
		Conversation: proto.String("Incoming confess!"),
	}
	resp, err := c.Client.SendMessage(grpJid, pl)
	if err != nil {
		return err
	}

	ctxInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(resp.ID),
		Participant:   proto.String(resp.Sender.ToNonAD().String()),
		QuotedMessage: pl,
	}

	// Retrieving the message object
	additionalMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(args),
			ContextInfo: ctxInfo,
		},
	}
	additionalExists := false
	canGoWithoutArgs := false
	theConfess := c.Event.Message
	if quoted := theConfess.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage(); quoted != nil {
		theConfess = quoted
	}
	if conv := theConfess.GetConversation(); conv != "" {
		theConfess = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(args),
				ContextInfo: ctxInfo,
			},
		}
	} else if ext := theConfess.GetExtendedTextMessage(); ext != nil {
		ext.ContextInfo = ctxInfo
	} else if img := theConfess.GetImageMessage(); img != nil {
		img.Caption = proto.String(args)
		img.ContextInfo = ctxInfo
		canGoWithoutArgs = true
	} else if vid := theConfess.GetVideoMessage(); vid != nil {
		vid.Caption = proto.String(args)
		vid.ContextInfo = ctxInfo
		canGoWithoutArgs = true
	} else if aud := theConfess.GetAudioMessage(); aud != nil {
		aud.ContextInfo = ctxInfo

		canGoWithoutArgs = true
		additionalExists = true
	} else if doc := theConfess.GetDocumentMessage(); doc != nil {
		doc.Caption = proto.String(args)
		doc.ContextInfo = ctxInfo
		canGoWithoutArgs = true
	} else if docCap := theConfess.GetDocumentWithCaptionMessage().GetMessage().GetDocumentMessage(); docCap != nil {
		docCap.Caption = proto.String(args)
		docCap.ContextInfo = ctxInfo
		canGoWithoutArgs = true
	} else if stk := theConfess.GetStickerMessage(); stk != nil {
		stk.ContextInfo = ctxInfo

		canGoWithoutArgs = true
		additionalExists = true
	}

	if !canGoWithoutArgs && len(args) == 0 {
		c.QuoteReply("Give the confess message.")
		return nil
	} else {
		if len(args) == 0 {
			additionalExists = false
		}
	}

	c.Client.SendMessage(grpJid, theConfess)
	if additionalExists {
		c.Client.SendMessage(grpJid, additionalMsg)
	}

	// mar, _ := json.MarshalIndent(part, "", "  ")
	// c.QuoteReply("```%s```", string(mar))

	return nil
}
