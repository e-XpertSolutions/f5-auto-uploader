package main

import (
	"fmt"
	"io"
	"log"
)

type logger interface {
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Notice(v ...interface{})
	Noticef(format string, v ...interface{})
}

type defaultLogger struct {
	l *log.Logger
}

func newLogger(w io.Writer) logger {
	return &defaultLogger{l: log.New(w, "", log.LstdFlags)}
}

func (dl defaultLogger) Error(v ...interface{}) {
	dl.l.Print("[error] ", fmt.Sprint(v...))
}

func (dl defaultLogger) Errorf(format string, v ...interface{}) {
	dl.l.Print("[error] ", fmt.Sprintf(format, v...))
}

func (dl defaultLogger) Notice(v ...interface{}) {
	dl.l.Print("[notice] ", fmt.Sprint(v...))
}

func (dl defaultLogger) Noticef(format string, v ...interface{}) {
	dl.l.Print("[notice] ", fmt.Sprintf(format, v...))
}
