package rich

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Table 텍스트 기반 테이블을 출력하는 구조체입니다.
// 헤더 행은 볼드체로 강조되며, 각 열은 내용에 맞게 자동으로 너비가 조정됩니다.
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

	// 각 열의 표시 너비(전각 문자는 2칸) 계산
	widths := make([]int, cols)
	for i, h := range t.headers {
		widths[i] = DisplayWidth(h)
	}
	for _, row := range t.rows {
		for i := 0; i < cols; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			if n := DisplayWidth(val); n > widths[i] {
				widths[i] = n
			}
		}
	}

	sep := tableRowSep(widths)

	fmt.Fprintln(w, sep)
	// 헤더 행 (볼드 스타일 적용)
	Fprintln(w, "[bold]%s[/bold]", tableRow(t.headers, widths, cols))
	fmt.Fprintln(w, sep)

	// 데이터 행
	for _, row := range t.rows {
		fmt.Fprintln(w, tableRow(row, widths, cols))
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
func tableRow(row []string, widths []int, cols int) string {
	var b strings.Builder
	b.WriteString("|")
	for i := 0; i < cols; i++ {
		val := ""
		if i < len(row) {
			val = row[i]
		}
		runeLen := DisplayWidth(val)
		padding := widths[i] - runeLen
		b.WriteString(fmt.Sprintf(" %s%s |", val, strings.Repeat(" ", padding)))
	}
	return b.String()
}
