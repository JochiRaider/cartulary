// Package recoverytest supplies operation-scoped storage for Recovery tests.
package recoverytest

import (
	"os"
	"testing"
)

// ProofScratch lends a confined root until test cleanup, closing its handle
// before testing removes the owned temporary directory.
func ProofScratch(t testing.TB) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	return root
}
