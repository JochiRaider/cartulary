package producer

import (
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"os"
	"testing"
)

func encoded(t *testing.T, v any) []byte {
	t.Helper()
	b, err := canonicaljson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func testRoot(t *testing.T) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root
}
