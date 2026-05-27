package logging

import (
	"fmt"
	"io"
	"time"

	"github.com/seoyc/wcli/rich"
)

// LogLevel 로그 레벨을 나타내는 정수 타입
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String 로그 레벨의 문자열 표현을 반환합니다.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger wcli 내부 및 애플리케이션 실행 중 로그를 기록하는 인터페이스
type Logger interface {
	Log(level LogLevel, format string, args ...any)
}

// DefaultLogger 기본 콘솔/출력 로거 구현체
type DefaultLogger struct {
	Writer    io.Writer
	MinLevel  LogLevel
	UseMarkup bool
}

// NewDefaultLogger 새 DefaultLogger 인스턴스를 생성합니다.
func NewDefaultLogger(w io.Writer, minLevel LogLevel, useMarkup bool) *DefaultLogger {
	return &DefaultLogger{
		Writer:    w,
		MinLevel:  minLevel,
		UseMarkup: useMarkup,
	}
}

// Log 지정된 레벨로 로그를 출력합니다.
func (l *DefaultLogger) Log(level LogLevel, format string, args ...any) {
	if level < l.MinLevel || l.Writer == nil {
		return
	}

	timeStr := time.Now().Format("2006-01-02 15:04:05")
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	if l.UseMarkup {
		var levelStr string
		switch level {
		case LevelDebug:
			levelStr = "[dim]DEBUG[/dim]"
		case LevelInfo:
			levelStr = "[blue]INFO[/blue]"
		case LevelWarn:
			levelStr = "[yellow]WARN[/yellow]"
		case LevelError:
			levelStr = "[red][bold]ERROR[/bold][/red]"
		}
		rich.Fprintln(l.Writer, "%s [%s] %s", timeStr, levelStr, msg)
	} else {
		fmt.Fprintf(l.Writer, "%s [%s] %s\n", timeStr, level.String(), msg)
	}
}

// NoOpLogger 아무 작동도 하지 않는 빈 로거 (기본값 설정용)
type NoOpLogger struct{}

func (n *NoOpLogger) Log(level LogLevel, format string, args ...any) {}

// Global Logger (기본은 NoOpLogger)
var globalLogger Logger = &NoOpLogger{}

// SetLogger 전역 로거를 주입합니다.
func SetLogger(l Logger) {
	if l != nil {
		globalLogger = l
	} else {
		globalLogger = &NoOpLogger{}
	}
}

// GetLogger 전역 로거를 획득합니다.
func GetLogger() Logger {
	return globalLogger
}
