package main

import (
	"errors"
	"fmt"
)

// This file defines mock structures for testing. It does not contain test.

// discardLogger does not write any log.
type discardLogger struct{}

func (dl discardLogger) Error(...interface{})           {}
func (dl discardLogger) Errorf(string, ...interface{})  {}
func (dl discardLogger) Notice(...interface{})          {}
func (dl discardLogger) Noticef(string, ...interface{}) {}

// bufferedLogger does not write any log but keep them into a buffer instead.
type bufferedLogger struct {
	errBuf, noticeBuf string
}

func (bl *bufferedLogger) Error(v ...interface{}) {
	bl.errBuf = fmt.Sprint(v...)
}

func (bl *bufferedLogger) Errorf(format string, v ...interface{}) {
	bl.errBuf = fmt.Sprintf(format, v...)
}

func (bl *bufferedLogger) Notice(v ...interface{}) {
	bl.noticeBuf = fmt.Sprint(v...)
}

func (bl *bufferedLogger) Noticef(format string, v ...interface{}) {
	bl.noticeBuf = fmt.Sprintf(format, v...)
}

func (bl *bufferedLogger) GetLastError() error {
	if bl.errBuf == "" {
		return nil
	}
	return errors.New(bl.errBuf)
}

func (bl *bufferedLogger) GetLastNotice() string {
	return bl.noticeBuf
}
