package rich_test

import (
	"github.com/seoyc/wcli/rich"
	"strings"
	"testing"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rich.Markup(tt.input)
			if !strings.Contains(got, tt.expected) {
				t.Errorf("Markup() = %v, want to contain %v", got, tt.expected)
			}
		})
	}
}
