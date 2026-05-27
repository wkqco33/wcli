package wcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/seoyc/wcli/logging"
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
	Args   []string
	Logger logging.Logger
}

// Command CLI 명령어를 정의하는 구조체
type Command struct {
	Use          string
	Short        string
	Long         string
	Aliases      []string // 커맨드 이름의 별칭 목록
	Version      string   // 설정 시 --version 플래그 자동 등록
	HelpTemplate string   // 설정 시 기본 도움말 대신 템플릿으로 출력
	GroupName    string   // 부모 도움말에서 이 커맨드를 분류할 그룹 이름

	// Logger 명령어 실행 시 사용할 로거
	Logger logging.Logger

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

func (c *Command) getLogger() logging.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	if c.parent != nil {
		return c.parent.getLogger()
	}
	return logging.GetLogger()
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
	logger := c.getLogger()
	ctx := &Context{
		Context: context.Background(),
		Args:    args,
		Logger:  logger,
	}
	logger.Log(logging.LevelDebug, "Executing command %q with args: %v", c.Name(), args)

	err := c.execute(ctx)
	if err != nil {
		if errors.Is(err, ErrHelp) {
			logger.Log(logging.LevelDebug, "Command %q execution interrupted: help requested", c.Name())
			return nil // 도움말 출력 후 정상 종료
		}
		logger.Log(logging.LevelError, "Command %q failed: %v", c.Name(), err)
		if !c.SilenceErrors {
			rich.Fprintln(c.errWriter(), "[red][bold]Error:[/bold] %s[/red]", err.Error())
		}
	} else {
		logger.Log(logging.LevelDebug, "Command %q executed successfully", c.Name())
	}
	return err
}

func (c *Command) execute(ctx *Context) error {
	logger := ctx.Logger
	// 1. 하위 명령어가 있는지 확인 (O(1) 맵 조회)
	if len(ctx.Args) > 0 {
		if sub, ok := c.commandMap[ctx.Args[0]]; ok {
			logger.Log(logging.LevelDebug, "Routing command from %q to sub-command %q", c.Name(), sub.Name())
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

		// 1-1. commandMap miss + 서브커맨드가 있을 때 퍼지 매칭 제안
		input := ctx.Args[0]
		if len(c.commands) > 0 && len(input) > 0 && input[0] != '-' {
			if suggestions := suggestCommands(input, c.commandMap); len(suggestions) > 0 {
				return fmt.Errorf("unknown command %q\n\nDid you mean?\n  %s",
					input, strings.Join(suggestions, "\n  "))
			}
		}
	}

	// 2. --version 플래그 확인
	if c.Version != "" && c.isVersionRequested(ctx.Args) {
		logger.Log(logging.LevelDebug, "Version flag detected on command %q", c.Name())
		fmt.Fprintln(c.outWriter(), c.Version)
		return ErrHelp
	}

	// 3. -h, --help 플래그 확인 (플래그 값 오탐 방지)
	if c.isHelpRequested(ctx.Args) {
		logger.Log(logging.LevelDebug, "Help flag detected on command %q", c.Name())
		c.help(c.outWriter())
		return ErrHelp
	}

	// 4. 플래그 파싱 (조상의 persistent + 자신의 persistent + 로컬)
	combined := c.buildCombinedFlagSet()
	if combined != nil {
		logger.Log(logging.LevelDebug, "Parsing flags for command %q", c.Name())
		remainingArgs, err := combined.Parse(ctx.Args)
		if err != nil {
			if flagErr, ok := err.(*FlagError); ok {
				flagErr.CommandName = c.Name()
			}
			return err
		}
		if err := combined.Validate(); err != nil {
			if valErr, ok := err.(*ValidationError); ok {
				valErr.CommandName = c.Name()
			}
			return err
		}
		ctx.Args = remainingArgs
	}

	// 5. PersistentPreRun 훅 실행 (루트 → 현재 순서)
	for _, hook := range c.collectPersistentPreRuns() {
		logger.Log(logging.LevelDebug, "Running PersistentPreRun for command %q", c.Name())
		if err := hook(ctx); err != nil {
			return err
		}
	}

	// 6. PreRun 훅 실행
	if c.PreRun != nil {
		logger.Log(logging.LevelDebug, "Running PreRun for command %q", c.Name())
		if err := c.PreRun(ctx); err != nil {
			return err
		}
	}

	// 7. 명령어 실행
	if c.Run != nil {
		logger.Log(logging.LevelDebug, "Running main function for command %q", c.Name())
		if err := c.Run(ctx); err != nil {
			return err
		}
	} else if len(c.commands) > 0 {
		// Run이 없고 하위 명령어가 있는 경우 도움말 출력
		logger.Log(logging.LevelDebug, "No Run defined, showing help for command %q", c.Name())
		c.help(c.outWriter())
		return ErrHelp
	} else {
		return &CommandError{CommandName: c.Name(), Err: fmt.Errorf("no Run function defined for command: %s", c.Use)}
	}

	// 8. PostRun 훅 실행
	if c.PostRun != nil {
		logger.Log(logging.LevelDebug, "Running PostRun for command %q", c.Name())
		if err := c.PostRun(ctx); err != nil {
			return err
		}
	}

	// 9. PersistentPostRun 훅 실행 (현재 → 루트 순서)
	for _, hook := range c.collectPersistentPostRuns() {
		logger.Log(logging.LevelDebug, "Running PersistentPostRun for command %q", c.Name())
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
	// 역순으로 수집 후 제자리 뒤집기: O(n) (prepend 방식의 O(n²) 개선)
	var reversed []*Command
	for p := c.parent; p != nil; p = p.parent {
		reversed = append(reversed, p)
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
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

// levenshtein 두 문자열 간의 편집 거리(Levenshtein distance)를 계산합니다.
// rolling array 방식으로 O(m*n) 시간, O(min(m,n)) 공간을 사용합니다.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) < len(rb) {
		ra, rb = rb, ra
	}
	// rb가 더 짧거나 같은 길이
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr := make([]int, len(rb)+1)
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev = curr
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// suggestCommands input 과 유사한 커맨드 이름 목록을 반환합니다.
// 편집 거리 임계값: min(2, len(input)/2)
func suggestCommands(input string, cmdMap map[string]*Command) []string {
	if len(input) == 0 {
		return nil
	}
	threshold := len(input) / 2
	if threshold > 2 {
		threshold = 2
	}
	if threshold == 0 {
		threshold = 1
	}

	// 중복 제거를 위해 커맨드 이름 집합 구성
	seen := make(map[string]bool)
	var suggestions []string
	for _, cmd := range cmdMap {
		name := cmd.Name()
		if seen[name] {
			continue
		}
		seen[name] = true
		if levenshtein(input, name) <= threshold {
			suggestions = append(suggestions, name)
		}
	}
	return suggestions
}
