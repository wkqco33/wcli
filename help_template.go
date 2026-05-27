package wcli

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"

	"github.com/seoyc/wcli/rich"
)

// DefaultHelpTemplate wcli의 기본 도움말을 구성하는 텍스트 템플릿
const DefaultHelpTemplate = `{{if .Long}}{{.Long}}{{else}}{{.Short}}{{end}}

{{if .Version}}[bold]Version:[/bold] {{.Version}}
{{end}}
[bold][yellow]Usage:[/yellow][/bold]
  {{.UsageLine}}

{{if .Aliases}}[bold][yellow]Aliases:[/yellow][/bold]
  {{.Name}}, {{join .Aliases ", "}}
{{end}}
{{if .HasSubCommands}}[bold][yellow]Available Commands:[/yellow][/bold]
{{range .SubCommands}}  [cyan]{{.Name | pad .MaxNameLen}}[/cyan]   {{.Short}}{{if .Aliases}} ({{join .Aliases ", "}}){{end}}
{{end}}
{{end}}
{{if .HasLocalFlags}}[bold][yellow]Flags:[/yellow][/bold]
{{range .LocalFlags}}  {{.ShortPart}}{{.FlagPart}}{{.RequiredMark}}   {{.Usage}}{{.DefaultPart}}
{{end}}  [green]-h[/green], [green]--help[/green]          print help
{{if .Version}}      [green]--version[/green]       print version
{{end}}
{{end}}
{{if .HasGlobalFlags}}[bold][yellow]Global Flags:[/yellow][/bold]
{{range .GlobalFlags}}  {{.ShortPart}}{{.FlagPart}}{{.RequiredMark}}   {{.Usage}}{{.DefaultPart}}
{{end}}
{{end}}
{{if .HasSubCommands}}Use "[cyan]{{cleanUse .Use}} [command] --help[/cyan]" for more information about a command.
{{end}}`

var templateCache sync.Map // map[string]*template.Template

type helpData struct {
	Use            string
	Name           string
	Short          string
	Long           string
	Version        string
	UsageLine      string
	Aliases        []string
	HasSubCommands bool
	SubCommands    []subCommandHelpData
	HasLocalFlags  bool
	LocalFlags     []flagHelpData
	HasGlobalFlags bool
	GlobalFlags    []flagHelpData
}

type subCommandHelpData struct {
	Name       string
	Short      string
	Aliases    []string
	MaxNameLen int
}

type flagHelpData struct {
	ShortPart    string
	FlagPart     string
	RequiredMark string
	Usage        string
	DefaultPart  string
}

// renderHelpTemplate 주어진 템플릿과 커맨드 데이터를 사용해 도움말을 렌더링하여 w에 씁니다.
func renderHelpTemplate(w io.Writer, tmplStr string, cmd *Command) error {
	tmpl, err := getOrCompileTemplate(tmplStr)
	if err != nil {
		return err
	}

	data := buildHelpData(cmd)
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	// 템플릿 렌더링 결과에 rich 마크업 적용하여 출력
	rich.Fprint(w, "%s", buf.String())
	return nil
}

func getOrCompileTemplate(tmplStr string) (*template.Template, error) {
	if cached, ok := templateCache.Load(tmplStr); ok {
		return cached.(*template.Template), nil
	}

	funcMap := template.FuncMap{
		"join": func(a []string, sep string) string {
			return strings.Join(a, sep)
		},
		"pad": func(width int, s string) string {
			if len(s) >= width {
				return s
			}
			return s + strings.Repeat(" ", width-len(s))
		},
		"cleanUse": func(use string) string {
			// use 문자열의 첫 공백 이전 토큰만 추출
			if i := strings.IndexByte(use, ' '); i >= 0 {
				return use[:i]
			}
			return use
		},
	}

	tmpl, err := template.New("help").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("help template compile error: %w", err)
	}

	templateCache.Store(tmplStr, tmpl)
	return tmpl, nil
}

func escapeMarkupBrackets(s string) string {
	return strings.ReplaceAll(s, "[", "\\[")
}

func buildHelpData(cmd *Command) helpData {
	data := helpData{
		Use:            escapeMarkupBrackets(cmd.Use),
		Name:           cmd.Name(),
		Short:          cmd.Short,
		Long:           cmd.Long,
		Version:        cmd.Version,
		UsageLine:      escapeMarkupBrackets(cmd.UsageLine()),
		Aliases:        cmd.Aliases,
		HasSubCommands: len(cmd.commands) > 0,
	}

	if data.HasSubCommands {
		maxLen := 0
		for _, sub := range cmd.commands {
			if n := len(sub.Name()); n > maxLen {
				maxLen = n
			}
		}

		data.SubCommands = make([]subCommandHelpData, len(cmd.commands))
		for i, sub := range cmd.commands {
			data.SubCommands[i] = subCommandHelpData{
				Name:       sub.Name(),
				Short:      sub.Short,
				Aliases:    sub.Aliases,
				MaxNameLen: maxLen,
			}
		}
	}

	localFlags := cmd.ownFlags()
	data.HasLocalFlags = len(localFlags) > 0
	if data.HasLocalFlags {
		data.LocalFlags = make([]flagHelpData, len(localFlags))
		for i, f := range localFlags {
			data.LocalFlags[i] = buildFlagHelpData(f)
		}
	}

	inheritedFlags := cmd.inheritedPersistentFlags()
	data.HasGlobalFlags = len(inheritedFlags) > 0
	if data.HasGlobalFlags {
		data.GlobalFlags = make([]flagHelpData, len(inheritedFlags))
		for i, f := range inheritedFlags {
			data.GlobalFlags[i] = buildFlagHelpData(f)
		}
	}

	return data
}

func buildFlagHelpData(f *Flag) flagHelpData {
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

	return flagHelpData{
		ShortPart:    shortPart,
		FlagPart:     flagPart,
		RequiredMark: requiredMark,
		Usage:        f.Usage,
		DefaultPart:  defaultPart,
	}
}
