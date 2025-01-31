package tg

import (
	"bytes"
	"errors"
	"strings"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type AdminCMDHandlerFunc func(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error

type AdminCMDHelpFunc func() string

type TGHaveAdminCommands struct {
	Handlers map[string]AdminCMDHandlerFunc
}

var (
	ErrCommandArgsInvalid = errors.New("not enough arguments for command")
)

func (r *TGHaveAdminCommands) HandleTGCommands(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	supportedCommands := make([]string, 0, len(r.Handlers))
	for cmd := range r.Handlers {
		supportedCommands = append(supportedCommands, cmd)
	}
	logger.Debug("handling tg command", zap.Any("message", message), zap.Strings("supported_commands", supportedCommands))
	if tokens == nil {
		logger.Warn("empty tokens for tg command")
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("Sub-command was not specified.\nAvailable subcommands:\n\n")
		for _, prefix := range supportedCommands {
			buf.WriteString("   " + prefix + "\n")
		}
		_ = SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
		return ErrCommandArgsInvalid
	}
	logger.Debug("handling tg command", zap.String("command", tokens[0]))
	if h, ok := r.Handlers[strings.ToLower(tokens[0])]; ok {
		if len(tokens) > 1 {
			return h(logger, bot, message, tokens[1:])
		} else {
			return h(logger, bot, message, nil)
		}
	}
	logger.Warn("unsupported command", zap.Strings("tokens", tokens), zap.Strings("supported_commands", supportedCommands))
	return stateful.ErrNotSupported
}
