package tg

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
)

type TgAPI interface {
	Start()
	Stop()
	GetBot() *telego.Bot
}

type Telego struct {
	token  string
	logger *zap.Logger

	bot     *telego.Bot
	filters []interfaces.StatefulFilter
}

func (t *Telego) Start() {
	botUser, err := t.bot.GetMe()
	if err != nil {
		t.logger.Error("bot cannot identify itself", zap.Error(err))
		return
	}
	t.logger.Info("bot user info", zap.Any("bot_user", botUser))

	updates, _ := t.bot.UpdatesViaLongPolling(nil)
	defer t.bot.StopLongPolling()

	bh, _ := th.NewBotHandler(t.bot, updates)
	defer bh.Stop()

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		chatID := tu.ID(message.Chat.ID)
		userID := message.From.ID
		logger := t.logger.With(
			zap.Int64("chat_id", message.Chat.ID),
			zap.Int64("from_user_id", userID),
		)
		logger.Info("got message")
		if logger.Level().Enabled(zap.DebugLevel) {
			logger.Debug("message content", zap.Any("message", message))
			_, err := bot.SendMessage(
				tu.Messagef(chatID, "got a message that have message_id=%d, type(media_group_id)=%T media_group_id=%+v\n",
					message.MessageID, message.MediaGroupID,
					message.MediaGroupID),
			)
			if err != nil {
				logger.Error("error sending metadata", zap.Error(err))
			}
			newMsgId, err := bot.CopyMessage(
				tu.CopyMessage(chatID, chatID, message.MessageID),
			)
			if err != nil {
				logger.Error("error copying message", zap.Error(err))
				return
			}
			logger.Info("message copied", zap.Int("new_message_id", newMsgId.MessageID))
		}
		for _, f := range t.filters {
			logger.Info("applying filter",
				zap.String("filter_name", f.GetName()),
			)
			score := f.Score(message)
			if score > 0 {
				logger.Info("message got scored",
					zap.Int("score", score),
				)
				if score >= 100 && f.IsFinal() {
					logger.Info("stop scoring, as")
					break
				}
			}
		}

	})

	bh.Start()
}

func (t *Telego) Stop() {
	t.bot.StopLongPolling()
}

func (t *Telego) GetBot() *telego.Bot {
	return t.bot
}

func NewTelego(logger *zap.Logger, token string, filters []interfaces.StatefulFilter) (TgAPI, error) {
	t := &Telego{
		logger:  logger,
		token:   token,
		filters: filters,
	}

	bot, err := telego.NewBot(t.token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, err
	}
	t.bot = bot

	return t, nil
}
