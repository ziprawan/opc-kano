package handles

import (
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func Vo(c *messageutil.MessageContext) error {
	c.QuoteReply("Not implemented yet. Wait for future update.")
	return ErrNotImplemented
}

func NVo(c *messageutil.MessageContext) (err error) {
	repliedMsg := c.GetRepliedMessage()
	if repliedMsg == nil {
		_, err = c.QuoteReply("Please reply to a view-once message.")
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

	msg := &waE2E.Message{}
	var downloadable whatsmeow.DownloadableMessage

	if i := repliedMsg.GetImageMessage(); i != nil && i.GetViewOnce() {
		i.ViewOnce = proto.Bool(false)
		downloadable = i
		msg.ImageMessage = i
	} else if v := repliedMsg.GetVideoMessage(); v != nil && v.GetViewOnce() {
		v.ViewOnce = proto.Bool(false)
		downloadable = v
		msg.VideoMessage = v
	} else if a := repliedMsg.GetAudioMessage(); a != nil && a.GetViewOnce() {
		a.ViewOnce = proto.Bool(false)
		downloadable = a
		msg.AudioMessage = a
	}

	if downloadable == nil {
		_, err = c.QuoteReply("Replied message is not a view-once message.")
		return
	}

	if val := c.ValidateDownloadableMessage(downloadable); val != nil {
		_, err = c.QuoteReply("Replied view once has missing fields. Most likely you are using the latest WhatsApp version.\nDebug info: %s", val.Error())
		return
	}

	_, err = c.Client.SendMessage(c.GetChat(), msg)
	return
}
