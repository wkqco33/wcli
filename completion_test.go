package wcli_test

import (
	"strings"
	"testing"

	"github.com/seoyc/wcli"
)

func newTestRootForCompletion() *wcli.Command {
	root := &wcli.Command{Use: "myapp", Short: "테스트 앱"}

	deploy := &wcli.Command{Use: "deploy", Short: "배포 실행"}
	var env string
	deploy.Flags().StringVar(&env, "env", "e", "", "환경 (prod/staging)")
	var dryRun bool
	deploy.Flags().BoolVar(&dryRun, "dry-run", "", false, "실제 실행 안 함")

	root.AddCommand(deploy)
	root.AddCommand(&wcli.Command{Use: "status", Short: "상태 확인"})
	return root
}

func TestGenFishCompletion_Basic(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 앱 이름 포함
	if !strings.Contains(out, "complete -c myapp") {
		t.Error("앱 이름 'myapp' 이 출력에 없음")
	}
	// 서브커맨드 등록
	if !strings.Contains(out, "-a deploy") {
		t.Error("서브커맨드 'deploy' 가 출력에 없음")
	}
	if !strings.Contains(out, "-a status") {
		t.Error("서브커맨드 'status' 가 출력에 없음")
	}
}

func TestGenFishCompletion_SubFlags(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// deploy 서브커맨드 플래그
	if !strings.Contains(out, "__fish_seen_subcommand_from deploy") {
		t.Error("deploy 서브커맨드 조건이 출력에 없음")
	}
	if !strings.Contains(out, "-l env") {
		t.Error("--env 플래그가 출력에 없음")
	}
	if !strings.Contains(out, "-l dry-run") {
		t.Error("--dry-run 플래그가 출력에 없음")
	}
}

func TestGenFishCompletion_NotSeenCondition(t *testing.T) {
	root := newTestRootForCompletion()

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 루트 레벨 완성에 not __fish_seen_subcommand_from 조건 포함
	if !strings.Contains(out, "not __fish_seen_subcommand_from") {
		t.Error("not __fish_seen_subcommand_from 조건이 출력에 없음")
	}
}

func TestGenFishCompletion_EscapeDesc(t *testing.T) {
	root := &wcli.Command{Use: "app", Short: "앱"}
	root.AddCommand(&wcli.Command{Use: "run", Short: "it's a command"})

	var buf strings.Builder
	if err := wcli.GenFishCompletion(root, &buf); err != nil {
		t.Fatalf("GenFishCompletion 오류: %v", err)
	}
	out := buf.String()

	// 작은따옴표 이스케이프 확인
	if strings.Contains(out, "it's") {
		t.Error("작은따옴표가 이스케이프되지 않음")
	}
}
