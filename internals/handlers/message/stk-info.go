package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

var StkInfoMan = CommandMan{
	Name:     "stkinfo - Lihat info mentahan stiker",
	Synopsis: []string{"stkinfo"},
	Description: []string{
		"Melihat info mentahan stiker yang di-reply dalam format JSON. Info stiker yang dimaksud adalah metadata yang ditaruh pada bagian EXIF di dalam stiker khususnya pada entry tag 0x4157 (WA).",
		"Tidak ada argumen yang diperlukan untuk perintah ini.",
	},

	SeeAlso: []SeeAlso{
		{Content: "stk", Type: SeeAlsoTypeCommand},
	},
	Source: "stk-info.go",
}

func StkInfoHandler(ctx *MessageContext) {
	repliedMsg, err := ctx.Instance.ResolveReplyMessage(false)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengambil pesan yang di-reply", true)
		return
	}
	if repliedMsg == nil {
		ctx.Instance.Reply("Mana reply-nya", true)
		return
	}

	msg := repliedMsg.Event.RawMessage

	if stk := msg.StickerMessage; stk != nil {
		downloaded_bytes, err := ctx.Instance.Client.Download(context.Background(), msg.GetStickerMessage())
		if err != nil {
			ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat mengunduh media %s", err), true)
			return
		}

		header := "EXIF"
		read := false
		verified := 0
		content := []byte{}

		for idx, b := range downloaded_bytes {
			if b == 'E' && !read {
				fmt.Println("Pertama", read, verified, idx, b)
				read = true
				verified = 1
				continue
			}

			if read && verified < 4 {
				if b == header[verified] {
					fmt.Println("Kedua", read, verified, idx, b)
					verified++
				} else {
					fmt.Println("Ketiga", read, verified, idx, b)
					verified = 0
					read = false
				}
			} else if read && verified < 18 {
				fmt.Println("Keempat", read, verified, idx, b)
				verified++
			} else if read && verified < 20 {
				if verified == 18 && b == 'A' {
					fmt.Println("Kelima", read, verified, idx, b)
					verified++
				} else if verified == 19 && b == 'W' {
					fmt.Println("Keenam", read, verified, idx, b)
					verified++
				} else {
					fmt.Println("Ketujuh", read, verified, idx, b)
					break
				}
			} else if read && verified < 30 {
				fmt.Println("Kedelapan", read, verified, idx, b)
				verified++
			} else if read && verified >= 30 {
				fmt.Println("Kesembilan", read, verified, idx, b)

				// Ignore the exif padding, or maybe we should stop?
				if b == '\x00' {
					continue
				}

				content = append(content, b)
			}

			fmt.Println("Kesepuluh", read, verified, idx, b)

		}

		if len(content) == 0 {
			ctx.Instance.Reply("Tidak dapat mengambil metadata stiker", true)
			return
		}

		var dest bytes.Buffer
		err = json.Indent(&dest, content, "", "  ")
		if err != nil {
			ctx.Instance.Reply("Metadata stiker tidak mengikuti format JSON", true)
			fmt.Println(err)
			return
		}

		ctx.Instance.Reply(dest.String(), true)

	} else {
		ctx.Instance.Reply("Bukan stiker", true)
	}
}
