package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger(&buf, LevelInfo, false)

	// Debug should be ignored
	logger.Log(LevelDebug, "debug message %d", 1)
	if buf.Len() > 0 {
		t.Errorf("expected debug to be filtered out, got %q", buf.String())
	}

	// Info should be printed
	logger.Log(LevelInfo, "info message %s", "hello")
	infoOutput := buf.String()
	if !strings.Contains(infoOutput, "[INFO]") || !strings.Contains(infoOutput, "info message hello") {
		t.Errorf("unexpected info output: %q", infoOutput)
	}

	buf.Reset()
	// Error should be printed
	logger.Log(LevelError, "error message")
	errOutput := buf.String()
	if !strings.Contains(errOutput, "[ERROR]") || !strings.Contains(errOutput, "error message") {
		t.Errorf("unexpected error output: %q", errOutput)
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}
	// Should not panic, should do nothing
	logger.Log(LevelError, "test message")
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewDefaultLogger(&buf, LevelDebug, false)
	SetLogger(l)

	if GetLogger() != l {
		t.Errorf("SetLogger/GetLogger mismatch")
	}

	GetLogger().Log(LevelDebug, "test global")
	if !strings.Contains(buf.String(), "test global") {
		t.Errorf("global logger did not output properly: %q", buf.String())
	}

	SetLogger(nil)
	if _, ok := GetLogger().(*NoOpLogger); !ok {
		t.Errorf("expected GetLogger to return NoOpLogger after setting nil")
	}
}
