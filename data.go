package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func getData(dataURL string) (GeckoterminalResponse, error) {
	var data GeckoterminalResponse

	err := getJson(dataURL, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func getJson(dataURL string, target interface{}) error {
	req, err := http.NewRequest("GET", dataURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return err
	}

	return nil
}
