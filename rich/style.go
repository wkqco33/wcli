package rich

import (
	"fmt"
	"strings"
)

// ANSI 컬러 코드
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	// 기본 색상
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

var tagMap = map[string]string{
	"bold":      Bold,
	"dim":       Dim,
	"underline": Underline,
	"black":     Black,
	"red":       Red,
	"green":     Green,
	"yellow":    Yellow,
	"blue":      Blue,
	"magenta":   Magenta,
	"cyan":      Cyan,
	"white":     White,
}

// Markup 간단한 마크업 파싱을 통해 텍스트에 ANSI 스타일을 적용합니다.
func Markup(text string) string {
	var result strings.Builder
	runes := []rune(text)
	inTag := false
	var currentTag strings.Builder
	var activeTags []string

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		if char == '[' {
			// 이스케이프 지원 (백슬래시 다음에 오는 '[')
			if i > 0 && runes[i-1] == '\\' {
				// 이미 이전에 백슬래시를 결과에 썼으므로 제거하고 '[' 추가
				resultStr := result.String()
				if len(resultStr) > 0 && resultStr[len(resultStr)-1] == '\\' {
					// 마지막 글자(백슬래시) 제거 (룬 단위가 아닌 단순 바이트 기준 슬라이싱이지만 ASCII 백슬래시이므로 안전)
					result.Reset()
					result.WriteString(resultStr[:len(resultStr)-1])
				}
				result.WriteRune('[')
				continue
			}
			inTag = true
			currentTag.Reset()
			continue
		}

		if char == ']' && inTag {
			inTag = false
			tag := currentTag.String()

			if strings.HasPrefix(tag, "/") {
				// 닫는 태그
				result.WriteString(Reset)
				if len(activeTags) > 0 {
					activeTags = activeTags[:len(activeTags)-1] // 스택 pop
					for _, t := range activeTags {
						if code, ok := tagMap[t]; ok {
							result.WriteString(code)
						}
					}
				}
			} else {
				// 여는 태그
				if code, ok := tagMap[tag]; ok {
					activeTags = append(activeTags, tag)
					result.WriteString(code)
				} else {
					// 알 수 없는 태그는 무시하고 그대로 출력
					result.WriteString(fmt.Sprintf("[%s]", tag))
				}
			}
			continue
		}

		if inTag {
			currentTag.WriteRune(char)
		} else {
			result.WriteRune(char)
		}
	}

	if len(activeTags) > 0 {
		result.WriteString(Reset)
	}

	return result.String()
}

// Print 마크업이 적용된 텍스트를 출력합니다.
func Print(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	fmt.Print(Markup(text))
}

// Println 마크업이 적용된 텍스트를 출력하고 줄바꿈합니다.
func Println(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	fmt.Println(Markup(text))
}
