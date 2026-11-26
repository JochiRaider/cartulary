//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// fakeProducer builds a binary and its receipt from content; its key is the
// content's digest.
type fakeProducer struct {
	content string
	builds  int
}

func (p *fakeProducer) ID() string { return "fake" }

func (p *fakeProducer) Key() (string, error) {
	sum := sha256.Sum256([]byte(p.content))
	return hex.EncodeToString(sum[:]), nil
}

func (p *fakeProducer) Build(staging string) error {
	p.builds++
	if err := os.WriteFile(filepath.Join(staging, "tool"), []byte(p.content), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(staging, "tool.crypto.json"), []byte("receipt:"+p.content), 0o600)
}

func fixture(t *testing.T) (Store, map[string]string) {
	t.Helper()
	root := t.TempDir()
	t.Cleanup(func() { _ = removeTree(root) })
	views := filepath.Join(root, "build", "bin")
	return Store{Root: filepath.Join(root, "store")}, map[string]string{
		"tool": filepath.Join(views, "tool"), "tool.crypto.json": filepath.Join(views, "tool.crypto.json"),
	}
}

func produce(t *testing.T, s Store, p Producer, views map[string]string, controls Controls) Result {
	t.Helper()
	result, err := s.Produce(p, views, controls)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// A root is built once, sealed read-only and reused while its key holds. Its
// views are single-link copies, since product code opens only single-link
// files, and a reuse leaves a current view untouched.
func TestProduceSealsReusesAndPublishesViews(t *testing.T) {
	s, views := fixture(t)
	p := &fakeProducer{content: "one"}
	first := produce(t, s, p, views, Controls{})
	published, err := os.Stat(views["tool"])
	if err != nil {
		t.Fatal(err)
	}
	second := produce(t, s, p, views, Controls{})
	if first.Reused || !second.Reused || first.Root != second.Root || p.builds != 1 {
		t.Fatalf("first %+v second %+v builds %d", first, second, p.builds)
	}
	if again, err := os.Stat(views["tool"]); err != nil || !os.SameFile(published, again) {
		t.Fatal("a reuse replaced a current view", err)
	}
	for source, view := range views {
		rootInfo, err := os.Stat(filepath.Join(s.RootPath(first.Root), source))
		if err != nil || rootInfo.Mode().Perm()&0o222 != 0 {
			t.Fatalf("root file %s is not sealed: %v", source, err)
		}
		var stat unix.Stat_t
		if err := unix.Stat(view, &stat); err != nil || stat.Nlink != 1 {
			t.Fatalf("view %s is not a single-link file: %v", view, err)
		}
		if viewInfo, err := os.Stat(view); err != nil || viewInfo.Mode().Perm() != rootInfo.Mode().Perm()|0o200 {
			t.Fatalf("view %s mode %v, %v", view, viewInfo.Mode(), err)
		}
	}
	// A removed view is republished from the root without a build.
	if err := os.Remove(views["tool"]); err != nil {
		t.Fatal(err)
	}
	if result := produce(t, s, p, views, Controls{}); !result.Reused || p.builds != 1 {
		t.Fatalf("republish %+v builds %d", result, p.builds)
	}
	if data, err := os.ReadFile(views["tool"]); err != nil || string(data) != "one" {
		t.Fatalf("republished view %q, %v", data, err)
	}
}

// A changed key builds a new root, and both views move to it together.
func TestChangedKeyPublishesItsNewRoot(t *testing.T) {
	s, views := fixture(t)
	p := &fakeProducer{content: "one"}
	first := produce(t, s, p, views, Controls{})
	p.content = "two"
	second := produce(t, s, p, views, Controls{})
	if second.Reused || second.Root == first.Root || p.builds != 2 {
		t.Fatalf("first %+v second %+v builds %d", first, second, p.builds)
	}
	for view, want := range map[string]string{views["tool"]: "two", views["tool.crypto.json"]: "receipt:two"} {
		if data, err := os.ReadFile(view); err != nil || string(data) != want {
			t.Fatalf("view %s = %q, %v; want %q", view, data, err, want)
		}
	}
}

// A root altered after sealing never supplies an artifact: the next produce
// rebuilds it and quarantines the altered copy.
func TestAlteredRootIsRebuiltAndQuarantined(t *testing.T) {
	s, views := fixture(t)
	p := &fakeProducer{content: "one"}
	first := produce(t, s, p, views, Controls{})
	sealed := filepath.Join(s.RootPath(first.Root), "tool")
	if err := os.Chmod(sealed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sealed, []byte("altered"), 0o755); err != nil {
		t.Fatal(err)
	}
	second := produce(t, s, p, views, Controls{})
	if second.Reused || second.Root != first.Root || p.builds != 2 {
		t.Fatalf("second %+v builds %d", second, p.builds)
	}
	if data, err := os.ReadFile(views["tool"]); err != nil || string(data) != "one" {
		t.Fatalf("restored view %q, %v", data, err)
	}
	quarantined, _ := filepath.Glob(filepath.Join(s.Root, "objects", ".quarantine-"+first.Root+"-*"))
	if len(quarantined) != 1 || !s.verify(first.Root) {
		t.Fatalf("quarantine %v, root verifies %v", quarantined, s.verify(first.Root))
	}
}

// A view is published only from bytes with the manifest's digest, and every
// view names a file of its root.
func TestPublishRefusesUnverifiedBytes(t *testing.T) {
	s, views := fixture(t)
	first := produce(t, s, &fakeProducer{content: "one"}, views, Controls{})
	target := filepath.Join(t.TempDir(), "copy")
	if err := copyFile(filepath.Join(s.RootPath(first.Root), "tool"), target, 0o755, strings.Repeat("0", 64)); err == nil {
		t.Fatal("a copy with another digest was accepted")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("a refused copy remained", err)
	}
	if err := s.publish(first.Root, map[string]string{"missing": filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatal("a view naming no root file was accepted")
	}
}

// Disable neither reads nor records keys; Force rebuilds and records.
func TestCacheControls(t *testing.T) {
	s, views := fixture(t)
	p := &fakeProducer{content: "one"}
	produce(t, s, p, views, Controls{Disable: true})
	produce(t, s, p, views, Controls{Disable: true})
	if records, _ := filepath.Glob(filepath.Join(s.Root, "keys", "*", "*.json")); p.builds != 2 || len(records) != 0 {
		t.Fatalf("disabled builds %d records %v", p.builds, records)
	}
	produce(t, s, p, views, Controls{Force: true})
	if result := produce(t, s, p, views, Controls{}); !result.Reused || p.builds != 3 {
		t.Fatalf("after force %+v builds %d", result, p.builds)
	}
}

// Collection removes records unused for the age, then roots no record names,
// within its limit. A collected root's views are copies and stay readable.
func TestCollect(t *testing.T) {
	s, views := fixture(t)
	p := &fakeProducer{content: "one"}
	kept := produce(t, s, p, views, Controls{})
	other, otherViews := fixture(t)
	other.Root = s.Root
	p.content = "two"
	orphan := produce(t, other, p, otherViews, Controls{})
	key, err := p.Key()
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(s.recordPath(p.ID(), key), old, old); err != nil {
		t.Fatal(err)
	}
	// A freshly sealed root outlives its record for the grace period.
	if removed, err := s.Collect(time.Now(), time.Hour, 16); err != nil || removed != 1 || !s.verify(orphan.Root) {
		t.Fatalf("collection within the grace removed %d, %v", removed, err)
	}
	if err := os.Chtimes(filepath.Join(s.Root, "objects", orphan.Root+".json"), old, old); err != nil {
		t.Fatal(err)
	}
	if removed, err := s.Collect(time.Now(), time.Hour, 0); err != nil || removed != 0 || !s.verify(orphan.Root) {
		t.Fatalf("bounded collection removed %d, %v", removed, err)
	}
	if removed, err := s.Collect(time.Now(), time.Hour, 16); err != nil || removed != 1 {
		t.Fatalf("collection after the grace removed %d, %v", removed, err)
	}
	if _, err := os.Stat(s.RootPath(orphan.Root)); !os.IsNotExist(err) {
		t.Fatal("an unreferenced root survived", err)
	}
	if !s.verify(kept.Root) {
		t.Fatal("a recorded root was collected")
	}
	if data, err := os.ReadFile(otherViews["tool"]); err != nil || string(data) != "two" {
		t.Fatalf("a collected root's view %q, %v", data, err)
	}
	if entries, _ := os.ReadDir(filepath.Join(s.Root, "objects")); len(entries) != 2 || !strings.HasPrefix(entries[0].Name(), kept.Root) {
		t.Fatalf("objects %v", entries)
	}
}
