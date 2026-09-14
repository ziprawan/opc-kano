package sticker

const STICKER_SIZE_PX = 512

type WhatsAppStickerMetadata struct {
	StickerPackId        string `json:"sticker-pack-id,omitempty"`
	StickerPackName      string `json:"sticker-pack-name,omitempty"`
	StickerPackPublisher string `json:"sticker-pack-publisher,omitempty"`
	AndroidAppStoreLink  string `json:"android-app-store-link,omitempty"`
	IosAppStoreLink      string `json:"ios-app-store-link,omitempty"`
	Emojis               string `json:"emojis,omitempty"`
	AccessibilityText    string `json:"accessibility-text,omitempty"`
}

type FFProbeResult struct {
	Streams []Stream `json:"streams,omitempty"`
}

type Stream struct {
	CodecType string `json:"codec_type,omitempty"`
	Duration  string `json:"duration,omitempty"`
}
