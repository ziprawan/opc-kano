package six

import (
	"fmt"
	"kano/internal/utils/messageutil"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func helpHandler(c *messageutil.MessageContext) error {
	sender := c.GetChat()
	if sender.Server != types.HiddenUserServer {
		c.QuoteReply("Utilitas SIX. Gunakan di private chat untuk informasi lebih lanjut.")
		return nil
	}

	args := c.Parser.Args
	if len(args) > 1 {
		first := args[0].Content.Data
		second := args[1].Content.Data

		if first == "help" {
			theFunc, ok := helpMap[second]
			if ok {
				theFunc(c)
				return nil
			}
		}
	}

	theCmd := fmt.Sprintf("%s%s", c.Parser.Command.UsedPrefix, c.Parser.Command.Name.Data)

	var msg strings.Builder
	fmt.Fprintf(&msg, "Penggunaan: %s [perintah] [opsi perintah]\n", theCmd)
	fmt.Fprintf(&msg, "Utilitas SIX (ITB). Untuk saat ini hanya tersedia utilitas yang berkaitan dengan jadwal kuliah dan info kelas.\n")
	fmt.Fprintf(&msg, "Gunakan `%s <perintah>` untuk melihat bantuan lebih lanjut.\n", theCmd)
	fmt.Fprintf(&msg, "\n")
	fmt.Fprintf(&msg, "Perintah tersedia:\n")
	fmt.Fprintf(&msg, "\t`reminder` (Alias: `r`)\n")
	fmt.Fprintf(&msg, "\tMenambahkan reminder waktu kelas dimulai dan/atau berakhir.\n")
	fmt.Fprintf(&msg, "\n")
	fmt.Fprintf(&msg, "\t`follow` (Alias: `f`)\n")
	fmt.Fprintf(&msg, "\tIkuti perubahan kelas (kuota, dosen, jadwal, hingga ruangan).\n")
	fmt.Fprintf(&msg, "\n")
	fmt.Fprintf(&msg, "\t`help`\n")
	fmt.Fprintf(&msg, "\tMenampilkan pesan bantuan ini. Gunakan %s help [command] untuk pesan bantuan per perintah.\n", theCmd)
	fmt.Fprintf(&msg, "\n")
	fmt.Fprintf(&msg, "\t`update` (Hanya pemilik bot)\n")
	fmt.Fprintf(&msg, "\tJadwal SIX biasa diperbarui setiap jam, tepat di menit 00. Dalam keadaan tertentu, perintah ini dapat digunakan untuk memaksa perbarui jadwal.\n")

	c.QuoteReply("%s", msg.String())
	return nil
}
