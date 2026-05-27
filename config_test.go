package wcli_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/seoyc/wcli"
)

func TestJSONConfig(t *testing.T) {
	jsonContent := `{
		"app": {
			"name": "testapp",
			"port": 8080,
			"debug": true
		},
		"database": "postgresql"
	}`

	tmpFile, err := ioutil.TempFile("", "config-*.json")
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

	// 1. 단순 값 검증
	dbVal := wcli.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	// 2. 중첩 맵 값 검증
	appName := wcli.Get("app.name")
	if appName != "testapp" {
		t.Errorf("app.name 값 불일치: 예상 testapp, 실제 %v", appName)
	}

	// float64(json 파싱 기본 숫자형) 검증
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

	tmpFile, err := ioutil.TempFile("", "config-*.ini")
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

	// 1. 글로벌 키 검증
	dbVal := wcli.Get("database")
	if dbVal != "postgresql" {
		t.Errorf("database 값 불일치: 예상 postgresql, 실제 %v", dbVal)
	}

	// 2. 섹션 키 검증
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
		t.Errorf("server.ssl 값 불일치 (쌍따옴표 제거 확인): 예상 false, 실제 %v", sslVal)
	}
}
