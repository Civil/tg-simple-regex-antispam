package hasLinks

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
	logger           *zap.Logger
	chainName        string
	isFinal          bool
	numLinks         int
	mentionNotInChat bool

	linkTypes map[string]bool
}

var ErrThresholdTooLow = errors.New("threshold for number of links must be positive")

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "hasLinks"))
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	mentionNotInChat, err := config2.GetOptionBoolWithDefault(config, "mentionNotInChat", false)
	if err != nil {
		return nil, err
	}

	numLinks, err := config2.GetOptionIntWithDefault(config, "numLinks", 1)
	if err != nil {
		return nil, err
	}

	if numLinks <= 0 {
		return nil, ErrThresholdTooLow
	}

	return &Filter{
		logger:           logger,
		chainName:        chainName,
		isFinal:          isFinal,
		numLinks:         numLinks,
		mentionNotInChat: mentionNotInChat,

		linkTypes: map[string]bool{
			"url":          true,
			"text_link":    true,
			"mention":      true,
			"text_mention": true,
			"email":        true,
		},
	}, nil
}

func Help() string {
	return "hasLinks checks if the message have too many links/emails/mentions"
}

func (r *Filter) Score(bot *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Error("panic in f", zap.Any("panic", rec))
		}
	}()
	r.logger.Debug("checking hasLinks", zap.Int("numLinks", r.numLinks), zap.Bool("mentionNotInChat", r.mentionNotInChat), zap.Int("entities", len(msg.Entities)))
	res := &scoringResult.ScoringResult{}
	var linkCount int
	for _, entity := range msg.Entities {
		r.logger.Debug("checking entity", zap.Any("entity", entity))
		if _, ok := r.linkTypes[entity.Type]; ok {
			r.logger.Debug("entity type is in the list", zap.Any("entity", entity))
			if entity.Type == "mention" && r.mentionNotInChat {
				r.logger.Debug("checking mention", zap.Any("entity", entity))
				// mention starts with '@' and we don't need that.
				start := entity.Offset + 1
				end := entity.Offset + entity.Length - 1
				if start >= end {
					r.logger.Error("invalid mention entity", zap.Int("offset", entity.Offset), zap.Int("length", entity.Length))
					continue
				}
				username := msg.Text[entity.Offset+1 : entity.Offset+entity.Length-1]
				getChatParams := &telego.GetChatParams{
					ChatID: telego.ChatID{ID: msg.Chat.ID},
				}
				info, err := bot.GetChat(getChatParams)
				if err != nil {
					r.logger.Error("failed to get chat info", zap.Error(err))
					continue
				}
				r.logger.Debug("got chat info", zap.Any("chat_info", info))
				ok = false
				for _, member := range info.ActiveUsernames {
					if member == username {
						ok = true
					}
				}
				if !ok {
					res.Reason = fmt.Sprintf("mentioned user '%s' is not in the chat, which is not allowed.", username)
					res.Score = 100
					return res
				}
			}
			linkCount++
		}
	}

	if linkCount >= r.numLinks {
		res.Reason = fmt.Sprintf("found %d links, which is more or equal to the threshold %d.", linkCount, r.numLinks)
		res.Score = 100
	}
	return res
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "hasLinks"
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
