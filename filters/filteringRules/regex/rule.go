package regex

import (
	"errors"
	"regexp"

	"github.com/mymmrac/telego"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	regex   *regexp.Regexp
	isFinal bool
}

func (r *Filter) Score(msg telego.Message) int {
	if r.regex.MatchString(msg.Caption) || r.regex.MatchString(msg.Text) {
		return 100
	}
	return 0
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "regex"
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

var (
	ErrRequiresRegexpParameter = errors.New(
		"regexp filter requires `regexp` parameter to work properly",
	)
	ErrFilterNotString = errors.New("filter is not a string")
	ErrRegexpEmpty     = errors.New("regexp cannot be empty")
)

func New(config map[string]any) (interfaces.FilteringRule, error) {
	filterI, ok := config["regex"]
	if !ok {
		return nil, ErrRequiresRegexpParameter
	}
	regex, ok := filterI.(string)
	if !ok {
		return nil, ErrFilterNotString
	}
	if regex == "" {
		return nil, ErrRegexpEmpty
	}

	isFinal, err := config2.GetOptionBool(config, "isFinal")
	if err != nil {
		return nil, err
	}

	res := Filter{
		isFinal: isFinal,
	}

	res.regex, err = regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func Help() string {
	return "regexp requires `regexp` parameter"
}
