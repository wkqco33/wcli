//go:build darwin

package rich

import "syscall"

// ioctl 터미널 제어용 syscall 상수 (macOS/BSD)
const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)
