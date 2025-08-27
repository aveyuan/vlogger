package vlogger

import "github.com/go-kratos/kratos/v2/log"

type Recovery struct {
	logger *log.Helper
}

func NewRecoverLog(logger *log.Helper) *Recovery {
	return &Recovery{
		logger: logger,
	}
}

// Write 实现Recovery写入日志
func (t *Recovery) Write(p []byte) (n int, err error) {
	t.logger.Error(string(p))
	return 0, nil
}
