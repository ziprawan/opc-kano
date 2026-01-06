package handles

import (
	"kano/internal/config"
	"kano/internal/utils/messageutil"
)

func Test(ctx *messageutil.MessageContext) error {
	if !ctx.IsSenderSame(config.GetConfig().OwnerJID) {
		return nil
	}

	return nil
}
