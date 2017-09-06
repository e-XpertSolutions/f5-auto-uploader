package main

import (
	"bytes"
	"strings"
	"testing"
)

func trimDatetime(s string) string {
	idx := strings.Index(s, "[")
	if idx == -1 {
		return s
	}
	return s[idx:]
}

func TestDefaultLogger_Error(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := newLogger(buf)
	logger.Error("test")
	want := "[error] test\n"
	if got := trimDatetime(buf.String()); got != want {
		t.Errorf("defaultLogger.Error(%q): got %q; want %q", "test", got, want)
	}
}

func TestDefaultLogger_Errorf(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := newLogger(buf)
	logger.Errorf("%s", "test")
	want := "[error] test\n"
	if got := trimDatetime(buf.String()); got != want {
		t.Errorf("defaultLogger.Errorf(%q, %q): got %q; want %q", "%s", "test", got, want)
	}
}

func TestDefaultLogger_Notice(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := newLogger(buf)
	logger.Notice("test")
	want := "[notice] test\n"
	if got := trimDatetime(buf.String()); got != want {
		t.Errorf("defaultLogger.Notice(%q): got %q; want %q", "test", got, want)
	}
}

func TestDefaultLogger_Noticef(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := newLogger(buf)
	logger.Noticef("%s", "test")
	want := "[notice] test\n"
	if got := trimDatetime(buf.String()); got != want {
		t.Errorf("defaultLogger.Noticef(%q, %q): got %q; want %q", "%s", "test", got, want)
	}
}
