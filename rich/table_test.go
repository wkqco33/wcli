package rich_test

import (
	"strings"
	"testing"

	"github.com/seoyc/wcli/rich"
)

func TestTable_Render(t *testing.T) {
	t.Run("기본 테이블 출력", func(t *testing.T) {
		var buf strings.Builder
		tbl := rich.NewTable("이름", "나이", "직업")
		tbl.AddRow("홍길동", "30", "개발자")
		tbl.AddRow("김철수", "25", "디자이너")
		tbl.Render(&buf)

		output := buf.String()
		// 구분선 포함 확인
		if !strings.Contains(output, "+") {
			t.Error("테이블 출력에 구분선(+)이 없습니다")
		}
		// 헤더 내용 확인
		if !strings.Contains(output, "이름") {
			t.Error("테이블 출력에 헤더 '이름'이 없습니다")
		}
		// 데이터 행 확인
		if !strings.Contains(output, "홍길동") {
			t.Error("테이블 출력에 '홍길동'이 없습니다")
		}
		if !strings.Contains(output, "디자이너") {
			t.Error("테이블 출력에 '디자이너'가 없습니다")
		}
	})

	t.Run("빈 헤더는 아무것도 출력하지 않음", func(t *testing.T) {
		var buf strings.Builder
		tbl := rich.NewTable()
		tbl.Render(&buf)
		if buf.String() != "" {
			t.Errorf("빈 헤더 테이블은 출력이 없어야 합니다, got: %q", buf.String())
		}
	})

	t.Run("열 수가 헤더보다 적은 행", func(t *testing.T) {
		var buf strings.Builder
		tbl := rich.NewTable("A", "B", "C")
		tbl.AddRow("1", "2") // C 열 누락
		tbl.Render(&buf)
		output := buf.String()
		if !strings.Contains(output, "1") || !strings.Contains(output, "2") {
			t.Error("부분 행이 올바르게 출력되지 않았습니다")
		}
	})

	t.Run("메서드 체이닝", func(t *testing.T) {
		var buf strings.Builder
		rich.NewTable("X", "Y").AddRow("1", "2").AddRow("3", "4").Render(&buf)
		output := buf.String()
		if !strings.Contains(output, "X") || !strings.Contains(output, "3") {
			t.Error("메서드 체이닝이 올바르게 동작하지 않습니다")
		}
	})

	t.Run("셀 마크업이 있어도 정렬이 어긋나지 않음", func(t *testing.T) {
		// 마크업 태그는 표시 폭 계산에서 제외되어야 하므로, 터미널이 아닌
		// 환경(color 미적용)에서는 태그가 제거된 채로 다른 셀과 폭이 맞아야 한다.
		var buf strings.Builder
		tbl := rich.NewTable("이름", "상태")
		tbl.AddRow("김철수", "[green]정상[/green]")
		tbl.AddRow("이영희", "[red]오류[/red]")
		tbl.Render(&buf)

		output := buf.String()
		if strings.Contains(output, "[green]") || strings.Contains(output, "[/red]") {
			t.Errorf("터미널이 아닌 환경에서는 마크업 태그가 제거되어야 합니다: %q", output)
		}

		lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
		want := rich.DisplayWidth(lines[0])
		for i, l := range lines {
			if w := rich.DisplayWidth(l); w != want {
				t.Errorf("줄 %d 표시폭=%d, 기대=%d (%q)", i, w, want, l)
			}
		}
	})
}

func TestTable_AlignRight(t *testing.T) {
	var buf strings.Builder
	tbl := rich.NewTable("항목", "값")
	tbl.SetAlign(rich.AlignLeft, rich.AlignRight)
	tbl.AddRow("나이", "30")
	tbl.AddRow("키", "175.5")
	tbl.Render(&buf)

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, l := range lines {
		if strings.Contains(l, "30") || strings.Contains(l, "175.5") {
			if !strings.HasPrefix(strings.TrimSpace(l), "|") {
				t.Errorf("우측 정렬 행 형식이 예상과 다름: %q", l)
			}
		}
	}
}

func TestTable_AlignCenter(t *testing.T) {
	var buf strings.Builder
	tbl := rich.NewTable("이름", "역할")
	tbl.SetAlign(rich.AlignCenter, rich.AlignCenter)
	tbl.AddRow("홍길동", "개발자")
	tbl.Render(&buf)

	output := buf.String()
	if !strings.Contains(output, "홍길동") || !strings.Contains(output, "개발자") {
		t.Error("가운데 정렬 테이블에 데이터가 없습니다")
	}
}

func TestTable_MaxWidth(t *testing.T) {
	var buf strings.Builder
	tbl := rich.NewTable("이름", "설명")
	tbl.AddRow("짧음", "짧은텍스트")
	tbl.AddRow("김철수", "이것은매우긴설명텍스트입니다")
	tbl.SetMaxWidth(30)
	tbl.Render(&buf)

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for _, l := range lines {
		if w := rich.DisplayWidth(l); w > 30 {
			t.Errorf("줄 표시폭=%d, 최대 30을 초과함: %q", w, l)
		}
	}
}

func TestTable_ChainingWithAlign(t *testing.T) {
	var buf strings.Builder
	rich.NewTable("A", "B", "C").
		AddRow("1", "2", "3").
		SetAlign(rich.AlignRight, rich.AlignCenter, rich.AlignLeft).
		Render(&buf)
	output := buf.String()
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Error("체이닝 + 정렬 조합이 올바르게 동작하지 않습니다")
	}
}
