package helpers

import (
	"github.com/mymmrac/telego"
)

func SendMessage(bot *telego.Bot, chatID telego.ChatID, messageID *int, text string) error {
	sendMessageParams := &telego.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	if messageID != nil {
		sendMessageParams.ReplyParameters = &telego.ReplyParameters{
			MessageID: *messageID,
		}
	}
	_, err := bot.SendMessage(
		sendMessageParams,
	)
	return err
}

func BanUser(bot *telego.Bot, chatID telego.ChatID, userID int64, deleteAll bool) error {
	req := &telego.BanChatMemberParams{
		UserID:         userID,
		RevokeMessages: deleteAll,
	}
	req = req.WithChatID(chatID)
	err := bot.BanChatMember(req)
	if err != nil {
		return err
	}
	return nil
}
