package addReportButton

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	isAnonymousReport bool
	msgPrefix         string
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) Apply(_ interfaces2.StatefulFilter, _ *scoringResult.ScoringResult, _ telego.ChatID, _ []int64,
	_ int64, _ any) error {
	return ErrNotSupported
}

func (r *Action) ApplyToMessage(_ interfaces2.StatefulFilter, _ *scoringResult.ScoringResult,
	message *telego.Message, extraParams any) error {
	vacationAdmins, ok := extraParams.(map[string]time.Time)
	if !ok {
		r.logger.Error("extraParams is not a map[string]time.Time")
		vacationAdmins = make(map[string]time.Time)
	}
	r.logger.Debug("adding report button to message",
		zap.Any("message", message),
		zap.Any("chat", message.Chat),
	)
	params := &telego.GetChatAdministratorsParams{
		ChatID: message.Chat.ChatID(),
	}
	admins, err := r.bot.GetChatAdministrators(params)
	if err != nil {
		return fmt.Errorf("getting chat administrators: %w", err)
	}
	msgBuf := bytes.NewBuffer([]byte{})
	if r.msgPrefix != "" {
		msgBuf.WriteString(r.msgPrefix + " ")
	}
	if r.isAnonymousReport {
		msgBuf.WriteString("Spam or chat rules violation: ")
	} else {
		msgBuf.WriteString("User @" + message.From.Username + " reported a spam or rules violation: ")
	}
	firstAdmin := true
	for _, admin := range admins {
		adminUsername := admin.MemberUser().Username
		if _, ok = vacationAdmins[adminUsername]; ok {
			continue
		}
		if strings.HasSuffix(strings.ToLower(adminUsername), "bot") {
			continue
		}
		if !firstAdmin {
			msgBuf.WriteString(", ")
		}
		msgBuf.WriteRune('@')
		msgBuf.WriteString(adminUsername)
		firstAdmin = false
	}

	sendMessageParams := &telego.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   msgBuf.String(),
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

func (r *Action) GetName() string {
	return "addReportButton"
}

func (r *Action) PerMessage() bool {
	return true
}

func New(logger *zap.Logger, bot *telego.Bot, config map[string]any) (interfaces.Action, error) {
	anonymousReport, err := config2.GetOptionBoolWithDefault(config, "isAnonymousReport", true)
	if err != nil {
		return nil, err
	}
	msgPrefix := config2.GetOptionStringWithDefault(config, "msgPrefix", "")
	return &Action{
		logger: logger,
		bot:    bot,

		isAnonymousReport: anonymousReport,
		msgPrefix:         msgPrefix,
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
