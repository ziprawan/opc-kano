package contactutil

import (
	"errors"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

var log = config.GetLogger().Sub("ContactUtil")

type Contact struct {
	ID         uint
	JID        types.JID
	Pushname   string
	CustomName string
}

func Init(jid types.JID, pushname string) (Contact, error) {
	contact := Contact{}
	model, err := initDb(jid, pushname)
	if err != nil {
		return contact, err
	}

	contact.ID = model.ID
	contact.JID = model.JID
	contact.Pushname = model.PushName
	contact.CustomName = model.CustomName

	return contact, nil
}

func GetIDs(jids []types.JID) (map[types.JID]uint, error) {
	if len(jids) == 0 {
		return map[types.JID]uint{}, nil
	}

	db := database.GetInstance()
	var exists []models.Contact
	tx := db.Where("jid IN ?", jids).Find(&exists)
	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.Errorf("Failed to query contact IDs: %s", tx.Error.Error())
		return nil, tx.Error
	}

	fmt.Println("JIDS2", jids)

	log.Debugf("Found %d contact data(s) at database", len(exists))

	ret := map[types.JID]uint{}
	for _, ex := range exists {
		ret[ex.JID] = ex.ID
	}

	inserts := make([]models.Contact, len(jids)-len(exists))
	if len(inserts) > 0 {
		log.Infof("Need %d new inserts", len(inserts))

		i := 0
		for _, jid := range jids {
			exist := false
			for _, e := range exists {
				if e.JID == jid {
					exist = true
					break
				}
			}
			if exist {
				continue
			}

			inserts[i].JID = jid
			i++
		}
		fmt.Printf("See here blud: %+v\n", inserts)

		tx := db.Create(&inserts)
		if tx.Error != nil {
			log.Errorf("Failed to insert contact data")
			return nil, tx.Error
		}

		for _, ins := range inserts {
			ret[ins.JID] = ins.ID
		}
	}

	return ret, nil
}
