package producer

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	verisinputs "github.com/JochiRaider/cartulary/referencepacks/veris/inputs"
)

type verisRejectingFetcher struct{ calls int }

func (f *verisRejectingFetcher) FetchVERIS(context.Context, VERISSource) (VERISDownload, error) {
	f.calls++
	return VERISDownload{Bytes: []byte("changed"), RetrievedAt: "2026-11-07T00:00:00Z", ContentType: "text/plain"}, nil
}

func TestVERISSourceAdmission_Unit(t *testing.T) {
	b, err := verisinputs.Files.ReadFile("inventory.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := verisInventory(b)
	if err != nil || len(inventory.Sources) != 11 {
		t.Fatal("immutable inventory", err)
	}
	for _, bad := range [][]byte{append(append([]byte{}, b...), '\n'), []byte(strings.Replace(string(b), "1.4.2", "1.4.3", 1)), []byte(`{"schema_id":"cartulary.veris_inventory.v1","sources":[]}`)} {
		if _, err := verisInventory(bad); err == nil {
			t.Fatal("changed/incomplete input admitted")
		}
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	fetch := &verisRejectingFetcher{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := VERISAcquire(ctx, root, b, fetch); !errors.Is(err, context.Canceled) || fetch.calls != 0 {
		t.Fatal("cancelled acquisition performed IO", err)
	}
	if _, err := VERISAcquire(context.Background(), root, b, fetch); err == nil {
		t.Fatal("changed source accepted")
	}
	if _, err := root.Stat("source-lock.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("failed acquisition published lock")
	}
	if _, err := VERISCheck(root, b); err == nil {
		t.Fatal("missing retained source admitted")
	}
	bad := inventory.Sources[0]
	bad.URL = "https://example.test/unapproved"
	if _, err := (VERISHTTPFetcher{}).FetchVERIS(context.Background(), bad); err == nil {
		t.Fatal("unlisted network source admitted")
	}
	for _, media := range []string{"text/html", "application/json"} {
		if err := verisDownloadValid(inventory.Sources[0], VERISDownload{[]byte("wrong"), "2026-11-07T00:00:00Z", media}); err == nil {
			t.Fatal("untrusted receipt admitted")
		}
	}
}
