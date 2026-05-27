package rich

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// ANSI 컬러 코드
const (
	Reset = "\033[0m"

	// 텍스트 스타일
	Bold          = "\033[1m"
	Dim           = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Blink         = "\033[5m"
	Reverse       = "\033[7m"
	Strikethrough = "\033[9m"

	// 전경 색상
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// 배경 색상
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// NoColor를 true로 설정하면 ANSI 코드를 출력하지 않습니다.
// NO_COLOR 환경변수가 설정된 경우 자동으로 활성화됩니다.
var NoColor bool

func init() {
	if os.Getenv("NO_COLOR") != "" {
		NoColor = true
	}
}

var tagMap = map[string]string{
	"bold":          Bold,
	"dim":           Dim,
	"underline":     Underline,
	"italic":        Italic,
	"blink":         Blink,
	"reverse":       Reverse,
	"strikethrough": Strikethrough,
	"black":         Black,
	"red":           Red,
	"green":         Green,
	"yellow":        Yellow,
	"blue":          Blue,
	"magenta":       Magenta,
	"cyan":          Cyan,
	"white":         White,
	"bg-black":      BgBlack,
	"bg-red":        BgRed,
	"bg-green":      BgGreen,
	"bg-yellow":     BgYellow,
	"bg-blue":       BgBlue,
	"bg-magenta":    BgMagenta,
	"bg-cyan":       BgCyan,
	"bg-white":      BgWhite,
}

// isTerminal w가 터미널(character device)인지 확인합니다.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		fi, err := f.Stat()
		return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
	}
	return false
}

// stripMarkup 마크업 태그를 제거하고 일반 텍스트를 반환합니다.
func stripMarkup(text string) string {
	var result strings.Builder
	runes := []rune(text)
	inTag := false
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		if char == '[' {
			if i > 0 && runes[i-1] == '\\' {
				resultStr := result.String()
				if len(resultStr) > 0 && resultStr[len(resultStr)-1] == '\\' {
					result.Reset()
					result.WriteString(resultStr[:len(resultStr)-1])
				}
				result.WriteRune('[')
				continue
			}
			inTag = true
			continue
		}
		if char == ']' && inTag {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// markupCache Markup() 결과를 캐시합니다. 동일 문자열은 한 번만 파싱됩니다.
var markupCache sync.Map // map[string]string

// Markup 간단한 마크업 파싱을 통해 텍스트에 ANSI 스타일을 적용합니다.
// 동일한 입력에 대한 결과는 캐시됩니다.
func Markup(text string) string {
	if cached, ok := markupCache.Load(text); ok {
		return cached.(string)
	}
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
				resultStr := result.String()
				if len(resultStr) > 0 && resultStr[len(resultStr)-1] == '\\' {
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
				// 닫는 태그: 이름으로 스택에서 찾아 제거 후 나머지 태그 재적용
				closingTagName := tag[1:]
				result.WriteString(Reset)
				for j := len(activeTags) - 1; j >= 0; j-- {
					if activeTags[j] == closingTagName {
						activeTags = append(activeTags[:j], activeTags[j+1:]...)
						break
					}
				}
				for _, t := range activeTags {
					if code, ok := tagMap[t]; ok {
						result.WriteString(code)
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

	output := result.String()
	// LoadOrStore: 동시에 같은 키가 들어와도 한 값만 저장됨
	if actual, loaded := markupCache.LoadOrStore(text, output); loaded {
		return actual.(string)
	}
	return output
}

// shouldColor w에 ANSI 코드를 출력해야 하는지 결정합니다.
func shouldColor(w io.Writer) bool {
	return !NoColor && isTerminal(w)
}

// Fprint 마크업이 적용된 텍스트를 w에 출력합니다.
// w가 터미널이 아니거나 NoColor가 true이면 ANSI 코드를 제거합니다.
func Fprint(w io.Writer, format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	if shouldColor(w) {
		fmt.Fprint(w, Markup(text))
	} else {
		fmt.Fprint(w, stripMarkup(text))
	}
}

// Fprintln 마크업이 적용된 텍스트를 w에 출력하고 줄바꿈합니다.
func Fprintln(w io.Writer, format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	if shouldColor(w) {
		fmt.Fprintln(w, Markup(text))
	} else {
		fmt.Fprintln(w, stripMarkup(text))
	}
}

// Print 마크업이 적용된 텍스트를 os.Stdout에 출력합니다.
func Print(format string, a ...any) {
	Fprint(os.Stdout, format, a...)
}

// Println 마크업이 적용된 텍스트를 os.Stdout에 출력하고 줄바꿈합니다.
func Println(format string, a ...any) {
	Fprintln(os.Stdout, format, a...)
}

// Sprint 마크업을 파싱하여 결과 문자열을 반환합니다.
// NoColor가 true인 경우 ANSI 코드를 제거합니다.
func Sprint(format string, a ...any) string {
	text := fmt.Sprintf(format, a...)
	if NoColor {
		return stripMarkup(text)
	}
	return Markup(text)
}
