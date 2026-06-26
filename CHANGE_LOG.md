# CHANGE LOG

## [Unreleased]

### 추가 (Added)
- **다양한 스피너(Spinner) 스타일 프리셋 추가**
  - 브라이유 점자 외에 `SpinnerDots`, `SpinnerLine`, `SpinnerCircle`, `SpinnerArrow`, `SpinnerBouncing` 프리셋 추가.
  - `SpinnerStyle` 구조체를 노출하여 사용자 정의 프레임/주기를 적용 가능하도록 함.
  - 스피너 런타임에 스타일을 변경할 수 있도록 `SetStyle(style SpinnerStyle)` API 추가.
- **프로그레스바(ProgressBar) 기능 강화 및 테마 프리셋 추가**
  - 블록 외에 `ThemeLine`, `ThemeDoubleLine`, `ThemeBullet`, `ThemeArrow`, `ThemeStar` 등 6가지 진행률 테마 프리셋 추가.
  - 진행 표시줄 채움색(`FillColor`) 및 배경색(`EmptyColor`) 커스텀 마크업 지정 가능.
  - 퍼센트 표시 생략(`ShowPercent`) 및 전체 단계 대비 현재 진행 카운터(`ShowCounter`) 표시 옵션 추가.
  - 진행률 기준 남은 예상 완료 시간을 추정하는 ETA(`ShowETA`, `Start()`) 기능 추가.
- **시연용 데모 예제 추가**
  - `examples/rich_demo/main.go`를 신설하여 새로운 스피너 및 프로그레스바 테마 전체 시연 가능하도록 구성.
- **전각 문자 표시 폭 계산 유틸 추가**
  - `DisplayWidth(s string) int`를 추가하여 한글/한자/가나/전각기호 및 주요 이모지를 2칸으로 계산.

### 수정 (Fixed)
- **Box/Table 정렬 오류 수정**
  - `Box`, `Table`이 폭을 룬 개수(`utf8.RuneCountInString`)로 계산해 한글/이모지 등 전각 문자가 섞이면 오른쪽 테두리가 어긋나던 문제를 `DisplayWidth` 적용으로 해결.
