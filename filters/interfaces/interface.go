package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	actions "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type FilteringRule interface {
	Score(telego.Message) int
	IsStateful() bool
	GetName() string
	IsFinal() bool
}

type InitFunc func(map[string]interface{}) (FilteringRule, error)

type HelpFunc func() string

type StatefulFilter interface {
	stateful.Stateful
	FilteringRule
}

type StatefulInitFunc func(*zap.Logger, bannedDB.BanDB, map[string]interface{}, []FilteringRule, []actions.Action) (StatefulFilter,
	error)
