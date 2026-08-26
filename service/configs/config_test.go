package configs

import (
	"testing"
)

const testConfigPath = "../temp/config.json"

func TestConfig(t *testing.T) {
	config := &Config{}

	err := config.LoadOrCreate(testConfigPath)

	if err != nil {
		panic(err)
	}
}

func TestWrite(t *testing.T) {
	config := &Config{}

	config.FontSize = 12

	err := config.SaveConfig(testConfigPath)

	if err != nil {
		panic(err)
	}
}
