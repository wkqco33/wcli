package rich_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/seoyc/wcli/rich"
)

func TestSpinner_NonTerminal(t *testing.T) {
	// bytes.Buffer는 비터미널 → 정적 텍스트 출력 후 goroutine 없이 종료
	var buf bytes.Buffer
	s := rich.NewSpinner(&buf)
	s.Start("로딩 중")
	s.Stop("[green]✓ 완료[/green]")

	out := buf.String()
	if !strings.Contains(out, "로딩 중") {
		t.Errorf("Start 텍스트가 출력되어야 함, 실제: %q", out)
	}
}

func TestSpinner_DoubleStop(t *testing.T) {
	// 중복 Stop 호출 시 패닉/데드락 없음
	var buf bytes.Buffer
	s := rich.NewSpinner(&buf)
	s.Start("test")
	s.Stop("ok")
	s.Stop("ok2") // no-op이어야 함
}

func TestSpinner_UpdateText(t *testing.T) {
	var buf bytes.Buffer
	s := rich.NewSpinner(&buf)
	s.Start("초기 텍스트")
	s.UpdateText("변경된 텍스트")
	s.Stop("")
}

func TestSpinner_NilWriter(t *testing.T) {
	// nil Writer → os.Stderr로 대체 (패닉 없음)
	s := rich.NewSpinner(nil)
	s.Start("nil 테스트")
	time.Sleep(10 * time.Millisecond)
	s.Stop("")
}

func TestSpinner_StartAlreadyRunning(t *testing.T) {
	// 이미 실행 중에 Start 재호출 시 텍스트만 업데이트, goroutine 중복 없음
	var buf bytes.Buffer
	s := rich.NewSpinner(&buf)
	s.Start("첫 번째")
	s.Start("두 번째") // 중복 Start
	s.Stop("")
}

func TestSpinner_StylesAndPreset(t *testing.T) {
	// 다양한 빌트인 스타일 설정 시 문제 없이 동작하는지 테스트
	styles := []rich.SpinnerStyle{
		rich.SpinnerDefault,
		rich.SpinnerDots,
		rich.SpinnerLine,
		rich.SpinnerCircle,
		rich.SpinnerArrow,
		rich.SpinnerBouncing,
	}

	for _, style := range styles {
		var buf bytes.Buffer
		s := rich.NewSpinner(&buf)
		s.SetStyle(style)
		s.Start("로딩")
		time.Sleep(5 * time.Millisecond)
		s.Stop("완료")
	}
}

func TestSpinner_CustomStyle(t *testing.T) {
	// 커스텀 스타일 테스트
	customStyle := rich.SpinnerStyle{
		Frames:   []string{"A", "B", "C"},
		Interval: 2 * time.Millisecond,
	}

	var buf bytes.Buffer
	s := rich.NewSpinner(&buf)
	s.SetStyle(customStyle)
	s.Start("동작")
	time.Sleep(10 * time.Millisecond)
	s.Stop("")
}
