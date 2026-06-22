package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/seoyc/wcli/config"
)

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
	if err := config.Load(&cfg,
		config.WithDotEnv(".test_bind.env"),
		config.WithFiles("test_bind.yaml"),
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
	if err := config.Load(&cfg, config.WithFiles("test_bind.toml")); err != nil {
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
	if err := config.Load(&cfg, config.WithEnv(), config.WithPrefix("WCLI_TEST")); err != nil {
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
	if err := config.WriteDefault(&cfg, tmpFile.Name()); err != nil {
		t.Fatalf("WriteDefault 실패: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil || info.Size() == 0 {
		t.Errorf("WriteDefault로 생성된 파일이 비어있음")
	}
}

func TestLoad_NilTarget(t *testing.T) {
	err := config.Load(nil)
	if err == nil {
		t.Fatal("nil target에서 에러가 발생해야 합니다")
	}
}

func TestLoad_NonPointer(t *testing.T) {
	type cfg struct{ Port int }
	err := config.Load(cfg{})
	if err == nil {
		t.Fatal("비포인터 target에서 에러가 발생해야 합니다")
	}
}

func TestLoadPointer(t *testing.T) {
	yamlContent := "PORT: 9090\nTEMP: 0.85\nSTOP: AI assistant\n"
	if err := os.WriteFile("test_pointer.yaml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_pointer.yaml")

	type ptrCfg struct {
		Port *int     `wcli:"PORT"`
		Temp *float64 `wcli:"TEMP"`
		Stop *string  `wcli:"STOP"`
		Seed *int     `wcli:"SEED"` // 지정하지 않은 필드는 nil이어야 함
	}

	var cfg ptrCfg
	if err := config.Load(&cfg, config.WithFiles("test_pointer.yaml")); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	if cfg.Port == nil || *cfg.Port != 9090 {
		t.Errorf("Port: 예상 9090, 실제 %v", cfg.Port)
	}
	if cfg.Temp == nil || *cfg.Temp != 0.85 {
		t.Errorf("Temp: 예상 0.85, 실제 %v", cfg.Temp)
	}
	if cfg.Stop == nil || *cfg.Stop != "AI assistant" {
		t.Errorf("Stop: 예상 'AI assistant', 실제 %v", cfg.Stop)
	}
	if cfg.Seed != nil {
		t.Errorf("Seed: 지정하지 않았으므로 nil이어야 함, 실제 %v", cfg.Seed)
	}
}

func TestLoadCaseInsensitivity(t *testing.T) {
	yamlContent := "host: case.example.com\ndebug: false\ndatabase:\n  pass: my-secret-key\n"
	if err := os.WriteFile("test_case.yaml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_case.yaml")

	var cfg testBindConfig
	if err := config.Load(&cfg, config.WithFiles("test_case.yaml")); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	if cfg.Host != "case.example.com" {
		t.Errorf("Host (소문자 파싱/대문자 바인딩): 예상 case.example.com, 실제 %s", cfg.Host)
	}
	if cfg.DB.Pass != "my-secret-key" {
		t.Errorf("DB.Pass (소문자 중첩 파싱): 예상 my-secret-key, 실제 %s", cfg.DB.Pass)
	}
}

func TestLoadSliceBinding(t *testing.T) {
	yamlContent := "ips: 127.0.0.1, 10.0.0.1, 192.168.0.1\nports: 80, 443, 8080\n"
	if err := os.WriteFile("test_slice.yaml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_slice.yaml")

	type sliceCfg struct {
		IPs   []string `wcli:"IPS"`
		Ports []int    `wcli:"PORTS"`
	}

	var cfg sliceCfg
	if err := config.Load(&cfg, config.WithFiles("test_slice.yaml")); err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	expectedIPs := []string{"127.0.0.1", "10.0.0.1", "192.168.0.1"}
	if !reflect.DeepEqual(cfg.IPs, expectedIPs) {
		t.Errorf("IPs 바인딩 실패: %v, 기대: %v", cfg.IPs, expectedIPs)
	}

	expectedPorts := []int{80, 443, 8080}
	if !reflect.DeepEqual(cfg.Ports, expectedPorts) {
		t.Errorf("Ports 바인딩 실패: %v, 기대: %v", cfg.Ports, expectedPorts)
	}
}
