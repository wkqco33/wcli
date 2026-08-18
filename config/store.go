package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Store는 전역 상태와 분리된 인스턴스 기반 설정 저장소입니다.
type Store struct {
	state *configStore
}

// NewStore 새 설정 저장소 인스턴스를 생성합니다.
func NewStore() *Store {
	return &Store{
		state: newConfigStore(),
	}
}

// SetLookupEnv 환경변수 조회 함수를 주입합니다 (테스트 격리 지원).
func (s *Store) SetLookupEnv(fn func(string) (string, bool)) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if fn != nil {
		s.state.lookupEnvFunc = fn
	} else {
		s.state.lookupEnvFunc = os.LookupEnv
	}
}

// SetReadFileFunc 파일 읽기 함수를 주입합니다 (테스트 격리 지원).
func (s *Store) SetReadFileFunc(fn func(string) ([]byte, error)) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if fn != nil {
		s.state.readFileFunc = fn
	} else {
		s.state.readFileFunc = os.ReadFile
	}
}

// SetStatFunc 파일 상태 조회 함수를 주입합니다 (테스트 격리 지원).
func (s *Store) SetStatFunc(fn func(string) (os.FileInfo, error)) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if fn != nil {
		s.state.statFunc = fn
	} else {
		s.state.statFunc = os.Stat
	}
}

// SetUserHomeDirFunc 홈 디렉터리 조회 함수를 주입합니다 (테스트 격리 지원).
func (s *Store) SetUserHomeDirFunc(fn func() (string, error)) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if fn != nil {
		s.state.userHomeDirFunc = fn
	} else {
		s.state.userHomeDirFunc = os.UserHomeDir
	}
}

func (s *Store) getLookupEnv() func(string) (string, bool) {
	if s.state.lookupEnvFunc != nil {
		return s.state.lookupEnvFunc
	}
	return os.LookupEnv
}

func (s *Store) getReadFile() func(string) ([]byte, error) {
	if s.state.readFileFunc != nil {
		return s.state.readFileFunc
	}
	return os.ReadFile
}

func (s *Store) getStat() func(string) (os.FileInfo, error) {
	if s.state.statFunc != nil {
		return s.state.statFunc
	}
	return os.Stat
}

func (s *Store) getUserHomeDir() func() (string, error) {
	if s.state.userHomeDirFunc != nil {
		return s.state.userHomeDirFunc
	}
	return os.UserHomeDir
}

func (s *Store) SetConfigFile(path string) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.configPath = path

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		s.state.configType = "json"
	case ".ini", ".cfg", ".conf":
		s.state.configType = "ini"
	case ".yaml", ".yml":
		s.state.configType = "yaml"
	case ".toml":
		s.state.configType = "toml"
	case ".env":
		s.state.configType = "env"
	}
}

func (s *Store) SetConfigType(inType string) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.configType = strings.ToLower(inType)
}

func (s *Store) SetEnvPrefix(prefix string) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.envPrefix = strings.ToUpper(prefix)
}

func (s *Store) AutomaticEnv() {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.autoEnv = true
}

func (s *Store) SetStrictParsing(strict bool) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.strict = strict
}

func (s *Store) Reset() {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	s.state.configPath = ""
	s.state.configType = ""
	s.state.data = make(map[string]any)
	s.state.autoEnv = false
	s.state.envPrefix = ""
	s.state.strict = false
}

func (s *Store) ReadInConfig() error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if s.state.configPath == "" {
		return fmt.Errorf("config file path is not set")
	}

	readFile := s.getReadFile()
	content, err := readFile(s.state.configPath)
	if err != nil {
		return fmt.Errorf("read config file error: %w", err)
	}

	var parsed map[string]any
	switch s.state.configType {
	case "json":
		if err := json.Unmarshal(content, &parsed); err != nil {
			return fmt.Errorf("parse json config error: %w", err)
		}
	case "ini":
		parsed, err = parseINIContent(string(content), s.state.strict)
	case "yaml", "yml":
		parsed, err = parseYAMLContent(string(content), s.state.strict)
	case "toml":
		parsed, err = parseTOMLContent(string(content), s.state.strict)
	case "env":
		parsed, err = parseDotEnvContent(string(content))
	default:
		return fmt.Errorf("unsupported config type: %q", s.state.configType)
	}
	if err != nil {
		return err
	}

	s.state.data = NormalizeKeys(parsed).(map[string]any)
	return nil
}

func (s *Store) ReloadConfig() error {
	return s.ReadInConfig()
}

func (s *Store) AutoDiscoverConfig(appName string, extraPaths ...string) error {
	candidates := make([]string, 0, len(extraPaths)+len(configExtensions)*2)

	// 1. 사용자 지정 경로
	candidates = append(candidates, extraPaths...)

	// 2. ./config.*
	for _, ext := range configExtensions {
		candidates = append(candidates, "config"+ext)
	}

	// 3. ~/.appname.*
	userHome := s.getUserHomeDir()
	if home, err := userHome(); err == nil {
		for _, ext := range configExtensions {
			candidates = append(candidates, filepath.Join(home, "."+appName+ext))
		}
	}

	// 4. /etc/appname/config.*
	for _, ext := range configExtensions {
		candidates = append(candidates, filepath.Join("/etc", appName, "config"+ext))
	}

	stat := s.getStat()
	for _, path := range candidates {
		if _, err := stat(path); err == nil {
			s.SetConfigFile(path)
			return s.ReadInConfig()
		}
	}

	return fmt.Errorf("config file not found (app: %s)", appName)
}

func (s *Store) Get(key string) any {
	s.state.mu.RLock()
	defer s.state.mu.RUnlock()

	if s.state.autoEnv {
		envKey := strings.ReplaceAll(strings.ToUpper(key), ".", "_")
		if s.state.envPrefix != "" {
			envKey = s.state.envPrefix + "_" + envKey
		}
		lookup := s.getLookupEnv()
		if envVal, exists := lookup(envKey); exists {
			return envVal
		}
	}
	return getNestedVal(s.state.data, key)
}

func (s *Store) Set(key string, value any) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	setNestedMap(s.state.data, key, value)
}

func (s *Store) SetDefault(key string, value any) {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if getNestedVal(s.state.data, key) == nil {
		setNestedMap(s.state.data, key, value)
	}
}

func (s *Store) GetString(key string) string {
	val := s.Get(key)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func (s *Store) GetInt(key string) int {
	val := s.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

func (s *Store) GetBool(key string) bool {
	val := s.Get(key)
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return false
}

func (s *Store) GetFloat64(key string) float64 {
	val := s.Get(key)
	if val == nil {
		return 0.0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

func (s *Store) GetDuration(key string) time.Duration {
	val := s.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case time.Duration:
		return v
	case int:
		return time.Duration(v)
	case int64:
		return time.Duration(v)
	case float64:
		return time.Duration(v)
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 0
}

func (s *Store) GetStringSlice(key string) []string {
	val := s.Get(key)
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		res := make([]string, len(v))
		for i, item := range v {
			res[i] = fmt.Sprintf("%v", item)
		}
		return res
	case string:
		parts := strings.Split(v, ",")
		res := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				res = append(res, trimmed)
			}
		}
		return res
	}
	return []string{fmt.Sprintf("%v", val)}
}
