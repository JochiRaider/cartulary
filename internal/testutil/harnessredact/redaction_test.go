package harnessredact

import (
	"strings"
	"testing"
)

func TestStringUsesSharedHarnessRules(t *testing.T) {
	input := strings.Join([]string{
		"postgres://cartulary:supersecret@127.0.0.1:5432/postgres?sslmode=disable",
		"password=supersecret",
		"Authorization: Bearer abc.def",
		"-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
	}, "\n")

	got := String(input)
	for _, secret := range []string{"supersecret", "abc.def", "BEGIN PRIVATE KEY", "END PRIVATE KEY"} {
		if strings.Contains(got, secret) {
			t.Fatalf("redacted output still contains %q: %s", secret, got)
		}
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("redacted output must contain replacement marker: %s", got)
	}
}
