package wcli

import (
	"context"
	"errors"
	"fmt"

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
	Use   string
	Short string
	Long  string

	// 실행 훅: PreRun → Run → PostRun 순서로 호출됨
	PreRun  func(ctx *Context) error
	Run     func(ctx *Context) error
	PostRun func(ctx *Context) error

	// SilenceErrors true이면 Execute()에서 에러를 자동으로 출력하지 않음
	SilenceErrors bool

	// 플래그 관리 (명령어별 고유의 플래그 셋 보유)
	flags *FlagSet

	commands []*Command
	parent   *Command
}

// Flags 커맨드의 플래그 셋을 반환 (초기화 보장)
func (c *Command) Flags() *FlagSet {
	if c.flags == nil {
		c.flags = NewFlagSet()
	}
	return c.flags
}

// AddCommand 하위 명령어를 추가
func (c *Command) AddCommand(cmds ...*Command) {
	for _, sub := range cmds {
		if sub == c {
			panic("Command can't be a child of itself")
		}
		sub.parent = c
		c.commands = append(c.commands, sub)
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
			rich.Println("[red][bold]Error:[/bold] %s[/red]", err.Error())
		}
	}
	return err
}

func (c *Command) execute(ctx *Context) error {
	// 1. 하위 명령어가 있는지 확인 (가장 먼저 확인)
	if len(ctx.Args) > 0 {
		subCmdName := ctx.Args[0]
		for _, sub := range c.commands {
			if sub.Use == subCmdName {
				ctx.Args = ctx.Args[1:]
				return sub.execute(ctx)
			}
		}
	}

	// 2. -h, --help 플래그 확인 (서브 커맨드 라우팅 이후)
	// 참고: --help가 다른 플래그의 값으로 쓰이더라도 도움말이 우선 출력됩니다.
	// 이는 표준 CLI 동작과 일치합니다.
	for _, arg := range ctx.Args {
		if arg == "-h" || arg == "--help" {
			c.Help()
			return ErrHelp
		}
	}

	// 3. 현재 명령어의 플래그 파싱
	if c.flags != nil {
		remainingArgs, err := c.flags.Parse(ctx.Args)
		if err != nil {
			return err
		}
		ctx.Args = remainingArgs
	}

	// 4. PreRun 훅 실행
	if c.PreRun != nil {
		if err := c.PreRun(ctx); err != nil {
			return err
		}
	}

	// 5. 명령어 실행
	if c.Run != nil {
		if err := c.Run(ctx); err != nil {
			return err
		}
	} else if len(c.commands) > 0 {
		// Run이 없고 하위 명령어가 있는 경우 도움말 출력
		c.Help()
		return ErrHelp
	} else {
		return fmt.Errorf("명령어 실행 함수가 정의되지 않음: %s", c.Use)
	}

	// 6. PostRun 훅 실행
	if c.PostRun != nil {
		if err := c.PostRun(ctx); err != nil {
			return err
		}
	}

	return nil
}
