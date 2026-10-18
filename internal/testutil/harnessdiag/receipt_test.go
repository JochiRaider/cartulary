package harnessdiag

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestReceiptPublicationFailure(t *testing.T) {
	for _, value := range []string{"16384.01 0.00\n", "32768.03 0.00\n", "123456.78 0.00\n"} {
		ms, err := parseBootMilliseconds(value)
		if err != nil || !safeInteger(ms, 0, 9007199254740991) {
			t.Fatalf("boot clock lost integer precision: %q %v %v", value, ms, err)
		}
	}
	if ms, err := parseBootMilliseconds("16384.01 0.00\n"); err != nil || ms != 16384010 {
		t.Fatalf("boot clock changed units: %v %v", ms, err)
	}
	for _, value := range []string{"", "-1.00", "1.001", "NaN", "9007199254741.00", "1e4.00"} {
		if _, err := parseBootMilliseconds(value); err == nil {
			t.Fatalf("accepted invalid boot clock %q", value)
		}
	}
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
