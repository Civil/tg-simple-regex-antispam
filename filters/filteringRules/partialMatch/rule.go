package partialMatch

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	partialMatch  string
	caseSensitive bool
	isFinal       bool
}

func (r *Filter) Score(msg telego.Message) int {
	if strings.Contains(msg.Caption, r.partialMatch) || strings.Contains(msg.Text, r.partialMatch) {
		return 100
	}
	return 0
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "partialMatch"
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

func New(config map[string]interface{}) (interfaces.FilteringRule, error) {
	filterI, ok := config["match"]
	if !ok {
		return nil, fmt.Errorf("partialMatch filter requires `match` configuration parameter")
	}
	filter, ok := filterI.(string)
	if !ok {
		return nil, fmt.Errorf("filter is not a string")
	}
	if filter == "" {
		return nil, fmt.Errorf("`match` cannot be empty")
	}

	isFinal, err := config2.GetOptionBool(config, "isFinal")
	if err != nil {
		return nil, err
	}

	caseSensitive := false
	caseSensitiveI, ok := config["case_sensitive"]
	if ok {
		caseSensitive, ok = caseSensitiveI.(bool)
		if !ok {
			return nil, fmt.Errorf("case_sensitive is not a bool")
		}
	}

	return &Filter{
		partialMatch:  filter,
		caseSensitive: caseSensitive,
		isFinal:       isFinal,
	}, nil
}

func Help() string {
	return "partialMatch requires `match` parameter"
}
