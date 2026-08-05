package logger

import "log"

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type StdLogger struct{}

func (StdLogger) Info(msg string)  { log.Println("[INFO]", msg) }
func (StdLogger) Error(msg string) { log.Println("[ERROR]", msg) }
