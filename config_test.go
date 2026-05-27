package wcli_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/seoyc/wcli"
)

// --- 글로벌 config store 테스트 (SetConfigFile / ReadInConfig / Get) ---

func TestJSONConfig(t *testing.T) {
	jsonContent := `{
		"app": {
			"name": "testapp",
			"port": 8080,
			"debug": true
		},
		"database": "postgresql"
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(jsonContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	wcli.SetConfigFile(tmpFile.Name())
	wcli.SetConfigType("json")

	if err := wcli.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	dbVal := wcli.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	appName := wcli.Get("app.name")
	if appName != "testapp" {
		t.Errorf("app.name 값 불일치: 예상 testapp, 실제 %v", appName)
	}

	portVal := wcli.Get("app.port")
	if portVal != 8080.0 && portVal != float64(8080) {
		t.Errorf("app.port 값 불일치: 예상 8080.0, 실제 %v (타입: %v)", portVal, reflect.TypeOf(portVal))
	}

	debugVal := wcli.Get("app.debug")
	if debugVal != true {
		t.Errorf("app.debug 값 불일치: 예상 true, 실제 %v", debugVal)
	}
}

func TestINIConfig(t *testing.T) {
	iniContent := `
# 글로벌 코멘트
database = postgresql

[server]
host = localhost
port = 9000
ssl = "false"
`
	tmpFile, err := os.CreateTemp("", "config-*.ini")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(iniContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	wcli.SetConfigFile(tmpFile.Name())
	wcli.SetConfigType("ini")

	if err := wcli.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	dbVal := wcli.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	hostVal := wcli.Get("server.host")
	if hostVal != "localhost" {
		t.Errorf("server.host 값 불일치: 예상 localhost, 실제 %v", hostVal)
	}

	portVal := wcli.Get("server.port")
	if portVal != "9000" {
		t.Errorf("server.port 값 불일치: 예상 9000, 실제 %v", portVal)
	}

	sslVal := wcli.Get("server.ssl")
	if sslVal != "false" {
		t.Errorf("server.ssl 값 불일치: 예상 false, 실제 %v", sslVal)
	}
}

func TestYAMLConfig(t *testing.T) {
	yamlContent := `
app:
  name: testapp
  port: 8080
database: postgresql
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	wcli.SetConfigFile(tmpFile.Name())
	if err := wcli.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if wcli.Get("database") != "postgresql" {
		t.Errorf("database 값 불일치: %v", wcli.Get("database"))
	}
	if wcli.Get("app.name") != "testapp" {
		t.Errorf("app.name 값 불일치: %v", wcli.Get("app.name"))
	}
	if wcli.Get("app.port") != "8080" {
		t.Errorf("app.port 값 불일치: %v", wcli.Get("app.port"))
	}
}

func TestTOMLConfig(t *testing.T) {
	tomlContent := `
database = "postgresql"

[app]
name = "testapp"
port = "8080"
`
	tmpFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(tomlContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	wcli.SetConfigFile(tmpFile.Name())
	if err := wcli.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if wcli.Get("database") != "postgresql" {
		t.Errorf("database 값 불일치: %v", wcli.Get("database"))
	}
	if wcli.Get("app.name") != "testapp" {
		t.Errorf("app.name 값 불일치: %v", wcli.Get("app.name"))
	}
}

func TestDotEnvConfig(t *testing.T) {
	envContent := `
# 환경 설정
DATABASE=postgresql
APP_NAME=testapp
APP_PORT=8080
`
	tmpFile, err := os.CreateTemp("", "config-*.env")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(envContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	wcli.SetConfigFile(tmpFile.Name())
	if err := wcli.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if wcli.Get("DATABASE") != "postgresql" {
		t.Errorf("DATABASE 값 불일치: %v", wcli.Get("DATABASE"))
	}
	if wcli.Get("APP_NAME") != "testapp" {
		t.Errorf("APP_NAME 값 불일치: %v", wcli.Get("APP_NAME"))
	}
}

// --- struct binding 테스트 (Load / WriteDefault) ---

type testBindConfig struct {
	Port    int     `wcli:"PORT" default:"8080"`
	Host    string  `wcli:"HOST" default:"localhost"`
	Debug   bool    `wcli:"DEBUG"`
	Timeout float64 `default:"30.5"`
	DB      struct {
		User string `wcli:"USER"`
		Pass string `wcli:"PASS"`
	} `wcli:"DATABASE"`
}

func TestLoadDotEnvAndYAML(t *testing.T) {
	dotenvContent := "PORT=9090\nDATABASE_USER=admin\n"
	if err := os.WriteFile(".test_bind.env", []byte(dotenvContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(".test_bind.env")

	yamlContent := "HOST: example.com\nDEBUG: true\nDATABASE:\n  PASS: secret\n"
	if err := os.WriteFile("test_bind.yaml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_bind.yaml")

	var cfg testBindConfig
	if err := wcli.Load(&cfg,
		wcli.WithDotEnv(".test_bind.env"),
		wcli.WithFiles("test_bind.yaml"),
	); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port: 예상 9090, 실제 %d", cfg.Port)
	}
	if cfg.Host != "example.com" {
		t.Errorf("Host: 예상 example.com, 실제 %s", cfg.Host)
	}
	if !cfg.Debug {
		t.Errorf("Debug: 예상 true, 실제 %v", cfg.Debug)
	}
	if cfg.Timeout != 30.5 {
		t.Errorf("Timeout: 예상 30.5, 실제 %f", cfg.Timeout)
	}
	if cfg.DB.User != "admin" {
		t.Errorf("DB.User: 예상 admin, 실제 %s", cfg.DB.User)
	}
	if cfg.DB.Pass != "secret" {
		t.Errorf("DB.Pass: 예상 secret, 실제 %s", cfg.DB.Pass)
	}
}

func TestLoadTOML(t *testing.T) {
	tomlContent := "PORT = 8888\n[DATABASE]\nUSER = \"root\"\n"
	if err := os.WriteFile("test_bind.toml", []byte(tomlContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_bind.toml")

	var cfg testBindConfig
	if err := wcli.Load(&cfg, wcli.WithFiles("test_bind.toml")); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	if cfg.Port != 8888 {
		t.Errorf("Port: 예상 8888, 실제 %d", cfg.Port)
	}
	if cfg.DB.User != "root" {
		t.Errorf("DB.User: 예상 root, 실제 %s", cfg.DB.User)
	}
}

func TestLoadEnvWithPrefix(t *testing.T) {
	os.Setenv("WCLI_TEST_PORT", "1234")
	os.Setenv("WCLI_TEST_DATABASE_USER", "env_user")
	defer os.Unsetenv("WCLI_TEST_PORT")
	defer os.Unsetenv("WCLI_TEST_DATABASE_USER")

	type envCfg struct {
		Port int `wcli:"PORT"`
		DB   struct {
			User string `wcli:"USER"`
		} `wcli:"DATABASE"`
	}

	var cfg envCfg
	if err := wcli.Load(&cfg, wcli.WithEnv(), wcli.WithPrefix("WCLI_TEST")); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	if cfg.Port != 1234 {
		t.Errorf("Port: 예상 1234, 실제 %d", cfg.Port)
	}
	if cfg.DB.User != "env_user" {
		t.Errorf("DB.User: 예상 env_user, 실제 %s", cfg.DB.User)
	}
}

func TestWriteDefault(t *testing.T) {
	type defaultCfg struct {
		Name string `wcli:"NAME" default:"myapp"`
		Port int    `wcli:"PORT" default:"9000"`
	}

	tmpFile, err := os.CreateTemp("", "defaults-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	var cfg defaultCfg
	if err := wcli.WriteDefault(&cfg, tmpFile.Name()); err != nil {
		t.Fatalf("WriteDefault 실패: %v", err)
	}

	// 파일이 생성됐는지 확인
	info, err := os.Stat(tmpFile.Name())
	if err != nil || info.Size() == 0 {
		t.Errorf("WriteDefault로 생성된 파일이 비어있음")
	}
}
