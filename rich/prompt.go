package rich

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Prompt 레이블을 출력하고 os.Stdin에서 한 줄 입력을 받습니다.
// 빈 입력이면 defaultVal을 반환합니다.
func Prompt(label, defaultVal string) (string, error) {
	return FPrompt(os.Stderr, os.Stdin, label, defaultVal)
}

// Confirm 레이블을 출력하고 Y/N 입력을 받습니다.
// 빈 입력이면 defaultVal을 반환합니다.
func Confirm(label string, defaultVal bool) (bool, error) {
	return FConfirm(os.Stderr, os.Stdin, label, defaultVal)
}

// Select 레이블과 선택지를 출력하고 번호 선택 입력을 받습니다.
// 범위 초과 시 최대 3회 재시도합니다.
func Select(label string, choices []string) (string, error) {
	return FSelect(os.Stderr, os.Stdin, label, choices)
}

// FPrompt 테스트 가능한 Prompt 내부 구현입니다.
func FPrompt(out io.Writer, in io.Reader, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		Fprint(out, "%s [%s]: ", label, defaultVal)
	} else {
		Fprint(out, "%s: ", label)
	}

	line, err := readLine(in)
	if err != nil {
		return "", err
	}
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

// FConfirm 테스트 가능한 Confirm 내부 구현입니다.
func FConfirm(out io.Writer, in io.Reader, label string, defaultVal bool) (bool, error) {
	hint := "[y/N]"
	if defaultVal {
		hint = "[Y/n]"
	}
	Fprint(out, "%s %s: ", label, hint)

	line, err := readLine(in)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(line) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return defaultVal, nil
	default:
		return false, fmt.Errorf("올바른 값을 입력하세요 (y/n)")
	}
}

// FSelect 테스트 가능한 Select 내부 구현입니다.
// 잘못된 입력 시 최대 3회 재시도합니다.
func FSelect(out io.Writer, in io.Reader, label string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("선택지가 없습니다")
	}

	Fprintln(out, "[bold]%s[/bold]", label)
	for i, c := range choices {
		Fprintln(out, "  [cyan]%d[/cyan]. %s", i+1, c)
	}

	const maxRetry = 3
	for attempt := 0; attempt < maxRetry; attempt++ {
		Fprint(out, "번호 선택 (1-%d): ", len(choices))
		line, err := readLine(in)
		if err != nil {
			return "", err
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(choices) {
			Fprintln(out, "[yellow]1~%d 사이의 숫자를 입력하세요.[/yellow]", len(choices))
			continue
		}
		return choices[n-1], nil
	}
	return "", fmt.Errorf("유효한 선택을 %d회 안에 입력하지 않았습니다", maxRetry)
}

// readLine in에서 한 줄을 읽어 공백을 제거한 후 반환합니다.
func readLine(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
