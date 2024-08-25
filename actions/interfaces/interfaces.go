package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
)

type Action interface {
	Apply(callbackStatefulFilter interfaces.StatefulFilter, chatID telego.ChatID, messageIDs []int64, userID int64) error
	ApplyToMessage(message telego.Message) error
}

type InitFunc func(*zap.Logger, *telego.Bot, map[string]any) (Action, error)

type HelpFunc func() string
