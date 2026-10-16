package performancefixturelifecycle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestPerformanceFixtureObjectSnapshotIntegrity(t *testing.T) {
	t.Run("proof fields are explicit even for empty populations", func(t *testing.T) {
		for _, raw := range []string{
			`{"manifest_sha256":"sha256:` + strings.Repeat("a", 64) + `"}`,
			`{"manifest_sha256":"sha256:` + strings.Repeat("a", 64) + `","object_count":null,"total_bytes":0}`,
			`{"manifest_sha256":"sha256:` + strings.Repeat("a", 64) + `","object_count":0,"object_count":0,"total_bytes":0}`,
			`{"manifest_sha256":"sha256:` + strings.Repeat("a", 64) + `","object_count":0,"total_bytes":1}`,
		} {
			var proof ObjectSnapshotSummary
			if err := json.Unmarshal([]byte(raw), &proof); err == nil {
				t.Fatal("ambiguous object proof admitted")
			}
		}
	})
	ctx := context.Background()
	key := strings.Repeat("a", 64)
	newSource := func(t *testing.T) (string, objectstore.Store, ObjectSnapshotSummary) {
		t.Helper()
		root := t.TempDir()
		store, err := objectstore.NewFilesystemStore(filepath.Join(root, objectSnapshotDirectory))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = store.Close() })
		for _, item := range []struct{ key, body, kind string }{{"blob/a", "first", "text/plain"}, {"blob/b", "second", "image/png"}, {"empty", "", "application/octet-stream"}} {
			if err := store.PutObject(ctx, item.key, strings.NewReader(item.body), int64(len(item.body)), item.kind); err != nil {
				t.Fatal(err)
			}
		}
		proof, err := sealObjectSnapshot(ctx, root, key, store)
		if err != nil {
			t.Fatal(err)
		}
		return root, store, proof
	}
	newDestination := func(t *testing.T) objectstore.Store {
		t.Helper()
		store, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = store.Close() })
		return store
	}
	t.Run("sealed payloads restore into distinct verified namespaces", func(t *testing.T) {
		root, source, proof := newSource(t)
		if proof.ObjectCount != 3 || proof.TotalBytes != 11 || !validObjectSnapshotSummary(proof) {
			t.Fatalf("invalid proof: %#v", proof)
		}
		first, second := newDestination(t), newDestination(t)
		second = snapshotListingStore{Store: second}
		for _, destination := range []objectstore.Store{first, second} {
			if err := restoreObjectSnapshot(ctx, root, key, proof, destination); err != nil {
				t.Fatal(err)
			}
			body, info, err := destination.ReadObject(ctx, "blob/b", objectstore.ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			value, err := io.ReadAll(body)
			_ = body.Close()
			if err != nil || string(value) != "second" || info.Size != 6 || info.ContentType != "image/png" {
				t.Fatal("restored bytes or metadata diverged")
			}
		}
		if err := first.PutObject(ctx, "clone-only", strings.NewReader("private"), 7, "text/plain"); err != nil {
			t.Fatal(err)
		}
		for _, isolated := range []objectstore.Store{source, second} {
			if _, err := isolated.StatObject(ctx, "clone-only"); err == nil {
				t.Fatal("clone mutation crossed namespace")
			}
		}
		if _, err := sealObjectSnapshot(ctx, root, key, source); err == nil {
			t.Fatal("sealed manifest was overwritten")
		}
	})
	t.Run("empty snapshots are explicit and restorable", func(t *testing.T) {
		root := t.TempDir()
		source, err := objectstore.NewFilesystemStore(filepath.Join(root, objectSnapshotDirectory))
		if err != nil {
			t.Fatal(err)
		}
		defer source.Close()
		proof, err := sealObjectSnapshot(ctx, root, key, source)
		if err != nil || proof.ObjectCount != 0 || proof.TotalBytes != 0 {
			t.Fatalf("empty proof: %#v, %v", proof, err)
		}
		if err := restoreObjectSnapshot(ctx, root, key, proof, newDestination(t)); err != nil {
			t.Fatal(err)
		}
	})
	for _, change := range []string{"payload bytes", "missing payload", "extra payload", "metadata", "manifest", "symlink", "wrong key", "wrong proof", "nonempty destination", "cancellation", "failed transfer", "corrupt destination"} {
		t.Run(change, func(t *testing.T) {
			root, source, proof := newSource(t)
			destination := newDestination(t)
			active := ctx
			selectedKey := key
			switch change {
			case "payload bytes":
				if err := source.PutObject(ctx, "blob/a", strings.NewReader("other"), 5, "text/plain"); err != nil {
					t.Fatal(err)
				}
			case "missing payload":
				if err := source.DeleteObject(ctx, "blob/a"); err != nil {
					t.Fatal(err)
				}
			case "extra payload":
				if err := source.PutObject(ctx, "extra", strings.NewReader("extra"), 5, "text/plain"); err != nil {
					t.Fatal(err)
				}
			case "metadata":
				if err := source.PutObject(ctx, "blob/a", strings.NewReader("first"), 5, "image/png"); err != nil {
					t.Fatal(err)
				}
			case "manifest":
				if err := os.WriteFile(filepath.Join(root, objectSnapshotFile), []byte("{}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				payload := filepath.Join(root, objectSnapshotDirectory, "blob/a")
				if err := os.Remove(payload); err != nil {
					t.Fatal(err)
				}
				outside := filepath.Join(t.TempDir(), "outside")
				if err := os.WriteFile(outside, []byte("first"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, payload); err != nil {
					t.Fatal(err)
				}
			case "wrong key":
				selectedKey = strings.Repeat("b", 64)
			case "wrong proof":
				proof.TotalBytes++
			case "nonempty destination":
				if err := destination.PutObject(ctx, "preexisting", strings.NewReader("keep"), 4, "text/plain"); err != nil {
					t.Fatal(err)
				}
			case "cancellation":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				active = cancelled
			case "failed transfer":
				destination = snapshotFaultStore{Store: destination, failPut: true}
			case "corrupt destination":
				destination = snapshotFaultStore{Store: destination, corruptPut: true}
			}
			if err := restoreObjectSnapshot(active, root, selectedKey, proof, destination); err == nil {
				t.Fatal("invalid clone was accepted")
			}
			if change == "nonempty destination" {
				entries, err := destination.ListObjects(ctx, "")
				if err != nil || len(entries) != 1 || entries[0].Key != "preexisting" {
					t.Fatal("borrowed destination changed")
				}
			}
		})
	}
	t.Run("closed manifest grammar and reference boundaries", func(t *testing.T) {
		root, _, proof := newSource(t)
		original, err := os.ReadFile(filepath.Join(root, objectSnapshotFile))
		if err != nil {
			t.Fatal(err)
		}
		for _, change := range []string{"unknown", "duplicate", "trailing", "retired", "null population", "unsorted", "duplicate key", "traversal", "absolute", "reserved", "negative", "too large", "bad digest", "control content type"} {
			t.Run(change, func(t *testing.T) {
				var manifest objectSnapshotManifest
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatal(err)
				}
				switch change {
				case "retired":
					manifest.SchemaID = "cartulary.performance_fixture_objects.v0"
				case "null population":
					manifest.Objects = nil
				case "unsorted":
					manifest.Objects[0], manifest.Objects[1] = manifest.Objects[1], manifest.Objects[0]
				case "duplicate key":
					manifest.Objects[1].Key = manifest.Objects[0].Key
				case "traversal":
					manifest.Objects[0].Key = "../escape"
				case "absolute":
					manifest.Objects[0].Key = "/escape"
				case "reserved":
					manifest.Objects[0].Key = "a.meta.json"
				case "negative":
					manifest.Objects[0].ByteSize = -1
				case "too large":
					manifest.Objects[0].ByteSize = maxSnapshotObjectBytes + 1
				case "bad digest":
					manifest.Objects[0].SHA256 = "sha256:bad"
				case "control content type":
					manifest.Objects[0].ContentType = "text/plain\ninjected"
				}
				raw, err := canonicaljson.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				raw = append(raw, '\n')
				if change == "unknown" {
					raw = bytes.Replace(raw, []byte(`"objects":`), []byte(`"unknown":true,"objects":`), 1)
				}
				if change == "duplicate" {
					raw = bytes.Replace(raw, []byte(`"snapshot_key":`), []byte(`"snapshot_key":"`+key+`","snapshot_key":`), 1)
				}
				if change == "trailing" {
					raw = append(raw, []byte("{}")...)
				}
				if err := os.WriteFile(filepath.Join(root, objectSnapshotFile), raw, 0o600); err != nil {
					t.Fatal(err)
				}
				// Even a caller-supplied matching byte digest cannot admit an invalid
				// schema, ambiguous JSON, unsafe key or unsupported resource bound.
				digest := sha256.Sum256(raw)
				candidate := proof
				candidate.ManifestSHA256 = "sha256:" + hex.EncodeToString(digest[:])
				if _, err := readObjectSnapshot(root, key, candidate); err == nil {
					t.Fatal("invalid manifest admitted")
				}
			})
		}
	})
}

type snapshotFaultStore struct {
	objectstore.Store
	failPut    bool
	corruptPut bool
}

func (store snapshotFaultStore) PutObject(ctx context.Context, key string, body io.Reader, size int64, kind string) error {
	if store.failPut {
		return errors.New("injected transfer failure")
	}
	if store.corruptPut {
		if _, err := io.Copy(io.Discard, body); err != nil {
			return err
		}
		return store.Store.PutObject(ctx, key, strings.NewReader(strings.Repeat("x", int(size))), size, kind)
	}
	return store.Store.PutObject(ctx, key, body, size, kind)
}

// S3 listings omit content types; verified reads supply that metadata.
type snapshotListingStore struct{ objectstore.Store }

func (store snapshotListingStore) ListObjects(ctx context.Context, prefix string) ([]objectstore.ObjectInfo, error) {
	entries, err := store.Store.ListObjects(ctx, prefix)
	for i := range entries {
		entries[i].ContentType = ""
	}
	return entries, err
}
