package multiplex

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Level

type LogLevel uint8

const (
	LogLevelDebug LogLevel = iota + 1
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

func MapLogLevel(level LogLevel) string {
	return levelMapper[level]
}

var levelMapper = map[LogLevel]string{
	LogLevelDebug: "debug",
	LogLevelInfo:  "info",
	LogLevelWarn:  "warn",
	LogLevelError: "error",
	LogLevelFatal: "fatal",
}

// Logger

type Logger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})

	ShallowCopy() Logger
	SetLogLevel(level LogLevel)
	WithCallDepth(externalDepth uint) Logger
	WithMessageId(msgId string) Logger
}

var (
	defaultLogger Logger = NewLogger(0, LogLevelDebug)
)

func SetDefaultLogger(l Logger) {
	defaultLogger = l
}

func DefaultLogger() Logger {
	return defaultLogger
}

func NewLogger(externalDepth uint, level LogLevel) stdLogger {
	flag := log.Ltime | log.LUTC | log.Lmsgprefix | log.Lshortfile
	internalDepth := 2
	return stdLogger{
		debug:             log.New(os.Stdout, "[Debug]: ", flag),
		info:              log.New(os.Stdout, "[Info]: ", flag),
		warn:              log.New(os.Stdout, "[Warn]: ", flag),
		err:               log.New(os.Stdout, "[Error]: ", flag),
		fatal:             log.New(os.Stdout, "[Fatal]: ", flag),
		internalCallDepth: internalDepth + int(externalDepth),
		logLevel:          &level,
	}
}

type stdLogger struct {
	debug             *log.Logger
	info              *log.Logger
	warn              *log.Logger
	err               *log.Logger
	fatal             *log.Logger
	internalCallDepth int
	logLevel          *LogLevel

	msgId string
}

func (l stdLogger) Debug(format string, a ...interface{}) {
	if *l.logLevel > LogLevelDebug {
		return
	}
	format = l.processPreformat(format)
	l.debug.Output(l.internalCallDepth, fmt.Sprintf(format, a...))
}

func (l stdLogger) Info(format string, a ...interface{}) {
	if *l.logLevel > LogLevelInfo {
		return
	}
	format = l.processPreformat(format)
	l.info.Output(l.internalCallDepth, fmt.Sprintf(format, a...))
}

func (l stdLogger) Warn(format string, a ...interface{}) {
	if *l.logLevel > LogLevelWarn {
		return
	}
	format = l.processPreformat(format)
	l.warn.Output(l.internalCallDepth, fmt.Sprintf(format, a...))
}

func (l stdLogger) Error(format string, a ...interface{}) {
	if *l.logLevel > LogLevelError {
		return
	}
	format = l.processPreformat(format)
	l.err.Output(l.internalCallDepth, fmt.Sprintf(format, a...))
}

func (l stdLogger) Fatal(format string, a ...interface{}) {
	if *l.logLevel > LogLevelFatal {
		return
	}
	format = l.processPreformat(format)
	l.fatal.Output(l.internalCallDepth, fmt.Sprintf(format, a...))
	os.Exit(1)
}

func (l stdLogger) ShallowCopy() Logger {
	return l.shallowCopy()
}

// shallowCopy 每次加上新的欄位, 要記得修改 shallowCopy 內部實現方式
func (l stdLogger) shallowCopy() stdLogger {
	return stdLogger{
		debug:             l.debug,
		info:              l.info,
		warn:              l.warn,
		err:               l.err,
		fatal:             l.fatal,
		internalCallDepth: l.internalCallDepth,
		logLevel:          l.logLevel,
		msgId:             l.msgId,
	}
}

func (l stdLogger) SetLogLevel(level LogLevel) {
	*l.logLevel = level
}

func (l stdLogger) LogLevel() LogLevel {
	return *l.logLevel
}

func (l stdLogger) WithCallDepth(externalDepth uint) Logger {
	l2 := l.shallowCopy()
	l2.internalCallDepth += int(externalDepth)
	return l2
}

func (l stdLogger) processPreformat(format string) string {
	b := strings.Builder{}

	if l.msgId != "" {
		b.WriteString("msg_id=")
		b.WriteString(l.msgId)
		b.WriteString(" ")
	}

	b.WriteString(format)
	return b.String()
}

func (l stdLogger) WithMessageId(msgId string) Logger {
	if msgId == "" {
		return l
	}

	l2 := l.shallowCopy()
	l2.msgId = msgId
	return l2
}

//

func SilentLogger() Logger {
	return silentLogger{}
}

type silentLogger struct{}

func (silentLogger) Debug(format string, a ...interface{}) {
	return
}

func (silentLogger) Info(format string, a ...interface{}) {
	return
}

func (silentLogger) Warn(format string, a ...interface{}) {
	return
}

func (silentLogger) Error(format string, a ...interface{}) {
	return
}

func (silentLogger) Fatal(format string, a ...interface{}) {
	return
}

func (l silentLogger) ShallowCopy() Logger {
	return l
}

func (silentLogger) SetLogLevel(level LogLevel) {
	return
}

func (l silentLogger) WithCallDepth(externalDepth uint) Logger {
	return l
}

func (l silentLogger) WithMessageId(msgId string) Logger {
	return l
}
