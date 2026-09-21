package configs

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

const shortcutSlotCount = 10

type ShortcutScript struct {
	ScriptID       string   `yaml:"id" json:"id" mapstructure:"id"`
	Command        []string `yaml:"command" json:"command" mapstructure:"command"`
	EnvironmentsID []string `yaml:"environments_id" json:"environments_id" mapstructure:"environments_id"`
}

type ShortcutSlot struct {
	Key     string           `yaml:"key" json:"key" mapstructure:"key"`
	Scripts []ShortcutScript `yaml:"scripts" json:"scripts" mapstructure:"scripts"`
}

type Config struct {
	FontSize  int            `yaml:"font_size" json:"font_size"`
	Shortcuts []ShortcutSlot `yaml:"shortcuts" json:"shortcuts"`
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
	viperInstance.SetConfigType("yaml")

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return err
		}

		file, err = os.OpenFile(path, os.O_CREATE, 0644)

		value, err := yaml.Marshal(getDefaultConfig())
		if err != nil {
			return err
		}

		count, err := file.Write(value)

		if count != len(value) {
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

	globalConfig = fromViperToConfig()
	configPath = path
	initialized = true
	return nil
}

func getDefaultConfig() *Config {
	return &Config{
		FontSize:  14,
		Shortcuts: normalizeShortcuts(nil),
	}
}

func fromViperToConfig() *Config {
	var cfg Config

	cfg.FontSize = viperInstance.GetInt("font_size")
	_ = viperInstance.UnmarshalKey("shortcuts", &cfg.Shortcuts)
	cfg.Shortcuts = normalizeShortcuts(cfg.Shortcuts)
	return &cfg
}

func fromConfigToViper(cfg *Config) {
	viperInstance.Set("font_size", cfg.FontSize)
	viperInstance.Set("shortcuts", cfg.Shortcuts)
}

func normalizeShortcuts(shortcuts []ShortcutSlot) []ShortcutSlot {
	normalized := make([]ShortcutSlot, shortcutSlotCount)
	for i := range normalized {
		normalized[i] = ShortcutSlot{
			Scripts: []ShortcutScript{},
		}
	}
	for i := 0; i < len(shortcuts) && i < shortcutSlotCount; i++ {
		slot := shortcuts[i]
		scripts := make([]ShortcutScript, 0, len(slot.Scripts))
		for _, script := range slot.Scripts {
			if script.Command == nil {
				script.Command = []string{}
			}
			if script.EnvironmentsID == nil {
				script.EnvironmentsID = []string{}
			}
			scripts = append(scripts, script)
		}
		slot.Scripts = scripts
		normalized[i] = slot
	}
	return normalized
}

func GetConfig() *Config {
	return globalConfig
}

func SetConfig(cfg *Config) error {
	globalConfig = cfg
	fromConfigToViper(cfg)
	return saveConfig()
}

func saveConfig() error {
	if viperInstance == nil {
		return errors.New("viper not initialized")
	}
	return viperInstance.WriteConfig()
}
