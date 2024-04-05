package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func getData() (GeckoterminalResponse, error) {
	var data GeckoterminalResponse

	err := getJson(DATA_URL, TOKEN, INCLUDE, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func getJson(url, token, include string, target interface{}) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s%s", url, token, include), nil)
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
