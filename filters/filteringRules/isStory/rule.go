package isStory

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	logger    *zap.Logger
	chainName string
	isFinal   bool
}

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "isStory"))
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", true)
	if err != nil {
		return nil, err
	}

	return &Filter{
		logger:    logger,
		chainName: chainName,
		isFinal:   isFinal,
	}, nil
}

func Help() string {
	return "isForward checks if the message is forwarded"
}

func (r *Filter) Score(_ *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	res := &scoringResult.ScoringResult{}
	if msg.Story != nil && msg.Story.ID != 0 {
		res.Reason = "this message is a story"
		res.Score = 100
	}
	return res
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "isStory"
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
