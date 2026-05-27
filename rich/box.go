package rich

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Box 텍스트를 유니코드 테두리 박스로 감싸 출력하는 구조체입니다.
// 내용(Content)에 마크업 태그가 포함된 경우 그대로 파싱됩니다.
type Box struct {
	Title   string // 상단 테두리에 표시할 제목 (빈 문자열이면 생략)
	Content string // 박스 내부 텍스트
}

// NewBox 내용을 받아 새 Box를 생성합니다.
func NewBox(content string) *Box {
	return &Box{Content: content}
}

// WithTitle 제목을 설정하고 자신을 반환합니다 (메서드 체이닝 지원).
func (b *Box) WithTitle(title string) *Box {
	b.Title = title
	return b
}

// Render 박스를 w에 출력합니다.
// 테두리 문자에는 볼드 스타일이 적용되며, 내용의 마크업 태그도 파싱됩니다.
func (b *Box) Render(w io.Writer) {
	lines := strings.Split(b.Content, "\n")

	// 최대 시각 너비 계산 (마크업 태그 제외)
	maxWidth := utf8.RuneCountInString(b.Title)
	for _, line := range lines {
		if n := utf8.RuneCountInString(stripMarkup(line)); n > maxWidth {
			maxWidth = n
		}
	}

	// inner = 내용 영역 너비 (좌우 패딩 각 1칸)
	inner := maxWidth + 2

	// 상단 테두리
	if b.Title != "" {
		titleStr := " " + b.Title + " "
		titleLen := utf8.RuneCountInString(titleStr)
		if titleLen > inner {
			inner = titleLen
			maxWidth = inner - 2
		}
		left := (inner - titleLen) / 2
		right := inner - titleLen - left
		top := "┌" + strings.Repeat("─", left) + titleStr + strings.Repeat("─", right) + "┐"
		Fprintln(w, "[bold]%s[/bold]", top)
	} else {
		Fprintln(w, "[bold]%s[/bold]", "┌"+strings.Repeat("─", inner)+"┐")
	}

	// 내용 줄 (내용 자체의 마크업도 그대로 파싱됨)
	for _, line := range lines {
		stripped := stripMarkup(line) // maxWidth 계산에서 이미 호출됐으므로 재계산 방지
		visualLen := utf8.RuneCountInString(stripped)
		padding := strings.Repeat(" ", maxWidth-visualLen)
		Fprintln(w, "[bold]│[/bold] %s%s [bold]│[/bold]", line, padding)
	}

	// 하단 테두리
	Fprintln(w, "[bold]%s[/bold]", "└"+strings.Repeat("─", inner)+"┘")
}

// Print os.Stdout에 박스를 출력합니다.
func (b *Box) Print() {
	b.Render(os.Stdout)
}
