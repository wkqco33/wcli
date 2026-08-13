package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/seoyc/wcli/config"
)

func TestStrictParsingINIErrorIncludesLine(t *testing.T) {
	resetConfigState(t)
	tmpFile, err := os.CreateTemp("", "strict-*.ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "ok=value\ninvalid_line\n"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	config.SetConfigFile(tmpFile.Name())
	config.SetConfigType("ini")
	config.SetStrictParsing(true)

	err = config.ReadInConfig()
	if err == nil {
		t.Fatal("엄격 모드에서 INI 구문 오류가 감지되어야 함")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("에러에 줄 번호가 포함되어야 함: %v", err)
	}
}

func TestLoadWithStrictParsing(t *testing.T) {
	resetConfigState(t)
	if err := os.WriteFile("strict.yaml", []byte("app:\n  ok: yes\n  bad\n"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("strict.yaml")

	type cfg struct {
		App struct {
			Ok string `wcli:"ok"`
		} `wcli:"app"`
	}
	var c cfg
	err := config.Load(&c, config.WithFiles("strict.yaml"), config.WithStrictParsing(true))
	if err == nil {
		t.Fatal("엄격 모드에서 YAML 구문 오류가 감지되어야 함")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("에러에 줄 번호가 포함되어야 함: %v", err)
	}
}

func TestStoreIsolation(t *testing.T) {
	s1 := config.NewStore()
	s2 := config.NewStore()

	s1.Set("app.name", "one")
	s2.Set("app.name", "two")

	if got := s1.GetString("app.name"); got != "one" {
		t.Fatalf("s1 격리 실패: %s", got)
	}
	if got := s2.GetString("app.name"); got != "two" {
		t.Fatalf("s2 격리 실패: %s", got)
	}
}
