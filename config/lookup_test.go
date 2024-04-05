package config

import (
	"os"
	"testing"
)

func TestLookupEnvOrString(t *testing.T) {
	// Test case 1: When the environment variable exists
	os.Setenv("KEY", "VALUE")
	result := lookupEnvOrString("KEY", "DEFAULT")
	if result != "VALUE" {
		t.Errorf("Expected 'VALUE', but got '%s'", result)
	}

	// Test case 2: When the environment variable does not exist
	os.Unsetenv("KEY")
	result = lookupEnvOrString("KEY", "DEFAULT")
	if result != "DEFAULT" {
		t.Errorf("Expected 'DEFAULT', but got '%s'", result)
	}
}

func TestLookupEnvOrInt(t *testing.T) {
	// Test case 1: When the environment variable exists and is a valid integer
	os.Setenv("KEY", "123")
	result := lookupEnvOrInt("KEY", 0)
	if result != 123 {
		t.Errorf("Expected 123, but got %d", result)
	}

	// Test case 2: When the environment variable exists but is not a valid integer
	os.Setenv("KEY", "abc")
	result = lookupEnvOrInt("KEY", 0)
	if result != 0 {
		t.Errorf("Expected 0, but got %d", result)
	}

	// Test case 3: When the environment variable does not exist
	os.Unsetenv("KEY")
	result = lookupEnvOrInt("KEY", 456)
	if result != 456 {
		t.Errorf("Expected 456, but got %d", result)
	}
}

func TestLookupEnvOrBool(t *testing.T) {
	// Test case 1: When the environment variable exists and is a valid boolean
	os.Setenv("KEY", "true")
	result := lookupEnvOrBool("KEY", false)
	if result != true {
		t.Errorf("Expected true, but got %t", result)
	}

	// Test case 2: When the environment variable exists but is not a valid boolean
	os.Setenv("KEY", "abc")
	result = lookupEnvOrBool("KEY", false)
	if result != false {
		t.Errorf("Expected false, but got %t", result)
	}

	// Test case 3: When the environment variable does not exist
	os.Unsetenv("KEY")
	result = lookupEnvOrBool("KEY", true)
	if result != true {
		t.Errorf("Expected true, but got %t", result)
	}
}
