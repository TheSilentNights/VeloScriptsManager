package configs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	FontSize int `json:"fontSize"`
}

func (config *Config) LoadConfig(path string) error {
	result, errConfig := os.ReadFile(path)

	if errConfig != nil {
		return errConfig
	}

	//prase json to obj
	errJson := json.Unmarshal(result, config)

	if errJson != nil {
		return errJson
	}

	return nil
}

func (config *Config) LoadOrCreate(path string) error {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)

	if err != nil {
		return err
	}

	_, osErr := os.Stat(path)
	if os.IsNotExist(osErr) {
		osCreateErr := os.WriteFile(path, []byte("{}"), 0755)

		if osCreateErr != nil {
			return osCreateErr
		}
	}

	return config.LoadConfig(path)
}

func (config *Config) SaveConfig(path string) error {

	data, errJson := json.Marshal(config)

	if errJson != nil {
		return errJson
	}

	if errOS := os.WriteFile(path, data, os.ModePerm); errOS != nil {
		return errOS
	}

	return nil

}
