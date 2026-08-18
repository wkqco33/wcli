package wcli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wkqco33/wcli/config"
)

// FlagType 플래그 값의 타입을 정의
type FlagType int

const (
	TypeString FlagType = iota
	TypeInt
	TypeBool
	TypeFloat64
	TypeDuration
	TypeStringSlice
)

// Flag CLI 명령어의 플래그 정보를 담는 구조체
type Flag struct {
	Name       string
	Shorthand  string
	Category   string
	Usage      string
	Type       FlagType
	DefaultVal string // 기본값 (도움말 출력용)

	// 값 바인딩용 포인터
	valueStr         *string
	valueInt         *int
	valueBool        *bool
	valueFloat64     *float64
	valueDuration    *time.Duration
	valueStringSlice *[]string

	// 기본값 저장 (Reset용)
	defaultStr         string
	defaultInt         int
	defaultBool        bool
	defaultFloat64     float64
	defaultDuration    time.Duration
	defaultStringSlice []string

	required  bool               // MarkRequired로 설정
	validate  func(string) error // SetValidation으로 설정
	wasSet    bool               // Parse 후 실제로 값이 설정됐는지 여부
	envName   string             // BindEnv로 바인딩된 환경변수명
	configKey string             // BindConfig로 바인딩된 구성파일 키
}

// FlagSet 플래그들의 모음과 파싱 로직을 담당
type FlagSet struct {
	flags            map[string]*Flag // Name을 키로
	shorts           map[string]*Flag // Shorthand를 키로
	sorted           []*Flag          // All() 정렬 캐시 (addFlag/merge 호출 시 무효화)
	exclusiveGroups  [][]string       // 상호 배제 플래그 그룹 목록
	requiredTogether [][]string       // 필수 동반 지정 플래그 그룹 목록
	lookupEnv        func(string) (string, bool)
	configGetter     func(string) any
}

// NewFlagSet 새로운 FlagSet을 생성
func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags:        make(map[string]*Flag),
		shorts:       make(map[string]*Flag),
		lookupEnv:    os.LookupEnv,
		configGetter: config.Get,
	}
}

// SetLookupEnv 환경변수 조회 함수를 주입합니다 (테스트 격리 및 모킹 지원).
func (f *FlagSet) SetLookupEnv(fn func(string) (string, bool)) {
	if fn != nil {
		f.lookupEnv = fn
	} else {
		f.lookupEnv = os.LookupEnv
	}
}

// SetConfigGetter 설정 값 조회 함수를 주입합니다 (테스트 격리 및 모킹 지원).
func (f *FlagSet) SetConfigGetter(fn func(string) any) {
	if fn != nil {
		f.configGetter = fn
	} else {
		f.configGetter = config.Get
	}
}

func (f *FlagSet) getLookupEnv() func(string) (string, bool) {
	if f.lookupEnv != nil {
		return f.lookupEnv
	}
	return os.LookupEnv
}

func (f *FlagSet) getConfigGetter() func(string) any {
	if f.configGetter != nil {
		return f.configGetter
	}
	return config.Get
}

// addFlag 플래그를 세트에 등록. 같은 이름/단축키가 이미 등록돼 있으면 panic합니다(개발 단계 실수 방지).
func (f *FlagSet) addFlag(flag *Flag) {
	if _, exists := f.flags[flag.Name]; exists {
		panic(fmt.Sprintf("wcli: flag --%s redefined", flag.Name))
	}
	if flag.Shorthand != "" {
		if existing, exists := f.shorts[flag.Shorthand]; exists {
			panic(fmt.Sprintf("wcli: shorthand -%s redefined (flags --%s and --%s)", flag.Shorthand, existing.Name, flag.Name))
		}
	}
	f.flags[flag.Name] = flag
	if flag.Shorthand != "" {
		f.shorts[flag.Shorthand] = flag
	}
	f.sorted = nil // 정렬 캐시 무효화
}

// MarkRequired 플래그를 필수로 표시합니다. 파싱 후 Validate() 호출 시 확인됩니다.
func (f *FlagSet) MarkRequired(name string) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	flag.required = true
	return nil
}

// SetValidation 플래그에 검증 함수를 등록합니다.
// fn은 파싱된 값의 문자열 표현을 받아 유효하지 않으면 에러를 반환해야 합니다.
func (f *FlagSet) SetValidation(name string, fn func(string) error) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	flag.validate = fn
	return nil
}

// SetCategory 플래그 도움말 분류 이름을 설정합니다.
func (f *FlagSet) SetCategory(name, category string) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	flag.Category = category
	return nil
}

// BindEnv 플래그에 환경변수를 바인딩합니다. 플래그가 지정되지 않았을 때 환경변수에서 읽어옵니다.
func (f *FlagSet) BindEnv(name, envName string) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	flag.envName = envName
	return nil
}

// BindConfig 플래그에 설정 파일 키를 바인딩합니다. 플래그와 환경변수가 지정되지 않았을 때 설정 파일에서 읽어옵니다.
func (f *FlagSet) BindConfig(name, configKey string) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	flag.configKey = configKey
	return nil
}

// MarkAllowed 플래그에 허용 가능한 값 목록을 설정합니다.
// 파싱된 값이 목록에 없으면 ValidationError를 반환합니다.
func (f *FlagSet) MarkAllowed(name string, allowed ...string) error {
	flag, ok := f.flags[name]
	if !ok {
		return &FlagError{FlagName: name, Err: fmt.Errorf("flag '%s' not found", name)}
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}
	flag.validate = func(val string) error {
		if _, ok := allowedSet[val]; !ok {
			return fmt.Errorf("value must be one of: %s", strings.Join(allowed, ", "))
		}
		return nil
	}
	return nil
}

// MarkFlagsMutuallyExclusive 플래그 목록을 상호 배제하도록 지정합니다. 지정된 플래그들 중 하나만 설정되어야 합니다.
func (f *FlagSet) MarkFlagsMutuallyExclusive(names ...string) {
	f.exclusiveGroups = append(f.exclusiveGroups, names)
}

// MarkFlagsRequiredTogether 플래그 목록을 동반 지정하도록 지정합니다. 하나라도 지정되면 모두 지정되어야 합니다.
func (f *FlagSet) MarkFlagsRequiredTogether(names ...string) {
	f.requiredTogether = append(f.requiredTogether, names)
}

// Validate required 플래그 누락 및 검증 함수 실행을 확인합니다.
func (f *FlagSet) Validate() error {
	lookup := f.getLookupEnv()
	configGet := f.getConfigGetter()

	// 1. 환경변수 및 설정파일 바인딩 처리 (우선순위 체인)
	for _, flag := range f.flags {
		if flag.wasSet {
			continue
		}

		// (1) 환경변수 바인딩 검사 및 설정
		if flag.envName != "" {
			if val, exists := lookup(flag.envName); exists {
				if err := f.setFlagValue(flag, val, fmt.Sprintf("env %s", flag.envName)); err != nil {
					return err
				}
				continue // 환경변수가 세팅되면 설정파일보다 우선순위가 높음
			}
		}

		// (2) 설정파일 바인딩 검사 및 설정
		if flag.configKey != "" {
			if val := configGet(flag.configKey); val != nil {
				if err := f.setConfigFlagValue(flag, val, fmt.Sprintf("config %s", flag.configKey)); err != nil {
					return err
				}
			}
		}
	}

	// 2. 상호 배제 검증
	for _, group := range f.exclusiveGroups {
		var setNames []string
		for _, name := range group {
			if flag, ok := f.flags[name]; ok && flag.wasSet {
				setNames = append(setNames, fmt.Sprintf("--%s", name))
			}
		}
		if len(setNames) > 1 {
			return &ValidationError{
				FlagName: group[0],
				Err:      fmt.Errorf("flags %s are mutually exclusive", strings.Join(setNames, " and ")),
			}
		}
	}

	// 3. 필수 동반 지정 검증
	for _, group := range f.requiredTogether {
		var anySet bool
		var allSet = true
		var missing []string
		for _, name := range group {
			if flag, ok := f.flags[name]; ok {
				if flag.wasSet {
					anySet = true
				} else {
					allSet = false
					missing = append(missing, fmt.Sprintf("--%s", name))
				}
			}
		}
		if anySet && !allSet {
			return &ValidationError{
				FlagName: group[0],
				Err: fmt.Errorf("if any of %s are set, all must be set (missing: %s)",
					strings.Join(group, ", "), strings.Join(missing, ", ")),
			}
		}
	}

	// 4. 필수 플래그 누락 및 개별 검증 검사
	// f.flags(map) 대신 이름순으로 정렬된 f.All()을 순회해야 한다: 맵 순회
	// 순서는 실행마다 무작위이므로, 필수 플래그가 둘 이상 동시에 누락된 경우
	// map을 그대로 순회하면 완전히 동일한 커맨드 실행에도 매번 다른 플래그가
	// "누락됨"으로 보고되어 에러 메시지가 비결정적이 된다.
	for _, flag := range f.All() {
		if flag.required && !flag.wasSet {
			return &ValidationError{
				FlagName: flag.Name,
				Err:      fmt.Errorf("required flag '--%s' not set", flag.Name),
			}
		}
		if flag.wasSet && flag.validate != nil {
			val := flag.stringVal()
			if err := flag.validate(val); err != nil {
				return &ValidationError{
					FlagName: flag.Name,
					Err:      err,
				}
			}
		}
	}
	return nil
}

// merge 다른 FlagSet의 모든 플래그를 이 FlagSet에 병합합니다 (충돌 시 other 우선).
func (f *FlagSet) merge(other *FlagSet) {
	if other == nil {
		return
	}
	for _, flag := range other.flags {
		f.flags[flag.Name] = flag
		if flag.Shorthand != "" {
			f.shorts[flag.Shorthand] = flag
		}
	}
	// 그룹 제약(상호 배제/필수 동반)도 함께 병합해야 combined FlagSet의 Validate()에서 검증됨
	f.exclusiveGroups = append(f.exclusiveGroups, other.exclusiveGroups...)
	f.requiredTogether = append(f.requiredTogether, other.requiredTogether...)
	if other.lookupEnv != nil {
		f.lookupEnv = other.lookupEnv
	}
	if other.configGetter != nil {
		f.configGetter = other.configGetter
	}
	f.sorted = nil // 정렬 캐시 무효화
}

// StringVar 문자열 플래그 등록
func (f *FlagSet) StringVar(p *string, name, shorthand, value, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:       name,
		Shorthand:  shorthand,
		Usage:      usage,
		Type:       TypeString,
		DefaultVal: value,
		defaultStr: value,
		valueStr:   p,
	})
}

// IntVar 정수 플래그 등록
func (f *FlagSet) IntVar(p *int, name, shorthand string, value int, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:       name,
		Shorthand:  shorthand,
		Usage:      usage,
		Type:       TypeInt,
		DefaultVal: strconv.Itoa(value),
		defaultInt: value,
		valueInt:   p,
	})
}

// BoolVar 불리언 플래그 등록
func (f *FlagSet) BoolVar(p *bool, name, shorthand string, value bool, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:        name,
		Shorthand:   shorthand,
		Usage:       usage,
		Type:        TypeBool,
		defaultBool: value,
		valueBool:   p,
	})
}

// Float64Var float64 플래그 등록
func (f *FlagSet) Float64Var(p *float64, name, shorthand string, value float64, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:           name,
		Shorthand:      shorthand,
		Usage:          usage,
		Type:           TypeFloat64,
		DefaultVal:     strconv.FormatFloat(value, 'f', -1, 64),
		defaultFloat64: value,
		valueFloat64:   p,
	})
}

// DurationVar time.Duration 플래그 등록 (예: "30s", "1h30m")
func (f *FlagSet) DurationVar(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:            name,
		Shorthand:       shorthand,
		Usage:           usage,
		Type:            TypeDuration,
		DefaultVal:      value.String(),
		defaultDuration: value,
		valueDuration:   p,
	})
}

// StringSliceVar 문자열 슬라이스 플래그 등록 (--flag val1 --flag val2 형식으로 누적)
func (f *FlagSet) StringSliceVar(p *[]string, name, shorthand string, value []string, usage string) {
	*p = append([]string(nil), value...)
	defaultVal := ""
	if len(value) > 0 {
		defaultVal = "[" + strings.Join(value, ",") + "]"
	}
	f.addFlag(&Flag{
		Name:               name,
		Shorthand:          shorthand,
		Usage:              usage,
		Type:               TypeStringSlice,
		DefaultVal:         defaultVal,
		defaultStringSlice: append([]string(nil), value...),
		valueStringSlice:   p,
	})
}

// reset 플래그의 상태를 초기 등록 당시의 기본값으로 복원합니다.
func (f *Flag) reset() {
	f.wasSet = false
	switch f.Type {
	case TypeString:
		if f.valueStr != nil {
			*f.valueStr = f.defaultStr
		}
	case TypeInt:
		if f.valueInt != nil {
			*f.valueInt = f.defaultInt
		}
	case TypeBool:
		if f.valueBool != nil {
			*f.valueBool = f.defaultBool
		}
	case TypeFloat64:
		if f.valueFloat64 != nil {
			*f.valueFloat64 = f.defaultFloat64
		}
	case TypeDuration:
		if f.valueDuration != nil {
			*f.valueDuration = f.defaultDuration
		}
	case TypeStringSlice:
		if f.valueStringSlice != nil {
			*f.valueStringSlice = append([]string(nil), f.defaultStringSlice...)
		}
	}
}

// Reset 등록된 모든 플래그의 wasSet 상태를 false로 초기화하고 바인딩된 변수 값을 초기 기본값으로 복원합니다.
func (f *FlagSet) Reset() {
	if f == nil {
		return
	}
	for _, flag := range f.flags {
		flag.reset()
	}
}

// Parse 인자를 파싱하여 플래그 값을 바인딩하고 남은 인자를 반환
func (f *FlagSet) Parse(args []string) ([]string, error) {
	var remainingArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// -- 종결자: 이후 모든 인자는 positional로 처리
		if arg == "--" {
			remainingArgs = append(remainingArgs, args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			var inlineVal string
			hasInline := false
			if eqIdx := strings.IndexByte(name, '='); eqIdx >= 0 {
				inlineVal = name[eqIdx+1:]
				name = name[:eqIdx]
				hasInline = true
			}
			flag, ok := f.flags[name]
			if !ok {
				return nil, &FlagError{FlagName: name, Err: fmt.Errorf("unknown flag: --%s", name)}
			}
			var err error
			if hasInline {
				err = f.setFlagValue(flag, inlineVal, arg)
			} else {
				i, err = f.parseValue(i, args, flag)
			}
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]
			// -o=value 형태의 인라인 값
			if eqIdx := strings.IndexByte(short, '='); eqIdx >= 0 {
				name := short[:eqIdx]
				flag, ok := f.shorts[name]
				if !ok {
					return nil, &FlagError{FlagName: name, Err: fmt.Errorf("unknown flag: -%s", name)}
				}
				if err := f.setFlagValue(flag, short[eqIdx+1:], arg); err != nil {
					return nil, err
				}
				continue
			}
			// 등록된 단축키면 단일 처리(다중 문자 단축키 호환), 아니면 결합 단축 플래그(-vh)로 파싱
			var err error
			if flag, ok := f.shorts[short]; ok {
				i, err = f.parseValue(i, args, flag)
			} else {
				i, err = f.parseShortCluster(i, args, short)
			}
			if err != nil {
				return nil, err
			}
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}
	return remainingArgs, nil
}

func (f *FlagSet) parseValue(idx int, args []string, flag *Flag) (int, error) {
	if flag.Type == TypeBool {
		*flag.valueBool = true
		flag.wasSet = true
		return idx, nil // 현재 인덱스에서 끝 (값 없음)
	}

	if idx+1 >= len(args) {
		return idx, &FlagError{FlagName: flag.Name, Err: fmt.Errorf("flag '%s' requires a value", args[idx])}
	}

	val := args[idx+1]
	if err := f.setFlagValue(flag, val, args[idx]); err != nil {
		return idx, err
	}
	return idx + 1, nil
}

// parseShortCluster 결합 단축 플래그(-vh, -vofile 등)를 파싱합니다.
// 각 문자를 bool 플래그로 처리하다가 비-bool 플래그를 만나면 나머지 문자열(또는 다음 인자)을 값으로 사용하고 종료합니다.
func (f *FlagSet) parseShortCluster(idx int, args []string, cluster string) (int, error) {
	for k := 0; k < len(cluster); k++ {
		ch := cluster[k : k+1]
		flag, ok := f.shorts[ch]
		if !ok {
			return idx, &FlagError{FlagName: ch, Err: fmt.Errorf("unknown flag: -%s", ch)}
		}
		if flag.Type == TypeBool {
			*flag.valueBool = true
			flag.wasSet = true
			continue
		}
		// 비-bool 플래그: 뒤에 남은 문자가 있으면 값으로(-ofile), 없으면 다음 인자를 값으로(-o file)
		if rest := cluster[k+1:]; rest != "" {
			return idx, f.setFlagValue(flag, rest, "-"+ch)
		}
		if idx+1 >= len(args) {
			return idx, &FlagError{FlagName: flag.Name, Err: fmt.Errorf("flag '-%s' requires a value", ch)}
		}
		return idx + 1, f.setFlagValue(flag, args[idx+1], "-"+ch)
	}
	return idx, nil
}

func (f *FlagSet) setFlagValue(flag *Flag, val string, flagArg string) error {
	flag.wasSet = true
	switch flag.Type {
	case TypeBool:
		switch strings.ToLower(val) {
		case "true", "1", "yes":
			*flag.valueBool = true
		case "false", "0", "no":
			*flag.valueBool = false
		default:
			return &FlagError{FlagName: flag.Name, Err: fmt.Errorf("invalid boolean value for flag '%s': %s", flagArg, val)}
		}
	case TypeString:
		*flag.valueStr = val
	case TypeInt:
		parsedInt, err := strconv.Atoi(val)
		if err != nil {
			return &FlagError{FlagName: flag.Name, Err: fmt.Errorf("invalid integer value for flag '%s': %s", flagArg, val)}
		}
		*flag.valueInt = parsedInt
	case TypeFloat64:
		parsedFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return &FlagError{FlagName: flag.Name, Err: fmt.Errorf("invalid float64 value for flag '%s': %s", flagArg, val)}
		}
		*flag.valueFloat64 = parsedFloat
	case TypeDuration:
		parsedDur, err := time.ParseDuration(val)
		if err != nil {
			return &FlagError{FlagName: flag.Name, Err: fmt.Errorf("invalid duration value for flag '%s': %s (example: 30s, 1h30m)", flagArg, val)}
		}
		*flag.valueDuration = parsedDur
	case TypeStringSlice:
		*flag.valueStringSlice = append(*flag.valueStringSlice, val)
	}
	return nil
}

func (f *FlagSet) setConfigFlagValue(flag *Flag, val any, flagArg string) error {
	if flag.Type != TypeStringSlice {
		return f.setFlagValue(flag, fmt.Sprintf("%v", val), flagArg)
	}

	flag.wasSet = true
	items := configValueToStringSlice(val)
	*flag.valueStringSlice = append((*flag.valueStringSlice)[:0], items...)
	return nil
}

func configValueToStringSlice(val any) []string {
	switch v := val.(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), v...)
	case []any:
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = fmt.Sprintf("%v", item)
		}
		return items
	case string:
		parts := strings.Split(v, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				items = append(items, trimmed)
			}
		}
		return items
	default:
		return []string{fmt.Sprintf("%v", val)}
	}
}

// stringVal 플래그의 현재 값을 문자열로 반환합니다 (검증 콜백용).
func (f *Flag) stringVal() string {
	switch f.Type {
	case TypeString:
		return *f.valueStr
	case TypeInt:
		return strconv.Itoa(*f.valueInt)
	case TypeBool:
		return strconv.FormatBool(*f.valueBool)
	case TypeFloat64:
		return strconv.FormatFloat(*f.valueFloat64, 'f', -1, 64)
	case TypeDuration:
		return f.valueDuration.String()
	case TypeStringSlice:
		return "[" + strings.Join(*f.valueStringSlice, ",") + "]"
	}
	return ""
}

// All 등록된 모든 플래그를 이름 순서로 반환합니다.
// 결과는 캐시되며 플래그가 추가될 때만 재정렬됩니다.
func (f *FlagSet) All() []*Flag {
	if f.sorted != nil {
		return f.sorted
	}
	result := make([]*Flag, 0, len(f.flags))
	for _, flag := range f.flags {
		result = append(result, flag)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	f.sorted = result
	return result
}

// Changed 해당 이름의 플래그가 CLI 인자, 환경변수 또는 설정파일 등을 통해 명시적으로 설정되었는지 여부를 반환합니다.
func (f *FlagSet) Changed(name string) bool {
	if flag, ok := f.flags[name]; ok {
		return flag.wasSet
	}
	return false
}
