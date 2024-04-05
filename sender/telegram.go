package sender

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ad/anonstickerbot/config"
	"github.com/go-telegram/bot"
	bm "github.com/go-telegram/bot/models"
)

type Sender struct {
	sync.RWMutex
	lgr              *slog.Logger
	config           *config.Config
	Bot              *bot.Bot
	Config           *config.Config
	deferredMessages map[int64]chan DeferredMessage
	lastMessageTimes map[int64]int64
}

func InitSender(lgr *slog.Logger, config *config.Config) (*Sender, error) {
	sender := &Sender{
		lgr:              lgr,
		config:           config,
		deferredMessages: make(map[int64]chan DeferredMessage),
		lastMessageTimes: make(map[int64]int64),
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(sender.handler),
		bot.WithSkipGetMe(),
	}

	b, newBotError := bot.New(config.TelegramToken, opts...)
	if newBotError != nil {
		return nil, fmt.Errorf("start bot error: %s", newBotError)
	}

	go b.Start(context.Background())
	go sender.sendDeferredMessages()

	sender.Bot = b

	return sender, nil
}

func (s *Sender) handler(ctx context.Context, b *bot.Bot, update *bm.Update) {
	if s.config.Debug {
		s.lgr.Debug(formatUpdateForLog(update))
	}
}
