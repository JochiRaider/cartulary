package packformat

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// The complete fixture has a valid Base manifest, payload, original license,
// and an additional 256 MiB notice. It is streamed from a compact recipe: no
// test or production limit is increased and no 256 MiB allocation is needed.
func testCompleteArchiveFileByteBoundary(t *testing.T) {
	t.Helper()
	const limit = int64(268435456)
	const noticePath = "notices/boundary.txt"
	// Independently calculated SHA-256 of 268435455 ASCII 'a' bytes and LF.
	const noticeSHA = "3f77e4d9918931a1937694be0700e22f8a609fb5514f83f9f3ddce1de2185ce7"
	raw, err := os.ReadFile("../../../../../contracts/reference-packs/builtins/release.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Packs []struct {
			Members map[string]string `json:"members"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(raw, &release); err != nil {
		t.Fatal(err)
	}
	members := release.Packs[0].Members
	_, err = DecodeManifest([]byte(members["manifest.json"]), false)
	if err != nil {
		t.Fatal(err)
	}
	value, err := canonicaljson.DecodeStrict([]byte(members["manifest.json"]))
	if err != nil {
		t.Fatal(err)
	}
	object := value.(map[string]any)
	files := object["files"].([]any)
	files = append(files, map[string]any{"path": noticePath, "role": "notice", "media_type": "text/plain", "size_bytes": limit, "sha256": noticeSHA})
	slices.SortFunc(files, func(a, b any) int {
		return strings.Compare(a.(map[string]any)["path"].(string), b.(map[string]any)["path"].(string))
	})
	object["files"] = files
	license := object["license"].(map[string]any)
	license["notice_paths"] = []string{"notices/LICENSE.txt", noticePath}
	manifest, err := canonicaljson.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	m, err := DecodeManifest(manifest, false)
	if err != nil {
		t.Fatal(err)
	}
	members["manifest.json"] = string(manifest)
	names := []string{noticePath}
	for name := range members {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, size := range []int64{limit, limit + 1} {
		archive, err := os.CreateTemp(t.TempDir(), "boundary-*.tar")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = archive.Close() })
		writer := tar.NewWriter(archive)
		for _, name := range names {
			var source io.Reader = strings.NewReader(members[name])
			length := int64(len(members[name]))
			if name == noticePath {
				length = size
				source = boundaryNotice(size)
			}
			if err := writer.WriteHeader(&tar.Header{Name: name, Size: length, Mode: 0600, Format: tar.FormatUSTAR}); err != nil {
				t.Fatal(err)
			}
			if n, err := io.Copy(writer, source); err != nil || n != length {
				t.Fatal(n, err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		info, err := archive.Stat()
		if err != nil {
			t.Fatal(err)
		}
		destination := &countingDestination{}
		inventory, err := Extract(context.Background(), archive, info.Size(), DefaultArchiveLimits(), destination)
		if size > limit {
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != "archive_structure_invalid" || destination.members != 0 {
				t.Fatal("one-over file published", err, destination.members)
			}
			continue
		}
		if err != nil {
			t.Fatal("at-limit archive", err)
		}
		if inventory[noticePath].SHA256 != noticeSHA || inventory[noticePath].Size != limit || destination.largestWrite > 65536 {
			t.Fatal("at-limit digest, size or bounded extraction", inventory[noticePath], destination.largestWrite)
		}
		if err := ValidateInventory(m, inventory, false); err != nil {
			t.Fatal(err)
		}
		if err := ValidateNoticeStream(context.Background(), boundaryNotice(size)); err != nil {
			t.Fatal("at-limit notice", err)
		}
		var rows contentRows
		if err := ValidateContent(context.Background(), m, contentMemory(members), &rows); err != nil || len(rows) == 0 {
			t.Fatal("complete profile", err)
		}
		if err := archive.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Clean(archive.Name())); err != nil {
			t.Fatal(err)
		}
	}
}

type repeatedASCII struct{}

func (repeatedASCII) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'a'
	}
	return len(p), nil
}
func boundaryNotice(size int64) io.Reader {
	return io.MultiReader(io.LimitReader(repeatedASCII{}, size-1), bytes.NewReader([]byte{'\n'}))
}

type countingDestination struct {
	members      int
	largestWrite int
}

func (d *countingDestination) Write(_ context.Context, _ string, write func(io.Writer) error) error {
	d.members++
	return write(countingMember{d})
}

type countingMember struct{ destination *countingDestination }

func (w countingMember) Write(p []byte) (int, error) {
	if len(p) > w.destination.largestWrite {
		w.destination.largestWrite = len(p)
	}
	return len(p), nil
}
