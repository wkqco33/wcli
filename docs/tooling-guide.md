# CLI 도구 및 개발 워크플로우

[← README로 돌아가기](../README.md)

`cmd/wcli` 도구와 이 저장소의 개발용 Taskfile 명령을 정리한 문서입니다.

## wcli CLI 도구 (`cmd/wcli`)

`wcli` 바이너리는 wcli 기반 프로젝트의 스캐폴딩 및 상태 점검 도구입니다.

### 프로젝트 초기화

```bash
# 새 wcli 프로젝트 생성 (go.mod, main.go, Makefile 자동 생성)
wcli init github.com/myorg/myapp

# wcli 라이브러리 경로 직접 지정
wcli init --lib-path ./wcli github.com/myorg/myapp
```

라이브러리 경로를 지정하지 않으면 `.gitmodules`에서 wcli 서브모듈 경로를 자동으로 탐지합니다.

### 서브커맨드 추가

```bash
# create.go 파일 생성 + main.go에 rootCmd.AddCommand(CreateCmd) 자동 주입
wcli add create
```

`main.go`에 `// wcli:commands` 마커가 있어야 자동 등록이 됩니다.

### 프로젝트 상태 점검 (doctor)

```bash
wcli doctor
```

현재 디렉토리의 wcli 프로젝트 상태를 점검합니다:

| 점검 항목 | 설명 |
|---|---|
| main.go 존재 | wcli 프로젝트 루트인지 확인 |
| go.mod 존재 | Go 모듈 파일 유무 |
| wcli 의존성 | go.mod에 seoyc/wcli 의존성 포함 여부 |
| replace 경로 유효성 | go.mod replace 지시어의 경로 존재 여부 |
| wcli:commands 마커 | main.go에 자동 등록 마커 포함 여부 |
## Taskfile 타겟

이 저장소의 개발 워크플로우는 [Task](https://taskfile.dev)(`Taskfile.yml`)로 관리됩니다. `task --list`로 전체 목록과 설명을 확인할 수 있습니다.

```
task build       # 컴파일 오류 확인
task test        # 전체 테스트
task test-v      # 상세 테스트 출력
task bench       # 벤치마크
task cover       # 커버리지 HTML 리포트
task check       # fmt-check + vet + test
task fmt         # 코드 포맷 적용
task vet         # 정적 분석
task tidy        # go mod tidy
task clean       # 빌드 아티팩트 제거
```
