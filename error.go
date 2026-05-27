package wcli

import "fmt"

// FlagError 플래그 파싱 또는 존재하지 않는 플래그 등 플래그 입력 구문 오류를 나타냅니다.
type FlagError struct {
	FlagName    string
	CommandName string
	Err         error
}

func (e *FlagError) Error() string {
	if e.CommandName != "" {
		return fmt.Sprintf("command %q flag --%s: %v", e.CommandName, e.FlagName, e.Err)
	}
	return fmt.Sprintf("flag --%s: %v", e.FlagName, e.Err)
}

func (e *FlagError) Unwrap() error {
	return e.Err
}

// ValidationError 필수 플래그 누락 또는 사용자 정의 검증(SetValidation) 실패 오류를 나타냅니다.
type ValidationError struct {
	FlagName    string
	CommandName string
	Err         error
}

func (e *ValidationError) Error() string {
	if e.CommandName != "" {
		return fmt.Sprintf("command %q flag --%s validation failed: %v", e.CommandName, e.FlagName, e.Err)
	}
	return fmt.Sprintf("flag --%s validation failed: %v", e.FlagName, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// CommandError 명령어 자체의 실행(예: Run 함수 부재) 또는 훅 실행 오류를 나타냅니다.
type CommandError struct {
	CommandName string
	Err         error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command %q execution error: %v", e.CommandName, e.Err)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}
