package cron

import (
	"fmt"
	cronfinaid "kano/internals/handlers/cron/finaid"

	"github.com/go-co-op/gocron/v2"
	"go.mau.fi/whatsmeow"
)

var (
	AlreadyStarted bool             = false
	sched          gocron.Scheduler = nil
)

func StartAllCron(cli *whatsmeow.Client) (gocron.Scheduler, error) {
	if AlreadyStarted {
		return sched, nil
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	j, err := s.NewJob(
		gocron.CronJob("0 * * * *", false),
		gocron.NewTask(cronfinaid.FinaidScholarshipCronFunc, cli),
	)
	if err != nil {
		return nil, err
	}

	fmt.Println(j.ID())

	AlreadyStarted = true
	sched = s

	s.Start()

	return s, nil
}
