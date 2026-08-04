package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type configStore struct {
	mu         sync.RWMutex
	configPath string
	configType string
	data       map[string]any
	autoEnv    bool
	envPrefix  string
}

var globalConfig = &configStore{
	data: make(map[string]any),
}

var (
	lookupEnvFunc   = os.LookupEnv
	userHomeDirFunc = os.UserHomeDir
	readFileFunc    = os.ReadFile
	statFunc        = os.Stat
)

// Reset 전역 설정 상태를 초기화합니다.
// 테스트 격리와 재초기화가 필요한 장기 실행 프로세스에서 사용할 수 있습니다.
func Reset() {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.configPath = ""
	globalConfig.configType = ""
	globalConfig.data = make(map[string]any)
	globalConfig.autoEnv = false
	globalConfig.envPrefix = ""
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

// SetEnvPrefix 환경변수 연동 시 사용할 글로벌 접두사를 설정합니다.
func SetEnvPrefix(prefix string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.envPrefix = strings.ToUpper(prefix)
}

// AutomaticEnv 설정 키 조회 시 대응하는 환경변수가 존재하는 경우, 환경변수 값을 최우선으로 연동하도록 활성화합니다.
func AutomaticEnv() {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.autoEnv = true
}

// ReadInConfig 설정 파일을 읽어 메모리에 로드합니다.
func ReadInConfig() error {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()

	if globalConfig.configPath == "" {
		return fmt.Errorf("config file path is not set")
	}

	content, err := readFileFunc(globalConfig.configPath)
	if err != nil {
		return fmt.Errorf("read config file error: %w", err)
	}

	var parsed map[string]any
	switch globalConfig.configType {
	case "json":
		if err := json.Unmarshal(content, &parsed); err != nil {
			return fmt.Errorf("parse json config error: %w", err)
		}
	case "ini":
		parsed, err = parseINIContent(string(content))
	case "yaml", "yml":
		parsed, err = parseYAMLContent(string(content))
	case "toml":
		parsed, err = parseTOMLContent(string(content))
	case "env":
		parsed, err = parseDotEnvContent(string(content))
	default:
		return fmt.Errorf("unsupported config type: %q", globalConfig.configType)
	}

	if err != nil {
		return err
	}

	globalConfig.data = NormalizeKeys(parsed).(map[string]any)
	return nil
}

// ReloadConfig 이미 지정된 설정 파일 경로와 파일 형식 정보를 사용하여 설정을 디스크에서 다시 읽어옵니다.
func ReloadConfig() error {
	return ReadInConfig()
}

// Get 설정 맵에서 계층형 점 표기법(예: "database.port")으로 값을 획득합니다.
func Get(key string) any {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()

	if globalConfig.autoEnv {
		envKey := strings.ReplaceAll(strings.ToUpper(key), ".", "_")
		if globalConfig.envPrefix != "" {
			envKey = globalConfig.envPrefix + "_" + envKey
		}
		if envVal, exists := lookupEnvFunc(envKey); exists {
			return envVal
		}
	}

	return getNestedVal(globalConfig.data, key)
}

// GetString 설정 키 값을 string으로 획득하며, 존재하지 않거나 변환할 수 없으면 빈 문자열을 반환합니다.
func GetString(key string) string {
	val := Get(key)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetInt 설정 키 값을 int로 획득하며, 존재하지 않거나 변환할 수 없으면 0을 반환합니다.
func GetInt(key string) int {
	val := Get(key)
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

// GetBool 설정 키 값을 bool로 획득하며, 존재하지 않거나 변환할 수 없으면 false를 반환합니다.
func GetBool(key string) bool {
	val := Get(key)
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

// GetFloat64 설정 키 값을 float64로 획득하며, 존재하지 않거나 변환할 수 없으면 0.0을 반환합니다.
func GetFloat64(key string) float64 {
	val := Get(key)
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

// GetDuration 설정 키 값을 time.Duration으로 획득하며, 존재하지 않거나 변환할 수 없으면 0을 반환합니다.
func GetDuration(key string) time.Duration {
	val := Get(key)
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

// GetStringSlice 설정 키 값을 []string으로 획득합니다.
// 값이 슬라이스인 경우 string으로 변환해 반환하고,
// 쉼표(,)로 구분된 단일 문자열인 경우 분리하여 반환합니다.
func GetStringSlice(key string) []string {
	val := Get(key)
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
		val, exists := m[strings.ToLower(part)]
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
		part := strings.ToLower(parts[i])
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
	current[strings.ToLower(parts[len(parts)-1])] = val
}

// NormalizeKeys 맵 데이터를 재귀적으로 돌며 모든 키를 소문자로 변환하여 정규화합니다.
func NormalizeKeys(val any) any {
	switch m := val.(type) {
	case map[string]any:
		res := make(map[string]any, len(m))
		for k, v := range m {
			res[strings.ToLower(k)] = NormalizeKeys(v)
		}
		return res
	case []any:
		res := make([]any, len(m))
		for i, v := range m {
			res[i] = NormalizeKeys(v)
		}
		return res
	default:
		return val
	}
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
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	type stackItem struct {
		indent int
		m      map[string]any
	}
	stack := []stackItem{{indent: -1, m: data}}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := countLeadingSpaces(line)
		for len(stack) > 1 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}

		pair := strings.SplitN(trimmed, ":", 2)
		if len(pair) != 2 {
			continue
		}

		key := strings.TrimSpace(pair[0])
		val := strings.TrimSpace(pair[1])
		if val != "" {
			parsed, err := parseConfigScalarOrArray(val)
			if err != nil {
				return nil, err
			}
			stack[len(stack)-1].m[key] = parsed
			continue
		}

		nextTrimmed, nextIndent, found := nextYAMLSignificantLine(lines, i+1)
		if found && nextIndent > indent && strings.HasPrefix(nextTrimmed, "- ") {
			items := make([]any, 0)
			for j := i + 1; j < len(lines); j++ {
				listLine := lines[j]
				listTrimmed := strings.TrimSpace(listLine)
				if listTrimmed == "" || strings.HasPrefix(listTrimmed, "#") {
					continue
				}

				listIndent := countLeadingSpaces(listLine)
				if listIndent <= indent {
					i = j - 1
					break
				}
				if !strings.HasPrefix(listTrimmed, "- ") {
					i = j - 1
					break
				}

				itemVal := strings.TrimSpace(strings.TrimPrefix(listTrimmed, "- "))
				parsed, err := parseConfigScalarOrArray(itemVal)
				if err != nil {
					return nil, err
				}
				items = append(items, parsed)
				i = j
			}
			stack[len(stack)-1].m[key] = items
			continue
		}

		newMap := make(map[string]any)
		stack[len(stack)-1].m[key] = newMap
		stack = append(stack, stackItem{indent: indent, m: newMap})
	}

	return data, nil
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
			parsed, err := parseConfigScalarOrArray(val)
			if err != nil {
				return nil, err
			}
			currentMap[key] = parsed
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

func parseConfigScalarOrArray(val string) (any, error) {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		return parseInlineArray(val)
	}
	return strings.Trim(val, `"'`), nil
}

func parseInlineArray(val string) ([]any, error) {
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(val, "["), "]"))
	if inner == "" {
		return []any{}, nil
	}

	parts := splitInlineArrayItems(inner)
	items := make([]any, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		items = append(items, strings.Trim(part, `"'`))
	}
	return items, nil
}

func splitInlineArrayItems(val string) []string {
	var (
		items   []string
		current strings.Builder
		quote   rune
	)

	for _, r := range val {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
			current.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			current.WriteRune(r)
		case r == ',':
			items = append(items, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	items = append(items, current.String())
	return items
}

func nextYAMLSignificantLine(lines []string, start int) (string, int, bool) {
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return trimmed, countLeadingSpaces(lines[i]), true
	}
	return "", 0, false
}

func countLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
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
	if home, err := userHomeDirFunc(); err == nil {
		for _, ext := range configExtensions {
			candidates = append(candidates, filepath.Join(home, "."+appName+ext))
		}
	}

	// 4. /etc/appname/config.*
	for _, ext := range configExtensions {
		candidates = append(candidates, filepath.Join("/etc", appName, "config"+ext))
	}

	for _, path := range candidates {
		if _, err := statFunc(path); err == nil {
			SetConfigFile(path)
			return ReadInConfig()
		}
	}

	return fmt.Errorf("config file not found (app: %s)", appName)
}

// OnConfigChange 설정 파일 변경 시 호출될 콜백 타입
type OnConfigChange func()

// watcherMu WatchConfig/StopWatching 동기화용 뮤텍스
var watcherMu sync.Mutex
var watcherRunning bool
var watcherStop chan struct{}

// WatchConfig 설정 파일 변경을 폴링 방식으로 감시합니다.
// 변경이 감지되면 ReadInConfig()를 다시 호출하고 callback을 실행합니다.
// interval이 0이면 기본값 2초를 사용합니다.
// 이미 실행 중인 감시가 있으면 중단 후 새로 시작합니다.
func WatchConfig(interval time.Duration, callback OnConfigChange) error {
	watcherMu.Lock()
	defer watcherMu.Unlock()

	if watcherRunning {
		close(watcherStop)
		watcherRunning = false
	}

	globalConfig.mu.RLock()
	configPath := globalConfig.configPath
	globalConfig.mu.RUnlock()

	if configPath == "" {
		return fmt.Errorf("config file path is not set")
	}

	if interval <= 0 {
		interval = 2 * time.Second
	}

	watcherStop = make(chan struct{})
	watcherRunning = true

	go func() {
		lastMod := getFileModTime(configPath)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-watcherStop:
				return
			case <-ticker.C:
				if mod := getFileModTime(configPath); mod != lastMod {
					lastMod = mod
					if err := ReloadConfig(); err == nil && callback != nil {
						callback()
					}
				}
			}
		}
	}()

	return nil
}

// StopWatching 설정 파일 감시를 중단합니다.
func StopWatching() {
	watcherMu.Lock()
	defer watcherMu.Unlock()
	if watcherRunning {
		close(watcherStop)
		watcherRunning = false
	}
}

func getFileModTime(path string) time.Time {
	info, err := statFunc(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
