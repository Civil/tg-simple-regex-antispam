package filters

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/partialMatch"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/regex"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/statefulFilters/checkNevents"
)

var (
	supportedFilteringRules = map[string]interfaces.InitFunc{
		"regexp":       regex.New,
		"partialMatch": partialMatch.New,
	}
	supportedFilteringRulesHelp = map[string]interfaces.HelpFunc{
		"regexp": regex.Help,
		"regex":  partialMatch.Help,
	}
)

var (
	supportedStatefulFilters = map[string]interfaces.StatefulInitFunc{
		"checkNevents": checkNevents.New,
	}
	supportedStatefulFiltersHelp = map[string]interfaces.HelpFunc{
		"checkNevents": checkNevents.Help,
	}
)

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
