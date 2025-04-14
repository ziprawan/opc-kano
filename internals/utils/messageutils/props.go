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
	return &m.Event.Info.Chat
}

func (m Message) SenderJID() *types.JID {
	nonAD := m.Event.Info.Sender.ToNonAD()
	return &nonAD
}

func (m Message) Conversation() string {
	var text string

	msg := m.Event.Message

	if c := m.Event.Message.GetConversation(); c != "" {
		text = c
	} else if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		text = *msg.ExtendedTextMessage.Text
	}

	return strings.TrimSpace(text)
}

func (m Message) Caption() string {
	var caption string

	if d := m.Event.RawMessage.DocumentWithCaptionMessage; d != nil {
		if d.Message.DocumentMessage.Caption != nil {
			caption = *d.Message.DocumentMessage.Caption
		}
	} else if i := m.Event.Message.ImageMessage; i != nil {
		if i.Caption != nil {
			caption = *i.Caption
		}
	} else if v := m.Event.Message.VideoMessage; v != nil {
		if v.Caption != nil {
			caption = *v.Caption
		}
	}

	return caption
}

func (m Message) Text() string {
	var text string

	if con := m.Conversation(); con != "" {
		text = con
	} else if cap := m.Caption(); cap != "" {
		text = cap
	}

	return text
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

func raw(ev *events.Message) (ret string, err error) {
	mar, err := json.Marshal(ev)
	return string(mar), err
}

func (m Message) Raw() (string, error) {
	return raw(m.Event)
}

func (m Message) ResolveReplyMessage(force_save bool) (*Message, error) {
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
				}

				if force_save {
					replMsg.SaveToDatabase()
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
		return nil, nil
	}

	key := react.Key
	if key == nil {
		return nil, nil
	}

	actualSender := *key.RemoteJID
	if key.Participant != nil {
		actualSender = *key.Participant
	}

	sender, err := types.ParseJID(actualSender)
	if err != nil {
		return nil, err
	}

	fmt.Println("VARS:", *key.ID, m.ChatJID().String(), acc.ID)
	var dbMsg events.Message
	var raw string
	err = db.QueryRow("SELECT raw FROM message_with_jid WHERE message_id = $1 AND entity_jid = $2 AND account_id = $3", *key.ID, m.ChatJID().String(), acc.ID).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(raw), &dbMsg)
	if err != nil {
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
	}

	return &rectMsg, nil
}
