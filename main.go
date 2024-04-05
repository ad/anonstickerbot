package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ad/anonstickerbot/config"
	"github.com/ad/anonstickerbot/logger"
	"github.com/ad/anonstickerbot/sender"
	"github.com/fogleman/gg"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/golang/freetype/truetype"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/font/gofont/goregular"
)

func main() {
	conf, errConfig := config.InitConfig(os.Args)
	if errConfig != nil {
		fmt.Println(errConfig)

		return
	}

	lgr := logger.InitLogger(conf.Debug)

	// Recovery
	defer func() {
		if p := recover(); p != nil {
			lgr.Error(fmt.Sprintf("panic recovered: %s; stack trace: %s", p, string(debug.Stack())))
		}
	}()

	s, errInitSender := sender.InitSender(lgr, conf)
	if errInitSender != nil {
		fmt.Println(errInitSender)

		return
	}

	me, err := s.Bot.GetMe(context.Background())
	if err != nil {
		fmt.Println(err)

		return
	}

	stickerSetName := fmt.Sprintf("stickers_by_%s", me.Username)

	if len(conf.TelegramAdminIDsList) != 0 {
		s.MakeRequestDeferred(sender.DeferredMessage{
			Method: "sendMessage",
			ChatID: conf.TelegramAdminIDsList[0],
			Text:   "Bot restarted: " + me.Username,
		}, s.SendResult)
	}

	err = updateSticker(
		s.Bot,
		stickerSetName,
		conf.DATA_URL,
		conf.IMG_IN_PATH,
		conf.IMG_OUT_PATH,
		me.Username,
		conf.TelegramAdminIDsList[0],
	)
	if err != nil {
		fmt.Println(err)

		// return
	}

	updateTicker := time.NewTicker(time.Duration(conf.UPDATE_DELAY) * time.Second)
	for {
		select {
		case <-updateTicker.C:
			err = updateSticker(
				s.Bot,
				stickerSetName,
				conf.DATA_URL,
				conf.IMG_IN_PATH,
				conf.IMG_OUT_PATH,
				me.Username,
				conf.TelegramAdminIDsList[0],
			)
			if err != nil {
				fmt.Println(err)

				// return
			}
		}
	}
}

// Comma produces a string form of the given number in base 10 with
// commas after every three orders of magnitude.
//
// e.g. Comma(834142) -> 834,142
func Comma(v int64) string {
	sign := ""

	// Min int64 can't be negated to a usable value, so it has to be special cased.
	if v == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}

	if v < 0 {
		sign = "-"
		v = 0 - v
	}

	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)
		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))

	return sign + strings.Join(parts[j:], ",")
}

func updateSticker(b *bot.Bot, stickerSetName, dataURL, imgInPath, imgOutPath, botUsername string, telegramID int64) error {
	data, err := getData(dataURL)
	if err != nil {
		return err
	}

	fmt.Println("-------------------------------------")

	m5Volume := int64(0)
	m5VolumeFloat, err := strconv.ParseFloat(data.Data.Attributes.VolumeUsd.M5, 64)
	if err == nil {
		m5Volume = int64(m5VolumeFloat)
	}

	h1Volume := int64(0)
	h1VolumeFloat, err := strconv.ParseFloat(data.Data.Attributes.VolumeUsd.H1, 64)
	if err == nil {
		h1Volume = int64(h1VolumeFloat)
	}

	h24Volume := int64(0)
	h24VolumeFloat, err := strconv.ParseFloat(data.Data.Attributes.VolumeUsd.H24, 64)
	if err == nil {
		h24Volume = int64(h24VolumeFloat)
	}

	m5PricePercentage, err := strconv.ParseFloat(data.Data.Attributes.PriceChangePercentage.M5, 64)
	if err != nil {
		m5PricePercentage = 0
	}

	h1PricePercentage, err := strconv.ParseFloat(data.Data.Attributes.PriceChangePercentage.H1, 64)
	if err != nil {
		h1PricePercentage = 0
	}

	h24PricePercentage, err := strconv.ParseFloat(data.Data.Attributes.PriceChangePercentage.H24, 64)
	if err != nil {
		h24PricePercentage = 0
	}

	fmt.Printf("Name: %s\n", data.Data.Attributes.Name)
	fmt.Printf("Base token price USD: %s\n", data.Data.Attributes.BaseTokenPriceUsd)
	fmt.Printf("Quote token price USD: %s\n", data.Data.Attributes.QuoteTokenPriceUsd)
	fmt.Printf("Base token price quote token: %s\n", data.Data.Attributes.BaseTokenPriceQuoteToken)
	fmt.Printf("Quote token price base token: %s\n", data.Data.Attributes.QuoteTokenPriceBaseToken)
	fmt.Printf("Price change percentage M5: %.2f%%, volume %d, buy %d/sell %d\n", m5PricePercentage, m5Volume, data.Data.Attributes.Transactions.M5.Buys, data.Data.Attributes.Transactions.M5.Sells)
	fmt.Printf("Price change percentage H1: %.2f%%, volume %d, buy %d/sell %d\n", h1PricePercentage, h1Volume, data.Data.Attributes.Transactions.H1.Buys, data.Data.Attributes.Transactions.H1.Sells)
	// fmt.Printf("Price change percentage H6: %.2f%%, volume %.2f \n", data.Data.Attributes.PriceChangePercentage.H6, data.Data.Attributes.VolumeUsd.H6)
	fmt.Printf("Price change percentage H24: %.2f%%, volume %d, buy %d/sell %d\n", h24PricePercentage, h24Volume, data.Data.Attributes.Transactions.H24.Buys, data.Data.Attributes.Transactions.H24.Sells)
	fmt.Printf("Reserve in USD: %s\n", data.Data.Attributes.ReserveInUsd)

	inputFile, err := os.Open(imgInPath)
	if err != nil {
		return fmt.Errorf("error loading image: %v", err)
	}
	defer inputFile.Close()

	img, err := webp.Decode(inputFile, &decoder.Options{})
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	dc := gg.NewContextForImage(img)

	font, _ := truetype.Parse(goregular.TTF)
	face18 := truetype.NewFace(font, &truetype.Options{Size: 18})
	face24 := truetype.NewFace(font, &truetype.Options{Size: 24})
	face26 := truetype.NewFace(font, &truetype.Options{Size: 26})
	face32 := truetype.NewFace(font, &truetype.Options{Size: 32})

	// dc.Clear()

	dc.SetRGB(1, 1, 1)
	dc.SetFontFace(face32)
	dc.DrawString(data.Data.Attributes.Name, 70, 58)

	dc.SetFontFace(face24)
	dc.DrawStringAnchored(data.Data.Attributes.QuoteTokenPriceBaseToken, 370, 54, 1, 0)

	dc.SetFontFace(face18)
	dc.DrawStringAnchored(data.Data.Attributes.BaseTokenPriceQuoteToken, 490, 54, 1, 0)

	dc.SetFontFace(face26)
	// data.Data.Attributes.PriceChangePercentage.M5,
	dc.DrawStringWrapped(
		fmt.Sprintf(
			"5M\n$%s\n%s/%s",
			Comma(m5Volume),
			Comma(int64(data.Data.Attributes.Transactions.M5.Buys)),
			Comma(int64(data.Data.Attributes.Transactions.M5.Sells)),
		),
		24,
		140,
		0,
		0,
		150,
		1.25,
		gg.AlignLeft,
	)

	m5PricePercentageColor := color.RGBA{128, 128, 128, 255}
	if m5PricePercentage > 0 {
		m5PricePercentageColor = color.RGBA{126, 211, 33, 255}
	} else if m5PricePercentage < 0 {
		m5PricePercentageColor = color.RGBA{208, 2, 27, 255}
	}

	dc.SetColor(m5PricePercentageColor)
	dc.DrawStringAnchored(
		fmt.Sprintf("%.2f%%", math.Abs(m5PricePercentage)),
		64,
		166,
		0,
		0,
	)

	dc.SetRGB(1, 1, 1)
	// data.Data.Attributes.PriceChangePercentage.M5,
	dc.DrawStringWrapped(
		fmt.Sprintf(
			"1H\n$%s\n%s/%s",
			Comma(h1Volume),
			Comma(int64(data.Data.Attributes.Transactions.H1.Buys)),
			Comma(int64(data.Data.Attributes.Transactions.H1.Sells)),
		),
		184,
		140,
		0,
		0,
		150,
		1.25,
		gg.AlignLeft,
	)

	h1PricePercentageColor := color.RGBA{128, 128, 128, 255}
	if h1PricePercentage > 0 {
		h1PricePercentageColor = color.RGBA{126, 211, 33, 255}
	} else if h1PricePercentage < 0 {
		h1PricePercentageColor = color.RGBA{208, 2, 27, 255}
	}

	dc.SetColor(h1PricePercentageColor)
	dc.DrawStringAnchored(
		fmt.Sprintf("%.2f%%", math.Abs(h1PricePercentage)),
		222,
		166,
		0,
		0,
	)

	dc.SetRGB(1, 1, 1)

	dc.DrawStringWrapped(
		fmt.Sprintf(
			"24H\n$%s\n%s/%s",
			Comma(h24Volume),
			Comma(int64(data.Data.Attributes.Transactions.H24.Buys)),
			Comma(int64(data.Data.Attributes.Transactions.H24.Sells)),
		),
		334,
		140,
		0,
		0,
		150,
		1.25,
		gg.AlignLeft,
	)

	h24PricePercentageColor := color.RGBA{128, 128, 128, 255}
	if h24PricePercentage > 0 {
		h24PricePercentageColor = color.RGBA{126, 211, 33, 255}
	} else if h24PricePercentage < 0 {
		h24PricePercentageColor = color.RGBA{208, 2, 27, 255}
	}

	dc.SetColor(h24PricePercentageColor)
	dc.DrawStringAnchored(
		fmt.Sprintf("%.2f%%", math.Abs(h24PricePercentage)),
		384,
		166,
		0,
		0,
	)

	dc.SetRGB(1, 1, 1)

	dc.SetFontFace(face18)
	dc.DrawStringAnchored(time.Now().Format(time.RFC850), 24, 490, 0, 0)

	outputFile, err := os.Create(imgOutPath)
	if err != nil {
		return fmt.Errorf("error creating image: %v", err)
	}
	defer outputFile.Close()

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		return fmt.Errorf("error NewLossyEncoderOptions: %v", err)
	}

	if err := webp.Encode(outputFile, dc.Image(), options); err != nil {
		return fmt.Errorf("error Encode: %v", err)
	}

	fmt.Println("-------------------------------------")

	fileContent, _ := os.ReadFile(imgOutPath)

	file, err := b.UploadStickerFile(context.Background(), &bot.UploadStickerFileParams{
		UserID: telegramID,
		PngSticker: &models.InputFileUpload{
			Filename: imgOutPath,
			Data:     bytes.NewReader(fileContent),
		},
	})

	if err != nil {
		return err
	}

	fmt.Printf("FileID: %s\n", file.FileID)

	stickerSet, err := b.GetStickerSet(context.Background(), &bot.GetStickerSetParams{
		Name: stickerSetName,
	})

	// time.Sleep(1 * time.Second)

	if err != nil {
		if err.Error() == "bad request, Bad Request: STICKERSET_INVALID" {
			result, err := b.CreateNewStickerSet(context.Background(), &bot.CreateNewStickerSetParams{
				UserID:          telegramID,
				Name:            stickerSetName,
				Title:           "Stickers by " + botUsername,
				NeedsRepainting: false,
				Stickers: []models.InputSticker{
					{
						Sticker: &models.InputFileString{
							Data: file.FileID,
						},
						EmojiList: []string{"🚀"},
						Format:    "static",
						MaskPosition: models.MaskPosition{
							Point: "forehead",
						},
					},
				},
			})

			if err != nil {
				return err
			}

			fmt.Printf("result: %+v, stickerset created check out: https://t.me/addstickers/%s\n", result, stickerSetName)

			return nil
		} else {
			return err
		}
	} else {
		fmt.Printf("StickerSet https://t.me/addstickers/%s exists\n", stickerSetName)
	}

	// time.Sleep(1 * time.Second)

	// fmt.Printf("StickerSet: %+v\n", stickerSet)
	if len(stickerSet.Stickers) == 0 {
		fmt.Println("StickerSet is empty")
	} else {
		for _, sticker := range stickerSet.Stickers {
			fmt.Println(sticker.Emoji, sticker.FileID)
			result, err := b.DeleteStickerFromSet(context.Background(), &bot.DeleteStickerFromSetParams{
				Sticker: sticker.FileID,
			})

			if err != nil {
				fmt.Println(err)
			}

			if result {
				fmt.Printf("Sticker %s deleted\n", sticker.FileID)
			}
		}
	}

	result, err := b.AddStickerToSet(context.Background(), &bot.AddStickerToSetParams{
		UserID: telegramID,
		Name:   stickerSetName,
		Sticker: models.InputSticker{
			Sticker: &models.InputFileString{
				Data: file.FileID,
			},
			Format:    "static",
			EmojiList: []string{"🚀"},
			MaskPosition: models.MaskPosition{
				Point: "forehead",
			},
		},
	})

	// result, err := b.ReplaceStickerInSet(context.Background(), &bot.ReplaceStickerInSetParams{
	// 	UserID:     telegramID,
	// 	Name:       stickerSetName,
	// 	OldSticker: stickerSet.Stickers[0].FileID,
	// 	Sticker: models.InputSticker{
	// 		Sticker: &models.InputFileString{
	// 			Data: file.FileID,
	// 		},
	// 		EmojiList: []string{"🚀"},
	// 		Format:    "static",
	// 		MaskPosition: models.MaskPosition{
	// 			Point: "forehead",
	// 		},
	// 	},
	// })

	if err != nil {
		return err
	}

	fmt.Printf("result: %+v, stickerset updated check out: https://t.me/addstickers/%s\n", result, stickerSetName)

	// _, _ = s.Bot.SendSticker(context.Background(), &bot.SendStickerParams{
	// 	ChatID: telegramID,
	// 	Sticker: &models.InputFileString{
	// 		Data: file.FileID,
	// 	},
	// })

	return nil
}
