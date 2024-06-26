package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/ad/anonstickerbot/config"
	"github.com/ad/anonstickerbot/logger"
	sndr "github.com/ad/anonstickerbot/sender"
	su "github.com/ad/anonstickerbot/stickerUpdater"
)

func Run(ctx context.Context, w io.Writer, args []string) error {
	conf, errInitConfig := config.InitConfig(os.Args)
	if errInitConfig != nil {
		return errInitConfig
	}

	lgr := logger.InitLogger(conf.Debug)

	// Recovery
	defer func() {
		if p := recover(); p != nil {
			lgr.Error(fmt.Sprintf("panic recovered: %s; stack trace: %s", p, string(debug.Stack())))
		}
	}()

	sender, errInitSender := sndr.InitSender(ctx, lgr, conf)
	if errInitSender != nil {
		return errInitSender
	}

	me, err := sender.Bot.GetMe(ctx)
	if err != nil {
		return err
	}

	if len(conf.TelegramAdminIDsList) != 0 {
		sender.MakeRequestDeferred(sndr.DeferredMessage{
			Method: "sendMessage",
			ChatID: conf.TelegramAdminIDsList[0],
			Text:   "Bot restarted: " + me.Username,
		}, sender.SendResult)
	}

	stickerUpdater, err := su.InitStickerUpdater(lgr, conf, sender.Bot, sender)
	if err != nil {
		return err
	}

	err = stickerUpdater.RunAll()
	if err != nil {
		fmt.Println(err)
	}

	updateTicker := time.NewTicker(time.Duration(conf.UPDATE_DELAY) * time.Second)

	go func() {
		for range updateTicker.C {
			err = stickerUpdater.RunAll()
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	return nil
}
