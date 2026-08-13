//go:build windows

package rich

import "os"

func isTerminalFD(fd int) bool {
	return false
}

func readPasswordNoEcho() (string, error) {
	return readLine(getLineReader(os.Stdin))
}
