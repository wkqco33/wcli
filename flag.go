package wcli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// FlagType 플래그 값의 타입을 정의
type FlagType int

const (
	TypeString FlagType = iota
	TypeInt
	TypeBool
)

// Flag CLI 명령어의 플래그 정보를 담는 구조체
type Flag struct {
	Name       string
	Shorthand  string
	Usage      string
	Type       FlagType
	DefaultVal string // 기본값 (도움말 출력용)

	// 값 바인딩용 포인터
	valueStr  *string
	valueInt  *int
	valueBool *bool
}

// FlagSet 플래그들의 모음과 파싱 로직을 담당
type FlagSet struct {
	flags  map[string]*Flag // Name을 키로
	shorts map[string]*Flag // Shorthand를 키로
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

// Parse 인자를 파싱하여 플래그 값을 바인딩하고 남은 인자를 반환
func (f *FlagSet) Parse(args []string) ([]string, error) {
	var remainingArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			flag, ok := f.flags[name]
			if !ok {
				return nil, fmt.Errorf("알 수 없는 플래그: %s", arg)
			}
			var err error
			i, err = f.parseValue(i, args, flag)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]
			flag, ok := f.shorts[short]
			if !ok {
				return nil, fmt.Errorf("알 수 없는 플래그: %s", arg)
			}
			var err error
			i, err = f.parseValue(i, args, flag)
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
		return idx, nil // 현재 인덱스에서 끝 (값 없음)
	}

	if idx+1 >= len(args) {
		return idx, fmt.Errorf("플래그 '%s'에 값이 필요합니다", args[idx])
	}

	val := args[idx+1]
	switch flag.Type {
	case TypeString:
		*flag.valueStr = val
	case TypeInt:
		parsedInt, err := strconv.Atoi(val)
		if err != nil {
			return idx, fmt.Errorf("플래그 '%s'의 값이 정수가 아닙니다: %s", args[idx], val)
		}
		*flag.valueInt = parsedInt
	}

	return idx + 1, nil // 값 인덱스까지 소비했으므로 idx+1 반환 (for 루프에서 i++ 되므로 실질적으로 idx+2 위치로 이동)
}

// All 등록된 모든 플래그를 이름 순서로 반환
func (f *FlagSet) All() []*Flag {
	result := make([]*Flag, 0, len(f.flags))
	for _, flag := range f.flags {
		result = append(result, flag)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
