package configs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// resetTest 重置全局状态，确保每个测试独立
func resetTest() {
	configLock.Lock()
	defer configLock.Unlock()
	globalConfig = nil
	configPath = ""
	viperInstance = nil
	initialized = false
}

// TestInitConfig_NewFile 文件不存在时应创建默认配置并加载
func TestInitConfig_NewFile(t *testing.T) {
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	// 确保文件不存在
	_ = os.Remove(cfgFile)

	err := InitConfig(cfgFile)
	if err != nil {
		t.Fatalf("InitConfig should succeed, got: %v", err)
	}

	// 验证 globalConfig 已设置
	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	if cfg.FontSize != 14 {
		t.Errorf("expected FontSize=14, got %d", cfg.FontSize)
	}
	if len(cfg.FrequentlyUsedCommands) != 0 {
		t.Errorf("expected empty commands, got %v", cfg.FrequentlyUsedCommands)
	}

	// 验证文件确实被创建
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

// TestInitConfig_ExistingFile 文件存在且内容合法时应正确加载
func TestInitConfig_ExistingFile(t *testing.T) {
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	// 写入有效配置
	initialData := Config{
		FontSize:               20,
		FrequentlyUsedCommands: []string{"cmd1", "cmd2"},
	}
	data, _ := json.Marshal(initialData)
	if err := os.WriteFile(cfgFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	err := InitConfig(cfgFile)
	if err != nil {
		t.Fatalf("InitConfig should succeed, got: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	if cfg.FontSize != 20 {
		t.Errorf("expected FontSize=20, got %d", cfg.FontSize)
	}
	if len(cfg.FrequentlyUsedCommands) != 2 || cfg.FrequentlyUsedCommands[0] != "cmd1" {
		t.Errorf("unexpected commands: %v", cfg.FrequentlyUsedCommands)
	}
}

// TestInitConfig_InvalidJSON 文件内容格式错误时应返回错误
func TestInitConfig_InvalidJSON(t *testing.T) {
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	// 写入非法 JSON
	if err := os.WriteFile(cfgFile, []byte(`{"font_size":}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := InitConfig(cfgFile)
	if err == nil {
		t.Error("expected error due to invalid JSON, got nil")
	}

	// 确保状态未初始化，globalConfig 为 nil
	if GetConfig() != nil {
		t.Error("globalConfig should be nil on error")
	}
	if initialized {
		t.Error("initialized should be false on error")
	}
}

// TestSetConfig 应更新内存并持久化到文件
func TestSetConfig(t *testing.T) {
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	// 先初始化默认配置
	if err := InitConfig(cfgFile); err != nil {
		t.Fatal(err)
	}

	// 更新配置
	newCfg := Config{
		FontSize:               30,
		FrequentlyUsedCommands: []string{"newcmd"},
	}
	if err := SetConfig(newCfg); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// 检查内存中更新
	cfg := GetConfig()
	if cfg.FontSize != 30 || len(cfg.FrequentlyUsedCommands) != 1 {
		t.Errorf("memory config mismatch: %+v", cfg)
	}

	// 重新加载文件验证持久化
	var reloaded Config
	fileData, _ := os.ReadFile(cfgFile)
	if err := json.Unmarshal(fileData, &reloaded); err != nil {
		t.Fatal(err)
	}
	if reloaded.FontSize != 30 || reloaded.FrequentlyUsedCommands[0] != "newcmd" {
		t.Errorf("file content mismatch: %+v", reloaded)
	}
}

// TestSaveConfigWithoutInit 未初始化时调用应 panic（或返回错误，但当前实现会 panic）
func TestSaveConfigWithoutInit(t *testing.T) {
	resetTest()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic because viperInstance is nil, but got none")
		}
	}()
	// 直接调用 SaveConfig，此时 viperInstance == nil，会 panic
	_ = SaveConfig()
}

// TestConcurrentAccess 检测并发读写是否存在数据竞争（需 go test -race）
func TestConcurrentAccess(t *testing.T) {
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	if err := InitConfig(cfgFile); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	// 启动多个读 goroutine
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = GetConfig()
		}()
	}

	// 启动多个写 goroutine（修改配置）
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cfg := Config{
				FontSize:               20 + i,
				FrequentlyUsedCommands: []string{"cmd"},
			}
			_ = SetConfig(cfg) // SetConfig 会写文件，但这里只测试竞争
		}(i)
	}

	wg.Wait()
	// 若存在 data race，go test -race 会报告
}

// TestInitConfig_ErrorNotFileNotFound 模拟其他错误（此处通过只读目录模拟）
// 注意：此测试依赖系统行为，在 Linux/macOS 上有效
func TestInitConfig_ReadError(t *testing.T) {
	if os.Getenv("SKIP_PERMISSION_TEST") != "" {
		t.Skip("Skipping permission test")
	}
	resetTest()
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")

	// 创建文件
	if err := os.WriteFile(cfgFile, []byte(`{"font_size":10}`), 0644); err != nil {
		t.Fatal(err)
	}
	// 将文件权限设为只读，导致 viper 读取失败（在某些系统上）
	if err := os.Chmod(cfgFile, 0000); err != nil {
		t.Skip("无法修改文件权限，跳过此测试")
	}
	defer os.Chmod(cfgFile, 0644) // 恢复权限以便清理

	err := InitConfig(cfgFile)
	// 期望返回错误（非文件不存在错误）
	if err == nil {
		t.Error("expected error due to permission, got nil")
	}
	if initialized {
		t.Error("initialized should remain false on error")
	}
}
