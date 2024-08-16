package deleteAndBan

import (
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot
}

func (r *Action) Apply(chatID telego.ChatID, messageIDs []int64, userID int64) error {
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
	return nil
}

func (r *Action) ApplyToMessage(_ telego.Message) error {
	return fmt.Errorf("not supported")
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
