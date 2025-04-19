package message

import (
	"fmt"
	"kano/internals/database"
	"slices"
	"strings"

	"github.com/lib/pq"
	"go.mau.fi/whatsmeow/types"
)

func (ctx MessageContext) TaggedHandler() {
	if ctx.Instance.ChatJID().Server != types.GroupServer {
		return
	}

	tagged := ctx.Parser.Tagged()

	if len(tagged) == 0 {
		return
	}

	db := database.GetDB()
	var jids []string

	grp := ctx.Instance.Group
	if grp == nil {
		fmt.Println("Chat is not a group, ignoring...")
		return
	}

	if slices.Contains(tagged, "all") {
		rows, err := db.Query("SELECT c.jid FROM participant p JOIN contact c ON c.id = p.contact_id AND p.group_id = $1", grp.ID)
		if err != nil {
			fmt.Println("Failed to query all participants", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var jid string
			err := rows.Scan(&jid)
			if err != nil {
				fmt.Println("Failed to scan participant jid", err)
				return
			}
			jids = append(jids, jid)
		}

		ctx.Instance.ReplyWithTags("@all", jids)
		return
	}

	allKeys := make(map[string]bool)
	var allTitles []string
	specials := []string{"member", "manager", "admin", "superadmin"}
	for _, special := range specials {
		if slices.Contains(tagged, special) {
			rows, err := db.Query("SELECT c.jid FROM participant p JOIN contact c ON c.id = p.contact_id AND p.group_id = $1 AND p.role = $2", grp.ID, strings.ToUpper(special))
			if err != nil {
				fmt.Println("Failed to query all participants", err)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var jid string
				err := rows.Scan(&jid)
				if err != nil {
					fmt.Println("Failed to scan participant jid", err)
					return
				}
				jids = append(jids, jid)
				allKeys[special] = true
			}
		}
	}

	rows, err := db.Query("SELECT c.jid, gt.title_name FROM group_title gt JOIN group_title_holder gth ON gt.id = gth.group_title_id JOIN participant p ON gth.participant_id = p.id AND gth.holding = true JOIN contact c ON p.contact_id = c.id WHERE gt.group_id = $1 AND gt.title_name = ANY($2::varchar[]) GROUP BY c.jid, gt.title_name", grp.ID, pq.Array(tagged))
	if err != nil {
		fmt.Println("Something went wrong when querying tag jids", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var jid, title_name string
		err := rows.Scan(&jid, &title_name)
		if err != nil {
			fmt.Println("Something went wreng when scanning result jid", err)
			return
		}

		jids = append(jids, jid)
		if _, val := allKeys[title_name]; !val {
			allKeys[title_name] = true
			allTitles = append(allTitles, title_name)
		}
	}

	allKeys = make(map[string]bool)
	var uniqJids []string
	for _, item := range jids {
		if _, val := allKeys[item]; !val {
			allKeys[item] = true
			uniqJids = append(uniqJids, item)
		}
	}

	if len(uniqJids) == 0 {
		return
	}

	ctx.Instance.ReplyWithTags(fmt.Sprintf("@%s", strings.Join(allTitles, " @")), uniqJids)
}
