package account

import (
	"errors"
	"nopi/internals/database"
	"sync"

	"go.mau.fi/whatsmeow/types"
)

type NopiAccount struct {
	ID       int64
	Name     string
	JID      *types.JID
	PushName string
}

var (
	accInstance *NopiAccount
	initErr     error
	once        sync.Once

	ErrAccNotFound = errors.New("cannot find account at your database")
)

func InitAccount(accountName string) *NopiAccount {
	once.Do(func() {
		var (
			acc *NopiAccount
			jid string
		)

		acc = &NopiAccount{}

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

	accInstance = &NopiAccount{
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

func GetData() (*NopiAccount, error) {
	return accInstance, initErr
}
