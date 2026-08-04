package rich

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Align 열 정렬 방식을 나타냅니다.
type Align int

const (
	AlignLeft   Align = iota // 좌측 정렬 (기본값)
	AlignRight               // 우측 정렬
	AlignCenter              // 가운데 정렬
)

// Table 텍스트 기반 테이블을 출력하는 구조체입니다.
// 헤더 행은 볼드체로 강조되며, 각 열은 내용에 맞게 자동으로 너비가 조정됩니다.
// 헤더/셀 값에 포함된 마크업 태그([red]...[/red] 등)도 파싱되어 스타일이 적용됩니다.
type Table struct {
	headers  []string
	rows     [][]string
	aligns   []Align // 열별 정렬 (nil이면 전부 좌측 정렬)
	maxWidth int     // 최대 표시 너비 (0이면 제한 없음)
}

// NewTable 헤더를 받아 새 Table을 생성합니다.
func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

// AddRow 행을 추가합니다. 열 개수가 헤더보다 적은 경우 렌더링 시 빈 문자열로 채워집니다.
func (t *Table) AddRow(cols ...string) *Table {
	t.rows = append(t.rows, cols)
	return t
}

// SetAlign 열별 정렬 방식을 설정합니다. 헤더 개수보다 적게 지정하면 나머지는 좌측 정렬됩니다.
func (t *Table) SetAlign(aligns ...Align) *Table {
	t.aligns = aligns
	return t
}

// SetMaxWidth 테이블 전체 최대 표시 너비를 설정합니다.
// 너비를 초과하는 열은 말줄임표(…)로 축약됩니다.
// 0이면 제한 없음 (기본값).
func (t *Table) SetMaxWidth(width int) *Table {
	t.maxWidth = width
	return t
}

// Render w에 테이블을 출력합니다.
// w가 터미널이면 헤더에 볼드 스타일이 적용됩니다.
func (t *Table) Render(w io.Writer) {
	cols := len(t.headers)
	if cols == 0 {
		return
	}

	// 각 열의 표시 너비 계산 (마크업 태그 제외, 전각 문자는 2칸)
	widths := make([]int, cols)
	for i, h := range t.headers {
		widths[i] = DisplayWidth(stripMarkup(h))
	}
	for _, row := range t.rows {
		for i := 0; i < cols; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			if n := DisplayWidth(stripMarkup(val)); n > widths[i] {
				widths[i] = n
			}
		}
	}

	// 최대 너비 제한 적용
	if t.maxWidth > 0 {
		// 구분선 포함 전체 너비 = (cols*3 + 1) + sum(widths)
		total := cols*3 + 1
		for _, w := range widths {
			total += w
		}
		if total > t.maxWidth {
			overflow := total - t.maxWidth
			// 가장 넓은 열부터 줄임
			for overflow > 0 {
				// 가장 넓은 열 찾기
				maxIdx := 0
				for i := 1; i < cols; i++ {
					if widths[i] > widths[maxIdx] {
						maxIdx = i
					}
				}
				if widths[maxIdx] <= 3 {
					break
				}
				shrink := widths[maxIdx] - 3
				if shrink > overflow {
					shrink = overflow
				}
				widths[maxIdx] -= shrink
				overflow -= shrink
			}
		}
	}

	sep := tableRowSep(widths)

	fmt.Fprintln(w, sep)
	// 헤더 행 (볼드 스타일 적용, 셀 마크업도 함께 파싱됨)
	Fprintln(w, "[bold]%s[/bold]", tableRow(t.headers, widths, cols, t.aligns))
	fmt.Fprintln(w, sep)

	// 데이터 행 (셀 마크업 파싱)
	for _, row := range t.rows {
		Fprintln(w, "%s", tableRow(row, widths, cols, t.aligns))
	}
	fmt.Fprintln(w, sep)
}

// Print os.Stdout에 테이블을 출력합니다.
func (t *Table) Print() {
	t.Render(os.Stdout)
}

// tableRowSep 각 열 너비에 맞는 +---+---+ 형식의 구분선을 반환합니다.
func tableRowSep(widths []int) string {
	var b strings.Builder
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		b.WriteString("+")
	}
	return b.String()
}

// tableRow 각 셀 값을 열 너비에 맞게 패딩한 | val | ... | 형식의 행 문자열을 반환합니다.
// 패딩은 마크업 태그를 제외한 표시 너비 기준으로 계산되므로, 셀에 마크업이
// 포함되어 있어도(예: "[green]OK[/green]") 정렬이 어긋나지 않습니다.
func tableRow(row []string, widths []int, cols int, aligns []Align) string {
	var b strings.Builder
	b.WriteString("|")
	for i := 0; i < cols; i++ {
		val := ""
		if i < len(row) {
			val = row[i]
		}
		visualLen := DisplayWidth(stripMarkup(val))

		// 열 너비 초과 시 말줄임
		if visualLen > widths[i] {
			val = truncateDisplay(val, widths[i])
			visualLen = widths[i]
		}

		padding := widths[i] - visualLen

		align := AlignLeft
		if i < len(aligns) {
			align = aligns[i]
		}

		switch align {
		case AlignRight:
			fmt.Fprintf(&b, " %s%s |", strings.Repeat(" ", padding), val)
		case AlignCenter:
			left := padding / 2
			right := padding - left
			fmt.Fprintf(&b, " %s%s%s |", strings.Repeat(" ", left), val, strings.Repeat(" ", right))
		default:
			fmt.Fprintf(&b, " %s%s |", val, strings.Repeat(" ", padding))
		}
	}
	return b.String()
}

// truncateDisplay 문자열을 주어진 표시 너비에 맞게 자르고 말줄임표(…)를 붙입니다.
func truncateDisplay(s string, maxWidth int) string {
	if maxWidth <= 1 {
		return ""
	}
	// 말줄임표(…)는 표시 폭 2
	ellipsisWidth := 2
	keepWidth := maxWidth - ellipsisWidth
	if keepWidth <= 0 {
		return "…"
	}

	var result strings.Builder
	currentWidth := 0
	for _, r := range s {
		rw := runeWidth(r)
		if currentWidth+rw > keepWidth {
			break
		}
		result.WriteRune(r)
		currentWidth += rw
	}
	result.WriteRune('…')
	return result.String()
}
