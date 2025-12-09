package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[provider_management] ", log.LstdFlags|log.Lshortfile),
	}
}
