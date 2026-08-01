package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

// wcliCommandsMarker wcli add 가 커맨드를 자동 주입할 때 사용하는 마커 주석
const wcliCommandsMarker = "// wcli:commands"

var (
	getwdFunc    = os.Getwd
	statFileFunc = os.Stat
	openFileFunc = os.Open
	readFileFunc = os.ReadFile
	writeFileFunc = os.WriteFile
)

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

	rootCmd.AddCommand(buildInitCmd(), buildAddCmd(), buildDoctorCmd())

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func buildInitCmd() *wcli.Command {
	var libPath string

	cmd := &wcli.Command{
		Use:   "init [module_name]",
		Short: "현재 디렉토리에 새 wcli CLI 프로젝트 생성",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("모듈 명칭을 지정해 주세요. 예: wcli init myapp")
			}

			modName := ctx.Args[0]
			appName := filepath.Base(modName)

			// wcli 라이브러리 경로 결정: 플래그 > .gitmodules 자동 탐지
			resolvedPath := libPath
			if resolvedPath == "" {
				detected, err := detectWcliPath(".")
				if err != nil {
					return fmt.Errorf(
						"wcli 라이브러리 경로를 찾을 수 없습니다.\n"+
							"  .gitmodules에 wcli submodule이 등록되어 있는지 확인하거나\n"+
							"  --lib-path 플래그로 경로를 직접 지정해 주세요.\n"+
							"  예: wcli init --lib-path ./wcli %s", modName)
				}
				resolvedPath = detected
				rich.Println("[dim]  wcli 경로 자동 탐지: %s[/dim]", resolvedPath)
			}

			// 절대 경로가 들어온 경우 현재 디렉토리 기준 상대 경로로 변환
			if filepath.IsAbs(resolvedPath) {
				wd, err := getwdFunc()
				if err != nil {
					return fmt.Errorf("현재 디렉토리 획득 실패: %w", err)
				}
				rel, err := filepath.Rel(wd, resolvedPath)
				if err != nil {
					return fmt.Errorf("상대 경로 변환 실패: %w", err)
				}
				resolvedPath = "./" + rel
			}

			rich.Println("[cyan]프로젝트 초기화 중...[/cyan] (모듈명: %s, 앱명: %s)", modName, appName)

			data := initData{
				ModuleName:  modName,
				LibraryPath: resolvedPath,
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

	cmd.Flags().StringVar(&libPath, "lib-path", "l", "", "wcli 라이브러리 경로 (기본값: .gitmodules 자동 탐지)")
	return cmd
}

// detectWcliPath .gitmodules 파일을 파싱해 wcli submodule의 상대 경로를 반환합니다.
func detectWcliPath(dir string) (string, error) {
	f, err := openFileFunc(filepath.Join(dir, ".gitmodules"))
	if err != nil {
		return "", fmt.Errorf(".gitmodules 파일을 열 수 없습니다: %w", err)
	}
	defer f.Close()

	// 현재 파싱 중인 submodule의 path 값
	var currentPath string
	isWcli := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 새 섹션 시작 - 이전 섹션이 wcli였으면 path 반환
		if strings.HasPrefix(line, "[submodule") {
			if isWcli && currentPath != "" {
				return "./" + currentPath, nil
			}
			isWcli = strings.Contains(strings.ToLower(line), "wcli")
			currentPath = ""
			continue
		}

		if !isWcli {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "path":
			currentPath = val
		case "url":
			// url에 wcli가 포함되면 wcli submodule로 확정
			if strings.Contains(strings.ToLower(val), "wcli") {
				isWcli = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf(".gitmodules 읽기 오류: %w", err)
	}

	// 파일 끝에서 마지막 섹션 처리
	if isWcli && currentPath != "" {
		return "./" + currentPath, nil
	}

	return "", fmt.Errorf(".gitmodules에서 wcli submodule을 찾을 수 없습니다")
}

func buildAddCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "add [command_name]",
		Short: "새 서브커맨드 파일 생성 및 바인딩 자동 등록",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("추가할 커맨드 명칭을 입력해 주세요. 예: wcli add create")
			}

			cmdName := strings.ToLower(strings.TrimSpace(ctx.Args[0]))
			if cmdName == "" {
				return fmt.Errorf("커맨드 이름이 비어 있습니다")
			}
			fileName := cmdName + ".go"

			// main.go 유무 체크하여 wcli 프로젝트인지 검증
			if _, err := statFileFunc("main.go"); os.IsNotExist(err) {
				return fmt.Errorf("wcli 프로젝트의 루트 디렉토리가 아닙니다 (main.go가 존재하지 않습니다)")
			}

			// 구조체 이름 생성 (첫 글자 대문자화 + Cmd 접미사, 멀티바이트 안전)
			runes := []rune(cmdName)
			structName := strings.ToUpper(string(runes[:1])) + string(runes[1:]) + "Cmd"

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
	if _, err := statFileFunc(fileName); err == nil {
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

	if err := writeFileFunc(fileName, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("파일 기록 에러: %w", err)
	}

	rich.Println("  [dim]+ 파일 생성 완료: %s[/dim]", fileName)
	return nil
}

func injectCommandToMain(structName string) error {
	content, err := readFileFunc("main.go")
	if err != nil {
		return err
	}

	mainStr := string(content)

	if !strings.Contains(mainStr, wcliCommandsMarker) {
		return fmt.Errorf("main.go 파일에 '%s' 마커 주석이 보이지 않습니다", wcliCommandsMarker)
	}

	bindingCode := wcliCommandsMarker + "\n\trootCmd.AddCommand(" + structName + ")"
	newMainStr := strings.Replace(mainStr, wcliCommandsMarker, bindingCode, 1)

	return writeFileFunc("main.go", []byte(newMainStr), 0644)
}

// checkResult doctor 점검 결과 항목
type checkResult struct {
	Name   string
	Status string // "ok" | "warn" | "fail"
	Detail string
}

func buildDoctorCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "doctor",
		Short: "현재 디렉토리의 wcli 프로젝트 상태 점검",
		Run: func(ctx *wcli.Context) error {
			results := runDoctor()

			tbl := rich.NewTable("점검 항목", "상태", "상세")
			for _, r := range results {
				var status string
				switch r.Status {
				case "ok":
					status = "[green]ok[/green]"
				case "warn":
					status = "[yellow]warn[/yellow]"
				default:
					status = "[red]fail[/red]"
				}
				tbl.AddRow(r.Name, status, r.Detail)
			}
			tbl.Print()
			return nil
		},
	}
}

func runDoctor() []checkResult {
	var results []checkResult

	// 1. main.go 존재 여부
	if _, err := statFileFunc("main.go"); err == nil {
		results = append(results, checkResult{"main.go 존재", "ok", "main.go 발견"})
	} else {
		results = append(results, checkResult{"main.go 존재", "fail", "main.go 없음 — wcli 프로젝트 루트가 아닌 것 같습니다"})
	}

	// 2. go.mod 존재 + wcli 의존성 포함 여부
	goModContent, err := os.ReadFile("go.mod")
	if err != nil {
		results = append(results, checkResult{"go.mod 존재", "fail", "go.mod 없음"})
	} else {
		results = append(results, checkResult{"go.mod 존재", "ok", "go.mod 발견"})
		if strings.Contains(string(goModContent), "seoyc/wcli") {
			results = append(results, checkResult{"wcli 의존성", "ok", "go.mod에 wcli 의존성 포함"})
		} else {
			results = append(results, checkResult{"wcli 의존성", "warn", "go.mod에 seoyc/wcli 의존성이 없습니다"})
		}

		// 3. replace 경로 유효성 검사
		for _, line := range strings.Split(string(goModContent), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "replace") {
				continue
			}
			// "replace X => ./path" 형식에서 경로 추출
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			replacePath := parts[len(parts)-1]
			if strings.HasPrefix(replacePath, ".") || strings.HasPrefix(replacePath, "/") {
				if _, err := os.Stat(replacePath); err == nil {
					results = append(results, checkResult{"replace 경로 유효성", "ok", replacePath + " 존재"})
				} else {
					results = append(results, checkResult{"replace 경로 유효성", "fail", replacePath + " 경로를 찾을 수 없습니다"})
				}
			}
		}
	}

	// 4. main.go 에 wcliCommandsMarker 존재 여부
	mainContent, err := os.ReadFile("main.go")
	if err == nil {
		if strings.Contains(string(mainContent), wcliCommandsMarker) {
			results = append(results, checkResult{"wcli:commands 마커", "ok", "main.go에 마커 존재"})
		} else {
			results = append(results, checkResult{"wcli:commands 마커", "warn", "main.go에 '// wcli:commands' 마커가 없습니다 (wcli add 자동 등록 불가)"})
		}
	}

	return results
}
