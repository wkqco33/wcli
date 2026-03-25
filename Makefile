MODULE  := github.com/seoyc/wcli
GOBIN   ?= $(shell go env GOPATH)/bin

# 기본 타겟
.DEFAULT_GOAL := help

##@ 개발

.PHONY: build
build: ## 패키지 빌드 (컴파일 오류 확인)
	go build ./...

.PHONY: test
test: ## 전체 테스트 실행
	go test ./...

.PHONY: test-v
test-v: ## 전체 테스트 실행 (상세 출력)
	go test -v ./...

.PHONY: bench
bench: ## 벤치마크 실행
	go test -bench=. -benchmem ./...

.PHONY: cover
cover: ## 테스트 커버리지 측정 및 HTML 리포트 생성
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "커버리지 리포트: coverage.html"

.PHONY: cover-pct
cover-pct: ## 테스트 커버리지 퍼센트 출력
	go test -cover ./...

##@ 코드 품질

.PHONY: vet
vet: ## go vet 정적 분석 실행
	go vet ./...

.PHONY: fmt
fmt: ## gofmt 코드 포맷 적용
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## 포맷 규칙 준수 여부 확인 (수정 없음)
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "포맷이 맞지 않는 파일:"; \
		echo "$$out"; \
		exit 1; \
	fi

.PHONY: lint
lint: ## staticcheck 린터 실행 (없으면 설치 안내)
	@command -v staticcheck >/dev/null 2>&1 || { \
		echo "staticcheck가 설치되어 있지 않습니다."; \
		echo "설치: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		exit 1; \
	}
	staticcheck ./...

.PHONY: check
check: fmt-check vet test ## fmt-check + vet + test 한번에 실행

##@ 설치 / 삭제

.PHONY: install
install: ## 현재 모듈을 GOPATH/bin에 설치 (예제 바이너리가 있을 경우)
	@if ls cmd/*/main.go 1>/dev/null 2>&1; then \
		go install ./cmd/...; \
		echo "설치 완료: $(GOBIN)"; \
	else \
		echo "설치할 바이너리가 없습니다 (cmd/ 디렉토리 없음)."; \
		echo "라이브러리로 사용: go get $(MODULE)"; \
	fi

.PHONY: uninstall
uninstall: ## GOPATH/bin에서 설치된 바이너리 제거
	@if ls cmd/*/main.go 1>/dev/null 2>&1; then \
		for dir in cmd/*/; do \
			bin=$(GOBIN)/$$(basename $$dir); \
			if [ -f "$$bin" ]; then rm -f "$$bin" && echo "삭제: $$bin"; fi; \
		done; \
	else \
		echo "제거할 바이너리가 없습니다."; \
	fi

##@ 의존성

.PHONY: tidy
tidy: ## go mod tidy (의존성 정리)
	go mod tidy

.PHONY: deps
deps: ## 의존성 목록 출력
	go list -m all

##@ 정리

.PHONY: clean
clean: ## 빌드 아티팩트 및 커버리지 파일 제거
	go clean ./...
	rm -f coverage.out coverage.html

##@ 도움말

.PHONY: help
help: ## 사용 가능한 타겟 목록 출력
	@awk 'BEGIN {FS = ":.*##"; printf "\n사용법:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
