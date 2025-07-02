package message

import (
	"database/sql"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"kano/internals/utils/saveutils"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
)

var ShippingMan = CommandMan{
	Name:     "shipping - Pasangan 2 orang harian",
	Synopsis: []string{"shipping"},
	Description: []string{
		"_*PERINGATAN: Perintah ini benar-benar memilih 2 orang secara acak semata dan diharapkan tidak ada kegaduhan yang disebabkan oleh fitur ini.*_",
		"Memilih 2 orang secara acak dalam partisipan grup dan dijadikan sebagai pasangan. Perintah ini hanya dapat digunakan di dalam grup dan menyimpannya di basis data. Pasangan akan di-reset pada pukul 00.00 zona waktu UTC.",
		"Tidak ada argumen yang diperlukan untuk perintah ini.",
	},

	SeeAlso: []SeeAlso{},
	Source:  "shipping.go",
}

func randomSelectAndRemove(s []string) ([]string, string) {
	selectedIdx := rand.Intn(len(s))
	selected := s[selectedIdx]

	ret := make([]string, 0)
	ret = append(ret, s[:selectedIdx]...)
	return append(ret, s[selectedIdx+1:]...), selected
}

func ShippingHandler(ctx *MessageContext) {
	if ctx.Instance.ChatJID().Server != types.GroupServer {
		return
	}

	acc, _ := account.GetData()
	now := time.Now()
	var settings saveutils.GroupSettings
	if ctx.Instance.Group.Settings != nil {
		settings = *ctx.Instance.Group.Settings
	} else {
		settings = saveutils.GroupSettings{}
	}

	needUpdate := false
	if settings.LastShippingTime.Valid {
		lastChosen := settings.LastShippingTime.Time
		lastChosenExpire := time.Date(
			lastChosen.Year(),
			lastChosen.Month(),
			lastChosen.Day()+1,
			0, 0, 0, 0,
			time.UTC,
		)
		if now.Unix() >= lastChosenExpire.Unix() {
			needUpdate = true
		}
	} else {
		if len(settings.ChosenShipping) != 2 {
			needUpdate = true
		}
	}

	adds := ":"
	if needUpdate {
		adds = " berhasil di-update!"
		var jids []string
		db := database.GetDB()
		rows, err := db.Query("SELECT c.jid FROM participant p INNER JOIN contact c ON c.id = p.contact_id AND p.group_id = $1 AND p.role != 'LEFT'", ctx.Instance.Group.ID)
		if err != nil {
			ctx.Instance.Reply(fmt.Sprintf("Ada yang salah saat mengambil semua data member\nErr: %s", err), true)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var jid string
			err := rows.Scan(&jid)
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Ada yang salah saat nge scan hasil kueri\nErr: %s", err), true)
				return
			}
			if jid == acc.JID.ToNonAD().String() {
				continue
			}
			jids = append(jids, jid)
		}
		var leftSide, rightSide string
		jids, leftSide = randomSelectAndRemove(jids)
		_, rightSide = randomSelectAndRemove(jids)

		settings.ChosenShipping = []string{leftSide, rightSide}
		settings.LastShippingTime = sql.NullTime{
			Time:  now,
			Valid: true,
		}
		ctx.Instance.Group.Settings = &settings
		err = ctx.Instance.Group.SaveGroupSettings()
		if err != nil {
			ctx.Instance.Reply(fmt.Sprintf("Ada yang salah saat menyimpan hasil shipping\nErr: %s", err), true)
			return
		}
	}

	expireTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+1,
		0, 0, 0, 0,
		time.UTC,
	)
	dur := expireTime.Sub(now)
	durTot := int(dur.Seconds())
	durSecs := durTot % 60
	durMins := int(durTot/60) % 60
	durHours := int(durTot/3600) % 24
	num1, _ := strconv.ParseInt(strings.Split(settings.ChosenShipping[0], "@")[0], 10, 0)
	num2, _ := strconv.ParseInt(strings.Split(settings.ChosenShipping[1], "@")[0], 10, 0)
	ctx.Instance.ReplyWithTags(fmt.Sprintf("Pasangan hari ini%s:\n@%d ❤ @%d\n\nPasangan baru akan dipilih dalam %d jam %d menit %d detik", adds, num1, num2, durHours, durMins, durSecs), []string{
		fmt.Sprintf("%d@s.whatsapp.net", num1),
		fmt.Sprintf("%d@s.whatsapp.net", num2),
	})
}
