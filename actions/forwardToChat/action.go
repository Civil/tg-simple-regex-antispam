package forwardToChat

import (
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	forwardToChatID int64
}

func (r *Action) Apply(_ interfaces2.StatefulFilter, _ *scoringResult.ScoringResult, _ telego.ChatID, _ []int64, _ int64) error {
	return ErrNotSupported
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) GetName() string {
	return "forwardToChat"
}

func (r *Action) PerMessage() bool {
	return true
}

func (r *Action) ApplyToMessage(_ interfaces2.StatefulFilter, score *scoringResult.ScoringResult, msg *telego.Message) error {
	forwardParams := &telego.ForwardMessageParams{
		ChatID:              telego.ChatID{ID: r.forwardToChatID},
		FromChatID:          msg.Chat.ChatID(),
		MessageID:           msg.MessageID,
		DisableNotification: true,
	}

	forwardedMsg, err := r.bot.ForwardMessage(forwardParams)
	if err != nil {
		r.logger.Warn("failed to forward message, trying to copy it...", zap.Int64("forwardToChatID", r.forwardToChatID), zap.Int("messageID", msg.MessageID), zap.Error(err))
		var link string
		if msg.Chat.Type == telego.ChatTypeSupergroup {
			link = fmt.Sprintf("https://t.me/c/%v/%v", msg.Chat.Username, msg.MessageID)
		} else {
			link = fmt.Sprintf("https://t.me/c/%v/%v", msg.Chat.ID, msg.MessageID)
		}
		msgParams := &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: r.forwardToChatID},
			Text:   fmt.Sprintf("Message (%v) from user %v:\n%v", link, msg.From.Username, msg.Text),
		}
		forwardedMsg, err = r.bot.SendMessage(msgParams)
	}

	if forwardedMsg == nil {
		return err
	}

	msgText := fmt.Sprintf("used_id: %v\nmessage_spam_score: %v\n\nban_reason:\n%v", forwardedMsg.From.ID, score.Score,
		score.Reason)
	err = tg.SendMarkdownMessage(r.bot, telego.ChatID{ID: r.forwardToChatID}, &forwardedMsg.MessageID, msgText)
	if err != nil {
		err = tg.SendMessage(r.bot, telego.ChatID{ID: r.forwardToChatID}, &forwardedMsg.MessageID, msgText)
	}

	return err
}

func New(logger *zap.Logger, bot *telego.Bot, config map[string]any) (interfaces.Action, error) {
	forwardToChatID, err := config2.GetOptionInt(config, "forwardToChatID")
	if err != nil {
		return nil, err
	}

	return &Action{
		logger:          logger,
		bot:             bot,
		forwardToChatID: int64(forwardToChatID),
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
