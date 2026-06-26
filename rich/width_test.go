package rich

import "testing"

func TestDisplayWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"한글", 4},     // 한글 2자 = 4칸
		{"a한b", 4},    // 1 + 2 + 1
		{"포맷", 4},     // 라벨 예시
		{"프레임", 6},    // 한글 3자 = 6칸
		{"500 x 226", 9},
		{"🎬", 2},      // 이모지 = 2칸
		{"📄 t.png", 8}, // 2 + 1 + 5
	}
	for _, c := range cases {
		if got := DisplayWidth(c.in); got != c.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
