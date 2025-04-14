package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"kano/internals/database"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func GroupInfoHandler(client *whatsmeow.Client, event *events.GroupInfo) {
	db := database.GetDB()
	var foundGroupID uint64
	err := db.QueryRow("SELECT id FROM \"group\" WHERE jid = $1", event.JID.String()).Scan(&foundGroupID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println("Errored when querying group info from database!")
			return
		} else {
			// nigga, I don't know what should even I do???
		}
	}
}
