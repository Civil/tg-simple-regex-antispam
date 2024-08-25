package tg

import (
	"errors"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
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
	filters *[]interfaces.StatefulFilter
}

func (t *Telego) Start() {
	t.logger.Info("starting telego...")
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
		userID := message.From.ID
		logger := t.logger.With(
			zap.Int64("chat_id", message.Chat.ID),
			zap.Int64("from_user_id", userID),
		)
		logger.Debug("got message", zap.Any("message", message))
		for _, f := range *t.filters {
			logger.Debug("applying filter",
				zap.String("filter_name", f.GetFilterName()),
				zap.String("filter_type", f.GetName()),
			)
			score := f.Score(&message)
			if score > 0 {
				logger.Info("message got scored",
					zap.Int("score", score),
					zap.Any("message", message),
				)
				if score >= 100 && f.IsFinal() {
					logger.Info("stop scoring")
					break
				}
			}
		}

	})

	t.logger.Info("telego initialized")
	bh.Start()
}

func (t *Telego) Stop() {
	t.bot.StopLongPolling()
}

func (t *Telego) GetBot() *telego.Bot {
	return t.bot
}

func NewTelego(logger *zap.Logger, token string, filters *[]interfaces.StatefulFilter) (TgAPI, error) {
	if token == "" || token == "your_telegram_bot_token" {
		logger.Error("no token provided")
		return nil, errors.New("no token provided")
	}

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
