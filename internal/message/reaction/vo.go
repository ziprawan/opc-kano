package reaction

import (
	"context"
	"errors"
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/word"
	"slices"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var mediaTypeToMMSType = map[whatsmeow.MediaType]string{
	whatsmeow.MediaImage:         "image",
	whatsmeow.MediaAudio:         "audio",
	whatsmeow.MediaVideo:         "video",
	whatsmeow.MediaDocument:      "document",
	whatsmeow.MediaHistory:       "md-msg-hist",
	whatsmeow.MediaAppState:      "md-app-state",
	whatsmeow.MediaStickerPack:   "sticker-pack",
	whatsmeow.MediaLinkThumbnail: "thumbnail-link",
}

var (
	APPROVAL_EMOJIS    = []string{"✅", "✔️", "☑️"}
	DISAPPROVAL_EMOJIS = []string{"❌", "✖️", "🚫"}
)

func downloadMediaFromVoRequest(c *messageutil.MessageContext, req models.VoRequest) ([]byte, error) {
	mediaType := (whatsmeow.MediaType)(req.MediaType)

	var url string
	var isWebWhatsappNetURL bool
	if req.Url.Valid {
		url = req.Url.String
		isWebWhatsappNetURL = strings.HasPrefix(url, "https://web.whatsapp.net")
	}
	if len(url) > 0 && !isWebWhatsappNetURL {
		return c.Client.GetClient().DangerousInternals().DownloadAndDecrypt(
			context.Background(),
			url,
			[]byte(word.FromBase64(req.MediaKey)),
			mediaType,
			-1,
			[]byte(word.FromBase64(req.FileEncSha256)),
			[]byte(word.FromBase64(req.FileSha256)),
		)
	} else if len(req.DirectPath.String) > 0 {
		return c.Client.GetClient().DownloadMediaWithPath(
			context.Background(),
			req.DirectPath.String,
			[]byte(word.FromBase64(req.FileEncSha256)),
			[]byte(word.FromBase64(req.FileSha256)),
			[]byte(word.FromBase64(req.MediaKey)),
			-1,
			mediaType,
			mediaTypeToMMSType[mediaType],
		)
	} else {
		return nil, whatsmeow.ErrNoURLPresent
	}
}

func VoReactApproval(c *messageutil.MessageContext) error {
	c.Logger.Debugf("Entered VoReactApproval function")

	reactContent := c.GetReaction()
	reactKey := c.GetReactionKey()

	isReactedToMe := reactKey.GetFromMe()
	if !isReactedToMe {
		c.Logger.Infof("Reacted message is not to me")
		return nil
	}

	reactedId := reactKey.GetID()

	isApproved := slices.Contains(APPROVAL_EMOJIS, reactContent)
	isDisapproved := slices.Contains(DISAPPROVAL_EMOJIS, reactContent)

	if !isApproved && !isDisapproved {
		c.Logger.Infof("No supported emojis")
		return nil
	}

	db := database.GetInstance()
	req := models.VoRequest{ChatJid: c.GetChat(), ApprovalMessageId: reactedId}
	tx := db.Where(&req).First(&req)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.Logger.Infof("No record found")
			return nil
		}

		c.Logger.Errorf("%s", tx.Error)
		return tx.Error
	}

	if !c.IsSenderSame(req.MessageOwnerJid) {
		c.Logger.Debugf("The reaction sender is not same as the vo message owner")
		return nil
	}

	ctxInfo := &waE2E.ContextInfo{
		StanzaID:    proto.String(reactKey.GetID()),
		Participant: proto.String(reactKey.GetParticipant()),
		QuotedMessage: &waE2E.Message{
			Conversation: proto.String("Reply placeholder. If you are seeing this, maybe your app is broken for some reason."),
		},
	}

	if isApproved {
		mediaBytes, err := downloadMediaFromVoRequest(c, req)
		if err != nil {
			c.Logger.Errorf("Failed to download: %s", tx.Error)
			return err
		}

		resp, err := c.Client.Upload(mediaBytes, whatsmeow.MediaImage)
		if err != nil {
			c.Logger.Errorf("Failed to upload: %s", tx.Error)
			return err
		}

		msg := &waE2E.Message{}
		if req.MediaType == string(whatsmeow.MediaImage) {
			now := time.Now().Unix()
			msg.ImageMessage = &waE2E.ImageMessage{
				Mimetype:          proto.String("image/jpeg"),
				URL:               proto.String(resp.URL),
				DirectPath:        proto.String(resp.DirectPath),
				MediaKey:          resp.MediaKey,
				FileEncSHA256:     resp.FileEncSHA256,
				FileSHA256:        resp.FileSHA256,
				FileLength:        proto.Uint64(resp.FileLength),
				MediaKeyTimestamp: proto.Int64(now),
				ContextInfo:       ctxInfo,
			}
		} else if req.MediaType == string(whatsmeow.MediaVideo) {
			now := time.Now().Unix()
			msg.VideoMessage = &waE2E.VideoMessage{
				Mimetype:          proto.String("video/mp4"),
				URL:               proto.String(resp.URL),
				DirectPath:        proto.String(resp.DirectPath),
				MediaKey:          resp.MediaKey,
				FileEncSHA256:     resp.FileEncSHA256,
				FileSHA256:        resp.FileSHA256,
				FileLength:        proto.Uint64(resp.FileLength),
				MediaKeyTimestamp: proto.Int64(now),
				ContextInfo:       ctxInfo,
			}
		} else if req.MediaType == string(whatsmeow.MediaAudio) {
			now := time.Now().Unix()
			msg.AudioMessage = &waE2E.AudioMessage{
				Mimetype:          proto.String("audio/ogg"),
				URL:               proto.String(resp.URL),
				DirectPath:        proto.String(resp.DirectPath),
				MediaKey:          resp.MediaKey,
				FileEncSHA256:     resp.FileEncSHA256,
				FileSHA256:        resp.FileSHA256,
				FileLength:        proto.Uint64(resp.FileLength),
				MediaKeyTimestamp: proto.Int64(now),
				ContextInfo:       ctxInfo,
			}
		} else {
			return fmt.Errorf("unexpected media type %s", req.MediaType)
		}

		c.SendMessage(msg)
	} else {
		c.SendMessage(&waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String("Request denied."),
				ContextInfo: ctxInfo,
			},
		})
	}

	return nil
}
