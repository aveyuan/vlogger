package vlogger

import "log/slog"

type Recovery struct {
	logger *slog.Logger
}

func NewRecoverLog(logger *slog.Logger) *Recovery {
	return &Recovery{
		logger: logger,
	}
}

// Write 实现Recovery写入日志
func (t *Recovery) Write(p []byte) (n int, err error) {
	t.logger.Error(string(p))
	return len(p), nil
}
