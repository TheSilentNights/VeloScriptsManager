package configs

import (
	"errors"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	FontSize               int      `json:"font_size"`
	FrequentlyUsedCommands []string `json:"frequently_used_commands"`
}

var (
	globalConfig  *Config
	configPath    string
	viperInstance *viper.Viper
	configLock    sync.RWMutex
	initialized   bool
)

// InitConfig loads (or creates) the config file. A failed attempt is not
// sticky: the next call retries instead of silently reporting success.
func InitConfig(path string) error {

	if initialized {
		return nil
	}

	configLock.Lock()
	defer configLock.Unlock()

	viperInstance = viper.New()
	viperInstance.SetConfigFile(path)
	viperInstance.SetConfigType("json")

	injectDefaultValue()

	var fileNotFoundError viper.ConfigFileNotFoundError

	if err := viperInstance.ReadInConfig(); err != nil {
		if errors.As(err, &fileNotFoundError) {
			// generate config file
			viperInstance.SafeWriteConfig()
		}
	}

	var cfg Config

	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return err
	}

	globalConfig = &cfg
	configPath = path
	initialized = true
	return nil
}

func injectDefaultValue() {
	viperInstance.SetDefault("font_size", 14)
	viperInstance.SetDefault("frequently_used_commands", make([]string, 0))
}

func GetConfig() *Config {
	return globalConfig
}

func SetConfig(cfg Config) error {
	globalConfig = &cfg
	return SaveConfig()
}

func SaveConfig() error {
	return viperInstance.WriteConfig()
}
