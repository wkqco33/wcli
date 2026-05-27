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
