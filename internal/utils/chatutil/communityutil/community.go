package communityutil

import (
	"context"
	"errors"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

var caches = map[string]*models.Community{}
var log = config.GetLogger().Sub("CommunityUtil")

func Init(cli *whatsmeow.Client, commJid types.JID) (*models.Community, error) {
	if c, o := caches[commJid.String()]; o && c != nil {
		log.Debugf("Found community data at cache with name %s", c.Name)
		return c, nil
	}

	if commJid.Server != types.GroupServer {
		log.Errorf("given jid server is not a group")
		return nil, fmt.Errorf("given jid server is not a group")
	}

	db := database.GetInstance()
	comm := models.Community{JID: commJid}
	tx := db.Where(&comm).First(&comm)

	if tx.Error == nil {
		log.Debugf("Found community data at database with name %s", comm.Name)
		return &comm, nil
	}

	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to get community data from database: %s", tx.Error.Error())
		return nil, tx.Error
	}

	commInfo, err := cli.GetGroupInfo(context.Background(), commJid)
	if err != nil {
		log.Errorf("Failed to get community info: %s", err.Error())
		return nil, nil
	}
	if commInfo.LinkedParentJID.String() != "" {
		return nil, fmt.Errorf("given jid is not a community because it has linked_parent_jid: %s", commInfo.LinkedParentJID.String())
	}

	comm.Name = commInfo.Name
	tx = db.Create(&comm)

	if tx.Error != nil {
		log.Errorf("Failed to insert community to database: %s", tx.Error)
		return nil, tx.Error
	}

	caches[commJid.String()] = &comm
	return &comm, nil
}

func Insert(commInfo types.GroupInfo) (*models.Community, error) {
	if !commInfo.IsParent {
		return nil, fmt.Errorf("jid is a group, not community")
	}
	if commInfo.LinkedParentJID.String() != "" {
		return nil, fmt.Errorf("given jid is not a community because it has linked_parent_jid: %s", commInfo.LinkedParentJID.String())
	}

	commJid := commInfo.JID
	if c, o := caches[commJid.String()]; o && c != nil {
		log.Debugf("Found community data at cache with name %s", c.Name)
		return c, nil
	}

	if commJid.Server != types.GroupServer {
		log.Errorf("given jid server is not a group")
		return nil, fmt.Errorf("given jid server is not a group")
	}

	db := database.GetInstance()
	comm := models.Community{JID: commJid}
	tx := db.Where(&comm).First(&comm)

	if tx.Error == nil {
		log.Debugf("Found community data at database with name %s", comm.Name)
		return &comm, nil
	}

	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to get community data from database: %s", tx.Error.Error())
		return nil, tx.Error
	}

	comm.Name = commInfo.Name
	tx = db.Create(&comm)

	if tx.Error != nil {
		log.Errorf("Failed to insert community to database: %s", tx.Error)
		return nil, tx.Error
	}

	caches[commJid.String()] = &comm
	return &comm, nil
}

func Get(commJid types.JID) (*models.Community, error) {
	if c, o := caches[commJid.String()]; o && c != nil {
		log.Debugf("Found community data at cache with name %s", c.Name)
		return c, nil
	}

	db := database.GetInstance()
	comm := models.Community{JID: commJid}
	tx := db.Where(&comm).First(&comm)
	if tx.Error != nil {
		return nil, tx.Error
	} else {
		return &comm, nil
	}
}
