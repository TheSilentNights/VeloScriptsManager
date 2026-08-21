package test

import (
	"github/TheSilentNights/VeloScriptsManager/service/configs"
	"testing"
)

const testConfigPath = "../temp/config.json"

func TestConfig(t *testing.T) {
	config := &configs.Config{}

	err := config.LoadOrCreate(testConfigPath)

	if err != nil {
		panic(err)
	}
}

func TestWrite(t *testing.T) {
	config := &configs.Config{}

	config.FontSize = 12

	err := config.SaveConfig(testConfigPath)

	if err != nil {
		panic(err)
	}
}
