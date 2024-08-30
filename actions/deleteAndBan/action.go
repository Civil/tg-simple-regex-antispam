package deleteAndBan

import (
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	cleanState bool
	dryRun     bool
	deleteAll  bool
}

func (r *Action) Apply(callback interfaces2.StatefulFilter, chatID telego.ChatID, messageIDs []int64, userID int64) error {
	if r.dryRun {
		r.logger.Debug("applying action in dry run mode")
		sendMessageParams := &telego.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("ban conditions for user with id=%v has been met, but dryRun is enabled", userID),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: int(messageIDs[0]),
			},
		}

		_, err := r.bot.SendMessage(sendMessageParams)
		if err != nil {
			r.logger.Error("failed to send dryRun message", zap.Int64("userID", userID))
		}
		return err
	}

	r.logger.Debug("applying action in normal mode")
	for _, messageID := range messageIDs {
		err := r.bot.DeleteMessage(tu.Delete(chatID, int(messageID)))
		if err != nil {
			return err
		}
	}

	err := tg.BanUser(r.bot, chatID, userID, r.deleteAll)
	if err != nil {
		r.logger.Error("failed to ban user", zap.Int64("userID", userID), zap.Error(err))
	}

	msgIds := make([]int, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		msgIds = append(msgIds, int(messageID))
	}
	deleteParams := &telego.DeleteMessagesParams{
		ChatID:     chatID,
		MessageIDs: msgIds,
	}
	err = r.bot.DeleteMessages(deleteParams)
	if err != nil {
		return err
	}

	if r.cleanState {
		err = callback.RemoveState(userID)
		if err != nil {
			r.logger.Error("failed to remove state", zap.Int64("userID", userID), zap.Error(err))
			return nil
		}
	}

	return nil
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) ApplyToMessage(_ interfaces2.StatefulFilter, _ *telego.Message) error {
	return ErrNotSupported
}

func New(logger *zap.Logger, bot *telego.Bot, config map[string]any) (interfaces.Action, error) {
	cleanState, err := config2.GetOptionBoolWithDefault(config, "cleanState", false)
	if err != nil {
		return nil, err
	}
	dryRyn, err := config2.GetOptionBoolWithDefault(config, "dryRun", true)
	if err != nil {
		return nil, err
	}
	deleteAll, err := config2.GetOptionBoolWithDefault(config, "deleteAll", true)
	if err != nil {
		return nil, err
	}
	return &Action{
		logger:     logger,
		bot:        bot,
		dryRun:     dryRyn,
		cleanState: cleanState,
		deleteAll:  deleteAll,
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
