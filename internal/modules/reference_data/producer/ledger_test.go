package producer

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestATTACKPublicationLedger(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := allocatePublication(dir, "fixture.20.0", Digest([]byte("same explicit inputs")))
			if err != nil || p.Sequence != 1 {
				t.Errorf("concurrent allocation: %+v %v", p, err)
			}
		}()
	}
	wg.Wait()
	if _, err := allocatePublication(dir, "fixture.20.0", Digest([]byte("changed"))); err == nil {
		t.Fatal("publication collision accepted")
	}
	p, err := allocatePublication(dir, "fixture.20.1", Digest([]byte("next")))
	if err != nil || p.Sequence != 2 {
		t.Fatal("sequence continuity", p, err)
	}
	if err = os.WriteFile(filepath.Join(dir, "0000000000000002.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = allocatePublication(dir, "third", Digest(nil)); err == nil {
		t.Fatal("tampered ledger admitted")
	}
}
