package client

import (
	"context"
	"io"
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

func (c *ClientContext) UploadReader(plaintext io.Reader, appInfo whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
	return c.client.UploadReader(context.Background(), plaintext, nil, appInfo)
}

func (c *ClientContext) GetGroupInfo(jid types.JID) (*types.GroupInfo, error) {
	return c.client.GetGroupInfo(context.Background(), jid)
}

func (c *ClientContext) GetSubGroups(community types.JID) ([]*types.GroupLinkTarget, error) {
	return c.client.GetSubGroups(context.Background(), community)
}

func (c *ClientContext) GetJoinedGroups() ([]*types.GroupInfo, error) {
	return c.client.GetJoinedGroups(context.Background())
}

func (c *ClientContext) BuildEdit(chat types.JID, id types.MessageID, newContent *waE2E.Message) *waE2E.Message {
	return c.client.BuildEdit(chat, id, newContent)
}
