package six

import (
	"fmt"
	"kano/internal/utils/messageutil"
	"strings"
)

func reminderHelp(c *messageutil.MessageContext) {
	theCmd := fmt.Sprintf("%s%s", c.Parser.Command.UsedPrefix, c.Parser.Command.Name.Data)

	var msg strings.Builder
	fmt.Fprintln(&msg, "*Pengingat jadwal kelas SIX*")
	fmt.Fprintln(&msg, "Membuat pengingat baru untuk jadwal kelas. Pengingat dapat diatur tepat saat kelas dimulai maupun saat kelas berakhir, serta dapat digeser sesuai dengan kebutuhan.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "*Ringkasan:*")
	fmt.Fprintf(&msg, "`%s r|reminder <code>-<number> [[^][+|-]<time>]`\n", theCmd)
	fmt.Fprintln(&msg, "- Dapat menggunakan perintah `r` ataupun `reminder`.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "Penjelasan parameter:")
	fmt.Fprintln(&msg, "`<code>-<number>`")
	fmt.Fprintln(&msg, "Kode matkul dan nomor kelas. Nomor kelas dapat ditulis dengan `1` ataupun `01`.")
	fmt.Fprintln(&msg, "Contoh: ET2202-01; ET2201-2")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "`[[^][+|-]<time>]` (Opsional)")
	fmt.Fprintln(&msg, "Menggeser waktu pengingat. Jika tidak diberikan, secara default akan menambahkan pengingat tepat saat kelas dimulai.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "`<time>`: Besar penggeseran waktu dalam satuan menit. Dapat mengubah satuan dengan menambah huruf m (menit), h (jam), atau d (hari). Dapat dikombinasikan dengan syarat harus urut (d > h > m). Nilai paling minimum adalah 1 menit, nilai paling maksimum adalah 1 pekan (7d = 168h = 10080m).")
	fmt.Fprintln(&msg, "Contoh: `10` => 10 menit; `20m` => 20 menit; `3h` => 3 jam, 4d => 4 hari, 2d12h20m => 2 hari 12 jam 20 menit;")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "`[+|-]`: Opsional. Menggunakan ataupun tidak menggunakan simbol plus (+), akan menggeser waktu lebih lambat dari waktu dimulainya kelas. Sebaliknya, untuk simbol minus (-) akan menggeser waktu lebih cepat dari waktu dimulainya kelas.")
	fmt.Fprintln(&msg, "Contoh: `-20m` => 20 menit sebelum kelas dimulai; `1h` => 1 jam setelah kelas dimulai; `+1h16m` =>1 jam 16 menit setelah kelas dimulai.")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "`[^]`: Opsional. Jika digunakan, waktu berakhirnya kelas akan digunakan sebagai referensi waktu, bukan waktu dimulainya kelas. Mungkin berguna dalam beberapa kasus.")
	fmt.Fprintln(&msg, "Contoh: `^-10m` => 10 menit sebelum kelas berakhir; `^1h` => 1 jam setelah kelas berakhir;")
	fmt.Fprintln(&msg, "")
	fmt.Fprintln(&msg, "Contoh lengkap:")
	fmt.Fprintf(&msg, "`%s reminder ET1201-01 ^+1h15m`\n", theCmd)
	fmt.Fprintln(&msg, "Membuat reminder 1 jam 15 menit setelah kelas berakhir untuk matkul ET1201 kelas 01.")
	fmt.Fprintf(&msg, "`%s r ET2202-01 -10`\n", theCmd)
	fmt.Fprintln(&msg, "Membuat reminder 10 menit sebelum kelas dimulai untuk matkul ET2202 kelas 01.")
	fmt.Fprintf(&msg, "`%s r ET2201-02 ^`\n", theCmd)
	fmt.Fprintln(&msg, "Membuat reminder tepat saat kelas berakhir untuk matkul ET2201 kelas 02.")
	fmt.Fprintf(&msg, "`%s r ET2201-02`\n", theCmd)
	fmt.Fprintln(&msg, "Membuat reminder tepat saat kelas dimulai untuk matkul ET2201 kelas 02.")

	c.QuoteReply("%s", msg.String())
}
