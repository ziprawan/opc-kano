package message

import (
	"kano/internals/utils/messageutils"
	"kano/internals/utils/parser"
)

type MessageContext struct {
	Instance *messageutils.Message
	Parser   *parser.Parser
}

func InitHandlerContext(instance *messageutils.Message, parser *parser.Parser) *MessageContext {
	return &MessageContext{
		Instance: instance,
		Parser:   parser,
	}
}
