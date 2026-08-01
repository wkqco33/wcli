# Rich 출력 가이드

[← README로 돌아가기](../README.md)

`rich` 패키지의 마크업 기반 출력과 터미널 UI 컴포넌트를 정리한 문서입니다.

## Rich 마크업

`rich` 패키지는 `[태그]텍스트[/태그]` 형식의 마크업을 ANSI 코드로 변환합니다.

### 지원 태그

| 태그 | 효과 |
|---|---|
| `[bold]` | 굵게 |
| `[dim]` | 흐리게 |
| `[underline]` | 밑줄 |
| `[italic]` | 기울임 |
| `[blink]` | 깜빡임 |
| `[reverse]` | 반전 |
| `[strikethrough]` | 취소선 |
| `[red]` `[green]` `[yellow]` | 텍스트 색상 |
| `[blue]` `[magenta]` `[cyan]` `[white]` | 텍스트 색상 |
| `[bg_red]` `[bg_green]` 등 | 배경색 |

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
## Spinner

비동기 작업 진행 중 터미널에 회전 애니메이션을 출력합니다. 비터미널 환경(파이프/리다이렉션)에서는 정적 텍스트로 자동 전환됩니다.

```go
s := rich.NewSpinner(os.Stderr) // nil이면 os.Stderr 사용

// 다양한 스타일 지정 가능 (기본값: SpinnerDefault)
// 지원 프리셋: SpinnerDefault, SpinnerDots, SpinnerLine, SpinnerCircle, SpinnerArrow, SpinnerBouncing
s.SetStyle(rich.SpinnerDots)

s.Start("데이터 로딩 중...")
// ... 비동기 작업 ...
s.UpdateText("거의 완료...")     // 실행 중 텍스트 변경 (thread-safe)
s.Stop("[green]✓ 완료[/green]") // 멈추고 완료 메시지 출력
```
## ProgressBar

진행 상황을 진행 표시줄(Progress Bar)로 렌더링합니다. 다양한 빌트인 테마, 마크업 색상, 카운터 및 예상 완료 시간(ETA) 표시 옵션을 지원합니다.

```go
pb := rich.NewProgressBar(100) // 전체 단계 100 지정

// 1. 다양한 스타일 테마 지원 (기본값: ThemeBlock)
// 지원 프리셋: ThemeBlock(█/░), ThemeLine(━/─), ThemeDoubleLine(═/─), ThemeBullet(●/○), ThemeArrow(>/ ), ThemeStar(★/☆)
pb.SetTheme(rich.ThemeDoubleLine)

// 2. 색상 마크업 지정
pb.FillColor = "blue"
pb.EmptyColor = "dim"

// 3. 다양한 옵션 지원
pb.ShowCounter = true // "(30/100)" 카운터 표시 활성화
pb.ShowETA = true     // "ETA: 5s" 예상 남은 시간 표시 활성화

// 4. 타이머 시작 (ETA 계산에 필요)
pb.Start()

for i := 0; i <= 100; i += 20 {
    pb.Print(i) // os.Stdout에 렌더링 결과 출력
    time.Sleep(100 * time.Millisecond)
}
```
## Interactive Prompts

`rich` 패키지의 대화형 입력 함수들입니다.

```go
// 텍스트 입력 (빈 입력 시 기본값 반환)
name, err := rich.Prompt("이름을 입력하세요", "홍길동")

// Y/N 확인
ok, err := rich.Confirm("계속하시겠습니까?", true) // 기본값: Y

// 번호 선택 (잘못된 입력 시 최대 3회 재시도)
env, err := rich.Select("환경 선택", []string{"dev", "staging", "prod"})
```

테스트 시에는 `io.Reader`/`io.Writer`를 직접 주입하는 내부 함수를 사용합니다:

```go
result, err := rich.FPrompt(&buf, strings.NewReader("입력값\n"), "레이블", "기본값")
```
