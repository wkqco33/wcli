# wcli

`wcli`는 Go 언어로 작성된 가볍고 성능 중심적인 CLI 명령어 라이브러리입니다.
Cobra에서 영감을 받아 제작되었으며, 복잡성을 줄이고 직관적인 데이터 구조와 함수형 접근 방식을 채택했습니다.
추가적으로 Python의 `rich` 라이브러리에서 영감을 받은 **스타일링 텍스트 출력**을 지원합니다.

## 특징

- **성능 최우선**: 무거운 파싱 라이브러리 없이 단순하고 효율적인 구조
- **데이터 중심**: 직관적인 `Command` 구조체를 통한 명령어 트리 구성 및 플래그 바인딩
- **명시적 에러 처리**: 숨겨진 패닉 없이 명확한 에러 반환
- **Rich 텍스트 출력**: 내장된 `rich` 패키지로 마크업 기반의 터미널 컬러 출력 지원

## 시작하기

```go
package main

import (
 "fmt"
 "os"

 "github.com/seoyc/wcli"
 "github.com/seoyc/wcli/rich"
)

func main() {
 var verbose bool
 var name string

 rootCmd := &wcli.Command{
  Use:   "app",
  Short: "앱 설명",
  Run: func(ctx *wcli.Context) error {
   if verbose {
    rich.Println("[blue]상세 모드로 실행합니다.[/blue]")
   }
   rich.Println("[green]안녕하세요, %s님![/green]", name)
   return nil
  },
 }

 // 플래그 설정
 rootCmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력 활성화")
 rootCmd.Flags().StringVar(&name, "name", "n", "User", "사용자 이름")

 if err := rootCmd.Execute(os.Args[1:]); err != nil {
  rich.Println("[red]Error: %v[/red]", err)
  os.Exit(1)
 }
}
```

## 마크업 문법 (Rich)

`rich` 패키지는 간단한 대괄호(`[]`) 마크업을 사용해 텍스트 색상 및 스타일을 적용합니다.

- `[red]텍스트[/red]` : 빨간색 텍스트
- `[bold][green]텍스트[/green][/bold]` : 굵은 녹색 텍스트
- `\[red]` : 대괄호 문자 그대로 출력 (이스케이프)
