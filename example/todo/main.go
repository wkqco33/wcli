// todo는 wcli 라이브러리 사용법을 보여주는 예제 CLI 앱입니다.
// 간단한 할 일(Todo) 관리 도구로, 주요 기능을 모두 시연합니다:
//   - 루트 커맨드 + 서브커맨드 구성
//   - 로컬/Persistent 플래그
//   - 필수 플래그 & 검증
//   - PreRun/PostRun 훅
//   - rich 마크업 출력
//   - 커스텀 에러 처리
//
// 실행 예시:
//
//	go run ./example/todo add --title "밥 먹기" --priority high
//	go run ./example/todo list --all
//	go run ./example/todo done 1
//	go run ./example/todo --help
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

// Todo 할 일 항목
type Todo struct {
	ID        int
	Title     string
	Priority  string
	Done      bool
	CreatedAt time.Time
}

// 간단한 인메모리 저장소
var todos = []*Todo{
	{ID: 1, Title: "wcli 예제 코드 작성", Priority: "high", Done: false, CreatedAt: time.Now().Add(-2 * time.Hour)},
	{ID: 2, Title: "README 읽기", Priority: "medium", Done: true, CreatedAt: time.Now().Add(-1 * time.Hour)},
	{ID: 3, Title: "커피 마시기", Priority: "low", Done: false, CreatedAt: time.Now()},
}

var nextID = 4

func main() {
	// --verbose는 모든 서브커맨드에서 공유되는 Persistent 플래그
	var verbose bool

	rootCmd := &wcli.Command{
		Use:     "todo",
		Short:   "간단한 할 일 관리 CLI",
		Long:    "wcli 라이브러리로 만든 할 일 관리 예제 앱입니다.\n서브커맨드로 할 일을 추가, 조회, 완료 처리할 수 있습니다.",
		Version: "1.0.0",

		// PersistentPreRun은 모든 서브커맨드 실행 전에 호출됨
		PersistentPreRun: func(ctx *wcli.Context) error {
			if verbose {
				rich.Println("[dim]─── verbose 모드 활성화 ───[/dim]")
			}
			return nil
		},
	}

	// --verbose / -v 플래그를 모든 서브커맨드에서 사용 가능하도록 Persistent 등록
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", "v", false, "상세 출력 활성화")

	// 서브커맨드 등록
	rootCmd.AddCommand(
		buildAddCmd(&verbose),
		buildListCmd(&verbose),
		buildDoneCmd(),
	)

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

// buildAddCmd "add" 서브커맨드를 생성합니다.
func buildAddCmd(verbose *bool) *wcli.Command {
	var (
		title    string
		priority string
		dueIn    time.Duration
	)

	cmd := &wcli.Command{
		Use:   "add",
		Short: "새 할 일 추가",
		Long:  "새로운 할 일 항목을 목록에 추가합니다.",

		PreRun: func(ctx *wcli.Context) error {
			if *verbose {
				rich.Println("[dim]입력값 검증 중...[/dim]")
			}
			return nil
		},

		Run: func(ctx *wcli.Context) error {
			todo := &Todo{
				ID:        nextID,
				Title:     title,
				Priority:  priority,
				Done:      false,
				CreatedAt: time.Now(),
			}
			todos = append(todos, todo)
			nextID++

			rich.Println("[bold][green]✓ 할 일이 추가되었습니다![/green][/bold]")
			rich.Println("  ID      : [cyan]%d[/cyan]", todo.ID)
			rich.Println("  제목    : %s", todo.Title)
			rich.Println("  우선순위: %s", priorityColored(todo.Priority))
			if dueIn > 0 {
				deadline := time.Now().Add(dueIn)
				rich.Println("  마감    : %s 후 (%s)", dueIn, deadline.Format("15:04:05"))
			}
			return nil
		},

		PostRun: func(ctx *wcli.Context) error {
			if *verbose {
				rich.Println("[dim]현재 총 %d개의 할 일이 있습니다.[/dim]", len(todos))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "t", "", "할 일 제목")
	cmd.Flags().StringVar(&priority, "priority", "p", "medium", "우선순위 (low/medium/high)")
	cmd.Flags().DurationVar(&dueIn, "due", "d", 0, "마감까지 남은 시간 (예: 30m, 2h)")

	// --title은 필수 플래그
	_ = cmd.Flags().MarkRequired("title")

	// --priority 값 검증
	_ = cmd.Flags().SetValidation("priority", func(val string) error {
		switch val {
		case "low", "medium", "high":
			return nil
		default:
			return fmt.Errorf("'%s'는 유효하지 않습니다. low/medium/high 중 하나를 선택하세요", val)
		}
	})

	return cmd
}

// buildListCmd "list" 서브커맨드를 생성합니다.
func buildListCmd(verbose *bool) *wcli.Command {
	var showAll bool

	cmd := &wcli.Command{
		Use:     "list",
		Short:   "할 일 목록 조회",
		Aliases: []string{"ls", "l"},

		Run: func(ctx *wcli.Context) error {
			filtered := make([]*Todo, 0, len(todos))
			for _, t := range todos {
				if showAll || !t.Done {
					filtered = append(filtered, t)
				}
			}

			if len(filtered) == 0 {
				rich.Println("[yellow]표시할 할 일이 없습니다.[/yellow]")
				return nil
			}

			rich.Println("[bold][yellow]── 할 일 목록 ──[/yellow][/bold]")
			for _, t := range filtered {
				status := "[red][ ][/red]"
				if t.Done {
					status = "[green][✓][/green]"
				}
				rich.Println("  %s [cyan]#%d[/cyan] %s  %s", status, t.ID, priorityColored(t.Priority), t.Title)
			}

			if *verbose {
				done := 0
				for _, t := range todos {
					if t.Done {
						done++
					}
				}
				fmt.Println()
				rich.Println("[dim]전체 %d개 중 %d개 완료[/dim]", len(todos), done)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", "a", false, "완료된 항목도 함께 표시")
	return cmd
}

// buildDoneCmd "done" 서브커맨드를 생성합니다.
func buildDoneCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "done [id]",
		Short: "할 일을 완료 처리",

		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("ID를 입력해주세요. 예: todo done 1")
			}

			id, err := strconv.Atoi(ctx.Args[0])
			if err != nil {
				return fmt.Errorf("유효하지 않은 ID: %s", ctx.Args[0])
			}

			for _, t := range todos {
				if t.ID == id {
					if t.Done {
						rich.Println("[yellow]이미 완료된 항목입니다: #%d %s[/yellow]", t.ID, t.Title)
						return nil
					}
					t.Done = true
					rich.Println("[bold][green]✓ 완료 처리되었습니다![/green][/bold]")
					rich.Println("  [cyan]#%d[/cyan] %s", t.ID, t.Title)
					return nil
				}
			}

			return fmt.Errorf("ID %d를 찾을 수 없습니다", id)
		},
	}
}

// priorityColored 우선순위 문자열에 색상 마크업을 적용합니다.
func priorityColored(p string) string {
	switch p {
	case "high":
		return rich.Markup("[red][bold]high[/bold][/red]")
	case "medium":
		return rich.Markup("[yellow]medium[/yellow]")
	default:
		return rich.Markup("[dim]low[/dim]")
	}
}
