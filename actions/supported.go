package actions

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/actions/addReportButton"
	"github.com/Civil/tg-simple-regex-antispam/actions/deleteAndBan"
	"github.com/Civil/tg-simple-regex-antispam/actions/forwardToChat"
	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/actions/storeAsSpam"
)

var (
	supportedActions = map[string]interfaces.InitFunc{
		"deleteAndBan":    deleteAndBan.New,
		"addReportButton": addReportButton.New,
		"forwardToChat":   forwardToChat.New,
		"storeAsSpam":     storeAsSpam.New,
	}
	supportedActionsHelp = map[string]interfaces.HelpFunc{
		"deleteAndBan":    deleteAndBan.Help,
		"addReportButton": addReportButton.Help,
		"forwardToChat":   forwardToChat.Help,
		"storeAsSpam":     storeAsSpam.Help,
	}
)

var ErrUnknownAction = errors.New("unknown action")

func GetAction(name string) (interfaces.InitFunc, error) {
	action, ok := supportedActions[name]
	if !ok {
		return nil, ErrUnknownAction
	}
	return action, nil
}

func GetActions() map[string]interfaces.InitFunc {
	return supportedActions
}

func GetHelp() map[string]interfaces.HelpFunc {
	return supportedActionsHelp
}
