package wcli_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/seoyc/wcli"
)

func TestCommandName(t *testing.T) {
	tests := []struct {
		use      string
		expected string
	}{
		{"serve", "serve"},
		{"serve [address]", "serve"},
		{"get [resource] [name]", "get"},
		{"", ""},
	}
	for _, tt := range tests {
		cmd := &wcli.Command{Use: tt.use}
		if got := cmd.Name(); got != tt.expected {
			t.Errorf("Name() for Use=%q = %q, want %q", tt.use, got, tt.expected)
		}
	}
}

func TestCommandAliases(t *testing.T) {
	var ran bool
	root := &wcli.Command{Use: "root", Run: func(ctx *wcli.Context) error { return nil }}
	sub := &wcli.Command{
		Use:     "server",
		Aliases: []string{"serve", "s"},
		Run: func(ctx *wcli.Context) error {
			ran = true
			return nil
		},
	}
	root.AddCommand(sub)

	for _, name := range []string{"server", "serve", "s"} {
		ran = false
		err := root.Execute([]string{name})
		if err != nil {
			t.Errorf("alias %q 실행 실패: %v", name, err)
		}
		if !ran {
			t.Errorf("alias %q 로 Run이 실행되지 않음", name)
		}
	}
}

func TestOutWriter(t *testing.T) {
	var buf bytes.Buffer
	cmd := &wcli.Command{
		Use:       "testapp",
		Short:     "test",
		OutWriter: &buf,
		Run:       func(ctx *wcli.Context) error { return nil },
	}

	err := cmd.Execute([]string{"--help"})
	if err != nil {
		t.Fatalf("--help 실행 실패: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("OutWriter로 출력된 도움말에 'Usage:'가 없음. 출력: %q", output)
	}
}

func TestPersistentFlags(t *testing.T) {
	var verbose bool
	var output string

	root := &wcli.Command{
		Use:           "root",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	root.PersistentFlags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")

	sub := &wcli.Command{
		Use: "sub",
		Run: func(ctx *wcli.Context) error { return nil },
	}
	sub.Flags().StringVar(&output, "output", "o", "", "출력 파일")
	root.AddCommand(sub)

	// 서브커맨드에서 부모의 persistent 플래그 사용
	err := root.Execute([]string{"sub", "--verbose", "--output", "file.txt"})
	if err != nil {
		t.Fatalf("persistent flag 테스트 실패: %v", err)
	}
	if !verbose {
		t.Error("verbose가 true여야 함")
	}
	if output != "file.txt" {
		t.Errorf("output 불일치. 예상: file.txt, 실제: %s", output)
	}
}

func TestPersistentHooks(t *testing.T) {
	var order []string

	root := &wcli.Command{
		Use:           "root",
		SilenceErrors: true,
		PersistentPreRun: func(ctx *wcli.Context) error {
			order = append(order, "root-pre")
			return nil
		},
		PersistentPostRun: func(ctx *wcli.Context) error {
			order = append(order, "root-post")
			return nil
		},
		Run: func(ctx *wcli.Context) error { return nil },
	}

	sub := &wcli.Command{
		Use: "sub",
		PersistentPreRun: func(ctx *wcli.Context) error {
			order = append(order, "sub-pre")
			return nil
		},
		PersistentPostRun: func(ctx *wcli.Context) error {
			order = append(order, "sub-post")
			return nil
		},
		Run: func(ctx *wcli.Context) error {
			order = append(order, "run")
			return nil
		},
	}
	root.AddCommand(sub)

	err := root.Execute([]string{"sub"})
	if err != nil {
		t.Fatalf("실행 실패: %v", err)
	}
	// 루트→서브 순으로 PersistentPreRun, 서브→루트 순으로 PersistentPostRun
	expected := []string{"root-pre", "sub-pre", "run", "sub-post", "root-post"}
	if !reflect.DeepEqual(order, expected) {
		t.Errorf("훅 실행 순서 불일치.\n예상: %v\n실제: %v", expected, order)
	}
}

func TestRequiredFlag(t *testing.T) {
	var name string
	cmd := &wcli.Command{
		Use:           "testcmd",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().StringVar(&name, "name", "n", "", "이름")
	if err := cmd.Flags().MarkRequired("name"); err != nil {
		t.Fatalf("MarkRequired 실패: %v", err)
	}

	// required 플래그 미설정 → 에러
	err := cmd.Execute([]string{})
	if err == nil {
		t.Error("required 플래그 미설정 시 에러가 발생해야 함")
	}

	// required 플래그 설정 → 정상
	err = cmd.Execute([]string{"--name", "alice"})
	if err != nil {
		t.Errorf("required 플래그 설정 후 에러 없어야 함: %v", err)
	}
}

func TestFlagValidation(t *testing.T) {
	var count int
	cmd := &wcli.Command{
		Use:           "testcmd",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().IntVar(&count, "count", "c", 0, "카운트")
	cmd.Flags().SetValidation("count", func(val string) error {
		if val == "0" {
			return fmt.Errorf("count must be > 0")
		}
		return nil
	})

	err := cmd.Execute([]string{"--count", "0"})
	if err == nil {
		t.Error("검증 실패 시 에러가 발생해야 함")
	}

	err = cmd.Execute([]string{"--count", "5"})
	if err != nil {
		t.Errorf("검증 통과 후 에러 없어야 함: %v", err)
	}
}

func TestVersionFlag(t *testing.T) {
	var buf strings.Builder
	cmd := &wcli.Command{
		Use:       "myapp",
		Version:   "1.2.3",
		OutWriter: &buf,
		Run:       func(ctx *wcli.Context) error { return nil },
	}

	err := cmd.Execute([]string{"--version"})
	if err != nil {
		t.Fatalf("--version 실행 실패: %v", err)
	}
	if !strings.Contains(buf.String(), "1.2.3") {
		t.Errorf("버전 출력에 '1.2.3'이 없음: %q", buf.String())
	}
}

func TestHelpFunc(t *testing.T) {
	var customCalled bool
	var buf strings.Builder
	cmd := &wcli.Command{
		Use:       "myapp",
		OutWriter: &buf,
		HelpFunc: func(cmd *wcli.Command, w io.Writer) {
			customCalled = true
			fmt.Fprintln(w, "custom help output")
		},
		Run: func(ctx *wcli.Context) error { return nil },
	}

	err := cmd.Execute([]string{"--help"})
	if err != nil {
		t.Fatalf("--help 실행 실패: %v", err)
	}
	if !customCalled {
		t.Error("HelpFunc가 호출되지 않음")
	}
	if !strings.Contains(buf.String(), "custom help output") {
		t.Errorf("커스텀 도움말 출력 없음: %q", buf.String())
	}
}

func TestCommandExecute(t *testing.T) {
	rootCmd := &wcli.Command{
		Use:   "root",
		Short: "루트 명령어",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) > 0 && ctx.Args[0] == "fail" {
				return wcli.ErrCommandNotFound
			}
			return nil
		},
	}

	subCmd := &wcli.Command{
		Use:   "sub",
		Short: "하위 명령어",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 1 || ctx.Args[0] != "arg1" {
				t.Errorf("예상되는 인자가 다름: %v", ctx.Args)
			}
			return nil
		},
	}

	rootCmd.AddCommand(subCmd)

	t.Run("루트 명령어 실행", func(t *testing.T) {
		err := rootCmd.Execute([]string{})
		if err != nil {
			t.Errorf("루트 명령어 실행 실패: %v", err)
		}
	})

	t.Run("하위 명령어 실행", func(t *testing.T) {
		err := rootCmd.Execute([]string{"sub", "arg1"})
		if err != nil {
			t.Errorf("하위 명령어 실행 실패: %v", err)
		}
	})

	t.Run("정의되지 않은 하위 명령어 - 루트의 인자로 전달", func(t *testing.T) {
		// sub가 아닌 unknown을 입력하면 루트 명령어의 Run이 실행되며 args로 전달됨
		err := rootCmd.Execute([]string{"unknown"})
		if err != nil {
			t.Errorf("에러 발생: %v", err)
		}
	})

	t.Run("루트 명령어 명시적 실패", func(t *testing.T) {
		err := rootCmd.Execute([]string{"fail"})
		if err != wcli.ErrCommandNotFound {
			t.Errorf("기대되는 에러가 발생하지 않음: %v", err)
		}
	})
}
