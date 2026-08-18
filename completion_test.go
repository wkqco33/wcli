package wcli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
)

func newTestRootForCompletion() *wcli.Command {
	root := &wcli.Command{Use: "myapp", Short: "테스트 앱"}

	deploy := &wcli.Command{Use: "deploy", Short: "배포 실행"}
	var env string
	deploy.Flags().StringVar(&env, "env", "e", "", "환경 (prod/staging)")
	var dryRun bool
	deploy.Flags().BoolVar(&dryRun, "dry-run", "", false, "실제 실행 안 함")

	root.AddCommand(deploy)
	root.AddCommand(&wcli.Command{Use: "status", Short: "상태 확인"})
	return root
}

func TestGenFishCompletion_Basic(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 앱 이름 포함
	if !strings.Contains(out, "complete -c myapp") {
		t.Error("앱 이름 'myapp' 이 출력에 없음")
	}
	// 서브커맨드 등록
	if !strings.Contains(out, "-a deploy") {
		t.Error("서브커맨드 'deploy' 가 출력에 없음")
	}
	if !strings.Contains(out, "-a status") {
		t.Error("서브커맨드 'status' 가 출력에 없음")
	}
}

func TestGenFishCompletion_SubFlags(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// deploy 서브커맨드 플래그
	if !strings.Contains(out, "__fish_seen_subcommand_from deploy") {
		t.Error("deploy 서브커맨드 조건이 출력에 없음")
	}
	if !strings.Contains(out, "-l env") {
		t.Error("--env 플래그가 출력에 없음")
	}
	if !strings.Contains(out, "-l dry-run") {
		t.Error("--dry-run 플래그가 출력에 없음")
	}
}

func TestGenFishCompletion_NotSeenCondition(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 루트 레벨 완성에 not __fish_seen_subcommand_from 조건 포함
	if !strings.Contains(out, "not __fish_seen_subcommand_from") {
		t.Error("not __fish_seen_subcommand_from 조건이 출력에 없음")
	}
}

func TestGenFishCompletion_EscapeDesc(t *testing.T) {
	root := &wcli.Command{Use: "app", Short: "앱"}
	root.AddCommand(&wcli.Command{Use: "run", Short: "it's a command"})

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 작은따옴표 이스케이프 확인
	if strings.Contains(out, "it's") {
		t.Error("작은따옴표가 이스케이프되지 않음")
	}
}

func TestGenBashCompletion_Basic(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenBashCompletion(root, &buf); err != nil {
		t.Fatalf("GenBashCompletion 오류: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"_myapp_bash_autocomplete()",
		`complete -F _myapp_bash_autocomplete myapp`,
		"deploy",
		"status",
		"--env",
		"--dry-run",
		"-e",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("bash completion 출력에 %q 없음: %q", want, out)
		}
	}
}

func TestGenZshCompletion_Basic(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenZshCompletion(root, &buf); err != nil {
		t.Fatalf("GenZshCompletion 오류: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"#compdef myapp",
		`"deploy:배포 실행"`,
		`"status:상태 확인"`,
		"'--env[환경 (prod/staging)]'",
		"'--dry-run[실제 실행 안 함]'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("zsh completion 출력에 %q 없음: %q", want, out)
		}
	}
}

func TestGenZshCompletion_EscapeDesc(t *testing.T) {
	root := &wcli.Command{Use: "app", Short: "앱"}
	root.AddCommand(&wcli.Command{Use: "run", Short: `it'"'"'s "$danger" [cmd]`})

	var buf strings.Builder
	if err := wcli.GenZshCompletion(root, &buf); err != nil {
		t.Fatalf("GenZshCompletion 오류: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, `it's`) || strings.Contains(out, `$danger`) || strings.Contains(out, "[cmd]") {
		t.Errorf("zsh 설명 문자열의 위험 문자가 제거되어야 함: %q", out)
	}
	if !strings.Contains(out, `run:its danger \[cmd\]`) {
		t.Errorf("zsh 설명 문자열 이스케이프 결과가 예상과 다름: %q", out)
	}
}

func TestNewCompletionCommand(t *testing.T) {
	root := newTestRootForCompletion()

	t.Run("shell 미지정", func(t *testing.T) {
		cmd := wcli.NewCompletionCommand(root)
		cmd.SilenceErrors = true
		err := cmd.Execute(nil)
		if err == nil || !strings.Contains(err.Error(), "please specify a shell") {
			t.Fatalf("shell 미지정 에러가 필요함, 실제: %v", err)
		}
	})

	t.Run("미지원 shell", func(t *testing.T) {
		cmd := wcli.NewCompletionCommand(root)
		cmd.SilenceErrors = true
		err := cmd.Execute([]string{"powershell"})
		if err == nil || !strings.Contains(err.Error(), "unsupported shell type") {
			t.Fatalf("미지원 shell 에러가 필요함, 실제: %v", err)
		}
	})

	t.Run("bash 출력은 root writer 사용", func(t *testing.T) {
		var buf bytes.Buffer
		root.OutWriter = &buf
		cmd := wcli.NewCompletionCommand(root)
		cmd.SilenceErrors = true
		if err := cmd.Execute([]string{"bash"}); err != nil {
			t.Fatalf("bash completion 실행 실패: %v", err)
		}
		if !strings.Contains(buf.String(), "_myapp_bash_autocomplete") {
			t.Errorf("root writer로 bash completion이 출력되어야 함: %q", buf.String())
		}
	})
}
