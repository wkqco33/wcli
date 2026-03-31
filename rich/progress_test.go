package rich_test

import (
	"strings"
	"testing"

	"github.com/seoyc/wcli/rich"
)

func TestProgressBar_Render(t *testing.T) {
	tests := []struct {
		name         string
		total        int
		current      int
		wantContains string
		wantEmpty    bool
	}{
		{
			name:         "0% 진행률",
			total:        100,
			current:      0,
			wantContains: "  0%",
		},
		{
			name:         "50% 진행률",
			total:        100,
			current:      50,
			wantContains: " 50%",
		},
		{
			name:         "100% 진행률",
			total:        100,
			current:      100,
			wantContains: "100%",
		},
		{
			name:         "초과값은 100%로 처리",
			total:        10,
			current:      20,
			wantContains: "100%",
		},
		{
			name:      "total이 0이면 빈 문자열",
			total:     0,
			current:   5,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := rich.NewProgressBar(tt.total)
			got := pb.Render(tt.current)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("Render()=%q, want empty string", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("Render()=%q, want to contain %q", got, tt.wantContains)
			}
		})
	}
}

func TestProgressBar_FillEmpty(t *testing.T) {
	pb := rich.NewProgressBar(10)
	pb.Fill = "#"
	pb.Empty = "-"
	result := pb.Render(5)
	if !strings.Contains(result, "#") {
		t.Errorf("커스텀 Fill 문자가 출력에 없습니다: %q", result)
	}
	if !strings.Contains(result, "-") {
		t.Errorf("커스텀 Empty 문자가 출력에 없습니다: %q", result)
	}
}
