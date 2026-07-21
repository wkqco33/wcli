package rich

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Table 텍스트 기반 테이블을 출력하는 구조체입니다.
// 헤더 행은 볼드체로 강조되며, 각 열은 내용에 맞게 자동으로 너비가 조정됩니다.
// 헤더/셀 값에 포함된 마크업 태그([red]...[/red] 등)도 파싱되어 스타일이 적용됩니다.
type Table struct {
	headers []string
	rows    [][]string
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

	sep := tableRowSep(widths)

	fmt.Fprintln(w, sep)
	// 헤더 행 (볼드 스타일 적용, 셀 마크업도 함께 파싱됨)
	Fprintln(w, "[bold]%s[/bold]", tableRow(t.headers, widths, cols))
	fmt.Fprintln(w, sep)

	// 데이터 행 (셀 마크업 파싱)
	for _, row := range t.rows {
		Fprintln(w, "%s", tableRow(row, widths, cols))
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
func tableRow(row []string, widths []int, cols int) string {
	var b strings.Builder
	b.WriteString("|")
	for i := 0; i < cols; i++ {
		val := ""
		if i < len(row) {
			val = row[i]
		}
		visualLen := DisplayWidth(stripMarkup(val))
		padding := widths[i] - visualLen
		b.WriteString(fmt.Sprintf(" %s%s |", val, strings.Repeat(" ", padding)))
	}
	return b.String()
}
