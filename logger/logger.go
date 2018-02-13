package logger

import (
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"fmt"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	logger log.Logger
	Level  Level
}

func New(logLevel string, writer io.Writer) *Logger {
	myTimestamp := log.Valuer(func() interface{} {
		return time.Now().Format(time.RFC3339)
	})
	l := log.NewLogfmtLogger(log.NewSyncWriter(writer))
	l = log.With(l, "ts", myTimestamp)
	return &Logger{
		logger: l,
		Level:  parseLevel(logLevel),
	}
}

func (l *Logger) SetLevel(logLevel string) {
	l.Level = parseLevel(logLevel)
}

func parseLevel(level string) Level {
	switch level {
	case "error":
		return LevelError
	case "warning":
		return LevelWarn
	case "info":
		return LevelInfo
	default:
		return LevelDebug
	}
}

func NewNopLogger() *Logger {
	return &Logger{
		logger: log.NewNopLogger(),
	}
}

func With(l *Logger, keyvals ...interface{}) *Logger {
	return &Logger{
		logger: log.With(l.logger, keyvals...),
		Level:  l.Level,
	}
}

func (l *Logger) Debug(keyvals ...interface{}) {
	if l.Level == LevelDebug {
		log.With(l.logger, "lvl", "debug").Log(keyvals...)
	}
}

func (l *Logger) Info(keyvals ...interface{}) {
	if l.Level <= LevelInfo {
		log.With(l.logger, "lvl", "info").Log(keyvals...)
	}
}

func (l *Logger) Warn(keyvals ...interface{}) {
	if l.Level <= LevelWarn {
		log.With(l.logger, "lvl", "warning").Log(keyvals...)
	}
}

func (l *Logger) Error(keyvals ...interface{}) {
	if l.Level <= LevelError {
		log.With(l.logger, "lvl", "error").Log(keyvals...)
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Debug("msg", fmt.Sprintf(format, v))
}
