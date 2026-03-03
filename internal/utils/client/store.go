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

func (c *ClientContext) GetManyLIDsForPNs(pns []types.JID) (map[types.JID]types.JID, error) {
	return c.Store.LIDs.GetManyLIDsForPNs(context.Background(), pns)
}

func (c *ClientContext) GetJID() types.JID {
	return c.Store.GetJID()
}
func (c *ClientContext) GetLID() types.JID {
	return c.Store.GetLID()
}
