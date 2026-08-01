package logging

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestLogLevelString 모든 레벨의 문자열 표현과 미정의 레벨 처리를 검증한다.
func TestLogLevelString(t *testing.T) {
	cases := map[LogLevel]string{
		LevelDebug:    "DEBUG",
		LevelInfo:     "INFO",
		LevelWarn:     "WARN",
		LevelError:    "ERROR",
		LogLevel(999): "UNKNOWN",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("LogLevel(%d).String() = %q, 기대 %q", level, got, want)
		}
	}
}

// TestDefaultLoggerMarkup UseMarkup=true 경로(markup())와 WARN 레벨을 검증한다.
func TestDefaultLoggerMarkup(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger(&buf, LevelDebug, true)
	logger.Log(LevelWarn, "조심 %d", 7)
	out := buf.String()
	if !strings.Contains(out, "WARN") || !strings.Contains(out, "조심 7") {
		t.Errorf("markup 출력에 WARN/메시지가 없음: %q", out)
	}
	// markup 모드에서는 대괄호가 리터럴로 출력되어야 함 (이스케이프된 \[)
	if !strings.Contains(out, "[WARN]") {
		t.Errorf("markup 모드에서 '[WARN]' 리터럴 기대: %q", out)
	}
}

// TestPackageLevelHelpers Debug/Info/Warn/Error 전역 헬퍼를 검증한다.
func TestPackageLevelHelpers(t *testing.T) {
	var buf bytes.Buffer
	SetLogger(NewDefaultLogger(&buf, LevelDebug, false))
	defer SetLogger(nil)

	Debug("d%d", 1)
	Info("i%d", 2)
	Warn("w%d", 3)
	Error("e%d", 4)

	out := buf.String()
	for _, want := range []string{"DEBUG", "d1", "INFO", "i2", "WARN", "w3", "ERROR", "e4"} {
		if !strings.Contains(out, want) {
			t.Errorf("전역 헬퍼 출력에 %q 없음: %q", want, out)
		}
	}
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

// TestGlobalLoggerConcurrentAccess SetLogger와 Debug/Info/Warn/Error(내부적으로
// GetLogger를 거침)를 여러 goroutine에서 동시에 호출해도 데이터 레이스가 없는지
// 확인합니다. config 패키지의 configStore처럼 뮤텍스로 보호되기 전에는 이 테스트를
// `go test -race`로 돌리면 검출됐어야 합니다.
func TestGlobalLoggerConcurrentAccess(t *testing.T) {
	defer SetLogger(nil)

	// io.Discard는 동시 쓰기에 안전(내부 상태 없이 바이트를 버림)하므로,
	// 이 테스트에서 검출하려는 레이스가 Writer 자체가 아니라 오직
	// SetLogger/GetLogger의 전역 상태 접근에 있도록 한다.
	var wg sync.WaitGroup
	const n = 50

	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			SetLogger(NewDefaultLogger(io.Discard, LevelDebug, false))
		}()
		go func(i int) {
			defer wg.Done()
			Debug("goroutine %d", i)
		}(i)
	}
	wg.Wait()
}

func TestDefaultLoggerUsesInjectableClock(t *testing.T) {
	oldNow := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	}
	defer func() { nowFunc = oldNow }()

	var buf bytes.Buffer
	logger := NewDefaultLogger(&buf, LevelInfo, false)
	logger.Log(LevelInfo, "clock test")

	if !strings.Contains(buf.String(), "2026-08-01 10:00:00 [INFO] clock test") {
		t.Fatalf("고정 시계가 반영되어야 함: %q", buf.String())
	}
}
