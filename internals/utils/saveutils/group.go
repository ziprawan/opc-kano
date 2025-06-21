package saveutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"os"
	"time"

	"github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

var (
	ErrNilArguments = errors.New("given argument is a nil pointer")
)

const (
	INSERT_GROUP_QUERY string = `
	INSERT INTO
		"group"
	VALUES
		(
			DEFAULT,
			$1,
			$2,
			$3,
			DEFAULT,
			DEFAULT,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,
			$20,
			$21,
			$22,
			$23,
			$24,
			$25,
			$26
		)
	ON CONFLICT (ENTITY_ID) DO UPDATE
	SET
		ACCOUNT_ID = EXCLUDED.ACCOUNT_ID,
		ENTITY_ID = EXCLUDED.ENTITY_ID,
		JID = EXCLUDED.JID,
		UPDATED_AT = NOW(),
		OWNER_JID = EXCLUDED.OWNER_JID,
		NAME = EXCLUDED.NAME,
		NAME_SET_AT = EXCLUDED.NAME_SET_AT,
		NAME_SET_BY = EXCLUDED.NAME_SET_BY,
		TOPIC = EXCLUDED.TOPIC,
		TOPIC_ID = EXCLUDED.TOPIC_ID,
		TOPIC_SET_AT = EXCLUDED.TOPIC_SET_AT,
		TOPIC_SET_BY = EXCLUDED.TOPIC_SET_BY,
		TOPIC_DELETED = EXCLUDED.TOPIC_DELETED,
		IS_LOCKED = EXCLUDED.IS_LOCKED,
		IS_ANNOUNCE = EXCLUDED.IS_ANNOUNCE,
		ANNOUNCE_VERSION_ID = EXCLUDED.ANNOUNCE_VERSION_ID,
		IS_EPHEMERAL = EXCLUDED.IS_EPHEMERAL,
		DISAPPEARING_TIMER = EXCLUDED.DISAPPEARING_TIMER,
		IS_INCOGNITO = EXCLUDED.IS_INCOGNITO,
		IS_PARENT = EXCLUDED.IS_PARENT,
		DEFAULT_MEMBERSHIP_APPROVAL_MODE = EXCLUDED.DEFAULT_MEMBERSHIP_APPROVAL_MODE,
		LINKED_PARENT_JID = EXCLUDED.LINKED_PARENT_JID,
		IS_DEFAULT_SUBGROUP = EXCLUDED.IS_DEFAULT_SUBGROUP,
		IS_JOIN_APPROVAL_REQUIRED = EXCLUDED.IS_JOIN_APPROVAL_REQUIRED,
		GROUP_CREATED = EXCLUDED.GROUP_CREATED,
		PARTICIPANT_VERSION_ID = EXCLUDED.PARTICIPANT_VERSION_ID,
		MEMBER_ADD_MODE = EXCLUDED.MEMBER_ADD_MODE
	RETURNING
		ID,
		CREATED_AT,
		UPDATED_AT`

	UPDATE_GROUP_QUERY string = `
UPDATE "group"
SET
	ACCOUNT_ID = $1,
	ENTITY_ID = $2,
	JID = $3,
	UPDATED_AT = NOW(),
	OWNER_JID = $4,
	NAME = $5,
	NAME_SET_AT = $6,
	NAME_SET_BY = $7,
	TOPIC = $8,
	TOPIC_ID = $9,
	TOPIC_SET_AT = $10,
	TOPIC_SET_BY = $11,
	TOPIC_DELETED = $12,
	IS_LOCKED = $13,
	IS_ANNOUNCE = $14,
	ANNOUNCE_VERSION_ID = $15,
	IS_EPHEMERAL = $16,
	DISAPPEARING_TIMER = $17,
	IS_INCOGNITO = $18,
	IS_PARENT = $19,
	DEFAULT_MEMBERSHIP_APPROVAL_MODE = $20,
	LINKED_PARENT_JID = $21,
	IS_DEFAULT_SUBGROUP = $22,
	IS_JOIN_APPROVAL_REQUIRED = $23,
	GROUP_CREATED = $24,
	PARTICIPANT_VERSION_ID = $25,
	MEMBER_ADD_MODE = $26
WHERE
	ID = $27`

	INSERT_CONTACT_WITH_ENTITY_QUERY string = `
WITH
	"inserted_entity" AS (
		INSERT INTO
			"entity"
		VALUES
			(
				DEFAULT,
				'CONTACT'::"chat_type",
				$1,
				$2
			)
		ON CONFLICT (jid, account_id) DO UPDATE
		SET
			"jid" = EXCLUDED."jid"
		RETURNING
			"id",
			"jid",
			"account_id"
	)
INSERT INTO
	"contact" ("entity_id", "jid", "account_id")
SELECT
	*
FROM
	"inserted_entity"
ON CONFLICT (entity_id) DO UPDATE
SET
	"jid" = EXCLUDED."jid"
RETURNING
	"id"`

	UPDATE_GROUP_SETTINGS string = `
INSERT INTO
	"group_settings"
VALUES (
	$1,
	$2,
	$3
)
ON CONFLICT
	("id")
DO UPDATE SET
	chosen_shipping = $2,
	last_shipping_time = $3`
)

func scanInt64FromRow(row *sql.Row) (*int64, error) {
	var num int64
	err := row.Scan(&num)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &num, nil
}

func getGroupSettings(groupID int64) (*GroupSettings, error) {
	db := database.GetDB()
	var groupSettings GroupSettings
	var rawr sql.NullString
	err := db.QueryRow("SELECT to_json(chosen_shipping), last_shipping_time FROM group_settings WHERE id = $1", groupID).Scan(&rawr, &groupSettings.LastShippingTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if rawr.Valid {
		err := json.Unmarshal([]byte(rawr.String), &groupSettings.ChosenShipping)
		if err != nil {
			fmt.Println("Gagal unmarshal chosen", err)
		}
	}

	return &groupSettings, nil
}

// Don't use this for non @g.us JID
func getGroup(jid *types.JID) (group *Group, err error) {
	if jid == nil {
		return nil, ErrNilArguments
	}
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		return
	}
	var g Group
	var db_jid string
	err = db.QueryRow("SELECT * FROM \"group\" WHERE account_id = $1 AND jid = $2", acc.ID, jid.String()).
		Scan(
			&g.ID,
			&g.AccountID,
			&g.EntityID,
			&db_jid,
			&g.CreatedAt,
			&g.UpdatedAt,
			&g.OwnerJID,
			&g.Name,
			&g.NameSetAt,
			&g.NameSetBy,
			&g.Topic,
			&g.TopicID,
			&g.TopicSetAt,
			&g.TopicSetBy,
			&g.TopicDeleted,
			&g.IsLocked,
			&g.IsAnnounce,
			&g.AnnonceVersionID,
			&g.IsEphemeral,
			&g.DisappearingTimer,
			&g.IsIncognito,
			&g.IsParent,
			&g.DefaultMembershipApprovalMode,
			&g.LinkedParentJID,
			&g.IsDefaultSubgroup,
			&g.IsJoinApprovalRequired,
			&g.GroupCreated,
			&g.ParticipantVersionID,
			&g.MemberAddMode,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	parsedJID, err := types.ParseJID(db_jid)
	if err != nil {
		return nil, err
	}

	groupSettings, err := getGroupSettings(g.ID)
	if err != nil {
		fmt.Println("Failed to get group settings:", err)
	}
	g.Settings = groupSettings
	g.JID = &parsedJID
	group = &g

	return
}

// Don't use this outside this file
func addOrUpdateGroupParticipant(tx *sql.Tx, group *Group, jid *types.JID, role ParticipantRole) (*Participant, error) {
	if tx == nil || jid == nil || group == nil {
		return nil, ErrNilArguments
	}

	var contactID int64
	scannedContactID, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM contact WHERE account_id = $1 AND jid = $2", group.AccountID, jid.String()))
	if err != nil {
		return nil, fmt.Errorf("error scanning contact id: %s", err)
	}

	if scannedContactID == nil {
		scannedContactID, err = scanInt64FromRow(tx.QueryRow(INSERT_CONTACT_WITH_ENTITY_QUERY, jid.String(), group.AccountID))
		if err != nil {
			return nil, fmt.Errorf("error scanning insert contact id %s", err)
		}
		if scannedContactID == nil {
			return nil, fmt.Errorf("failed to insert contact")
		} else {
			contactID = *scannedContactID
		}
	} else {
		contactID = *scannedContactID
	}

	// I assume contactID is always valid
	var participant Participant = Participant{
		GroupID:   group.ID,
		ContactID: contactID,
	}
	// Check if the participant already exists
	err = tx.QueryRow("SELECT id, role FROM participant WHERE group_id = $1 AND contact_id = $2", group.ID, contactID).Scan(&participant.ID, &participant.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Participant doesn't exist, insert it
			err = tx.QueryRow("INSERT INTO participant (group_id, contact_id, role) VALUES ($1, $2, $3) RETURNING id, role", group.ID, contactID, role).Scan(&participant.ID, &participant.Role)
			if err != nil {
				return nil, fmt.Errorf("error when scanning insert participant %s", err)
			}
		} else {
			return nil, fmt.Errorf("error scanning select participant %s", err)
		}
	}

	// Update the participant role if it is different
	if participant.Role != role {
		// This is a negation of "participant.Role == ParticipantRoleManager && role == ParticipantRoleMember"
		// This means that if the participant is a manager and the new role is member, we don't update it
		if participant.Role != ParticipantRoleManager || role != ParticipantRoleMember {
			// Update the role
			_, err = tx.Exec("UPDATE participant SET role = $1 WHERE id = $2", role, participant.ID)
			if err != nil {
				return nil, fmt.Errorf("error executing update participant %s", err)
			}
			participant.Role = role
		}
	}

	return &participant, nil
}

func (grp *Group) AddOrUpdateParticipant(tx *sql.Tx, jid *types.JID, role ParticipantRole) (*Participant, error) {
	return addOrUpdateGroupParticipant(tx, grp, jid, role)
}

func (grp *Group) SaveGroupSettings() error {
	db := database.GetDB()
	if grp.Settings == nil {
		grp.Settings = &GroupSettings{}
		_, err := db.Exec("INSERT INTO group_settings (id) VALUES id = $1", grp.ID)
		if err != nil {
			return err
		}
		return nil
	}
	groupSettings := *grp.Settings

	_, err := db.Exec(UPDATE_GROUP_SETTINGS, grp.ID, pq.Array(groupSettings.ChosenShipping), groupSettings.LastShippingTime)
	if err != nil {
		return err
	}

	return nil
}

func SaveOrUpdateGroup(client *whatsmeow.Client, jid *types.JID) (*Group, error) {
	fmt.Println("save or update group called")
	if client == nil || jid == nil {
		return nil, ErrNilArguments
	}

	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get account data", err)
		return nil, err
	}

	p, err := client.GetGroupInfo(*jid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get group info from server", err)
		return nil, err
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create transaction", err)
		return nil, err
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Get the entity id first, if exists, use it. If not, create it
	var entId int64
	scannedEntId, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM entity e WHERE account_id = $1 AND jid = $2", acc.ID, jid.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to scan SELECT ENTITY result", err)
		return nil, err
	}
	if scannedEntId == nil {
		err = tx.QueryRow("INSERT INTO entity VALUES (DEFAULT, 'GROUP'::chat_type, $2, $1) ON CONFLICT (jid, account_id) DO UPDATE SET jid = $2 RETURNING id", acc.ID, jid.String()).Scan(&entId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Fprintln(os.Stderr, "failed to scan INSERT ENTITY result", err)
				return nil, err
			}
		}
	} else {
		entId = *scannedEntId
	}

	// Check again for the group data first, if exist, update it
	var grpID int64
	scannedGrpID, err := scanInt64FromRow(tx.QueryRow("SELECT id FROM \"group\" WHERE account_id = $1 AND jid = $2", acc.ID, jid.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to scan SELECT \"GROUP\" result", err)
		return nil, err
	}
	var createdAt, updatedAt time.Time
	if scannedGrpID == nil {
		// Oh, it doesn't exists! Let's insert it!
		insertStatement, err := tx.Prepare(INSERT_GROUP_QUERY)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to prepare INSERT \"GROUP\" statement", err)
			return nil, err
		}
		err = insertStatement.QueryRow(
			&acc.ID,
			&entId,
			&p.JID,
			&p.OwnerJID,
			&p.Name,
			&p.NameSetAt,
			&p.NameSetBy,
			&p.Topic,
			&p.TopicID,
			&p.TopicSetAt,
			&p.TopicSetBy,
			&p.TopicDeleted,
			&p.IsLocked,
			&p.IsAnnounce,
			&p.AnnounceVersionID,
			&p.IsEphemeral,
			&p.DisappearingTimer,
			&p.IsIncognito,
			&p.IsParent,
			&p.DefaultMembershipApprovalMode,
			&p.LinkedParentJID,
			&p.IsDefaultSubGroup,
			&p.IsJoinApprovalRequired,
			&p.GroupCreated,
			&p.ParticipantVersionID,
			&p.MemberAddMode,
		).Scan(&grpID, &createdAt, &updatedAt)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to scan INSERT \"GROUP\" result", err)
			return nil, err
		}
	} else {
		// Eh, it actually exists? That's weird, I think we will just update it
		updateStatement, err := tx.Prepare(UPDATE_GROUP_QUERY)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to prepare UPDATE \"GROUP\" statement", err)
			return nil, err
		}

		_, err = updateStatement.Exec(
			acc.ID,
			entId,
			p.JID,
			p.OwnerJID,
			p.Name,
			p.NameSetAt,
			p.NameSetBy,
			p.Topic,
			p.TopicID,
			p.TopicSetAt,
			p.TopicSetBy,
			p.TopicDeleted,
			p.IsLocked,
			p.IsAnnounce,
			p.AnnounceVersionID,
			p.IsEphemeral,
			p.DisappearingTimer,
			p.IsIncognito,
			p.IsParent,
			p.DefaultMembershipApprovalMode,
			p.LinkedParentJID,
			p.IsDefaultSubGroup,
			p.IsJoinApprovalRequired,
			p.GroupCreated,
			p.ParticipantVersionID,
			p.MemberAddMode,
			grpID,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to execute UPDATE \"GROUP\"", err)
			return nil, err
		}
	}

	group := Group{
		ID:                            grpID,
		AccountID:                     acc.ID,
		EntityID:                      entId,
		CreatedAt:                     createdAt,
		UpdatedAt:                     updatedAt,
		OwnerJID:                      p.OwnerJID.String(),
		Name:                          p.Name,
		NameSetAt:                     p.NameSetAt,
		NameSetBy:                     p.NameSetBy.String(),
		Topic:                         sql.NullString{String: p.Topic, Valid: len(p.Topic) != 0},
		TopicID:                       sql.NullString{String: p.TopicID, Valid: len(p.TopicID) != 0},
		TopicSetAt:                    sql.NullTime{Time: p.TopicSetAt, Valid: true},
		TopicSetBy:                    sql.NullString{String: p.TopicSetBy.String(), Valid: len(p.TopicSetBy.String()) != 0},
		TopicDeleted:                  p.TopicDeleted,
		IsLocked:                      p.IsLocked,
		IsAnnounce:                    p.IsAnnounce,
		AnnonceVersionID:              p.AnnounceVersionID,
		IsEphemeral:                   p.IsEphemeral,
		DisappearingTimer:             int32(p.DisappearingTimer),
		IsIncognito:                   p.IsIncognito,
		IsParent:                      p.IsParent,
		DefaultMembershipApprovalMode: p.DefaultMembershipApprovalMode,
		LinkedParentJID: sql.NullString{
			String: p.LinkedParentJID.String(),
			Valid:  len(p.LinkedParentJID.String()) != 0,
		},
		IsDefaultSubgroup:      p.IsDefaultSubGroup,
		IsJoinApprovalRequired: p.IsJoinApprovalRequired,
		GroupCreated:           p.GroupCreated,
		ParticipantVersionID:   p.ParticipantVersionID,
		MemberAddMode:          string(p.MemberAddMode),
	}

	for _, participant := range p.Participants {
		role := ParticipantRoleMember
		if participant.IsSuperAdmin {
			role = ParticipantRoleSuperAdmin
		}
		if participant.IsAdmin {
			role = ParticipantRoleAdmin
		}
		_, err = addOrUpdateGroupParticipant(tx, &group, &participant.JID, role)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to add or update group participant", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("failed to commit group transaction:", err)
		return nil, err
	}

	return &group, nil
}

func GetOrSaveGroup(client *whatsmeow.Client, jid *types.JID) (*Group, error) {
	if client == nil || jid == nil {
		return nil, ErrNilArguments
	}
	foundGroup, err := getGroup(jid)
	if err != nil {
		return nil, err
	}

	// foundGroup is a pointer, check for nil pointer first
	if foundGroup == nil {
		return SaveOrUpdateGroup(client, jid)
	}

	// Just return it
	return foundGroup, nil
}
