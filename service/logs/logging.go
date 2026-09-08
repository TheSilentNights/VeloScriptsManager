package logs

import (
	"github/TheSilentNights/VeloScriptsManager/service/utils"

	"github.com/DeRuina/timberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	roller *timberjack.Logger
	Logger *zap.Logger
)

func InitLogger(logPath string) {
	//TODO: init logger
	roller = &timberjack.Logger{
		Filename: logPath, // Choose an appropriate path
		MaxSize:  500,     // megabytes
		// MaxBackups:         3,       // backups
		MaxAge:           28,     // days. tips: the normal February has 28 days
		Compression:      "gzip", // "none" | "gzip" | "zstd" (preferred over legacy Compress)
		LocalTime:        true,   // default: false (use UTC)
		BackupTimeFormat: "2006-01-02-15-04-05",
		FileMode:         0o644, // Custom permissions for newly created files. If unset or 0, defaults to 640.
	}
	var encoderCfg zapcore.EncoderConfig

	if utils.IsDevEnv() {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// 3. 创建 zap 的核心（Core），将日志写入 timberjack 轮转器
	var logLevel zapcore.Level

	if utils.IsDev {
		logLevel = zap.DebugLevel
	} else {
		logLevel = zap.InfoLevel
	}

	core := zapcore.NewCore(encoder, roller, logLevel)

	// 4. 创建 zap Logger
	Logger = zap.New(core, zap.AddCaller()) // zap.AddCaller() 会记录调用者的文件名和行号
}

func CloseResource() {
	Logger.Sync()
	roller.Close()
}
