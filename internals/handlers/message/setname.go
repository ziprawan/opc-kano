package message

import "fmt"

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
