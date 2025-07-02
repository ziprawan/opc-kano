package message

import (
	"encoding/json"
	"fmt"
)

var JsonMan = CommandMan{
	Name:     "json - lihat struktur pesan",
	Synopsis: []string{"json"},
	Description: []string{
		"Melihat struktur pesan dalam format JSON dan biasanya digunakan untuk keperluan debugging. Struktur pesan hanya mengikuti _struct_ yang sudah terdefinisi di dalam package go.mau.fi/whatsmeow/proto/waE2E (lihat bagian SEE ALSO nomor 1)",
		"Tidak ada argumen yang diperlukan untuk perintah ini.",
	},

	SeeAlso: []SeeAlso{
		{Content: "https://pkg.go.dev/go.mau.fi/whatsmeow/proto/waE2E", Type: SeeAlsoTypeExternalLink},
	},
	Source: "json.go",
}

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
