package wcli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SourceType 설정 소스 유형
type SourceType int

const (
	sourceEnv SourceType = iota
	sourceDotEnv
	sourceYAML
	sourceTOML
)

type configSource struct {
	stype SourceType
	path  string
}

type configBindLoader struct {
	sources []configSource
	tagName string
	prefix  string
}

// BindOption Load() 함수에 전달하는 옵션 타입
type BindOption func(*configBindLoader)

// WithEnv 시스템 환경변수를 소스로 추가합니다.
func WithEnv() BindOption {
	return func(l *configBindLoader) {
		l.sources = append(l.sources, configSource{stype: sourceEnv})
	}
}

// WithDotEnv .env 파일을 소스로 추가합니다.
func WithDotEnv(path string) BindOption {
	return func(l *configBindLoader) {
		l.sources = append(l.sources, configSource{stype: sourceDotEnv, path: path})
	}
}

// WithFiles YAML 또는 TOML 파일들을 소스로 추가합니다.
func WithFiles(paths ...string) BindOption {
	return func(l *configBindLoader) {
		for _, path := range paths {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".yaml", ".yml":
				l.sources = append(l.sources, configSource{stype: sourceYAML, path: path})
			case ".toml":
				l.sources = append(l.sources, configSource{stype: sourceTOML, path: path})
			}
		}
	}
}

// WithTag 구조체 태그 이름을 커스텀 지정합니다 (기본값: "wcli").
func WithTag(tag string) BindOption {
	return func(l *configBindLoader) {
		l.tagName = tag
	}
}

// WithPrefix 환경변수 조회 시 사용할 접두사를 설정합니다.
func WithPrefix(prefix string) BindOption {
	return func(l *configBindLoader) {
		l.prefix = prefix
	}
}

// Load 설정 소스들로부터 데이터를 로드해 target 구조체에 바인딩합니다.
// 소스는 옵션 순서대로 병합되며, 뒤에 오는 소스가 앞의 값을 덮어씁니다.
//
// 예시:
//
//	var cfg AppConfig
//	err := wcli.Load(&cfg,
//	    wcli.WithDotEnv(".env"),
//	    wcli.WithFiles("config.yaml"),
//	    wcli.WithEnv(),
//	    wcli.WithPrefix("APP"),
//	)
func Load(target any, options ...BindOption) error {
	loader := &configBindLoader{tagName: "wcli"}
	for _, opt := range options {
		opt(loader)
	}

	merged := make(map[string]any)
	for _, src := range loader.sources {
		data, err := loadBindSource(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to load source: %w", err)
		}
		mergeConfigMap(merged, data)
	}

	return bindStruct(reflect.ValueOf(target).Elem(), merged, merged, loader.tagName, loader.prefix)
}

// WriteDefault target 구조체의 `default` 태그 값을 파일로 저장합니다.
// 파일 확장자(.env, .yaml, .toml)로 포맷을 결정합니다.
func WriteDefault(target any, path string) error {
	data, err := extractDefaults(target, "wcli")
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return saveYAML(path, data)
	case ".toml":
		return saveTOML(path, data)
	case ".env":
		return saveDotEnv(path, data)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// --- 내부 로직 ---

func loadBindSource(src configSource) (map[string]any, error) {
	switch src.stype {
	case sourceEnv:
		return loadSystemEnv()
	case sourceDotEnv:
		return loadDotEnvFile(src.path)
	case sourceYAML:
		content, err := os.ReadFile(src.path)
		if err != nil {
			return nil, err
		}
		return parseYAMLContent(string(content))
	case sourceTOML:
		content, err := os.ReadFile(src.path)
		if err != nil {
			return nil, err
		}
		return parseTOMLContent(string(content))
	default:
		return nil, fmt.Errorf("unsupported source type: %v", src.stype)
	}
}

func loadSystemEnv() (map[string]any, error) {
	data := make(map[string]any)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			data[pair[0]] = pair[1]
		}
	}
	return data, nil
}

func loadDotEnvFile(path string) (map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[string]any)
	scanner := bufio.NewScanner(file)
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

func mergeConfigMap(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				mergeConfigMap(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func bindStruct(structVal reflect.Value, data map[string]any, rootData map[string]any, tagName, prefix string) error {
	typ := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			tag = fieldType.Name
		}

		fullKey := tag
		if prefix != "" {
			fullKey = prefix + "_" + tag
		}

		var rawValue any
		var exists bool

		// 1. 중첩 구조 (파일 데이터)에서 조회
		rawValue, exists = data[tag]

		// 2. 시스템 환경변수 조회 (최우선)
		if envVal, envExists := os.LookupEnv(strings.ToUpper(fullKey)); envExists {
			rawValue = envVal
			exists = true
		}

		// 3. rootData 평탄화 조회
		if !exists {
			rawValue, exists = rootData[strings.ToUpper(fullKey)]
		}

		// 4. default 태그 fallback
		if !exists {
			if defaultVal := fieldType.Tag.Get("default"); defaultVal != "" {
				rawValue = defaultVal
				exists = true
			}
		}

		if field.Kind() == reflect.Struct {
			nestedData, ok := rawValue.(map[string]any)
			if !ok {
				nestedData = make(map[string]any)
			}
			if err := bindStruct(field, nestedData, rootData, tagName, fullKey); err != nil {
				return err
			}
			continue
		}

		if exists {
			if err := setFieldValue(field, rawValue); err != nil {
				return fmt.Errorf("field %s: %w", fieldType.Name, err)
			}
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value any) error {
	valStr := fmt.Sprintf("%v", value)

	switch field.Kind() {
	case reflect.String:
		field.SetString(valStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(valStr)
			if err == nil {
				field.SetInt(int64(d))
				return nil
			}
		}
		v, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(valStr)
		if err != nil {
			return err
		}
		field.SetBool(v)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}

func extractDefaults(target any, tagName string) (map[string]any, error) {
	val := reflect.ValueOf(target)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target must be a struct or pointer to a struct")
	}
	return extractStructDefaults(val, tagName), nil
}

func extractStructDefaults(structVal reflect.Value, tagName string) map[string]any {
	data := make(map[string]any)
	typ := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := typ.Field(i)

		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			tag = fieldType.Name
		}

		if field.Kind() == reflect.Struct {
			data[tag] = extractStructDefaults(field, tagName)
			continue
		}

		if defaultVal := fieldType.Tag.Get("default"); defaultVal != "" {
			data[tag] = defaultVal
		}
	}
	return data
}

func saveYAML(path string, data map[string]any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeYAMLMap(file, data, 0)
}

func writeYAMLMap(file *os.File, data map[string]any, indent int) error {
	spaces := strings.Repeat("  ", indent)
	for k, v := range data {
		if nested, ok := v.(map[string]any); ok {
			if _, err := file.WriteString(spaces + k + ":\n"); err != nil {
				return err
			}
			if err := writeYAMLMap(file, nested, indent+1); err != nil {
				return err
			}
		} else {
			if _, err := file.WriteString(fmt.Sprintf("%s%s: %v\n", spaces, k, v)); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveTOML(path string, data map[string]any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 최상위 키-값 먼저
	for k, v := range data {
		if _, ok := v.(map[string]any); !ok {
			if _, err := file.WriteString(fmt.Sprintf("%s = %v\n", k, v)); err != nil {
				return err
			}
		}
	}
	// 섹션
	for k, v := range data {
		if section, ok := v.(map[string]any); ok {
			if _, err := file.WriteString(fmt.Sprintf("\n[%s]\n", k)); err != nil {
				return err
			}
			for sk, sv := range section {
				if _, err := file.WriteString(fmt.Sprintf("%s = %v\n", sk, sv)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func saveDotEnv(path string, data map[string]any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for k, v := range data {
		if _, err := file.WriteString(fmt.Sprintf("%s=%v\n", k, v)); err != nil {
			return err
		}
	}
	return nil
}
