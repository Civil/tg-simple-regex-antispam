package types

import (
	"go.uber.org/zap"

	actions "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
)

type StatefulInitFunc func(*zap.Logger, bannedDB.BanDB, map[string]any, []interfaces.FilteringRule, []actions.Action) (interfaces.StatefulFilter,
	error)
