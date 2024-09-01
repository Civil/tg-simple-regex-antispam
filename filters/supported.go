package filters

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/isForward"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/partialMatch"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/regex"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/statefulFilters/checkNevents"
	"github.com/Civil/tg-simple-regex-antispam/filters/statefulFilters/report"
	"github.com/Civil/tg-simple-regex-antispam/filters/types"
)

var (
	supportedFilteringRules = map[string]interfaces.InitFunc{
		"regex":        regex.New,
		"partialMatch": partialMatch.New,
		"isForward":    isForward.New,
	}
	supportedFilteringRulesHelp = map[string]interfaces.HelpFunc{
		"regex":        regex.Help,
		"partialMatch": partialMatch.Help,
		"isForward":    isForward.Help,
	}
)

var (
	supportedStatefulFilters = map[string]types.StatefulInitFunc{
		"checkNevents": checkNevents.New,
		"report":       report.New,
	}
	supportedStatefulFiltersHelp = map[string]interfaces.HelpFunc{
		"checkNevents": checkNevents.Help,
		"report":       report.Help,
	}
)

var ErrUknownStatefulFilter = errors.New("unknown stateful filter")

func GetStatefulFilter(name string) (types.StatefulInitFunc, error) {
	initFunc, ok := supportedStatefulFilters[name]
	if !ok {
		return nil, ErrUknownStatefulFilter
	}
	return initFunc, nil
}

func GetStatefulFilters() map[string]types.StatefulInitFunc {
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
