package client

import (
	"context"
	"kano/internal/config"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

var log = config.GetLogger().Sub("Client")

func (c *ClientContext) SendMessage(to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	log.Debugf("Sending message to %s", to.String())
	return c.client.SendMessage(context.Background(), to, message, extra...)
}

func (c *ClientContext) Download(msg whatsmeow.DownloadableMessage) ([]byte, error) {
	return c.client.Download(context.Background(), msg)
}

func (c *ClientContext) Upload(content []byte, mediaType whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
	return c.client.Upload(context.Background(), content, mediaType)
}
