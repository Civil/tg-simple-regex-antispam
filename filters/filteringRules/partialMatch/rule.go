package partialMatch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

var (
	ErrRequiresMatchParam = errors.New(
		"partialMatch filter requires `match` configuration parameter",
	)
	ErrMatchEmpty           = errors.New("`match` cannot be empty")
	ErrCaseSensitiveNotBool = errors.New("case_sensitive is not a bool")
)

type Filter struct {
	logger        *zap.Logger
	chainName     string
	partialMatch  string
	caseSensitive bool
	isFinal       bool
}

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "partialMatch"))
	filter, err := config2.GetOptionString(config, "match")
	if err != nil {
		return nil, ErrRequiresMatchParam
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
		logger:        logger,
		chainName:     chainName,
		partialMatch:  filter,
		caseSensitive: caseSensitive,
		isFinal:       isFinal,
	}, nil
}

func Help() string {
	return "partialMatch requires `match` parameter"
}

func (r *Filter) Score(_ *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	res := &scoringResult.ScoringResult{}
	if strings.Contains(msg.Caption, r.partialMatch) || strings.Contains(msg.Text, r.partialMatch) {
		res.Reason = fmt.Sprintf("Partial match found: %s", r.partialMatch)
		res.Score = 100
	}
	return res
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

func (r *Filter) TGAdminPrefix() string {
	return ""
}

func (r *Filter) HandleTGCommands(_ *zap.Logger, _ *telego.Bot, _ *telego.Message, _ []string) error {
	return nil
}
