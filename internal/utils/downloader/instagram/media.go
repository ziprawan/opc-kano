package instagram

import (
	"encoding/json"
	"fmt"
	"strings"
)

type igParsed struct {
	caption string
	medias  []igParsed_Media
}

type igParsed_Media struct {
	Url        string
	Dimensions igParsed_Dimensions

	IsVideo  bool
	Duration float64

	// Special case where HD video is detected
	VideoId  string
	AudioId  string
	AudioUrl string
}

type igParsed_Dimensions struct {
	Height int
	Width  int
}

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

func parseGeneralInfo(data []byte) (igParsed, error) {
	var general igResponseGeneral
	err := json.Unmarshal(data, &general)
	if err != nil {
		return igParsed{}, fmt.Errorf("failed to get graphql general info: %s", err)
	}

	// Parse the common fields
	res := igParsed{}
	var caption strings.Builder

	media := general.Data.Media
	if cap := media.Caption.Edges; len(cap) > 0 {
		fmt.Fprintf(&caption, "%s", cap[0].Node.Text)
	}

	if accCap := media.AccessibilityCaption; accCap != nil {
		fmt.Fprintf(&caption, "\n\n%s", *accCap)
	}

	owner := media.Owner
	fmt.Fprintf(&caption, "\n\nUploaded by @%s (%s) - %d 🎥 %d 🫂", owner.Username, owner.FullName, owner.Timelines.Count, owner.Followers.Count)

	if tagged := media.Tagged.Edges; len(tagged) > 0 {
		fmt.Fprint(&caption, "\nAdditional user(s):")
		for _, tag := range tagged {
			user := tag.Node.User
			fmt.Fprintf(&caption, "\n- @%s (%s)", user.Username, user.FullName)
		}
	}

	fmt.Fprintf(&caption, "\n\n")

	// Parse specific fields
	switch t := general.Data.Media.Typename; t {
	case "XDTGraphSidecar":
		var sidecar igResponse__GraphSidecar
		err := json.Unmarshal(data, &sidecar)
		if err != nil {
			return res, fmt.Errorf("failed to parse XDTGraphSidecar object: %s", err)
		}

		childs := sidecar.Data.Media.Childs.Edges
		fmt.Println("found", len(childs), "sidecar childs")
		for _, child := range childs {
			node := child.Node
			if node.Image != nil {
				addImage(&res, *node.Image)
			} else if node.Video != nil {
				addVideo(&res, *node.Video)
			} else {
				return res, fmt.Errorf("XDTGraphSidecar childrens doesn't contains XDTGraphVideo or XDTGraphImage")
			}
		}

	case "XDTGraphImage":
		var image igResponse__GraphImage
		err := json.Unmarshal(data, &image)
		if err != nil {
			return res, fmt.Errorf("failed to parse XDTGraphImage object: %s", err)
		}

		media := image.Data.Media
		addImage(&res, media)

	case "XDTGraphVideo":
		var video igResponse__GraphVideo
		err := json.Unmarshal(data, &video)
		if err != nil {
			return res, fmt.Errorf("failed to parse XDTGraphVideo object: %s", err)
		}

		media := video.Data.Media
		addVideo(&res, media)

		if media.MusicAttribution != nil {
			attr := *media.MusicAttribution
			if attr.IsOriginal {
				fmt.Fprint(&caption, "Uses original audio")
			} else {
				fmt.Fprintf(&caption, "Used audio: %q by %q", attr.SongName, attr.Artist)
			}

			fmt.Fprintf(&caption, " (https://www.instagram.com/reels/audio/%s)\n", attr.AudioId)
			if attr.Muted {
				fmt.Fprintf(&caption, "_Audio is muted (%s)_\n", attr.MutedReason)
			}

			fmt.Fprint(&caption, "\n")
		}

		fmt.Fprintf(&caption, "%d 👀 %d ▶️ ", media.Views, media.Plays)

	default:
		return res, fmt.Errorf("unexpected graphql __typename %q", t)
	}

	fmt.Fprintf(&caption, "%d ❤️ %d 💬", media.Likes.Count, media.Comments.Count)
	res.caption = strings.TrimSpace(caption.String())

	return res, nil
}

func addVideo(res *igParsed, media igGraph_Video) {
	var parsedMedia igParsed_Media

	parsedMedia.IsVideo = true
	parsedMedia.Duration = media.Duration
	parsedMedia.Dimensions = media.Dimensions
	parsedMedia.Url = media.VideoUrl

	// Dash manifest offers more and higher qualities
	if dash := media.DashInfo; dash.Eligible {
		dash := parseDashInfo(dash.Manifest)
		fps := 0.0

		for _, video := range dash.Video {
			if (video.Height > parsedMedia.Dimensions.Height || video.Width > parsedMedia.Dimensions.Width) && video.FrameRate > fps {
				parsedMedia.Dimensions = igParsed_Dimensions{video.Height, video.Width}
				parsedMedia.Url = video.Url

				fps = video.FrameRate
				parsedMedia.VideoId = video.Id
			}
		}

		if fps > 0 { // I assume that there is a video that is more HD than the video_url
			bitrate := 0

			// There might be a reel that doesn't have audio
			// So I don't need to throw an error if audio is not found
			for _, audio := range dash.Audio {
				if audio.Bitrate > bitrate {
					bitrate = audio.Bitrate
					parsedMedia.AudioId = audio.Id
					parsedMedia.AudioUrl = audio.Url
				}
			}
		}
	}

	res.medias = append(res.medias, parsedMedia)
}

func addImage(res *igParsed, media igGraph_Image) {
	var parsedMedia igParsed_Media

	parsedMedia.Url = media.DisplayUrl
	parsedMedia.Dimensions = media.Dimensions

	for _, rsrc := range media.DisplayResources {
		if rsrc.Height > parsedMedia.Dimensions.Height || rsrc.Width > parsedMedia.Dimensions.Width {
			parsedMedia.Dimensions = igParsed_Dimensions{rsrc.Height, rsrc.Width}
			parsedMedia.Url = rsrc.Src
		}
	}

	res.medias = append(res.medias, parsedMedia)
}
