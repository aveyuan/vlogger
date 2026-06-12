package vlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewJSONLogger(t *testing.T) {
	logDir := t.TempDir()
	logger := New(&LogConfig{
		AppName:    "app",
		AppVersion: "v1",
		Env:        "test",
		ID:         "node-1",
		LogPath:    logDir,
		Level:      "debug",
	})

	ctx := WithRequestID(context.Background(), "rid")
	logger.InfoContext(ctx, "hello")

	data, err := os.ReadFile(filepath.Join(logDir, "app", "app_json.log"))
	if err != nil {
		t.Fatal(err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["msg"] != "hello" {
		t.Fatalf("msg = %v", entry["msg"])
	}
	if entry["appName"] != "app" || entry["version"] != "v1" || entry["env"] != "test" || entry["id"] != "node-1" {
		t.Fatalf("missing app fields: %#v", entry)
	}
	if entry["datetime"] == "" || entry["lineNumber"] == "" {
		t.Fatalf("missing datetime or lineNumber: %#v", entry)
	}
	if _, ok := entry["ext"]; !ok {
		t.Fatalf("missing ext: %#v", entry)
	}
	if entry["requestId"] != "rid" {
		t.Fatalf("requestId = %v", entry["requestId"])
	}
}

func TestNewTextLogger(t *testing.T) {
	logDir := t.TempDir()
	logger := New(&LogConfig{
		AppName:  "app",
		LogPath:  logDir,
		Format:   "text",
		NoStdout: true,
	})

	logger.Info("hello")

	data, err := os.ReadFile(filepath.Join(logDir, "app", "app_text.log"))
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, `msg=hello`) || !strings.Contains(output, `appName=app`) {
		t.Fatalf("unexpected text output: %s", output)
	}
}

func TestLevelFiltering(t *testing.T) {
	logDir := t.TempDir()
	logger := New(&LogConfig{
		AppName:  "app",
		LogPath:  logDir,
		Level:    "error",
		NoStdout: true,
	})

	logger.Warn("hidden")
	logger.Error("visible")

	data, err := os.ReadFile(filepath.Join(logDir, "app", "app_json.log"))
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if strings.Contains(output, "hidden") || !strings.Contains(output, "visible") {
		t.Fatalf("unexpected level output: %s", output)
	}
}

func TestContextHandlerRequestIDFromExt(t *testing.T) {
	var buf bytes.Buffer
	handler := NewContextHandler(slog.NewJSONHandler(&buf, nil))
	logger := slog.New(handler)

	logger.InfoContext(context.Background(), "hello")

	if !strings.Contains(buf.String(), `"requestId":"rid"`) {
		t.Fatalf("missing requestId output: %s", buf.String())
	}
}
