package sender

import (
	"fmt"
	"strings"

	bm "github.com/go-telegram/bot/models"
)

func formatUpdateForLog(message *bm.Update) string {
	switch {
	case message.Message != nil:
		return formatMessageForLog(message)
	case message.EditedMessage != nil:
		return formatMessageEditForLog(message)
	case message.ChannelPost != nil:
		return formatChannelpostForLog(message)
	case message.MessageReaction != nil:
		return formatMessageReactionForLog(message)
	case message.MessageReactionCount != nil:
		return formatMessageReactionCountForLog(message)
	case message.CallbackQuery != nil:
		return formatCallbackQueryForLog(message)
	}

	// jsonData, _ := json.Marshal(message)
	// s.lgr.Debug(fmt.Sprintf("Message %s", string(jsonData)))

	return fmt.Sprintf("%+v", message)
}

func formatMessageForLog(message *bm.Update) string {
	if message.Message.Chat.Type == "private" {
		return fmt.Sprintf("Message from %d (%s): %s", message.Message.Chat.ID, getChatDataFromMessage(&message.Message.Chat), message.Message.Text)
	}

	return fmt.Sprintf(
		"Message from %d (%s) -> %d in (https://t.me/%s/%d): %s",
		message.Message.From.ID,
		getUserDataFromMessage(message.Message.From),
		message.Message.Chat.ID,
		message.Message.Chat.Username,
		message.Message.ID,
		message.Message.Text,
	)
}

func formatMessageEditForLog(message *bm.Update) string {
	if message.EditedMessage.Chat.Type == "private" {
		return fmt.Sprintf("MessageEdit from %d (%s): %s", message.EditedMessage.Chat.ID, getChatDataFromMessage(&message.EditedMessage.Chat), message.EditedMessage.Text)
	}

	return fmt.Sprintf(
		"MessageEdit from %d (%s) -> %d in (https://t.me/%s/%d): %s",
		message.EditedMessage.From.ID,
		getUserDataFromMessage(message.EditedMessage.From),
		message.EditedMessage.Chat.ID,
		message.EditedMessage.Chat.Username,
		message.EditedMessage.ID,
		message.EditedMessage.Text,
	)
}

func formatChannelpostForLog(message *bm.Update) string {
	if message.ChannelPost.Chat.Type == "private" {
		return fmt.Sprintf("ChannelPost from %d (%s): %s", message.ChannelPost.Chat.ID, getChatDataFromMessage(&message.ChannelPost.Chat), message.ChannelPost.Text)
	}

	return fmt.Sprintf(
		"ChannelPost from %d (%s) -> %d in (https://t.me/%s/%d): %s",
		message.ChannelPost.From.ID,
		getUserDataFromMessage(message.ChannelPost.From),
		message.ChannelPost.Chat.ID,
		message.ChannelPost.Chat.Username,
		message.ChannelPost.ID,
		message.ChannelPost.Text,
	)
}

func formatMessageReactionForLog(message *bm.Update) string {
	oldEmoji := []string{}
	newEmoji := []string{}
	for _, reaction := range message.MessageReaction.OldReaction {
		oldEmoji = append(oldEmoji, reaction.ReactionTypeEmoji.Emoji)
	}
	for _, reaction := range message.MessageReaction.NewReaction {
		newEmoji = append(newEmoji, reaction.ReactionTypeEmoji.Emoji)
	}

	if message.MessageReaction.Chat.Type == "private" {
		return fmt.Sprintf("MessageReaction %s -> %s, by %d (%s)", strings.Join(oldEmoji, ","), strings.Join(newEmoji, ","), message.MessageReaction.User.ID, getChatDataFromMessage(&message.MessageReaction.Chat))
	}

	return fmt.Sprintf("MessageReaction %s -> %s, by %d (%s) in (https://t.me/%s/%d)", strings.Join(oldEmoji, ","), strings.Join(newEmoji, ","), message.MessageReaction.User.ID, getUserDataFromMessage(message.MessageReaction.User), message.MessageReaction.Chat.Username, message.MessageReaction.MessageID)
}

func formatCallbackQueryForLog(message *bm.Update) string {
	return fmt.Sprintf("CallbackQuery %s", message.CallbackQuery.Data)
}

func formatMessageReactionCountForLog(message *bm.Update) string {
	return fmt.Sprintf("MessageReactionCount %#v", message.MessageReactionCount)
}

func getChatDataFromMessage(user *bm.Chat) string {
	if user == nil {
		return "Unknown"
	}

	return strings.Join([]string{user.FirstName, user.LastName, user.Username}, " ")
}

func getUserDataFromMessage(user *bm.User) string {
	if user == nil {
		return "Unknown"
	}

	return strings.Join([]string{user.FirstName, user.LastName, user.Username}, " ")
}
