package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffolderInit(t *testing.T) {
	// 임시 테스트 디렉토리 생성
	tmpDir, err := ioutil.TempDir("", "wcli-test-*")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 현재 작업 디렉토리 백업 및 변경
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("현재 디렉토리 획득 실패: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("작업 디렉토리 변경 실패: %v", err)
	}
	defer os.Chdir(origWd)

	// wcli init testapp 실행 시뮬레이션
	data := initData{
		ModuleName:  "testapp",
		LibraryPath: "/mock/wcli",
		AppName:     "testapp",
	}

	if err := renderToFile("go.mod", GoModTemplate, data); err != nil {
		t.Fatalf("go.mod 생성 실패: %v", err)
	}
	if err := renderToFile("main.go", MainTemplate, data); err != nil {
		t.Fatalf("main.go 생성 실패: %v", err)
	}
	if err := renderToFile("Makefile", MakefileTemplate, data); err != nil {
		t.Fatalf("Makefile 생성 실패: %v", err)
	}

	// 생성 파일 존재 여부 및 키워드 검증
	files := []string{"go.mod", "main.go", "Makefile"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("파일이 생성되지 않음: %s", file)
		}
	}

	goModContent, _ := ioutil.ReadFile("go.mod")
	if !strings.Contains(string(goModContent), "module testapp") {
		t.Errorf("go.mod에 module testapp이 없음, 실제: %q", string(goModContent))
	}

	mainContent, _ := ioutil.ReadFile("main.go")
	if !strings.Contains(string(mainContent), "// wcli:commands") {
		t.Errorf("main.go에 wcli:commands 마커가 없음, 실제: %q", string(mainContent))
	}
}

func TestScaffolderAdd(t *testing.T) {
	// 임시 테스트 디렉토리 생성
	tmpDir, err := ioutil.TempDir("", "wcli-test-*")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("현재 디렉토리 획득 실패: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("작업 디렉토리 변경 실패: %v", err)
	}
	defer os.Chdir(origWd)

	// 초기 main.go 모형 생성
	mainBoilerplate := `package main
func main() {
	// wcli:commands
}
`
	if err := ioutil.WriteFile("main.go", []byte(mainBoilerplate), 0644); err != nil {
		t.Fatalf("임시 main.go 기록 실패: %v", err)
	}

	// wcli add create 시뮬레이션
	data := cmdData{
		CmdName:       "create",
		CmdStructName: "CreateCmd",
	}

	if err := renderToFile("create.go", CommandTemplate, data); err != nil {
		t.Fatalf("create.go 생성 실패: %v", err)
	}

	if err := injectCommandToMain("CreateCmd"); err != nil {
		t.Fatalf("명령어 자동 등록 실패: %v", err)
	}

	// 검증
	createContent, _ := ioutil.ReadFile("create.go")
	if !strings.Contains(string(createContent), "var CreateCmd = &wcli.Command") {
		t.Errorf("create.go 내용 비정상, 실제: %q", string(createContent))
	}

	mainContent, _ := ioutil.ReadFile("main.go")
	if !strings.Contains(string(mainContent), "rootCmd.AddCommand(CreateCmd)") {
		t.Errorf("main.go 자동 바인딩 결과 비정상, 실제: %q", string(mainContent))
	}
}
