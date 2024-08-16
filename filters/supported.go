package filters

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/partialMatch"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/regex"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/statefulFilters/checkNevents"
)

var (
	supportedFilteringRules     map[string]interfaces.InitFunc
	supportedFilteringRulesHelp map[string]interfaces.HelpFunc
)

var (
	supportedStatefulFilters     map[string]interfaces.StatefulInitFunc
	supportedStatefulFiltersHelp map[string]interfaces.HelpFunc
)

func init() {
	// Stateless filters
	supportedFilteringRules["regexp"] = regex.New
	supportedFilteringRulesHelp["regexp"] = regex.Help

	supportedFilteringRules["partialMatch"] = partialMatch.New
	supportedFilteringRulesHelp["regex"] = partialMatch.Help

	// Stateful filters
	supportedStatefulFilters["checkNevents"] = checkNevents.New
	supportedStatefulFiltersHelp["checkNevents"] = checkNevents.Help
}

var ErrUknownStatefulFilter = errors.New("unknown stateful filter")

func GetStatefulFilter(name string) (interfaces.StatefulInitFunc, error) {
	initFunc, ok := supportedStatefulFilters[name]
	if !ok {
		return nil, ErrUknownStatefulFilter
	}
	return initFunc, nil
}

func GetStatefulFilters() map[string]interfaces.StatefulInitFunc {
	return supportedStatefulFilters
}

func GetStatefulFiltersHelp() map[string]interfaces.HelpFunc {
	return supportedStatefulFiltersHelp
}

func GetFilteringRules() map[string]interfaces.InitFunc {
	return supportedFilteringRules
}

func GetFilteringRulesHelp() map[string]interfaces.HelpFunc {
	return supportedFilteringRulesHelp
}
