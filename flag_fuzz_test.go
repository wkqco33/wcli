package wcli

import "testing"

func FuzzFlagSetParse(f *testing.F) {
	f.Add("--name", "alice", "--count", "3", "-v")
	f.Add("--", "--not-flag", "", "", "")
	f.Add("-vc", "10", "", "", "")

	f.Fuzz(func(t *testing.T, a, b, c, d, e string) {
		var (
			name  string
			count int
			verb  bool
		)
		fs := NewFlagSet()
		fs.StringVar(&name, "name", "n", "", "name")
		fs.IntVar(&count, "count", "c", 0, "count")
		fs.BoolVar(&verb, "verbose", "v", false, "verbose")

		args := []string{a, b, c, d, e}
		_, _ = fs.Parse(args)
		_ = fs.Validate()
	})
}
