package cronjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"kano/internal/utils/six"
	"kano/internal/utils/six/schedules"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func SixUpdateSchedules(cli *whatsmeow.Client) func() {
	conf := config.GetConfig()
	send := func(msg string) {
		if conf.OwnerJID.User == "" {
			fmt.Println(msg)
			return
		}

		cli.SendMessage(context.Background(), conf.OwnerJID, &waE2E.Message{
			Conversation: proto.String(msg),
		})
	}

	return func() {
		fmt.Println("Running SixUpdateSchedules")
		send("Starting schedule update...")
		subjects, err := six.GetAllSchedules()
		if err != nil {
			send(fmt.Sprintf("Failed to fetch schedules: %s", err))
			return
		}

		diff, err := schedules.GetScheduleDiff(subjects)
		if err != nil {
			send(fmt.Sprintf("Failed to generate diff: %s", err))
			return
		}

		err = schedules.ApplyDiff(diff)
		if err != nil {
			send(fmt.Sprintf("Failed to apply diff: %s", err))
			return
		}

		schedules.CleanupTmpFiles()

		addedSubjects := 0
		removedSubjects := 0
		modifiedSubjects := 0
		addedClasses := 0
		removedClasses := 0
		modifiedClasses := 0
		for _, sems := range diff {
			addedSubjects += len(sems.AddedSubjects)
			removedSubjects += len(sems.RemovedSubjects)
			modifiedSubjects += len(sems.ModifiedSubjects)

			for _, sub := range sems.ModifiedSubjects {
				addedClasses += len(sub.AddedClasses)
				removedClasses += len(sub.RemovedClasses)
				modifiedClasses += len(sub.ModifiedClasses)
			}
		}

		send(
			fmt.Sprintf(
				"Schedules updated, with AddedSubjects=%d, RemovedSubjects=%d, ModifiedSubjects={%d, AddedClasses=%d, RemovedClasses=%d, ModifiedClasses=%d}",
				addedSubjects, removedSubjects, modifiedSubjects, addedClasses, removedClasses, modifiedClasses,
			),
		)

		mar, err := json.MarshalIndent(diff, "", "\t")
		if err != nil {
			return
		}

		resp, err := cli.Upload(context.Background(), mar, whatsmeow.MediaDocument)
		if err != nil {
			return
		}

		now := time.Now()
		cli.SendMessage(context.Background(), conf.OwnerJID, &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				Mimetype:          proto.String("application/json"),
				FileName:          proto.String(fmt.Sprintf("scheddiff_%4d-%02d-%02d_%02d.json", now.Year(), now.Month(), now.Day(), now.Hour())),
				URL:               proto.String(resp.URL),
				DirectPath:        proto.String(resp.DirectPath),
				MediaKey:          resp.MediaKey,
				FileEncSHA256:     resp.FileEncSHA256,
				FileSHA256:        resp.FileSHA256,
				FileLength:        proto.Uint64(resp.FileLength),
				MediaKeyTimestamp: proto.Int64(now.Unix()),
			},
		})
	}
}
