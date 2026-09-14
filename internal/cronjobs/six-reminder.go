package cronjobs

import (
	"context"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"
	"math"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errorSent = false

func SixReminder(cli *whatsmeow.Client) func() {
	return func() {
		send := func(target types.JID, msg string) {
			cli.SendMessage(context.Background(), target, &waE2E.Message{Conversation: proto.String(msg)})
		}
		owner := config.GetConfig().OwnerJID

		now := time.Now().In(config.Jakarta)
		// now := time.Unix(1770631200, 0).In(config.Jakarta)
		dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.Jakarta)

		db := database.GetInstance()

		res, err := gorm.G[models.ClassReminderView](db).
			Joins(clause.LeftJoin.Association("Delivery"), models.NoopJoin).
			Joins(clause.LeftJoin.Association("Schedule"), models.NoopJoin).
			Joins(clause.InnerJoin.Association("SubjectClass.Subject"), models.NoopJoin).
			Where(
				// Limit at start of day until current time only
				"alert_time_unix >= ? AND alert_time_unix <= ? AND \"Delivery\".schedule_id IS NULL",
				dayStart.Unix(), now.Unix(),
			).
			Order("alert_time_unix").
			Find(context.Background())
		if err != nil {
			if !errorSent {
				send(owner, fmt.Sprintf("SixReminder: Gagal mengambil data reminder: %s", err))
				errorSent = true
			} else {
				fmt.Println(err)
			}

			return
		}
		errorSent = false

		if len(res) == 0 {
			return
		}

		jids := map[types.JID]*strings.Builder{}
		toInsert := make([]models.ClassReminderDelivery, 0, len(res))

		for _, remView := range res {
			toInsert = append(toInsert, models.ClassReminderDelivery{
				ScheduleId:       remView.ScheduleId,
				Jid:              remView.Jid,
				DeliveredForUnix: remView.AlertTimeUnix,
			})

			jid := remView.Jid
			if _, ok := jids[jid]; !ok {
				jids[jid] = &strings.Builder{}
			} else {
				fmt.Fprintln(jids[jid], "")
			}

			alertTime := time.Unix(remView.AlertTimeUnix, 0)
			schedTime := remView.Schedule.Start

			diff := alertTime.Sub(schedTime)

			kapan := ""
			if diff < 0 {
				kapan = "sebelum"
			} else {
				kapan = "setelah"
			}
			ref := ""
			if remView.AnchorAtEnd {
				ref = "berakhir"
			} else {
				ref = "dimulai"
			}

			diff = diff.Abs()
			hari := int(math.Floor(diff.Hours() / 24))
			jam := int(math.Floor(diff.Hours())) % 24
			menit := int(math.Floor(diff.Minutes())) % 60

			fmt.Fprintf(jids[jid], "Reminder: ")

			if hari > 0 {
				fmt.Fprintf(jids[jid], "%d hari ", hari)
			}
			if jam > 0 {
				fmt.Fprintf(jids[jid], "%d jam ", jam)
			}
			if menit > 0 {
				fmt.Fprintf(jids[jid], "%d menit ", menit)
			}

			fmt.Fprintf(jids[jid], "%s kelas %s-%02d (%s) %s",
				kapan,
				remView.SubjectClass.Subject.Code,
				remView.SubjectClass.Number,
				remView.SubjectClass.Subject.Name,
				ref,
			)
		}

		tx := db.WithContext(context.Background()).CreateInBatches(&toInsert, 1000)
		if tx.Error != nil {
			if !errorSent {
				send(owner, fmt.Sprintf("SixReminder: Gagal menyimpan hasil delivery: %s", err))
				errorSent = true
			} else {
				fmt.Println(err)
			}

			return // Avoid to continue sending updates to users
		}
		errorSent = false

		for jid, builder := range jids {
			send(jid, builder.String())
		}
	}
}
