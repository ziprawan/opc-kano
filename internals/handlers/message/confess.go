package message

import (
	"context"
	"database/sql"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"kano/internals/utils/messageutils"
	"strconv"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type GroupIDJID struct {
	ID   int64     `json:"id"`
	JID  types.JID `json:"jid"`
	Name string    `json:"name"`
}

func getUserConfessTarget(jid types.JID) (*types.JID, error) {
	acc, err := account.GetData()
	if err != nil {
		return nil, err
	}

	// Check if the user is in the database
	db := database.GetDB()
	var target sql.NullString
	err = db.QueryRow("SELECT g.jid FROM contact c LEFT JOIN contact_settings cs ON cs.id = c.id LEFT JOIN \"group\" g ON g.id = cs.confess_target_id WHERE c.account_id = $1 AND c.jid = $2", acc.ID, jid.String()).Scan(&target)
	if err != nil {
		return nil, err
	}
	if target.Valid {
		// Parse the target JID
		targetJID, err := types.ParseJID(target.String)
		if err != nil {
			return nil, err
		}
		return &targetJID, nil
	} else {
		return nil, nil
	}
}

func saveUserConfessTarget(jid types.JID, targetID int64) error {
	acc, err := account.GetData()
	if err != nil {
		return err
	}
	// Check if the user is in the database
	db := database.GetDB()
	var contactId int64
	err = db.QueryRow("SELECT id FROM contact WHERE account_id = $1 AND jid = $2", acc.ID, jid.String()).Scan(&contactId)
	if err != nil {
		return err
	}
	// Check if the target is in the database
	var targetJID string
	err = db.QueryRow("SELECT jid FROM \"group\" WHERE id = $1", targetID).Scan(&targetJID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("target group not found")
		} else {
			return err
		}
	}
	// Insert or update the confess target
	_, err = db.Exec("INSERT INTO contact_settings (id, confess_target_id) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET confess_target_id = $2", contactId, targetID)
	if err != nil {
		return err
	}
	return nil
}

func sendConfessMessage(ctx *MessageContext, jid types.JID) {
	var mediaMessage *waE2E.Message
	var hasReply, needReply bool

	msg := ctx.Instance.Event.RawMessage
	if ext := msg.ExtendedTextMessage; ext != nil {
		if ctxInfo := ext.ContextInfo; ctxInfo != nil {
			if replied := ctxInfo.QuotedMessage; replied != nil {
				msg = ctxInfo.QuotedMessage
				hasReply = true
			}
		}
	}

	args := ctx.Parser.GetArgs()
	instance := messageutils.Message{
		Event: &events.Message{
			RawMessage: msg,
		},
	}
	confessCaption := "Ada konfes dari seseorang nih!\n"
	if cap := instance.Caption(); hasReply && len(cap) != 0 {
		confessCaption += cap
	} else {
		if len(args) == 0 {
			ctx.Instance.Reply("Beri pesannya dong kak (dukungan media sudah bisa namun masih tahap uji coba)", true)
			return
		} else {
			confessCaption += ctx.Parser.Text[args[0].Start:]
		}
	}

	if vid := msg.GetVideoMessage(); vid != nil {
		mediaMessage = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           vid.URL,
				DirectPath:    vid.DirectPath,
				MediaKey:      vid.MediaKey,
				FileEncSHA256: vid.FileEncSHA256,
				FileSHA256:    vid.FileSHA256,
				FileLength:    vid.FileLength,
				Caption:       &confessCaption,
				Mimetype:      vid.Mimetype,
			},
		}
	} else if img := msg.GetImageMessage(); img != nil {
		mediaMessage = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           img.URL,
				DirectPath:    img.DirectPath,
				MediaKey:      img.MediaKey,
				FileEncSHA256: img.FileEncSHA256,
				FileSHA256:    img.FileSHA256,
				FileLength:    img.FileLength,
				Caption:       &confessCaption,
				Mimetype:      img.Mimetype,
			},
		}
	} else if aud := msg.GetAudioMessage(); aud != nil {
		needReply = true
		mediaMessage = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           aud.URL,
				DirectPath:    aud.DirectPath,
				MediaKey:      aud.MediaKey,
				FileEncSHA256: aud.FileEncSHA256,
				FileSHA256:    aud.FileSHA256,
				FileLength:    aud.FileLength,
				Mimetype:      aud.Mimetype,
			},
		}
	} else if stk := msg.GetStickerMessage(); stk != nil {
		needReply = true
		mediaMessage = &waE2E.Message{
			StickerMessage: &waE2E.StickerMessage{
				URL:           stk.URL,
				DirectPath:    stk.DirectPath,
				MediaKey:      stk.MediaKey,
				FileEncSHA256: stk.FileEncSHA256,
				FileSHA256:    stk.FileSHA256,
				FileLength:    stk.FileLength,
				Mimetype:      stk.Mimetype,
			},
		}
	}

	if mediaMessage != nil {
		resp, err := ctx.Instance.Client.SendMessage(context.Background(), jid, mediaMessage)
		if err != nil {
			fmt.Println("confess: Failed to send media message")
			return
		}
		if needReply {
			ctx.Instance.Client.SendMessage(context.Background(), jid, &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: &confessCaption,
					ContextInfo: &waE2E.ContextInfo{
						StanzaID:      proto.String(resp.ID),
						Participant:   proto.String(resp.Sender.String()),
						QuotedMessage: mediaMessage,
					},
				},
			})
		}
	} else {
		ctx.Instance.Client.SendMessage(context.Background(), jid, &waE2E.Message{
			Conversation: &confessCaption,
		})
	}

}

func getUserJoinedGroups(senderJid string) ([]GroupIDJID, error) {
	db := database.GetDB()
	acc, _ := account.GetData()
	rows, err := db.Query("SELECT g.id, g.jid, g.name FROM participant p INNER JOIN contact c ON c.id = p.contact_id INNER JOIN \"group\" g ON g.id = p.group_id AND g.is_incognito != true AND p.role != 'LEFT' WHERE g.account_id = $1 AND c.jid = $2 ORDER BY p.group_id ASC", acc.ID, senderJid)
	if err != nil {
		fmt.Println("confess: getUserJoinedGroups: Failed to build query:", err)
		return nil, err
	}
	defer rows.Close()

	var jid []GroupIDJID
	for rows.Next() {
		var id int64
		var j, name string
		if err := rows.Scan(&id, &j, &name); err != nil {
			fmt.Println("confess: getUserJoinedGroups: Failed to scan result:", err)
			return nil, err
		}
		parsed, err := types.ParseJID(j)
		if err != nil {
			fmt.Println("confess: getUserJoinedGroups: Failed to parse JID", err)
			return nil, err
		}
		jid = append(jid, GroupIDJID{
			ID:   id,
			JID:  parsed,
			Name: name,
		})
	}

	return jid, nil
}

func ConfessTargetHandler(ctx *MessageContext) {
	if ctx.Instance.Event.Info.Chat.Server != types.DefaultUserServer {
		return
	}

	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		jids, err := getUserJoinedGroups(ctx.Instance.SenderJID().String())
		if err != nil {
			ctx.Instance.Reply("Something went wrong", true)
			return
		}

		msg := "Kirimkan ID grup yang ingin dijadikan target confess\nBerikut list grup yang ada _(format: [id]. [nama grup])_ :\n\n"
		for _, j := range jids {
			msg += fmt.Sprintf("%d. %s\n", j.ID, j.Name)
		}

		ctx.Instance.Reply(msg, true)
		return
	}

	id, err := strconv.ParseInt(args[0].Content, 10, 0)
	if err != nil {
		ctx.Instance.Reply("ID grup bukan angka atau tidak valid", true)
		return
	}

	db := database.GetDB()
	var name string
	err = db.QueryRow("SELECT g.name FROM participant p INNER JOIN \"group\" g ON p.group_id = $1 AND p.role != 'LEFT' AND p.contact_id = $2", id, ctx.Instance.Contact.ID).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Instance.Reply("Grup tidak ditemukan", true)
			return
		} else {
			fmt.Println(err)
			ctx.Instance.Reply("Internal server error [5]", true)
			return
		}
	}

	senderJID := ctx.Instance.SenderJID()
	err = saveUserConfessTarget(*senderJID, id)
	if err != nil {
		fmt.Println(err)
		ctx.Instance.Reply("Internal server error [7]", true)
		return
	}

	ctx.Instance.Reply(fmt.Sprintf("Berhasil mengatur target confess ke %s", name), true)
}

func ConfessHandler(ctx *MessageContext) {
	if ctx.Instance.Event.Info.Chat.Server != types.DefaultUserServer {
		return
	}

	// if ctx.Instance.SenderJID().User != conf.OwnerJID.User {
	// 	ctx.Instance.Reply("Under maintenance.", true)
	// 	return
	// }

	target, err := getUserConfessTarget(*ctx.Instance.SenderJID())
	if err != nil {
		fmt.Println(err)
		ctx.Instance.Reply("Internal server error [0]", true)
		return
	}
	if target != nil {
		sendConfessMessage(ctx, *target)
		return
	}

	jid, err := getUserJoinedGroups(ctx.Instance.SenderJID().String())
	if err != nil {
		ctx.Instance.Reply("Internal server error [2]", true)
		return
	}
	if len(jid) == 0 {
		ctx.Instance.Reply("Saya ga pernah lihat kamu di grup manapun, hmm", true)
		return
	}
	if len(jid) > 1 {
		msg := "Saya lihat kamu ada di beberapa grup. Atur target confess dengan menggunakan *.confesstarget [id]* tanpa kurung siku\n\nBerikut list grup yang ada _(format: [id]. [nama grup])_ :\n\n"
		for _, j := range jid {
			msg += fmt.Sprintf("%d. %s\n", j.ID, j.Name)
		}
		ctx.Instance.Reply(msg, true)
		return
	}
	if len(jid) == 1 {
		sendConfessMessage(ctx, jid[0].JID)
		return
	}

	ctx.Instance.Reply("Internal server error [4]", true)
}
