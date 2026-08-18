// Package testutil provides lightweight, zero-dependency test helpers and assertions for wcli.
package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
)

// ExecuteCommand runs a wcli.Command with captured stdout and stderr buffers.
// It returns the stdout output, stderr output, and any execution error.
func ExecuteCommand(cmd *wcli.Command, args ...string) (stdout string, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	origOut := cmd.OutWriter
	origErr := cmd.ErrWriter

	cmd.OutWriter = &outBuf
	cmd.ErrWriter = &errBuf

	defer func() {
		cmd.OutWriter = origOut
		cmd.ErrWriter = origErr
	}()

	err = cmd.Execute(args)
	return outBuf.String(), errBuf.String(), err
}

// AssertEqual checks if got equals want.
func AssertEqual[T comparable](t *testing.T, got, want T, msg ...string) {
	t.Helper()
	if got != want {
		m := getMsg("AssertEqual failed", msg...)
		t.Fatalf("%s: got %v, want %v", m, got, want)
	}
}

// AssertEqualf checks if got equals want with formatted message.
func AssertEqualf[T comparable](t *testing.T, got, want T, format string, args ...any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", fmt.Sprintf(format, args...), got, want)
	}
}

// AssertNotEqual checks if got does not equal want.
func AssertNotEqual[T comparable](t *testing.T, got, want T, msg ...string) {
	t.Helper()
	if got == want {
		m := getMsg("AssertNotEqual failed", msg...)
		t.Fatalf("%s: expected value not to equal %v", m, want)
	}
}

// AssertContains checks if s contains substr.
func AssertContains(t *testing.T, s, substr string, msg ...string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		m := getMsg("AssertContains failed", msg...)
		t.Fatalf("%s: string %q does not contain %q", m, s, substr)
	}
}

// AssertNotContains checks if s does not contain substr.
func AssertNotContains(t *testing.T, s, substr string, msg ...string) {
	t.Helper()
	if strings.Contains(s, substr) {
		m := getMsg("AssertNotContains failed", msg...)
		t.Fatalf("%s: string %q contains %q (expected not to)", m, s, substr)
	}
}

// AssertNoError checks if err is nil.
func AssertNoError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err != nil {
		m := getMsg("AssertNoError failed", msg...)
		t.Fatalf("%s: unexpected error: %v", m, err)
	}
}

// AssertNoErrorf checks if err is nil with formatted message.
func AssertNoErrorf(t *testing.T, err error, format string, args ...any) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", fmt.Sprintf(format, args...), err)
	}
}

// AssertError checks if err is not nil.
func AssertError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err == nil {
		m := getMsg("AssertError failed", msg...)
		t.Fatalf("%s: expected an error, got nil", m)
	}
}

// AssertErrorf checks if err is not nil with formatted message.
func AssertErrorf(t *testing.T, err error, format string, args ...any) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected an error, got nil", fmt.Sprintf(format, args...))
	}
}

// AssertErrorIs checks if gotErr wraps or matches targetErr using errors.Is.
func AssertErrorIs(t *testing.T, gotErr, targetErr error, msg ...string) {
	t.Helper()
	if !errors.Is(gotErr, targetErr) {
		m := getMsg("AssertErrorIs failed", msg...)
		t.Fatalf("%s: error %v is not target error %v", m, gotErr, targetErr)
	}
}

// AssertErrorAs checks if gotErr can be unwrapped into target.
func AssertErrorAs(t *testing.T, gotErr error, target any, msg ...string) bool {
	t.Helper()
	if !errors.As(gotErr, target) {
		m := getMsg("AssertErrorAs failed", msg...)
		t.Fatalf("%s: error %v cannot be cast to %T", m, gotErr, target)
		return false
	}
	return true
}

// AssertTrue checks if condition is true.
func AssertTrue(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		m := getMsg("AssertTrue failed", msg...)
		t.Fatalf("%s: expected true, got false", m)
	}
}

// AssertTruef checks if condition is true with formatted message.
func AssertTruef(t *testing.T, condition bool, format string, args ...any) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true, got false", fmt.Sprintf(format, args...))
	}
}

// AssertFalse checks if condition is false.
func AssertFalse(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		m := getMsg("AssertFalse failed", msg...)
		t.Fatalf("%s: expected false, got true", m)
	}
}

// AssertLen checks if the length of slice matches expectedLen.
func AssertLen[T any](t *testing.T, slice []T, expectedLen int, msg ...string) {
	t.Helper()
	if len(slice) != expectedLen {
		m := getMsg("AssertLen failed", msg...)
		t.Fatalf("%s: got length %d, want %d", m, len(slice), expectedLen)
	}
}

// AssertPanics checks if fn panics.
func AssertPanics(t *testing.T, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			m := getMsg("AssertPanics failed", msg...)
			t.Fatalf("%s: expected function to panic, but it did not", m)
		}
	}()
	fn()
}

// AssertNotPanics checks that fn does not panic.
func AssertNotPanics(t *testing.T, fn func(), msg ...string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			m := getMsg("AssertNotPanics failed", msg...)
			t.Fatalf("%s: expected function not to panic, but recovered: %v", m, r)
		}
	}()
	fn()
}

// SetEnv sets an environment variable for the duration of the test.
func SetEnv(t *testing.T, key, val string) {
	t.Helper()
	t.Setenv(key, val)
}

func getMsg(defaultMsg string, msg ...string) string {
	if len(msg) > 0 && msg[0] != "" {
		return msg[0]
	}
	return defaultMsg
}
