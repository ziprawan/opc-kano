package instagram

import (
	"encoding/json"
	"fmt"
)

type igResponseGeneral struct {
	Data struct {
		Media struct {
			Typename             string            `json:"__typename"`
			Owner                igResponse__Owner `json:"owner"`
			AccessibilityCaption *string           `json:"accessibility_caption"`
			Comments             igResponse__Count `json:"edge_media_preview_comment"`
			Likes                igResponse__Count `json:"edge_media_preview_like"`

			Caption igResponse__Edges[igResponse__Node[igResponse__Caption]] `json:"edge_media_to_caption"`
			Tagged  igResponse__Edges[igResponse__Node[igResponse__Tagged]]  `json:"edge_media_to_tagged_user"`
		} `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type igResponse__Edges[T any] struct {
	Edges []T `json:"edges"`
}

type igResponse__Node[T comparable] struct {
	Node T `json:"node"`
}

type igResponse__Tagged struct {
	User igResponse__User `json:"user"`
}

type igResponse__Caption struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text"`
}

type igResponse__User struct {
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	IsVerified bool   `json:"is_verified"`
}

type igResponse__Owner struct {
	igResponse__User

	Timelines igResponse__Count `json:"edge_owner_to_timeline_media"`
	Followers igResponse__Count `json:"edge_followed_by"`
}

type igResponse__Count struct {
	Count int `json:"count"`
}

type igResponse__GraphSidecar struct {
	Data struct {
		Media igGraph_Sidecar `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type igGraph_Sidecar struct {
	Childs igResponse__Edges[igResponse__Node[igGraph_SidecarChild]] `json:"edge_sidecar_to_children"`
}

type igGraph_SidecarChild struct {
	Image *igGraph_Image
	Video *igGraph_Video
}

func (ch *igGraph_SidecarChild) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch t := raw["__typename"]; t {
	case "XDTGraphImage":
		var i igGraph_Image
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		ch.Image = &i
	case "XDTGraphVideo":
		var v igGraph_Video
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		ch.Video = &v
	default:
		return fmt.Errorf("unexpected graphql __typename %q", t)
	}

	return nil
}

type igResponse__GraphVideo struct {
	Data struct {
		Media igGraph_Video `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type igGraph_Video struct {
	Dimensions igParsed_Dimensions `json:"dimensions"`
	VideoUrl   string              `json:"video_url"`
	Duration   float64             `json:"video_duration"`

	Views int `json:"video_view_count"`
	Plays int `json:"video_play_count"`

	DashInfo struct {
		Eligible bool   `json:"is_dash_eligible"`
		Manifest string `json:"video_dash_manifest"`
		Total    int    `json:"number_of_qualities"`
	} `json:"dash_info"`

	MusicAttribution *igGraph_MusicAttribution `json:"clips_music_attribution_info"`
}

type igGraph_MusicAttribution struct {
	Artist      string `json:"artist_name"`
	SongName    string `json:"song_name"`
	IsOriginal  bool   `json:"uses_original_audio"`
	Muted       bool   `json:"should_mute_audio"`
	MutedReason string `json:"should_mute_audio_reason"`
	AudioId     string `json:"audio_id"`
}

type igResponse__GraphImage struct {
	Data struct {
		Media igGraph_Image `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type igGraph_Image struct {
	Dimensions       igParsed_Dimensions `json:"dimensions"`
	DisplayUrl       string              `json:"display_url"`
	DisplayResources []struct {
		igParsed_Dimensions

		Src string `json:"src"`
	} `json:"display_resources"`
}
