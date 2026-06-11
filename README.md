# vlog日志组件
提供日志组件，基于 Go 官方 slog 实现，提供日志级别、日志格式、日志路径、日志最大尺寸、日志最大天数、日志最大备份数量等配置。

封装了GORM，kafka相关日志实现，其他可以自行扩展。

## 实现逻辑

### 初始化流程

`New(config *LogConfig)` 是日志入口，返回 `*slog.Logger`。

1. 先通过 `normalizeConfig` 补齐默认值：
   - `app_name` 为空时使用 `no_app_name`
   - `max_size` 默认 `1024`
   - `max_age` 默认 `7`
   - `format` 支持 `json` 和 `text`，其他值默认按 `json` 处理
2. 如果 `log_path` 不为空，会创建目录：
   - `log_path/app_name`
3. 文件日志使用 `github.com/DeRuina/timberjack` 做切割：
   - 文件名为 `app_name_format.log`
   - 按 `max_size`、`max_age`、`max_backups` 控制切割和保留
   - 归档压缩格式为 `gzip`
4. 如果 `no_stdout=false`，会同时输出到控制台。
5. 文件和控制台通过 `io.MultiWriter` 组合，所以可以同时写入。
6. 根据 `format` 创建 slog handler：
   - `json` 使用 `slog.NewJSONHandler`
   - `text` 使用 `slog.NewTextHandler`
7. 外层包一层 `ContextHandler`，用于从 `context.Context` 注入扩展字段。

### 日志字段

初始化时会通过 `logger.With` 固定写入应用信息：

- `appName`
- `version`
- `env`
- `id`

slog 默认字段会经过 `ReplaceAttr` 转换：

- `time` 转为 `datetime`，格式为 `2006-01-02 15:04:05.000`
- `source` 转为 `lineNumber`，格式为 `file.go:line`

最终 JSON 日志示例：

```json
{"datetime":"2026-06-11 11:45:30.778","level":"INFO","lineNumber":"logger_test.go:29","msg":"hello","appName":"app","version":"v1","env":"test","id":"node-1","requestId":"rid","ext":{"request_id":"rid"}}
```

### 日志级别

`level` 配置映射到 slog 级别：

- `debug` 或未知值：`slog.LevelDebug`
- `info`：`slog.LevelInfo`
- `warn` / `warning`：`slog.LevelWarn`
- `error`：`slog.LevelError`
- `panic`：按 `error` 阈值处理

### 上下文字段

组件支持从 `context.Context` 读取扩展字段。

单独写入 request id：

```go
ctx := vlogger.WithRequestID(context.Background(), "rid")
logger.InfoContext(ctx, "hello")
```

写入扩展字段和 request id：

```go
ctx := vlogger.WithExt(context.Background(), &vlogger.ExtLogValue{
    RequestID: "rid",
    Ext: map[string]string{
        "request_id": "rid",
    },
})
logger.InfoContext(ctx, "hello")
```

如果同时使用 `WithRequestID` 和 `WithExt.RequestID`，`WithRequestID` 的值优先。

### 配置示例

```go
logger := vlogger.New(&vlogger.LogConfig{
    AppName:    "app",
    AppVersion: "v1",
    Env:        "prod",
    ID:         "node-1",
    LogPath:    "./logs",
    Level:      "info",
    Format:     "json",
    MaxSize:    1024,
    MaxAge:     7,
    MaxBackups: 10,
    NoStdout:   false,
})

logger.Info("service started")
```
