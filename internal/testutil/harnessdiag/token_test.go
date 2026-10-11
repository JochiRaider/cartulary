package harnessdiag

import (
	"regexp"
	"strings"
	"testing"
)

// The independent regexp oracle covers input punctuation, every ASCII byte,
// UTF-8 rejection and length boundaries without reusing the byte predicates.
func TestDiagnosticTokens(t *testing.T) {
	for _, c := range []struct {
		pattern    string
		maximum    int
		underscore bool
	}{
		{`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`, 128, false},
		{`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,254}$`, 255, true},
	} {
		reference := regexp.MustCompile(c.pattern)
		values := []string{"", "a", "_", ".", "-", "a.b_c-d", "é", "😀", "a\n"}
		for code := 0; code < 128; code++ {
			values = append(values, string(rune(code)), "a"+string(rune(code)))
		}
		for _, n := range []int{c.maximum - 1, c.maximum, c.maximum + 1} {
			values = append(values, strings.Repeat("a", n))
		}
		for _, value := range values {
			if validToken(value, c.maximum, c.underscore) != reference.MatchString(value) {
				t.Fatalf("token mismatch: %q", value)
			}
		}
	}
}
