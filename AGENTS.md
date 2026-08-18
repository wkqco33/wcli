# AGENTS.md

이 문서는 `wcli` 프로젝트에서 작업하는 모든 AI 에이전트 및 개발자를 위한 아키텍처, 개발 철학, TDD 개발 규칙 가이드입니다.

---

## 1. 핵심 개발 철학

- **탈객체지향 & 데이터 중심 설계**: 복잡한 클래스/인터페이스 상속 계층 대신 명확한 구조체(스키마)와 순수 함수형 I/O 파이프라인으로 구현합니다.
- **성능 최우선 (Performance First)**: 불필요한 메모리 할당(0 allocs/op 지향)과 불필요한 추상화 레이어를 철저히 배제합니다.
- **Zero-Dependency**: 외부 서드파티 라이브러리(testify, cobra, viper 등)를 일절 추가하지 않고, Go 표준 라이브러리만으로 완성도를 유지합니다.
- **철저한 TDD (Test-Driven Development)**: 모든 기능 추가 및 버그 수정은 **실패하는 단위 테스트(Red) → 최소 구현(Green) → 최적화/정돈(Refactor)** 절차를 따릅니다.
- **일관된 문서화**: 모든 문서와 변경 사항 기록(`CHANGELOG.md`, `README.md`)은 명확하고 간결한 한국어로 작성합니다.

---

## 2. 패키지 구조 및 역할

```
wcli/
├── command.go, flag.go, error.go, help.go, completion.go  # 핵심 CLI 라우팅 & 플래그 엔진
├── config/                                                # 설정 로더 (JSON, YAML, TOML, INI, ENV)
│   ├── config.go                                          # 전역 설정 래퍼
│   ├── store.go                                           # 격리 가능한 인스턴스 기반 Store
│   └── bind.go                                            # 구조체 태그 바인딩
├── logging/                                               # 경량 콘솔 로깅 & 마크업 지원
├── rich/                                                  # 터미널 스타일링, 폭 계산, 프롬프트, 테이블, 박스, 프로그레스바
├── internal/testutil/                                     # TDD 전용 경량 테스트 하네스 및 단언 헬퍼
├── cmd/wcli/                                              # wcli CLI 스캐폴딩 도구 (init, add, doctor)
└── Taskfile.yml                                           # 빌드, 테스트, 린트, 벤치마크 자동화 태스크
```

---

## 3. TDD 개발 가이드라인

### (1) 단위 테스트 작성 규칙
1. **테스트 헬퍼 활용**: 외부 라이브러리 대신 `internal/testutil`의 헬퍼를 사용합니다.
   - `testutil.ExecuteCommand(cmd, args...)`: 커맨드 실행 및 `stdout`, `stderr`, `err` 캡처
   - `testutil.AssertEqual(t, got, want)` / `AssertEqualf(t, got, want, format, args...)`
   - `testutil.AssertContains(t, s, substr)` / `AssertNotContains(t, s, substr)`
   - `testutil.AssertNoError(t, err)` / `AssertError(t, err)`
   - `testutil.AssertErrorIs(t, gotErr, targetErr)` / `AssertErrorAs(t, gotErr, target)`
   - `testutil.AssertTrue(t, cond)` / `AssertFalse(t, cond)`
   - `testutil.AssertLen(t, slice, length)`
   - `testutil.AssertPanics(t, fn)` / `AssertNotPanics(t, fn)`

2. **전역 상태 조작 금지 & 의존성 주입(DI)**:
   - `os.Setenv`나 전역 `config`를 직접 조작하지 말고, `FlagSet.SetLookupEnv`, `FlagSet.SetConfigGetter`, `config.NewStore()`를 활용하여 인스턴스 단위로 격리합니다.
   - 모든 단위 테스트에는 반드시 `t.Parallel()`을 명시하여 독립성과 동시성을 보장합니다.

3. **Table-Driven Test 구조 준수**:
   ```go
   func TestFeature_TableDriven(t *testing.T) {
       t.Parallel()

       tests := []struct {
           name       string
           args       []string
           wantStdout string
           wantErr    bool
       }{
           {
               name:       "성공 케이스",
               args:       []string{"--name", "Alice"},
               wantStdout: "Hello, Alice!",
               wantErr:    false,
           },
           {
               name:    "에러 케이스",
               args:    []string{},
               wantErr: true,
           },
       }

       for _, tt := range tests {
           tt := tt
           t.Run(tt.name, func(t *testing.T) {
               t.Parallel()

               cmd := newTestCommand()
               stdout, _, err := testutil.ExecuteCommand(cmd, tt.args...)

               if tt.wantErr {
                   testutil.AssertError(t, err)
               } else {
                   testutil.AssertNoError(t, err)
                   testutil.AssertContains(t, stdout, tt.wantStdout)
               }
           })
       }
   }
   ```

4. **에러 타입 정형화**:
   - `FlagError`, `ValidationError`, `CommandError`, `ConfigError` 등 구조체 기반 에러를 사용하고 `errors.Is`/`errors.As` 검증을 수행합니다.

---

## 4. 필수 검증 명령어

작업 완료 전 다음 명령어가 모두 통과하는지 확인해야 합니다:

```bash
# 1. 포맷, 정적분석, 린트, 기본 테스트 일괄 검사
task check

# 2. 캐시 없는 초고속 단위 테스트 실행 (TDD 피드백)
task test:fast # 또는 task tf

# 3. 8스레드 병렬 실행 및 레이스 컨디션 정밀 검증
task test:parallel # 또는 task tp

# 4. 성능 벤치마크 (0 allocs/op 및 회귀 여부 점검)
task bench
```

---

## 5. 릴리스 및 버전 태깅 규칙

- **시맨틱 버저닝(SemVer) 준수**: `vMAJOR.MINOR.PATCH` 형식(예: `v0.2.0`)을 따릅니다.
- **Annotated Tag 필수 사용**: 새 버전을 릴리스할 때는 경량(lightweight) 태그 대신 **반드시 어노테이티드 태그(`git tag -a`)**를 생성하여 서명/메타데이터와 릴리스 요약 메시지를 포함해야 합니다.
- **태그 생성 명령어 예시**:
  ```bash
  # 1. CHANGELOG.md 버전 및 날짜 업데이트 완료 후 커밋
  git commit -am "chore(release): bump version to v0.2.0"

  # 2. Annotated tag 생성 (릴리스 요약 메시지 포함)
  git tag -a v0.2.0 -m "Release v0.2.0: TDD 개발 체계 개선 및 테스트 하네스 구축"

  # 3. 원격 푸시
  git push origin main
  git push origin v0.2.0
  ```

---

## 6. 작업 완료 시 체크리스트

- [ ] 신규 기능/수정 사항에 대한 단위 테스트가 먼저 작성되었는가?
- [ ] `internal/testutil`을 사용하여 간결하고 명확하게 단언했는가?
- [ ] `t.Parallel()` 병렬 테스트 시 레이스 컨디션(`task test:parallel`)이 없는가?
- [ ] `task check` (gofmt, go vet, staticcheck, go test)를 100% 통과했는가?
- [ ] 불필요한 외부 의존성이 추가되지 않았는가?
- [ ] `CHANGELOG.md`의 `[Unreleased]` 섹션에 변경 내역을 한국어로 간결하게 기록했는가?
- [ ] 릴리스 시 경량 태그 대신 `git tag -a` (Annotated Tag)를 사용했는가?

