package types

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	actions "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
)

type StatefulInitFunc func(*zap.Logger, string, bannedDB.BanDB, *telego.Bot, map[string]any, []interfaces.FilteringRule, []actions.Action) (interfaces.StatefulFilter,
	error)
