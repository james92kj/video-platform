package logger 

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New() *Logger {
	return &Logger {
		Logger: log.New(os.Stdout, "VIDEO-API:", log.LstdFlags),
	}
}

func (l *Logger) Info(msg string) {
	l.Printf("INFO: %s",msg)
}

func (l *Logger) Error(msg string) {
	l.Printf("ERROR: %s",msg)
}


