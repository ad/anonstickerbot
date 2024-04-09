package stickerUpdater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ad/anonstickerbot/config"
	"github.com/ad/anonstickerbot/sender"

	"github.com/dustin/go-humanize"
	"github.com/fogleman/gg"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/golang/freetype/truetype"
	"github.com/nickalie/go-webpbin"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/webp"
)

type StickerUpdater struct {
	logger   *slog.Logger
	config   *config.Config
	sender   *sender.Sender
	bot      *bot.Bot
	stickers map[string]*StickerConfig
}

type StickerConfig struct {
	Name    string      `json:"name"`
	Address string      `json:"address"`
	Emoji   string      `json:"emoji"`
	image   image.Image `json:"-"`
}

func InitStickerUpdater(logger *slog.Logger, config *config.Config, bot *bot.Bot, sender *sender.Sender) (*StickerUpdater, error) {
	stickerUpdater := &StickerUpdater{
		logger:   logger,
		config:   config,
		bot:      bot,
		sender:   sender,
		stickers: make(map[string]*StickerConfig),
	}

	// read directory with tokens and load configs from json
	dirs, err := os.ReadDir(config.TOKENS_PATH)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		file, err := os.ReadFile(fmt.Sprintf("%s/%s/info.json", config.TOKENS_PATH, dir.Name()))
		if err != nil {
			continue
		}

		stickerConfig := &StickerConfig{}
		if err := json.Unmarshal(file, stickerConfig); err != nil {
			continue
		}

		webpFile, err := os.ReadFile(fmt.Sprintf("%s/%s/sticker.webp", config.TOKENS_PATH, dir.Name()))
		if err != nil {
			continue
		}

		inputFile, err := webp.Decode(bytes.NewReader(webpFile))
		if err != nil {
			continue
		}

		stickerConfig.image = inputFile

		stickerUpdater.stickers[stickerConfig.Name] = stickerConfig
	}

	fmt.Printf("stickerUpdater.stickers: %d\n", len(stickerUpdater.stickers))

	return stickerUpdater, nil
}

func (su *StickerUpdater) Run(name string) error {
	stickerConfig, ok := su.stickers[name]
	if !ok {
		return fmt.Errorf("sticker with name %q not found", name)
	}

	return su.updateSticker(stickerConfig)
}

func (su *StickerUpdater) RunAll() error {
	for _, stickerConfig := range su.stickers {
		if err := su.updateSticker(stickerConfig); err != nil {
			return err
		}
	}

	return nil
}

func (su *StickerUpdater) updateSticker(stickerConfig *StickerConfig) error {
	dataURL := strings.Replace(su.config.DATA_URL, "%s", stickerConfig.Address, 1)

	data, err := getData(dataURL)
	if err != nil {
		return fmt.Errorf("%s:%s getData error: %w (%s)", stickerConfig.Name, stickerConfig.Address, err, dataURL)
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

	dc := gg.NewContextForImage(stickerConfig.image)

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

	mCap, err := strconv.ParseInt(data.Data.Attributes.FdvUsd, 10, 64)
	if err != nil {
		mCap = 0
	}

	dc.SetFontFace(face26)
	dc.DrawStringAnchored("$"+humanize.Comma(mCap), 490, 30, 1, 1)

	templateFileImage := dc.Image()
	if su.config.DATA_OHLCV_URL != "" {
		dataOhlcvURL := strings.Replace(su.config.DATA_OHLCV_URL, "%s", stickerConfig.Address, 1)
		ohlcvData, err := getOHLCVData(dataOhlcvURL)
		if err == nil {
			candles := getCandles(ohlcvData)
			if len(candles) > 0 {
				imgNRGBA := image.NewNRGBA(image.Rect(0, 0, 512, 512))
				draw.Draw(imgNRGBA, templateFileImage.Bounds(), templateFileImage, image.Point{0, 0}, draw.Over)

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
	}

	buf := new(bytes.Buffer)

	if err := webpbin.Encode(buf, templateFileImage); err != nil {
		return err
	}

	if su.config.Debug {
		fmt.Println("-------------------------------------")
	}

	msg, err := su.bot.SendSticker(context.Background(), &bot.SendStickerParams{
		ChatID:         su.config.TelegramTargetChatID,
		ProtectContent: false,
		Emoji:          stickerConfig.Emoji,
		Sticker: &models.InputFileUpload{
			Filename: "sticker.webp",
			Data:     bytes.NewReader(buf.Bytes()),
		},
	})

	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return err
	}

	su.sender.Lock()
	defer su.sender.Unlock()

	su.sender.LastStickers[stickerConfig.Name] = msg.Sticker.FileID

	return nil
}
