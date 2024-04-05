package sender

import (
	"context"
	"reflect"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	sendCooldownPerUser = int64(time.Second / 3)
	sendInterval        = time.Second
)

type DeferredMessage struct {
	Method string

	ChatID int64  // for copyMessage, forwardMessage and sendMessage
	Text   string // for sendMessage

	fromChatID string // for copyMessage and forwardMessage
	messageID  int    // for copyMessage and forwardMessage

	messageThreadID  int // for copyMessage and sendMessage
	replyToMessageID int // for copyMessage and sendMessage

	replyMarkup models.ReplyMarkup // for sendMessage

	callback func(SendResult) error
}

type SendResult struct {
	ChatID      int64
	Msg         string
	Error       error
	MessageID   int64
	ForwardDate int
}

func (s *Sender) MakeRequestDeferred(dm DeferredMessage, callback func(s SendResult) error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.deferredMessages[dm.ChatID]; !ok {
		s.deferredMessages[dm.ChatID] = make(chan DeferredMessage, 100)
	}

	dm.callback = callback

	s.deferredMessages[dm.ChatID] <- dm
}

func (s *Sender) sendDeferredMessages() {
	timer := time.NewTicker(sendInterval)

	for range timer.C {
		var cases []reflect.SelectCase

		for userID, ch := range s.deferredMessages {
			if s.userCanReceiveMessage(userID) && len(ch) > 0 {
				sc := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
				cases = append(cases, sc)
			}
		}

		if len(cases) > 0 {
			_, value, ok := reflect.Select(cases)

			if ok {
				dm := value.Interface().(DeferredMessage)

				var (
					err         error
					messageID   int64
					forwardDate int
				)

				switch dm.Method {
				case "sendMessage":
					resultMessage, errSendMessage := s.Bot.SendMessage(context.Background(), &bot.SendMessageParams{
						ChatID:          dm.ChatID,
						Text:            dm.Text,
						MessageThreadID: dm.messageThreadID,
						ReplyParameters: &models.ReplyParameters{MessageID: dm.replyToMessageID},
						ReplyMarkup:     dm.replyMarkup,
					})

					messageID = int64(resultMessage.ID)
					err = errSendMessage
				case "sendMessageHTML":
					resultMessage, errSendMessage := s.Bot.SendMessage(context.Background(), &bot.SendMessageParams{
						ChatID:    dm.ChatID,
						Text:      dm.Text,
						ParseMode: "HTML",
						LinkPreviewOptions: &models.LinkPreviewOptions{
							IsDisabled: bot.True(),
						},
					})

					messageID = int64(resultMessage.ID)
					err = errSendMessage

				case "copyMessage":
					resultMessage, errCopyMessage := s.Bot.CopyMessage(context.Background(), &bot.CopyMessageParams{
						ChatID:          dm.ChatID,
						FromChatID:      dm.fromChatID,
						MessageID:       dm.messageID,
						MessageThreadID: dm.messageThreadID,
						ReplyParameters: &models.ReplyParameters{MessageID: dm.replyToMessageID},
					})

					messageID = int64(resultMessage.ID)
					err = errCopyMessage

				case "forwardMessage":
					resultMessage, errForwardMessage := s.Bot.ForwardMessage(context.Background(), &bot.ForwardMessageParams{
						ChatID:     dm.ChatID,
						FromChatID: dm.fromChatID,
						MessageID:  dm.messageID,
					})

					if resultMessage.ForwardOrigin.MessageOriginHiddenUser == nil {
						forwardDate = resultMessage.ForwardOrigin.MessageOriginUser.Date
					} else {
						forwardDate = resultMessage.ForwardOrigin.MessageOriginHiddenUser.Date
					}

					messageID = int64(resultMessage.ID)
					err = errForwardMessage
					dm.Text = resultMessage.Text
				}

				if dm.callback != nil {
					_ = dm.callback(
						SendResult{
							ChatID:      dm.ChatID,
							Msg:         dm.Text,
							Error:       err,
							MessageID:   messageID,
							ForwardDate: forwardDate,
						},
					)
				}

				s.lastMessageTimes[dm.ChatID] = time.Now().UnixNano()
			}
		}
	}
}

func (s *Sender) userCanReceiveMessage(userID int64) bool {
	t, ok := s.lastMessageTimes[userID]

	return !ok || t+sendCooldownPerUser <= time.Now().UnixNano()
}
