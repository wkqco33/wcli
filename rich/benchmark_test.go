package rich_test

import (
	"testing"

	"github.com/seoyc/wcli/rich"
)

func BenchmarkDisplayWidth_ASCII(b *testing.B) {
	s := "hello world this is a test string"
	for i := 0; i < b.N; i++ {
		rich.DisplayWidth(s)
	}
}

func BenchmarkDisplayWidth_CJK(b *testing.B) {
	s := "안녕하세요 이것은 한글 테스트 문자열입니다"
	for i := 0; i < b.N; i++ {
		rich.DisplayWidth(s)
	}
}

func BenchmarkDisplayWidth_Mixed(b *testing.B) {
	s := "hello 안녕 123 🌍 world"
	for i := 0; i < b.N; i++ {
		rich.DisplayWidth(s)
	}
}

func BenchmarkMarkup_Simple(b *testing.B) {
	s := "[red]에러 메시지[/red]"
	for i := 0; i < b.N; i++ {
		rich.Markup(s)
	}
}

func BenchmarkMarkup_Complex(b *testing.B) {
	s := "[bold][red]Error:[/red][/bold] [yellow]warning[/yellow] [green]ok[/green]"
	for i := 0; i < b.N; i++ {
		rich.Markup(s)
	}
}

func BenchmarkMarkup_TrueColor(b *testing.B) {
	s := "[#ff4500]주황[/#ff4500] [color(208)]256색[/color(208)]"
	for i := 0; i < b.N; i++ {
		rich.Markup(s)
	}
}

func BenchmarkTable_Render(b *testing.B) {
	tbl := rich.NewTable("이름", "나이", "직업", "지역")
	for i := 0; i < 100; i++ {
		tbl.AddRow("홍길동", "30", "개발자", "서울")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tbl.Render(blackhole{})
	}
}

type blackhole struct{}

func (blackhole) Write(p []byte) (int, error) { return len(p), nil }
