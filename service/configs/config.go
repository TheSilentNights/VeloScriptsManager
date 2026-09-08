package configs

import (
	"errors"
	"os"
	"path/filepath"
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

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return err
		}

		file, err = os.OpenFile(path, os.O_CREATE, 0644)
		count, err := file.Write([]byte("{}"))

		if count != 2 {
			return errors.New("写入配置文件失败: 配置文件内容为空")
		}

		if err != nil {
			return err
		}
		file.Close()
	}

	if err := viperInstance.ReadInConfig(); err != nil {
		return err
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
	viperInstance.SetDefault("font_size", 12)
}

func GetConfig() *Config {
	return globalConfig
}

func SetConfig(cfg Config) error {
	globalConfig = &cfg
	return SaveConfig()
}

func SaveConfig() error {
	if viperInstance == nil {
		return errors.New("viper not initialized")
	}
	return viperInstance.WriteConfig()
}
