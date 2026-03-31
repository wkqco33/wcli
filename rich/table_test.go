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
}
