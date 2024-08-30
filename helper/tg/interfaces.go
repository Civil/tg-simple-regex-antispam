package tg

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type AdminCMDHandlerFunc func(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error

type AdminCMDHelpFunc func() string
