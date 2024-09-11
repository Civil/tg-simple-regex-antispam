package tg

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
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

func SendMarkdownMessage(bot *telego.Bot, chatID telego.ChatID, messageID *int, text string) error {
	sendMessageParams := &telego.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: telego.ModeMarkdownV2,
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

func DeleteMessage(bot *telego.Bot, msg *telego.Message) error {
	return bot.DeleteMessage(tu.Delete(msg.Chat.ChatID(), msg.MessageID))
}
