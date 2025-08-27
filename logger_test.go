package vlogger

import (
	"testing"
)

func TestZap(t *testing.T) {
	log := New(&LogConfig{})
	log.Info("hhx")

}
