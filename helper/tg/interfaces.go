package tg

import (
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
	if tokens == nil {
		logger.Error("empty tokens for tg command")
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
	supportedCommands := make([]string, 0, len(r.Handlers))
	for cmd := range r.Handlers {
		supportedCommands = append(supportedCommands, cmd)
	}
	logger.Warn("unsupported command", zap.Strings("tokens", tokens), zap.Strings("supported_commands", supportedCommands))
	return stateful.ErrNotSupported
}
