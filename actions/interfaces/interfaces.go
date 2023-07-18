package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type Action interface {
	Apply(chatID telego.ChatID, messageIDs []int64, userID int64) error
	ApplyToMessage(message telego.Message) error
}

type InitFunc func(*zap.Logger, *telego.Bot, map[string]interface{}) (Action, error)

type HelpFunc func() string
