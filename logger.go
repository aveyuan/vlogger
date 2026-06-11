package vlogger

import (
	"context"
	"log/slog"
)

type LogWith struct {
	AppName    string //应用名称
	AppVersion string //版本号
	Env        string // 环境
	ID         string // ID
}

const (
	ExtLogKey       = "extlogkey"
	RequestIDLogKey = "requestidlogkey"
)

type ExtLogValue struct {
	Ext       interface{}
	RequestID string
	ctx       context.Context
}

func NewLogger(lg *slog.Logger, with *LogWith) *slog.Logger {
	if lg == nil {
		lg = slog.Default()
	}
	if with == nil {
		return lg
	}
	return lg.With(
		"appName", with.AppName,
		"version", with.AppVersion,
		"env", with.Env,
		"id", with.ID,
	)
}

func WithExt(ctx context.Context, ExtLogValue *ExtLogValue) context.Context {
	if ctx == nil {
		return nil
	}
	ExtLogValue.ctx = ctx

	return context.WithValue(ctx, ExtLogKey, ExtLogValue)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, RequestIDLogKey, requestID)
}


// slog handler实现
type contextHandler struct {
	handler slog.Handler
}

func NewContextHandler(handler slog.Handler) slog.Handler {
	return &contextHandler{handler: handler}
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	requestID := ""
	if ext, ok := ctx.Value(ExtLogKey).(*ExtLogValue); ok {
		record.AddAttrs(slog.Any("ext", ext.Ext))
		requestID = ext.RequestID
	}
	if value, ok := ctx.Value(RequestIDLogKey).(string); ok && value != "" {
		requestID = value
	}
	if requestID != "" {
		record.AddAttrs(slog.String("requestId", requestID))
	}
	return h.handler.Handle(ctx, record)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{handler: h.handler.WithGroup(name)}
}
