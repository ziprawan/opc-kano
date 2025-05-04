package message

import (
	"encoding/json"
	"fmt"
)

func JSONHandler(ctx *MessageContext) {
	if ctx == nil {
		return
	}
	msg := ctx.Instance.Event.RawMessage
	marshal, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat membuat JSON dari pesan: %s", err.Error()), true)
		return
	}

	ctx.Instance.Reply(fmt.Sprintf("```%s```", string(marshal)), true)
}
