package stickerUpdater

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log/slog"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/ad/anonstickerbot/config"
	"github.com/ad/anonstickerbot/sender"

	"github.com/dustin/go-humanize"
	"github.com/fogleman/gg"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
	xwebp "golang.org/x/image/webp"
)

type StickerUpdater struct {
	logger *slog.Logger
	config *config.Config
	sender *sender.Sender
	bot    *bot.Bot
}

func InitStickerUpdater(logger *slog.Logger, config *config.Config, bot *bot.Bot, sender *sender.Sender) (*StickerUpdater, error) {
	stickerUpdater := &StickerUpdater{
		logger: logger,
		config: config,
		bot:    bot,
		sender: sender,
	}

	return stickerUpdater, nil
}

func (su *StickerUpdater) Run() error {
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

	inputFile, err := os.ReadFile(imgInPath)
	if err != nil {
		return err
	}

	img, err := xwebp.Decode(bytes.NewReader(inputFile))
	if err != nil {
		return err
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

	baseTokenPriceUsd, err := strconv.ParseFloat(data.Data.Attributes.BaseTokenPriceUsd, 64)
	if err != nil {
		baseTokenPriceUsd = 0
	}

	quoteTokenPriceBaseToken, err := strconv.ParseFloat(data.Data.Attributes.QuoteTokenPriceBaseToken, 64)
	if err != nil {
		quoteTokenPriceBaseToken = 0
	}

	baseTokenPriceQuoteToken, err := strconv.ParseFloat(data.Data.Attributes.BaseTokenPriceQuoteToken, 64)
	if err != nil {
		baseTokenPriceQuoteToken = 0
	}

	dc.SetFontFace(face24)
	dc.DrawStringAnchored(
		fmt.Sprintf(
			"$%s   A%s   T%s",
			humanize.CommafWithDigits(baseTokenPriceUsd, 5),
			humanize.CommafWithDigits(quoteTokenPriceBaseToken, 2),
			humanize.CommafWithDigits(baseTokenPriceQuoteToken, 5),
		),
		256, 280, 0.5, 0)

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
	dc.DrawStringAnchored(time.Now().Format(time.RFC822), 490, 100, 1, 0.5)

	dc.SetFontFace(face32)
	dc.DrawStringAnchored("t.me/anon_club", 24, 110, 0, 0)

	mCap, err := strconv.ParseInt(data.Data.Attributes.FdvUsd, 10, 64)
	if err != nil {
		mCap = 0
	}

	dc.SetFontFace(face26)
	dc.DrawStringAnchored("$"+humanize.Comma(mCap), 490, 30, 1, 1)

	templateFileImage := dc.Image()
	if su.config.DATA_OHLCV_URL != "" {
		ohlcvData, err := getOHLCVData(su.config.DATA_OHLCV_URL)
		if err == nil {
			candles := getCandles(ohlcvData)

			imgNRGBA := image.NewNRGBA(image.Rect(0, 0, 512, 512))
			draw.Draw(imgNRGBA, imgNRGBA.Bounds(), &image.Uniform{BG_COLOR}, image.ZP, draw.Src)
			draw.Draw(imgNRGBA, templateFileImage.Bounds(), templateFileImage, templateFileImage.Bounds().Min, draw.Over)

			createAxes(
				imgNRGBA,
				candles,
				Options{
					YOffset:     300,
					Width:       512,
					Height:      512,
					CandleWidth: 6,
					Rows:        20,
					Columns:     20,
				},
			)

			templateFileImage = imgNRGBA
		}
	}

	f, err := os.Create(imgOutPath)
	if err != nil {
		return err
	}

	defer f.Close()

	err = png.Encode(f, templateFileImage)
	if err != nil {
		return err
	}

	if su.config.Debug {
		fmt.Println("-------------------------------------")
	}

	fileContent, _ := os.ReadFile(imgOutPath)

	msg, err := su.bot.SendSticker(context.Background(), &bot.SendStickerParams{
		ChatID:         -4154669576,
		ProtectContent: true,
		Emoji:          "🎱",
		Sticker: &models.InputFileUpload{
			Filename: imgOutPath,
			Data:     bytes.NewReader(fileContent),
		},
	})

	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return err
	}

	su.sender.Lock()
	defer su.sender.Unlock()

	su.sender.LastStickers["anon"] = msg.Sticker.FileID

	return nil
}
