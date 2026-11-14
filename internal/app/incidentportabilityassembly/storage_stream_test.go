package incidentportabilityassembly_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
)

func TestIncidentBundleRootStorageStreamsBoundedArchives_Unit(t *testing.T) {
	temporary, exports := t.TempDir(), t.TempDir()
	storage, err := incidentportabilityassembly.NewRootStorage(temporary, exports)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	const size = int64(65 << 20)
	// Independent SHA-256 of exactly 65 MiB of zero bytes.
	wantContent := incidentbundles.ArchiveContent{ByteSize: size, SHA256: "25631f11bd18756ec0029380ec886af0c8824dc6b2706bbdb1d9451c7cf45f42"}
	source := &boundedZeroSource{remaining: size}
	staged, err := storage.Stage(context.Background(), source, size)
	if err != nil || staged.ByteSize != size || source.maximumRead > 32<<10 {
		t.Fatalf("streamed upload: size=%d max read=%d error=%v", staged.ByteSize, source.maximumRead, err)
	}
	if staged.ArchiveContent != wantContent {
		t.Fatalf("staged content=%+v want=%+v", staged.ArchiveContent, wantContent)
	}
	reader, actual, err := storage.OpenStaged(staged.BundleStagingRef, size)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	digest := sha256.New()
	if n, err := io.CopyBuffer(digest, io.NewSectionReader(reader, 0, actual), make([]byte, 32<<10)); err != nil || n != size {
		t.Fatal(n, err)
	}
	if hex.EncodeToString(digest.Sum(nil)) != staged.SHA256 {
		t.Fatal("upload seal differs from retained bytes")
	}
	published, err := storage.Publish(context.Background(), "22222222-2222-4222-8222-222222222222", size, func(destination io.Writer) error {
		_, err := io.CopyBuffer(destination, io.NewSectionReader(reader, 0, actual), make([]byte, 32<<10))
		return err
	})
	_ = reader.Close()
	if err != nil {
		t.Fatal("publication failed", err)
	}
	if published.ArchiveContent != wantContent {
		t.Fatalf("published content=%+v want=%+v", published.ArchiveContent, wantContent)
	}
	public, err := storage.OpenPublished(context.Background(), published.BundleStorageRef, published.ArchiveContent)
	if err != nil {
		t.Fatal(err)
	}
	defer public.Close()
	publishedDigest := sha256.New()
	if n, err := io.CopyBuffer(publishedDigest, public, make([]byte, 32<<10)); err != nil || n != size || hex.EncodeToString(publishedDigest.Sum(nil)) != wantContent.SHA256 {
		t.Fatalf("published bytes=%d digest=%x error=%v", n, publishedDigest.Sum(nil), err)
	}
	_ = public.Close()
	wrong := published.ArchiveContent
	wrong.SHA256 = strings.Repeat("0", 64)
	if reader, err := storage.OpenPublished(context.Background(), published.BundleStorageRef, wrong); err == nil {
		reader.Close()
		t.Fatal("wrong published hash admitted")
	}
	wrong = published.ArchiveContent
	wrong.ByteSize--
	if reader, err := storage.OpenPublished(context.Background(), published.BundleStorageRef, wrong); err == nil {
		reader.Close()
		t.Fatal("wrong published size admitted")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if reader, err := storage.OpenPublished(canceled, published.BundleStorageRef, published.ArchiveContent); !errors.Is(err, context.Canceled) {
		if reader != nil {
			reader.Close()
		}
		t.Fatal("canceled public read", err)
	}
	oversized, _, oversizeErr := storage.OpenStaged(staged.BundleStagingRef, size-1)
	if oversized != nil {
		_ = oversized.Close()
	}
	if !errors.Is(oversizeErr, incidentbundles.ErrArchiveBytesExceeded) {
		t.Fatal("oversize open admitted", oversizeErr)
	}
	if err := storage.RemoveStaged(staged.BundleStagingRef); err != nil {
		t.Fatal(err)
	}
	if err := storage.RemovePublished(published.BundleStorageRef); err != nil {
		t.Fatal(err)
	}

	t.Run("upload limit and interrupted source leave no files", func(t *testing.T) {
		for _, tc := range []struct {
			source io.Reader
			want   error
		}{
			{strings.NewReader("too long"), incidentbundles.ErrArchiveBytesExceeded},
			{io.MultiReader(strings.NewReader("ok"), failedArchiveReader{}), io.ErrUnexpectedEOF},
		} {
			if _, err := storage.Stage(context.Background(), tc.source, 3); !errors.Is(err, tc.want) {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(filepath.Join(temporary, "incident-bundles", "imports"))
			if err != nil || len(entries) != 0 {
				t.Fatal("failed upload leaked bytes", entries, err)
			}
		}
	})
	t.Run("ignored writer failure cannot seal", func(t *testing.T) {
		_, err := storage.Publish(context.Background(), "33333333-3333-4333-8333-333333333333", 3, func(destination io.Writer) error {
			_, _ = io.WriteString(destination, "too long")
			return nil
		})
		if !errors.Is(err, incidentbundles.ErrArchiveBytesExceeded) {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join(exports, "incident-bundles"))
		if err != nil || len(entries) != 0 {
			t.Fatal("failed publication leaked bytes", entries, err)
		}
	})
}

type boundedZeroSource struct {
	remaining   int64
	maximumRead int
}

func (s *boundedZeroSource) Read(p []byte) (int, error) {
	s.maximumRead = max(s.maximumRead, len(p))
	if s.remaining == 0 {
		return 0, io.EOF
	}
	n := min(int64(len(p)), s.remaining)
	clear(p[:n])
	s.remaining -= n
	return int(n), nil
}

type failedArchiveReader struct{}

func (failedArchiveReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
