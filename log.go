package vlogger

import "github.com/go-kratos/kratos/v2/log"

func New(config *LogConfig) *log.Helper {
	zapLog := NewZapLog(config)
	logger := NewLogger(zapLog, &LogWith{AppName: config.AppName, AppVersion: config.AppVersion, Env: config.Env, ID: config.ID})
	return log.NewHelper(logger)
}

type LogConfig struct {
	AppName    string `yaml:"app_name" json:"app_name"`       //应用名称
	AppVersion string `yaml:"app_version" json:"app_version"` //版本号
	Env        string `yaml:"env" json:"env"`                 // 环境
	ID         string `yaml:"id" json:"id"`                   // 主机ID
	LogPath    string `yaml:"log_path" json:"log_path"`       //日志路径
	Level      string `yaml:"level" json:"level"`             //日志级别
	MaxSize    int    `yaml:"max_size" json:"max_size"`       //日志最大尺寸
	MaxAge     int    `yaml:"max_age" json:"max_age"`         //日志最大天数
	Stdout     bool   `yaml:"stdout" json:"stdout"`           //是否向控制台输出
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` //最大备份数量
}
