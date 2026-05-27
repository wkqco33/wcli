package wcli

import (
	"errors"
	"fmt"
	"testing"
)

func TestFlagError(t *testing.T) {
	origErr := errors.New("invalid syntax")
	err := &FlagError{
		FlagName:    "timeout",
		CommandName: "run",
		Err:         origErr,
	}

	expectedStr := `command "run" flag --timeout: invalid syntax`
	if err.Error() != expectedStr {
		t.Errorf("expected %q, got %q", expectedStr, err.Error())
	}

	if !errors.Is(err, origErr) {
		t.Errorf("expected Unwrap to yield origErr")
	}

	errNoCmd := &FlagError{
		FlagName: "port",
		Err:      origErr,
	}
	expectedNoCmdStr := `flag --port: invalid syntax`
	if errNoCmd.Error() != expectedNoCmdStr {
		t.Errorf("expected %q, got %q", expectedNoCmdStr, errNoCmd.Error())
	}
}

func TestValidationError(t *testing.T) {
	origErr := fmt.Errorf("must be positive")
	err := &ValidationError{
		FlagName:    "port",
		CommandName: "start",
		Err:         origErr,
	}

	expectedStr := `command "start" flag --port validation failed: must be positive`
	if err.Error() != expectedStr {
		t.Errorf("expected %q, got %q", expectedStr, err.Error())
	}

	if !errors.Is(err, origErr) {
		t.Errorf("expected Unwrap to yield origErr")
	}
}

func TestCommandError(t *testing.T) {
	origErr := errors.New("timeout reached")
	err := &CommandError{
		CommandName: "build",
		Err:         origErr,
	}

	expectedStr := `command "build" execution error: timeout reached`
	if err.Error() != expectedStr {
		t.Errorf("expected %q, got %q", expectedStr, err.Error())
	}

	if !errors.Is(err, origErr) {
		t.Errorf("expected Unwrap to yield origErr")
	}
}
