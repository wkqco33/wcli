package config_test

import (
	"testing"

	"github.com/seoyc/wcli/config"
)

func resetConfigState(t *testing.T) {
	t.Helper()
	config.Reset()
	t.Cleanup(config.Reset)
}
