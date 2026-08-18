package wcli_test

import (
	"testing"
	"time"

	"github.com/wkqco33/wcli"
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

// TestShortFlagCluster 결합 단축 플래그(-vh, -ab, -ofile, -vo file)를 검증한다.
func TestShortFlagCluster(t *testing.T) {
	t.Run("bool 묶음", func(t *testing.T) {
		var a, b, c bool
		cmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
		cmd.Flags().BoolVar(&a, "all", "a", false, "")
		cmd.Flags().BoolVar(&b, "bose", "b", false, "")
		cmd.Flags().BoolVar(&c, "cat", "c", false, "")
		if err := cmd.Execute([]string{"-ab"}); err != nil {
			t.Fatalf("실행 실패: %v", err)
		}
		if !a || !b || c {
			t.Errorf("기대 a=true b=true c=false, 실제 a=%v b=%v c=%v", a, b, c)
		}
	})

	t.Run("bool 후 값 플래그 (붙은 값)", func(t *testing.T) {
		var v bool
		var out string
		cmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
		cmd.Flags().BoolVar(&v, "verbose", "v", false, "")
		cmd.Flags().StringVar(&out, "out", "o", "", "")
		if err := cmd.Execute([]string{"-vofile.txt"}); err != nil {
			t.Fatalf("실행 실패: %v", err)
		}
		if !v || out != "file.txt" {
			t.Errorf("기대 v=true out=file.txt, 실제 v=%v out=%q", v, out)
		}
	})

	t.Run("bool 후 값 플래그 (다음 인자)", func(t *testing.T) {
		var v bool
		var out string
		cmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
		cmd.Flags().BoolVar(&v, "verbose", "v", false, "")
		cmd.Flags().StringVar(&out, "out", "o", "", "")
		if err := cmd.Execute([]string{"-vo", "file.txt"}); err != nil {
			t.Fatalf("실행 실패: %v", err)
		}
		if !v || out != "file.txt" {
			t.Errorf("기대 v=true out=file.txt, 실제 v=%v out=%q", v, out)
		}
	})

	t.Run("미등록 문자는 에러", func(t *testing.T) {
		var a bool
		cmd := &wcli.Command{Use: "app", SilenceErrors: true, Run: func(ctx *wcli.Context) error { return nil }}
		cmd.Flags().BoolVar(&a, "all", "a", false, "")
		if err := cmd.Execute([]string{"-ax"}); err == nil {
			t.Error("미등록 단축키 -x로 에러가 발생해야 함")
		}
	})

	t.Run("다중 문자 단축키 우선", func(t *testing.T) {
		// "ab"가 단일 단축키로 등록된 경우 결합으로 분해하지 않아야 함
		var v bool
		cmd := &wcli.Command{Use: "app", Run: func(ctx *wcli.Context) error { return nil }}
		cmd.Flags().BoolVar(&v, "verbose", "ab", false, "")
		if err := cmd.Execute([]string{"-ab"}); err != nil {
			t.Fatalf("실행 실패: %v", err)
		}
		if !v {
			t.Error("다중 문자 단축키 -ab가 단일로 처리되어야 함")
		}
	})
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

// TestRequiredFlagValidationDeterministic 필수 플래그가 둘 이상 동시에
// 누락된 경우, Validate()가 항상 동일한(이름순으로 가장 앞선) 플래그를
// 보고하는지 확인합니다. 내부적으로 map을 그대로 순회하면 Go의 무작위
// 맵 순회 순서 때문에 실행마다 다른 플래그가 보고될 수 있습니다.
func TestRequiredFlagValidationDeterministic(t *testing.T) {
	var zeta, alpha, mid string

	fs := wcli.NewFlagSet()
	// 등록 순서를 알파벳 순서와 다르게 섞어, 등록 순서가 아니라 정렬 순서로
	// 보고되는지도 함께 검증한다.
	fs.StringVar(&zeta, "zeta", "", "", "z")
	fs.StringVar(&alpha, "alpha", "", "", "a")
	fs.StringVar(&mid, "mid", "", "", "m")
	for _, name := range []string{"zeta", "alpha", "mid"} {
		if err := fs.MarkRequired(name); err != nil {
			t.Fatalf("MarkRequired(%s) 실패: %v", name, err)
		}
	}

	for i := 0; i < 50; i++ {
		err := fs.Validate()
		if err == nil {
			t.Fatal("세 플래그 모두 미설정 상태이므로 에러가 발생해야 함")
		}
		valErr, ok := err.(*wcli.ValidationError)
		if !ok {
			t.Fatalf("*wcli.ValidationError 기대, 실제: %T", err)
		}
		if valErr.FlagName != "alpha" {
			t.Fatalf("반복 #%d: 이름순으로 가장 앞선 'alpha'가 보고되어야 하지만 실제: %q", i, valErr.FlagName)
		}
	}
}
