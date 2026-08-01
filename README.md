# wcli

`wcli`는 Go 언어로 작성된 가볍고 성능 중심적인 CLI 명령어 라이브러리입니다.
[Cobra](https://github.com/spf13/cobra)에서 영감을 받아 제작되었으며, Python의 [`rich`](https://github.com/Textualize/rich) 라이브러리처럼 **마크업 기반 터미널 컬러 출력**을 내장합니다.

## 특징

- **제로 의존성**: 외부 패키지 없이 순수 Go 표준 라이브러리만 사용
- **성능 최우선**: O(1) 커맨드 라우팅 맵, 정렬 캐시, 마크업 파싱 캐시
- **데이터 중심**: 직관적인 `Command` 구조체로 명령어 트리 구성
- **Rich 텍스트 출력**: 마크업 기반 터미널 컬러/스타일 + Spinner, ProgressBar, Table, Box 내장
- **Cobra 호환 API**: `Persistent` 플래그/훅, 서브커맨드 별칭, `--name=value` 문법 등
- **구조화된 에러 처리**: `FlagError`, `ValidationError` 등 정밀한 디버깅 및 세부 속성 추출 지원
- **경량 로깅 서브패키지**: 런타임 오버헤드가 극히 적은 콘솔 로깅 및 동적 로그 레벨 제어 분리 제공
- **Fuzzy 커맨드 매칭**: 오타 입력 시 유사 커맨드 자동 제안
- **셸 자동 완성**: Bash / Zsh / Fish 완성 스크립트 생성

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

## 문서 안내

긴 README 대신 주제별 문서로 나눠서 관리합니다. 필요한 기능부터 바로 찾아보세요.

- [커맨드 및 플래그 가이드](docs/command-guide.md): 커맨드 구조, 플래그, 출력 제어, 도움말, 에러 처리, 로깅, 자동 완성
- [Rich 출력 가이드](docs/rich-guide.md): 마크업, Spinner, ProgressBar, Interactive Prompt
- [설정 및 고급 플래그 가이드](docs/config-guide.md): 환경 변수, 설정 파일, 플래그 제약 조건, 구조체 바인딩
- [CLI 도구 및 개발 워크플로우](docs/tooling-guide.md): `cmd/wcli` 도구와 `Taskfile.yml` 사용법
- [변경 이력](CHANGE_LOG.md)

## 문서 사용 순서

1. 처음 시작할 때는 이 README의 빠른 시작 예제를 확인합니다.
2. 실제 CLI 구조를 설계할 때는 [커맨드 및 플래그 가이드](docs/command-guide.md)를 봅니다.
3. 터미널 출력 꾸미기가 필요하면 [Rich 출력 가이드](docs/rich-guide.md)를 봅니다.
4. 설정 파일, 환경 변수, 고급 바인딩이 필요하면 [설정 및 고급 플래그 가이드](docs/config-guide.md)를 봅니다.
5. 프로젝트 스캐폴딩이나 저장소 개발 작업은 [CLI 도구 및 개발 워크플로우](docs/tooling-guide.md)를 봅니다.
