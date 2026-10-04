package referenceassembly_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
)

type countedStagingReader struct {
	reader io.Reader
	count  int
}

func (r *countedStagingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.count += n
	return n, err
}

type stagingReadFailure struct{}

func (stagingReadFailure) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

type stagingProbeFailure struct{}

func (stagingProbeFailure) Read(p []byte) (int, error) {
	p[0] = 'x'
	return 1, io.ErrUnexpectedEOF
}

func testStagingBounds(t *testing.T) {
	root := t.TempDir()
	storage, err := referenceassembly.NewRootStorage(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, length := range []int{127, 128, 129, 4096} {
		source := &countedStagingReader{reader: strings.NewReader(strings.Repeat("x", length))}
		ref, _, size, err := storage.StageStream(context.Background(), source, 128)
		if length <= 128 {
			if err != nil || size != int64(length) {
				t.Fatal(size, err)
			}
			if err := storage.RemoveStaged(ref); err != nil {
				t.Fatal(err)
			}
		} else {
			var bound *reference_data.StagingLimitError
			if !errors.As(err, &bound) || bound.Maximum != 128 || bound.Observed != 129 || source.count != 129 || ref.String() != "" {
				t.Fatalf("bound: %v consumed=%d", err, source.count)
			}
		}
	}
	_, _, _, err = storage.StageStream(context.Background(), io.MultiReader(strings.NewReader("short"), stagingReadFailure{}), 128)
	var bound *reference_data.StagingLimitError
	if !errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &bound) {
		t.Fatal("read failure classified as content", err)
	}
	_, _, _, err = storage.StageStream(context.Background(), io.MultiReader(strings.NewReader(strings.Repeat("x", 128)), stagingProbeFailure{}), 128)
	if !errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &bound) {
		t.Fatal("probe read failure classified as content", err)
	}
	files, err := os.ReadDir(filepath.Join(root, "reference-packs/imports"))
	if err != nil || len(files) != 0 {
		t.Fatal("temporary files retained", files, err)
	}
}
