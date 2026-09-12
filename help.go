package wcli

import (
	"io"
	"strings"

	"github.com/wkqco33/wcli/rich"
)

// CommandPath 루트부터 현재 커맨드까지의 이름 경로를 반환합니다 (예: "ncli ai create").
// Use에 포함된 인자 표기(예: "[id]")는 제외하고 커맨드 이름만 사용합니다.
func (c *Command) CommandPath() string {
	name := c.Name()
	if c.parent == nil {
		return name
	}
	parent := c.parent.CommandPath()
	if parent == "" {
		return name
	}
	return parent + " " + name
}

// usagePrefix Usage 줄에 사용할 경로를 반환합니다.
// Use에 포함된 인자 표기(예: "[id]")를 그대로 유지합니다.
func (c *Command) usagePrefix() string {
	if c.parent == nil {
		return c.Use
	}

	parent := c.parent.CommandPath()
	switch {
	case parent == "":
		return c.Use
	case c.Use == "":
		return parent
	default:
		return parent + " " + c.Use
	}
}

// UsageLine 명령어의 사용법 한 줄을 반환합니다.
func (c *Command) UsageLine() string {
	use := c.usagePrefix()
	if use == "" {
		use = "command"
	}
	var parts []string
	parts = append(parts, use)
	if len(c.commands) > 0 {
		parts = append(parts, "[command]")
	}
	hasFlags := (c.flags != nil && len(c.flags.flags) > 0) ||
		(c.persistentFlags != nil && len(c.persistentFlags.flags) > 0) ||
		len(c.inheritedPersistentFlags()) > 0
	if hasFlags {
		parts = append(parts, "[flags]")
	}
	return strings.Join(parts, " ")
}

// help 도움말 텍스트를 생성하고 w에 출력합니다.
func (c *Command) help(w io.Writer) {
	// HelpFunc가 설정된 경우 커스텀 도움말 사용
	if c.HelpFunc != nil {
		c.HelpFunc(c, w)
		return
	}

	tmpl := c.HelpTemplate
	if tmpl == "" {
		tmpl = DefaultHelpTemplate
	}

	if err := renderHelpTemplate(w, tmpl, c); err != nil {
		// 템플릿 오류 시 안전장치로 표준 에러 출력
		rich.Fprintln(c.errWriter(), "[red]Help template error: %v[/red]", err)
	}
}

// ownFlags 이 커맨드 자신의 로컬 플래그와 persistent 플래그를 합쳐 반환합니다.
func (c *Command) ownFlags() []*Flag {
	combined := NewFlagSet()
	if c.flags != nil {
		combined.merge(c.flags)
	}
	if c.persistentFlags != nil {
		combined.merge(c.persistentFlags)
	}
	return combined.All()
}
