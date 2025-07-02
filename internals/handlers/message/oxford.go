package message

import (
	"fmt"
	"kano/internals/utils/kanoutils"
	"strings"
)

var OxfordMan = CommandMan{
	Name:     "oxford - Definisi kata (english) dari kamus Oxford",
	Synopsis: []string{".oxford KATA ..."},
	Description: []string{
		"Mencari definisi dari sebuah atau lebih kata dari kamus Oxford (SEE ALSO nomor 1). Hasil yang akan dikirim ada infleksinya, bagaimana cara mengucapkannya (disertai link audio), definisi (serta definisi singkatnya), dan catatan dari sang penerjemah.",
		"*KATA* (Wajib)\n{SPACE}Kata atau dua lebih kata yang ingin dicari.",
	},

	SeeAlso: []SeeAlso{
		{Content: "https://dict-api.com/api/od/", Type: SeeAlsoTypeExternalLink},
	},
	Source: "oxford.go",
}

func OxfordHandler(ctx *MessageContext) {
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Give a word to find the definition at Oxford University Press", true)
		return
	}

	word := ctx.Parser.Text[args[0].Start:]
	res, err := kanoutils.FindDefinition(word)
	if err != nil {
		ctx.Instance.Reply(err.Error(), true)
		return
	}

	if res.LastUpdated == 0 || len(res.Results) == 0 {
		ctx.Instance.Reply("Server returned a null data", true)
		return
	}

	msg := ""

	for resultIdx, result := range res.Results {
		msg += fmt.Sprintf("*Result #%d: %s (%s)*\n", resultIdx+1, result.Word, result.Type)

		for lexicalIdx, lexicalEntry := range result.LexicalEntries {
			for entryIdx, entry := range lexicalEntry.Entries {
				msg += fmt.Sprintf("*Lexical #%d - Entry #%d*\n", lexicalIdx+1, entryIdx+1)

				if len(entry.Etymologies) > 0 {
					msg += fmt.Sprintf("Etymology:\n> %s\n", strings.Join(entry.Etymologies, "; "))
				}

				if len(entry.Inflections) > 0 {
					var inflections []string
					for _, inflect := range entry.Inflections {
						inflections = append(inflections, inflect.InflectedForm)
					}
					msg += fmt.Sprintf("Inflections: _%s_\n", strings.Join(inflections, "; "))
				}

				if len(entry.Pronunciations) > 0 {
					msg += "Pronunciations:\n"
					for _, pronun := range entry.Pronunciations {
						msg += fmt.Sprintf("- %s (%s)\n", pronun.PhoneticSpelling, pronun.AudioFile)
					}
				}

				if len(entry.Senses) > 0 {
					for senseIdx, sense := range entry.Senses {
						msg += fmt.Sprintf("*Sense #%d*\n", senseIdx+1)

						if len(sense.Definitions) > 0 {
							msg += fmt.Sprintf("*Definition*: _%s_\n", strings.Join(sense.Definitions, "; "))
						}

						if len(sense.ShortDefinitions) > 0 {
							msg += fmt.Sprintf("*Short definition*: _%s_\n", strings.Join(sense.ShortDefinitions, "; "))
						}

						if len(sense.Examples) > 0 {
							msg += "*Examples:*\n"
							for _, exm := range sense.Examples {
								msg += fmt.Sprintf("- %s\n", exm.Text)
							}
						}

						if len(sense.Registers) > 0 {
							var registers []string
							for _, reg := range sense.Registers {
								registers = append(registers, reg.Text)
							}
							msg += fmt.Sprintf("_Registered as: %s_\n", strings.Join(registers, "; "))
						}

					}
				}

				if len(entry.Notes) > 0 {
					var notes []string
					for _, note := range entry.Notes {
						notes = append(notes, note.Text)
					}
					msg += fmt.Sprintf("Notes: \n> %s\n", strings.Join(notes, "; "))
				}

				msg += "\n"
			}
		}

		msg += "\n\n"
	}

	ctx.Instance.Reply(strings.TrimSpace(msg), true)
}
