package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	IMG_IN_PATH  = "./stickerAnon.webp"
	IMG_OUT_PATH = "./stickerpack.webp"
	DATA_URL     = "https://api.geckoterminal.com/api/v2/networks/ton/pools/"
	TOKEN        = "EQAjeq_aW_fSP7XqoF15ZZ7zUYiWLqv6UccN-jJlliomy-B3"
	INCLUDE      = "?include=dex%2Cdex.network.explorers%2Cdex_link_services%2Cnetwork_link_services%2Cpairs%2Ctoken_link_services%2Ctokens.token_security_metric%2Ctokens.tags&base_token=0"
)

type GeckoterminalResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			BaseTokenPriceUsd             string      `json:"base_token_price_usd"`
			BaseTokenPriceNativeCurrency  string      `json:"base_token_price_native_currency"`
			QuoteTokenPriceUsd            string      `json:"quote_token_price_usd"`
			QuoteTokenPriceNativeCurrency string      `json:"quote_token_price_native_currency"`
			BaseTokenPriceQuoteToken      string      `json:"base_token_price_quote_token"`
			QuoteTokenPriceBaseToken      string      `json:"quote_token_price_base_token"`
			Address                       string      `json:"address"`
			Name                          string      `json:"name"`
			PoolCreatedAt                 time.Time   `json:"pool_created_at"`
			FdvUsd                        string      `json:"fdv_usd"`
			MarketCapUsd                  interface{} `json:"market_cap_usd"`
			PriceChangePercentage         struct {
				M5  string `json:"m5"`
				H1  string `json:"h1"`
				H6  string `json:"h6"`
				H24 string `json:"h24"`
			} `json:"price_change_percentage"`
			Transactions struct {
				M5 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m5"`
				M15 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m15"`
				M30 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"m30"`
				H1 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"h1"`
				H24 struct {
					Buys    int `json:"buys"`
					Sells   int `json:"sells"`
					Buyers  int `json:"buyers"`
					Sellers int `json:"sellers"`
				} `json:"h24"`
			} `json:"transactions"`
			VolumeUsd struct {
				M5  string `json:"m5"`
				H1  string `json:"h1"`
				H6  string `json:"h6"`
				H24 string `json:"h24"`
			} `json:"volume_usd"`
			ReserveInUsd string `json:"reserve_in_usd"`
		} `json:"attributes"`
		Relationships struct {
			BaseToken struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"base_token"`
			QuoteToken struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"quote_token"`
			Dex struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"dex"`
		} `json:"relationships"`
	} `json:"data"`
}

func main() {
	data, err := getData()
	if err != nil {
		fmt.Println(err)
		return
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

	inputFile, err := os.Open(IMG_IN_PATH)
	if err != nil {
		fmt.Printf("Error loading image: %v\n", err)
		return
	}
	defer inputFile.Close()

	img, err := webp.Decode(inputFile, &decoder.Options{})
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return
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

	outputFile, err := os.Create(IMG_OUT_PATH)
	if err != nil {
		fmt.Printf("Error creating image: %v\n", err)
		return
	}
	defer outputFile.Close()

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		fmt.Printf("Error NewLossyEncoderOptions: %v\n", err)
		return
	}

	if err := webp.Encode(outputFile, dc.Image(), options); err != nil {
		fmt.Printf("Error Encode: %v\n", err)
		return
	}

	fmt.Println("-------------------------------------")
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
