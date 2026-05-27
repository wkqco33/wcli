// logging_error는 wcli의 경량 로깅 시스템과 구조화된 에러 처리 기능을 보여주는 예제 앱입니다.
//
// 실행 예시:
//
//	# 정상 실행 (로깅 연동 확인)
//	go run ./examples/logging_error --name "길동" --debug
//
//	# 플래그 구문 에러 유도 (FlagError 발생)
//	go run ./examples/logging_error --invalid-flag
//
//	# 필수 플래그 누락 에러 유도 (ValidationError 발생)
//	go run ./examples/logging_error
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/logging"
	"github.com/seoyc/wcli/rich"
)

func main() {
	var (
		name  string
		debug bool
	)

	// 1. 기본 NoOpLogger 대신 DefaultLogger 생성 및 주입
	// stdout에 출력하며, INFO 레벨 이상만, rich 마크업 서식을 적용해 출력하도록 설정합니다.
	logger := logging.NewDefaultLogger(os.Stderr, logging.LevelInfo, true)
	logging.SetLogger(logger)

	rootCmd := &wcli.Command{
		Use:   "app",
		Short: "로깅 및 에러 테스트 데모 앱",
		Long:  "wcli의 경량 로깅 패키지 및 구조화된 에러 핸들링 기능을 확인해보는 데모 CLI입니다.",
		// 에러를 직접 처리하고 포맷팅하기 위해 자동 출력을 꺼줍니다.
		SilenceErrors: true,

		PersistentPreRun: func(ctx *wcli.Context) error {
			// --debug 플래그가 설정된 경우 로거의 최소 출력 레벨을 DEBUG로 동적 완화합니다.
			if debug {
				if dl, ok := ctx.Logger.(*logging.DefaultLogger); ok {
					dl.MinLevel = logging.LevelDebug
					ctx.Logger.Log(logging.LevelDebug, "디버그 모드가 켜졌습니다. 모든 레벨의 로그를 출력합니다.")
				}
			}
			return nil
		},

		Run: func(ctx *wcli.Context) error {
			ctx.Logger.Log(logging.LevelInfo, "비즈니스 로직 시작...")
			rich.Println("[green]안녕하세요, %s님![/green]", name)
			ctx.Logger.Log(logging.LevelInfo, "비즈니스 로직 정상 종료.")
			return nil
		},
	}

	rootCmd.Flags().StringVar(&name, "name", "n", "", "사용자 이름")
	rootCmd.Flags().BoolVar(&debug, "debug", "d", false, "디버그 로그 활성화")

	// --name은 필수 플래그로 설정
	_ = rootCmd.Flags().MarkRequired("name")

	// 2. 명령어 실행 및 에러 분기 처리
	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		fmt.Println()
		rich.Println("[bold][red]❌ 에러 발생 분석 결과:[/red][/bold]")

		// 에러 타입에 따른 세부 분기 및 데이터 스키마 접근
		var flagErr *wcli.FlagError
		var valErr *wcli.ValidationError
		var cmdErr *wcli.CommandError

		if errors.As(err, &flagErr) {
			rich.Println("  [yellow][구문 오류][/yellow] 잘못된 옵션 입력")
			rich.Println("    - 대상 플래그: [cyan]--%s[/cyan]", flagErr.FlagName)
			rich.Println("    - 오류 내용: %v", flagErr.Err)
		} else if errors.As(err, &valErr) {
			rich.Println("  [yellow][검증 오류][/yellow] 제약 조건 위반")
			rich.Println("    - 대상 플래그: [cyan]--%s[/cyan]", valErr.FlagName)
			rich.Println("    - 오류 내용: %v", valErr.Err)
		} else if errors.As(err, &cmdErr) {
			rich.Println("  [yellow][실행 오류][/yellow] 시스템 처리 실패")
			rich.Println("    - 명령어 명칭: [cyan]%s[/cyan]", cmdErr.CommandName)
			rich.Println("    - 오류 내용: %v", cmdErr.Err)
		} else {
			rich.Println("  일반 에러: %v", err)
		}
		os.Exit(1)
	}
}
