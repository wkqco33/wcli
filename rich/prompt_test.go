package rich_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wkqco33/wcli/rich"
)

func TestFPrompt_Default(t *testing.T) {
	in := strings.NewReader("\n") // 빈 입력
	var out bytes.Buffer
	result, err := rich.FPrompt(&out, in, "이름", "홍길동")
	if err != nil {
		t.Fatalf("FPrompt 오류: %v", err)
	}
	if result != "홍길동" {
		t.Errorf("기본값 반환 기대: 홍길동, 실제: %q", result)
	}
}

func TestFPrompt_UserInput(t *testing.T) {
	in := strings.NewReader("철수\n")
	var out bytes.Buffer
	result, err := rich.FPrompt(&out, in, "이름", "홍길동")
	if err != nil {
		t.Fatalf("FPrompt 오류: %v", err)
	}
	if result != "철수" {
		t.Errorf("입력값 반환 기대: 철수, 실제: %q", result)
	}
}

func TestFConfirm_Yes(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer
	result, err := rich.FConfirm(&out, in, "계속하시겠습니까?", false)
	if err != nil {
		t.Fatalf("FConfirm 오류: %v", err)
	}
	if !result {
		t.Error("'y' 입력 시 true 반환 기대")
	}
}

func TestFConfirm_No(t *testing.T) {
	in := strings.NewReader("no\n")
	var out bytes.Buffer
	result, err := rich.FConfirm(&out, in, "계속하시겠습니까?", true)
	if err != nil {
		t.Fatalf("FConfirm 오류: %v", err)
	}
	if result {
		t.Error("'no' 입력 시 false 반환 기대")
	}
}

func TestFConfirm_Default(t *testing.T) {
	in := strings.NewReader("\n") // 빈 입력
	var out bytes.Buffer
	result, err := rich.FConfirm(&out, in, "계속하시겠습니까?", true)
	if err != nil {
		t.Fatalf("FConfirm 오류: %v", err)
	}
	if !result {
		t.Error("빈 입력 시 defaultVal(true) 반환 기대")
	}
}

func TestFSelect_ValidChoice(t *testing.T) {
	in := strings.NewReader("2\n")
	var out bytes.Buffer
	result, err := rich.FSelect(&out, in, "환경 선택", []string{"dev", "staging", "prod"})
	if err != nil {
		t.Fatalf("FSelect 오류: %v", err)
	}
	if result != "staging" {
		t.Errorf("2번 선택 시 'staging' 기대, 실제: %q", result)
	}
}

func TestFSelect_EmptyChoices(t *testing.T) {
	in := strings.NewReader("1\n")
	var out bytes.Buffer
	_, err := rich.FSelect(&out, in, "선택", nil)
	if err == nil {
		t.Error("선택지 없을 때 에러 기대")
	}
}

func TestFSelect_ExceedRetry(t *testing.T) {
	// 항상 잘못된 번호 입력 (3회 초과)
	in := strings.NewReader("99\n99\n99\n")
	var out bytes.Buffer
	_, err := rich.FSelect(&out, in, "선택", []string{"a", "b"})
	if err == nil {
		t.Fatal("3회 실패 후 에러 기대")
	}
	// 재시도 루프가 매 시도마다 bufio.Reader를 새로 만들면 두 번째 시도부터
	// 미리 읽혀 버려진 입력 때문에 EOF가 나서 "3회 재시도" 에러가 아니라
	// 엉뚱한(예: EOF) 에러로 조기 종료된다. 정확히 3회 시도 후의 에러인지 확인한다.
	if !strings.Contains(err.Error(), "3 attempts") {
		t.Errorf("3회 재시도 소진 에러를 기대했지만 실제: %v", err)
	}
}

func TestFPrompt_ReuseSameReaderAcrossCalls(t *testing.T) {
	// 하나의 io.Reader에 여러 줄이 한 번에 준비되어 있고(파이프/스크립트 입력을
	// 흉내), 이를 여러 번의 별도 FPrompt 호출에 걸쳐 순서대로 소비하는 시나리오.
	// 과거에는 호출마다 새 bufio.Reader를 만들어 미리 읽힌 뒷줄이 유실되었다.
	in := strings.NewReader("first\nsecond\nthird\n")
	var out bytes.Buffer

	a, err := rich.FPrompt(&out, in, "이름1", "")
	if err != nil {
		t.Fatalf("첫 번째 FPrompt 오류: %v", err)
	}
	if a != "first" {
		t.Errorf("첫 번째 값 기대: first, 실제: %q", a)
	}

	b, err := rich.FPrompt(&out, in, "이름2", "")
	if err != nil {
		t.Fatalf("두 번째 FPrompt 오류: %v", err)
	}
	if b != "second" {
		t.Errorf("두 번째 값 기대: second, 실제: %q", b)
	}

	c, err := rich.FPrompt(&out, in, "이름3", "")
	if err != nil {
		t.Fatalf("세 번째 FPrompt 오류: %v", err)
	}
	if c != "third" {
		t.Errorf("세 번째 값 기대: third, 실제: %q", c)
	}
}

func TestFPasswordPrompt(t *testing.T) {
	in := strings.NewReader("mysecret\n")
	var out bytes.Buffer
	result, err := rich.FPasswordPrompt(&out, in, "비밀번호")
	if err != nil {
		t.Fatalf("FPasswordPrompt 오류: %v", err)
	}
	if result != "mysecret" {
		t.Errorf("비밀번호 값 기대: mysecret, 실제: %q", result)
	}
}
