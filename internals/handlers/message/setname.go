package message

import "fmt"

var SetnameMan = CommandMan{
	Name:     "setname - Atur nama custom",
	Synopsis: []string{"setname [NAMA_BARU ...]"},
	Description: []string{
		"Ubah nama custom yang tersimpan di basis data bot. Hal ini bisa digunakan dalam beberapa kasus seperti tampilan nama di dalam website bot, nama publisher dari stk, leaderboard, dan lainnya.",
		"*NAMA_BARU*\n{SPACE}Nama baru yang ingin diatur. Jika argumen ini tidak diberikan, maka bot akan memberitahu nama custom pengirim saat ini.",
	},

	SeeAlso: []SeeAlso{
		{Content: "stk", Type: SeeAlsoTypeCommand},
		{Content: "leaderboard", Type: SeeAlsoTypeCommand},
	},
	Source: "setname.go",
}

func SetNameHandler(ctx *MessageContext) {
	if ctx.Instance.Contact == nil {
		fmt.Println("setname: Contact is nil")
		return
	}

	ctc := ctx.Instance.Contact

	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		msg := "Berikan nama custom baru yang kamu inginkan"
		if ctc.CustomName.Valid {
			msg += fmt.Sprintf("\nNama kamu saat ini: %s", ctc.CustomName.String)
		}
		ctx.Instance.Reply(msg, true)
		return
	}

	ctc.CustomName.Valid = true
	ctc.CustomName.String = ctx.Parser.Text[args[0].Start:]
	err := ctc.Save()
	if err != nil {
		fmt.Println("setname: Failed to save name to database:", err)
		ctx.Instance.Reply("Terjadi kesalahan saat menyimpan nama", true)
	} else {
		ctx.Instance.Reply(fmt.Sprintf("Berhasil menyimpan nama sebagai: %s", ctc.CustomName.String), true)
	}
}
