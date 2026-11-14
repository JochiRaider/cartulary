package policy

import (
	"strings"
	"testing"
)

func TestRecoveryBlobMetadata(t *testing.T) {
	t.Parallel()
	size := func(v int64) *int64 { return &v }
	digest := func(v string) *string { return &v }
	sha := strings.Repeat("a", 64)
	for _, tc := range []struct {
		name          string
		declared      int64
		observed      *int64
		sha, expected *string
		valid         bool
	}{
		{"empty finalized object", 0, size(0), digest(sha), nil, true},
		{"exact metadata", 12, size(12), digest(sha), digest(sha), true},
		{"negative declaration", -1, size(-1), digest(sha), nil, false},
		{"missing observation", 12, nil, digest(sha), nil, false},
		{"size mismatch", 12, size(13), digest(sha), nil, false},
		{"missing digest", 12, size(12), nil, nil, false},
		{"empty digest", 12, size(12), digest(""), nil, false},
		{"short digest", 12, size(12), digest(strings.Repeat("a", 63)), nil, false},
		{"uppercase digest", 12, size(12), digest(strings.Repeat("A", 64)), nil, false},
		{"invalid hex", 12, size(12), digest(strings.Repeat("g", 64)), nil, false},
		{"prefixed digest", 12, size(12), digest("sha256:" + sha), nil, false},
		{"whitespace digest", 12, size(12), digest(" " + sha), nil, false},
		{"expected mismatch", 12, size(12), digest(sha), digest(strings.Repeat("b", 64)), false},
		{"empty expected", 12, size(12), digest(sha), digest(""), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidRecoveryBlobMetadata(tc.declared, tc.observed, tc.sha, tc.expected); got != tc.valid {
				t.Fatalf("metadata admission=%v want=%v", got, tc.valid)
			}
		})
	}
}
