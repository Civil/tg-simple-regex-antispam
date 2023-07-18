package actions

import (
	"fmt"

	"github.com/Civil/tg-simple-regex-antispam/actions/addReportButton"
	"github.com/Civil/tg-simple-regex-antispam/actions/deleteAndBan"
	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
)

var supportedActions map[string]interfaces.InitFunc
var supportedActionsHelp map[string]interfaces.HelpFunc

func init() {
	supportedActions["deleteAndBan"] = deleteAndBan.New
	supportedActionsHelp["deleteAndBan"] = deleteAndBan.Help

	supportedActions["addReportButton"] = addReportButton.New
	supportedActionsHelp["addReportButton"] = addReportButton.Help
}

func GetAction(name string) (interfaces.InitFunc, error) {
	action, ok := supportedActions[name]
	if !ok {
		return nil, fmt.Errorf("unknown action: %v", name)
	}
	return action, nil
}

func GetActions() map[string]interfaces.InitFunc {
	return supportedActions
}

func GetHelp() map[string]interfaces.HelpFunc {
	return supportedActionsHelp
}
