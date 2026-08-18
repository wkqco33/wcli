package wcli

import (
	"fmt"
	"io"
	"strings"
)

// NewCompletionCommand 셸 자동 완성 스크립트를 생성하는 'completion' 명령어를 생성합니다.
func NewCompletionCommand(root *Command) *Command {
	return &Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "셸 자동 완성 스크립트 생성",
		Long: `지정된 셸(bash, zsh 또는 fish)에 대한 자동 완성 스크립트를 생성하여 표준 출력으로 덤프합니다.

사용 방법:
  # Bash 완성 등록
  source <(your_app completion bash)

  # Zsh 완성 등록
  source <(your_app completion zsh)

  # Fish 완성 등록
  your_app completion fish > ~/.config/fish/completions/your_app.fish`,
		Run: func(ctx *Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("please specify a shell (bash, zsh, or fish)")
			}

			shell := strings.ToLower(ctx.Args[0])
			switch shell {
			case "bash":
				return GenBashCompletion(root, root.outWriter())
			case "zsh":
				return GenZshCompletion(root, root.outWriter())
			case "fish":
				return GenFishCompletion(root, root.outWriter())
			default:
				return fmt.Errorf("unsupported shell type: %s (must be one of: bash, zsh, fish)", shell)
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
	fmt.Fprintln(&b, "    local cur prev cword opts")
	fmt.Fprintln(&b, "    COMPREPLY=()")
	fmt.Fprintln(&b, "    cur=\"${COMP_WORDS[COMP_CWORD]}\"")
	fmt.Fprintln(&b, "    prev=\"${COMP_WORDS[COMP_CWORD-1]}\"")
	fmt.Fprintln(&b, "    cword=${COMP_CWORD}")
	fmt.Fprintln(&b)

	if len(cmd.commands) > 0 {
		fmt.Fprintln(&b, "    local subcmd=\"\"")
		fmt.Fprintln(&b, "    local i")
		fmt.Fprintln(&b, "    for ((i=1; i < cword; i++)); do")
		fmt.Fprintln(&b, "        local w=\"${COMP_WORDS[i]}\"")
		fmt.Fprintln(&b, "        if [[ \"${w}\" != -* ]]; then")
		fmt.Fprintln(&b, "            subcmd=\"${w}\"")
		fmt.Fprintln(&b, "            break")
		fmt.Fprintln(&b, "        fi")
		fmt.Fprintln(&b, "    done")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "    case \"${subcmd}\" in")

		for _, sub := range cmd.commands {
			var patterns []string
			patterns = append(patterns, sub.Name())
			patterns = append(patterns, sub.Aliases...)
			patternStr := strings.Join(patterns, "|")

			var subOpts []string
			for _, sf := range sub.ownFlags() {
				subOpts = append(subOpts, "--"+sf.Name)
				if sf.Shorthand != "" {
					subOpts = append(subOpts, "-"+sf.Shorthand)
				}
			}
			for _, child := range sub.commands {
				subOpts = append(subOpts, child.Name())
				subOpts = append(subOpts, child.Aliases...)
			}
			if len(subOpts) > 0 {
				fmt.Fprintf(&b, "        %s)\n", patternStr)
				fmt.Fprintf(&b, "            opts=\"%s\"\n", strings.Join(subOpts, " "))
				fmt.Fprintln(&b, "            ;;")
			}
		}

		// 루트 레벨 옵션 (서브커맨드 목록 + 루트 플래그)
		var rootOpts []string
		for _, sub := range cmd.commands {
			rootOpts = append(rootOpts, sub.Name())
			rootOpts = append(rootOpts, sub.Aliases...)
		}
		for _, f := range cmd.ownFlags() {
			rootOpts = append(rootOpts, "--"+f.Name)
			if f.Shorthand != "" {
				rootOpts = append(rootOpts, "-"+f.Shorthand)
			}
		}

		fmt.Fprintln(&b, "        *)")
		fmt.Fprintf(&b, "            opts=\"%s\"\n", strings.Join(rootOpts, " "))
		fmt.Fprintln(&b, "            ;;")
		fmt.Fprintln(&b, "    esac")
	} else {
		var rootOpts []string
		for _, f := range cmd.ownFlags() {
			rootOpts = append(rootOpts, "--"+f.Name)
			if f.Shorthand != "" {
				rootOpts = append(rootOpts, "-"+f.Shorthand)
			}
		}
		fmt.Fprintf(&b, "    opts=\"%s\"\n", strings.Join(rootOpts, " "))
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "    COMPREPLY=( $(compgen -W \"${opts}\" -- \"${cur}\") )")
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

// GenFishCompletion fish 자동 완성 스크립트를 생성합니다.
func GenFishCompletion(cmd *Command, w io.Writer) error {
	name := cmd.Name()
	var b strings.Builder

	// 모든 서브커맨드 이름 목록 (not __fish_seen_subcommand_from 조건용)
	var allSubNames []string
	for _, sub := range cmd.commands {
		allSubNames = append(allSubNames, sub.Name())
	}
	notSeenCond := ""
	if len(allSubNames) > 0 {
		notSeenCond = fmt.Sprintf(" -n 'not __fish_seen_subcommand_from %s'", strings.Join(allSubNames, " "))
	}

	// 루트 플래그 등록
	for _, f := range cmd.ownFlags() {
		desc := escapeFishDesc(f.Usage)
		if f.Shorthand != "" {
			fmt.Fprintf(&b, "complete -c %s%s -s %s -l %s -d '%s'\n", name, notSeenCond, f.Shorthand, f.Name, desc)
		} else {
			fmt.Fprintf(&b, "complete -c %s%s -l %s -d '%s'\n", name, notSeenCond, f.Name, desc)
		}
	}

	// 서브커맨드 등록
	for _, sub := range cmd.commands {
		desc := escapeFishDesc(sub.Short)
		fmt.Fprintf(&b, "complete -c %s -f%s -a %s -d '%s'\n", name, notSeenCond, sub.Name(), desc)

		// 서브커맨드 플래그 등록
		seenCond := fmt.Sprintf(" -n '__fish_seen_subcommand_from %s'", sub.Name())
		for _, f := range sub.ownFlags() {
			fDesc := escapeFishDesc(f.Usage)
			if f.Shorthand != "" {
				fmt.Fprintf(&b, "complete -c %s%s -s %s -l %s -d '%s'\n", name, seenCond, f.Shorthand, f.Name, fDesc)
			} else {
				fmt.Fprintf(&b, "complete -c %s%s -l %s -d '%s'\n", name, seenCond, f.Name, fDesc)
			}
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// escapeFishDesc fish complete -d 인수의 작은따옴표 이스케이프
func escapeFishDesc(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

func escapeZshUsage(s string) string {
	// zsh _arguments 설명 문자열 내 특수문자 처리
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "'", "")  // 스크립트 구조 파괴 방지
	s = strings.ReplaceAll(s, "\"", "") // 이중 인용부호 제거
	s = strings.ReplaceAll(s, "`", "")  // 백틱: 명령 치환 방지
	s = strings.ReplaceAll(s, "$", "")  // 달러: 변수 치환 방지
	return s
}
