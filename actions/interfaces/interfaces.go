package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
)

type Action interface {
	Apply(callbackStatefulFilter interfaces.StatefulFilter, score *scoringResult.ScoringResult, chatID telego.ChatID,
		messageIDs []int64, userID int64, extraParams any) error
	ApplyToMessage(callbackStatefulFilter interfaces.StatefulFilter, score *scoringResult.ScoringResult,
		message *telego.Message, extraParams any) error
	GetName() string
	PerMessage() bool
}

type InitFunc func(*zap.Logger, *telego.Bot, map[string]any) (Action, error)

type HelpFunc func() string
