package rich

import (
	"strings"
	"testing"
	"time"
)

func TestProgressBar_UsesInjectableClock(t *testing.T) {
	oldNow := nowFunc
	current := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return current }
	defer func() { nowFunc = oldNow }()

	pb := NewProgressBar(100)
	pb.ShowETA = true
	pb.Start()

	current = current.Add(10 * time.Second)
	out := pb.Render(50)
	if !strings.Contains(out, "ETA: 10s") {
		t.Fatalf("고정 시계 기반 ETA가 예상과 다름: %q", out)
	}
}
