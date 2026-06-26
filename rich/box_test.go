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

	t.Run("전각 문자 정렬", func(t *testing.T) {
		// 한글/이모지가 섞여 줄마다 글자 수가 달라도 모든 줄의 표시 폭이
		// 같아야 테두리가 맞는다(전각=2칸 반영).
		var buf strings.Builder
		rich.NewBox("포맷  GIF\n프레임  10  (Ctrl+C로 종료)").
			WithTitle("🎬 t.gif").Render(&buf)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		want := rich.DisplayWidth(lines[0])
		for i, l := range lines {
			if w := rich.DisplayWidth(l); w != want {
				t.Errorf("줄 %d 표시폭=%d, 기대=%d (%q)", i, w, want, l)
			}
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
