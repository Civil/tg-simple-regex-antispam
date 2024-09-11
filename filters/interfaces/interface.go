package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type FilteringRule interface {
	Score(*telego.Message) *scoringResult.ScoringResult
	IsStateful() bool
	GetName() string
	GetFilterName() string
	IsFinal() bool
	TGAdminPrefix() string
	HandleTGCommands(*zap.Logger, *telego.Bot, *telego.Message, []string) error
}

type InitFunc func(*zap.Logger, map[string]any, string) (FilteringRule, error)

type HelpFunc func() string

type StatefulFilter interface {
	stateful.Stateful
	FilteringRule
	RemoveState(int64) error
}
