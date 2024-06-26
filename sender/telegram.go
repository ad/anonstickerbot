package sender

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ad/anonstickerbot/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	bm "github.com/go-telegram/bot/models"
)

type Sender struct {
	sync.RWMutex
	logger           *slog.Logger
	config           *config.Config
	Bot              *bot.Bot
	Config           *config.Config
	LastStickers     map[string]string
	deferredMessages map[int64]chan DeferredMessage
	lastMessageTimes map[int64]int64
}

func InitSender(ctx context.Context, logger *slog.Logger, config *config.Config) (*Sender, error) {
	sender := &Sender{
		logger:           logger,
		config:           config,
		LastStickers:     make(map[string]string),
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

	go b.Start(ctx)
	go sender.sendDeferredMessages()

	sender.Bot = b

	return sender, nil
}

func (s *Sender) handler(ctx context.Context, b *bot.Bot, update *bm.Update) {
	if s.config.Debug {
		s.logger.Debug(formatUpdateForLog(update))
	}

	if update.InlineQuery == nil {
		return
	}

	s.RLock()
	defer s.RUnlock()

	results := []models.InlineQueryResult{}

	for name, fileID := range s.LastStickers {
		results = append(results, &models.InlineQueryResultCachedSticker{ID: name, StickerFileID: fileID})
	}

	_, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
		CacheTime:     0,
	})

	if err != nil {
		fmt.Printf("answer inline query error %s\n", err.Error())
	}
}
