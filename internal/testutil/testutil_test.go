package testutil_test

import (
	"errors"
	"testing"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/internal/testutil"
)

func TestExecuteCommand(t *testing.T) {
	cmd := &wcli.Command{
		Use: "hello [name]",
	}
	cmd.Run = func(ctx *wcli.Context) error {
		name := "world"
		if len(ctx.Args) > 0 {
			name = ctx.Args[0]
		}
		cmd.OutWriter.Write([]byte("Hello, " + name + "!\n"))
		return nil
	}

	stdout, stderr, err := testutil.ExecuteCommand(cmd, "gopher")
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, stderr, "")
	testutil.AssertContains(t, stdout, "Hello, gopher!")
}

func TestAssertions(t *testing.T) {
	testutil.AssertEqual(t, 10, 10)
	testutil.AssertNotEqual(t, 10, 20)
	testutil.AssertContains(t, "apple banana cherry", "banana")
	testutil.AssertNotContains(t, "apple banana cherry", "grape")

	var nilErr error
	testutil.AssertNoError(t, nilErr)

	customErr := errors.New("something went wrong")
	testutil.AssertError(t, customErr)
	testutil.AssertErrorIs(t, customErr, customErr)

	flagErr := &wcli.FlagError{FlagName: "foo", Err: errors.New("bad")}
	var target *wcli.FlagError
	testutil.AssertErrorAs(t, flagErr, &target)
	testutil.AssertEqual(t, target.FlagName, "foo")

	testutil.AssertTrue(t, true)
	testutil.AssertFalse(t, false)
	testutil.AssertLen(t, []int{1, 2, 3}, 3)
	testutil.AssertPanics(t, func() { panic("oops") })
	testutil.AssertNotPanics(t, func() {})
}
