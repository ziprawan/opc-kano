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

	text := msgInstance.Text()
	parsed, err := msgInstance.InitParser([]string{".", "!", "/"})

	if err != nil {
		fmt.Println(err)
	}

	handlerCtx := handler.InitHandlerContext(&msgInstance, &parsed)
	cmd := parsed.GetCommand()

	// jsonByte, err := json.MarshalIndent(event.RawMessage, "", " ")

	// fmt.Printf("%+v\n", event)

	// if err != nil {
	// 	client.SendMessage(context.Background(), event.Info.Chat, &waE2E.Message{
	// 		Conversation: proto.String(err.Error()),
	// 	})
	// }

	// client.SendMessage(context.Background(), event.Info.Chat, &waE2E.Message{
	// 	Conversation: proto.String(string(jsonByte)),
	// })

	handlerCtx.TaggedHandler()

	if msgInstance.Reaction() != nil {
		handlerCtx.VoHandler()
	}

	if cmd.Command == "ping" {
		handlerCtx.PingHandler()
	}

	if cmd.Command == "vo" {
		handlerCtx.VoHandler()
	}

	if cmd.Command == "stk" {
		handlerCtx.StkHandler()
	}

	if cmd.Command == "stkinfo" {
		handlerCtx.StkInfoHandler()
	}

	if cmd.Command == "confess" {
		handlerCtx.ConfessHandler()
	}

	if cmd.Command == "login" {
		handlerCtx.LoginHandler()
	}

	fmt.Println("Received conversation: ", text)

	return nil
}
