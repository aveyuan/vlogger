package vlogger

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"
)

func replaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}
	switch attr.Key {
	case slog.TimeKey:
		if t, ok := attr.Value.Any().(time.Time); ok {
			return slog.String("datetime", t.Format("2006-01-02 15:04:05.000"))
		}
	case slog.SourceKey:
		if source, ok := attr.Value.Any().(*slog.Source); ok && source != nil {
			return slog.String("lineNumber", fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line))
		}
	}
	return attr
}
