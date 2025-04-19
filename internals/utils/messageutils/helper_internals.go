package messageutils

import (
	"errors"
	"kano/internals/database"
)

var (
	ErrUnsupportedJidServer = errors.New("unsupported jid server")
)

// Internal function of SaveToDatabase()
// Parameter required
func saveToDatabase(m Message) error {
	db := database.GetDB()

	raw, err := m.Marshal()
	if err != nil {
		return err
	}

	if m.Group != nil {
		entId := m.Group.EntityID

		_, err := db.Exec("INSERT INTO \"message\" VALUES (DEFAULT, DEFAULT, DEFAULT, $1, $2, $3, DEFAULT, $4)", m.ID(), entId, raw, m.Text())
		if err != nil {
			return err
		}

		return nil
	} else if m.Contact != nil {
		entId := m.Contact.EntityID

		_, err := db.Exec("INSERT INTO \"message\" VALUES (DEFAULT, DEFAULT, DEFAULT, $1, $2, $3, DEFAULT, $4)", m.ID(), entId, raw, m.Text())
		if err != nil {
			return err
		}

		return nil
	} else {
		return ErrUnsupportedJidServer
	}
}
