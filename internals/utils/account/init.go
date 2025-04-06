package account

import (
	"errors"
	"kano/internals/database"
	"sync"

	"go.mau.fi/whatsmeow/types"
)

type KanoAccount struct {
	ID       int64
	Name     string
	JID      *types.JID
	PushName string
}

var (
	accInstance *KanoAccount
	initErr     error
	once        sync.Once

	ErrAccNotFound = errors.New("cannot find account at your database")
)

func InitAccount(accountName string) *KanoAccount {
	once.Do(func() {
		var (
			acc *KanoAccount
			jid string
		)

		acc = &KanoAccount{}

		db := database.GetDB()
		stmt := db.QueryRow("SELECT * FROM account WHERE name = $1", accountName)
		err := stmt.Scan(
			&acc.ID,
			&acc.Name,
			&jid,
		)

		if err != nil {
			initErr = err
		}

		parsed, err := types.ParseJID(jid)
		if err != nil {
			initErr = err
		}

		acc.JID = &parsed
		accInstance = acc
	})

	return accInstance
}

func SaveAccount(accountName string, jid *types.JID) error {
	jidStr := jid.String()
	db := database.GetDB()

	var accID int64
	stmt, err := db.Prepare("INSERT INTO account VALUES (DEFAULT, $1, $2) ON CONFLICT (name) DO UPDATE SET jid = $2 RETURNING id")
	if err != nil {
		return err
	}

	err = stmt.QueryRow(accountName, jidStr).Scan(&accID)
	if err != nil {
		return err
	}

	accInstance = &KanoAccount{
		ID:   accID,
		Name: accountName,
		JID:  jid,
	}

	return nil
}

func SetPushName(pushName string) {
	if accInstance != nil {
		accInstance.PushName = pushName
	}
}

func GetData() (*KanoAccount, error) {
	return accInstance, initErr
}
