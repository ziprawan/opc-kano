package message

import (
	"fmt"
	cronfinaid "kano/internals/handlers/cron/finaid"
	finaiditb "kano/internals/utils/finaid-itb"
	"kano/internals/utils/kanoutils"
	"math"
	"slices"
	"strconv"
	"strings"
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
	"down": MessageHandler{
		Func: DownloaderHandler,
		Man:  DownloaderMan,
	},
	"c1": MessageHandler{
		Func: func(ctx *MessageContext) {
			args := ctx.Parser.GetArgs()

			if len(args) == 0 {
				ctx.Instance.Reply("Berikan argumen", true)
				return
			}

			if len(args)%2 == 1 {
				ctx.Instance.Reply("Jumlah argumen tidak genap", true)
				return
			}

			nums := []uint64{}

			for i, arg := range args {
				n, err := strconv.ParseUint(arg.Content, 10, 0)
				if err != nil {
					ctx.Instance.Reply(fmt.Sprintf("Argumen ke %d bukan bilangan bulat positif", i), true)
					return
				}
				nums = append(nums, n)
			}

			var (
				l uint64 = 0
				r uint64 = 0

				mults   [2]float64 = [2]float64{math.Inf(-1), math.Inf(1)}
				predict [2]uint64  = [2]uint64{}
			)

			dbg := ""

			for i := range len(nums) / 2 {
				idx := i * 2

				l += nums[idx]
				r += nums[idx+1]

				var res_limit [2]float64 = [2]float64{float64(r), float64(r)}
				if r%2 == 0 {
					res_limit[0] -= 0.5
					res_limit[1] += 0.5
				} else {
					res_limit[0] -= 0.4
					res_limit[1] += 0.4
				}

				var cur_mults [2]float64 = [2]float64{res_limit[0] / float64(l), res_limit[1] / float64(l)}
				if cur_mults[0] > mults[0] {
					mults[0] = cur_mults[0]
				}
				if cur_mults[1] < mults[1] {
					mults[1] = cur_mults[1]
				}

				predict = [2]uint64{uint64(math.Ceil(100 / mults[1])), uint64(math.Ceil(100 / mults[0]))}

				dbg += fmt.Sprintf("Subject #%d => [%d expanded into %d]\n", i+1, l, r)
				dbg += fmt.Sprintf("Mults: %.3f => %.3f\n", mults[0], mults[1])
				dbg += fmt.Sprintf("Predicted: %d => %d\n\n", predict[0], predict[1])

				if predict[0] > predict[1] {
					ctx.Instance.Reply(dbg+"No solve.", true)
					return
				}
			}

			pred_list := []string{}
			for p := range predict[1] - predict[0] + 1 {
				pred_list = append(pred_list, strconv.FormatUint(predict[0]+p, 10))
			}

			ctx.Instance.Reply(dbg+fmt.Sprintf("Predicts: %s", strings.Join(pred_list, ", ")), true)
		},
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
		Func: func(ctx *MessageContext) {},
	},
	"latestbeasiswa": MessageHandler{
		Man: CommandMan{
			Name:        "latestbeasiswa - Beasiswa Terbaru dari Finaid ITB",
			Synopsis:    []string{"latestbeasiswa"},
			Description: []string{"-"},
			SeeAlso:     []SeeAlso{},
			Source:      "COMMANDS.go",
		},
		Func: func(ctx *MessageContext) {
			sc, err := finaiditb.FetchScholarships(1)
			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan: %s", err), true)
				return
			}

			if len(sc.Data) == 0 {
				ctx.Instance.Reply("No data.", true)
				return
			}
			data := sc.Data[0]

			msg := kanoutils.GenerateFinaidScholarshipMessage(data)

			ctx.Instance.Reply(msg, true)
		},
		Aliases: []string{"bea"},
	},
	"regis": MessageHandler{
		Func: func(ctx *MessageContext) {
			available := []string{"beasiswa"}

			cmd := ctx.Parser.GetCommand().Command
			args := ctx.Parser.GetArgs()
			if len(args) == 0 {
				ctx.Instance.Reply("Expected 1 argument", true)
				return
			}

			t := args[0].Content
			if !slices.Contains(available, t) {
				ctx.Instance.Reply(fmt.Sprintf("Allowed argument: %s", strings.Join(available, ", ")), true)
				return
			}

			var err error
			if t == "beasiswa" {
				if cmd == "regis" || cmd == "follow" {
					err = cronfinaid.RegisterNewJID(*ctx.Instance.ChatJID())
				} else {
					err = cronfinaid.UnregisterJID(*ctx.Instance.ChatJID())
				}

			}

			if err != nil {
				ctx.Instance.Reply(fmt.Sprintf("Err: %s", err), true)
			} else {
				ctx.Instance.Reply("success", true)
			}
		},
		Aliases: []string{"unregis", "follow", "unfollow"},
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
