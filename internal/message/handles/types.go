package handles

import "kano/internal/utils/messageutil"

type CommandHandlerFunc func(ctx *messageutil.MessageContext) error

type CommandHandler struct {
	Func    CommandHandlerFunc
	Aliases []string
	Man     CommandMan
}

type CommandMan struct {
	Name        string
	Synopsis    []string
	Description []string

	SourceFilename string
	SeeAlso        []SeeAlso
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
