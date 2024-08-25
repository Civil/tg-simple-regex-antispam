package interfaces

import (
	"github.com/mymmrac/telego"

	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type FilteringRule interface {
	Score(*telego.Message) int
	IsStateful() bool
	GetName() string
	IsFinal() bool
}

type InitFunc func(map[string]any) (FilteringRule, error)

type HelpFunc func() string

type StatefulFilter interface {
	stateful.Stateful
	FilteringRule
	RemoveState(userID int64) error
}
