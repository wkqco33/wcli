package wcli

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrCommandNotFound 하위 명령어를 찾을 수 없을 때 반환하는 에러
	ErrCommandNotFound = errors.New("command not found")
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

	Run func(ctx *Context) error

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
	return c.execute(ctx)
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

	// 2. 현재 명령어의 플래그 파싱 (하위 명령어가 아니면 이 명령어의 플래그로 파싱)
	if c.flags != nil {
		remainingArgs, err := c.flags.Parse(ctx.Args)
		if err != nil {
			return err
		}
		ctx.Args = remainingArgs
	}

	// 3. 명령어 실행
	if c.Run != nil {
		return c.Run(ctx)
	}

	return fmt.Errorf("명령어 실행 함수가 정의되지 않음: %s", c.Use)
}
