package config

import (
	"os"
	"testing"
)

func TestInitConfig(t *testing.T) {
	// Test case 1: When the config file exists and can be successfully read
	_, _ = os.Create(ConfigFileName)
	defer os.Remove(ConfigFileName)

	_, err := InitConfig([]string{""})
	if err != nil {
		t.Errorf("Expected no error, but got %v", err.Error())
	}

	// Test case 2: When the config file does not exist
	os.Remove(ConfigFileName)

	_, err = InitConfig([]string{""})
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test case 3: When environment variables are provided
	os.Setenv("TELEGRAM_TOKEN", "token123")
	os.Setenv("TELEGRAM_ADMIN_IDS", "7890")

	config, err := InitConfig([]string{""})
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if config.TelegramToken != "token123" {
		t.Errorf("Expected TelegramToken to be 'token123', but got '%s'", config.TelegramToken)
	}
	if config.TelegramAdminIDsList[0] != 7890 {
		t.Errorf("Expected TelegramAdmin to be '7890', but got '%s'", config.TelegramAdminIDs)
	}

	// Clean up environment variables
	os.Unsetenv("TELEGRAM_TOKEN")
	os.Unsetenv("TELEGRAM_ADMIN_ID")
}
