# wcli 기여 가이드 (Contributing Guide)

`wcli` 프로젝트에 관심을 가져주셔서 감사합니다! 버그 수정, 기능 추가, 문서 개선 등 모든 기여를 환영합니다.

---

## 🛠️ 개발 환경 설정

### 요구 사항
- **Go**: 1.22 이상 (표준 라이브러리 기반, 외부 의존성 없음)
- **Task** (선택 권장): 태스크 러너 ([설치 가이드](https://taskfile.dev/installation/))
- **staticcheck** 또는 **golangci-lint** (코드 린팅용)

### 저장소 복제 및 준비
```bash
git clone https://github.com/wkqco33/wcli.git
cd wcli

# 전체 빌드 및 테스트 확인
task check
# 또는 task가 없는 경우:
go test -race ./...
```

---

## 🧪 개발 워크플로우

`Taskfile.yml`을 통해 편리한 개발 명령어를 제공합니다:

| 명령어 | 설명 |
|---|---|
| `task test` | 전체 패키지 단위 테스트 실행 |
| `task test-race` | 경쟁 상태(Race Condition) 검지기 적용 테스트 |
| `task cover` | 테스트 커버리지 측정 및 HTML 리포트 생성 (`coverage.html`) |
| `task fmt` | `gofmt` 코드 포맷팅 일괄 적용 |
| `task fmt-check` | 포맷팅 준수 여부 검사 |
| `task vet` | `go vet` 정적 분석 실행 |
| `task lint` | `staticcheck` / `golangci-lint` 린터 실행 |
| `task check` | 포맷 확인 + vet + lint + test 일괄 실행 (PR 전 필수 실행 권장) |
| `task clean` | 빌드 산출물 및 커버리지 임시 파일 정리 |

---

## 📐 기여 원칙

1. **제로 의존성 (Zero Dependencies) 유지**:
   - `wcli`는 외부 패키지(`third-party`) 의존성 없이 **Go 표준 라이브러리만 사용**하는 것을 핵심 철학으로 합니다. 새로운 기능을 추가할 때 외부 라이브러리 추가는 지양해 주세요.
2. **테스트 동반**:
   - 버그 수정이나 새 기능 추가 시 반드시 관련 유닛 테스트를 함께 작성해 주세요.
   - `go test -race ./...`를 통과해야 합니다.
3. **코드 포맷팅**:
   - 제출 전 `task fmt` 또는 `gofmt -w .`를 실행하여 Go 표준 포맷을 맞춰주세요.

---

## 🚀 Pull Request 절차

1. 저장소를 Fork하고 기능 브랜치를 생성합니다. (`git checkout -b feature/awesome-feature` 또는 `fix/issue-description`)
2. 변경 사항을 구현하고 테스트를 추가합니다.
3. `task check` 명령어로 포맷, 린트, 테스트를 확인합니다.
4. 커밋 메시지는 간결하고 명확하게 작성합니다. (예: `feat: add ...`, `fix: resolve ...`)
5. Pull Request를 생성하고 변경 배경과 테스트 결과를 설명합니다.
