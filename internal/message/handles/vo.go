package handles

import (
	"database/sql"
	"errors"
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/word"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type downloadableMessageWithURL interface {
	whatsmeow.DownloadableMessage
	GetURL() string
}

func Vo(c *messageutil.MessageContext) (err error) {
	db := database.GetInstance()
	msgId, senderJid, repliedMsg := c.GetRepliedMessage()
	if repliedMsg == nil {
		_, err = c.QuoteReply("Please reply to a view-once message.")
		return
	}

	repliedToThemself := c.IsSenderSame(senderJid)
	fmt.Println(repliedToThemself, c.GetSender(), senderJid)

	// Check if the sender is replied to themself
	found := models.VoRequest{ChatJid: c.GetChat(), MessageId: msgId}
	tx := db.Where(&found).First(&found)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			err = tx.Error
			c.QuoteReply("Failed to query to vo_request table: %s", tx.Error)
			return
		}
	} else {
		msg := fmt.Sprintf("View-once message is already requested by @%s", found.RequesterJid.User)
		if found.Accepted.Valid {
			if found.Accepted.Bool {
				msg += " and already accepted by the sender"
			} else {
				msg += " and already denied by the sender"
			}
		}

		ctxInfo := c.BuildReplyContextInfo()
		ctxInfo.MentionedJID = []string{found.RequesterJid.String()}
		_, err = c.SendMessage(&waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        &msg,
				ContextInfo: ctxInfo,
			},
		})
		return
	}

	// Check if it was from ViewOnce*
	if vo := repliedMsg.GetViewOnceMessage().GetMessage(); vo != nil {
		repliedMsg = vo
	} else if vo2 := repliedMsg.GetViewOnceMessageV2().GetMessage(); vo2 != nil {
		repliedMsg = vo2
	} else if vo2ext := repliedMsg.GetViewOnceMessageV2Extension().GetMessage(); vo2ext != nil {
		repliedMsg = vo2ext
	}

	// A message object to send if the requested vo is the sender itself
	selfMsg := &waE2E.Message{}
	var mediaType whatsmeow.MediaType
	var downloadable whatsmeow.DownloadableMessage

	if i := repliedMsg.GetImageMessage(); i != nil && i.GetViewOnce() {
		i.ViewOnce = proto.Bool(false)
		downloadable = i
		selfMsg.ImageMessage = i
		mediaType = whatsmeow.MediaImage
	} else if v := repliedMsg.GetVideoMessage(); v != nil && v.GetViewOnce() {
		v.ViewOnce = proto.Bool(false)
		downloadable = v
		selfMsg.VideoMessage = v
		mediaType = whatsmeow.MediaVideo
	} else if a := repliedMsg.GetAudioMessage(); a != nil && a.GetViewOnce() {
		a.ViewOnce = proto.Bool(false)
		downloadable = a
		selfMsg.AudioMessage = a
		mediaType = whatsmeow.MediaAudio
	}

	if downloadable == nil {
		_, err = c.QuoteReply("Replied message is not a view-once message.")
		return
	}

	if val := c.ValidateDownloadableMessage(downloadable); val != nil {
		_, err = c.QuoteReply("Replied view once has missing fields. Most likely you are using the latest WhatsApp version.\nDebug info: %s", val.Error())
		return
	}

	insert := models.VoRequest{
		ChatJid:   c.GetChat(),
		MessageId: msgId,

		RequesterJid:    c.GetSender(),
		MessageOwnerJid: senderJid,
		// ApprovalMessageId is assigned after sending the approval message

		// Url and DirectPath is assigned below by a condition
		MediaKey:      word.ToBase64(string(downloadable.GetMediaKey())),
		FileSha256:    word.ToBase64(string(downloadable.GetFileSHA256())),
		FileEncSha256: word.ToBase64(string(downloadable.GetFileEncSHA256())),
		MediaType:     string(mediaType),
	}

	if dp := downloadable.GetDirectPath(); len(dp) > 0 {
		insert.DirectPath = sql.NullString{
			String: dp,
			Valid:  true,
		}
	}
	if u, ok := downloadable.(downloadableMessageWithURL); ok && len(u.GetURL()) > 0 {
		insert.Url = sql.NullString{
			String: u.GetURL(),
			Valid:  true,
		}
	}

	if repliedToThemself {
		sent, err := c.SendMessage(selfMsg)
		if err != nil {
			return err
		} else {
			insert.ApprovalMessageId = sent.ID
			insert.Accepted = sql.NullBool{Bool: true, Valid: true}
		}
	} else {
		textMsg := fmt.Sprintf("Waiting for approval from @%s (react with ✅ to approve, ❌ or just leave it alone to reject)", senderJid.User)
		ctxInfo := c.BuildReplyContextInfo()
		ctxInfo.MentionedJID = []string{senderJid.String()}
		sent, err := c.SendMessage(&waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        &textMsg,
				ContextInfo: ctxInfo,
			},
		})

		if err != nil {
			return err
		} else {
			insert.ApprovalMessageId = sent.ID
		}
	}

	tx = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&insert)
	if tx.Error != nil {
		err = tx.Error
		c.Reply(fmt.Sprintf("Failed to save into vo_request table: %s", tx.Error))
		return
	}

	return
}

var VoMan = CommandMan{
	Name:     "vo - unwrap a view-once message",
	Synopsis: []string{"*vo*"},
	Description: []string{
		"Unwraps a replied view-once message. Supports all currently known types of view-once messages, including images, videos, and audio. In certain cases, this command may fail to execute if:" +
			"\n- The command sender is using the WhatsApp desktop client" +
			"\n- The command sender is using the latest version of WhatsApp",
		"If the owner of the view-once message is the same as the command sender, the bot will immediately unwrap the message. If the owner is different (i.e., another user attempts to view the message again), the bot will send a request message in the chat.",
		"The owner of the view-once message can react to the request message with:" +
			"\n- ✅ ✔️ ☑️ to approve" +
			"\n- ❌ ✖️ 🚫 to reject" +
			"\n- Or ignore it to take no action",
		"A view-once message that has already been requested cannot be requested again, even if the owner has not yet approved or rejected the request.",
	},
	SourceFilename: "vo.go",
	SeeAlso:        []SeeAlso{},
}
