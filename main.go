package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"kano/internals/database"
	handlers "kano/internals/handlers"
	projectConfig "kano/internals/project_config"
	"kano/internals/utils/account"
	webhandlers "kano/web_handlers"
)

// Still copied from whatsmeow's example
// Gonna tidy up later after learn enough about Go
func main() {
	conf := projectConfig.LoadConfig()
	db := database.GetDB()
	acc := account.InitAccount(conf.SessionName)

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("postgres", conf.DatabaseURL, dbLog)

	if err != nil {
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetDevice(*acc.JID)
	if err != nil {
		panic(err)
	}
	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	var eventHandler = func(evt any) {
		// data, err := json.MarshalIndent(evt, "", "  ")
		// if err != nil {
		// 	panic(err)
		// }
		// os.WriteFile(fmt.Sprintf("json/%d_%T.json", time.Now().UnixMilli(), evt), data, 0644)
		switch v := evt.(type) {
		case *events.Message:
			go handlers.MessageEventHandler(client, v)
		case *events.Connected:
			fmt.Println("Connected to WA Web")

			account.SetPushName(deviceStore.PushName)
			client.SetForceActiveDeliveryReceipts(true)
			client.SendPresence(types.PresenceAvailable)
		case *events.Contact:
			marshal, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Println("Nigga")
			}
			fmt.Println("Got new contact event: ", string(marshal))
		case *events.PairSuccess:
			err := account.SaveAccount(conf.SessionName, &v.ID)
			if err != nil {
				panic(err)
			}

			fmt.Println("Pair QR Success and account added to database!")
		case *events.LoggedOut:
			if v.Reason == events.ConnectFailureLoggedOut {
				res, err := db.Exec("DELETE FROM account WHERE name = $1", conf.SessionName)
				if err != nil {
					panic(err)
				}

				fmt.Println(res)
				syscall.Exit(0)
			}
		case *events.GroupInfo:
			marsh, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(marsh))
		}
	}

	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		client.Connect()
	}

	webhandlers.Web()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
