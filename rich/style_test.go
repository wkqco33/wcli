package rich_test

import (
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
