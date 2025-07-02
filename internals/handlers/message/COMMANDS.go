package message

import (
	"bytes"
	"context"
	"fmt"
	"kano/internals/utils/kanoutils"
	"os"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type MessageHandlerFunc func(ctx *MessageContext)

// Inspired by man page :D
type CommandMan struct {
	Name        string
	Synopsis    []string
	Description []string

	Source  string
	SeeAlso []SeeAlso
}

type SeeAlsoType string

const (
	SeeAlsoTypeCommand      SeeAlsoType = "command"
	SeeAlsoTypeExternalLink SeeAlsoType = "external_link"
)

type SeeAlso struct {
	Content string
	Type    SeeAlsoType
}

type MessageHandler struct {
	Func    MessageHandlerFunc
	Aliases []string
	Man     CommandMan
}

type MessageHandlerFuncMap map[string]MessageHandler

var MESSAGE_HANDLERS MessageHandlerFuncMap = MessageHandlerFuncMap{
	"ping": MessageHandler{
		Func:    PingHandler,
		Man:     PingMan,
		Aliases: []string{"p"},
	},
	"vo": MessageHandler{
		Func: VoHandler,
		Man:  VoMan,
	},
	"confess": MessageHandler{
		Func:    ConfessHandler,
		Man:     ConfessMan,
		Aliases: []string{"fess"},
	},
	"stk": MessageHandler{
		Func:    StkHandler,
		Man:     StkMan,
		Aliases: []string{"s"},
	},
	"stkinfo": MessageHandler{
		Func: StkInfoHandler,
		Man:  StkInfoMan,
	},
	"confesstarget": MessageHandler{
		Func:    ConfessTargetHandler,
		Man:     ConfessTargetMan,
		Aliases: []string{"fesstarget", "targetconfess", "ct"},
	},
	"login": MessageHandler{
		Func: LoginHandler,
		Man:  LoginMan,
	},
	"hsr": MessageHandler{
		Func: HSRHandler,
		Man:  HSRMan,
	},
	"gi": MessageHandler{
		Func: GIHandler,
		Man:  GIMan,
	},
	"zzz": MessageHandler{
		Func: ZZZHandler,
		Man:  ZZZMan,
	},
	"nim": MessageHandler{
		Func: NIMHandler,
		Man:  NimMan,
	},
	"pddikti": MessageHandler{
		Func:    PDDIKTIHandler,
		Man:     DiddyMan,
		Aliases: []string{"diddy"},
	},
	"shipping": MessageHandler{
		Func: ShippingHandler,
		Man:  ShippingMan,
	},
	"wordle": MessageHandler{
		Func:    WorldeHandler,
		Man:     WordleMan,
		Aliases: []string{"worlde", "word"},
	},
	"json": MessageHandler{
		Func: JSONHandler,
		Man:  JsonMan,
	},
	"oxford": MessageHandler{
		Func: OxfordHandler,
		Man:  OxfordMan,
	},
	"leaderboard": MessageHandler{
		Func:    LeaderboardHandler,
		Man:     LeaderboardMan,
		Aliases: []string{"lb"},
	},
	"setname": MessageHandler{
		Func:    SetNameHandler,
		Man:     SetnameMan,
		Aliases: []string{"name"},
	},
	"test": MessageHandler{
		Man: CommandMan{
			Name:     "test - Test",
			Synopsis: []string{"test"},
			Description: []string{
				"Buat ngetes.",
				"Argumen mungkin diperlukan dalam beberapa kasus.",
			},
			SeeAlso: []SeeAlso{},
			Source:  "COMMANDS.go",
		},
		Func: func(ctx *MessageContext) {
			vidBytes, err := os.ReadFile("assets/birdbrain.mp4")
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Errored when reading assets/birdbrain.mp4: %s", err), true)
				return
			}
			first := time.Now().UnixMilli()

			reader := bytes.NewReader(vidBytes)
			fmt.Println(reader.Size())
			buf := bytes.NewBuffer(nil)
			cmd := ffmpeg.Input("pipe:0", ffmpeg.KwArgs{
				"analyzeduration": "100M",
				"probesize":       "100M",
			}).
				Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", 1)}).
				Output("pipe:1", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg", "pix_fmt": "yuvj420p"}).
				WithInput(reader).WithOutput(buf, os.Stdout)

			fmt.Println("Executing", cmd.Compile())
			err = cmd.Run()
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Errored when getting the first frame of the video: %s", err), true)
				return
			}

			frameInfo, err := kanoutils.GenerateImageInfo(buf.Bytes())
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Errored when generating image info: %s", err), true)
				return
			}

			resp, err := ctx.Instance.Upload(vidBytes, whatsmeow.MediaVideo)
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Errored when uploading the media: %s", err), true)
				return
			}
			last := time.Now().UnixMilli()

			fmt.Printf("Response: %+v\n", resp)

			vidMsg := &waE2E.VideoMessage{
				URL:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				Caption:       proto.String(fmt.Sprintf("Upload time: %d ms", last-first)),
				Mimetype:      proto.String("video/mp4"),

				Width:         &frameInfo.Width,
				Height:        &frameInfo.Height,
				JPEGThumbnail: frameInfo.JPEGThumbnail,
			}

			targetJID := ctx.Instance.ChatJID()
			fmt.Printf("Vidmsg: %+v\nTarget JID: %s", vidMsg, targetJID.String())

			resp2, err := ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), &waE2E.Message{
				VideoMessage: vidMsg,
			})
			if err != nil {
				fmt.Println("Error nich")
			} else {
				fmt.Printf("%+v\n", resp2)
			}
		},
	},
}
var mappedCommands map[string]MessageHandler = map[string]MessageHandler{}

func (ctx *MessageContext) Handle() {
	cmd := ctx.Parser.GetCommand()

	if len(mappedCommands) < len(MESSAGE_HANDLERS) {
		for key, val := range MESSAGE_HANDLERS {
			mappedCommands[key] = val
			for _, alias := range val.Aliases {
				mappedCommands[alias] = val
			}
		}
	}

	detectedFunc, ok := mappedCommands[cmd.Command]
	if ok {
		detectedFunc.Func(ctx)
	}
}
