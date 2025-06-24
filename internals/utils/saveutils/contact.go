package saveutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"

	"github.com/lib/pq"
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
const UPDATE_CONTACT_SETTINGS string = `
INSERT INTO
	"contact_settings"
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7
)
ON CONFLICT
	("id")
DO UPDATE SET
	confess_target_id = $2,
	current_wordle = $3,
	wordle_generated_at = $4,
	wordle_guesses = $5,
	wordle_streaks = $6,
	game_points = $7`

func getContactSettings(contactID int64) (*ContactSettings, error) {
	db := database.GetDB()
	var contactSettings ContactSettings
	var wordleGuessesRaw sql.NullString
	err := db.QueryRow("SELECT confess_target_id, current_wordle, wordle_generated_at, to_json(wordle_guesses) FROM contact_settings WHERE id = $1", contactID).Scan(
		&contactSettings.ConfessTargetID,
		&contactSettings.CurrentWordle,
		&contactSettings.WordleGeneratedAt,
		&wordleGuessesRaw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if wordleGuessesRaw.Valid {
		err := json.Unmarshal([]byte(wordleGuessesRaw.String), &contactSettings.WordleGuesses)
		if err != nil {
			fmt.Println("Gagal unmarshal wordle_guesses di contact_settings", err)
		}
	}

	return &contactSettings, nil
}

func getContact(jid *types.JID) (*Contact, error) {
	if jid == nil {
		return nil, ErrNilArguments
	}

	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Println("Error getting account data:", err)
		return nil, err
	}

	var contact Contact
	var db_jid string
	err = db.QueryRow("SELECT id, entity_id, account_id, created_at, updated_at, jid, custom_name, push_name, login_request_id, login_expiration_date, login_redirect FROM contact WHERE account_id = $1 AND jid = $2", acc.ID, jid).Scan(&contact.ID, &contact.EntityID, &contact.AccountID, &contact.CreatedAt, &contact.UpdatedAt, &db_jid, &contact.CustomName, &contact.PushName, &contact.LoginRequestID, &contact.LoginExpirationDate, &contact.LoginRedirect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		fmt.Println("Error getting contact:", err)
		return nil, err
	}
	parsedJID, err := types.ParseJID(db_jid)
	if err != nil {
		fmt.Println("Error parsing JID:", err)
		return nil, err
	}

	contactSettings, err := getContactSettings(contact.ID)
	if err != nil {
		fmt.Println("Failed to get contact settings:", err)
	}
	contact.Settings = contactSettings
	contact.JID = &parsedJID

	return &contact, nil
}

func SaveOrUpdateContact(contact *Contact) (*Contact, error) {
	fmt.Println("SaveOrUpdateContact() called")
	if contact == nil {
		return nil, ErrNilArguments
	}

	fmt.Println("Getting database conn and acc data")
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Println("Error getting account data:", err)
		return nil, err
	}

	fmt.Println("Creating transaction")
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return nil, err
	}
	fmt.Println("Creating defer func")
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	fmt.Println("Getting entity ID")
	var entId int64
	scannedEntId, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM entity e WHERE account_id = $1 AND jid = $2", acc.ID, contact.JID))
	if err != nil {
		fmt.Println("Error scanning entity ID:", err)
		return nil, err
	}
	fmt.Println("Checking if entity ID is nil")
	if scannedEntId == nil {
		fmt.Println("entity ID is nil")
		err = tx.QueryRow("INSERT INTO entity VALUES (DEFAULT, 'CONTACT'::chat_type, $2, $1) ON CONFLICT (jid, account_id) DO UPDATE SET jid = $2 RETURNING id", acc.ID, contact.JID).Scan(&entId)
		if err != nil {
			fmt.Println("Error inserting entity:", err)
			return nil, err
		}
	} else {
		fmt.Println("dereferencing scannedEntId")
		entId = *scannedEntId
	}

	fmt.Println("Getting contact ID")
	var contactId int64
	scannedContactId, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM contact c WHERE account_id = $1 AND entity_id = $2", acc.ID, entId))
	if err != nil {
		fmt.Println("Error scanning contact ID:", err)
		return nil, err
	}
	if scannedContactId == nil {
		fmt.Println("Contact ID is nil")
		// It doesn't exists, create it
		var jid string // ga kepake wok
		err = tx.QueryRow(INSERT_CONTACT_QUERY, entId, acc.ID, contact.JID.String(), contact.CustomName, contact.PushName).Scan(&contactId, &entId, &acc.ID, &contact.CreatedAt, &contact.UpdatedAt, &jid, &contact.CustomName, &contact.PushName, &contact.LoginRequestID, &contact.LoginExpirationDate, &contact.LoginRedirect)
		if err != nil {
			fmt.Println("Error inserting contact:", err)
			return nil, err
		}
	} else {
		fmt.Println("Contact ID is not nil")
		contactId = *scannedContactId
		// It exists, update it
		_, err = tx.Exec(UPDATE_CONTACT_QUERY, acc.ID, entId, contact.JID.String(), contact.CustomName, contact.PushName, contactId)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("Failed to commit transaction:", err)
		return nil, err
	}

	contact.ID = contactId
	contact.EntityID = entId
	contact.AccountID = acc.ID

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

func (c *Contact) SaveContactSettings() error {
	if c == nil {
		return errors.New("contact: Contact is nil")
	}
	db := database.GetDB()
	if c.Settings == nil {
		c.Settings = &ContactSettings{}
		_, err := db.Exec("INSERT INTO contact_settings (id) VALUES id = $1", c.ID)
		if err != nil {
			return err
		}
	}
	cSettings := *c.Settings

	_, err := db.Exec(
		UPDATE_CONTACT_SETTINGS,
		c.ID,
		cSettings.ConfessTargetID,
		cSettings.CurrentWordle,
		cSettings.WordleGeneratedAt,
		pq.Array(cSettings.WordleGuesses),
		cSettings.WordleStreaks,
		cSettings.GamePoints,
	)
	if err != nil {
		return err
	}
	return nil
}
