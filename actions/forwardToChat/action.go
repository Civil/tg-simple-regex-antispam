package forwardToChat

import (
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	forwardToChatID int64
}

func (r *Action) Apply(_ interfaces2.StatefulFilter, _ telego.ChatID, _ []int64, _ int64) error {
	return ErrNotSupported
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) GetName() string {
	return "forwardToChat"
}

func (r *Action) PerMessage() bool {
	return true
}

func (r *Action) ApplyToMessage(_ interfaces2.StatefulFilter, msg *telego.Message) error {
	forwardParams := &telego.ForwardMessageParams{
		ChatID:              telego.ChatID{ID: r.forwardToChatID},
		FromChatID:          msg.Chat.ChatID(),
		MessageID:           msg.MessageID,
		DisableNotification: true,
	}

	_, err := r.bot.ForwardMessage(forwardParams)
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
		_, err = r.bot.SendMessage(msgParams)
		return err
	}

	return nil
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