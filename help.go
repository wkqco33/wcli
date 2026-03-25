package wcli

import (
	"fmt"
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
	if c.flags != nil && len(c.flags.flags) > 0 {
		parts = append(parts, "[flags]")
	}
	return strings.Join(parts, " ")
}

// Help 도움말 텍스트를 생성하고 출력합니다.
func (c *Command) Help() {
	// 설명 출력
	desc := c.Long
	if desc == "" {
		desc = c.Short
	}
	if desc != "" {
		fmt.Println(rich.Markup(desc))
		fmt.Println()
	}

	// 사용법
	rich.Println("[bold][yellow]Usage:[/yellow][/bold]")
	fmt.Println("  " + c.UsageLine())
	fmt.Println()

	// 하위 명령어 목록
	if len(c.commands) > 0 {
		rich.Println("[bold][yellow]Available Commands:[/yellow][/bold]")
		maxLen := 0
		for _, sub := range c.commands {
			if len(sub.Use) > maxLen {
				maxLen = len(sub.Use)
			}
		}
		for _, sub := range c.commands {
			padding := strings.Repeat(" ", maxLen-len(sub.Use))
			rich.Println("  [cyan]%s[/cyan]%s   %s", sub.Use, padding, sub.Short)
		}
		fmt.Println()
	}

	// 플래그 목록
	rich.Println("[bold][yellow]Flags:[/yellow][/bold]")
	if c.flags != nil {
		for _, f := range c.flags.All() {
			printFlagHelp(f)
		}
	}
	rich.Println("  [green]-h[/green], [green]--help[/green]          도움말 출력")
	fmt.Println()

	// 하위 명령어가 있을 경우 추가 안내
	if len(c.commands) > 0 {
		rich.Println(`Use "[cyan]%s [command] --help[/cyan]" for more information about a command.`, c.Use)
	}
}

// printFlagHelp 단일 플래그의 도움말 줄을 출력합니다.
func printFlagHelp(f *Flag) {
	typeStr := ""
	switch f.Type {
	case TypeString:
		typeStr = "string"
	case TypeInt:
		typeStr = "int"
	case TypeBool:
		typeStr = ""
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

	defaultPart := ""
	if f.DefaultVal != "" {
		defaultPart = fmt.Sprintf(" (기본값: %s)", f.DefaultVal)
	}

	rich.Println("  %s%s   %s%s", shortPart, flagPart, f.Usage, defaultPart)
}
