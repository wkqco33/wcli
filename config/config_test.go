package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/seoyc/wcli/config"
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

	config.SetConfigFile(tmpFile.Name())
	config.SetConfigType("json")

	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	dbVal := config.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	appName := config.Get("app.name")
	if appName != "testapp" {
		t.Errorf("app.name 값 불일치: 예상 testapp, 실제 %v", appName)
	}

	portVal := config.Get("app.port")
	if portVal != 8080.0 && portVal != float64(8080) {
		t.Errorf("app.port 값 불일치: 예상 8080.0, 실제 %v (타입: %v)", portVal, reflect.TypeOf(portVal))
	}

	debugVal := config.Get("app.debug")
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

	config.SetConfigFile(tmpFile.Name())
	config.SetConfigType("ini")

	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	dbVal := config.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	hostVal := config.Get("server.host")
	if hostVal != "localhost" {
		t.Errorf("server.host 값 불일치: 예상 localhost, 실제 %v", hostVal)
	}

	portVal := config.Get("server.port")
	if portVal != "9000" {
		t.Errorf("server.port 값 불일치: 예상 9000, 실제 %v", portVal)
	}

	sslVal := config.Get("server.ssl")
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

	config.SetConfigFile(tmpFile.Name())
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if config.Get("database") != "postgresql" {
		t.Errorf("database 값 불일치: %v", config.Get("database"))
	}
	if config.Get("app.name") != "testapp" {
		t.Errorf("app.name 값 불일치: %v", config.Get("app.name"))
	}
	if config.Get("app.port") != "8080" {
		t.Errorf("app.port 값 불일치: %v", config.Get("app.port"))
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

	config.SetConfigFile(tmpFile.Name())
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if config.Get("database") != "postgresql" {
		t.Errorf("database 값 불일치: %v", config.Get("database"))
	}
	if config.Get("app.name") != "testapp" {
		t.Errorf("app.name 값 불일치: %v", config.Get("app.name"))
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

	config.SetConfigFile(tmpFile.Name())
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if config.Get("DATABASE") != "postgresql" {
		t.Errorf("DATABASE 값 불일치: %v", config.Get("DATABASE"))
	}
	if config.Get("APP_NAME") != "testapp" {
		t.Errorf("APP_NAME 값 불일치: %v", config.Get("APP_NAME"))
	}
}

func TestTOMLNestedSection(t *testing.T) {
	tomlContent := `
[server]
host = "localhost"
port = "8080"

[server.tls]
cert = "/etc/ssl/cert.pem"
key = "/etc/ssl/key.pem"
`
	tmpFile, err := os.CreateTemp("", "config-nested-*.toml")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(tomlContent)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	config.SetConfigFile(tmpFile.Name())
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if config.Get("server.host") != "localhost" {
		t.Errorf("server.host 불일치: %v", config.Get("server.host"))
	}
	if config.Get("server.tls.cert") != "/etc/ssl/cert.pem" {
		t.Errorf("server.tls.cert 불일치: %v", config.Get("server.tls.cert"))
	}
	if config.Get("server.tls.key") != "/etc/ssl/key.pem" {
		t.Errorf("server.tls.key 불일치: %v", config.Get("server.tls.key"))
	}
}

func TestAutoDiscoverConfig_Found(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.json"
	os.WriteFile(cfgPath, []byte(`{"key":"value"}`), 0644)

	err := config.AutoDiscoverConfig("testapp", cfgPath)
	if err != nil {
		t.Fatalf("AutoDiscoverConfig 오류: %v", err)
	}
	if config.Get("key") != "value" {
		t.Errorf("'key' 값 기대: value, 실제: %v", config.Get("key"))
	}
}

func TestAutoDiscoverConfig_NotFound(t *testing.T) {
	err := config.AutoDiscoverConfig("nonexistent_app_xyz123")
	if err == nil {
		t.Error("파일이 없을 때 에러 기대")
	}
}
