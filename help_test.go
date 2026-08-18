package wcli_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
)

func TestCommandHelp(t *testing.T) {
	cmd := &wcli.Command{
		Use:           "testapp",
		Short:         "테스트 앱",
		Long:          "이것은 테스트 앱입니다.",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}

	var strVal string
	cmd.Flags().StringVar(&strVal, "name", "n", "world", "이름 설정")

	t.Run("--help 플래그로 도움말 출력 후 nil 반환", func(t *testing.T) {
		err := cmd.Execute([]string{"--help"})
		if err != nil {
			t.Errorf("--help 실행 후 nil 에러 기대, 실제: %v", err)
		}
	})

	t.Run("-h 플래그로 도움말 출력 후 nil 반환", func(t *testing.T) {
		err := cmd.Execute([]string{"-h"})
		if err != nil {
			t.Errorf("-h 실행 후 nil 에러 기대, 실제: %v", err)
		}
	})
}

func TestSubCommandHelp(t *testing.T) {
	rootCmd := &wcli.Command{
		Use:           "root",
		Short:         "루트 명령어",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	subCmd := &wcli.Command{
		Use:   "sub",
		Short: "하위 명령어",
		Run:   func(ctx *wcli.Context) error { return nil },
	}
	rootCmd.AddCommand(subCmd)

	t.Run("하위 명령어의 --help 처리", func(t *testing.T) {
		err := rootCmd.Execute([]string{"sub", "--help"})
		if err != nil {
			t.Errorf("sub --help 실행 후 nil 에러 기대, 실제: %v", err)
		}
	})

	t.Run("Run 없는 루트는 도움말 출력 후 nil 반환", func(t *testing.T) {
		noRunRoot := &wcli.Command{
			Use:           "norun",
			Short:         "실행 함수 없는 루트",
			SilenceErrors: true,
		}
		noRunRoot.AddCommand(&wcli.Command{
			Use:   "child",
			Short: "자식 명령어",
			Run:   func(ctx *wcli.Context) error { return nil },
		})
		err := noRunRoot.Execute([]string{})
		if err != nil {
			t.Errorf("Run 없는 루트 실행 후 nil 에러 기대, 실제: %v", err)
		}
	})
}

func TestCommandHooks(t *testing.T) {
	var order []string

	cmd := &wcli.Command{
		Use:           "test",
		SilenceErrors: true,
		PreRun: func(ctx *wcli.Context) error {
			order = append(order, "pre")
			return nil
		},
		Run: func(ctx *wcli.Context) error {
			order = append(order, "run")
			return nil
		},
		PostRun: func(ctx *wcli.Context) error {
			order = append(order, "post")
			return nil
		},
	}

	t.Run("PreRun → Run → PostRun 순서 실행", func(t *testing.T) {
		order = nil
		err := cmd.Execute([]string{})
		if err != nil {
			t.Errorf("실행 실패: %v", err)
		}
		expected := []string{"pre", "run", "post"}
		if !reflect.DeepEqual(order, expected) {
			t.Errorf("실행 순서 불일치. 예상: %v, 실제: %v", expected, order)
		}
	})

	t.Run("PreRun 에러 시 Run, PostRun 미실행", func(t *testing.T) {
		order = nil
		errCmd := &wcli.Command{
			Use:           "errcmd",
			SilenceErrors: true,
			PreRun: func(ctx *wcli.Context) error {
				order = append(order, "pre")
				return wcli.ErrCommandNotFound
			},
			Run: func(ctx *wcli.Context) error {
				order = append(order, "run")
				return nil
			},
			PostRun: func(ctx *wcli.Context) error {
				order = append(order, "post")
				return nil
			},
		}
		err := errCmd.Execute([]string{})
		if err != wcli.ErrCommandNotFound {
			t.Errorf("PreRun 에러가 전파되어야 함, 실제: %v", err)
		}
		expected := []string{"pre"}
		if !reflect.DeepEqual(order, expected) {
			t.Errorf("실행 순서 불일치. 예상: %v, 실제: %v", expected, order)
		}
	})

	t.Run("Run 에러 시 PostRun 미실행", func(t *testing.T) {
		order = nil
		errCmd := &wcli.Command{
			Use:           "errcmd2",
			SilenceErrors: true,
			PreRun: func(ctx *wcli.Context) error {
				order = append(order, "pre")
				return nil
			},
			Run: func(ctx *wcli.Context) error {
				order = append(order, "run")
				return wcli.ErrCommandNotFound
			},
			PostRun: func(ctx *wcli.Context) error {
				order = append(order, "post")
				return nil
			},
		}
		err := errCmd.Execute([]string{})
		if err != wcli.ErrCommandNotFound {
			t.Errorf("Run 에러가 전파되어야 함, 실제: %v", err)
		}
		expected := []string{"pre", "run"}
		if !reflect.DeepEqual(order, expected) {
			t.Errorf("실행 순서 불일치. 예상: %v, 실제: %v", expected, order)
		}
	})
}

func TestUsageLine(t *testing.T) {
	cmdWithSub := &wcli.Command{Use: "app"}
	cmdWithSub.AddCommand(&wcli.Command{
		Use:   "child",
		Short: "자식",
		Run:   func(ctx *wcli.Context) error { return nil },
	})

	cmdWithFlags := &wcli.Command{Use: "app"}
	var strVal string
	cmdWithFlags.Flags().StringVar(&strVal, "name", "n", "", "이름")

	tests := []struct {
		name     string
		cmd      *wcli.Command
		contains string
	}{
		{
			name:     "단순 명령어",
			cmd:      &wcli.Command{Use: "app"},
			contains: "app",
		},
		{
			name:     "하위 명령어 있는 경우 [command] 포함",
			cmd:      cmdWithSub,
			contains: "[command]",
		},
		{
			name:     "플래그 있는 경우 [flags] 포함",
			cmd:      cmdWithFlags,
			contains: "[flags]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cmd.UsageLine()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("UsageLine() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}
