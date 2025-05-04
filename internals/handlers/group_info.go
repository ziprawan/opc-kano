package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/saveutils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type JIDRole struct {
	JID  *types.JID
	Role saveutils.ParticipantRole
}

func GroupInfoHandler(client *whatsmeow.Client, event *events.GroupInfo) {
	if event.JID.Server != types.GroupServer {
		fmt.Println("Unsupported jid server:", event.JID.Server)
		return
	}
	grp, err := saveutils.GetOrSaveGroup(client, &event.JID)
	if err != nil {
		fmt.Println("Something went wrong when resolving group info", err)
		return
	}

	var updates []JIDRole
	if len(event.Join) > 0 {
		msg := "Halo, "
		var joins []string
		for _, jid := range event.Join {
			msg += fmt.Sprintf("@%s ", jid.User)
			joins = append(joins, jid.String())
			updates = append(updates, JIDRole{
				JID:  &jid,
				Role: "MEMBER",
			})
		}
		msg += "👋"
		client.SendMessage(context.Background(), event.JID, &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: &msg,
				ContextInfo: &waE2E.ContextInfo{
					MentionedJID: joins,
				},
			},
		})
	}
	if len(event.Leave) > 0 {
		msg := "Bye, "
		var leaves []string
		for _, jid := range event.Leave {
			msg += fmt.Sprintf("@%s ", jid.User)
			leaves = append(leaves, jid.String())
			updates = append(updates, JIDRole{
				JID:  &jid,
				Role: "LEFT",
			})
		}
		client.SendMessage(context.Background(), event.JID, &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: &msg,
				ContextInfo: &waE2E.ContextInfo{
					MentionedJID: leaves,
				},
			},
		})
	}
	if len(event.Promote) > 0 {
		for _, jid := range event.Promote {
			updates = append(updates, JIDRole{
				JID:  &jid,
				Role: "ADMIN",
			})
		}
	}
	if len(event.Demote) > 0 {
		for _, jid := range event.Demote {
			updates = append(updates, JIDRole{
				JID:  &jid,
				Role: "MEMBER",
			})
		}
	}

	if len(updates) > 0 {
		db := database.GetDB()
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			fmt.Println("Error creating a transaction", err)
			return
		}
		defer func() {
			if tx != nil {
				tx.Rollback()
			}
		}()
		for _, up := range updates {
			p, err := grp.AddOrUpdateParticipant(tx, up.JID, up.Role)
			if err != nil {
				fmt.Println("Error adding or updating participant", err)
			} else if p != nil {
				fmt.Printf("Participant%+v\n", p)
			}
		}
		tx.Commit()
	}
}
