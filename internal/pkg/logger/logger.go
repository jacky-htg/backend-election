package logger

import (
	"context"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/bytedance/sonic"
)

type Logger struct {
	Log    *log.Logger
	Format LoggerFormat
}
type LoggerFormat struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

func New() *Logger {
	return &Logger{Log: log.New(os.Stdout, "", 0)}
}

func (l *Logger) Error(err error) error {
	if ok := l.format("ERROR", err.Error()); ok {
		message, _ := sonic.Marshal(l.Format)
		l.Log.Println(string(message))
	} else {
		l.Log.Println(err.Error())
	}
	return err
}
func (l *Logger) Info(msg string) {
	if ok := l.format("INFO", msg); ok {
		message, _ := sonic.Marshal(l.Format)
		l.Log.Println(string(message))
	} else {
		l.Log.Println(msg)
	}
}

func (l *Logger) Fatal(ctx context.Context, err error) {
	l.Error(err)
	os.Exit(1)
}

func (l *Logger) format(level string, msg string) bool {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		l.Format = LoggerFormat{
			Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			Level:     level,
			Message:   msg,
			File:      path.Base(file),
			Line:      line,
		}
	}
	return ok
}
