package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

const ConfigFileName = "/data/options.json"

// Config ...
type Config struct {
	TelegramToken        string  `json:"TELEGRAM_TOKEN"`
	TelegramAdminIDs     string  `json:"TELEGRAM_ADMIN_IDS"`
	TelegramAdminIDsList []int64 `json:"-"`

	IMG_IN_PATH  string `json:"IMG_IN_PATH"`
	IMG_OUT_PATH string `json:"IMG_OUT_PATH"`
	DATA_URL     string `json:"DATA_URL"`

	UPDATE_DELAY int `json:"UPDATE_DELAY"`

	Debug bool `json:"DEBUG"`
}

func InitConfig(args []string) (*Config, error) {
	var config = &Config{
		TelegramToken:        "",
		TelegramAdminIDs:     "",
		TelegramAdminIDsList: []int64{},

		Debug: false,
	}

	var initFromFile = false

	if _, err := os.Stat(ConfigFileName); err == nil {
		jsonFile, err := os.Open(ConfigFileName)
		if err == nil {
			byteValue, _ := io.ReadAll(jsonFile)
			if err = json.Unmarshal(byteValue, &config); err == nil {
				initFromFile = true
			} else {
				fmt.Printf("error on unmarshal config from file %s\n", err.Error())
			}
		}
	}

	if !initFromFile {
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.StringVar(&config.TelegramToken, "telegramToken", lookupEnvOrString("TELEGRAM_TOKEN", config.TelegramToken), "TELEGRAM_TOKEN")
		flags.StringVar(&config.TelegramAdminIDs, "telegramAdminIDs", lookupEnvOrString("TELEGRAM_ADMIN_IDS", config.TelegramAdminIDs), "TELEGRAM_ADMIN_IDS")

		flags.StringVar(&config.IMG_IN_PATH, "imgInPath", lookupEnvOrString("IMG_IN_PATH", config.IMG_IN_PATH), "IMG_IN_PATH")
		flags.StringVar(&config.IMG_OUT_PATH, "imgOutPath", lookupEnvOrString("IMG_OUT_PATH", config.IMG_OUT_PATH), "IMG_OUT_PATH")
		flags.StringVar(&config.DATA_URL, "dataUrl", lookupEnvOrString("DATA_URL", config.DATA_URL), "DATA_URL")

		flags.IntVar(&config.UPDATE_DELAY, "updateDelay", lookupEnvOrInt("UPDATE_DELAY", config.UPDATE_DELAY), "UPDATE_DELAY")

		flags.BoolVar(&config.Debug, "debug", lookupEnvOrBool("DEBUG", config.Debug), "Debug")

		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
	}

	if config.TelegramAdminIDs != "" {
		chatIDS := strings.Split(config.TelegramAdminIDs, ",")
		for _, chatID := range chatIDS {
			if chatIDInt, err := strconv.ParseInt(strings.Trim(chatID, "\n\t "), 10, 64); err == nil {
				config.TelegramAdminIDsList = append(config.TelegramAdminIDsList, chatIDInt)
			}
		}
	}

	return config, nil
}
