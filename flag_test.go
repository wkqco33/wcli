package wcli_test

import (
	"testing"

	"github.com/seoyc/wcli"
)

func TestFlagParsing(t *testing.T) {
	var strVal string
	var intVal int
	var boolVal bool

	cmd := &wcli.Command{
		Use:   "testcmd",
		Short: "플래그 테스트 커맨드",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 1 || ctx.Args[0] != "arg1" {
				t.Errorf("예상되는 인자가 다름: %v", ctx.Args)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&strVal, "name", "n", "default", "이름 설정")
	cmd.Flags().IntVar(&intVal, "count", "c", 0, "카운트 설정")
	cmd.Flags().BoolVar(&boolVal, "verbose", "v", false, "상세 출력")

	args := []string{"--name", "tester", "-c", "5", "-v", "arg1"}

	err := cmd.Execute(args)
	if err != nil {
		t.Fatalf("명령어 실행 실패: %v", err)
	}

	if strVal != "tester" {
		t.Errorf("strVal이 다름. 예상: tester, 실제: %s", strVal)
	}
	if intVal != 5 {
		t.Errorf("intVal이 다름. 예상: 5, 실제: %d", intVal)
	}
	if !boolVal {
		t.Errorf("boolVal이 다름. 예상: true, 실제: %v", boolVal)
	}
}

func TestFlagErrorHandling(t *testing.T) {
	var intVal int

	cmd := &wcli.Command{
		Use: "errcmd",
		Run: func(ctx *wcli.Context) error {
			return nil
		},
	}
	cmd.Flags().IntVar(&intVal, "count", "c", 0, "카운트 설정")

	// 값이 필요한 플래그에 값 미제공
	err := cmd.Execute([]string{"--count"})
	if err == nil {
		t.Errorf("에러가 발생해야 함 (값 누락)")
	}

	// 잘못된 타입 제공
	err = cmd.Execute([]string{"--count", "notanumber"})
	if err == nil {
		t.Errorf("에러가 발생해야 함 (잘못된 타입)")
	}

	// 알 수 없는 플래그
	err = cmd.Execute([]string{"--unknown"})
	if err == nil {
		t.Errorf("에러가 발생해야 함 (알 수 없는 플래그)")
	}
}
