package rich

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// readPasswordNoEcho 터미널 에코 없이 비밀번호를 읽습니다.
// 터미널이 아니면 일반 입력으로 폴백합니다.
func readPasswordNoEcho() (string, error) {
	fd := int(os.Stdin.Fd())
	if !isTerminalFD(fd) {
		return readLine(getLineReader(os.Stdin))
	}

	termios, err := getTermios(fd)
	if err != nil {
		return readLine(getLineReader(os.Stdin))
	}

	newTermios := *termios
	newTermios.Lflag &^= syscall.ECHO
	if err := setTermios(fd, &newTermios); err != nil {
		return readLine(getLineReader(os.Stdin))
	}

	line, err := readLine(getLineReader(os.Stdin))

	setTermios(fd, termios)
	fmt.Fprintln(os.Stderr)

	return line, err
}

func isTerminalFD(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlGetTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func getTermios(fd int) (*syscall.Termios, error) {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlGetTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	if err != 0 {
		return nil, err
	}
	return &termios, nil
}

func setTermios(fd int, termios *syscall.Termios) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlSetTermios, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	if err != 0 {
		return err
	}
	return nil
}

// PasswordPrompt 레이블을 출력하고 os.Stdin에서 에코 없이 비밀번호를 입력받습니다.
// 터미널이 아닌 환경에서는 일반 입력으로 폴백합니다.
func PasswordPrompt(label string) (string, error) {
	return FPasswordPrompt(os.Stderr, os.Stdin, label)
}

// FPasswordPrompt 테스트 가능한 PasswordPrompt 내부 구현입니다.
func FPasswordPrompt(out io.Writer, in io.Reader, label string) (string, error) {
	Fprint(out, "%s: ", label)

	if f, ok := in.(*os.File); ok && isTerminalFD(int(f.Fd())) {
		return readPasswordNoEcho()
	}
	return readLine(getLineReader(in))
}

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

	line, err := readLine(getLineReader(in))
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

	line, err := readLine(getLineReader(in))
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
		return false, fmt.Errorf("invalid value, expected y or n")
	}
}

// FSelect 테스트 가능한 Select 내부 구현입니다.
// 잘못된 입력 시 최대 3회 재시도합니다.
func FSelect(out io.Writer, in io.Reader, label string, choices []string) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices provided")
	}

	Fprintln(out, "[bold]%s[/bold]", label)
	for i, c := range choices {
		Fprintln(out, "  [cyan]%d[/cyan]. %s", i+1, c)
	}

	reader := getLineReader(in)

	const maxRetry = 3
	for attempt := 0; attempt < maxRetry; attempt++ {
		Fprint(out, "번호 선택 (1-%d): ", len(choices))
		line, err := readLine(reader)
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
	return "", fmt.Errorf("no valid selection made within %d attempts", maxRetry)
}

// lineReaderCache in(io.Reader)별로 재사용할 *bufio.Reader를 보관합니다.
// bufio.Reader는 내부적으로 미리 읽어들인(read-ahead) 바이트를 버퍼에 보관하므로,
// 같은 in에 대해 호출할 때마다 새 bufio.Reader를 만들면 이전에 미리 읽혀 버퍼에
// 남아있던 다음 줄들이 그대로 유실됩니다. 이 캐시는 같은 in 인스턴스에 대해
// 항상 동일한 bufio.Reader를 재사용해 그 문제를 막습니다.
//
// 이 라이브러리의 전형적인 사용 패턴(프로세스 생애주기 동안 소수의 리더,
// 대개 os.Stdin 하나)에서는 캐시가 무한정 커질 위험이 없습니다.
var lineReaderCache sync.Map // map[io.Reader]*bufio.Reader

// getLineReader in에 대응하는 *bufio.Reader를 반환합니다. 이미 생성된 적이 있으면
// 캐시에서 재사용하고, 없으면 새로 만들어 캐시에 등록합니다. in의 동적 타입이
// 비교 불가능한 경우(맵 키로 쓸 수 없는 경우)에는 캐시를 사용하지 않고 매번
// 새로 생성합니다(과거 동작과 동일하게 안전하게 폴백).
func getLineReader(in io.Reader) *bufio.Reader {
	if br, ok := in.(*bufio.Reader); ok {
		return br
	}
	if !reflect.TypeOf(in).Comparable() {
		return bufio.NewReader(in)
	}
	if cached, ok := lineReaderCache.Load(in); ok {
		return cached.(*bufio.Reader)
	}
	br := bufio.NewReader(in)
	actual, _ := lineReaderCache.LoadOrStore(in, br)
	return actual.(*bufio.Reader)
}

// readLine r에서 한 줄을 읽어 공백을 제거한 후 반환합니다.
func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
