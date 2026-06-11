package vlogger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeRuina/timberjack"
)

func New(config *LogConfig) *slog.Logger {
	config = normalizeConfig(config)

	if config.LogPath == "" && config.NoStdout {
		panic("日志路径不写，控制台日志也关闭，配置有问题，请检查")
	}

	var writers []io.Writer
	if config.LogPath != "" {
		logDir := filepath.Join(config.LogPath, config.AppName)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			panic(err)
		}
		writers = append(writers, &timberjack.Logger{
			Filename:    filepath.Join(logDir, config.AppName+"_"+config.Format+".log"),
			MaxSize:     config.MaxSize,
			MaxAge:      config.MaxAge,
			MaxBackups:  config.MaxBackups,
			LocalTime:   true,
			Compression: "gzip",
		})
	}

	if !config.NoStdout {
		writers = append(writers, os.Stdout)
	}
	writer := io.MultiWriter(writers...)

	opts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLevel(config.Level),
		ReplaceAttr: replaceAttr,
	}

	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	return NewLogger(slog.New(NewContextHandler(handler)), &LogWith{
		AppName:    config.AppName,
		AppVersion: config.AppVersion,
		Env:        config.Env,
		ID:         config.ID,
	})
}

type LogConfig struct {
	AppName    string `yaml:"app_name" json:"app_name"`       //应用名称
	AppVersion string `yaml:"app_version" json:"app_version"` //版本号
	Env        string `yaml:"env" json:"env"`                 // 环境
	ID         string `yaml:"id" json:"id"`                   // 主机ID
	LogPath    string `yaml:"log_path" json:"log_path"`       //日志路径,为空标识不写日志
	Level      string `yaml:"level" json:"level"`             //日志级别
	Format     string `yaml:"format" json:"format"`           //日志格式 json/text
	MaxSize    int    `yaml:"max_size" json:"max_size"`       //日志最大尺寸
	MaxAge     int    `yaml:"max_age" json:"max_age"`         //日志最大天数
	NoStdout   bool   `yaml:"no_stdout" json:"no_stdout"`     //是否向控制台输出,默认输出，除非设置为true
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` //最大备份数量
}

func normalizeConfig(config *LogConfig) *LogConfig {
	if config == nil {
		config = &LogConfig{}
	}
	c := *config
	if c.AppName == "" {
		c.AppName = "no_app_name"
	}
	if c.MaxSize <= 0 {
		c.MaxSize = 1024
	}
	if c.MaxAge <= 0 {
		c.MaxAge = 7
	}
	c.Format = strings.ToLower(c.Format)
	if c.Format != "text" {
		c.Format = "json"
	}
	return &c
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "panic":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
