package config

import "testing"

func FuzzLightweightParsers(f *testing.F) {
	f.Add("a: 1\nb:\n  - x\n")
	f.Add("port = 8080\n[db]\nname = \"app\"\n")
	f.Add("key=value\n[sec]\na=1\n")

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = parseYAMLContent(input, false)
		_, _ = parseTOMLContent(input, false)
		_, _ = parseINIContent(input, false)
	})
}
