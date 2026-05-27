package rich

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// spinnerFrames 브라이유 점자 기반 회전 프레임
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner 비동기 작업 중 회전하는 로딩 표시기입니다.
// 터미널 환경에서는 goroutine으로 애니메이션을 출력하고,
// 비터미널 환경에서는 텍스트를 한 줄만 정적으로 출력합니다.
type Spinner struct {
	out     io.Writer
	mu      sync.Mutex
	text    string
	running bool
	done    chan struct{}
}

// NewSpinner 새 Spinner를 생성합니다.
// w가 nil이면 os.Stderr를 기본 출력으로 사용합니다.
func NewSpinner(w io.Writer) *Spinner {
	if w == nil {
		w = os.Stderr
	}
	return &Spinner{out: w}
}

// Start 스피너 애니메이션을 시작합니다.
// 이미 실행 중이면 텍스트만 업데이트합니다.
func (s *Spinner) Start(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.text = text

	if s.running {
		return
	}
	s.running = true

	// 비터미널 환경: 정적 텍스트 한 줄 출력 후 종료
	if !shouldColor(s.out) {
		Fprintln(s.out, "%s...", text)
		return
	}

	s.done = make(chan struct{})
	go s.animate()
}

// Stop 스피너를 멈추고 완료 메시지를 출력합니다.
// msg가 빈 문자열이면 현재 줄만 지웁니다.
func (s *Spinner) Stop(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.running = false

	// 비터미널 환경은 goroutine이 없으므로 메시지만 출력
	if !shouldColor(s.out) {
		if msg != "" {
			Fprintln(s.out, "%s", msg)
		}
		return
	}

	// goroutine 종료 대기
	close(s.done)
	s.done = nil

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

func (s *Spinner) animate() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	idx := 0
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			text := s.text
			s.mu.Unlock()

			frame := spinnerFrames[idx%len(spinnerFrames)]
			idx++
			// \r로 줄 처음으로 돌아가 덮어쓰기
			fmt.Fprintf(s.out, "\r%s %s", Markup("[cyan]"+frame+"[/cyan]"), text)
		}
	}
}
