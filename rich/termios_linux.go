//go:build linux

package rich

import "syscall"

// ioctl 터미널 제어용 syscall 상수 (Linux)
const (
	ioctlGetTermios = syscall.TCGETS
	ioctlSetTermios = syscall.TCSETS
)
