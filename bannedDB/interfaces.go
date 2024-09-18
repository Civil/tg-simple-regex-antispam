package bannedDB

import (
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type BanDB interface {
	stateful.Stateful
	BanUser(userID int64) error
	UnbanUser(userID int64) error
	IsBanned(userID int64) bool
	ListUserIDs() ([]int64, error)
	SetStatefulFilters(filters []interfaces.StatefulFilter)
}
