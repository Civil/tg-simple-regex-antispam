package addReportButton

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	isAnonymousReport bool
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) Apply(_ interfaces2.StatefulFilter, _ telego.ChatID, _ []int64, _ int64) error {
	return ErrNotSupported
}

func (r *Action) ApplyToMessage(message telego.Message) error {
	params := &telego.GetChatAdministratorsParams{
		ChatID: message.Chat.ChatID(),
	}
	admins, err := r.bot.GetChatAdministrators(params)
	if err != nil {
		return fmt.Errorf("getting chat administrators: %w", err)
	}
	msgBuf := bytes.NewBuffer([]byte{})
	if r.isAnonymousReport {
		msgBuf.WriteString("Spam: ")
	} else {
		msgBuf.WriteString("User @" + message.From.Username + " reported a spam: ")
	}
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
	anonymousReport, ok := config["is_anonymous_report"].(bool)
	if !ok {
		anonymousReport = false
	}
	return &Action{
		logger: logger,
		bot:    bot,

		isAnonymousReport: anonymousReport,
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
