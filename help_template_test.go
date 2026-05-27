package wcli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/seoyc/wcli"
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
