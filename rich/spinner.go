package rich

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// SpinnerStyle 스피너 애니메이션의 프레임과 업데이트 주기 설정을 나타냅니다.
type SpinnerStyle struct {
	Frames   []string
	Interval time.Duration
}

// 빌트인 스피너 스타일 프리셋
var (
	SpinnerDefault = SpinnerStyle{
		Frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Interval: 80 * time.Millisecond,
	}
	SpinnerDots = SpinnerStyle{
		Frames:   []string{".  ", ".. ", "...", "   "},
		Interval: 250 * time.Millisecond,
	}
	SpinnerLine = SpinnerStyle{
		Frames:   []string{"-", "\\", "|", "/"},
		Interval: 100 * time.Millisecond,
	}
	SpinnerCircle = SpinnerStyle{
		Frames:   []string{"◐", "◓", "◑", "◒"},
		Interval: 150 * time.Millisecond,
	}
	SpinnerArrow = SpinnerStyle{
		Frames:   []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		Interval: 100 * time.Millisecond,
	}
	SpinnerBouncing = SpinnerStyle{
		Frames:   []string{"▖", "▘", "▝", "▗"},
		Interval: 100 * time.Millisecond,
	}
)

// Spinner 비동기 작업 중 회전하는 로딩 표시기입니다.
// 터미널 환경에서는 goroutine으로 애니메이션을 출력하고,
// 비터미널 환경에서는 텍스트를 한 줄만 정적으로 출력합니다.
type Spinner struct {
	out     io.Writer
	mu      sync.Mutex
	text    string
	running bool
	done    chan struct{}
	style   SpinnerStyle
}

// NewSpinner 새 Spinner를 생성합니다.
// w가 nil이면 os.Stderr를 기본 출력으로 사용합니다.
func NewSpinner(w io.Writer) *Spinner {
	if w == nil {
		w = os.Stderr
	}
	return &Spinner{
		out:   w,
		style: SpinnerDefault,
	}
}

// SetStyle 스피너의 스타일을 설정합니다.
func (s *Spinner) SetStyle(style SpinnerStyle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.style = style
}

// Start 스피너 애니메이션을 시작합니다.
// 이미 실행 중이면 텍스트만 업데이트합니다.
func (s *Spinner) Start(text string) {
	s.mu.Lock()
	s.text = text

	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// 비터미널 환경: 정적 텍스트 한 줄 출력 후 종료
	if !shouldColor(s.out) {
		Fprintln(s.out, "%s...", text)
		return
	}

	done := make(chan struct{})
	s.mu.Lock()
	s.done = done
	s.mu.Unlock()
	go s.animate(done)
}

// Stop 스피너를 멈추고 완료 메시지를 출력합니다.
// msg가 빈 문자열이면 현재 줄만 지웁니다.
func (s *Spinner) Stop(msg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false

	// 비터미널 환경은 goroutine이 없으므로 메시지만 출력
	if !shouldColor(s.out) {
		s.mu.Unlock()
		if msg != "" {
			Fprintln(s.out, "%s", msg)
		}
		return
	}

	close(s.done)
	s.done = nil
	s.mu.Unlock()

	// 현재 줄 지우고 완료 메시지 출력
	fmt.Fprint(s.out, "\r\033[K")
	if msg != "" {
		Fprintln(s.out, "%s", msg)
	}
}

// UpdateText 실행 중인 스피너의 텍스트를 변경합니다.
func (s *Spinner) UpdateText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text = text
}

// animate 스피너 애니메이션 루프입니다. done은 Start()에서 지역 변수로 캡처되어
// 전달되므로, 매 반복 s.mu 없이 안전하게 참조할 수 있습니다(Stop()이 s.done
// 필드를 락 하에 close/nil로 바꿔도 이 로컬 채널 값 자체는 영향받지 않음 —
// 필드를 직접 재참조하면 락 없는 읽기/쓰기 데이터 레이스가 됩니다).
func (s *Spinner) animate(done <-chan struct{}) {
	s.mu.Lock()
	style := s.style
	s.mu.Unlock()

	ticker := time.NewTicker(style.Interval)
	defer ticker.Stop()

	idx := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.mu.Lock()
			text := s.text
			frames := style.Frames
			s.mu.Unlock()

			if len(frames) == 0 {
				continue
			}
			frame := frames[idx%len(frames)]
			idx++
			// \r로 줄 처음으로 돌아가 덮어쓰기
			fmt.Fprintf(s.out, "\r%s %s", Markup("[cyan]"+frame+"[/cyan]"), text)
		}
	}
}
