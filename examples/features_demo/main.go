// features_demo는 wcli에 추가된 4대 핵심 기능들을 종합 시연하는 통합 예제 앱입니다.
//
// 1. 셸 자동완성 명령어 등록
// 2. 플래그 간 상호 배제 제약 검사
// 3. 플래그 필수 동반 지정 제약 검사
// 4. 환경 변수 자동 바인딩
// 5. 커스텀 템플릿 기반 도움말 출력
//
// 실행 가이드:
//
//	# 1. 템플릿 도움말 확인 (커스텀 템플릿 렌더링 확인)
//	go run ./examples/features_demo --help
//
//	# 2. 셸 자동완성 출력 확인
//	go run ./examples/features_demo completion zsh
//
//	# 3. 환경 변수 자동 바인딩 테스트
//	API_TOKEN="example-token" go run ./examples/features_demo --host localhost
//
//	# 4. 상호 배제 플래그 에러 발생 테스트 (--json과 --yaml을 같이 지정 시 에러)
//	API_TOKEN="example-token" go run ./examples/features_demo --host localhost --json --yaml
//
//	# 5. 필수 동반 지정 플래그 에러 발생 테스트 (--user를 지정했으나 --password 누락 시 에러)
//	API_TOKEN="example-token" go run ./examples/features_demo --host localhost --user alice
package main

import (
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

// 커스텀 도움말 템플릿 (미니멀하고 세련된 형태)
const customHelpTemplate = `[bold][cyan]⚡ {{.Name}} - 피처 통합 데모 앱[/cyan][/bold]
{{.Short}}

[yellow]💡 사용법:[/yellow]
  {{.UsageLine}}

[yellow]⚙️ 로컬 플래그 목록:[/yellow]
{{range .LocalFlags}}  [green]--{{.Name | pad 10}}[/green] : {{.Usage}}{{if .DefaultPart}} (기본값: {{.DefaultPart}}){{end}}
{{end}}
더 자세한 명령별 가이드는 --help 옵션을 참조하세요.
`

func main() {
	var (
		host     string
		apiToken string
		jsonOut  bool
		yamlOut  bool
		user     string
		password string
	)

	rootCmd := &wcli.Command{
		Use:   "features_demo",
		Short: "wcli 4대 신규 고급 피처 통합 시연 데모",
		Long:  "이 데모는 wcli의 확장된 기능(환경변수 자동 매핑, 상호배제, 동반입력, 셸자동완성, 커스텀 템플릿 도움말)을 시연합니다.",
		// 에러를 직접 예쁘게 덤프하기 위해 SilenceErrors를 켭니다.
		SilenceErrors: true,

		// 커스텀 도움말 템플릿을 루트 커맨드에 주입합니다.
		HelpTemplate: customHelpTemplate,

		Run: func(ctx *wcli.Context) error {
			rich.Println("[bold][green]🎉 성공적으로 검증을 통과했습니다![/green][/bold]")
			fmt.Printf("  - 호스트 주소  : %s\n", host)
			fmt.Printf("  - API 토큰     : %s (바인딩 완료)\n", apiToken)
			fmt.Printf("  - JSON 출력 여부: %t\n", jsonOut)
			fmt.Printf("  - YAML 출력 여부: %t\n", yamlOut)
			if user != "" {
				fmt.Printf("  - 계정/비밀번호: %s / %s (동반 바인딩 완료)\n", user, password)
			}
			return nil
		},
	}

	// 셸 자동 완성 명령어(completion)를 하위 커맨드로 등록합니다.
	rootCmd.AddCommand(wcli.NewCompletionCommand(rootCmd))

	// 플래그 설정
	rootCmd.Flags().StringVar(&host, "host", "h", "localhost", "대상 호스트 주소")
	rootCmd.Flags().StringVar(&apiToken, "token", "t", "", "API 접속 인증 토큰 (환경변수 API_TOKEN 바인딩)")
	rootCmd.Flags().BoolVar(&jsonOut, "json", "j", false, "출력 형식을 JSON으로 설정 (YAML과 상호 배제)")
	rootCmd.Flags().BoolVar(&yamlOut, "yaml", "y", false, "출력 형식을 YAML로 설정 (JSON과 상호 배제)")
	rootCmd.Flags().StringVar(&user, "user", "u", "", "로그인 계정 (password와 동반 필수)")
	rootCmd.Flags().StringVar(&password, "password", "p", "", "로그인 비밀번호 (user와 동반 필수)")

	// 1. 환경 변수 자동 매핑 (apiToken 플래그가 지정 안 되면 env API_TOKEN에서 탐색)
	_ = rootCmd.Flags().BindEnv("token", "API_TOKEN")

	// 2. 상호 배제 설정 (--json과 --yaml은 공존할 수 없음)
	rootCmd.Flags().MarkFlagsMutuallyExclusive("json", "yaml")

	// 3. 필수 동반 지정 설정 (--user와 --password는 하나라도 설정되면 반드시 쌍으로 와야 함)
	rootCmd.Flags().MarkFlagsRequiredTogether("user", "password")

	// 4. 루트에 지정된 token은 실행 시 필수 플래그로 검증함 (환경변수로 공급되든 플래그로 공급되든 무관)
	_ = rootCmd.Flags().MarkRequired("token")

	// 명령어 실행
	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		rich.Println("[bold][red]❌ 실행 에러 감지:[/red][/bold] %v", err)
		os.Exit(1)
	}
}
