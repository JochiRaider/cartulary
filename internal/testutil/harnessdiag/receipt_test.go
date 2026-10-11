package harnessdiag

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestReceiptPublicationFailure(t *testing.T) {
	roots := []string{"/private/one", "/private/two", "/private/three"}
	payload := []byte("closed receipt\n")
	full := errors.New("disk full")
	calls := 0
	published := map[string][]byte{}
	err := publishClosedReceipts(roots, payload, func(file string, data []byte) error {
		calls++
		if calls == 2 {
			return full
		}
		published[file] = bytes.Clone(data)
		return nil
	})
	if !errors.Is(err, full) || calls != 2 || len(published) != 1 {
		t.Fatal("publication failure was swallowed or later member was written")
	}
	if !bytes.Equal(published[filepath.Join(roots[0], "diagnostics/terminal-receipt.json")], payload) {
		t.Fatal("partial publication changed closed payload")
	}
}
