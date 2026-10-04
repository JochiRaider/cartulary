package packformat

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testCompleteCompressionRatio(t *testing.T, manifest Manifest, original contentMemory) {
	members := contentMemory{}
	for path, value := range original {
		members[path] = value
	}
	manifest.Files = slices.Clone(manifest.Files)
	notice := manifest.License.Notices[0]
	serialize := func(size int) int64 {
		members[notice] = strings.Repeat("a", size-1) + "\n"
		for i := range manifest.Files {
			if manifest.Files[i].Path == notice {
				manifest.Files[i].Size = int64(size)
				manifest.Files[i].SHA256 = Digest([]byte(members[notice]))
			}
		}
		raw, err := canonicaljson.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		members["manifest.json"] = string(raw)
		total := int64(0)
		for _, data := range members {
			total += int64(len(data))
		}
		return total
	}
	size := 4000000
	total := serialize(size)
	size += int((100 - total%100) % 100)
	total = serialize(size)
	if total%100 != 0 {
		t.Fatal("compression ratio recipe")
	}
	encode := func(comment int) []byte {
		var buffer bytes.Buffer
		writer := zip.NewWriter(&buffer)
		if err := writer.SetComment(strings.Repeat("a", comment)); err != nil {
			t.Fatal(err)
		}
		for _, path := range sortedTrustKeys(members) {
			member, err := writer.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := member.Write([]byte(members[path])); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return buffer.Bytes()
	}
	// ZIP's admitted comment changes only complete container length. At equality
	// extracted bytes are exactly 100 times that length, across all regular files.
	denominator := total / 100
	for _, over := range []bool{false, true} {
		if over {
			if got := serialize(size + 1); got != total+1 {
				t.Fatal("one-over numerator", got)
			}
		}
		comment := int(denominator) - len(encode(0))
		if comment < 0 || comment > 65535 {
			t.Fatal("unreachable ratio recipe", comment)
		}
		archive := encode(comment)
		if int64(len(archive)) != denominator {
			t.Fatal("denominator mismatch")
		}
		inventory, err := Extract(context.Background(), bytes.NewReader(archive), int64(len(archive)), DefaultArchiveLimits(), &countingDestination{})
		if over {
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != "archive_compression_ratio_exceeded" {
				t.Fatal("one-byte-over ratio", err)
			}
			continue
		}
		if err != nil {
			t.Fatal("exact compression ratio", err)
		}
		if err := ValidateInventory(manifest, inventory, false); err != nil {
			t.Fatal(err)
		}
		if err := ValidateContent(context.Background(), manifest, members, nil); err != nil {
			t.Fatal(err)
		}
		if err := ValidateNotice([]byte(members[notice])); err != nil {
			t.Fatal(err)
		}
	}
}
