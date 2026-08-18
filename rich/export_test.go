package rich

// SetTerminalHooksForTesting replaces terminal detection and password reading functions for testing.
func SetTerminalHooksForTesting(isTerm func(int) bool, readPass func() (string, error)) func() {
	origTerm := isTerminalFDFunc
	origPass := readPasswordNoEchoFunc

	if isTerm != nil {
		isTerminalFDFunc = isTerm
	}
	if readPass != nil {
		readPasswordNoEchoFunc = readPass
	}

	return func() {
		isTerminalFDFunc = origTerm
		readPasswordNoEchoFunc = origPass
	}
}
