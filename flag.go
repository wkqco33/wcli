package wcli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
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

	required bool               // MarkRequired로 설정
	validate func(string) error // SetValidation으로 설정
	wasSet   bool               // Parse 후 실제로 값이 설정됐는지 여부
}

// FlagSet 플래그들의 모음과 파싱 로직을 담당
type FlagSet struct {
	flags  map[string]*Flag // Name을 키로
	shorts map[string]*Flag // Shorthand를 키로
	sorted []*Flag          // All() 정렬 캐시 (addFlag/merge 호출 시 무효화)
}

// NewFlagSet 새로운 FlagSet을 생성
func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags:  make(map[string]*Flag),
		shorts: make(map[string]*Flag),
	}
}

// addFlag 플래그를 세트에 등록
func (f *FlagSet) addFlag(flag *Flag) {
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
		return fmt.Errorf("flag '%s' not found", name)
	}
	flag.required = true
	return nil
}

// SetValidation 플래그에 검증 함수를 등록합니다.
// fn은 파싱된 값의 문자열 표현을 받아 유효하지 않으면 에러를 반환해야 합니다.
func (f *FlagSet) SetValidation(name string, fn func(string) error) error {
	flag, ok := f.flags[name]
	if !ok {
		return fmt.Errorf("flag '%s' not found", name)
	}
	flag.validate = fn
	return nil
}

// Validate required 플래그 누락 및 검증 함수 실행을 확인합니다.
func (f *FlagSet) Validate() error {
	for _, flag := range f.flags {
		if flag.required && !flag.wasSet {
			return fmt.Errorf("required flag '--%s' not set", flag.Name)
		}
		if flag.wasSet && flag.validate != nil {
			val := flag.stringVal()
			if err := flag.validate(val); err != nil {
				return fmt.Errorf("flag '--%s': %w", flag.Name, err)
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
		valueInt:   p,
	})
}

// BoolVar 불리언 플래그 등록
func (f *FlagSet) BoolVar(p *bool, name, shorthand string, value bool, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Type:      TypeBool,
		valueBool: p,
	})
}

// Float64Var float64 플래그 등록
func (f *FlagSet) Float64Var(p *float64, name, shorthand string, value float64, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:         name,
		Shorthand:    shorthand,
		Usage:        usage,
		Type:         TypeFloat64,
		DefaultVal:   strconv.FormatFloat(value, 'f', -1, 64),
		valueFloat64: p,
	})
}

// DurationVar time.Duration 플래그 등록 (예: "30s", "1h30m")
func (f *FlagSet) DurationVar(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	*p = value
	f.addFlag(&Flag{
		Name:          name,
		Shorthand:     shorthand,
		Usage:         usage,
		Type:          TypeDuration,
		DefaultVal:    value.String(),
		valueDuration: p,
	})
}

// StringSliceVar 문자열 슬라이스 플래그 등록 (--flag val1 --flag val2 형식으로 누적)
func (f *FlagSet) StringSliceVar(p *[]string, name, shorthand string, value []string, usage string) {
	*p = value
	defaultVal := ""
	if len(value) > 0 {
		defaultVal = "[" + strings.Join(value, ",") + "]"
	}
	f.addFlag(&Flag{
		Name:             name,
		Shorthand:        shorthand,
		Usage:            usage,
		Type:             TypeStringSlice,
		DefaultVal:       defaultVal,
		valueStringSlice: p,
	})
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
				return nil, fmt.Errorf("unknown flag: --%s", name)
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
			var inlineVal string
			hasInline := false
			if eqIdx := strings.IndexByte(short, '='); eqIdx >= 0 {
				inlineVal = short[eqIdx+1:]
				short = short[:eqIdx]
				hasInline = true
			}
			flag, ok := f.shorts[short]
			if !ok {
				return nil, fmt.Errorf("unknown flag: -%s", short)
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
		return idx, fmt.Errorf("flag '%s' requires a value", args[idx])
	}

	val := args[idx+1]
	if err := f.setFlagValue(flag, val, args[idx]); err != nil {
		return idx, err
	}
	return idx + 1, nil
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
			return fmt.Errorf("invalid boolean value for flag '%s': %s", flagArg, val)
		}
	case TypeString:
		*flag.valueStr = val
	case TypeInt:
		parsedInt, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid integer value for flag '%s': %s", flagArg, val)
		}
		*flag.valueInt = parsedInt
	case TypeFloat64:
		parsedFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float64 value for flag '%s': %s", flagArg, val)
		}
		*flag.valueFloat64 = parsedFloat
	case TypeDuration:
		parsedDur, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid duration value for flag '%s': %s (example: 30s, 1h30m)", flagArg, val)
		}
		*flag.valueDuration = parsedDur
	case TypeStringSlice:
		*flag.valueStringSlice = append(*flag.valueStringSlice, val)
	}
	return nil
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
