package messageutils

import (
	"context"
	"database/sql"
	"fmt"
	"nopi/internals/database"
	"nopi/internals/utils/account"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

const INSERT_PARTICIPANT_QUERY string = `
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
		ON CONFLICT ("jid", "account_id") DO UPDATE
		SET
			"jid" = EXCLUDED."jid"
		RETURNING
			"id",
			"jid",
			"account_id"
	),
	"inserted_contact" AS (
		INSERT INTO
			"contact" ("entity_id", "jid", "account_id")
		SELECT
			*
		FROM
			"inserted_entity"
		ON CONFLICT ("entity_id") DO UPDATE
		SET
			"jid" = EXCLUDED."jid"
		RETURNING
			"id"
	)
INSERT INTO
	"participant" ("group_id", "contact_id", "role")
SELECT
	$3,
	"ic"."id",
	$4::"participant_role"
FROM
	"inserted_contact" AS "ic"
ON CONFLICT ("group_id", "contact_id") DO UPDATE
SET
	"role" = EXCLUDED."role"
`
const INSERT_CONTACT_QUERY string = `
WITH
	"inserted_entity" AS (
		INSERT INTO
			"entity"
		VALUES
			(DEFAULT, 'CONTACT'::"chat_type", $1, $2)
		ON CONFLICT ("jid", "account_id") DO UPDATE
		SET
			"jid" = EXCLUDED."jid"
		RETURNING
			"id",
			"jid",
			"account_id"
	)
INSERT INTO
	"contact" ("entity_id", "jid", "account_id", "push_name")
SELECT
	*,
	$3
FROM
	"inserted_entity"
ON CONFLICT ("entity_id") DO UPDATE
SET
	"jid" = EXCLUDED."jid"
RETURNING
	id,
	entity_id,
	jid,
	created_at,
	updated_at,
	custom_name,
	push_name,
	account_id`

func saveGroup(group *types.JID, sender *types.JID, ev *events.Message, client *whatsmeow.Client) (*Chat, error) {
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		panic(err)
	}

	server := group.Server
	chat := Chat{}

	if server == types.GroupServer {
		var jid string
		var g Group
		foundStmt := db.QueryRow("SELECT * FROM \"group\" WHERE account_id = $1 AND jid = $2", acc.ID, group.String())
		err := foundStmt.Scan(
			&g.ID,
			&g.AccountID,
			&g.EntityID,
			&jid,
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
			if err == sql.ErrNoRows {
				fmt.Println("Group not found, saving")
				p, e := client.GetGroupInfo(*group)
				if e != nil {
					fmt.Println("Errored at getting group info")
					return nil, e
				}

				fmt.Println(p)

				tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
				if err != nil {
					fmt.Println("Errored at creating transaction")
					return &chat, err
				}

				defer func() {
					if tx != nil {
						tx.Rollback()
					}
				}()

				var entId int64
				fmt.Printf("%+v\n", acc)
				err = tx.QueryRow("INSERT INTO entity VALUES (DEFAULT, 'GROUP'::chat_type, $1, $2) ON CONFLICT (jid, account_id) DO UPDATE SET jid = $1 RETURNING id", group.String(), acc.ID).Scan(&entId)
				if err != nil {
					fmt.Println("Errored at inserting entity")
					return &chat, err
				}

				// GODDAMN
				var grpID int64
				var crtAt, uptAt time.Time
				insStmt, err := tx.Prepare("INSERT INTO \"group\" VALUES (DEFAULT, $1, $2, $3, DEFAULT, DEFAULT, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) ON CONFLICT (entity_id) DO UPDATE SET account_id = EXCLUDED.account_id, entity_id = EXCLUDED.entity_id, jid = EXCLUDED.jid, updated_at = now(), owner_jid = EXCLUDED.owner_jid, name = EXCLUDED.name, name_set_at = EXCLUDED.name_set_at, name_set_by = EXCLUDED.name_set_by, topic = EXCLUDED.topic, topic_id = EXCLUDED.topic_id, topic_set_at = EXCLUDED.topic_set_at, topic_set_by = EXCLUDED.topic_set_by, topic_deleted = EXCLUDED.topic_deleted, is_locked = EXCLUDED.is_locked, is_announce = EXCLUDED.is_announce, announce_version_id = EXCLUDED.announce_version_id, is_ephemeral = EXCLUDED.is_ephemeral, disappearing_timer = EXCLUDED.disappearing_timer, is_incognito = EXCLUDED.is_incognito, is_parent = EXCLUDED.is_parent, default_membership_approval_mode = EXCLUDED.default_membership_approval_mode, linked_parent_jid = EXCLUDED.linked_parent_jid, is_default_subgroup = EXCLUDED.is_default_subgroup, is_join_approval_required = EXCLUDED.is_join_approval_required, group_created = EXCLUDED.group_created, participant_version_id = EXCLUDED.participant_version_id, member_add_mode = EXCLUDED.member_add_mode RETURNING id, created_at, updated_at")
				if err != nil {
					fmt.Println("Errored at preparing insert group statement")
					return &chat, err
				}
				err = insStmt.QueryRow(
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
				).Scan(&grpID, &crtAt, &uptAt)
				if err != nil {
					fmt.Println("Errored at scanning insert result")
					return nil, err
				}

				fmt.Println("Got entity and group id:", entId, grpID)

				for _, part := range p.Participants {
					fmt.Println("Inserting:", part.JID)
					parStmt, err := tx.Prepare(INSERT_PARTICIPANT_QUERY)
					if err != nil {
						fmt.Println("Errored at preparing participant insert")
						return nil, err
					}

					var role string = "MEMBER"

					if part.IsSuperAdmin {
						role = "SUPERADMIN"
					} else if part.IsAdmin {
						role = "ADMIN"
					}

					_, err = parStmt.Exec(part.JID, acc.ID, grpID, role)
					if err != nil {
						fmt.Println("Errored at executing participant insert")
						return nil, err
					}
				}

				if err = tx.Commit(); err != nil {
					fmt.Println("Errored at commit")
					return nil, err
				}

				chat.JID = &p.JID
				chat.Group = &Group{
					ID:                            grpID,
					AccountID:                     acc.ID,
					EntityID:                      entId,
					CreatedAt:                     crtAt,
					UpdatedAt:                     uptAt,
					OwnerJID:                      p.OwnerJID.String(),
					Name:                          p.Name,
					NameSetAt:                     p.NameSetAt,
					NameSetBy:                     p.NameSetBy.String(),
					Topic:                         p.Topic,
					TopicID:                       p.TopicID,
					TopicSetAt:                    p.TopicSetAt,
					TopicSetBy:                    p.TopicSetBy.String(),
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

				return &chat, nil
			} else {
				fmt.Println("Errored at scanning get group result")
				return nil, err
			}
		}

		parsed, err := types.ParseJID(jid)
		if err != nil {
			fmt.Println("Errored at parsing jid")
			return nil, err
		}

		chat.Group = &g
		chat.JID = &parsed

		return &chat, nil
	} else if server == types.DefaultUserServer {
		return saveContact(sender, ev)
	} else {
		return nil, ErrUnsupportedJidServer
	}
}

func saveContact(sender *types.JID, ev *events.Message) (*Chat, error) {
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		panic(err)
	}

	server := sender.Server
	chat := Chat{}

	if server == types.DefaultUserServer {
		p := ev.Info.PushName
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			fmt.Println("Errored at creating transaction")
			return nil, err
		}

		defer func() {
			if tx != nil {
				tx.Rollback()
			}
		}()

		var c Contact
		err = db.QueryRow(INSERT_CONTACT_QUERY, sender.String(), acc.ID, p).Scan(&c.ID, &c.EntityID, &c.JID, &c.CreatedAt, &c.UpdatedAt, &c.CustomName, &c.PushName, &c.AccountID)
		if err != nil {
			fmt.Println("Errored at executing contact insert")
			return nil, err
		}

		chat.Contact = &c
		chat.JID = sender

		return &chat, nil
	} else {
		return nil, ErrUnsupportedJidServer
	}
}
