package tg

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type TgAPI interface {
	Start()
	Stop()
	GetBot() *telego.Bot
}

type Telego struct {
	token  string
	logger *zap.Logger

	bot      *telego.Bot
	filters  *[]interfaces.StatefulFilter
	adminIDs map[int64]struct{}
	banDB    bannedDB.BanDB
}

func (t *Telego) HandleBanDBMessages(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) {
	logger.Debug("ban db command", zap.Strings("tokens", tokens))
	switch strings.ToLower(tokens[0]) {
	case "list":
		list, err := t.banDB.ListUserIDs()
		if err != nil {
			logger.Error("failed to list banned users", zap.Error(err))
			return
		}
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("Banned users:\n")
		for _, userID := range list {
			buf.WriteString(fmt.Sprintf("%v\n", userID))
		}
		err = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
	case "unban":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return
		}
		err = t.banDB.UnbanUser(userIDInt)
		if err != nil {
			logger.Error("failed to unban user", zap.String("userID", userID), zap.Error(err))
			_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot unban user: %s",
					err.Error()))
			return
		}
		_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v unbanned", userIDInt))
	case "bannodel":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return
		}
		err = t.banDB.BanUser(userIDInt)
		if err != nil {
			logger.Error("failed to ban user", zap.String("userID", userID), zap.Error(err))
			_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot ban user: %s",
					err.Error()))
			return
		}
		err = tg.BanUser(bot, message.Chat.ChatID(), userIDInt, false)
		_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v banned", userIDInt))
	case "ban":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return
		}
		err = t.banDB.BanUser(userIDInt)
		if err != nil {
			logger.Error("failed to ban user", zap.String("userID", userID), zap.Error(err))
			_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot ban user: %s",
					err.Error()))
			return
		}
		err = tg.BanUser(bot, message.Chat.ChatID(), userIDInt, true)
		_ = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v banned", userIDInt))
	case "help":
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("Available commands:\n")
		buf.WriteString(" `list` - list all banned users (IDs only)")
		buf.WriteString(" `ban` - ban user by ID and delete all messages")
		buf.WriteString(" `banNoDel` - ban user by ID but keep all messages")
		buf.WriteString(" `unban` - unban user by ID")
		buf.WriteString(" `help` - this help")

		err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
	default:
		logger.Warn("unsupported command", zap.Strings("tokens", tokens))
	}

}

func (t *Telego) HandleAdminMessages(logger *zap.Logger, bot *telego.Bot, message *telego.Message) {
	logger.Debug("admin command", zap.String("command", message.Text))
	tokens := strings.Split(message.Text, " ")
	if len(tokens) < 2 {
		logger.Warn("invalid command", zap.Any("message", message))
		err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("invalid command: %v",
			message.Text))
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
		return
	}

	switch tokens[1] {
	case t.banDB.TGAdminPrefix():
		err := t.banDB.HandleTGCommands(logger, bot, message, tokens)
		if err != nil {
			logger.Error("failed to handle ban db command", zap.Error(err))
		}
	default:
		logger.Warn("unsupported command", zap.Any("message", message))
		err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("unsupported command: %v", message.Text))
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
	}
}

func (t *Telego) HandleMessages(bot *telego.Bot, message telego.Message) {
	userID := message.From.ID
	logger := t.logger.With(
		zap.Int64("chat_id", message.Chat.ID),
		zap.Int64("from_user_id", userID),
	)
	logger.Debug("got message", zap.Any("message", message))
	if strings.HasPrefix(message.Text, "/admin ") {
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

func NewTelego(logger *zap.Logger, token string, filters *[]interfaces.StatefulFilter, adminIDs []int64, banDB bannedDB.BanDB) (TgAPI,
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
	}

	bot, err := telego.NewBot(t.token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, err
	}
	t.bot = bot

	return t, nil
}
