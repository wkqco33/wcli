package rich

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ProgressTheme 프로그레스바의 모양을 나타내는 테마 구조체입니다.
type ProgressTheme struct {
	Fill  string
	Empty string
}

// 빌트인 프로그레스바 테마 프리셋
var (
	ThemeBlock      = ProgressTheme{Fill: "█", Empty: "░"}
	ThemeLine       = ProgressTheme{Fill: "━", Empty: "─"}
	ThemeDoubleLine = ProgressTheme{Fill: "═", Empty: "─"}
	ThemeBullet     = ProgressTheme{Fill: "●", Empty: "○"}
	ThemeArrow      = ProgressTheme{Fill: ">", Empty: " "}
	ThemeStar       = ProgressTheme{Fill: "★", Empty: "☆"}
)

var nowFunc = time.Now

// ProgressBar 진행률을 시각적으로 표시하는 구조체입니다.
type ProgressBar struct {
	Total       int       // 전체 단계 수
	Width       int       // 막대 너비 (기본값: 40)
	Fill        string    // 채워진 부분 문자 (기본값: "█")
	Empty       string    // 빈 부분 문자 (기본값: "░")
	FillColor   string    // 채워진 부분 색상 마크업 (기본값: "green")
	EmptyColor  string    // 빈 부분 색상 마크업 (기본값: "")
	ShowPercent bool      // 백분율 표시 여부 (기본값: true)
	ShowCounter bool      // 카운터 표시 여부 (예: "(50/100)", 기본값: false)
	ShowETA     bool      // 남은 예상 시간 표시 여부 (기본값: false)
	startTime   time.Time // 시작 시간 (ETA 계산용)
}

// NewProgressBar 새 ProgressBar를 생성합니다.
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		Total:       total,
		Width:       40,
		Fill:        "█",
		Empty:       "░",
		FillColor:   "green",
		EmptyColor:  "",
		ShowPercent: true,
		ShowCounter: false,
		ShowETA:     false,
	}
}

// SetTheme 프로그레스바의 테마를 설정합니다.
func (p *ProgressBar) SetTheme(theme ProgressTheme) {
	p.Fill = theme.Fill
	p.Empty = theme.Empty
}

// Start 프로그레스바 타이머를 시작합니다. ETA 계산에 필요합니다.
func (p *ProgressBar) Start() {
	p.startTime = nowFunc()
}

// calculateETA 남은 예상 시간을 계산합니다.
func (p *ProgressBar) calculateETA(current int) string {
	if p.startTime.IsZero() || current <= 0 || current >= p.Total {
		return ""
	}
	elapsed := time.Since(p.startTime)
	pct := float64(current) / float64(p.Total)
	
	totalTime := time.Duration(float64(elapsed) / pct)
	eta := totalTime - elapsed
	
	seconds := int(eta.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("ETA: %ds", seconds)
}

// Render current/total 진행률을 마크업 문자열로 반환합니다.
func (p *ProgressBar) Render(current int) string {
	if p.Total <= 0 {
		return ""
	}
	width := p.Width
	if width <= 0 {
		width = 40
	}
	pct := float64(current) / float64(p.Total)
	if pct < 0 {
		pct = 0
	} else if pct > 1.0 {
		pct = 1.0
	}
	filled := int(pct * float64(width))

	var fillStr string
	if p.FillColor != "" && filled > 0 {
		fillStr = fmt.Sprintf("[%s]%s[/%s]", p.FillColor, strings.Repeat(p.Fill, filled), p.FillColor)
	} else {
		fillStr = strings.Repeat(p.Fill, filled)
	}

	var emptyStr string
	if p.EmptyColor != "" && width-filled > 0 {
		emptyStr = fmt.Sprintf("[%s]%s[/%s]", p.EmptyColor, strings.Repeat(p.Empty, width-filled), p.EmptyColor)
	} else {
		emptyStr = strings.Repeat(p.Empty, width-filled)
	}

	bar := fillStr + emptyStr

	var parts []string
	parts = append(parts, bar)

	if p.ShowPercent {
		parts = append(parts, fmt.Sprintf("%3.0f%%", pct*100))
	}

	if p.ShowCounter {
		parts = append(parts, fmt.Sprintf("(%d/%d)", current, p.Total))
	}

	if p.ShowETA {
		if etaStr := p.calculateETA(current); etaStr != "" {
			parts = append(parts, etaStr)
		}
	}

	return strings.Join(parts, " ")
}

// Fprint w에 진행 상황을 한 줄로 출력합니다.
func (p *ProgressBar) Fprint(w io.Writer, current int) {
	Fprintln(w, "%s", p.Render(current))
}

// Print os.Stdout에 진행 상황을 출력합니다.
func (p *ProgressBar) Print(current int) {
	p.Fprint(os.Stdout, current)
}
