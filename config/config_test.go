package config_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/seoyc/wcli/config"
)

// --- 글로벌 config store 테스트 (SetConfigFile / ReadInConfig / Get) ---

func TestJSONConfig(t *testing.T) {
	resetConfigState(t)
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
	resetConfigState(t)
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
	resetConfigState(t)
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
	resetConfigState(t)
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
	resetConfigState(t)
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
	resetConfigState(t)
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

func TestConfigArraySupportAcrossFormats(t *testing.T) {
	tests := []struct {
		name         string
		ext          string
		content      string
		expectedIPs  []string
		expectedTags []string
		expectedNone []string
	}{
		{
			name:         "json",
			ext:          ".json",
			content:      `{"ips":["127.0.0.1","10.0.0.1"],"tags":["prod"],"empty":[]}`,
			expectedIPs:  []string{"127.0.0.1", "10.0.0.1"},
			expectedTags: []string{"prod"},
			expectedNone: []string{},
		},
		{
			name: "yaml",
			ext:  ".yaml",
			content: "ips:\n" +
				"  - 127.0.0.1\n" +
				"  - 10.0.0.1\n" +
				"tags: [prod]\n" +
				"empty: []\n",
			expectedIPs:  []string{"127.0.0.1", "10.0.0.1"},
			expectedTags: []string{"prod"},
			expectedNone: []string{},
		},
		{
			name:         "toml",
			ext:          ".toml",
			content:      "ips = [\"127.0.0.1\", \"10.0.0.1\"]\ntags = [\"prod\"]\nempty = []\n",
			expectedIPs:  []string{"127.0.0.1", "10.0.0.1"},
			expectedTags: []string{"prod"},
			expectedNone: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigState(t)
			tmpFile, err := os.CreateTemp("", "config-array-*"+tt.ext)
			if err != nil {
				t.Fatalf("임시 파일 생성 실패: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("임시 파일 쓰기 실패: %v", err)
			}
			tmpFile.Close()

			config.SetConfigFile(tmpFile.Name())
			if tt.ext == ".json" {
				config.SetConfigType("json")
			}
			if err := config.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig 실패: %v", err)
			}

			if got := config.GetStringSlice("ips"); !reflect.DeepEqual(got, tt.expectedIPs) {
				t.Fatalf("ips 배열 파싱 실패: got=%v want=%v", got, tt.expectedIPs)
			}
			if got := config.GetStringSlice("tags"); !reflect.DeepEqual(got, tt.expectedTags) {
				t.Fatalf("tags 단일 배열 파싱 실패: got=%v want=%v", got, tt.expectedTags)
			}
			if got := config.GetStringSlice("empty"); !reflect.DeepEqual(got, tt.expectedNone) {
				t.Fatalf("empty 배열 파싱 실패: got=%v want=%v", got, tt.expectedNone)
			}
		})
	}
}

func TestAutoDiscoverConfig_Found(t *testing.T) {
	resetConfigState(t)
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
	resetConfigState(t)
	err := config.AutoDiscoverConfig("nonexistent_app_xyz123")
	if err == nil {
		t.Error("파일이 없을 때 에러 기대")
	}
}

func TestTypedGetHelpers(t *testing.T) {
	resetConfigState(t)
	config.Set("test.str", "hello")
	config.Set("test.int_str", "123")
	config.Set("test.int_raw", 456)
	config.Set("test.bool_str", "true")
	config.Set("test.bool_raw", false)
	config.Set("test.float_str", "1.23")
	config.Set("test.duration_str", "5s")
	config.Set("test.slice_str", "a, b, c")

	if config.GetString("test.str") != "hello" {
		t.Errorf("GetString 실패: %s", config.GetString("test.str"))
	}
	if config.GetInt("test.int_str") != 123 {
		t.Errorf("GetInt(str) 실패: %d", config.GetInt("test.int_str"))
	}
	if config.GetInt("test.int_raw") != 456 {
		t.Errorf("GetInt(raw) 실패: %d", config.GetInt("test.int_raw"))
	}
	if !config.GetBool("test.bool_str") {
		t.Errorf("GetBool(str) 실패: %v", config.GetBool("test.bool_str"))
	}
	if config.GetBool("test.bool_raw") {
		t.Errorf("GetBool(raw) 실패: %v", config.GetBool("test.bool_raw"))
	}
	if config.GetFloat64("test.float_str") != 1.23 {
		t.Errorf("GetFloat64 실패: %f", config.GetFloat64("test.float_str"))
	}
	if config.GetDuration("test.duration_str") != 5*time.Second {
		t.Errorf("GetDuration 실패: %v", config.GetDuration("test.duration_str"))
	}

	slice := config.GetStringSlice("test.slice_str")
	expectedSlice := []string{"a", "b", "c"}
	if !reflect.DeepEqual(slice, expectedSlice) {
		t.Errorf("GetStringSlice 실패: %v, 기대: %v", slice, expectedSlice)
	}
}

func TestAutomaticEnv(t *testing.T) {
	resetConfigState(t)
	config.AutomaticEnv()

	// 1. 환경변수 미설정 시 기존 세팅값
	config.Set("database.port", 3306)
	if config.GetInt("database.port") != 3306 {
		t.Errorf("기본 포트 반환 실패: %d", config.GetInt("database.port"))
	}

	// 2. 환경변수 설정 시 덮어쓰기 확인
	os.Setenv("DATABASE_PORT", "5432")
	defer os.Unsetenv("DATABASE_PORT")
	if config.GetInt("database.port") != 5432 {
		t.Errorf("환경변수 덮어쓰기 실패: %d", config.GetInt("database.port"))
	}

	// 3. 접두사(Prefix) 설정 시 동작 확인
	config.SetEnvPrefix("MYAPP")
	os.Setenv("MYAPP_DATABASE_PORT", "9999")
	defer os.Unsetenv("MYAPP_DATABASE_PORT")

	if config.GetInt("database.port") != 9999 {
		t.Errorf("글로벌 접두사 적용된 환경변수 적용 실패: %d", config.GetInt("database.port"))
	}
}

func TestReloadConfig(t *testing.T) {
	resetConfigState(t)
	tmpFile, err := os.CreateTemp("", "config-reload-*.json")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 1. 초기값 쓰기 및 로드
	if _, err := tmpFile.Write([]byte(`{"key": "value1"}`)); err != nil {
		t.Fatalf("임시 파일 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	config.SetConfigFile(tmpFile.Name())
	config.SetConfigType("json")
	if err := config.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig 실패: %v", err)
	}

	if config.GetString("key") != "value1" {
		t.Errorf("초기값 로드 실패: %s", config.GetString("key"))
	}

	// 2. 파일 변경
	if err := os.WriteFile(tmpFile.Name(), []byte(`{"key": "value2"}`), 0644); err != nil {
		t.Fatalf("임시 파일 갱신 실패: %v", err)
	}

	// 3. 리로드
	if err := config.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig 실패: %v", err)
	}

	if config.GetString("key") != "value2" {
		t.Errorf("리로드 후 변경값 로드 실패: %s", config.GetString("key"))
	}
}

func TestSetDefaultDoesNotOverrideExisting(t *testing.T) {
	resetConfigState(t)

	config.Set("app.port", 8080)
	config.SetDefault("app.port", 9090)

	if got := config.GetInt("app.port"); got != 8080 {
		t.Fatalf("SetDefault는 기존 값을 덮어쓰면 안 됨: %d", got)
	}
}

func TestResetClearsGlobalState(t *testing.T) {
	resetConfigState(t)

	config.Set("app.name", "from-store")
	config.SetEnvPrefix("APP")
	config.AutomaticEnv()
	t.Setenv("APP_APP_NAME", "from-env")

	if got := config.Get("app.name"); got != "from-env" {
		t.Fatalf("Reset 전 env 우선순위가 적용되어야 함: %v", got)
	}

	config.Reset()

	if got := config.Get("app.name"); got != nil {
		t.Fatalf("Reset 후 저장된 값이 남아 있으면 안 됨: %v", got)
	}
}
