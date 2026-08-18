package rich_test

import (
	"strings"
	"testing"
	"time"

	"github.com/wkqco33/wcli/rich"
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

func TestProgressBar_Themes(t *testing.T) {
	themes := []rich.ProgressTheme{
		rich.ThemeBlock,
		rich.ThemeLine,
		rich.ThemeDoubleLine,
		rich.ThemeBullet,
		rich.ThemeArrow,
		rich.ThemeStar,
	}

	for _, theme := range themes {
		pb := rich.NewProgressBar(10)
		pb.SetTheme(theme)
		result := pb.Render(5)
		if !strings.Contains(result, theme.Fill) && theme.Fill != " " {
			t.Errorf("테마 Fill 문자가 출력에 없습니다: %q (theme.Fill: %s)", result, theme.Fill)
		}
		if !strings.Contains(result, theme.Empty) && theme.Empty != " " {
			t.Errorf("테마 Empty 문자가 출력에 없습니다: %q (theme.Empty: %s)", result, theme.Empty)
		}
	}
}

func TestProgressBar_Colors(t *testing.T) {
	pb := rich.NewProgressBar(10)
	pb.FillColor = "blue"
	pb.EmptyColor = "dim"
	result := pb.Render(5)

	if !strings.Contains(result, "[blue]") {
		t.Errorf("커스텀 FillColor 마크업이 없습니다: %q", result)
	}
	if !strings.Contains(result, "[dim]") {
		t.Errorf("커스텀 EmptyColor 마크업이 없습니다: %q", result)
	}
}

func TestProgressBar_Options(t *testing.T) {
	pb := rich.NewProgressBar(100)
	pb.ShowPercent = false
	pb.ShowCounter = true
	result := pb.Render(30)

	if strings.Contains(result, "%") {
		t.Errorf("ShowPercent가 false일 때 퍼센트 표시가 없어야 합니다: %q", result)
	}
	if !strings.Contains(result, "(30/100)") {
		t.Errorf("ShowCounter가 true일 때 카운터 표시가 있어야 합니다: %q", result)
	}
}

func TestProgressBar_ETA(t *testing.T) {
	pb := rich.NewProgressBar(100)
	pb.ShowETA = true
	pb.Start()

	// 테스트 환경에서 시간 경과를 시뮬레이션하기 위해
	// 아주 살짝 대기하거나 렌더링 검사
	time.Sleep(10 * time.Millisecond)
	result := pb.Render(50)

	if !strings.Contains(result, "ETA:") {
		t.Errorf("ShowETA가 true이고 Start 되었을 때 ETA가 표시되어야 합니다: %q", result)
	}
}
