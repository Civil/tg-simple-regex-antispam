package filters

import (
	"errors"

	"github.com/Civil/tg-simple-regex-antispam/filters/chains/checkNevents"
	"github.com/Civil/tg-simple-regex-antispam/filters/chains/report"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/hasEmoji"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/hasLinks"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/isForward"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/isInSpam"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/isRepeatedMessage"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/isStory"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/partialMatch"
	"github.com/Civil/tg-simple-regex-antispam/filters/filteringRules/regex"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types"
)

var (
	supportedFilteringRules = map[string]interfaces.InitFunc{
		"regex":             regex.New,
		"partialMatch":      partialMatch.New,
		"isForward":         isForward.New,
		"hasEmoji":          hasEmoji.New,
		"hasLinks":          hasLinks.New,
		"isInSpam":          isInSpam.New,
		"isRepeatedMessage": isRepeatedMessage.New,
		"isStory":           isStory.New,
	}
	supportedFilteringRulesHelp = map[string]interfaces.HelpFunc{
		"regex":             regex.Help,
		"partialMatch":      partialMatch.Help,
		"isForward":         isForward.Help,
		"hasEmoji":          hasEmoji.Help,
		"hasLinks":          hasLinks.Help,
		"isInSpam":          isInSpam.Help,
		"isRepeatedMessage": isRepeatedMessage.Help,
		"isStory":           isStory.Help,
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
