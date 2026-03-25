# wcli

`wcli`는 Go 언어로 작성된 가볍고 성능 중심적인 CLI 명령어 라이브러리입니다.
[Cobra](https://github.com/spf13/cobra)에서 영감을 받아 제작되었으며, Python의 [`rich`](https://github.com/Textualize/rich) 라이브러리처럼 **마크업 기반 터미널 컬러 출력**을 내장합니다.

## 특징

- **제로 의존성**: 외부 패키지 없이 순수 Go 표준 라이브러리만 사용
- **성능 최우선**: O(1) 커맨드 라우팅 맵, 정렬 캐시, 마크업 파싱 캐시
- **데이터 중심**: 직관적인 `Command` 구조체로 명령어 트리 구성
- **Rich 텍스트 출력**: 마크업 기반 터미널 컬러/스타일 지원, TTY 자동 감지
- **Cobra 호환 API**: `Persistent` 플래그/훅, 서브커맨드 별칭, `--name=value` 문법 등

## 빠른 시작

```go
package main

import (
    "os"
    "github.com/seoyc/wcli"
    "github.com/seoyc/wcli/rich"
)

func main() {
    var verbose bool
    var name string

    rootCmd := &wcli.Command{
        Use:     "app",
        Short:   "앱 설명",
        Version: "1.0.0",
        Run: func(ctx *wcli.Context) error {
            if verbose {
                rich.Println("[blue]상세 모드 활성화[/blue]")
            }
            rich.Println("[green]안녕하세요, %s님![/green]", name)
            return nil
        },
    }

    rootCmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력 활성화")
    rootCmd.Flags().StringVar(&name, "name", "n", "User", "사용자 이름")

    if err := rootCmd.Execute(os.Args[1:]); err != nil {
        os.Exit(1)
    }
}
```

## 커맨드 구조

### 기본 구조

```go
cmd := &wcli.Command{
    Use:     "serve [address]",  // 첫 토큰이 커맨드 이름으로 사용됨
    Short:   "서버를 시작합니다",
    Long:    "지정된 주소에서 HTTP 서버를 시작합니다.",
    Aliases: []string{"server", "s"},
    Version: "2.1.0",
    Run: func(ctx *wcli.Context) error {
        // ctx.Args에 플래그 파싱 후 남은 positional 인자
        return nil
    },
}
```

### 서브커맨드

```go
rootCmd := &wcli.Command{Use: "app", Short: "앱"}

getCmd := &wcli.Command{
    Use:   "get [resource]",
    Short: "리소스 조회",
    Run:   func(ctx *wcli.Context) error { return nil },
}

rootCmd.AddCommand(getCmd)
rootCmd.Execute(os.Args[1:])
```

### 실행 훅

```go
cmd := &wcli.Command{
    Use: "deploy",
    PreRun: func(ctx *wcli.Context) error {
        // Run 실행 전
        return nil
    },
    Run: func(ctx *wcli.Context) error {
        return nil
    },
    PostRun: func(ctx *wcli.Context) error {
        // Run 성공 후
        return nil
    },
    // 모든 하위 커맨드에도 적용되는 훅
    PersistentPreRun: func(ctx *wcli.Context) error {
        return nil
    },
    PersistentPostRun: func(ctx *wcli.Context) error {
        return nil
    },
}
```

실행 순서: `(루트→현재) PersistentPreRun` → `PreRun` → `Run` → `PostRun` → `(현재→루트) PersistentPostRun`

## 플래그

### 지원 타입

| 메서드 | 타입 | 예시 |
|---|---|---|
| `StringVar` | `string` | `--name alice`, `--name=alice` |
| `IntVar` | `int` | `--count 5`, `--count=5` |
| `BoolVar` | `bool` | `--verbose`, `--verbose=false` |
| `Float64Var` | `float64` | `--ratio 3.14` |
| `DurationVar` | `time.Duration` | `--timeout 30s`, `--timeout=1h30m` |
| `StringSliceVar` | `[]string` | `--tag foo --tag bar` (누적) |

### 플래그 등록

```go
var (
    name    string
    count   int
    verbose bool
    ratio   float64
    timeout time.Duration
    tags    []string
)

cmd.Flags().StringVar(&name, "name", "n", "default", "사용자 이름")
cmd.Flags().IntVar(&count, "count", "c", 0, "횟수")
cmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")
cmd.Flags().Float64Var(&ratio, "ratio", "r", 1.0, "비율")
cmd.Flags().DurationVar(&timeout, "timeout", "t", 5*time.Second, "타임아웃")
cmd.Flags().StringSliceVar(&tags, "tag", "", nil, "태그 (반복 가능)")
```

### Persistent 플래그 (하위 커맨드 상속)

```go
// rootCmd에 등록된 persistent 플래그는 모든 하위 커맨드에서도 사용 가능
var verbose bool
rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")
```

help 출력 시 상속된 플래그는 `Global Flags:` 섹션으로 분리 표시됩니다.

### 필수 플래그 & 검증

```go
cmd.Flags().StringVar(&output, "output", "o", "", "출력 파일")
cmd.Flags().MarkRequired("output")  // 미설정 시 에러

cmd.Flags().IntVar(&port, "port", "p", 8080, "포트")
cmd.Flags().SetValidation("port", func(val string) error {
    n, _ := strconv.Atoi(val)
    if n < 1 || n > 65535 {
        return fmt.Errorf("포트는 1~65535 범위여야 합니다")
    }
    return nil
})
```

### 특수 문법

```bash
# --name=value 인라인 값
app --output=file.txt --count=5 --verbose=false

# -- 종결자: 이후 모든 인자는 플래그로 해석하지 않음
app --name foo -- --not-a-flag positional-arg
```

## 출력 제어

### OutWriter / ErrWriter

테스트나 파이프 출력에 유용합니다.

```go
var buf bytes.Buffer
cmd := &wcli.Command{
    Use:       "app",
    OutWriter: &buf,   // 기본값: os.Stdout
    ErrWriter: &buf,   // 기본값: os.Stderr
    Run:       func(ctx *wcli.Context) error { return nil },
}
```

서브커맨드는 부모의 Writer를 자동으로 상속합니다.

### NoColor / TTY 감지

```go
// 환경변수로 제어 (no-color.org 표준)
NO_COLOR=1 ./app

// 프로그래밍 방식
rich.NoColor = true
```

파이프(`./app | cat`)나 파일 리다이렉션 시 ANSI 코드가 자동으로 제거됩니다.

## Rich 마크업

`rich` 패키지는 `[태그]텍스트[/태그]` 형식의 마크업을 ANSI 코드로 변환합니다.

### 지원 태그

| 태그 | 효과 |
|---|---|
| `[bold]` | 굵게 |
| `[dim]` | 흐리게 |
| `[underline]` | 밑줄 |
| `[red]` `[green]` `[yellow]` | 텍스트 색상 |
| `[blue]` `[magenta]` `[cyan]` `[white]` | 텍스트 색상 |

### 사용법

```go
rich.Println("[bold][green]성공![/green][/bold]")
rich.Println("[red]오류: %s[/red]", err.Error())

// io.Writer에 출력
rich.Fprintln(w, "[cyan]%s[/cyan]", message)

// 마크업 문자열로만 변환 (출력 없음)
styled := rich.Markup("[yellow]경고[/yellow]")

// 이스케이프
rich.Println(`\[bold] 괄호 그대로 출력`)
```

중첩 태그는 닫힘 순서에 관계없이 올바르게 처리됩니다:
```go
rich.Println("[bold][red]텍스트[/bold][/red]")  // 정상 동작
```

## 커스텀 도움말

```go
cmd := &wcli.Command{
    Use: "app",
    HelpFunc: func(cmd *wcli.Command, w io.Writer) {
        fmt.Fprintln(w, "커스텀 도움말 출력")
        fmt.Fprintf(w, "버전: %s\n", cmd.Version)
    },
}
```

`HelpFunc`가 `nil`이면 기본 도움말이 출력됩니다.

## Makefile 타겟

```
make build       # 컴파일 오류 확인
make test        # 전체 테스트
make test-v      # 상세 테스트 출력
make bench       # 벤치마크
make cover       # 커버리지 HTML 리포트
make check       # fmt-check + vet + test
make fmt         # 코드 포맷 적용
make vet         # 정적 분석
make tidy        # go mod tidy
make clean       # 빌드 아티팩트 제거
```

## 에러 처리

```go
var (
    ErrCommandNotFound = errors.New("command not found")
    ErrHelp            = errors.New("help requested")  // Execute()에서 nil로 변환
)

cmd := &wcli.Command{
    SilenceErrors: true,  // 에러 자동 출력 억제
    Run: func(ctx *wcli.Context) error {
        return fmt.Errorf("커스텀 에러")
    },
}
```
