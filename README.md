# wcli

[![Go Reference](https://pkg.go.dev/badge/github.com/wkqco33/wcli.svg)](https://pkg.go.dev/github.com/wkqco33/wcli)
[![Go Report Card](https://goreportcard.com/badge/github.com/wkqco33/wcli)](https://goreportcard.com/report/github.com/wkqco33/wcli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`wcli`는 Go 언어로 작성된 가볍고 성능 중심적인 CLI 명령어 라이브러리입니다.
[Cobra](https://github.com/spf13/cobra)에서 영감을 받아 제작되었으며, Python의 [`rich`](https://github.com/Textualize/rich) 라이브러리처럼 **마크업 기반 터미널 컬러 출력**을 내장합니다.

## ✨ 특징

- **제로 의존성**: 외부 패키지 없이 순수 Go 표준 라이브러리만 사용
- **성능 최우선**: O(1) 커맨드 라우팅 맵, 정렬 캐시, 마크업 파싱 캐시
- **데이터 중심**: 직관적인 `Command` 구조체로 명령어 트리 구성
- **TDD 친화적 하네스**: 입출력 버퍼 캡처 및 제로 디펜던시 단언 헬퍼(`internal/testutil`) 내장
- **완전한 테스트 격리**: 의존성 주입(DI) 기반으로 `t.Parallel()` 병렬 테스트 완벽 지원
- **Rich 텍스트 출력**: 마크업 기반 터미널 컬러/스타일 + Spinner, ProgressBar, Table, Box 내장
- **Cobra 호환 API**: `Persistent` 플래그/훅, 서브커맨드 별칭, `--name=value` 문법 등
- **구조화된 에러 처리**: `FlagError`, `ValidationError`, `ConfigError` 등 정밀한 디버깅 및 세부 속성 추출 지원
- **경량 로깅 서브패키지**: 런타임 오버헤드가 극히 적은 콘솔 로깅 및 동적 로그 레벨 제어 분리 제공
- **Fuzzy 커맨드 매칭**: 오타 입력 시 유사 커맨드 자동 제안
- **셸 자동 완성**: Bash / Zsh / Fish 완성 스크립트 생성

## 📦 설치

### 라이브러리 설치

```bash
go get github.com/wkqco33/wcli
```

### CLI 스캐폴딩 도구 설치 (선택)

새로운 `wcli` 기반 CLI 프로젝트를 신속히 구성할 수 있는 스캐폴딩 도구입니다.

```bash
go install github.com/wkqco33/wcli/cmd/wcli@latest
```

## 🚀 빠른 시작

```go
package main

import (
    "os"

    "github.com/wkqco33/wcli"
    "github.com/wkqco33/wcli/rich"
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

## 📚 상세 문서

긴 README 대신 주제별 문서로 나눠서 관리합니다. 필요한 기능부터 바로 찾아보세요.

- [커맨드 및 플래그 가이드](docs/command-guide.md): 커맨드 구조, 플래그, 출력 제어, 도움말, 에러 처리, 로깅, 자동 완성
- [Rich 출력 가이드](docs/rich-guide.md): 마크업, Spinner, ProgressBar, Interactive Prompt
- [설정 및 고급 플래그 가이드](docs/config-guide.md): 환경 변수, 설정 파일, 플래그 제약 조건, 구조체 바인딩
- [CLI 도구 및 개발 워크플로우](docs/tooling-guide.md): `cmd/wcli` 도구와 `Taskfile.yml` 사용법
- [에이전트 및 TDD 개발 가이드](AGENTS.md): 아키텍처 원칙, TDD 작성법, 테스트 하네스, 릴리스 규칙
- [기여 가이드](CONTRIBUTING.md)
- [변경 이력](CHANGELOG.md)

## 🤝 기여 및 피드백

버그 제보나 기능 제안은 언제든지 환영합니다. 자세한 개발 워크플로우와 테스트 방법은 [기여 가이드 (CONTRIBUTING.md)](CONTRIBUTING.md)를 참고해 주세요.

## 📄 라이선스

이 프로젝트는 [MIT License](LICENSE)에 따라 배포됩니다.
