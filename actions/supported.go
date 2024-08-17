package actions

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/actions/addReportButton"
	"github.com/Civil/tg-simple-regex-antispam/actions/deleteAndBan"
	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
)

var (
	supportedActions = map[string]interfaces.InitFunc{
		"deleteAndBan":    deleteAndBan.New,
		"addReportButton": addReportButton.New,
	}
	supportedActionsHelp = map[string]interfaces.HelpFunc{
		"deleteAndBan":    deleteAndBan.Help,
		"addReportButton": addReportButton.Help,
	}
)

var ErrUknownAction = errors.New("unknown action")

func GetAction(name string) (interfaces.InitFunc, error) {
	action, ok := supportedActions[name]
	if !ok {
		return nil, ErrUknownAction
	}
	return action, nil
}

func GetActions() map[string]interfaces.InitFunc {
	return supportedActions
}

func GetHelp() map[string]interfaces.HelpFunc {
	return supportedActionsHelp
}
