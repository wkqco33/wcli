# Changelog

모든 주요 변경 사항은 이 파일에 기록됩니다.
이 프로젝트는 [Semantic Versioning](https://semver.org/spec/v2.0.0.html) 및 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/) 형식을 준수합니다.

## [v0.2.0] - 2026-08-18

### 추가 (Added)
- **TDD 경량 테스트 하네스 (`internal/testutil`)**
  - 표준 출력/에러 버퍼 자동 캡처 실행 헬퍼 (`ExecuteCommand`)
  - Go 표준 `t.Helper()` 기반 제로 디펜던시 단언 헬퍼 (`AssertEqual`, `AssertContains`, `AssertErrorIs`, `AssertErrorAs`, `AssertTrue`, `AssertLen` 등)
- **인스턴스 기반 의존성 주입(DI) 및 병렬 테스트(`t.Parallel()`) 지원**
  - `FlagSet` 환경변수 조회(`SetLookupEnv`) 및 설정 조회(`SetConfigGetter`) 주입 지원
  - `config.Store` 환경변수/파일 읽기/스탯 함수 주입 지원
  - `rich` 패키지 테스트용 터미널/비밀번호 판독 훅 지원
- **에러 스키마 확장**
  - 설정 로드/파싱 오류 표현을 위한 `ConfigError` 구조체 추가
- **TDD 워크플로우 및 가이드**
  - `Taskfile.yml`에 TDD 태스크 (`test:fast`, `test:parallel`, `test:watch`) 추가
  - `AGENTS.md` 작성 (개발 철학, TDD 표준 작성법, 검증 태스크, 릴리스 규칙 단일화)

### 변경 (Changed)
- `config` 패키지의 패키지 레벨 함수들을 `Store` 인스턴스 위임으로 일원화하여 코드 중복 제거 및 유지보수성 향상

## [v0.1.0] - 2026-08-15

### 추가 (Added)
- **Cobra 스타일의 커맨드 및 플래그 라우팅 엔진**
  - `Command` 구조체 기반 계층적 CLI 트리 구성
  - Persistent 플래그/훅 (`PersistentPreRun`, `PersistentPostRun`) 지원
  - 플래그 상호 배제(`MarkFlagsMutuallyExclusive`) 및 필수 동반 지정(`MarkFlagsRequiredTogether`) 제약 지원
  - 플래그 자동 환경변수 매핑(`BindEnv`) 및 커스텀 유효성 검사(`SetValidation`) 지원
- **터미널 Rich 마크업 및 UI 컴포넌트**
  - 마크업 기반 터미널 컬러/스타일링 (`[bold]`, `[cyan]`, `[green]` 등 ANSI 변환)
  - 6종의 스피너 스타일 프리셋 (`SpinnerDots`, `SpinnerLine`, `SpinnerCircle`, `SpinnerArrow`, `SpinnerBouncing`, `SpinnerDefault`)
  - 6종의 프로그레스바 테마 프리셋 (`ThemeLine`, `ThemeDoubleLine`, `ThemeBullet`, `ThemeArrow`, `ThemeStar`, 기본 블록) 및 ETA 예측
  - 전각 문자(한글/한자/이모지 등) 폭 계산 유틸(`DisplayWidth`)을 통한 정밀 정렬 `Box`, `Table`
  - 대화형 프롬프트 (`Prompt`, `PasswordPrompt`, `Confirm`, `Select`)
- **설정 관리 및 바인딩 (`config`)**
  - JSON, YAML, TOML, ENV 포맷 자동 감지 및 로드
  - 구조체 태그(`config:"..."`) 기반 자동 매핑
  - 단일 키-값 저장소 (`Store`)
- **경량 로깅 서브패키지 (`logging`)**
  - 런타임 오버헤드가 적은 레벨별 로깅 (DEBUG, INFO, WARN, ERROR, FATAL)
  - Rich 마크업 포맷팅 지원
- **CLI 스캐폴딩 도구 (`cmd/wcli`)**
  - 새 프로젝트 초기화 (`wcli init <module>`)
  - 서브커맨드 코드 템플릿 생성 (`wcli add <command>`)
  - 프로젝트 구성 점검 (`wcli doctor`)
- **개발 워크플로우 도구**
  - `Taskfile.yml`을 통한 빌드, 테스트, 커버리지, 포맷, 린트, 벤치마크 일괄 관리

### 수정 (Fixed)
- **커스텀 도움말 템플릿 플래그 필드 누락 해결**
  - 커스텀 템플릿에서 `{{.Name}}`, `{{.Shorthand}}`, `{{.TypeStr}}`, `{{.Required}}`를 직접 참조할 수 있도록 `flagHelpData` 필드 확장
- **Box/Table 전각 문자 정렬 오류 수정**
  - 전각 문자(한글, 이모지) 포함 시 오른쪽 테두리가 틀어지던 현상을 `DisplayWidth` 적용으로 교정
