//go:build linux

package performancefixturelifecycle

import (
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerformanceFixtureObjectStorageAdmission(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name         string
		bytes, block int64
		available    uint64
		valid        bool
	}{
		{"empty", 0, 4096, 0, true},
		{"exact", 4096, 4096, 1, true},
		{"partial block", 4097, 4096, 1, false},
		{"memory backed capacity", 20971520000, 4096, (16 << 30) / 4096, false},
		{"disk capacity", 20971520000, 4096, (64 << 30) / 4096, true},
		{"counter multiplication overflow", math.MaxInt64, 4096, math.MaxUint64, true},
		{"unknown block size", 1, 0, 1, false},
		{"negative population", -1, 4096, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateObjectSnapshotCapacity(tc.bytes, tc.block, tc.available); (err == nil) != tc.valid {
				t.Fatalf("capacity acceptance = %v, want %v", err == nil, tc.valid)
			}
		})
	}
	root := t.TempDir()
	if err := checkObjectSnapshotCapacity(root, 1); err != nil {
		t.Fatal(err)
	}
	if err := checkObjectSnapshotCapacity(filepath.Join(root, "missing-private-path"), 1); err == nil || strings.Contains(err.Error(), root) || strings.Contains(err.Error(), "missing-private-path") {
		t.Fatal("unavailable storage capacity was accepted or exposed a private path")
	}
}
