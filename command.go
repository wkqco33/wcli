package wcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/seoyc/wcli/rich"
)

var (
	// ErrCommandNotFound 하위 명령어를 찾을 수 없을 때 반환하는 에러
	ErrCommandNotFound = errors.New("command not found")
	// ErrHelp 도움말 요청 시 반환하는 에러
	ErrHelp = errors.New("help requested")
)

// Context 명령어 실행 시 전달되는 컨텍스트
type Context struct {
	context.Context
	Args []string
}

// Command CLI 명령어를 정의하는 구조체
type Command struct {
	Use     string
	Short   string
	Long    string
	Aliases []string // 커맨드 이름의 별칭 목록
	Version string   // 설정 시 --version 플래그 자동 등록

	// 실행 훅: PreRun → Run → PostRun 순서로 호출됨
	PreRun  func(ctx *Context) error
	Run     func(ctx *Context) error
	PostRun func(ctx *Context) error

	// PersistentPreRun/PersistentPostRun은 이 커맨드와 모든 하위 커맨드에서 실행됩니다.
	// 실행 순서: (루트→현재) PersistentPreRun → PreRun → Run → PostRun → (현재→루트) PersistentPostRun
	PersistentPreRun  func(ctx *Context) error
	PersistentPostRun func(ctx *Context) error

	// SilenceErrors true이면 Execute()에서 에러를 자동으로 출력하지 않음
	SilenceErrors bool

	// HelpFunc 커스텀 도움말 함수. 설정 시 기본 도움말 대신 호출됩니다.
	HelpFunc func(cmd *Command, w io.Writer)

	// OutWriter 표준 출력 대상 (nil이면 os.Stdout 사용)
	OutWriter io.Writer
	// ErrWriter 에러 출력 대상 (nil이면 os.Stderr 사용)
	ErrWriter io.Writer

	// 로컬 플래그 (이 커맨드에서만 유효)
	flags *FlagSet
	// Persistent 플래그 (이 커맨드 및 모든 하위 커맨드에서 유효)
	persistentFlags *FlagSet

	commands   []*Command
	commandMap map[string]*Command // 이름/별칭 → 커맨드 (O(1) 라우팅)
	parent     *Command
}

// outWriter OutWriter가 설정되어 있으면 반환, 아니면 os.Stdout
func (c *Command) outWriter() io.Writer {
	if c.OutWriter != nil {
		return c.OutWriter
	}
	return os.Stdout
}

// errWriter ErrWriter가 설정되어 있으면 반환, 아니면 os.Stderr
func (c *Command) errWriter() io.Writer {
	if c.ErrWriter != nil {
		return c.ErrWriter
	}
	return os.Stderr
}

// Name 커맨드 이름을 반환 (Use의 첫 번째 공백 이전 토큰)
func (c *Command) Name() string {
	name := c.Use
	if i := strings.IndexByte(name, ' '); i >= 0 {
		name = name[:i]
	}
	return name
}

// Flags 로컬 플래그 셋을 반환 (초기화 보장)
func (c *Command) Flags() *FlagSet {
	if c.flags == nil {
		c.flags = NewFlagSet()
	}
	return c.flags
}

// PersistentFlags Persistent 플래그 셋을 반환 (초기화 보장).
// 여기에 등록된 플래그는 이 커맨드 및 모든 하위 커맨드에서도 사용할 수 있습니다.
func (c *Command) PersistentFlags() *FlagSet {
	if c.persistentFlags == nil {
		c.persistentFlags = NewFlagSet()
	}
	return c.persistentFlags
}

// AddCommand 하위 명령어를 추가
func (c *Command) AddCommand(cmds ...*Command) {
	if c.commandMap == nil {
		c.commandMap = make(map[string]*Command)
	}
	for _, sub := range cmds {
		if sub == c {
			panic("Command can't be a child of itself")
		}
		sub.parent = c
		c.commands = append(c.commands, sub)
		c.commandMap[sub.Name()] = sub
		for _, alias := range sub.Aliases {
			c.commandMap[alias] = sub
		}
	}
}

// Execute 인자를 받아 플래그를 파싱하고 명령어를 실행
func (c *Command) Execute(args []string) error {
	ctx := &Context{
		Context: context.Background(),
		Args:    args,
	}
	err := c.execute(ctx)
	if err != nil {
		if errors.Is(err, ErrHelp) {
			return nil // 도움말 출력 후 정상 종료
		}
		if !c.SilenceErrors {
			rich.Fprintln(c.errWriter(), "[red][bold]Error:[/bold] %s[/red]", err.Error())
		}
	}
	return err
}

func (c *Command) execute(ctx *Context) error {
	// 1. 하위 명령어가 있는지 확인 (O(1) 맵 조회)
	if len(ctx.Args) > 0 {
		if sub, ok := c.commandMap[ctx.Args[0]]; ok {
			// 부모의 OutWriter/ErrWriter를 자식에게 전파
			if sub.OutWriter == nil && c.OutWriter != nil {
				sub.OutWriter = c.OutWriter
			}
			if sub.ErrWriter == nil && c.ErrWriter != nil {
				sub.ErrWriter = c.ErrWriter
			}
			ctx.Args = ctx.Args[1:]
			return sub.execute(ctx)
		}
	}

	// 2. --version 플래그 확인
	if c.Version != "" && c.isVersionRequested(ctx.Args) {
		fmt.Fprintln(c.outWriter(), c.Version)
		return ErrHelp
	}

	// 3. -h, --help 플래그 확인 (플래그 값 오탐 방지)
	if c.isHelpRequested(ctx.Args) {
		c.help(c.outWriter())
		return ErrHelp
	}

	// 4. 플래그 파싱 (조상의 persistent + 자신의 persistent + 로컬)
	combined := c.buildCombinedFlagSet()
	if combined != nil {
		remainingArgs, err := combined.Parse(ctx.Args)
		if err != nil {
			return err
		}
		if err := combined.Validate(); err != nil {
			return err
		}
		ctx.Args = remainingArgs
	}

	// 5. PersistentPreRun 훅 실행 (루트 → 현재 순서)
	for _, hook := range c.collectPersistentPreRuns() {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	// 6. PreRun 훅 실행
	if c.PreRun != nil {
		if err := c.PreRun(ctx); err != nil {
			return err
		}
	}

	// 7. 명령어 실행
	if c.Run != nil {
		if err := c.Run(ctx); err != nil {
			return err
		}
	} else if len(c.commands) > 0 {
		// Run이 없고 하위 명령어가 있는 경우 도움말 출력
		c.help(c.outWriter())
		return ErrHelp
	} else {
		return fmt.Errorf("no Run function defined for command: %s", c.Use)
	}

	// 8. PostRun 훅 실행
	if c.PostRun != nil {
		if err := c.PostRun(ctx); err != nil {
			return err
		}
	}

	// 9. PersistentPostRun 훅 실행 (현재 → 루트 순서)
	for _, hook := range c.collectPersistentPostRuns() {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// buildCombinedFlagSet 조상의 persistent 플래그 + 자신의 persistent 플래그 + 로컬 플래그를
// 하나의 FlagSet으로 병합하여 반환합니다. 플래그가 없으면 nil 반환.
func (c *Command) buildCombinedFlagSet() *FlagSet {
	combined := NewFlagSet()
	hasAny := false

	// 조상의 persistent 플래그 (루트 → 부모 순서로 병합)
	for _, anc := range c.ancestors() {
		if anc.persistentFlags != nil {
			combined.merge(anc.persistentFlags)
			hasAny = true
		}
	}
	// 자신의 persistent 플래그
	if c.persistentFlags != nil {
		combined.merge(c.persistentFlags)
		hasAny = true
	}
	// 자신의 로컬 플래그
	if c.flags != nil {
		combined.merge(c.flags)
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return combined
}

// ancestors 루트부터 부모까지의 조상 커맨드를 순서대로 반환합니다.
func (c *Command) ancestors() []*Command {
	var result []*Command
	for p := c.parent; p != nil; p = p.parent {
		result = append([]*Command{p}, result...)
	}
	return result
}

// collectPersistentPreRuns 루트 → 현재 순서로 PersistentPreRun 훅을 수집합니다.
func (c *Command) collectPersistentPreRuns() []func(*Context) error {
	var hooks []func(*Context) error
	for _, anc := range c.ancestors() {
		if anc.PersistentPreRun != nil {
			hooks = append(hooks, anc.PersistentPreRun)
		}
	}
	if c.PersistentPreRun != nil {
		hooks = append(hooks, c.PersistentPreRun)
	}
	return hooks
}

// collectPersistentPostRuns 현재 → 루트 순서로 PersistentPostRun 훅을 수집합니다.
func (c *Command) collectPersistentPostRuns() []func(*Context) error {
	var hooks []func(*Context) error
	if c.PersistentPostRun != nil {
		hooks = append(hooks, c.PersistentPostRun)
	}
	ancs := c.ancestors()
	for i := len(ancs) - 1; i >= 0; i-- {
		if ancs[i].PersistentPostRun != nil {
			hooks = append(hooks, ancs[i].PersistentPostRun)
		}
	}
	return hooks
}

// matchesName 커맨드 이름 또는 별칭이 일치하는지 확인
func (c *Command) matchesName(name string) bool {
	if c.Name() == name {
		return true
	}
	for _, alias := range c.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

// isVersionRequested args에서 --version 플래그를 감지합니다.
func (c *Command) isVersionRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--version" {
			return true
		}
	}
	return false
}

// isHelpRequested args에서 -h/--help 플래그를 감지합니다.
// 비-bool 플래그의 값으로 사용된 경우는 무시합니다.
func (c *Command) isHelpRequested(args []string) bool {
	combined := c.buildCombinedFlagSet()
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--" {
			return false
		}
		if arg == "-h" || arg == "--help" {
			return true
		}
		if strings.HasPrefix(arg, "--") && !strings.Contains(arg, "=") {
			name := arg[2:]
			if combined != nil {
				if flag, ok := combined.flags[name]; ok && flag.Type != TypeBool {
					skipNext = true
				}
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) == 2 && !strings.Contains(arg, "=") {
			short := arg[1:]
			if combined != nil {
				if flag, ok := combined.shorts[short]; ok && flag.Type != TypeBool {
					skipNext = true
				}
			}
		}
	}
	return false
}

// inheritedPersistentFlags 조상들의 persistent 플래그 목록을 반환합니다 (help 출력용).
func (c *Command) inheritedPersistentFlags() []*Flag {
	combined := NewFlagSet()
	for _, anc := range c.ancestors() {
		if anc.persistentFlags != nil {
			combined.merge(anc.persistentFlags)
		}
	}
	return combined.All()
}

// Help 도움말을 os.Stdout(또는 OutWriter)에 출력합니다.
func (c *Command) Help() {
	c.help(c.outWriter())
}
