package hasEmoji

import (
	"errors"
	"fmt"

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
	numEmojis int

	linkTypes map[string]bool
}

var ErrThresholdTooLow = errors.New("threshold for number of emojis must be positive")

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "hasEmoji"))
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	numEmojis, err := config2.GetOptionIntWithDefault(config, "numEmojis", 7)
	if err != nil {
		return nil, err
	}

	if numEmojis <= 0 {
		return nil, ErrThresholdTooLow
	}

	return &Filter{
		logger:    logger,
		chainName: chainName,
		isFinal:   isFinal,
		numEmojis: numEmojis,

		linkTypes: map[string]bool{
			"custom_emoji": true,
		},
	}, nil
}

func Help() string {
	return "hasEmoji checks if the message has too much emoji or stickers"
}

func (r *Filter) Score(_ *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	res := &scoringResult.ScoringResult{}
	var linkCount int
	for _, entity := range msg.Entities {
		if _, ok := r.linkTypes[entity.Type]; ok {
			linkCount++
		}
	}

	if linkCount >= r.numEmojis {
		res.Reason = fmt.Sprintf("found %d emoji/stickers, which is more than threshold of %d.", linkCount, r.numEmojis)
		res.Score = 100
	}
	return res
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "hasEmoji"
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
