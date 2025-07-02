package message

import (
	"fmt"
	"time"
)

var PingMan = CommandMan{
	Name:     "ping - cek latensi",
	Synopsis: []string{"ping"},
	Description: []string{
		"Cek latensi respon bot dalam milidetik ~ Biasa digunakan untuk mengecek apakah bot aktif atau tidak.",
		"Hasil latensi sangat tidak akurat karena mengandalkan waktu dari pesan yang dikirim. Sedangkan waktu yang ada di dalam pesan itu dalam detik, sehingga memiliki ketidakpastian sebesar 1 detik.",
		"Tidak ada argumen yang diperlukan untuk perintah ini.",
	},
	SeeAlso: []SeeAlso{},
	Source:  "ping.go",
}

func PingHandler(ctx *MessageContext) {
	msgTime := ctx.Instance.Event.Info.Timestamp.UnixMilli()
	currTime := time.Now().UnixMilli()
	diff := float64((currTime - msgTime)) / 1000

	if diff < 0 {
		diff = -diff
	}

	fmt.Println(msgTime, currTime)

	rep, err := ctx.Instance.Reply(fmt.Sprintf("Pong!\nDiff time: %.3f ± 1 s", diff), true)
	fmt.Printf("Rep: %+v\nErr: %+v", rep, err)
}
