package message

import (
	"database/sql"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"kano/internals/utils/messageutils"
	"log"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

var (
	ACCEPT string = "✅"
	DENY   string = "❌"
)

func voStartDownload(instance *messageutils.Message, selectedMsg *waE2E.Message) {
	var viewOnceMsg *waE2E.Message

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v\n", r)
		}
	}()

	if selectedMsg.ViewOnceMessage != nil {
		viewOnceMsg = selectedMsg.ViewOnceMessage.Message
	} else if selectedMsg.ViewOnceMessageV2 != nil {
		viewOnceMsg = selectedMsg.ViewOnceMessageV2.Message
	} else if selectedMsg.ViewOnceMessageV2Extension != nil {
		viewOnceMsg = selectedMsg.ViewOnceMessageV2Extension.Message
	}

	if viewOnceMsg == nil {
		viewOnceMsg = selectedMsg
	}

	var downloadableMsg whatsmeow.DownloadableMessage
	var caption, mimeType string

	if vid := viewOnceMsg.GetVideoMessage(); vid != nil && vid.ViewOnce != nil && *vid.ViewOnce {
		downloadableMsg = vid
		caption = vid.GetCaption()
		mimeType = vid.GetMimetype()
	} else if img := viewOnceMsg.GetImageMessage(); img != nil && img.ViewOnce != nil && *img.ViewOnce {
		downloadableMsg = img
		caption = img.GetCaption()
		mimeType = img.GetMimetype()
	} else if aud := viewOnceMsg.GetAudioMessage(); aud != nil && aud.ViewOnce != nil && *aud.ViewOnce {
		downloadableMsg = aud
		mimeType = aud.GetMimetype()
	}

	if downloadableMsg == nil {
		instance.Reply("Itu bukan vo", true)
		return
	}

	downloaded_bytes, err := instance.Client.Download(downloadableMsg)
	if err != nil {
		instance.Reply("Terjadi kesalahan saat mengunduh media!", true)
		return
	}

	mediaType := whatsmeow.GetMediaType(downloadableMsg)

	if mediaType == whatsmeow.MediaImage {
		instance.ReplyImage(downloaded_bytes, mimeType, caption, true)
		return
	} else if mediaType == whatsmeow.MediaVideo {
		instance.ReplyVideo(downloaded_bytes, mimeType, caption, true)
		return
	} else if mediaType == whatsmeow.MediaAudio {
		instance.ReplyAudio(downloaded_bytes, mimeType, true)
		return
	}

	instance.Reply("Unexpected reachable code", true)
}

func (ctx MessageContext) voReactHandler() (stop bool) {
	stop = false

	if !ctx.Instance.IsReaction() {
		fmt.Println("[vo] Message is not a reaction")
		return
	}

	fmt.Println("[vo] Woi")

	stop = true
	react := ""
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		fmt.Println("[vo] Failed to get account data with error: ", err)
		return
	}

	if r := ctx.Instance.Reaction(); r != nil {
		react = *r
	}

	reactedMsg, err := ctx.Instance.ResolveReactedMessage()
	if err != nil {
		fmt.Println("[vo] Failed to get reacted message with error: ", err)
		return
	}
	if reactedMsg == nil {
		fmt.Println("[vo] Reacted msg is nil")
		return
	}

	var reqID int64
	var requestedMsgID string
	var accepted sql.NullBool
	err = db.QueryRow("SELECT id, requested_msg_id, accepted FROM request_view_once_entity WHERE jid = $1 AND account_id = $2 AND confirm_msg_id = $3", ctx.Instance.ChatJID().String(), acc.ID, reactedMsg.ID()).Scan(&reqID, &requestedMsgID, &accepted)
	if err != nil {
		fmt.Println("[vo] Failed to get request data with error: ", err)
		return
	}

	if accepted.Valid {
		fmt.Println("[vo] Request already completed with accepted value = ", accepted.Bool)
		return
	}

	requestedMsg, err := ctx.Instance.GetMessage(requestedMsgID, ctx.Instance.ChatJID())
	if err != nil {
		fmt.Println("[vo] Failed to get requested message from database")
		return
	}
	if requestedMsg == nil {
		fmt.Println("[vo] Requested message is not found")
		return
	}

	if requestedMsg.SenderJID().String() != ctx.Instance.SenderJID().String() {
		fmt.Println("[vo] Reactor is not a sender of requested message")
		return
	}

	if react == DENY {
		fmt.Println("[vo] Denied")
		reactedMsg.EditText("Permintaan ditolak!")

		// Set request status into not accepted
		_, err := db.Exec("UPDATE request_view_once AS rvo SET accepted = false WHERE rvo.id = $1", reqID)
		if err != nil {
			fmt.Println("[vo] Failed to set accepted into false where request id is ", reqID)
			return
		}

		return
	} else if react == ACCEPT {
		fmt.Println("[vo] Accepted")
		reactedMsg.EditText("Permintaan diterima! Mohon ditunggu")

		// Set request status into accepted
		_, err := db.Exec("UPDATE request_view_once AS rvo SET accepted = true WHERE rvo.id = $1", reqID)
		if err != nil {
			fmt.Println("[vo] Failed to set accepted into true where request id is", reqID, "with error", err)
			return
		}

		voStartDownload(reactedMsg, requestedMsg.Event.RawMessage)
		return
	}

	fmt.Println("[vo] Invalid emoji")

	return
}

func (ctx MessageContext) VoHandler() {
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		return
	}

	stop := ctx.voReactHandler()
	if stop {
		return
	}

	repliedMsg, err := ctx.Instance.ResolveReplyMessage(true)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengambil pesan reply", true)
		return
	}

	if repliedMsg == nil {
		ctx.Instance.Reply("Mana reply-nya", true)
		return
	}

	part := ctx.Instance.Event.RawMessage.ExtendedTextMessage.ContextInfo.Participant
	if part == nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengolah data pengirim", true)
		return
	}

	if ctx.Instance.SenderJID().String() == *part {
		voStartDownload(ctx.Instance, repliedMsg.Event.RawMessage)
		return
	}

	var reqID int64
	var requestedMsgID string
	var accepted sql.NullBool
	var reqFound bool = true
	err = db.QueryRow("SELECT id, requested_msg_id, accepted FROM request_view_once_entity WHERE jid = $1 AND account_id = $2 AND requested_msg_id = $3", ctx.Instance.ChatJID().String(), acc.ID, repliedMsg.ID()).Scan(&reqID, &requestedMsgID, &accepted)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			ctx.Instance.Reply(fmt.Sprintf("Something went wrong: %s", err), true)
			return
		} else {
			reqFound = false
		}
	}

	if reqFound {
		var text string = "Permintaan untuk pesan tersebut sudah ada"
		if accepted.Valid {
			if accepted.Bool {
				text += " dan yang diminta sudah menerimanya"
			} else {
				text += " dan yang diminta sudah menolaknya"
			}
		}

		ctx.Instance.Reply(text, true)
		return
	}

	var entID *int64
	err = db.QueryRow("SELECT id FROM entity WHERE account_id = $1 AND jid = $2", acc.ID, ctx.Instance.ChatJID().String()).Scan(&entID)
	if err != nil {
		grp, err := ctx.Instance.SaveGroup(false)
		if err != nil {
			return
		}

		if grp.Group != nil {
			entID = &grp.Group.AccountID
		} else if grp.Contact != nil {
			entID = &grp.Contact.AccountID
		}
	}

	if entID == nil {
		return
	}

	confirm, err := ctx.Instance.Reply("React pesan ini dengan ✅ untuk menyetujui atau ❌ untuk menolak", true)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("INSERT INTO request_view_once VALUES(DEFAULT, DEFAULT, $1, $2, $3)", &entID, confirm.ID, repliedMsg.ID())
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat menyimpan permintaan ke basis data", true)
		return
	}
}
