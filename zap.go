package vlogger

import (
	"fmt"
	"os"
	"path/filepath"

	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var LogStr2Level = map[string]int8{
	"debug": -1,
	"info":  0,
	"warn":  1,
	"error": 2,
	"panic": 3,
}

// 初始化日志
func NewZapLog(LogConfig *LogConfig) *kzap.Logger {
	if LogConfig.AppName == "" {
		LogConfig.AppName = "no_app_name"
	}
	if LogConfig.LogPath == "" {
		LogConfig.LogPath = "./logs"
	}
	if LogConfig.MaxSize <= 0 {
		LogConfig.MaxSize = 1024
	}
	if LogConfig.MaxAge <= 0 {
		LogConfig.MaxAge = 7
	}

	// 根据配置类引入
	lum := lumberjack.Logger{
		Filename:   filepath.Join(LogConfig.LogPath, LogConfig.AppName, fmt.Sprintf("%v_json.log", LogConfig.AppName)), // 日志文件路径
		MaxSize:    LogConfig.MaxSize,                                                                                  // 每个日志文件保存的大小 单位:M
		MaxAge:     LogConfig.MaxAge,                                                                                   // 文件最多保存多少天
		MaxBackups: LogConfig.MaxBackups,                                                                               // 日志文件最多保存多少个备份
		LocalTime:  true,                                                                                               // 本地时间
		Compress:   true,                                                                                               // 是否压缩
	}

	// 设置日志级别
	l, ok := LogStr2Level[LogConfig.Level]
	if !ok {
		l = -1
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(l))
	var writes = []zapcore.WriteSyncer{zapcore.AddSync(&lum)}
	// 是否向控制台输出
	if LogConfig.Stdout {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			LevelKey:       "level",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.NewMultiWriteSyncer(writes...),
		atomicLevel,
	)

	return kzap.NewLogger(zap.New(core))
}
