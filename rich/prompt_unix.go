//go:build linux || darwin

package rich

import (
	"fmt"
	"os"
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
