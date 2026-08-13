package main

// GoModTemplate go.mod 파일 뼈대 템플릿
const GoModTemplate = `module {{.ModuleName}}

go 1.26.1

require github.com/seoyc/wcli v0.0.0

replace github.com/seoyc/wcli => {{.LibraryPath}}
`

// MainTemplate main.go 파일 뼈대 템플릿
const MainTemplate = `package main

import (
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/logging"
	"github.com/seoyc/wcli/rich"
)

func main() {
	var verbose bool

	// 기본 로거 주입
	logger := logging.NewDefaultLogger(os.Stderr, logging.LevelInfo, true)
	logging.SetLogger(logger)

	rootCmd := &wcli.Command{
		Use:     "{{.AppName}}",
		Short:   "{{.AppName}} CLI 애플리케이션",
		Long:    "wcli로 생성된 {{.AppName}} CLI 프로그램입니다.",
		Version: "0.1.0",
		PersistentPreRun: func(ctx *wcli.Context) error {
			if verbose {
				if dl, ok := ctx.Logger.(*logging.DefaultLogger); ok {
					dl.MinLevel = logging.LevelDebug
				}
			}
			return nil
		},
		Run: func(ctx *wcli.Context) error {
			rich.Println("[green][bold]안녕하세요! {{.AppName}}이(가) 작동할 준비가 되었습니다.[/bold][/green]")
			rich.Println("  도움말 확인: [cyan]go run main.go --help[/cyan]")
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", "v", false, "상세 출력 활성화")

	// wcli:commands

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
`

// CommandTemplate 서브커맨드 파일용 템플릿
const CommandTemplate = `package main

import (
	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

var {{.CmdStructName}} = &wcli.Command{
	Use:   "{{.CmdName}}",
	Short: "{{.CmdName}} 명령어 설명",
	Run: func(ctx *wcli.Context) error {
		rich.Println("[cyan]{{.CmdName}} 명령어가 성공적으로 실행되었습니다.[/cyan]")
		return nil
	},
}
`

// TaskfileTemplate 빌드 편의성을 돕는 Taskfile 템플릿
const TaskfileTemplate = `version: '3'

tasks:
  default:
    aliases: [help]
    silent: true
    cmds:
      - task --list

  build:
    desc: 컴파일 오류 확인
    cmds:
      - go build ./...

  test:
    desc: 전체 테스트 실행
    cmds:
      - go test ./...

  run:
    desc: 애플리케이션 실행
    cmds:
      - go run main.go
`
