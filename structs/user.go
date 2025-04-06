package structs

import "database/sql"

type User struct {
	ID         int64          `json:"id"`
	PushName   sql.NullString `json:"push_name"`
	CustomName sql.NullString `json:"custom_name"`
	JID        string         `json:"jid"`
}
