package configs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	FontSize int `json:"fontSize"`
}

var (
	globalConfig *Config
	configPath   string
	configMu     sync.RWMutex
	initialized  bool
)

// InitConfig loads (or creates) the config file. A failed attempt is not
// sticky: the next call retries instead of silently reporting success.
func InitConfig(path string) error {
	configMu.Lock()
	defer configMu.Unlock()

	if initialized {
		return nil
	}

	cfg := &Config{}
	if err := cfg.LoadOrCreate(path); err != nil {
		return err
	}

	globalConfig = cfg
	configPath = path
	initialized = true
	return nil
}

func GetConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return Config{}
	}

	return *globalConfig
}

func SetConfig(c Config) error {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		return os.ErrInvalid
	}

	*globalConfig = c
	return globalConfig.SaveConfig(configPath)
}

func Update(fn func(*Config)) error {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		return os.ErrInvalid
	}

	//fn updates the config
	fn(globalConfig)

	return globalConfig.SaveConfig(configPath)
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
