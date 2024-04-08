package stickerUpdater

import (
	"image"
	"image/color"
	"slices"
	"strconv"
	"time"
)

var (
	POSITIVE_COLOR = color.RGBA{R: 126, G: 211, B: 33, A: 255}
	NEGATIVE_COLOR = color.RGBA{R: 208, G: 2, B: 27, A: 255}
	PIKE_COLOR     = color.RGBA{R: 211, G: 211, B: 211, A: 255}
)

type GeckoterminalOHLCVResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			OhlcvList [][]float64 `json:"ohlcv_list"`
		} `json:"attributes"`
	} `json:"data"`
	Meta struct {
		Base struct {
			Address         string `json:"address"`
			Name            string `json:"name"`
			Symbol          string `json:"symbol"`
			CoingeckoCoinID any    `json:"coingecko_coin_id"`
		} `json:"base"`
		Quote struct {
			Address         string `json:"address"`
			Name            string `json:"name"`
			Symbol          string `json:"symbol"`
			CoingeckoCoinID any    `json:"coingecko_coin_id"`
		} `json:"quote"`
	} `json:"meta"`
}

type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func getOHLCVData(dataURL string) (GeckoterminalOHLCVResponse, error) {
	var data GeckoterminalOHLCVResponse

	err := getJson(dataURL, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func getCandles(data GeckoterminalOHLCVResponse) []Candle {
	var ohlcvData []Candle

	for _, ohlcv := range data.Data.Attributes.OhlcvList {
		ohlcvData = append(ohlcvData, Candle{
			Time:   time.Unix(int64(ohlcv[0]), 0),
			Open:   ohlcv[1],
			High:   ohlcv[2],
			Low:    ohlcv[3],
			Close:  ohlcv[4],
			Volume: ohlcv[5],
		})
	}

	slices.Reverse(ohlcvData)

	return ohlcvData
}

func (c Candle) getColor() color.RGBA {
	if c.Open > c.Close {
		return NEGATIVE_COLOR
	}

	return POSITIVE_COLOR
}

type Options struct {
	Width       int
	Height      int
	YOffset     int
	CandleWidth int
	Columns     int
	Rows        int
}

func createAxes(img *image.NRGBA, data []Candle, opts Options) {
	var (
		higherValue       float64
		lowerValue        float64
		chartHmin         int
		chartHmax         int
		startTimePosition int
		endTimePosition   int
		startTime         time.Time
		timeDiff          time.Duration
	)

	chartHmin = opts.Height
	chartHmax = 5 + opts.YOffset
	chartWmax := opts.Width - 5

	var max float64
	var min float64
	var maxLength int
	for i, d := range data {
		if i == 0 {
			max = d.High
			min = d.Low
		} else {
			if d.High > max {
				max = d.High
			}
			if d.Low < min {
				min = d.Low
			}
		}
		if len(strconv.Itoa(int(d.High))) > maxLength {
			maxLength = len(strconv.Itoa(int(d.High)))
		}
		if len(strconv.Itoa(int(d.Low))) > maxLength {
			maxLength = len(strconv.Itoa(int(d.Low)))
		}
	}

	chartSeparation := maxLength
	chartLineWidth := 5

	diff := ((max - min) * 5) / 100
	higherValue = max + diff
	lowerValue = min - diff

	startTime = data[0].Time
	endTime := data[len(data)-1].Time

	timeDiff = endTime.Sub(startTime)

	separation := (chartWmax - chartSeparation) / opts.Columns
	for i := 0; i < opts.Columns+1; i++ {
		xPosition := chartSeparation + chartLineWidth + separation*i
		if i == 0 {
			startTimePosition = xPosition
		} else if i == opts.Columns {
			endTimePosition = xPosition
		}
	}

	for _, d := range data {
		t := d.Time
		newXPosition := getXPointInChart(t, startTime, timeDiff, startTimePosition, endTimePosition)
		candleColor := d.getColor()
		candleHighYPosition := getYPointInChart(d.High, lowerValue, higherValue, chartHmin, chartHmax)
		candleOpenYpoint := getYPointInChart(d.Open, lowerValue, higherValue, chartHmin, chartHmax)
		candleCloseYpoint := getYPointInChart(d.Close, lowerValue, higherValue, chartHmin, chartHmax)
		candleLowYpoint := getYPointInChart(d.Low, lowerValue, higherValue, chartHmin, chartHmax)

		if candleColor == POSITIVE_COLOR {
			aux := candleCloseYpoint
			candleCloseYpoint = candleOpenYpoint
			candleOpenYpoint = aux

		}

		line(newXPosition, newXPosition, candleHighYPosition, candleOpenYpoint, PIKE_COLOR, img)
		halfCandleWidth := opts.CandleWidth / 2
		line(newXPosition-halfCandleWidth, newXPosition+halfCandleWidth, candleOpenYpoint, candleCloseYpoint, candleColor, img)
		line(newXPosition, newXPosition, candleCloseYpoint, candleLowYpoint, PIKE_COLOR, img)
	}
}

func line(x1, x2, y1, y2 int, col color.RGBA, img *image.NRGBA) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			img.Set(x, y, col)
		}
	}
}

func getYPointInChart(value, lowerValue, higherValue float64, chartHmin, chartHmax int) int {
	ypercent := ((value - lowerValue) * 100) / (higherValue - lowerValue)
	auxYPoint := float64(chartHmin-chartHmax) * (float64(ypercent) / 100)
	newYPoint := chartHmin - int(auxYPoint)
	return int(newYPoint)
}

func getXPointInChart(value, startTime time.Time, timeDiff time.Duration, startTimePosition, endTimePosition int) int {
	ypercent := ((value.Sub(startTime)) * 100) / (timeDiff)
	auxYPoint := float64(endTimePosition-startTimePosition) * (float64(ypercent) / 100)
	newYPoint := startTimePosition + int(auxYPoint)
	return int(newYPoint)
}
