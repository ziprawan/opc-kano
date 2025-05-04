package message

import (
	"context"
	"database/sql"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
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
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Beri pesannya dong kak (belum support media lagi)", true)
		return
	}

	confessMsg := "Ada konfes dari seseorang nih!\n" + strings.Replace(ctx.Parser.Text, ctx.Parser.GetCommand().FullCommand, "", 1)
	ctx.Instance.Client.SendMessage(context.Background(), jid, &waE2E.Message{
		Conversation: &confessMsg,
	})
}

func ConfessTargetHandler(ctx *MessageContext) {
	if ctx.Instance.Event.Info.Chat.Server != types.DefaultUserServer {
		return
	}

	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Kirimkan ID grup yang ingin dijadikan target confess", true)
		return
	}

	id, err := strconv.ParseInt(args[0].Content, 10, 0)
	if err != nil {
		ctx.Instance.Reply("ID grup bukan angka atau tidak valid", true)
		return
	}

	db := database.GetDB()
	var name string
	err = db.QueryRow("SELECT g.name FROM \"group\" g WHERE g.id = $1", id).Scan(&name)
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

	db := database.GetDB()
	// conf := projectconfig.GetConfig()
	acc, err := account.GetData()
	if err != nil {
		fmt.Println(err)
		ctx.Instance.Reply("Internal server error [-1]", true)
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

	var jid []GroupIDJID
	rows, err := db.Query("SELECT g.id, g.jid, g.name FROM participant p INNER JOIN contact c ON c.id = p.contact_id INNER JOIN \"group\" g ON g.id = p.group_id AND g.is_incognito != true WHERE g.account_id = $1 AND c.jid = $2 ORDER BY p.group_id ASC", acc.ID, ctx.Instance.SenderJID().String())
	if err != nil {
		fmt.Println(err)
		ctx.Instance.Reply("Internal server error [1]", true)
		return
	}
	for rows.Next() {
		var id int64
		var j, name string
		if err := rows.Scan(&id, &j, &name); err != nil {
			fmt.Println(err)
			ctx.Instance.Reply("Internal server error [2]", true)
			return
		}
		parsed, err := types.ParseJID(j)
		if err != nil {
			fmt.Println(err)
			ctx.Instance.Reply("Internal server error [3]", true)
			return
		}
		jid = append(jid, GroupIDJID{
			ID:   id,
			JID:  parsed,
			Name: name,
		})
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
