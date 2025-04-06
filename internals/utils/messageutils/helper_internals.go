package messageutils

import (
	"kano/internals/database"

	"go.mau.fi/whatsmeow/types"
)

// Internal function of SaveToDatabase()
// Parameter required
func saveToDatabase(m Message) error {
	db := database.GetDB()
	res, err := m.SaveEntities()
	if err != nil {
		return err
	}

	server := m.ChatJID().Server
	raw, err := m.Raw()
	if err != nil {
		return err
	}

	if server == types.GroupServer {
		entId := res[0].Group.EntityID

		_, err := db.Exec("INSERT INTO \"message\" VALUES (DEFAULT, DEFAULT, DEFAULT, $1, $2, $3, DEFAULT, $4)", m.ID(), entId, raw, m.Text())
		if err != nil {
			return err
		}

		return nil
	} else if server == types.DefaultUserServer {
		entId := res[1].Contact.EntityID

		_, err := db.Exec("INSERT INTO \"message\" VALUES (DEFAULT, DEFAULT, DEFAULT, $1, $2, $3, DEFAULT, $4)", m.ID(), entId, raw, m.Text())
		if err != nil {
			return err
		}

		return nil
	} else {
		return ErrUnsupportedJidServer
	}
}
