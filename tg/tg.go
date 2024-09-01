package tg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/helper/logs"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type TgAPI interface {
	Start()
	Stop()
	GetBot() *telego.Bot
	UpdatePrefixes()
}

type Telego struct {
	token  string
	logger *zap.Logger

	bot      *telego.Bot
	filters  *[]interfaces.StatefulFilter
	adminIDs map[int64]struct{}
	banDB    bannedDB.BanDB

	handlers map[string]tg.AdminCMDHandlerFunc
}

func New(logger *zap.Logger, token string, filters *[]interfaces.StatefulFilter, adminIDs []int64, banDB bannedDB.BanDB) (TgAPI,
	error) {
	if token == "" || token == "your_telegram_bot_token" {
		logger.Error("no token provided")
		return nil, errors.New("no token provided")
	}

	adminIDsMap := make(map[int64]struct{})
	for _, id := range adminIDs {
		adminIDsMap[id] = struct{}{}
	}

	t := &Telego{
		banDB:    banDB,
		logger:   logger,
		token:    token,
		filters:  filters,
		adminIDs: adminIDsMap,
		handlers: make(map[string]tg.AdminCMDHandlerFunc),
	}

	for _, filter := range *filters {
		logger.Info("registering filter", zap.String("filter", filter.GetFilterName()), zap.String("chain_name", filter.TGAdminPrefix()))
		prefix := filter.TGAdminPrefix()
		if prefix != "" {
			t.handlers[prefix] = filter.HandleTGCommands
		}
	}
	t.handlers[t.banDB.TGAdminPrefix()] = t.banDB.HandleTGCommands
	t.handlers["listCmds"] = t.listAdminPrefixes

	bot, err := telego.NewBot(t.token, telego.WithLogger(logs.New(t.logger)))
	if err != nil {
		return nil, err
	}
	t.bot = bot

	return t, nil
}

func (t *Telego) UpdatePrefixes() {
	for _, filter := range *t.filters {
		prefix := filter.TGAdminPrefix()
		keys := make([]string, 0)
		for k := range t.handlers {
			if strings.HasPrefix(k, prefix) {
				keys = append(keys, k)
			}
		}
		t.logger.Debug("checking filter",
			zap.String("filter", filter.GetFilterName()),
			zap.String("filter_type", filter.GetName()),
			zap.String("chain_name", prefix),
			zap.Any("handlers", keys),
		)

		if prefix != "" {
			t.logger.Debug("checking if filter was registered")
			if _, ok := t.handlers[prefix]; !ok {
				t.logger.Debug("registering filter", zap.String("filter", filter.GetFilterName()))
				t.handlers[prefix] = filter.HandleTGCommands
			}
		}
	}
}

func (t *Telego) listAdminPrefixes(logger *zap.Logger, bot *telego.Bot, message *telego.Message, _ []string) error {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("Available subcommands:\n\n")
	for prefix := range t.handlers {
		buf.WriteString("   " + prefix + "\n")
	}

	err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
	if err != nil {
		logger.Error("failed to send message", zap.Error(err))
	}
	return err
}

func (t *Telego) HandleAdminMessages(logger *zap.Logger, bot *telego.Bot, message *telego.Message) {
	logger.Debug("admin command", zap.String("command", message.Text))
	tokens := strings.Split(message.Text, " ")
	if len(tokens) < 2 {
		err := t.listAdminPrefixes(logger, bot, message, nil)
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
		return
	}

	if h, ok := t.handlers[tokens[1]]; ok {
		var err error
		if len(tokens) > 2 {
			err = h(logger, bot, message, tokens[2:])
		} else {
			err = h(logger, bot, message, nil)
		}
		if err != nil {
			logger.Error("failed to handle command", zap.Error(err))
		}
		return
	}

	logger.Warn("unsupported command", zap.Any("message", message))
	err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
		fmt.Sprintf("unsupported command: %v", message.Text))
	if err != nil {
		logger.Error("failed to send message", zap.Error(err))
	}
}

func (t *Telego) HandleMessages(bot *telego.Bot, message telego.Message) {
	userID := message.From.ID
	logger := t.logger.With(
		zap.Int64("chat_id", message.Chat.ID),
		zap.Int64("from_user_id", userID),
	)
	logger.Debug("got message", zap.Any("message", message))
	if message.Text == "/admin" || strings.HasPrefix(message.Text, "/admin ") {
		if _, ok := t.adminIDs[userID]; !ok {
			logger.Debug("user is not in list of extra super users, checking chat admins")
			params := &telego.GetChatAdministratorsParams{
				ChatID: message.Chat.ChatID(),
			}
			chatAdmins, err := bot.GetChatAdministrators(params)
			if err != nil {
				logger.Error("failed to get chat administrators", zap.Error(err))
			}
			ok = false
			for _, admin := range chatAdmins {
				if admin.MemberUser().ID == userID {
					ok = true
					logger.Debug("user is chat admin", zap.Any("user_id", userID))
					break
				}
			}
			if !ok {
				logger.Warn("user is not admin or chat admin", zap.Any("user_id", userID), zap.Any("message", message))
				return
			}
		}
		t.HandleAdminMessages(logger, bot, &message)
		return
	}
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

	bh.HandleMessage(t.HandleMessages)
	bh.HandleEditedMessage(t.HandleMessages)

	t.logger.Info("telego initialized")
	bh.Start()
}

func (t *Telego) Stop() {
	t.bot.StopLongPolling()
}

func (t *Telego) GetBot() *telego.Bot {
	return t.bot
}
