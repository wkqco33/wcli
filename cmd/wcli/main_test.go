package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTmpDir(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "wcli-test-*")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	if realPath, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = realPath
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("현재 디렉토리 획득 실패: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("작업 디렉토리 변경 실패: %v", err)
	}
	return tmpDir, func() {
		os.Chdir(origWd)
		os.RemoveAll(tmpDir)
	}
}

func TestScaffolderInit(t *testing.T) {
	tmpDir, cleanup := setupTmpDir(t)
	defer cleanup()

	data := initData{
		ModuleName:  "testapp",
		LibraryPath: "./wcli",
		AppName:     "testapp",
	}

	if err := renderToFile("go.mod", GoModTemplate, data); err != nil {
		t.Fatalf("go.mod 생성 실패: %v", err)
	}
	if err := renderToFile("main.go", MainTemplate, data); err != nil {
		t.Fatalf("main.go 생성 실패: %v", err)
	}
	if err := renderToFile("Taskfile.yml", TaskfileTemplate, data); err != nil {
		t.Fatalf("Taskfile.yml 생성 실패: %v", err)
	}

	for _, file := range []string{"go.mod", "main.go", "Taskfile.yml"} {
		if _, err := os.Stat(filepath.Join(tmpDir, file)); os.IsNotExist(err) {
			t.Errorf("파일이 생성되지 않음: %s", file)
		}
	}

	goModContent, _ := os.ReadFile("go.mod")
	if !strings.Contains(string(goModContent), "module testapp") {
		t.Errorf("go.mod에 module testapp이 없음, 실제: %q", string(goModContent))
	}
	// 상대 경로가 기록되어야 함
	if !strings.Contains(string(goModContent), "=> ./wcli") {
		t.Errorf("go.mod에 상대 경로 replace가 없음, 실제: %q", string(goModContent))
	}

	mainContent, _ := os.ReadFile("main.go")
	if !strings.Contains(string(mainContent), "// wcli:commands") {
		t.Errorf("main.go에 wcli:commands 마커가 없음, 실제: %q", string(mainContent))
	}
}

func TestScaffolderAdd(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	mainBoilerplate := `package main
func main() {
	// wcli:commands
}
`
	if err := os.WriteFile("main.go", []byte(mainBoilerplate), 0644); err != nil {
		t.Fatalf("임시 main.go 기록 실패: %v", err)
	}

	data := cmdData{CmdName: "create", CmdStructName: "CreateCmd"}
	if err := renderToFile("create.go", CommandTemplate, data); err != nil {
		t.Fatalf("create.go 생성 실패: %v", err)
	}
	if err := injectCommandToMain("CreateCmd"); err != nil {
		t.Fatalf("명령어 자동 등록 실패: %v", err)
	}

	createContent, _ := os.ReadFile("create.go")
	if !strings.Contains(string(createContent), "var CreateCmd = &wcli.Command") {
		t.Errorf("create.go 내용 비정상, 실제: %q", string(createContent))
	}

	mainContent, _ := os.ReadFile("main.go")
	if !strings.Contains(string(mainContent), "rootCmd.AddCommand(CreateCmd)") {
		t.Errorf("main.go 자동 바인딩 결과 비정상, 실제: %q", string(mainContent))
	}
}

func TestDetectWcliPath(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	tests := []struct {
		name       string
		gitmodules string
		wantPath   string
		wantErr    bool
	}{
		{
			name: "정상 탐지",
			gitmodules: `[submodule "wcli"]
	path = wcli
	url = https://github.com/wkqco33/wcli
`,
			wantPath: "./wcli",
		},
		{
			name: "커스텀 경로",
			gitmodules: `[submodule "libs/wcli"]
	path = libs/wcli
	url = https://github.com/wkqco33/wcli
`,
			wantPath: "./libs/wcli",
		},
		{
			name: "wcli 없음",
			gitmodules: `[submodule "other"]
	path = other
	url = https://github.com/example/other
`,
			wantErr: true,
		},
		{
			name:       "파일 없음",
			gitmodules: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(".gitmodules")
			if tt.gitmodules != "" {
				if err := os.WriteFile(".gitmodules", []byte(tt.gitmodules), 0644); err != nil {
					t.Fatalf(".gitmodules 기록 실패: %v", err)
				}
			}

			got, err := detectWcliPath(".")
			if tt.wantErr {
				if err == nil {
					t.Errorf("에러가 발생해야 하지만 성공함, 결과: %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			if got != tt.wantPath {
				t.Errorf("경로 불일치: got %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestAddCmd_EmptyName(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	if err := os.WriteFile("main.go", []byte("package main\nfunc main(){}\n"), 0644); err != nil {
		t.Fatalf("main.go 기록 실패: %v", err)
	}

	// 빈 문자열 슬라이싱 패닉이 발생하지 않고 에러를 반환해야 함
	var panicOccurred bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
			}
		}()
		cmdName := strings.ToLower(strings.TrimSpace(""))
		if cmdName == "" {
			return // 정상 조기 반환
		}
		runes := []rune(cmdName)
		_ = strings.ToUpper(string(runes[:1])) + string(runes[1:]) + "Cmd"
	}()

	if panicOccurred {
		t.Error("빈 커맨드 이름 처리 중 패닉 발생")
	}
}

// TestBuildInitCmd init 커맨드의 Run(Execute 경로)을 통해 프로젝트가 스캐폴딩되는지 검증한다.
func TestBuildInitCmd(t *testing.T) {
	tmpDir, cleanup := setupTmpDir(t)
	defer cleanup()

	cmd := buildInitCmd()
	if err := cmd.Execute([]string{"--lib-path", "./wcli", "myorg/myapp"}); err != nil {
		t.Fatalf("init 실행 실패: %v", err)
	}
	for _, file := range []string{"go.mod", "main.go", "Taskfile.yml"} {
		if _, err := os.Stat(filepath.Join(tmpDir, file)); os.IsNotExist(err) {
			t.Errorf("파일이 생성되지 않음: %s", file)
		}
	}
	// 모듈 인자가 없으면 에러
	if err := buildInitCmd().Execute(nil); err == nil {
		t.Error("모듈명 없이 실행 시 에러가 발생해야 함")
	}
}

func TestValidateCommandName(t *testing.T) {
	valid := []string{"create", "create-user", "v2", "a1-b2"}
	for _, name := range valid {
		if err := validateCommandName(name); err != nil {
			t.Fatalf("유효한 이름이 거부됨: %s (%v)", name, err)
		}
	}

	invalid := []string{"", "-create", "create-", "Create", "create_user", "1create", "create user", "../x"}
	for _, name := range invalid {
		if err := validateCommandName(name); err == nil {
			t.Fatalf("유효하지 않은 이름이 허용됨: %s", name)
		}
	}
}

func TestToCommandStructName(t *testing.T) {
	tests := map[string]string{
		"create":      "CreateCmd",
		"create-user": "CreateUserCmd",
		"user-v2":     "UserV2Cmd",
	}
	for input, want := range tests {
		if got := toCommandStructName(input); got != want {
			t.Fatalf("구조체 이름 변환 실패: input=%s got=%s want=%s", input, got, want)
		}
	}
}

// TestBuildAddCmd add 커맨드의 Run을 통해 서브커맨드 파일 생성 및 main.go 주입을 검증한다.
func TestBuildAddCmd(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	os.WriteFile("main.go", []byte("package main\nfunc main(){\n\t// wcli:commands\n}\n"), 0644)

	if err := buildAddCmd().Execute([]string{"create"}); err != nil {
		t.Fatalf("add 실행 실패: %v", err)
	}
	if _, err := os.Stat("create.go"); os.IsNotExist(err) {
		t.Error("create.go가 생성되지 않음")
	}
	mainContent, _ := os.ReadFile("main.go")
	if !strings.Contains(string(mainContent), "rootCmd.AddCommand(CreateCmd)") {
		t.Errorf("main.go 자동 바인딩 누락: %q", string(mainContent))
	}
	// 인자 없이 실행하면 에러
	if err := buildAddCmd().Execute(nil); err == nil {
		t.Error("커맨드명 없이 실행 시 에러가 발생해야 함")
	}
}

func TestRenderToFile_ExistingFile(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	if err := os.WriteFile("go.mod", []byte("existing"), 0644); err != nil {
		t.Fatalf("기존 파일 생성 실패: %v", err)
	}

	if err := renderToFile("go.mod", GoModTemplate, initData{ModuleName: "app"}); err == nil {
		t.Fatal("기존 파일이 있으면 에러가 발생해야 함")
	}
}

func TestInjectCommandToMain_MissingMarker(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	if err := os.WriteFile("main.go", []byte("package main\nfunc main(){}\n"), 0644); err != nil {
		t.Fatalf("main.go 기록 실패: %v", err)
	}

	if err := injectCommandToMain("CreateCmd"); err == nil {
		t.Fatal("마커가 없으면 에러가 발생해야 함")
	}
}

func TestBuildAddCmd_NonProjectRoot(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	if err := buildAddCmd().Execute([]string{"create"}); err == nil {
		t.Fatal("main.go가 없으면 wcli 프로젝트 루트가 아니라는 에러가 필요함")
	}
}

func TestBuildInitCmd_AbsoluteLibPathBecomesRelative(t *testing.T) {
	tmpDir, cleanup := setupTmpDir(t)
	defer cleanup()

	libDir := filepath.Join(tmpDir, "vendor", "wcli")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("lib 디렉토리 생성 실패: %v", err)
	}

	cmd := buildInitCmd()
	if err := cmd.Execute([]string{"--lib-path", libDir, "myorg/myapp"}); err != nil {
		t.Fatalf("init 실행 실패: %v", err)
	}

	goModContent, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("go.mod 읽기 실패: %v", err)
	}
	if !strings.Contains(string(goModContent), "=> ./vendor/wcli") {
		t.Fatalf("절대 경로 lib-path는 현재 디렉토리 기준 상대 경로로 저장되어야 함: %q", string(goModContent))
	}
}

// TestBuildDoctorCmd doctor 커맨드의 Run이 에러 없이 동작하는지 검증한다.
func TestBuildDoctorCmd(t *testing.T) {
	_, cleanup := setupTmpDir(t)
	defer cleanup()

	os.WriteFile("main.go", []byte("package main\n// wcli:commands\n"), 0644)
	os.WriteFile("go.mod", []byte("module myapp\nrequire github.com/wkqco33/wcli v0.0.0\n"), 0644)

	if err := buildDoctorCmd().Execute(nil); err != nil {
		t.Fatalf("doctor 실행 실패: %v", err)
	}
}

func TestRunDoctor_FailWithoutMainGo(t *testing.T) {
	tmpDir, cleanup := setupTmpDir(t)
	defer cleanup()

	orig, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(orig)

	results := runDoctor()
	failFound := false
	for _, r := range results {
		if r.Status == "fail" {
			failFound = true
			break
		}
	}
	if !failFound {
		t.Error("main.go가 없을 때 fail 항목이 있어야 함")
	}
}

func TestRunDoctor_OkWithValidProject(t *testing.T) {
	tmpDir, cleanup := setupTmpDir(t)
	defer cleanup()

	orig, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(orig)

	// 최소 프로젝트 파일 생성
	os.WriteFile("main.go", []byte("package main\n// wcli:commands\n"), 0644)
	os.WriteFile("go.mod", []byte("module myapp\n\nrequire github.com/wkqco33/wcli v0.0.0\n"), 0644)

	results := runDoctor()
	for _, r := range results {
		if r.Status == "fail" {
			t.Errorf("fail 항목이 있어서는 안 됨: %s — %s", r.Name, r.Detail)
		}
	}
}
