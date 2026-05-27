# wcli

`wcli`는 Go 언어로 작성된 가볍고 성능 중심적인 CLI 명령어 라이브러리입니다.
[Cobra](https://github.com/spf13/cobra)에서 영감을 받아 제작되었으며, Python의 [`rich`](https://github.com/Textualize/rich) 라이브러리처럼 **마크업 기반 터미널 컬러 출력**을 내장합니다.

## 특징

- **제로 의존성**: 외부 패키지 없이 순수 Go 표준 라이브러리만 사용
- **성능 최우선**: O(1) 커맨드 라우팅 맵, 정렬 캐시, 마크업 파싱 캐시
- **데이터 중심**: 직관적인 `Command` 구조체로 명령어 트리 구성
- **Rich 텍스트 출력**: 마크업 기반 터미널 컬러/스타일 지원, TTY 자동 감지
- **Cobra 호환 API**: `Persistent` 플래그/훅, 서브커맨드 별칭, `--name=value` 문법 등
- **구조화된 에러 처리**: `FlagError`, `ValidationError` 등 정밀한 디버깅 및 세부 속성 추출 지원
- **경량 로깅 서브패키지**: 런타임 오버헤드가 극히 적은 콘솔 로깅 및 동적 로그 레벨 제어 분리 제공

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

`wcli`는 구체적인 원인 추적이 가능하도록 구조화된 에러 타입들을 제공합니다. `errors.As`를 통해 에러 원인을 분기하여 구체적인 속성에 접근할 수 있습니다.

- `*wcli.FlagError`: 옵션/플래그 입력 구문 에러 (알 수 없는 플래그, 잘못된 형태 등)
- `*wcli.ValidationError`: 플래그 값 유효성 및 필수 제약 조건 실패 에러
- `*wcli.CommandError`: 훅 실행 도중 발생한 에러 및 라이브러리 내부 에러

```go
cmd := &wcli.Command{
    SilenceErrors: true,  // 에러 화면 자동 출력 억제
    Run: func(ctx *wcli.Context) error {
        return fmt.Errorf("일반 비즈니스 에러")
    },
}

if err := cmd.Execute(os.Args[1:]); err != nil {
    var flagErr *wcli.FlagError
    if errors.As(err, &flagErr) {
        fmt.Printf("플래그 에러 발생: %s\n", flagErr.FlagName)
    }
}
```

## 로깅 (Logging)

성능 최우선 원칙에 따라 설계된 경량 로거 서브패키지 `logging`을 제공합니다. 

```go
import "github.com/seoyc/wcli/logging"

func main() {
    // 1. DefaultLogger 생성 및 전역 설정
    logger := logging.NewDefaultLogger(os.Stderr, logging.LevelInfo, true)
    logging.SetLogger(logger)

    cmd := &wcli.Command{
        Use: "app",
        Run: func(ctx *wcli.Context) error {
            // 2. 훅 내에서 context를 통해 전달받은 로거 사용
            ctx.Logger.Log(logging.LevelInfo, "작업을 시작합니다.")
            return nil
        },
    }
    cmd.Execute(os.Args[1:])
}
```

## 플래그 확장 조건 및 검증

플래그 관계 제약 조건 및 환경 변수 자동 주입 기능을 내장하고 있습니다.

### 환경 변수 자동 매핑 (BindEnv)
특정 플래그가 지정되지 않은 경우, 대체해서 값을 읽어올 환경 변수를 지정합니다.
```go
cmd.Flags().StringVar(&token, "token", "t", "", "인증 토큰")
// --token이 공백일 시 환경변수 API_TOKEN을 조회하여 바인딩
_ = cmd.Flags().BindEnv("token", "API_TOKEN")
```

### 상호 배제 조건 지정 (Mutually Exclusive)
지정된 옵션 그룹 중 하나만 전달되어야 하는 조건을 검증합니다.
```go
cmd.Flags().BoolVar(&jsonOut, "json", "j", false, "JSON 출력")
cmd.Flags().BoolVar(&yamlOut, "yaml", "y", false, "YAML 출력")
// --json과 --yaml을 동시에 설정 시 Validate 단계에서 에러 발생
cmd.Flags().MarkFlagsMutuallyExclusive("json", "yaml")
```

### 필수 동반 조건 지정 (Required Together)
하나라도 설정되면 그룹 내 모든 플래그가 함께 설정되어야 하는 조건입니다.
```go
cmd.Flags().StringVar(&user, "user", "u", "", "계정")
cmd.Flags().StringVar(&pass, "password", "p", "", "비밀번호")
// --user 또는 --password 중 하나라도 주어지면 둘 다 필수로 동작
cmd.Flags().MarkFlagsRequiredTogether("user", "password")
```

### 설정 파일 매핑 (BindConfig)
외부 구성 파일(JSON, INI)의 값을 플래그에 매핑하여 자동으로 읽어올 수 있습니다. `wcli`는 다음과 같은 우선순위 체인(우선순위 연동 사슬)을 지원합니다:
`입력된 플래그 > 바인딩된 환경변수 > 설정 파일 내 매핑값 > 기본값(Default)`

의존성 오버헤드를 막기 위해 외부 YAML 파서 등을 사용하지 않고, Go 표준 패키지(`encoding/json`)와 한 줄씩 파싱하는 경량 INI 파서를 직접 내장 구현했습니다. 성능 하락을 방지하기 위해 파일 실시간 감지(Hot Reload)를 생략하고 CLI 실행 시점에 1회만 설정 파일을 파싱합니다.

```go
package main

import (
    "fmt"
    "os"
    "github.com/seoyc/wcli"
)

func main() {
    // 1. 설정 파일 설정 및 로드 (JSON 또는 INI 지원)
    wcli.SetConfigFile("config.json")
    wcli.SetConfigType("json")
    if err := wcli.ReadInConfig(); err != nil {
        fmt.Printf("설정 파일 읽기 실패: %v\n", err)
        os.Exit(1)
    }

    var dbHost string
    cmd := &wcli.Command{
        Use: "app",
        Run: func(ctx *wcli.Context) error {
            fmt.Printf("Database Host: %s\n", dbHost)
            return nil
        },
    }

    cmd.Flags().StringVar(&dbHost, "host", "H", "localhost", "데이터베이스 호스트")

    // 2. 환경변수 및 설정파일 키 매핑 (점 표기법 지원)
    _ = cmd.Flags().BindEnv("host", "DB_HOST")
    _ = cmd.Flags().BindConfig("host", "database.host")

    cmd.Execute(os.Args[1:])
}
```

#### INI 형식 매핑
INI 설정 파일인 경우 섹션 이름을 포함한 점 표기법(`section.key`)으로 매핑합니다.
```ini
[database]
host = config-db-host
port = 5432
```
```go
wcli.SetConfigFile("config.ini")
wcli.SetConfigType("ini")
_ = wcli.ReadInConfig()

cmd.Flags().BindConfig("host", "database.host")
```

## 셸 자동 완성 (Shell Autocomplete)

`wcli.NewCompletionCommand`를 사용하여 빌드된 명령어에 대한 Zsh/Bash용 자동 완성 코드를 생성할 수 있습니다.

```go
func main() {
    rootCmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
    
    // completion 서브커맨드 등록
    rootCmd.AddCommand(wcli.NewCompletionCommand(rootCmd))
    
    rootCmd.Execute(os.Args[1:])
}
```

## 커스텀 도움말 템플릿 (Template Help)

`text/template` 형식을 사용해 도움말 레이아웃을 마음대로 변경할 수 있습니다. 템플릿 컴파일 결과는 패키지 단에서 캐싱되므로 매 출력 시 성능 저하가 없습니다.

```go
const myHelpTemplate = `[bold][cyan]⚡ {{.Name}} 도움말[/cyan][/bold]
{{.Short}}

사용법:
  {{.UsageLine}}
`

cmd := &wcli.Command{
    Use:          "app",
    Short:        "설명",
    HelpTemplate: myHelpTemplate, // 템플릿 주입
}
```
