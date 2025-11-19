package handler

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func Handle(cli *whatsmeow.Client, evt any) error {
	LogEvent(evt)

	switch ev := evt.(type) {
	case *events.ChatPresence:
		return ChatPresence(ev)
	case *events.Connected:
		return Connected(cli)
	case *events.Disconnected:
		return Disconnected()
	case *events.LoggedOut:
		return Login(cli)
	case *events.Message:
		return Message(cli, ev)
	case *events.OfflineSyncCompleted:
		return OfflineSyncCompleted(ev)
	case *events.OfflineSyncPreview:
		return OfflineSyncPreview(ev)
	case *events.Picture:
		return Picture(ev)
	case *events.Receipt:
		return Receipt(ev)
	case *events.UndecryptableMessage:
		return UndecryptableMessage(cli, ev)
	default:
		return defaultHandler(ev)
	}
}
