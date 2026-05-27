package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

const defaultLibraryPath = "/home/wkqco/Workspace/wcli"

type initData struct {
	ModuleName  string
	LibraryPath string
	AppName     string
}

type cmdData struct {
	CmdName       string
	CmdStructName string
}

func main() {
	var rootCmd *wcli.Command
	rootCmd = &wcli.Command{
		Use:   "wcli",
		Short: "wcli 기반 프로젝트 스캐폴더 및 개발 도구",
		Long:  "wcli 명령어 생성 및 초기 보일러플레이트 구성을 지원하는 코드 생성 도구입니다.",
		Run: func(ctx *wcli.Context) error {
			rootCmd.Help()
			return nil
		},
	}

	rootCmd.AddCommand(buildInitCmd(), buildAddCmd())

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func buildInitCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "init [module_name]",
		Short: "현재 디렉토리에 새 wcli CLI 프로젝트 생성",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("모듈 명칭을 지정해 주세요. 예: wcli init myapp")
			}

			modName := ctx.Args[0]
			appName := filepath.Base(modName)

			rich.Println("[cyan]프로젝트 초기화 중...[/cyan] (모듈명: %s, 앱명: %s)", modName, appName)

			data := initData{
				ModuleName:  modName,
				LibraryPath: defaultLibraryPath,
				AppName:     appName,
			}

			// 1. go.mod 작성
			if err := renderToFile("go.mod", GoModTemplate, data); err != nil {
				return err
			}

			// 2. main.go 작성
			if err := renderToFile("main.go", MainTemplate, data); err != nil {
				return err
			}

			// 3. Makefile 작성
			if err := renderToFile("Makefile", MakefileTemplate, data); err != nil {
				return err
			}

			rich.Println("[green][bold]✓ 프로젝트가 성공적으로 초기화되었습니다![/bold][/green]")
			rich.Println("  실행 명령어:")
			rich.Println("    - 의존성 다운로드: [cyan]go mod tidy[/cyan]")
			rich.Println("    - 앱 실행        : [cyan]go run main.go[/cyan]")
			return nil
		},
	}
}

func buildAddCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "add [command_name]",
		Short: "새 서브커맨드 파일 생성 및 바인딩 자동 등록",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("추가할 커맨드 명칭을 입력해 주세요. 예: wcli add create")
			}

			cmdName := strings.ToLower(ctx.Args[0])
			fileName := cmdName + ".go"

			// main.go 유무 체크하여 wcli 프로젝트인지 검증
			if _, err := os.Stat("main.go"); os.IsNotExist(err) {
				return fmt.Errorf("wcli 프로젝트의 루트 디렉토리가 아닙니다 (main.go가 존재하지 않습니다)")
			}

			// 구조체 이름 생성 (첫 글자 대문자화 + Cmd 접미사)
			structName := strings.Title(cmdName) + "Cmd"

			rich.Println("[cyan]커맨드 추가 중...[/cyan] (파일명: %s, 구조체명: %s)", fileName, structName)

			data := cmdData{
				CmdName:       cmdName,
				CmdStructName: structName,
			}

			// 1. <cmd_name>.go 뼈대 작성
			if err := renderToFile(fileName, CommandTemplate, data); err != nil {
				return err
			}

			// 2. main.go 에 등록코드 자동 주입
			if err := injectCommandToMain(structName); err != nil {
				rich.Println("[yellow][경고] main.go에 명령어 자동 등록을 실패했습니다. 수동으로 등록하세요.[/yellow] (%v)", err)
				rich.Println("  등록 방법: main.go 에 [cyan]rootCmd.AddCommand(%s)[/cyan] 구문을 추가해 주세요.", structName)
			} else {
				rich.Println("[green]✓ main.go 에 %s가 성공적으로 자동 등록되었습니다.[/green]", structName)
			}

			rich.Println("[green][bold]✓ 커맨드가 정상적으로 추가되었습니다![/bold][/green]")
			return nil
		},
	}
}

func renderToFile(fileName, tmplStr string, data interface{}) error {
	// 이미 파일이 존재하는지 체크
	if _, err := os.Stat(fileName); err == nil {
		return fmt.Errorf("파일이 이미 존재합니다: %s", fileName)
	}

	tmpl, err := template.New(fileName).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("템플릿 파싱 에러: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("템플릿 실행 에러: %w", err)
	}

	if err := ioutil.WriteFile(fileName, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("파일 기록 에러: %w", err)
	}

	rich.Println("  [dim]+ 파일 생성 완료: %s[/dim]", fileName)
	return nil
}

func injectCommandToMain(structName string) error {
	content, err := ioutil.ReadFile("main.go")
	if err != nil {
		return err
	}

	mainStr := string(content)
	marker := "// wcli:commands"

	if !strings.Contains(mainStr, marker) {
		return fmt.Errorf("main.go 파일에 '%s' 마커 주석이 보이지 않습니다", marker)
	}

	bindingCode := fmt.Sprintf("// wcli:commands\n\trootCmd.AddCommand(%s)", structName)
	newMainStr := strings.Replace(mainStr, marker, bindingCode, 1)

	return ioutil.WriteFile("main.go", []byte(newMainStr), 0644)
}
