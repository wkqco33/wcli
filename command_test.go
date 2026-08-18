package wcli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/internal/testutil"
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
		testutil.AssertEqual(t, cmd.Name(), tt.expected)
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
		testutil.AssertNoErrorf(t, err, "alias %q 실행 실패", name)
		testutil.AssertTruef(t, ran, "alias %q 로 Run이 실행되지 않음", name)
	}
}

func TestOutWriter(t *testing.T) {
	cmd := &wcli.Command{
		Use:   "testapp",
		Short: "test",
		Run:   func(ctx *wcli.Context) error { return nil },
	}

	stdout, _, err := testutil.ExecuteCommand(cmd, "--help")
	testutil.AssertNoError(t, err)
	testutil.AssertContains(t, stdout, "Usage:")
}

// TestSubCommandInheritsWriter 하위 커맨드가 부모의 OutWriter를 상속하되,
// 자식 구조체 필드는 변경되지 않아야 한다(멱등성/동시성 안전).
func TestSubCommandInheritsWriter(t *testing.T) {
	var buf bytes.Buffer
	sub := &wcli.Command{
		Use: "child",
		Run: func(ctx *wcli.Context) error { return nil },
	}
	root := &wcli.Command{
		Use:       "root",
		OutWriter: &buf,
	}
	root.AddCommand(sub)

	if err := root.Execute([]string{"child", "--help"}); err != nil {
		t.Fatalf("실행 실패: %v", err)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("자식 도움말이 부모 OutWriter로 출력되어야 함. 출력: %q", buf.String())
	}
	// 자식 구조체 필드는 건드리면 안 됨
	if sub.OutWriter != nil {
		t.Error("자식 OutWriter 필드가 실행 중 변경됨(mutation 발생)")
	}

	// 두 번째 실행에 다른 Writer를 줘도 첫 실행의 Writer가 남아 오염시키지 않아야 함
	var buf2 bytes.Buffer
	root.OutWriter = &buf2
	if err := root.Execute([]string{"child", "--help"}); err != nil {
		t.Fatalf("재실행 실패: %v", err)
	}
	if !strings.Contains(buf2.String(), "Usage:") {
		t.Errorf("재실행 시 새 OutWriter로 출력되어야 함. 출력: %q", buf2.String())
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

// TestVersionNotTriggeredAsFlagValue --version이 비-bool 플래그의 값으로 전달된 경우
// 버전 출력으로 오인하지 않아야 한다.
func TestVersionNotTriggeredAsFlagValue(t *testing.T) {
	var buf strings.Builder
	var name string
	var ran bool
	cmd := &wcli.Command{
		Use:       "myapp",
		Version:   "1.2.3",
		OutWriter: &buf,
		Run:       func(ctx *wcli.Context) error { ran = true; return nil },
	}
	cmd.Flags().StringVar(&name, "name", "n", "", "이름")

	if err := cmd.Execute([]string{"--name", "--version"}); err != nil {
		t.Fatalf("실행 실패: %v", err)
	}
	if !ran {
		t.Error("Run이 실행되어야 함 (--version은 --name의 값)")
	}
	if name != "--version" {
		t.Errorf("--name 값이 '--version'이어야 함, 실제: %q", name)
	}
	if strings.Contains(buf.String(), "1.2.3") {
		t.Errorf("버전이 출력되면 안 됨: %q", buf.String())
	}
}

// TestDuplicateFlagPanics 같은 이름/단축키를 두 번 등록하면 panic해야 한다.
func TestDuplicateFlagPanics(t *testing.T) {
	t.Run("같은 이름", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("중복 이름 등록 시 panic해야 함")
			}
		}()
		var a, b string
		fs := wcli.NewFlagSet()
		fs.StringVar(&a, "out", "o", "", "")
		fs.StringVar(&b, "out", "", "", "")
	})
	t.Run("같은 단축키", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("중복 단축키 등록 시 panic해야 함")
			}
		}()
		var a, b string
		fs := wcli.NewFlagSet()
		fs.StringVar(&a, "out", "o", "", "")
		fs.StringVar(&b, "output", "o", "", "")
	})
}

// TestExecuteContext 주입한 컨텍스트가 Run에 전달되고 값/취소가 전파되는지 검증한다.
func TestExecuteContext(t *testing.T) {
	type ctxKey string
	const key ctxKey = "trace"

	// 1. 값 전파
	cmd := &wcli.Command{
		Use: "app",
		Run: func(ctx *wcli.Context) error {
			if v := ctx.Value(key); v != "abc" {
				t.Errorf("컨텍스트 값 전파 실패, 기대 'abc' 실제 %v", v)
			}
			return nil
		},
	}
	parent := context.WithValue(context.Background(), key, "abc")
	if err := cmd.ExecuteContext(parent, nil); err != nil {
		t.Fatalf("실행 실패: %v", err)
	}

	// 2. 취소 전파
	cancelCmd := &wcli.Command{
		Use: "app",
		Run: func(ctx *wcli.Context) error {
			return ctx.Err()
		},
	}
	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := cancelCmd.ExecuteContext(cctx, nil); err != context.Canceled {
		t.Errorf("취소 전파 실패, 기대 context.Canceled 실제 %v", err)
	}

	// 3. nil 컨텍스트는 Background로 대체되어 패닉하지 않음
	nilCmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
	var nilCtx context.Context
	if err := nilCmd.ExecuteContext(nilCtx, nil); err != nil {
		t.Errorf("nil 컨텍스트 실행 실패: %v", err)
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

func TestFuzzyCommandSuggestion(t *testing.T) {
	newRoot := func() *wcli.Command {
		root := &wcli.Command{Use: "app"}
		root.AddCommand(&wcli.Command{Use: "deploy", Short: "배포"})
		root.AddCommand(&wcli.Command{Use: "status", Short: "상태"})
		return root
	}

	t.Run("유사 커맨드 제안 포함", func(t *testing.T) {
		err := newRoot().Execute([]string{"deply"})
		if err == nil {
			t.Fatal("에러가 발생해야 함")
		}
		if !strings.Contains(err.Error(), "deploy") {
			t.Errorf("'deploy' 제안이 에러 메시지에 없음: %v", err)
		}
		if !strings.Contains(err.Error(), "Did you mean") {
			t.Errorf("'Did you mean' 문구가 에러 메시지에 없음: %v", err)
		}
	})

	t.Run("편집거리 큰 입력 - 제안 없이 도움말 출력", func(t *testing.T) {
		// Run이 없는 루트: unknown args → 도움말 출력 → ErrHelp → Execute()에서 nil 반환
		err := newRoot().Execute([]string{"xyz123"})
		if err != nil {
			// 제안이 있으면 에러, 없으면 nil (help)
			if strings.Contains(err.Error(), "Did you mean") {
				t.Errorf("제안이 없어야 하는데 포함됨: %v", err)
			}
		}
	})
}

func TestExecuteWritesErrorToErrWriter(t *testing.T) {
	var errBuf strings.Builder
	cmd := &wcli.Command{
		Use:       "app",
		ErrWriter: &errBuf,
		Run: func(ctx *wcli.Context) error {
			return fmt.Errorf("boom [danger]")
		},
	}

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("에러가 반환되어야 함")
	}
	out := errBuf.String()
	if !strings.Contains(out, "Error:") || !strings.Contains(out, "boom [danger]") {
		t.Fatalf("ErrWriter 출력이 예상과 다름: %q", out)
	}
}

func TestExecuteSilenceErrorsSuppressesErrWriter(t *testing.T) {
	var errBuf strings.Builder
	cmd := &wcli.Command{
		Use:           "app",
		ErrWriter:     &errBuf,
		SilenceErrors: true,
		Run: func(ctx *wcli.Context) error {
			return fmt.Errorf("boom")
		},
	}

	err := cmd.Execute(nil)
	if err == nil {
		t.Fatal("에러가 반환되어야 함")
	}
	if errBuf.Len() != 0 {
		t.Fatalf("SilenceErrors=true이면 ErrWriter 출력이 없어야 함: %q", errBuf.String())
	}
}
