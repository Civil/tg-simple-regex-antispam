package partialMatch

import (
	"errors"
	"strings"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	partialMatch  string
	caseSensitive bool
	isFinal       bool
}

func (r *Filter) Score(msg *telego.Message) int {
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

func (r *Filter) GetFilterName() string {
	return ""
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

var (
	ErrRequiresMatchParam = errors.New(
		"partialMatch filter requires `match` configuration parameter",
	)
	ErrFilterNotString      = errors.New("filter is not a string")
	ErrMatchEmpty           = errors.New("`match` cannot be empty")
	ErrCaseSensitiveNotBool = errors.New("case_sensitive is not a bool")
)

func New(config map[string]any) (interfaces.FilteringRule, error) {
	filterI, ok := config["match"]
	if !ok {
		return nil, ErrRequiresMatchParam
	}
	filter, ok := filterI.(string)
	if !ok {
		return nil, ErrFilterNotString
	}
	if filter == "" {
		return nil, ErrMatchEmpty
	}

	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	caseSensitive := false
	caseSensitiveI, ok := config["case_sensitive"]
	if ok {
		caseSensitive, ok = caseSensitiveI.(bool)
		if !ok {
			return nil, ErrCaseSensitiveNotBool
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

func (r *Filter) TGAdminPrefix() string {
	return ""
}

func (r *Filter) HandleTGCommands(_ *zap.Logger, _ *telego.Bot, _ *telego.Message, _ []string) error {
	return nil
}
