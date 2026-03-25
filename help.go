package wcli

import (
	"fmt"
	"io"
	"strings"

	"github.com/seoyc/wcli/rich"
)

// UsageLine 명령어의 사용법 한 줄을 반환합니다.
func (c *Command) UsageLine() string {
	use := c.Use
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

	// 설명 출력
	desc := c.Long
	if desc == "" {
		desc = c.Short
	}
	if desc != "" {
		rich.Fprintln(w, "%s", desc)
		fmt.Fprintln(w)
	}

	// 버전 출력 (설정된 경우)
	if c.Version != "" {
		rich.Fprintln(w, "[bold]Version:[/bold] %s", c.Version)
		fmt.Fprintln(w)
	}

	// 사용법
	rich.Fprintln(w, "[bold][yellow]Usage:[/yellow][/bold]")
	fmt.Fprintln(w, "  "+c.UsageLine())
	fmt.Fprintln(w)

	// 별칭 출력
	if len(c.Aliases) > 0 {
		rich.Fprintln(w, "[bold][yellow]Aliases:[/yellow][/bold]")
		fmt.Fprintln(w, "  "+c.Name()+", "+strings.Join(c.Aliases, ", "))
		fmt.Fprintln(w)
	}

	// 하위 명령어 목록
	if len(c.commands) > 0 {
		rich.Fprintln(w, "[bold][yellow]Available Commands:[/yellow][/bold]")
		maxLen := 0
		for _, sub := range c.commands {
			nameLen := len(sub.Name())
			if nameLen > maxLen {
				maxLen = nameLen
			}
		}
		for _, sub := range c.commands {
			name := sub.Name()
			padding := strings.Repeat(" ", maxLen-len(name))
			aliasHint := ""
			if len(sub.Aliases) > 0 {
				aliasHint = fmt.Sprintf(" (%s)", strings.Join(sub.Aliases, ", "))
			}
			rich.Fprintln(w, "  [cyan]%s[/cyan]%s   %s%s", name, padding, sub.Short, aliasHint)
		}
		fmt.Fprintln(w)
	}

	// 로컬 플래그 + 자신의 persistent 플래그
	localFlags := c.ownFlags()
	if len(localFlags) > 0 || true { // 항상 Flags 섹션 출력 (-h/--help 때문에)
		rich.Fprintln(w, "[bold][yellow]Flags:[/yellow][/bold]")
		for _, f := range localFlags {
			printFlagHelp(w, f)
		}
		rich.Fprintln(w, "  [green]-h[/green], [green]--help[/green]          print help")
		if c.Version != "" {
			rich.Fprintln(w, "      [green]--version[/green]       print version")
		}
		fmt.Fprintln(w)
	}

	// 조상의 persistent 플래그 (Global Flags)
	inherited := c.inheritedPersistentFlags()
	if len(inherited) > 0 {
		rich.Fprintln(w, "[bold][yellow]Global Flags:[/yellow][/bold]")
		for _, f := range inherited {
			printFlagHelp(w, f)
		}
		fmt.Fprintln(w)
	}

	// 하위 명령어가 있을 경우 추가 안내
	if len(c.commands) > 0 {
		rich.Fprintln(w, `Use "[cyan]%s [command] --help[/cyan]" for more information about a command.`, c.Use)
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

// printFlagHelp 단일 플래그의 도움말 줄을 w에 출력합니다.
func printFlagHelp(w io.Writer, f *Flag) {
	typeStr := ""
	switch f.Type {
	case TypeString:
		typeStr = "string"
	case TypeInt:
		typeStr = "int"
	case TypeBool:
		typeStr = ""
	case TypeFloat64:
		typeStr = "float64"
	case TypeDuration:
		typeStr = "duration"
	case TypeStringSlice:
		typeStr = "stringSlice"
	default:
		typeStr = "unknown"
	}

	shortPart := "    "
	if f.Shorthand != "" {
		shortPart = fmt.Sprintf("[green]-%s[/green], ", f.Shorthand)
	}

	flagPart := fmt.Sprintf("[green]--%s[/green]", f.Name)
	if typeStr != "" {
		flagPart += " " + typeStr
	}

	requiredMark := ""
	if f.required {
		requiredMark = " [red]*[/red]"
	}

	defaultPart := ""
	if f.DefaultVal != "" {
		defaultPart = fmt.Sprintf(" (default: %s)", f.DefaultVal)
	}

	rich.Fprintln(w, "  %s%s%s   %s%s", shortPart, flagPart, requiredMark, f.Usage, defaultPart)
}
