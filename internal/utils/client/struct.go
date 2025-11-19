package client

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
)

type ClientContext struct {
	client *whatsmeow.Client
	Store  *store.Device
}

func CreateContext(cli *whatsmeow.Client) *ClientContext {
	return &ClientContext{
		client: cli,
		Store:  cli.Store,
	}
}
