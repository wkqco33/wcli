package wcli_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/config"
)

func TestFlagBindEnv(t *testing.T) {
	var token string
	fs := wcli.NewFlagSet()
	fs.StringVar(&token, "token", "t", "", "API Token")

	err := fs.BindEnv("token", "WCLI_TEST_TOKEN")
	if err != nil {
		t.Fatalf("BindEnv 실패: %v", err)
	}

	// 1. 환경 변수가 없을 때는 빈 문자열 유지
	os.Unsetenv("WCLI_TEST_TOKEN")
	_, err = fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	if err := fs.Validate(); err != nil {
		t.Fatalf("Validate 실패: %v", err)
	}
	if token != "" {
		t.Errorf("token이 빈 값이어야 함, 실제: %q", token)
	}

	// 2. 환경 변수가 설정되면 자동으로 주입 확인
	os.Setenv("WCLI_TEST_TOKEN", "EnvValue123")
	defer os.Unsetenv("WCLI_TEST_TOKEN")

	// Parse 수행 후 Validate 시 환경변수 바인딩이 일어남
	fs = wcli.NewFlagSet()
	fs.StringVar(&token, "token", "t", "", "API Token")
	_ = fs.BindEnv("token", "WCLI_TEST_TOKEN")

	_, err = fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	if err := fs.Validate(); err != nil {
		t.Fatalf("Validate 실패: %v", err)
	}
	if token != "EnvValue123" {
		t.Errorf("token에 환경 변수 값이 바인딩되지 않음, 실제: %q", token)
	}
}

func TestFlagExclusive(t *testing.T) {
	var jsonOut bool
	var yamlOut bool

	fs := wcli.NewFlagSet()
	fs.BoolVar(&jsonOut, "json", "j", false, "JSON output")
	fs.BoolVar(&yamlOut, "yaml", "y", false, "YAML output")
	fs.MarkFlagsMutuallyExclusive("json", "yaml")

	// 1. 하나만 설정되었을 때는 성공
	_, err := fs.Parse([]string{"--json"})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	if err := fs.Validate(); err != nil {
		t.Errorf("Validate 실패하면 안 됨: %v", err)
	}

	// 2. 둘 다 설정되었을 때는 Validate 에러 발생
	fs = wcli.NewFlagSet()
	fs.BoolVar(&jsonOut, "json", "j", false, "JSON output")
	fs.BoolVar(&yamlOut, "yaml", "y", false, "YAML output")
	fs.MarkFlagsMutuallyExclusive("json", "yaml")

	_, err = fs.Parse([]string{"--json", "--yaml"})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	err = fs.Validate()
	if err == nil {
		t.Error("상호 배제 검증 에러가 발생해야 함")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("에러 메시지에 'mutually exclusive'가 포함되어야 함, 실제: %v", err)
	}
}

func TestFlagRequiredTogether(t *testing.T) {
	var user string
	var password string

	fs := wcli.NewFlagSet()
	fs.StringVar(&user, "user", "u", "", "Username")
	fs.StringVar(&password, "password", "p", "", "Password")
	fs.MarkFlagsRequiredTogether("user", "password")

	// 1. 둘 다 설정 안 됨 -> 정상
	_, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	if err := fs.Validate(); err != nil {
		t.Errorf("Validate 실패하면 안 됨: %v", err)
	}

	// 2. 하나만 설정됨 -> 에러
	fs = wcli.NewFlagSet()
	fs.StringVar(&user, "user", "u", "", "Username")
	fs.StringVar(&password, "password", "p", "", "Password")
	fs.MarkFlagsRequiredTogether("user", "password")

	_, err = fs.Parse([]string{"--user", "alice"})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	err = fs.Validate()
	if err == nil {
		t.Error("동반 지정 제약조건으로 인해 검증 에러가 발생해야 함")
	}

	// 3. 둘 다 설정됨 -> 정상
	fs = wcli.NewFlagSet()
	fs.StringVar(&user, "user", "u", "", "Username")
	fs.StringVar(&password, "password", "p", "", "Password")
	fs.MarkFlagsRequiredTogether("user", "password")

	_, err = fs.Parse([]string{"--user", "alice", "--password", "secret"})
	if err != nil {
		t.Fatalf("Parse 실패: %v", err)
	}
	if err := fs.Validate(); err != nil {
		t.Errorf("Validate 실패하면 안 됨: %v", err)
	}
}

// TestExclusiveViaExecute 그룹 제약이 Execute() 경로(combined FlagSet)에서도
// 동작하는지 검증한다. merge()가 그룹 정보를 복사하지 않으면 제약이 무력화된다.
func TestExclusiveViaExecute(t *testing.T) {
	var jsonOut, yamlOut bool
	cmd := &wcli.Command{
		Use:           "app",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	cmd.Flags().BoolVar(&jsonOut, "json", "j", false, "JSON output")
	cmd.Flags().BoolVar(&yamlOut, "yaml", "y", false, "YAML output")
	cmd.Flags().MarkFlagsMutuallyExclusive("json", "yaml")

	// 하나만 설정 -> 정상
	if err := cmd.Execute([]string{"--json"}); err != nil {
		t.Errorf("단일 플래그는 통과해야 함: %v", err)
	}
	// 둘 다 설정 -> 상호 배제 에러
	err := cmd.Execute([]string{"--json", "--yaml"})
	if err == nil {
		t.Fatal("상호 배제 검증 에러가 발생해야 함")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("에러 메시지에 'mutually exclusive' 포함 기대, 실제: %v", err)
	}
}

// TestRequiredTogetherViaPersistent persistent 플래그에 건 동반 제약이
// 하위 커맨드 Execute 경로에서도 동작하는지 검증한다.
func TestRequiredTogetherViaPersistent(t *testing.T) {
	var user, password string
	root := &wcli.Command{Use: "app"}
	root.PersistentFlags().StringVar(&user, "user", "u", "", "Username")
	root.PersistentFlags().StringVar(&password, "password", "p", "", "Password")
	root.PersistentFlags().MarkFlagsRequiredTogether("user", "password")

	sub := &wcli.Command{
		Use:           "run",
		SilenceErrors: true,
		Run:           func(ctx *wcli.Context) error { return nil },
	}
	root.AddCommand(sub)

	// 하나만 설정 -> 에러
	err := root.Execute([]string{"run", "--user", "alice"})
	if err == nil {
		t.Fatal("동반 지정 제약으로 검증 에러가 발생해야 함")
	}
	// 둘 다 설정 -> 정상
	if err := root.Execute([]string{"run", "--user", "alice", "--password", "secret"}); err != nil {
		t.Errorf("둘 다 설정 시 통과해야 함: %v", err)
	}
}

func TestFlagPriorityChain(t *testing.T) {
	// 임시 JSON 설정 파일 준비
	jsonContent := `{"server": {"host": "config-host"}}`
	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.Write([]byte(jsonContent))
	tmpFile.Close()

	config.SetConfigFile(tmpFile.Name())
	config.SetConfigType("json")
	_ = config.ReadInConfig()

	// 1. 설정파일 매핑 작동 검증 (플래그, 환경변수 없을 시)
	var host string
	fs := wcli.NewFlagSet()
	fs.StringVar(&host, "host", "h", "default-host", "Server Host")
	_ = fs.BindEnv("host", "WCLI_TEST_HOST")
	_ = fs.BindConfig("host", "server.host")

	os.Unsetenv("WCLI_TEST_HOST")
	_, _ = fs.Parse([]string{})
	_ = fs.Validate()
	if host != "config-host" {
		t.Errorf("설정파일 값 매핑 실패: 예상 config-host, 실제 %q", host)
	}

	// 2. 환경변수 우선순위 검증 (설정파일이 있어도 환경변수가 덮어씀)
	os.Setenv("WCLI_TEST_HOST", "env-host")
	defer os.Unsetenv("WCLI_TEST_HOST")

	fs = wcli.NewFlagSet()
	fs.StringVar(&host, "host", "h", "default-host", "Server Host")
	_ = fs.BindEnv("host", "WCLI_TEST_HOST")
	_ = fs.BindConfig("host", "server.host")

	_, _ = fs.Parse([]string{})
	_ = fs.Validate()
	if host != "env-host" {
		t.Errorf("환경변수 우선순위 적용 실패: 예상 env-host, 실제 %q", host)
	}

	// 3. CLI 플래그 최우선순위 검증 (설정파일, 환경변수가 있어도 CLI 값이 덮어씀)
	fs = wcli.NewFlagSet()
	fs.StringVar(&host, "host", "h", "default-host", "Server Host")
	_ = fs.BindEnv("host", "WCLI_TEST_HOST")
	_ = fs.BindConfig("host", "server.host")

	_, _ = fs.Parse([]string{"--host", "cli-host"})
	_ = fs.Validate()
	if host != "cli-host" {
		t.Errorf("CLI 플래그 최우선 적용 실패: 예상 cli-host, 실제 %q", host)
	}
}

func TestBindConfigStringSliceAcrossFormats(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		content  string
		expected []string
	}{
		{
			name:     "json",
			ext:      ".json",
			content:  `{"server":{"tags":["blue","green"]}}`,
			expected: []string{"blue", "green"},
		},
		{
			name: "yaml",
			ext:  ".yaml",
			content: "server:\n" +
				"  tags:\n" +
				"    - blue\n" +
				"    - green\n",
			expected: []string{"blue", "green"},
		},
		{
			name:     "toml",
			ext:      ".toml",
			content:  "[server]\ntags = [\"blue\", \"green\"]\n",
			expected: []string{"blue", "green"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Reset()
			tmpFile, err := os.CreateTemp("", "flag-config-*"+tt.ext)
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

			tags := []string{"default"}
			fs := wcli.NewFlagSet()
			fs.StringSliceVar(&tags, "tag", "", []string{"default"}, "태그")
			if err := fs.BindConfig("tag", "server.tags"); err != nil {
				t.Fatalf("BindConfig 실패: %v", err)
			}

			if _, err := fs.Parse(nil); err != nil {
				t.Fatalf("Parse 실패: %v", err)
			}
			if err := fs.Validate(); err != nil {
				t.Fatalf("Validate 실패: %v", err)
			}
			if !reflect.DeepEqual(tags, tt.expected) {
				t.Fatalf("BindConfig 배열 바인딩 실패: got=%v want=%v", tags, tt.expected)
			}
		})
	}
}
