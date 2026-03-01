package handles

import (
	"encoding/json"
	"kano/internal/config"
	"kano/internal/utils/chatutil/grouputil"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/types"
)

func Test(ctx *messageutil.MessageContext) error {
	if !ctx.IsSenderSame(config.GetConfig().OwnerJID) {
		return nil
	}

	jid, _ := types.ParseJID("120363320329260818@g.us")
	info, err := grouputil.InitDb(ctx.Client.GetClient(), jid)
	if err != nil {
		ctx.QuoteReply("Failed to init group: %s", err)
		return err
	}
	mar, _ := json.MarshalIndent(info, "", "  ")
	ctx.QuoteReply("```%s```", string(mar))

	return nil
}
