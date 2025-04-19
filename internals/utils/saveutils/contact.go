package saveutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"os"

	"go.mau.fi/whatsmeow/types"
)

const INSERT_CONTACT_QUERY string = `
INSERT INTO
	"contact" (
		"entity_id",
		"account_id",
		"jid",
		"custom_name",
		"push_name"
	)
VALUES
	($1, $2, $3, $4, $5)
ON CONFLICT (ENTITY_ID) DO UPDATE
SET
	ACCOUNT_ID = EXCLUDED.ACCOUNT_ID,
	ENTITY_ID = EXCLUDED.ENTITY_ID,
	JID = EXCLUDED.JID,
	CUSTOM_NAME = EXCLUDED.CUSTOM_NAME,
	PUSH_NAME = EXCLUDED.PUSH_NAME
RETURNING
	ID,
	ENTITY_ID,
	ACCOUNT_ID,
	CREATED_AT,
	UPDATED_AT,
	JID,
	CUSTOM_NAME,
	PUSH_NAME,
	LOGIN_REQUEST_ID,
	LOGIN_EXPIRATION_DATE,
	LOGIN_REDIRECT
`
const UPDATE_CONTACT_QUERY string = `
UPDATE "contact"
SET
	ACCOUNT_ID = $1,
	ENTITY_ID = $2,
	JID = $3,
	CUSTOM_NAME = $4,
	PUSH_NAME = $5,
	UPDATED_AT = NOW()
WHERE
	ID = $6`

func getContact(jid *types.JID) (*Contact, error) {
	if jid == nil {
		return nil, ErrNilArguments
	}

	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting account data:", err)
		return nil, err
	}

	var contact Contact
	var db_jid string
	err = db.QueryRow("SELECT id, entity_id, account_id, created_at, updated_at, jid, custom_name, push_name, login_request_id, login_expiration_date, login_redirect FROM contact WHERE account_id = $1 AND jid = $2", acc.ID, jid).Scan(&contact.ID, &contact.EntityID, &contact.AccountID, &contact.CreatedAt, &contact.UpdatedAt, &db_jid, &contact.CustomName, &contact.PushName, &contact.LoginRequestID, &contact.LoginExpirationDate, &contact.LoginRedirect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		fmt.Fprintln(os.Stderr, "Error getting contact:", err)
		return nil, err
	}
	parsedJID, err := types.ParseJID(db_jid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing JID:", err)
		return nil, err
	}
	contact.JID = &parsedJID

	return &contact, nil
}

func SaveOrUpdateContact(contact *Contact) (*Contact, error) {
	if contact == nil {
		return nil, ErrNilArguments
	}

	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting account data:", err)
		return nil, err
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting transaction:", err)
		return nil, err
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	var entId int64
	scannedEntId, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM entity e WHERE account_id = $1 AND jid = $2", acc.ID, contact.JID))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning entity ID:", err)
		return nil, err
	}
	if scannedEntId == nil {
		err = tx.QueryRow("INSERT INTO entity VALUES (DEFAULT, 'CONTACT'::chat_type, $2, $1) ON CONFLICT (jid, account_id) DO UPDATE SET jid = $2 RETURNING id", acc.ID, contact.JID).Scan(&entId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Fprintln(os.Stderr, "Error inserting entity:", err)
				return nil, err
			}
		}
	} else {
		entId = *scannedEntId
	}

	var contactId int64
	scannedContactId, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM contact c WHERE account_id = $1 AND entity_id = $2", acc.ID, entId))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error scanning contact ID:", err)
		return nil, err
	}
	if scannedContactId == nil {
		// It doesn't exists, create it
		err = tx.QueryRow(INSERT_CONTACT_QUERY, entId, acc.ID, contact.JID, contact.CustomName, contact.PushName).Scan(&contactId, &entId, &acc.ID, &contact.CreatedAt, &contact.UpdatedAt, &contact.JID, &contact.CustomName, &contact.PushName, &contact.LoginRequestID, &contact.LoginExpirationDate, &contact.LoginRedirect)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error inserting contact:", err)
			return nil, err
		}
	} else {
		contactId = *scannedContactId
		// It exists, update it
		_, err = tx.Exec(UPDATE_CONTACT_QUERY, acc.ID, entId, contact.JID, contact.CustomName, contact.PushName, contactId)
	}

	return contact, err
}

func GetOrSaveContact(jid *types.JID) (*Contact, error) {
	if jid == nil {
		return nil, ErrNilArguments
	}
	foundContact, err := getContact(jid)
	if err != nil {
		return nil, err
	}

	if foundContact == nil {
		fmt.Println("Contact not found")
		contact := &Contact{
			JID: jid,
		}
		return SaveOrUpdateContact(contact)
	}
	fmt.Println("Contact found")
	return foundContact, nil
}
