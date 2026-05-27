package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type configStore struct {
	mu         sync.RWMutex
	configPath string
	configType string
	data       map[string]any
}

var globalConfig = &configStore{
	data: make(map[string]any),
}

// SetConfigFile 설정 파일 경로를 지정합니다.
func SetConfigFile(path string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.configPath = path

	// 확장자를 바탕으로 타입 추정
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		globalConfig.configType = "json"
	case ".ini", ".cfg", ".conf":
		globalConfig.configType = "ini"
	case ".yaml", ".yml":
		globalConfig.configType = "yaml"
	case ".toml":
		globalConfig.configType = "toml"
	case ".env":
		globalConfig.configType = "env"
	}
}

// SetConfigType 설정 파일 형식을 명시적으로 지정합니다 ("json", "ini", "yaml", "toml", "env").
func SetConfigType(inType string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.configType = strings.ToLower(inType)
}

// ReadInConfig 설정 파일을 읽어 메모리에 로드합니다.
func ReadInConfig() error {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()

	if globalConfig.configPath == "" {
		return fmt.Errorf("config file path is not set")
	}

	content, err := os.ReadFile(globalConfig.configPath)
	if err != nil {
		return fmt.Errorf("read config file error: %w", err)
	}

	switch globalConfig.configType {
	case "json":
		var raw map[string]any
		if err := json.Unmarshal(content, &raw); err != nil {
			return fmt.Errorf("parse json config error: %w", err)
		}
		globalConfig.data = raw
	case "ini":
		parsed, err := parseINIContent(string(content))
		if err != nil {
			return fmt.Errorf("parse ini config error: %w", err)
		}
		globalConfig.data = parsed
	case "yaml", "yml":
		parsed, err := parseYAMLContent(string(content))
		if err != nil {
			return fmt.Errorf("parse yaml config error: %w", err)
		}
		globalConfig.data = parsed
	case "toml":
		parsed, err := parseTOMLContent(string(content))
		if err != nil {
			return fmt.Errorf("parse toml config error: %w", err)
		}
		globalConfig.data = parsed
	case "env":
		parsed, err := parseDotEnvContent(string(content))
		if err != nil {
			return fmt.Errorf("parse env config error: %w", err)
		}
		globalConfig.data = parsed
	default:
		return fmt.Errorf("unsupported config type: %q", globalConfig.configType)
	}

	return nil
}

// Get 설정 맵에서 계층형 점 표기법(예: "database.port")으로 값을 획득합니다.
func Get(key string) any {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return getNestedVal(globalConfig.data, key)
}

// Set 설정 맵에 계층형 점 표기법으로 값을 설정합니다.
func Set(key string, value any) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	setNestedMap(globalConfig.data, key, value)
}

// SetDefault 해당 키가 없을 때만 기본값을 설정합니다.
func SetDefault(key string, value any) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	if getNestedVal(globalConfig.data, key) == nil {
		setNestedMap(globalConfig.data, key, value)
	}
}

func getNestedVal(data map[string]any, key string) any {
	parts := strings.Split(key, ".")
	var current any = data
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		val, exists := m[part]
		if !exists {
			return nil
		}
		current = val
	}
	return current
}

func setNestedMap(m map[string]any, key string, val any) {
	parts := strings.Split(key, ".")
	current := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, exists := current[part]
		if !exists {
			nextMap := make(map[string]any)
			current[part] = nextMap
			current = nextMap
		} else {
			nextMap, ok := next.(map[string]any)
			if !ok {
				nextMap = make(map[string]any)
				current[part] = nextMap
				current = nextMap
			} else {
				current = nextMap
			}
		}
	}
	current[parts[len(parts)-1]] = val
}

func parseINIContent(content string) (map[string]any, error) {
	data := make(map[string]any)
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			val = val[1 : len(val)-1]
		}

		fullKey := key
		if currentSection != "" {
			fullKey = currentSection + "." + key
		}

		setNestedMap(data, fullKey, val)
	}

	return data, scanner.Err()
}

func parseYAMLContent(content string) (map[string]any, error) {
	data := make(map[string]any)
	scanner := bufio.NewScanner(strings.NewReader(content))

	type stackItem struct {
		indent int
		m      map[string]any
	}
	stack := []stackItem{{indent: -1, m: data}}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		for len(stack) > 1 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}

		pair := strings.SplitN(trimmed, ":", 2)
		if len(pair) < 1 {
			continue
		}

		key := strings.TrimSpace(pair[0])
		if len(pair) == 2 && strings.TrimSpace(pair[1]) != "" {
			val := strings.TrimSpace(pair[1])
			val = strings.Trim(val, `"'`)
			stack[len(stack)-1].m[key] = val
		} else {
			newMap := make(map[string]any)
			stack[len(stack)-1].m[key] = newMap
			stack = append(stack, stackItem{indent: indent, m: newMap})
		}
	}

	return data, scanner.Err()
}

func parseTOMLContent(content string) (map[string]any, error) {
	data := make(map[string]any)
	currentMap := data
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// [section] 또는 [section.subsection] 처리
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.Trim(line, "[]")
			currentMap = getOrCreateNestedMap(data, section)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			currentMap[key] = val
		}
	}

	return data, scanner.Err()
}

// getOrCreateNestedMap 점 표기법 경로에서 중첩 맵을 가져오거나 생성합니다.
func getOrCreateNestedMap(root map[string]any, path string) map[string]any {
	parts := strings.Split(path, ".")
	current := root
	for _, part := range parts {
		if m, ok := current[part].(map[string]any); ok {
			current = m
		} else {
			m := make(map[string]any)
			current[part] = m
			current = m
		}
	}
	return current
}

func parseDotEnvContent(content string) (map[string]any, error) {
	data := make(map[string]any)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			data[key] = val
		}
	}

	return data, scanner.Err()
}

// configExtensions 지원 설정 파일 확장자 목록 (우선순위 순서)
var configExtensions = []string{".yaml", ".yml", ".toml", ".ini", ".json", ".env"}

// AutoDiscoverConfig appName 기반으로 표준 경로를 순서대로 탐색해 설정 파일을 로드합니다.
// 탐색 순서: extraPaths → ./config.* → ~/.appname.* → /etc/appname/config.*
// 첫 번째로 존재하는 파일을 로드하며, 아무 파일도 없으면 에러를 반환합니다.
func AutoDiscoverConfig(appName string, extraPaths ...string) error {
	candidates := make([]string, 0, len(extraPaths)+len(configExtensions)*2)

	// 1. 사용자 지정 경로
	candidates = append(candidates, extraPaths...)

	// 2. ./config.*
	for _, ext := range configExtensions {
		candidates = append(candidates, "config"+ext)
	}

	// 3. ~/.appname.*
	if home, err := os.UserHomeDir(); err == nil {
		for _, ext := range configExtensions {
			candidates = append(candidates, filepath.Join(home, "."+appName+ext))
		}
	}

	// 4. /etc/appname/config.*
	for _, ext := range configExtensions {
		candidates = append(candidates, filepath.Join("/etc", appName, "config"+ext))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			SetConfigFile(path)
			return ReadInConfig()
		}
	}

	return fmt.Errorf("설정 파일을 찾을 수 없습니다 (앱명: %s)", appName)
}
