package deleteAndBan

import (
	"errors"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	cleanState bool
}

func (r *Action) Apply(callback interfaces2.StatefulFilter, chatID telego.ChatID, messageIDs []int64, userID int64) error {
	for _, messageID := range messageIDs {
		err := r.bot.DeleteMessage(tu.Delete(chatID, int(messageID)))
		if err != nil {
			return err
		}
	}
	req := &telego.BanChatMemberParams{
		UserID: userID,
	}
	req = req.WithChatID(chatID)
	err := r.bot.BanChatMember(req)
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
	cleanState, err := config2.GetOptionBoolWithDefault(config, "cleanState", true)
	if err != nil {
		return nil, err
	}
	return &Action{
		logger:     logger,
		bot:        bot,
		cleanState: cleanState,
	}, nil
}

func Help() string {
	return "deleteAndBan doesn't require any parameter"
}
