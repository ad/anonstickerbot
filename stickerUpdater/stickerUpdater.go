package stickerUpdater

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/ad/anonstickerbot/config"
	"github.com/dustin/go-humanize"
	"github.com/fogleman/gg"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/golang/freetype/truetype"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/font/gofont/goregular"
)

type StickerUpdater struct {
	logger         *slog.Logger
	config         *config.Config
	bot            *bot.Bot
	stickerSetName string
	botUsername    string
}

func InitStickerUpdater(logger *slog.Logger, config *config.Config, bot *bot.Bot, stickerSetName, botUsername string) (*StickerUpdater, error) {
	stickerUpdater := &StickerUpdater{
		logger:         logger,
		config:         config,
		bot:            bot,
		stickerSetName: stickerSetName,
		botUsername:    botUsername,
	}

	return stickerUpdater, nil
}

func (su *StickerUpdater) Run() error {
	telegramID := su.config.TelegramAdminIDsList[0]
	botUsername := su.botUsername
	stickerSetName := su.stickerSetName
	dataURL := su.config.DATA_URL
	imgInPath := su.config.IMG_IN_PATH
	imgOutPath := su.config.IMG_OUT_PATH

	data, err := getData(dataURL)
	if err != nil {
		return err
	}

	if su.config.Debug {
		fmt.Println("-------------------------------------")
	}

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

	if su.config.Debug {
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
	}

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

	dc.SetRGB(1, 1, 1)
	dc.SetFontFace(face32)
	dc.DrawString(data.Data.Attributes.Name, 70, 58)

	dc.SetFontFace(face24)
	dc.DrawStringAnchored(data.Data.Attributes.QuoteTokenPriceBaseToken, 370, 54, 1, 0)

	dc.SetFontFace(face18)
	dc.DrawStringAnchored(data.Data.Attributes.BaseTokenPriceQuoteToken, 490, 54, 1, 0)

	dc.SetFontFace(face26)

	dc.DrawStringWrapped(
		fmt.Sprintf(
			"5M\n$%s\n%s/%s",
			humanize.Comma(m5Volume),
			humanize.Comma(int64(data.Data.Attributes.Transactions.M5.Buys)),
			humanize.Comma(int64(data.Data.Attributes.Transactions.M5.Sells)),
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
		65,
		166,
		0,
		0,
	)

	dc.SetRGB(1, 1, 1)

	dc.DrawStringWrapped(
		fmt.Sprintf(
			"1H\n$%s\n%s/%s",
			humanize.Comma(h1Volume),
			humanize.Comma(int64(data.Data.Attributes.Transactions.H1.Buys)),
			humanize.Comma(int64(data.Data.Attributes.Transactions.H1.Sells)),
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
			humanize.Comma(h24Volume),
			humanize.Comma(int64(data.Data.Attributes.Transactions.H24.Buys)),
			humanize.Comma(int64(data.Data.Attributes.Transactions.H24.Sells)),
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
		385,
		166,
		0,
		0,
	)

	dc.SetRGB(1, 1, 1)

	dc.SetFontFace(face18)
	dc.DrawStringAnchored(time.Now().Format(time.RFC850), 24, 490, 0, 0)

	dc.SetFontFace(face32)
	dc.DrawStringAnchored("t.me/anon_club", 24, 110, 0, 0)

	mCap, err := strconv.ParseInt(data.Data.Attributes.FdvUsd, 10, 64)
	if err != nil {
		mCap = 0
	}

	dc.SetFontFace(face26)
	dc.DrawStringAnchored("Capitalization $"+humanize.Comma(mCap), 24, 290, 0, 0)

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

	if su.config.Debug {
		fmt.Println("-------------------------------------")
	}

	fileContent, _ := os.ReadFile(imgOutPath)

	file, err := su.bot.UploadStickerFile(context.Background(), &bot.UploadStickerFileParams{
		UserID: telegramID,
		PngSticker: &models.InputFileUpload{
			Filename: imgOutPath,
			Data:     bytes.NewReader(fileContent),
		},
	})

	if err != nil {
		return err
	}

	if su.config.Debug {
		fmt.Printf("FileID: %s\n", file.FileID)
	}

	stickerSet, err := su.bot.GetStickerSet(context.Background(), &bot.GetStickerSetParams{
		Name: stickerSetName,
	})

	// time.Sleep(1 * time.Second)

	if err != nil {
		if err.Error() == "bad request, Bad Request: STICKERSET_INVALID" {
			result, err := su.bot.CreateNewStickerSet(context.Background(), &bot.CreateNewStickerSetParams{
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

			if su.config.Debug {
				fmt.Printf("result: %+v, stickerset created check out: https://t.me/addstickers/%s\n", result, stickerSetName)
			}

			return nil
		} else {
			return err
		}
	} else {
		if su.config.Debug {
			fmt.Printf("StickerSet https://t.me/addstickers/%s exists\n", stickerSetName)
		}
	}

	if len(stickerSet.Stickers) == 0 {
		if su.config.Debug {
			fmt.Println("StickerSet is empty")
		}
	} else {
		for _, sticker := range stickerSet.Stickers {
			if su.config.Debug {
				fmt.Println(sticker.Emoji, sticker.FileID)
			}
			result, err := su.bot.DeleteStickerFromSet(context.Background(), &bot.DeleteStickerFromSetParams{
				Sticker: sticker.FileID,
			})

			if err != nil {
				fmt.Println(err)
			}

			if result {
				if su.config.Debug {
					fmt.Printf("Sticker %s deleted\n", sticker.FileID)
				}
			}
		}
	}

	result, err := su.bot.AddStickerToSet(context.Background(), &bot.AddStickerToSetParams{
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

	if su.config.Debug {
		fmt.Printf("result: %+v, stickerset updated check out: https://t.me/addstickers/%s\n", result, stickerSetName)
	}

	return nil
}
