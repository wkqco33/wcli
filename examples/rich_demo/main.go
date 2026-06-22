package main

import (
	"fmt"
	"os"
	"time"

	"github.com/seoyc/wcli/rich"
)

func main() {
	rich.Println("[bold][cyan]✨ wcli/rich 컴포넌트 업그레이드 데모 ✨[/cyan][/bold]\n")

	// 1. 다양한 스피너 시연
	rich.Println("[bold][yellow]1. 다양한 스피너(Spinner) 스타일 프리셋[/yellow][/bold]")
	rich.Println("스피너는 비동기 작업 시 터미널에서 생동감 넘치는 애니메이션을 제공합니다.\n")

	spinners := []struct {
		name  string
		style rich.SpinnerStyle
	}{
		{"기본 브라이유 점자 (Default)", rich.SpinnerDefault},
		{"도트 애니메이션 (Dots)", rich.SpinnerDots},
		{"클래식 텍스트 회전 (Line)", rich.SpinnerLine},
		{"원형 회전 (Circle)", rich.SpinnerCircle},
		{"화살표 회전 (Arrow)", rich.SpinnerArrow},
		{"바운싱 블록 (Bouncing)", rich.SpinnerBouncing},
	}

	for _, ts := range spinners {
		s := rich.NewSpinner(os.Stdout)
		s.SetStyle(ts.style)
		s.Start(fmt.Sprintf("%s 실행 중...", ts.name))
		time.Sleep(1500 * time.Millisecond)
		s.Stop(fmt.Sprintf("[green]✓ %s 완료[/green]", ts.name))
		fmt.Println()
	}

	// 2. 다양한 프로그레스바 시연
	rich.Println("[bold][yellow]2. 다양한 프로그레스바(ProgressBar) 테마 및 옵션[/yellow][/bold]")
	rich.Println("진행률 표시는 터미널 환경에 맞는 다양한 문자와 색상 마크업을 지원합니다.\n")

	themes := []struct {
		name  string
		theme rich.ProgressTheme
		color string
	}{
		{"기본 블록 테마 (Block)", rich.ThemeBlock, "green"},
		{"라인 테마 (Line)", rich.ThemeLine, "cyan"},
		{"더블 라인 테마 (DoubleLine)", rich.ThemeDoubleLine, "blue"},
		{"블릿 테마 (Bullet)", rich.ThemeBullet, "yellow"},
		{"화살표 테마 (Arrow)", rich.ThemeArrow, "magenta"},
		{"별 테마 (Star)", rich.ThemeStar, "white"},
	}

	for _, tt := range themes {
		rich.Print("[bold]%s[/bold]\n", tt.name)
		pb := rich.NewProgressBar(100)
		pb.SetTheme(tt.theme)
		pb.FillColor = tt.color
		pb.EmptyColor = "dim"
		pb.ShowCounter = true
		pb.ShowETA = true
		pb.Start()

		// 진행 상태 모사
		for i := 0; i <= 100; i += 20 {
			pb.Print(i)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("\n")
	}

	rich.Println("[bold][green]🎉 데모 시연이 끝났습니다![/green][/bold]")
}
