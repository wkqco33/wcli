package rich

import (
	"fmt"
	"io"
	"os"
	"strconv"
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

// isHexColor #rrggbb 형식의 16진수 컬러 코드인지 확인합니다.
func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
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
// Markup()과 동일한 규칙으로 태그 여부를 판별합니다: 알려진 여는 태그와 모든 닫는
// 태그([/xxx])만 제거되고, tagMap에 없는 대괄호 텍스트(예: "[red, blue] 중 선택")는
// 리터럴로 보존됩니다. 이 판별을 Markup()과 다르게 하면 색상 미지원 환경(stripMarkup
// 경로)에서 실제 마크업이 아닌 텍스트가 유실되거나, 폭 계산(DisplayWidth)이 실제
// 출력과 어긋나는 문제가 생깁니다.
func stripMarkup(text string) string {
	var result strings.Builder
	runes := []rune(text)
	inTag := false
	var currentTag strings.Builder
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
			currentTag.Reset()
			continue
		}
		if char == ']' && inTag {
			inTag = false
			tag := currentTag.String()
			if strings.HasPrefix(tag, "/") {
				// Markup()과 동일하게 닫는 태그는 알려진 이름 여부와 무관하게 항상 제거됨
				continue
			}
			if _, ok := tagMap[tag]; ok {
				continue
			}
			if isColorTag(tag) {
				continue
			}
			// 알 수 없는 여는 태그: Markup()과 동일하게 리터럴로 복원
			result.WriteRune('[')
			result.WriteString(tag)
			result.WriteRune(']')
			continue
		}
		if inTag {
			currentTag.WriteRune(char)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// isColorTag #rrggbb, color(N), bg-#rrggbb, bg-color(N) 형식의 컬러 태그인지 확인합니다.
func isColorTag(tag string) bool {
	switch {
	case strings.HasPrefix(tag, "bg-#") && len(tag) == 10:
		return isHexColor(tag[3:])
	case strings.HasPrefix(tag, "bg-color(") && strings.HasSuffix(tag, ")"):
		inner := tag[9 : len(tag)-1]
		if n, err := strconv.Atoi(inner); err == nil && n >= 0 && n <= 255 {
			return true
		}
		return false
	case strings.HasPrefix(tag, "#") && len(tag) == 7:
		return isHexColor(tag)
	case strings.HasPrefix(tag, "color(") && strings.HasSuffix(tag, ")"):
		inner := tag[6 : len(tag)-1]
		if n, err := strconv.Atoi(inner); err == nil && n >= 0 && n <= 255 {
			return true
		}
		return false
	}
	return false
}

// EscapeMarkup s에 포함된 '['를 이스케이프하여 Markup()/stripMarkup()이 태그로
// 해석하지 않도록 만듭니다. 에러 메시지, 사용자 입력 등 마크업을 의도하지 않은
// 동적 문자열을 마크업 포맷 문자열에 끼워 넣을 때 사용하세요. 그렇지 않으면
// 대괄호를 쓰는 흔한 표기(예: "허용값: [red] 또는 [blue]")가 실제 색상 태그로
// 오인되어 유실되거나 의도치 않게 스타일이 적용될 수 있습니다.
func EscapeMarkup(s string) string {
	return strings.ReplaceAll(s, "[", "\\[")
}

// markupCache Markup() 결과를 캐시합니다. 동일 문자열은 한 번만 파싱됩니다.
var markupCache sync.Map // map[string]string

// Markup 간단한 마크업 파싱을 통해 텍스트에 ANSI 스타일을 적용합니다.
// 동일한 입력에 대한 결과는 캐시됩니다.
// 지원 태그: [bold], [red], [bg-blue], [#ff4500], [color(208)], [bg-#ff4500], [bg-color(208)]
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
					} else if code := resolveColorTag(t); code != "" {
						result.WriteString(code)
					}
				}
			} else {
				// 여는 태그
				if code, ok := tagMap[tag]; ok {
					activeTags = append(activeTags, tag)
					result.WriteString(code)
				} else if code := resolveColorTag(tag); code != "" {
					activeTags = append(activeTags, tag)
					result.WriteString(code)
				} else {
					// 알 수 없는 태그는 무시하고 그대로 출력
					fmt.Fprintf(&result, "[%s]", tag)
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

// resolveColorTag #rrggbb, color(N), bg-#rrggbb, bg-color(N) 형식의 태그를 ANSI 코드로 변환합니다.
// 알 수 없는 형식이면 빈 문자열을 반환합니다.
func resolveColorTag(tag string) string {
	switch {
	case strings.HasPrefix(tag, "bg-#") && len(tag) == 10:
		// 배경색 TrueColor: bg-#rrggbb
		hex := tag[4:]
		r, _ := strconv.ParseUint(hex[0:2], 16, 64)
		g, _ := strconv.ParseUint(hex[2:4], 16, 64)
		b, _ := strconv.ParseUint(hex[4:6], 16, 64)
		return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
	case strings.HasPrefix(tag, "bg-color(") && strings.HasSuffix(tag, ")"):
		// 배경색 256색: bg-color(N)
		inner := tag[9 : len(tag)-1]
		if n, err := strconv.Atoi(inner); err == nil && n >= 0 && n <= 255 {
			return fmt.Sprintf("\033[48;5;%dm", n)
		}
	case strings.HasPrefix(tag, "#") && len(tag) == 7:
		// 전경색 TrueColor: #rrggbb
		r, _ := strconv.ParseUint(tag[1:3], 16, 64)
		g, _ := strconv.ParseUint(tag[3:5], 16, 64)
		b, _ := strconv.ParseUint(tag[5:7], 16, 64)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	case strings.HasPrefix(tag, "color(") && strings.HasSuffix(tag, ")"):
		// 전경색 256색: color(N)
		inner := tag[6 : len(tag)-1]
		if n, err := strconv.Atoi(inner); err == nil && n >= 0 && n <= 255 {
			return fmt.Sprintf("\033[38;5;%dm", n)
		}
	}
	return ""
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
