package addReportButton

import (
	"bytes"
	"fmt"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot
}

func (r *Action) Apply(_ telego.ChatID, _ []int64, _ int64) error {
	return fmt.Errorf("not supported")
}

func (r *Action) ApplyToMessage(message telego.Message) error {
	params := &telego.GetChatAdministratorsParams{
		ChatID: message.Chat.ChatID(),
	}
	admins, err := r.bot.GetChatAdministrators(params)
	if err != nil {
		return fmt.Errorf("getting chat administrators: %w", err)
	}
	msgBuf := bytes.NewBuffer([]byte("User @" + message.From.Username + " reported a spam: "))
	for i, admin := range admins {
		if i != 0 {
			msgBuf.WriteString(", ")
		}
		msgBuf.WriteRune('@')
		msgBuf.WriteString(admin.MemberUser().Username)
	}

	sendMessageParams := &telego.SendMessageParams{
		Text: msgBuf.String(),
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	}

	_, err = r.bot.SendMessage(
		sendMessageParams,
	)
	msgBuf.Reset()
	if err != nil {
		return err
	}

	return nil
}

func New(logger *zap.Logger, bot *telego.Bot, config map[string]any) (interfaces.Action, error) {
	_ = config
	return &Action{
		logger: logger,
		bot:    bot,
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
