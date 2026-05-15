package handles

import (
	"kano/internal/database/models"
	"kano/internal/utils/chatutil/contactutil"
	"kano/internal/utils/messageutil"
	"slices"

	"go.mau.fi/whatsmeow/types"
)

// A very inefficient group refresher
func RefreshGroups(c *messageutil.MessageContext) error {
	grps := []models.Group{}
	tx := db.Find(&grps)
	if tx.Error != nil {
		c.QuoteReply("Failed to get group list: %s", tx.Error)
		return tx.Error
	}

	for _, grp := range grps {
		jid := grp.JID
		dbParts := []models.Participant{}
		tx := db.Preload("Contact").Where("group_id", grp.ID).Find(&dbParts)
		if tx.Error != nil {
			c.QuoteReply("Failed to get participant list for group JID %s: %s", jid, tx.Error)
			return tx.Error
		}

		if jid.Server != types.GroupServer {
			continue
		}

		grpInfo, err := c.Client.GetGroupInfo(jid)
		if err != nil {
			c.QuoteReply("Failed to get group info for JID %s: %s", jid, tx.Error)
			return err
		}

		grp.Name = grpInfo.Name

		parts := grpInfo.Participants
		for _, part := range parts {
			idx := slices.IndexFunc(dbParts, func(p models.Participant) bool {
				return p.Contact.JID.ToNonAD().String() == part.JID.ToNonAD().String()
			})

			contact, err := contactutil.Init(part.JID, "")
			if err != nil {
				c.QuoteReply("Failed to get/create contact info for JID %s: %s", part.JID, err)
				return err
			}
			var role models.ParticipantRole = models.ParticipantRoleMember
			if part.IsSuperAdmin {
				role = models.ParticipantRoleSuperadmin
			} else if part.IsAdmin {
				role = models.ParticipantRoleAdmin
			}
			model := models.Participant{
				GroupID:   grp.ID,
				ContactID: contact.ID,
				Role:      role,
			}

			if idx == -1 {
				tx = db.Create(model)
				if tx.Error != nil {
					c.QuoteReply("Failed to save participant with contact %s group %s: %s", contact.JID, grp.JID, tx.Error)
				}
			} else {
				model.ID = dbParts[idx].ID
				tx = db.Save(model)
				if tx.Error != nil {
					c.QuoteReply("Failed to update participant with contact %s group %s: %s", contact.JID, grp.JID, tx.Error)
				}
			}
		}
	}

	return nil
}
