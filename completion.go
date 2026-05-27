package wcli

import (
	"fmt"
	"io"
	"strings"
)

// NewCompletionCommand 셸 자동 완성 스크립트를 생성하는 'completion' 명령어를 생성합니다.
func NewCompletionCommand(root *Command) *Command {
	return &Command{
		Use:   "completion [bash|zsh]",
		Short: "셸 자동 완성 스크립트 생성",
		Long: `지정된 셸(bash 또는 zsh)에 대한 자동 완성 스크립트를 생성하여 표준 출력으로 덤프합니다.

사용 방법:
  # Bash 완성 등록
  source <(your_app completion bash)

  # Zsh 완성 등록
  source <(your_app completion zsh)`,
		Run: func(ctx *Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("셸 이름을 지정해주세요 (bash 또는 zsh)")
			}

			shell := strings.ToLower(ctx.Args[0])
			switch shell {
			case "bash":
				return GenBashCompletion(root, root.outWriter())
			case "zsh":
				return GenZshCompletion(root, root.outWriter())
			default:
				return fmt.Errorf("지원하지 않는 셸 타입: %s (bash, zsh 중 하나를 지정하세요)", shell)
			}
		},
	}
}

// GenBashCompletion bash 자동 완성 스크립트를 생성합니다.
func GenBashCompletion(cmd *Command, w io.Writer) error {
	name := cmd.Name()
	var b strings.Builder

	// 기본 bash 완성 핸들러 작성
	fmt.Fprintf(&b, "_%s_bash_autocomplete() {\n", name)
	fmt.Fprintln(&b, "    local cur prev opts")
	fmt.Fprintln(&b, "    COMPREPLY=()")
	fmt.Fprintln(&b, "    cur=\"${COMP_WORDS[COMP_CWORD]}\"")
	fmt.Fprintln(&b, "    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"")
	fmt.Fprintln(&b)

	// 계층 분석 없이 모든 명령어와 플래그를 편평하게 후보군으로 추출 (성능 및 제로 의존성 준수)
	var allWords []string
	collectAllWords(cmd, &allWords)
	optsStr := strings.Join(allWords, " ")

	fmt.Fprintf(&b, "    opts=\"%s\"\n", optsStr)
	fmt.Fprintln(&b, "    COMPREPLY=( $(compgen -W \"${opts}\" -- ${cur}) )")
	fmt.Fprintln(&b, "    return 0")
	fmt.Fprintln(&b, "}")
	fmt.Fprintf(&b, "complete -F _%s_bash_autocomplete %s\n", name, name)

	_, err := io.WriteString(w, b.String())
	return err
}

// GenZshCompletion zsh 자동 완성 스크립트를 생성합니다.
func GenZshCompletion(cmd *Command, w io.Writer) error {
	name := cmd.Name()
	var b strings.Builder

	fmt.Fprintf(&b, "#compdef %s\n\n", name)
	fmt.Fprintf(&b, "_%s() {\n", name)
	fmt.Fprintln(&b, "    local context state state_descr line")
	fmt.Fprintln(&b, "    typeset -A opt_args")
	fmt.Fprintln(&b)

	// 루트 플래그 빌드
	var flags []string
	localFlags := cmd.ownFlags()
	for _, f := range localFlags {
		flags = append(flags, fmt.Sprintf("'--%s[%s]'", f.Name, escapeZshUsage(f.Usage)))
		if f.Shorthand != "" {
			flags = append(flags, fmt.Sprintf("'-%s[%s]'", f.Shorthand, escapeZshUsage(f.Usage)))
		}
	}

	if len(flags) > 0 {
		fmt.Fprintln(&b, "    _arguments \\")
		for _, f := range flags {
			fmt.Fprintf(&b, "        %s \\\n", f)
		}
		fmt.Fprintln(&b, "        '1: :->cmds' \\")
		fmt.Fprintln(&b, "        '*:: :->args'")
	} else {
		fmt.Fprintln(&b, "    _arguments '1: :->cmds' '*:: :->args'")
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "    case $state in")
	fmt.Fprintln(&b, "        cmds)")
	fmt.Fprintln(&b, "            local -a subcmds")
	fmt.Fprintln(&b, "            subcmds=(")
	for _, sub := range cmd.commands {
		fmt.Fprintf(&b, "                \"%s:%s\"\n", sub.Name(), escapeZshUsage(sub.Short))
	}
	fmt.Fprintln(&b, "            )")
	fmt.Fprintln(&b, "            _describe \"command\" subcmds")
	fmt.Fprintln(&b, "            ;;")
	fmt.Fprintln(&b, "        args)")
	fmt.Fprintln(&b, "            case $line[1] in")
	for _, sub := range cmd.commands {
		fmt.Fprintf(&b, "                %s)\n", sub.Name())
		fmt.Fprintf(&b, "                    _arguments ")
		// 서브 플래그 덤프
		var subFlags []string
		for _, sf := range sub.ownFlags() {
			subFlags = append(subFlags, fmt.Sprintf("'--%s[%s]'", sf.Name, escapeZshUsage(sf.Usage)))
		}
		if len(subFlags) > 0 {
			fmt.Fprintln(&b, "\\")
			for _, sf := range subFlags {
				fmt.Fprintf(&b, "                        %s \\\n", sf)
			}
			fmt.Fprintln(&b, "                        '*: :_files'")
		} else {
			fmt.Fprintln(&b, "'*: :_files'")
		}
		fmt.Fprintln(&b, "                    ;;")
	}
	fmt.Fprintln(&b, "            esac")
	fmt.Fprintln(&b, "            ;;")
	fmt.Fprintln(&b, "    esac")
	fmt.Fprintln(&b, "}")
	fmt.Fprintf(&b, "_%s \"$@\"\n", name)

	_, err := io.WriteString(w, b.String())
	return err
}

func collectAllWords(cmd *Command, words *[]string) {
	// 중복 방지를 위한 맵
	visited := make(map[string]bool)

	var collect func(c *Command)
	collect = func(c *Command) {
		name := c.Name()
		if name != "" && !visited[name] {
			visited[name] = true
			*words = append(*words, name)
		}
		for _, alias := range c.Aliases {
			if !visited[alias] {
				visited[alias] = true
				*words = append(*words, alias)
			}
		}

		// 플래그 이름 수집
		flags := c.ownFlags()
		for _, f := range flags {
			nameKey := "--" + f.Name
			if !visited[nameKey] {
				visited[nameKey] = true
				*words = append(*words, nameKey)
			}
			if f.Shorthand != "" {
				shKey := "-" + f.Shorthand
				if !visited[shKey] {
					visited[shKey] = true
					*words = append(*words, shKey)
				}
			}
		}

		for _, sub := range c.commands {
			collect(sub)
		}
	}

	collect(cmd)
}

func escapeZshUsage(s string) string {
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	return s
}
