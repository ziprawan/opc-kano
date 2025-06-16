package messageutils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"strings"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (m Message) ID() *string {
	return &m.Event.Info.ID
}

func (m Message) ChatJID() *types.JID {
	chat := m.Event.Info.Chat
	if chat.ToNonAD().String() == "status@broadcast" {
		return m.SenderJID()
	}

	return &chat
}

func (m Message) SenderJID() *types.JID {
	nonAD := m.Event.Info.Sender.ToNonAD()
	return &nonAD
}

func conversation(ev *events.Message) string {
	var text string

	if ev == nil || ev.Message == nil {
		return text
	}
	msg := ev.Message
	if c := msg.GetConversation(); c != "" {
		text = c
	} else if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		text = *msg.ExtendedTextMessage.Text
	}

	return strings.TrimSpace(text)
}

func (m Message) Conversation() string {
	return conversation(m.Event)
}

func caption(ev *events.Message) string {
	var caption string
	if ev == nil || ev.RawMessage == nil {
		return caption
	}

	if d := ev.RawMessage.DocumentWithCaptionMessage; d != nil {
		if e := d.Message; e != nil {
			if f := e.DocumentMessage; f != nil && f.Caption != nil {
				caption = *f.Caption
			}
		}
	} else if i := ev.RawMessage.ImageMessage; i != nil && i.Caption != nil {
		caption = *i.Caption
	} else if v := ev.RawMessage.VideoMessage; v != nil && v.Caption != nil {
		caption = *v.Caption
	}

	return strings.TrimSpace(caption)
}

func (m Message) Caption() string {
	return caption(m.Event)
}

func text(ev *events.Message) string {
	var text string
	if con := conversation(ev); con != "" {
		text = con
	} else if cap := caption(ev); cap != "" {
		text = cap
	}
	return strings.TrimSpace(text)
}

func (m Message) Text() string {
	return text(m.Event)
}

func (m Message) Reaction() *string {
	if react := m.Event.RawMessage.ReactionMessage; react != nil {
		return react.Text
	}

	return nil
}

func (m Message) IsReaction() bool {
	react := m.Event.RawMessage.ReactionMessage
	return react != nil
}

func (m Message) GetReactedMessageID() *types.MessageID {
	if react := m.Event.RawMessage.ReactionMessage; react != nil {
		return react.Key.ID
	}

	return nil
}

func marshal(ev *events.Message) (ret string, err error) {
	mar, err := json.Marshal(ev)
	return string(mar), err
}

func (m Message) Marshal() (string, error) {
	return marshal(m.Event)
}

func (m Message) ResolveReplyMessage(forceSave bool) (*Message, error) {
	msg := m.Event.RawMessage

	if extend := msg.ExtendedTextMessage; extend != nil {
		if context := extend.ContextInfo; context != nil {
			if quoted := context.QuotedMessage; quoted != nil {
				acc, err := account.GetData()
				if err != nil {
					return nil, err
				}
				sender, err := types.ParseJID(*context.Participant)
				if err != nil {
					return nil, err
				}

				replMsg := Message{
					Client: m.Client,
					Event: &events.Message{
						Info: types.MessageInfo{
							MessageSource: types.MessageSource{
								Chat:     *m.ChatJID(),
								Sender:   sender,
								IsFromMe: (*context.Participant == acc.JID.ToNonAD().String()),
							},
							ID: *context.StanzaID,
						},
						Message:    quoted,
						RawMessage: quoted,
					},
					Group:   m.Group,
					Contact: m.Contact,
				}

				if forceSave {
					err := replMsg.SaveToDatabase()
					if err != nil {
						fmt.Println("[ResolveReplyMessage()] Failed to save reply message to database", err)
					}
				}

				return &replMsg, nil
			}
		}
	}

	return nil, nil
}

func (m Message) ResolveReactedMessage() (*Message, error) {
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		return nil, err
	}

	msg := m.Event.RawMessage
	react := msg.ReactionMessage
	if react == nil {
		fmt.Println("[ResolveReactedMessage()] Message is not a ReactionMessage")
		return nil, nil
	}

	key := react.Key
	if key == nil {
		fmt.Println("[ResolveReactedMessage()] ReactionMessage has no Key field which is weird")
		return nil, nil
	}

	actualSender := *key.RemoteJID
	if key.Participant != nil {
		actualSender = *key.Participant
	}

	sender, err := types.ParseJID(actualSender)
	if err != nil {
		fmt.Println("[ResolveReactedMessage()] Failed to parse SenderJID")
		return nil, err
	}

	fmt.Println("VARS:", *key.ID, m.ChatJID().String(), acc.ID)
	var dbMsg events.Message
	var raw string
	err = db.QueryRow("SELECT raw FROM message_with_jid WHERE message_id = $1 AND entity_jid = $2 AND account_id = $3", *key.ID, m.ChatJID().String(), acc.ID).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("[ResolveReactedMessage()] Cannot find reacted message from database")
			return nil, nil
		}
		fmt.Println("[ResolveReactedMessage()] Errored when scanning result", err)
		return nil, err
	}

	err = json.Unmarshal([]byte(raw), &dbMsg)
	if err != nil {
		fmt.Println("[ResolveReactedMessage()] Errored when Unmarshal json")
		return nil, err
	}

	rectMsg := Message{
		Client: m.Client,
		Event: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Chat:     *m.ChatJID(),
					Sender:   sender,
					IsFromMe: *key.FromMe,
				},
				ID: *key.ID,
			},
			Message:    dbMsg.Message,
			RawMessage: dbMsg.RawMessage,
		},
		Group:   m.Group,
		Contact: m.Contact,
	}

	return &rectMsg, nil
}
