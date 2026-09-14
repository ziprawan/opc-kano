package six

import (
	"fmt"
	"kano/internal/utils/messageutil"
	"strings"
)

func followHelp(c *messageutil.MessageContext) {
	theCmd := fmt.Sprintf("%s%s", c.Parser.Command.UsedPrefix, c.Parser.Command.Name.Data)

	var msg strings.Builder
	fmt.Fprintln(&msg, "*Ikuti perubahan kelas*")
	fmt.Fprintln(&msg, "Mengikuti segala perubahan kelas, seperti perubahan jadwal kelas, ruangan, aktivitas, dan/atau metode. Serta, jika ada, perubahan informasi kuota dan dosen juga.")
	fmt.Fprintln(&msg, "Pengecekan perbaruan dilakukan perjam, sehingga adanya kemungkinan keterlambatan info maksimum 1 jam.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "*Ringkasan:*")
	fmt.Fprintf(&msg, "`%s f|follow <code>-<number>`\n", theCmd)
	fmt.Fprintln(&msg, "- Dapat menggunakan perintah `f` ataupun `follow`.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "*Penjelasan parameter:*")
	fmt.Fprintln(&msg, "`<code>-<number>`")
	fmt.Fprintln(&msg, "Kode matkul dan nomor kelas. Nomor kelas dapat ditulis dengan `1` ataupun `01`.")
	fmt.Fprintln(&msg, "Contoh: ET2202-01; ET2201-2")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "Contoh lengkap:")
	fmt.Fprintf(&msg, "`%s follow ET1201-01`\n", theCmd)
	fmt.Fprintln(&msg, "Mengikuti segala perubahan untuk kelas ET1201-01.")
	fmt.Fprintf(&msg, "`%s f ET2202-2`\n", theCmd)
	fmt.Fprintln(&msg, "Mengikuti segala perubahan untuk kelas ET2202-02.")

	c.QuoteReply("%s", msg.String())
}
