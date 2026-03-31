package rich

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// ProgressBar 진행률을 시각적으로 표시하는 구조체입니다.
type ProgressBar struct {
	Total int    // 전체 단계 수
	Width int    // 막대 너비 (기본값: 40)
	Fill  string // 채워진 부분 문자 (기본값: "█")
	Empty string // 빈 부분 문자 (기본값: "░")
}

// NewProgressBar 새 ProgressBar를 생성합니다.
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		Total: total,
		Width: 40,
		Fill:  "█",
		Empty: "░",
	}
}

// Render current/total 진행률을 마크업 문자열로 반환합니다.
// 반환값에는 채워진 부분에 green 마크업이 적용됩니다.
func (p *ProgressBar) Render(current int) string {
	if p.Total <= 0 {
		return ""
	}
	width := p.Width
	if width <= 0 {
		width = 40
	}
	pct := math.Max(0, math.Min(1.0, float64(current)/float64(p.Total)))
	filled := int(pct * float64(width))
	bar := "[green]" + strings.Repeat(p.Fill, filled) + "[/green]" + strings.Repeat(p.Empty, width-filled)
	return fmt.Sprintf("%s %3.0f%%", bar, pct*100)
}

// Fprint w에 진행 상황을 한 줄로 출력합니다.
func (p *ProgressBar) Fprint(w io.Writer, current int) {
	Fprintln(w, "%s", p.Render(current))
}

// Print os.Stdout에 진행 상황을 출력합니다.
func (p *ProgressBar) Print(current int) {
	p.Fprint(os.Stdout, current)
}
