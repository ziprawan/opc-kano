package client

import (
	"context"

	"go.mau.fi/whatsmeow/types"
)

func (c *ClientContext) GetLIDForPN(pn types.JID) (types.JID, error) {
	return c.Store.LIDs.GetLIDForPN(context.Background(), pn)
}

func (c *ClientContext) GetPNForLID(lid types.JID) (types.JID, error) {
	return c.Store.LIDs.GetPNForLID(context.Background(), lid)
}
