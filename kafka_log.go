package vlogger

import "github.com/go-kratos/kratos/v2/log"

type KafkaErrorLog struct {
	logger *log.Helper
}

func NewKafkaErrorLog(logger *log.Helper) *KafkaErrorLog {
	return &KafkaErrorLog{logger: logger}
}

func (t *KafkaErrorLog) Printf(format string, v ...interface{}) {
	t.logger.Errorf(format, v...)
}

type KafkaInfoLog struct {
	logger *log.Helper
}

func NewKafkaInfoLog(logger *log.Helper) *KafkaInfoLog {
	return &KafkaInfoLog{logger: logger}
}

func (t *KafkaInfoLog) Printf(format string, v ...interface{}) {
	t.logger.Infof(format, v...)
}
