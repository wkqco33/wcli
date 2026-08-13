# 커맨드 및 플래그 가이드

[← README로 돌아가기](../README.md)

`wcli.Command` 기반 CLI 구조, 플래그, 출력, 도움말, 에러 처리 기능을 정리한 문서입니다.

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

### 서브커맨드 그룹

도움말에서 서브커맨드를 논리적 그룹으로 묶어 표시합니다.

```go
rootCmd.AddCommand(
    &wcli.Command{Use: "deploy",   Short: "배포 실행",   GroupName: "배포 관리"},
    &wcli.Command{Use: "rollback", Short: "배포 롤백",   GroupName: "배포 관리"},
    &wcli.Command{Use: "logs",     Short: "로그 조회",   GroupName: "운영"},
    &wcli.Command{Use: "status",   Short: "상태 확인"},  // 그룹 없음 → Available Commands
)
```

도움말 출력 예:
```
배포 관리:
  deploy     배포 실행
  rollback   배포 롤백

운영:
  logs   로그 조회

Available Commands:
  status   상태 확인
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

### 컨텍스트 주입 (ExecuteContext)

취소·타임아웃·값 전파를 위해 외부 `context.Context`를 주입할 수 있습니다. 주입된 컨텍스트는 `ctx.Context`로 전달됩니다.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

cmd := &wcli.Command{
    Use: "app",
    Run: func(ctx *wcli.Context) error {
        select {
        case <-ctx.Done():
            return ctx.Err()  // 타임아웃/취소
        default:
            return doWork()
        }
    },
}
cmd.ExecuteContext(ctx, os.Args[1:])  // Execute(args)는 context.Background()를 사용
```

### Fuzzy 커맨드 매칭

오타가 있는 커맨드 입력 시 편집 거리 기반으로 유사 커맨드를 제안합니다.

```
$ app deply
Error: unknown command "deply"

Did you mean?
  deploy
```
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

### 도움말 카테고리 분류

도움말이 길어질 때는 플래그를 카테고리별로 나눠 표시할 수 있습니다.

```go
cmd.Flags().StringVar(&host, "host", "H", "localhost", "대상 호스트")
cmd.Flags().StringVar(&token, "token", "t", "", "인증 토큰")

cmd.Flags().SetCategory("host", "연결")
cmd.Flags().SetCategory("token", "인증")
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

# 결합 단축 플래그: bool 단축 플래그를 묶고, 마지막에 값 플래그를 붙일 수 있음
app -abc                 # -a -b -c (모두 bool)
app -vofile.txt          # -v(bool) + -o=file.txt
app -vo file.txt         # -v(bool) + -o file.txt

# -- 종결자: 이후 모든 인자는 플래그로 해석하지 않음
app --name foo -- --not-a-flag positional-arg
```

> 다중 문자 단축키(예: `Shorthand: "vv"`)가 등록돼 있으면 결합 분해보다 우선합니다.
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
    logger := logging.NewDefaultLogger(os.Stderr, logging.LevelInfo, true)
    logging.SetLogger(logger)

    cmd := &wcli.Command{
        Use: "app",
        Run: func(ctx *wcli.Context) error {
            ctx.Logger.Log(logging.LevelInfo, "작업을 시작합니다.")
            return nil
        },
    }
    cmd.Execute(os.Args[1:])
}
```

전역 상태 분리가 필요하면 인스턴스 기반 매니저를 사용할 수 있습니다.

```go
manager := logging.NewLoggerManager()
manager.SetLogger(logging.NewDefaultLogger(os.Stderr, logging.LevelDebug, true))
manager.GetLogger().Log(logging.LevelInfo, "instance logger")
```
## 셸 자동 완성 (Shell Autocomplete)

`wcli.NewCompletionCommand`로 Bash / Zsh / Fish 자동 완성 스크립트를 생성합니다.

```go
rootCmd.AddCommand(wcli.NewCompletionCommand(rootCmd))
```

```bash
# Bash
source <(app completion bash)

# Zsh
source <(app completion zsh)

# Fish
app completion fish > ~/.config/fish/completions/app.fish
```
## 커스텀 도움말 템플릿 (Template Help)

`text/template` 형식을 사용해 도움말 레이아웃을 변경할 수 있습니다. 템플릿 컴파일 결과는 패키지 단에서 캐싱됩니다.

```go
const myHelpTemplate = `[bold][cyan]⚡ {{.Name}} 도움말[/cyan][/bold]
{{.Short}}

사용법:
  {{.UsageLine}}
`

cmd := &wcli.Command{
    Use:          "app",
    Short:        "설명",
    HelpTemplate: myHelpTemplate,
}
```
