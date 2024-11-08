package log

import (
	"log"
	"os"
)

//go:generate mockgen -source=logger.go -destination=mock/logger.go
type Logger interface {
	Error(args ...interface{})
	Info(args ...interface{})
	Fatal(args ...interface{})
}

type logger struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

func New() Logger {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return &logger{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
	}
}

func (l *logger) Error(args ...interface{}) {
	l.ErrorLog.Println(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.InfoLog.Println(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.ErrorLog.Fatal(args...)
}
