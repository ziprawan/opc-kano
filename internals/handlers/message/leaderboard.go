package message

import (
	"database/sql"
	"fmt"
	"kano/internals/database"
	"strings"
)

var LeaderboardMan = CommandMan{
	Name:     "leaderboard - Papan peringkat grup",
	Synopsis: []string{"leaderboard"},
	Description: []string{
		"Menampilkan papan peringkat (leaderboard) dari partisipan di grup. Perintah ini hanya bisa digunakan di dalam grup.",
		"Papan peringkat akan mengurutkan dan mengambil 5 skor permainan terbesar dari partisipan grup. Serta mengambil rank dan skor permainan dari sang pengirim meskipun bukan masuk ke dalam 5 terbesar.\nPartisipan yang memiliki skor permainan 0 tidak akan dihitung di dalam rank.",
		"Nama yang tertera di dalam papan peringkat akan memprioritaskan nama custom yang pernah diatur oleh partisipan. Jika tidak ada, maka akan menggunakan nama WhatsApp-nya, kemudian nomor telepon dalam format internasional tanpa tanda tambah.",
		"Tidak ada argumen yang diperlukan untuk perintah ini.",
	},

	SeeAlso: []SeeAlso{
		{Content: "setname", Type: SeeAlsoTypeCommand},
	},
	Source: "leaderboard.go",
}

const LEADERBOARD_QUERY = `WITH
	"ranked" AS (
		SELECT
			"c"."jid",
			"c"."push_name",
			"c"."custom_name",
			"cs"."game_points",
			ROW_NUMBER() OVER (
				ORDER BY
					"cs"."game_points" DESC
			) AS "rank"
		FROM
			"participant" "p"
			INNER JOIN "contact_settings" "cs" ON "cs"."id" = "p"."contact_id"
			AND "cs"."game_points" != 0
			INNER JOIN "contact" "c" ON "c"."id" = "cs"."id"
		WHERE
			"p"."group_id" = $1
	),
	"top5" AS (
		SELECT
			*
		FROM
			"ranked"
		ORDER BY
			"rank"
		LIMIT
			5
	),
	"sender" AS (
		SELECT
			*
		FROM
			"ranked"
		WHERE
			"jid" = $2
	)
SELECT
	*
FROM
	"top5"
UNION
SELECT
	*
FROM
	"sender"
ORDER BY
	"rank"`

func LeaderboardHandler(ctx *MessageContext) {
	if ctx.Instance.Group == nil {
		ctx.Instance.Reply("Pakenya di grup aja ya bang", true)
		return
	}

	db := database.GetDB()
	rows, err := db.Query(LEADERBOARD_QUERY, ctx.Instance.Group.ID, ctx.Instance.Contact.JID.String())
	if err != nil {
		fmt.Println("leaderboard: Failed to create query:", err)
		ctx.Instance.Reply("Terjadi kesalahan saat membuat kueri", true)
		return
	}
	defer rows.Close()

	msg := "*===== LEADERBOARD =====*\n\n"

	for rows.Next() {
		var jid string
		var pushName, customName sql.NullString
		var gamePoints, rank int
		err := rows.Scan(&jid, &pushName, &customName, &gamePoints, &rank)
		if err != nil {
			fmt.Println("leaderboard: Failed to scan result:", err)
			ctx.Instance.Reply("Terjadi kesalahan saat mengambil data", true)
			return
		}

		name := strings.Split(jid, "@")[0]
		if customName.Valid && len(customName.String) > 0 {
			name = customName.String
		} else if pushName.Valid && len(pushName.String) > 0 {
			name = pushName.String
		}

		add := ""
		if jid == ctx.Instance.Contact.JID.String() {
			add = fmt.Sprintf("%d. *%s (%d points)*", rank, name, gamePoints)
		} else {
			add = fmt.Sprintf("%d. %s (%d points)", rank, name, gamePoints)
		}

		msg += add + "\n"
	}

	ctx.Instance.Reply(msg, true)
}
