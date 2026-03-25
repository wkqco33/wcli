package wcli_test

import (
	"testing"
	"time"

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

func TestFlagEqualSyntax(t *testing.T) {
	var strVal string
	var intVal int
	var boolVal bool

	cmd := &wcli.Command{
		Use: "testcmd",
		Run: func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().StringVar(&strVal, "output", "o", "", "출력 파일")
	cmd.Flags().IntVar(&intVal, "count", "c", 0, "카운트")
	cmd.Flags().BoolVar(&boolVal, "verbose", "v", false, "상세 출력")

	err := cmd.Execute([]string{"--output=file.txt", "--count=5", "--verbose=true"})
	if err != nil {
		t.Fatalf("--name=value 형식 실행 실패: %v", err)
	}
	if strVal != "file.txt" {
		t.Errorf("strVal 불일치. 예상: file.txt, 실제: %s", strVal)
	}
	if intVal != 5 {
		t.Errorf("intVal 불일치. 예상: 5, 실제: %d", intVal)
	}
	if !boolVal {
		t.Errorf("boolVal 불일치. 예상: true, 실제: %v", boolVal)
	}

	// --verbose=false 형식
	err = cmd.Execute([]string{"--verbose=false"})
	if err != nil {
		t.Fatalf("--verbose=false 실행 실패: %v", err)
	}
	if boolVal {
		t.Errorf("boolVal 불일치. 예상: false, 실제: %v", boolVal)
	}
}

func TestFlagDoubleDashTerminator(t *testing.T) {
	var strVal string
	cmd := &wcli.Command{
		Use: "testcmd",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 2 || ctx.Args[0] != "--not-a-flag" || ctx.Args[1] != "arg2" {
				t.Errorf("예상되는 positional args 불일치: %v", ctx.Args)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&strVal, "name", "n", "", "이름")

	err := cmd.Execute([]string{"--name", "test", "--", "--not-a-flag", "arg2"})
	if err != nil {
		t.Fatalf("-- 종결자 테스트 실패: %v", err)
	}
	if strVal != "test" {
		t.Errorf("strVal 불일치. 예상: test, 실제: %s", strVal)
	}
}

func TestHelpFlagValueNotMisdetected(t *testing.T) {
	var configVal string
	cmd := &wcli.Command{
		Use:           "testcmd",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().StringVar(&configVal, "config", "c", "", "설정값")

	// --config의 값으로 "--help"가 오는 경우 도움말이 오탐되면 안 됨
	err := cmd.Execute([]string{"--config", "--help"})
	if err != nil {
		t.Errorf("--config --help 실행 시 에러 없어야 함, 실제: %v", err)
	}
	if configVal != "--help" {
		t.Errorf("configVal 불일치. 예상: --help, 실제: %s", configVal)
	}
}

func TestNewFlagTypes(t *testing.T) {
	var f64Val float64
	var durVal time.Duration
	var sliceVal []string

	cmd := &wcli.Command{
		Use: "testcmd",
		Run: func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().Float64Var(&f64Val, "ratio", "r", 0.5, "비율")
	cmd.Flags().DurationVar(&durVal, "timeout", "t", 5*time.Second, "타임아웃")
	cmd.Flags().StringSliceVar(&sliceVal, "tag", "", nil, "태그")

	err := cmd.Execute([]string{"--ratio", "3.14", "--timeout", "1m30s", "--tag", "foo", "--tag", "bar"})
	if err != nil {
		t.Fatalf("새 플래그 타입 실행 실패: %v", err)
	}
	if f64Val != 3.14 {
		t.Errorf("f64Val 불일치. 예상: 3.14, 실제: %f", f64Val)
	}
	if durVal != 90*time.Second {
		t.Errorf("durVal 불일치. 예상: 1m30s, 실제: %v", durVal)
	}
	if len(sliceVal) != 2 || sliceVal[0] != "foo" || sliceVal[1] != "bar" {
		t.Errorf("sliceVal 불일치. 예상: [foo bar], 실제: %v", sliceVal)
	}

	// = 구문도 동작해야 함
	err = cmd.Execute([]string{"--ratio=2.71", "--timeout=10s"})
	if err != nil {
		t.Fatalf("= 구문 실행 실패: %v", err)
	}
	if f64Val != 2.71 {
		t.Errorf("f64Val(= 구문) 불일치. 예상: 2.71, 실제: %f", f64Val)
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
