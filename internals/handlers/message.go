package handlers

import (
	"fmt"
	handler "kano/internals/handlers/message"
	"kano/internals/utils/messageutils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

var workerLimit = 10
var sem = make(chan struct{}, workerLimit)

func MessageEventHandler(client *whatsmeow.Client, event *events.Message) error {
	sem <- struct{}{}        // Acquire a worker slot
	defer func() { <-sem }() // Release worker slot

	msgInstance := messageutils.CreateMessageInstance(client, event)
	msgInstance.MarkRead()
	err := msgInstance.SaveToDatabase()
	if err != nil {
		fmt.Printf("KESALAHAN SAAT MENYIMPAN PESAN: %+v\n", err)
	} else {
		fmt.Printf("PESAN BERHASIL DISIMPAN!\n")
	}

	parsed, err := msgInstance.InitParser([]string{".", "!", "/"})
	cmd := parsed.GetCommand()

	if err != nil {
		fmt.Println(err)
	}

	handlerCtx := handler.InitHandlerContext(msgInstance, &parsed)
	handlerCtx.TaggedHandler()

	if msgInstance.Reaction() != nil {
		handler.VoHandler(handlerCtx)
	}

	handlerCtx.Handle()
	if cmd.Command == "help" {
		handlerCtx.HelpHandler()
	}

	return nil
}
