package messageutil

import (
	"encoding/json"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type Button interface {
	GetButtonName() string
}

func GenericGetNativeFlowButton(btn Button) *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	mar, _ := json.Marshal(btn)

	return &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String(btn.GetButtonName()),
		ButtonParamsJSON: proto.String(string(mar)),
	}
}

type QuickReplyButton struct {
	ID          string `json:"id"`
	DisplayText string `json:"display_text"`
}

func (btn QuickReplyButton) GetButtonName() string {
	return "quick_reply"
}

func (btn QuickReplyButton) GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	return GenericGetNativeFlowButton(btn)
}

type SingleSelectButton struct {
	Title    string                       `json:"title"`
	Sections []SingleSelectButton_Section `json:"sections"`
}

func (btn SingleSelectButton) GetButtonName() string {
	return "single_select"
}

func (btn SingleSelectButton) GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	return GenericGetNativeFlowButton(btn)
}

type SingleSelectButton_Section struct {
	Title string                           `json:"title,omitempty"`
	Rows  []SingleSelectButton_Section_Row `json:"rows"`
}

type SingleSelectButton_Section_Row struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Header      string `json:"header,omitempty"`
}

type CtaUrlButton struct {
	DisplayText string `json:"display_text"`
	Url         string `json:"url"`
	MerchantUrl string `json:"merchant_url,omitempty"`
}

func (btn CtaUrlButton) GetButtonName() string {
	return "cta_url"
}

func (btn CtaUrlButton) GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	return GenericGetNativeFlowButton(btn)
}

type CtaCopyButton struct {
	DisplayText string `json:"display_text"`
	Code        string `json:"copy_code"`
}

func (btn CtaCopyButton) GetButtonName() string {
	return "cta_copy"
}

func (btn CtaCopyButton) GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	return GenericGetNativeFlowButton(btn)
}

type CtaCallButton struct {
	DisplayText string `json:"display_text"`
	PhoneNumber string `json:"phone_number"`
}

func (btn CtaCallButton) GetButtonName() string {
	return "cta_call"
}

func (btn CtaCallButton) GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	return GenericGetNativeFlowButton(btn)
}
