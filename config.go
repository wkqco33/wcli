package wcli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type configStore struct {
	mu         sync.RWMutex
	configPath string
	configType string // "json" or "ini"
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
	ext := strings.ToLower(filepathExt(path))
	if ext == ".json" {
		globalConfig.configType = "json"
	} else if ext == ".ini" || ext == ".cfg" || ext == ".conf" {
		globalConfig.configType = "ini"
	}
}

// SetConfigType 설정 파일 형식을 명시적으로 지정합니다 ("json", "ini").
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

	content, err := ioutil.ReadFile(globalConfig.configPath)
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
	default:
		return fmt.Errorf("unsupported config type: %q", globalConfig.configType)
	}

	return nil
}

// Get 설정 맵에서 계층형 점 표기법(예: "database.port")으로 값을 획득합니다.
func Get(key string) any {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()

	parts := strings.Split(key, ".")
	var current any = globalConfig.data

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

func parseINIContent(content string) (map[string]any, error) {
	data := make(map[string]any)
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 빈 줄 또는 주석 건너뜀
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 섹션 헤더 처리
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		// 키-밸류 추출
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 잘못된 INI 형식 라인은 건너뜀
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// 쌍따옴표 제거
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

func filepathExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}
