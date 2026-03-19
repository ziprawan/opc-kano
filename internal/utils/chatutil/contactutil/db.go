package contactutil

import (
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"

	"go.mau.fi/whatsmeow/types"
)

var db = database.GetInstance()

func initDb(jid types.JID, pushname string) (*models.Contact, error) {
	if jid.Server != types.HiddenUserServer && jid.Server != types.DefaultUserServer {
		return nil, fmt.Errorf("given jid server is not @lid")
	}

	contact := models.Contact{}
	tx := db.
		Where(models.Contact{JID: jid}).
		Assign(models.Contact{PushName: pushname}).
		FirstOrCreate(&contact)

	return &contact, tx.Error
}
