package ytdlpbind

type YtDlpJSON struct {
	Id          string `json:"id"`
	Title       string `json:"title,omitempty"`
	FullTitle   string `json:"fulltitle"`
	Description string `json:"description"`

	YtDlpFormat
	Formats          []YtDlpFormat `json:"formats"`
	RequestedFormats []YtDlpFormat `json:"requested_formats"`

	Duration     int `json:"duration"`
	ViewCount    int `json:"view_count"`
	LikeCount    int `json:"like_count"`
	CommentCount int `json:"comment_count"`

	Channel              string `json:"channel"`
	ChannelUrl           string `json:"channel_url"`
	ChannelFollowerCount int    `json:"channel_follower_count,omitempty"`

	Uploader    string `json:"uploader"`
	UploaderUrl string `json:"uploader_url"`
}

type YtDlpFormat struct {
	FormatId string `json:"format_id"`
	Url      string `json:"url"`
	FileSize int    `json:"filesize"`

	Ext    string `json:"ext,omitempty"`
	Height int    `json:"height"`
	Width  int    `json:"width"`

	Cookies     string            `json:"cookies,omitempty"`
	HttpHeaders map[string]string `json:"http_headers,omitempty"`
}
