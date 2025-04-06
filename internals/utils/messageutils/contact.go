package messageutils

import (
	"database/sql"
	"time"
)

type Contact struct {
	ID         int64
	EntityID   int64
	JID        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CustomName sql.NullString
	PushName   sql.NullString
	AccountID  int64
}

// func SaveContact(jid *types.JID, pushName string) error {
// 	jidStr := jid.String()
// 	db := database.GetDB()
// 	sessionName := projectconfig.GetConfig().SessionName

// 	accRow := db.QueryRow("SELECT id FROM account WHERE name = $1", sessionName)

// 	var accountId int64

// 	err := accRow.Scan(&accountId)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			panic(errors.New("SESSION_NAME is not found in your database!"))
// 		} else {
// 			fmt.Println(err)
// 		}
// 	}

// 	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})

// 	return nil
// }
