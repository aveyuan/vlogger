package vlogger

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type LogWith struct {
	AppName    string //应用名称
	AppVersion string //版本号
	Env        string // 环境
	ID         string // ID
}

const ExtLogKey = "extlogkey"

type ExtLogValue struct {
	Ext interface{}
	ctx context.Context
}

func NewLogger(lg log.Logger, with *LogWith) log.Logger {
	return log.With(lg,
		"datetime", log.Timestamp("2006-01-02 15:04:05.000"),
		// "env", with.Env,
		// "appName", with.AppName,
		// "version", with.AppVersion,
		// "id", with.ID,
		// "traceID", TraceID(),
		// "spanID", SpanID(),
		"lineNumber", log.DefaultCaller,
		"ext", Ext(),
	)
}

func WithExt(ctx context.Context, ExtLogValue *ExtLogValue) context.Context {
	if ctx == nil {
		return nil
	}
	ExtLogValue.ctx = ctx

	return context.WithValue(ctx, ExtLogKey, ExtLogValue)
}

func Ext() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ext, ok := ctx.Value(ExtLogKey).(*ExtLogValue); ok {
			return ext.Ext
		}
		return nil
	}
}
