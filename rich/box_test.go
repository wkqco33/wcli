package rich_test

import (
	"strings"
	"testing"

	"github.com/seoyc/wcli/rich"
)

func TestBox_Render(t *testing.T) {
	t.Run("기본 박스", func(t *testing.T) {
		var buf strings.Builder
		rich.NewBox("Hello, World!").Render(&buf)
		output := buf.String()
		// 테두리 문자 확인
		if !strings.Contains(output, "┌") || !strings.Contains(output, "┐") {
			t.Error("상단 테두리 문자가 없습니다")
		}
		if !strings.Contains(output, "└") || !strings.Contains(output, "┘") {
			t.Error("하단 테두리 문자가 없습니다")
		}
		if !strings.Contains(output, "│") {
			t.Error("측면 테두리 문자가 없습니다")
		}
		if !strings.Contains(output, "Hello, World!") {
			t.Error("내용이 출력에 없습니다")
		}
	})

	t.Run("제목 있는 박스", func(t *testing.T) {
		var buf strings.Builder
		rich.NewBox("내용입니다").WithTitle("제목").Render(&buf)
		output := buf.String()
		if !strings.Contains(output, "제목") {
			t.Error("제목이 출력에 없습니다")
		}
		if !strings.Contains(output, "내용입니다") {
			t.Error("내용이 출력에 없습니다")
		}
	})

	t.Run("여러 줄 내용", func(t *testing.T) {
		var buf strings.Builder
		rich.NewBox("첫 번째 줄\n두 번째 줄\n세 번째 줄").Render(&buf)
		output := buf.String()
		if !strings.Contains(output, "첫 번째 줄") {
			t.Error("첫 번째 줄이 없습니다")
		}
		if !strings.Contains(output, "세 번째 줄") {
			t.Error("세 번째 줄이 없습니다")
		}
	})

	t.Run("메서드 체이닝", func(t *testing.T) {
		var buf strings.Builder
		b := rich.NewBox("content").WithTitle("title")
		b.Render(&buf)
		output := buf.String()
		if !strings.Contains(output, "title") || !strings.Contains(output, "content") {
			t.Error("메서드 체이닝이 올바르게 동작하지 않습니다")
		}
	})
}
