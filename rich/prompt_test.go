package rich_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/seoyc/wcli/rich"
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
		t.Error("3회 실패 후 에러 기대")
	}
}
