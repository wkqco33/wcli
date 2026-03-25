package wcli_test

import (
	"testing"

	"github.com/seoyc/wcli"
)

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
