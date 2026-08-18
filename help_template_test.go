package wcli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
)

func TestCustomHelpTemplate(t *testing.T) {
	var buf bytes.Buffer
	cmd := &wcli.Command{
		Use:          "testapp",
		Short:        "짧은 설명",
		Long:         "긴 설명 내용",
		Version:      "1.2.3",
		OutWriter:    &buf,
		HelpTemplate: `Name: {{.Name}} / Version: {{.Version}}`,
		Run:          func(ctx *wcli.Context) error { return nil },
	}

	err := cmd.Execute([]string{"--help"})
	if err != nil {
		t.Fatalf("--help 실행 에러: %v", err)
	}

	output := buf.String()
	expected := "Name: testapp / Version: 1.2.3"
	if !strings.Contains(output, expected) {
		t.Errorf("예상되는 템플릿 결과 %q가 포함되지 않음, 실제 출력: %q", expected, output)
	}
}

func TestDefaultHelpTemplateRender(t *testing.T) {
	var buf bytes.Buffer
	cmd := &wcli.Command{
		Use:       "testapp",
		Short:     "이것은 기본 앱",
		OutWriter: &buf,
		Run:       func(ctx *wcli.Context) error { return nil },
	}
	var verbose bool
	cmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")

	err := cmd.Execute([]string{"--help"})
	if err != nil {
		t.Fatalf("--help 실행 에러: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Usage:") || !strings.Contains(output, "testapp [flags]") {
		t.Errorf("기본 도움말 출력 불량, 실제 출력: %q", output)
	}
}

func TestSubcommandGroups(t *testing.T) {
	root := &wcli.Command{Use: "app", Short: "테스트 앱"}
	root.AddCommand(&wcli.Command{Use: "deploy", Short: "배포", GroupName: "배포 관리"})
	root.AddCommand(&wcli.Command{Use: "rollback", Short: "롤백", GroupName: "배포 관리"})
	root.AddCommand(&wcli.Command{Use: "logs", Short: "로그 조회", GroupName: "운영"})
	root.AddCommand(&wcli.Command{Use: "status", Short: "상태 확인"}) // 그룹 없음

	var buf bytes.Buffer
	root.OutWriter = &buf
	root.Help()
	out := buf.String()

	// 그룹 헤더 출력
	if !strings.Contains(out, "배포 관리:") {
		t.Error("'배포 관리:' 그룹 헤더가 출력에 없음")
	}
	if !strings.Contains(out, "운영:") {
		t.Error("'운영:' 그룹 헤더가 출력에 없음")
	}

	// 그룹 내 커맨드 출력
	if !strings.Contains(out, "deploy") {
		t.Error("'deploy' 커맨드가 출력에 없음")
	}
	if !strings.Contains(out, "rollback") {
		t.Error("'rollback' 커맨드가 출력에 없음")
	}

	// 그룹 없는 커맨드는 Available Commands에 출력
	if !strings.Contains(out, "Available Commands") {
		t.Error("'Available Commands' 섹션이 출력에 없음")
	}
	if !strings.Contains(out, "status") {
		t.Error("'status' 커맨드가 출력에 없음")
	}
}

func TestFlagGroupNoteInHelp(t *testing.T) {
	cmd := &wcli.Command{Use: "app", Short: "테스트"}
	var jsonFlag, yamlFlag bool
	cmd.Flags().BoolVar(&jsonFlag, "json", "", false, "JSON 출력")
	cmd.Flags().BoolVar(&yamlFlag, "yaml", "", false, "YAML 출력")
	cmd.Flags().MarkFlagsMutuallyExclusive("json", "yaml")

	var buf bytes.Buffer
	cmd.OutWriter = &buf
	cmd.Help()
	out := buf.String()

	if !strings.Contains(out, "mutually exclusive") {
		t.Errorf("'mutually exclusive' 주석이 출력에 없음: %q", out)
	}
	if !strings.Contains(out, "json") || !strings.Contains(out, "yaml") {
		t.Errorf("플래그 이름이 출력에 없음: %q", out)
	}
}

func TestCategorizedLocalFlagsInHelp(t *testing.T) {
	cmd := &wcli.Command{Use: "app", Short: "테스트"}
	var host, token string
	var verbose bool

	cmd.Flags().StringVar(&host, "host", "H", "localhost", "대상 호스트")
	cmd.Flags().StringVar(&token, "token", "t", "", "인증 토큰")
	cmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")

	if err := cmd.Flags().SetCategory("host", "연결"); err != nil {
		t.Fatalf("host 카테고리 설정 실패: %v", err)
	}
	if err := cmd.Flags().SetCategory("token", "인증"); err != nil {
		t.Fatalf("token 카테고리 설정 실패: %v", err)
	}

	var buf bytes.Buffer
	cmd.OutWriter = &buf
	cmd.Help()
	out := buf.String()

	if !strings.Contains(out, "Flags:") {
		t.Fatalf("'Flags:' 섹션이 출력에 없음: %q", out)
	}
	if !strings.Contains(out, "연결:") || !strings.Contains(out, "인증:") {
		t.Fatalf("카테고리 헤더가 출력에 없음: %q", out)
	}
	if !strings.Contains(out, "host") || !strings.Contains(out, "token") || !strings.Contains(out, "verbose") {
		t.Fatalf("카테고리 플래그 출력 누락: %q", out)
	}

	if strings.Index(out, "verbose") > strings.Index(out, "연결:") {
		t.Fatalf("카테고리 없는 플래그는 카테고리 헤더보다 먼저 출력되어야 함: %q", out)
	}
}

func TestCategorizedGlobalFlagsInHelp(t *testing.T) {
	root := &wcli.Command{Use: "app", Short: "테스트 앱"}
	child := &wcli.Command{Use: "serve", Short: "서버 실행"}
	root.AddCommand(child)

	var configPath, profile string
	root.PersistentFlags().StringVar(&configPath, "config", "c", "", "설정 파일 경로")
	root.PersistentFlags().StringVar(&profile, "profile", "p", "default", "실행 프로필")

	if err := root.PersistentFlags().SetCategory("config", "입력"); err != nil {
		t.Fatalf("config 카테고리 설정 실패: %v", err)
	}
	if err := root.PersistentFlags().SetCategory("profile", "실행"); err != nil {
		t.Fatalf("profile 카테고리 설정 실패: %v", err)
	}

	var buf bytes.Buffer
	child.OutWriter = &buf
	child.Help()
	out := buf.String()

	if !strings.Contains(out, "Global Flags:") {
		t.Fatalf("'Global Flags:' 섹션이 출력에 없음: %q", out)
	}
	if !strings.Contains(out, "입력:") || !strings.Contains(out, "실행:") {
		t.Fatalf("글로벌 플래그 카테고리 헤더가 출력에 없음: %q", out)
	}
	if !strings.Contains(out, "config") || !strings.Contains(out, "profile") {
		t.Fatalf("글로벌 플래그 출력 누락: %q", out)
	}
}

func TestCustomHelpTemplateFlagFields(t *testing.T) {
	var buf bytes.Buffer
	cmd := &wcli.Command{
		Use:          "custom_app",
		Short:        "커스텀 템플릿 테스트",
		OutWriter:    &buf,
		HelpTemplate: `{{range .LocalFlags}}Flag: {{.Name}} ({{.Shorthand}}), Type: {{.TypeStr}}, Required: {{.Required}}, Usage: {{.Usage}}{{end}}`,
		Run:          func(ctx *wcli.Context) error { return nil },
	}

	var name string
	cmd.Flags().StringVar(&name, "name", "n", "", "사용자 이름")
	_ = cmd.Flags().MarkRequired("name")

	if err := cmd.Execute([]string{"--help"}); err != nil {
		t.Fatalf("--help 실행 에러: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Flag: name (n), Type: string, Required: true, Usage: 사용자 이름") {
		t.Errorf("커스텀 템플릿 플래그 필드 렌더링 실패, 실제 출력: %q", out)
	}
}
