package vlogger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// ErrRecordNotFound record not found error
var ErrRecordNotFound = errors.New("record not found")

// GormToZap 映射
// GORM接口定义了接口和错误级别，在此之上，需要做级别的转换
// Zap错误级别从-1 开始 依次为 debug=-1 info=0 warn=1 error=2
// GORM的级别为 Silent =1  error = 2  warn = 3  info = 4
// GORM的debug=zapinfo Silent
// 结论，废弃，因为始终只能捕获debug信息，无用，错误的信息可以在业务层面上去拿到
var ZapToGorm map[int]int = map[int]int{-1: 4, 0: 4, 1: 3, 2: 2}

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

// Config logger config
type Config struct {
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  glogger.LogLevel
}

var (
	// Discard Discard logger will print any log to io.Discard
	Discard = NewGormLog(nil, Config{})
	// Default Default logger
	Default = NewGormLog(nil, Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  glogger.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
	// Recorder Recorder logger records running SQL into a recorder instance
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

// New initialize logger
func NewGormLog(klog *klog.Helper, config Config) glogger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	// 非zap日志下使用彩色打印
	if config.Colorful && klog == nil {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}
	// 转换config级别
	config.LogLevel = glogger.LogLevel(ZapToGorm[int(config.LogLevel)])
	return &logger{
		Writer:       log.New(os.Stdout, "\r\n", log.LstdFlags),
		klog:         klog,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	Writer
	Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
	klog                                *klog.Helper
}

// LogMode log mode
func (l *logger) LogMode(level glogger.LogLevel) glogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Info {
		if l.klog != nil {
			l.klog.WithContext(ctx).Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
			return
		}
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Warn {
		if l.klog != nil {
			l.klog.WithContext(ctx).Warnf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
			return
		}
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Error {
		if l.klog != nil {
			l.klog.WithContext(ctx).Errorf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
			return
		}
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= glogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glogger.Error && (!l.IgnoreRecordNotFoundError):
		// 对错误信息高优先级匹配
		sql, rows := fc()
		if rows == -1 {
			if l.klog != nil {
				l.klog.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				return
			}
			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			if l.klog != nil {
				l.klog.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				return
			}
			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= glogger.Warn:
		// 慢查询警告打印
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			if l.klog != nil {
				l.klog.WithContext(ctx).Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				return
			}
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			if l.klog != nil {
				l.klog.WithContext(ctx).Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				return
			}
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == glogger.Info:
		// 其余的就是info信息打印
		sql, rows := fc()
		if rows == -1 {
			if l.klog != nil {
				l.klog.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
				return
			}
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			if l.klog != nil {
				l.klog.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
				return
			}
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// Trace print sql message
func (l logger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

type traceRecorder struct {
	glogger.Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

// New new trace recorder
func (l traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: l.Interface, BeginAt: time.Now()}
}

// Trace implement logger interface
func (l *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.BeginAt = begin
	l.SQL, l.RowsAffected = fc()
	l.Err = err
}
