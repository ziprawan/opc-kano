package main

import (
	"context"
	"fmt"
	"kano/internal/config"
	"kano/internal/cronjobs"
	"kano/internal/handler"
	_ "kano/internal/message/handles" // Triggering the command aliases indexing
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/netresearch/go-cron"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	config.Init()
	defer config.GetLogger().Close()

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour |
			cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "postgres", config.GetConfig().DatabaseURL, dbLog)
	if err != nil {
		panic(err)
	}

	var deviceStore *store.Device
	deviceStore, err = container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	var eventHandler = func(evt any) {
		err = handler.Handle(client, evt)
		if err != nil {
			logger := config.GetLogger()
			logger.Errorf("Event handler goes wrong: %s", err.Error())
		}
	}

	client.AddEventHandler(eventHandler)

	c.AddFunc("*/10 * * * * *", cronjobs.SixReminder(client))
	id, err := c.AddFunc("@hourly", cronjobs.SixUpdateSchedules(client))
	if err != nil {
		panic(err)
	}
	fmt.Println("Registered with ID", id)

	handler.Connect(client)
	c.Start()

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt, syscall.SIGTERM)
	<-sign

	client.Disconnect()
}
