package handler

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
)

func Login(cli *whatsmeow.Client) error {
	qrChan, _ := cli.GetQRChannel(context.Background())
	err := cli.Connect()
	if err != nil {
		panic(err)
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			fmt.Println("QR Code:", evt.Code)
		} else {
			fmt.Println("Login event:", evt.Event)
		}
	}

	return nil
}

func Connect(cli *whatsmeow.Client) {
	if cli.Store.ID == nil {
		Login(cli)
	} else {
		cli.Connect()
	}

	fmt.Println("Running app")
}
