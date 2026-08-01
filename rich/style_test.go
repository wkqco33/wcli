package rich_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/seoyc/wcli/rich"
)

func TestMarkup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "기본 색상 적용",
			input:    "[red]에러[/red]",
			expected: "\033[31m에러",
		},
		{
			name:     "다중 태그 (순차적)",
			input:    "[blue]정보[/blue] [green]성공[/green]",
			expected: "\033[34m정보",
		},
		{
			name:     "이스케이프",
			input:    `\[red] 일반 텍스트`,
			expected: "[red] 일반 텍스트",
		},
		{
			name:     "중첩 태그 - 닫힘 순서가 다를 때도 정상 동작",
			input:    "[bold][red]text[/bold][/red]",
			expected: "\033[1m\033[31mtext",
		},
		{
			name:     "닫힘 태그 이름으로 정확히 pop",
			input:    "[bold][red]text[/red] still bold[/bold]",
			expected: "\033[1m\033[31mtext",
		},
		{
			name:     "italic 태그",
			input:    "[italic]기울임[/italic]",
			expected: "\033[3m기울임",
		},
		{
			name:     "blink 태그",
			input:    "[blink]깜빡임[/blink]",
			expected: "\033[5m깜빡임",
		},
		{
			name:     "reverse 태그",
			input:    "[reverse]반전[/reverse]",
			expected: "\033[7m반전",
		},
		{
			name:     "strikethrough 태그",
			input:    "[strikethrough]취소선[/strikethrough]",
			expected: "\033[9m취소선",
		},
		{
			name:     "배경색 bg-red 태그",
			input:    "[bg-red]배경[/bg-red]",
			expected: "\033[41m배경",
		},
		{
			name:     "배경색 bg-green 태그",
			input:    "[bg-green]배경[/bg-green]",
			expected: "\033[42m배경",
		},
		{
			name:     "전경+배경 조합",
			input:    "[white][bg-blue]조합[/bg-blue][/white]",
			expected: "\033[37m\033[44m조합",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rich.Markup(tt.input)
			if !strings.Contains(got, tt.expected) {
				t.Errorf("Markup() = %q, want to contain %q", got, tt.expected)
			}
		})
	}
}

// TestFprintln_NonTagBracketsPreserved 비-터미널(색상 미지원) 출력 경로에서
// tagMap에 없는 대괄호 텍스트가 삭제되지 않고 그대로 보존되는지 확인합니다.
// bytes.Buffer는 터미널이 아니므로 Fprintln이 내부적으로 stripMarkup을 사용합니다.
func TestFprintln_NonTagBracketsPreserved(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "알 수 없는 여는 태그는 리터럴로 보존",
			input: "허용값: [TODO] 또는 [WIP]",
			want:  "허용값: [TODO] 또는 [WIP]",
		},
		{
			name:  "쉼표가 섞인 목록형 대괄호는 태그가 아니므로 보존",
			input: "다음 중 하나: [red, green, blue]",
			want:  "다음 중 하나: [red, green, blue]",
		},
		{
			name:  "실제 알려진 태그는 여전히 제거됨",
			input: "[bold]강조[/bold] 일반",
			want:  "강조 일반",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rich.Fprintln(&buf, "%s", tt.input)
			got := strings.TrimRight(buf.String(), "\n")
			if got != tt.want {
				t.Errorf("Fprintln output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEscapeMarkup(t *testing.T) {
	// 이스케이프하지 않으면 "[red]"처럼 실제 태그명과 우연히 일치하는 리터럴
	// 텍스트가 Markup()에 의해 ANSI 코드로 둔갑해 원래 글자가 사라진다.
	dynamic := "허용값: [red] 또는 [blue]"
	unescaped := rich.Markup(dynamic)
	if strings.Contains(unescaped, "red") || strings.Contains(unescaped, "blue") {
		t.Fatalf("전제 조건 실패: 이스케이프 없이도 리터럴이 보존됨(테스트 가정과 다름): %q", unescaped)
	}

	// EscapeMarkup을 적용하면 대괄호가 리터럴로 보존되어야 한다.
	escaped := rich.EscapeMarkup(dynamic)
	got := rich.Markup(escaped)
	if got != dynamic {
		t.Errorf("Markup(EscapeMarkup(s)) = %q, want %q", got, dynamic)
	}
}

func TestSprint(t *testing.T) {
	// NoColor 환경에서 ANSI 코드가 제거되는지 확인하기 위해
	// Markup 자체가 ANSI 코드를 포함하는지 확인
	result := rich.Markup("[red]에러[/red]")
	if !strings.Contains(result, "\033[31m") {
		t.Errorf("Markup([red]) should contain red ANSI code, got %q", result)
	}
	// Sprint도 NoColor가 false이면 Markup 결과와 동일해야 함
	sprintResult := rich.Sprint("[red]에러[/red]")
	if result != sprintResult {
		t.Errorf("Sprint() = %q, want %q", sprintResult, result)
	}
}

func TestMarkup_TrueColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "TrueColor hex 전경",
			input:    "[#ff4500]주황[/#ff4500]",
			expected: "\033[38;2;255;69;0m주황",
		},
		{
			name:     "TrueColor hex 배경",
			input:    "[bg-#ff0000]빨간배경[/bg-#ff0000]",
			expected: "\033[48;2;255;0;0m빨간배경",
		},
		{
			name:     "256색 전경",
			input:    "[color(208)]주황[/color(208)]",
			expected: "\033[38;5;208m주황",
		},
		{
			name:     "256색 배경",
			input:    "[bg-color(196)]빨간배경[/bg-color(196)]",
			expected: "\033[48;5;196m빨간배경",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rich.Markup(tt.input)
			if !strings.Contains(got, tt.expected) {
				t.Errorf("Markup() = %q, want to contain %q", got, tt.expected)
			}
		})
	}
}

func TestMarkup_TrueColorStrip(t *testing.T) {
	// 비터미널 환경에서 TrueColor/256색 태그가 제거되는지 확인
	var buf bytes.Buffer
	rich.Fprintln(&buf, "[#ff4500]컬러[/#ff4500] [color(208)]256색[/color(208)]")
	got := strings.TrimRight(buf.String(), "\n")
	if strings.Contains(got, "[#ff4500]") || strings.Contains(got, "[color(208)]") {
		t.Errorf("비터미널 환경에서 컬러 태그가 제거되어야 함: %q", got)
	}
	if !strings.Contains(got, "컬러") || !strings.Contains(got, "256색") {
		t.Errorf("컬러 태그 내 텍스트는 보존되어야 함: %q", got)
	}
}
