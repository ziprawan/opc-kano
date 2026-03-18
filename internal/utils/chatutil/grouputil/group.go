package grouputil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/chatutil/communityutil"
	"kano/internal/utils/chatutil/contactutil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

var caches = map[string]*models.Group{}
var log = config.GetLogger().Sub("GroupUtil")

type Group struct {
	models.Group

	GroupSettings *GroupSettings
}

func Init(cli *whatsmeow.Client, groupJid types.JID) (*Group, error) {
	group, err := InitDb(cli, groupJid)
	if group != nil {
		settings, err := InitSettings(group.ID)
		if err != nil {
			return &Group{}, err
		}
		return &Group{Group: *group, GroupSettings: settings}, nil
	}
	return &Group{}, err
}

func InitDb(cli *whatsmeow.Client, groupJid types.JID) (*models.Group, error) {
	if c, o := caches[groupJid.String()]; o && c != nil {
		log.Debugf("Found group data at cache with name %s", c.Name)
		return c, nil
	}

	if groupJid.Server != types.GroupServer {
		log.Errorf("given jid server is not a group")
		return nil, fmt.Errorf("given jid server is not a group")
	}

	db := database.GetInstance()
	grp := models.Group{JID: groupJid}
	tx := db.Where(&grp).First(&grp)

	if tx.Error == nil {
		log.Debugf("Found group data at database with name %s", grp.Name)
		return &grp, nil
	}

	// Has another error than ErrRecordNotFound
	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to get group data from database: %s", tx.Error.Error())
		return nil, tx.Error
	}

	grpInfo, err := cli.GetGroupInfo(context.Background(), groupJid)
	if err != nil {
		log.Errorf("Failed to get group info: %s", err.Error())
		return nil, err
	}

	if grpInfo.LinkedParentJID.Server == types.GroupServer {
		comm, err := communityutil.Init(cli, grpInfo.LinkedParentJID)
		if err != nil {
			log.Errorf("Failed to init community: %s", err.Error())
		}
		grp.CommunityID = sql.NullInt64{Valid: true, Int64: int64(comm.ID)}
	}

	grp.Name = grpInfo.Name
	if len(grpInfo.Participants) > 0 {
		log.Debugf("Got %d participant(s)", len(grpInfo.Participants))
		participants := make([]types.JID, len(grpInfo.Participants))
		for i, p := range grpInfo.Participants {
			jid := p.JID
			if jid.Server != types.HiddenUserServer {
				jid = p.LID
			}
			if jid.Server != types.HiddenUserServer {
				return nil, fmt.Errorf("given participant jid server is not @lid: %s", jid)
			}

			participants[i] = jid
		}

		ids, err := contactutil.GetIDs(participants)
		if err != nil {
			log.Errorf("Failed to get contact IDs: %s", err.Error())
			return nil, err
		}

		partModels := make([]models.Participant, len(participants))
		for i, p := range grpInfo.Participants {
			jid := p.JID
			if jid.Server != types.HiddenUserServer {
				jid = p.LID
			}
			if jid.Server != types.HiddenUserServer {
				return nil, fmt.Errorf("given participant jid server is not @lid: %s", jid)
			}

			partModels[i].ContactID = ids[jid]
			partModels[i].Role = models.ParticipantRoleMember
			if p.IsAdmin {
				partModels[i].Role = models.ParticipantRoleAdmin
			}
			// Not using else if, because IsAdmin is true when IsSuperAdmin is true
			if p.IsSuperAdmin {
				partModels[i].Role = models.ParticipantRoleSuperadmin
			}
		}

		grp.Participants = partModels
	}

	tx = db.Create(&grp)

	if tx.Error != nil {
		log.Errorf("Failed to insert group to database: %s", tx.Error)
		return nil, tx.Error
	}

	caches[groupJid.String()] = &grp
	return &grp, nil
}

func Insert(grpInfo types.GroupInfo) (*models.Group, error) {
	if grpInfo.IsParent {
		return nil, fmt.Errorf("jid is community, not group")
	}

	groupJid := grpInfo.JID
	if c, o := caches[groupJid.String()]; o && c != nil {
		log.Debugf("Found group data at cache with name %s", c.Name)
		return c, nil
	}

	if groupJid.Server != types.GroupServer {
		log.Errorf("given jid server is not a group")
		return nil, fmt.Errorf("given jid server is not a group")
	}

	db := database.GetInstance()
	grp := models.Group{JID: groupJid}
	tx := db.Where(&grp).First(&grp)

	if tx.Error == nil {
		log.Debugf("Found group data at database with name %s", grp.Name)
		return &grp, nil
	}

	// Has another error than ErrRecordNotFound
	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to get group data from database: %s", tx.Error.Error())
		return nil, tx.Error
	}

	if grpInfo.LinkedParentJID.Server == types.GroupServer {
		comm, err := communityutil.Get(grpInfo.LinkedParentJID)
		if err != nil {
			log.Errorf("Failed to get community info: %s", err.Error())
		}
		grp.CommunityID = sql.NullInt64{Valid: true, Int64: int64(comm.ID)}
	}

	grp.Name = grpInfo.Name
	if len(grpInfo.Participants) > 0 {
		log.Debugf("Got %d participant(s)", len(grpInfo.Participants))
		participants := make([]types.JID, len(grpInfo.Participants))
		ids, err := contactutil.GetIDs(participants)
		if err != nil {
			log.Errorf("Failed to get contact IDs: %s", err.Error())
			return nil, err
		}

		partModels := make([]models.Participant, len(participants))
		i := 0
		for _, id := range ids {
			partModels[i].ContactID = id
			i++
		}

		grp.Participants = partModels
	}

	tx = db.Create(&grp)

	if tx.Error != nil {
		log.Errorf("Failed to insert group to database: %s", tx.Error)
		return nil, tx.Error
	}

	caches[groupJid.String()] = &grp
	return &grp, nil
}

func Update(grpInfo types.GroupInfo) (*models.Group, error) {
	if grpInfo.IsParent {
		return nil, fmt.Errorf("jid is community, not group")
	}

	groupJid := grpInfo.JID
	if c, o := caches[groupJid.String()]; o && c != nil {
		log.Debugf("Found group data at cache with name %s", c.Name)
		return c, nil
	}

	if groupJid.Server != types.GroupServer {
		log.Errorf("given jid server is not a group")
		return nil, fmt.Errorf("given jid server is not a group")
	}

	db := database.GetInstance()
	grp := models.Group{JID: groupJid}
	tx := db.Where(&grp).First(&grp)

	if tx.Error == nil {
		log.Debugf("Found group data at database with name %s", grp.Name)
		return &grp, nil
	}

	// Has another error than ErrRecordNotFound
	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to get group data from database: %s", tx.Error.Error())
		return nil, tx.Error
	}

	if grpInfo.LinkedParentJID.Server == types.GroupServer {
		comm, err := communityutil.Get(grpInfo.LinkedParentJID)
		if err != nil {
			log.Errorf("Failed to get community info: %s", err.Error())
		}
		grp.CommunityID = sql.NullInt64{Valid: true, Int64: int64(comm.ID)}
	}

	grp.Name = grpInfo.Name
	if len(grpInfo.Participants) > 0 {
		log.Debugf("Got %d participant(s)", len(grpInfo.Participants))
		participants := make([]types.JID, len(grpInfo.Participants))
		ids, err := contactutil.GetIDs(participants)
		if err != nil {
			log.Errorf("Failed to get contact IDs: %s", err.Error())
			return nil, err
		}

		partModels := make([]models.Participant, len(participants))
		i := 0
		for _, id := range ids {
			partModels[i].ContactID = id
			i++
		}

		grp.Participants = partModels
	}

	tx = db.Create(&grp)

	if tx.Error != nil {
		log.Errorf("Failed to insert group to database: %s", tx.Error)
		return nil, tx.Error
	}

	caches[groupJid.String()] = &grp
	return &grp, nil
}

func (g Group) GetParticipantRole(contactId uint) (models.ParticipantRole, error) {
	part := models.Participant{GroupID: g.ID, ContactID: contactId}
	db := database.GetInstance()
	tx := db.
		Where("group_id = ?", g.ID).
		Where("contact_id = ?", contactId).
		Assign(models.Participant{Role: models.ParticipantRoleMember}).
		FirstOrCreate(&part)
	if tx.Error != nil {
		return models.ParticipantRoleLeft, tx.Error
	}
	fmt.Printf("%+v\n", part)

	return part.Role, nil
}
