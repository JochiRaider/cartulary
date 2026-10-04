package packformat

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"hash/crc32"
	"io"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type memoryDestination map[string]*bytes.Buffer

type countedArchiveReader struct {
	*bytes.Reader
	read int
}

func (r *countedArchiveReader) ReadAt(p []byte, at int64) (int, error) {
	n, err := r.Reader.ReadAt(p, at)
	r.read += n
	return n, err
}

func TestArchiveZIPMetadataHasBoundedWorkingReads_Unit(t *testing.T) {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for i := 0; i < 32; i++ {
		_, err := w.CreateHeader(&zip.FileHeader{Name: fmt.Sprintf("payload/%02d", i), Method: zip.Store, Comment: strings.Repeat("x", 60000)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reader := &countedArchiveReader{Reader: bytes.NewReader(b.Bytes())}
	inventory, err := Extract(context.Background(), reader, int64(b.Len()), DefaultArchiveLimits(), memoryDestination{})
	if err != nil || len(inventory) != 32 {
		t.Fatalf("metadata fixture: %v", err)
	}
	// The bounded EOCD tail scan may read one comment. Central parsing must
	// skip the remaining 1.8 MB instead of retaining comments in zip.Reader.
	// Preflight and extraction each perform one bounded scan.
	if reader.read > 262144 {
		t.Fatalf("retained or scanned unneeded ZIP comments: %d", reader.read)
	}
}

func TestArchiveZIPChecksDecodedCRCSizeAndCompressedEnd_Unit(t *testing.T) {
	payload := []byte("canonical payload")
	var compressed bytes.Buffer
	deflater, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := deflater.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := deflater.Close(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		crc     uint32
		decoded uint64
		suffix  []byte
		valid   bool
	}{
		{"valid", crc32.ChecksumIEEE(payload), uint64(len(payload)), nil, true},
		{"crc", 0, uint64(len(payload)), nil, false},
		{"short declaration", crc32.ChecksumIEEE(payload), uint64(len(payload) - 1), nil, false},
		{"long declaration", crc32.ChecksumIEEE(payload), uint64(len(payload) + 1), nil, false},
		{"trailing compressed byte", crc32.ChecksumIEEE(payload), uint64(len(payload)), []byte{0}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var b bytes.Buffer
			w := zip.NewWriter(&b)
			data := append(append([]byte{}, compressed.Bytes()...), test.suffix...)
			member, err := w.CreateRaw(&zip.FileHeader{Name: "payload/data", Method: zip.Deflate, CRC32: test.crc, UncompressedSize64: test.decoded, CompressedSize64: uint64(len(data))})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := member.Write(data); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			_, err = Extract(context.Background(), bytes.NewReader(b.Bytes()), int64(b.Len()), DefaultArchiveLimits(), memoryDestination{})
			if (err == nil) != test.valid {
				t.Fatalf("admission: %v", err)
			}
		})
	}
}

func TestArchiveTARPaddingHasIndependentStructuralBound_Unit(t *testing.T) {
	base := container(t, "tar", []string{"manifest.json"}, []byte("{}"))
	for _, blocks := range []int{2, 20, 21} {
		for _, gzipWrapped := range []bool{false, true} {
			data := append(append([]byte{}, base...), make([]byte, (blocks-2)*512)...)
			if gzipWrapped {
				var encoded bytes.Buffer
				w := gzip.NewWriter(&encoded)
				if _, err := w.Write(data); err != nil {
					t.Fatal(err)
				}
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				data = encoded.Bytes()
			}
			_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), memoryDestination{})
			if blocks <= 20 && err != nil {
				t.Fatalf("padding %d, gzip %v: %v", blocks, gzipWrapped, err)
			}
			if blocks > 20 {
				var failure *Failure
				if !errors.As(err, &failure) || failure.Code != "archive_structure_invalid" {
					t.Fatalf("excess padding: %v", err)
				}
			}
		}
	}
}

func (d memoryDestination) Write(_ context.Context, name string, write func(io.Writer) error) error {
	b := new(bytes.Buffer)
	d[name] = b
	return write(b)
}

func TestArchiveDirectoryMarkersHaveAnIndependentStructuralBound_Unit(t *testing.T) {
	for _, count := range []int{10000, 10001} {
		var b bytes.Buffer
		writer := tar.NewWriter(&b)
		for i := 0; i < count; i++ {
			if err := writer.WriteHeader(&tar.Header{Name: fmt.Sprintf("d%05d/", i), Typeflag: tar.TypeDir, Mode: 0700, Format: tar.FormatUSTAR}); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: 2, Format: tar.FormatUSTAR}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte("{}")); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		inventory, err := Extract(context.Background(), bytes.NewReader(b.Bytes()), int64(b.Len()), DefaultArchiveLimits(), memoryDestination{})
		if count == 10000 {
			if err != nil || len(inventory) != 1 {
				t.Fatalf("at-limit directories: %v", err)
			}
		} else {
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != "archive_structure_invalid" {
				t.Fatalf("over-limit directories: %v", err)
			}
		}
	}
}

func TestArchiveRatioUsesCompleteContainerOnceAndHandlesOverflow_Unit(t *testing.T) {
	t.Run("complete file byte boundary", testCompleteArchiveFileByteBoundary)
	data := container(t, "gzip", []string{"payload/one", "payload/two"}, bytes.Repeat([]byte("a"), 100000))
	limits := DefaultArchiveLimits()
	limits.CompressionRatio = 150000 / int64(len(data))
	if limits.CompressionRatio < 1 || limits.CompressionRatio > 1000 || int64(len(data))*limits.CompressionRatio >= 200000 || int64(len(data))*limits.CompressionRatio < 100000 {
		t.Fatal("ratio fixture no longer straddles the two-member boundary")
	}
	_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), limits, memoryDestination{})
	var failure *Failure
	if !errors.As(err, &failure) || failure.Code != "archive_compression_ratio_exceeded" {
		t.Fatalf("ratio denominator was multiplied per member: %v", err)
	}
	admission := archiveAdmission{limits: ArchiveLimits{Members: 1, ExtractedBytes: math.MaxInt64, CompressionRatio: 1000}, sourceBytes: math.MaxInt64, compressed: true, members: Inventory{}}
	if err := admission.reserve(268435456); err != nil {
		t.Fatalf("overflow-safe guard rejected within-bound input: %v", err)
	}
}

func container(t *testing.T, kind string, names []string, payload []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	if kind == "zip" {
		w := zip.NewWriter(&b)
		for _, name := range names {
			f, err := w.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.Write(payload); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return b.Bytes()
	}
	w := tar.NewWriter(&b)
	for _, name := range names {
		if err := w.WriteHeader(&tar.Header{Name: name, Size: int64(len(payload)), Mode: 0600, Format: tar.FormatUSTAR}); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if kind == "tar" {
		return b.Bytes()
	}
	var gz bytes.Buffer
	g := gzip.NewWriter(&gz)
	if _, err := g.Write(b.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
	return gz.Bytes()
}

func TestArchiveContainersHaveEquivalentLogicalBytes_Unit(t *testing.T) {
	var expected Inventory
	for _, kind := range []string{"zip", "tar", "gzip"} {
		t.Run(kind, func(t *testing.T) {
			data := container(t, kind, []string{"manifest.json", "payload/entries.ndjson"}, []byte("{}\n"))
			destination := memoryDestination{}
			limits := DefaultArchiveLimits()
			limits.ContainerBytes = int64(len(data))
			limits.ExtractedBytes = 6
			limits.Members = 2
			got, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), limits, destination)
			if err != nil {
				t.Fatal(err)
			}
			if expected == nil {
				expected = got
			} else if !reflect.DeepEqual(expected, got) {
				t.Fatal("container format changed logical identity")
			}
			if destination["payload/entries.ndjson"].String() != "{}\n" {
				t.Fatal("extracted bytes changed")
			}
		})
	}
}

func TestArchiveAdmissionRejectsAttacksAndLimits_Unit(t *testing.T) {
	for _, kind := range []string{"zip", "tar", "gzip"} {
		for _, test := range []struct {
			name  string
			names []string
			code  string
		}{
			{"traversal", []string{"a/../manifest.json"}, "path_traversal"},
			{"duplicate", []string{"manifest.json", "manifest.json"}, "path_collision"},
			{"case", []string{"manifest.json", "Manifest.json"}, "path_collision"},
			{"parent", []string{"payload", "payload/data"}, "path_collision"},
			{"parent reverse", []string{"payload/data", "payload"}, "path_collision"},
		} {
			t.Run(kind+"/"+test.name, func(t *testing.T) {
				data := container(t, kind, test.names, []byte("{}"))
				_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), memoryDestination{})
				var rejected *Failure
				if !errors.As(err, &rejected) || rejected.Code != test.code {
					t.Fatalf("error=%v, want %s", err, test.code)
				}
			})
		}
		data := container(t, kind, []string{"manifest.json"}, []byte("{}"))
		limits := DefaultArchiveLimits()
		limits.ExtractedBytes = 1
		destination := memoryDestination{}
		_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), limits, destination)
		var rejected *Failure
		if !errors.As(err, &rejected) || rejected.Code != "archive_extracted_bytes_exceeded" || len(destination) != 0 {
			t.Fatalf("limit checked after writing: %v", err)
		}
	}
	data := container(t, "gzip", []string{"manifest.json"}, []byte("{}"))
	data = append(data, container(t, "gzip", []string{"other"}, []byte("{}"))...)
	if _, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), memoryDestination{}); err == nil {
		t.Fatal("accepted concatenated gzip")
	}
}

func TestArchivePathDirectoryExceptionAndCancellation_Unit(t *testing.T) {
	if err := ValidatePath("empty/", true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"empty//", "/", "a/./", "a/../", "a\\b/"} {
		if ValidatePath(name, true) == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	data := container(t, "tar", []string{"manifest.json"}, []byte("{}"))
	if _, err := Extract(ctx, bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), memoryDestination{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
}

func TestArchiveRegistryPrecedenceAndNoRejectedWrites_Unit(t *testing.T) {
	for _, kind := range []string{"zip", "tar", "gzip"} {
		for _, tc := range []struct {
			name  string
			names []string
			check string
			count int
		}{
			{"path beats collisions", []string{"same", "Same", "a/../secret", "bad/./secret"}, "archive_path", 2},
			{"all collisions", []string{"same", "Same", "parent", "parent/child"}, "archive_collision", 4},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				var expected []byte
				for _, reverse := range []bool{false, true} {
					names := slices.Clone(tc.names)
					if reverse {
						slices.Reverse(names)
					}
					data := container(t, kind, names, []byte("payload"))
					destination := memoryDestination{}
					_, err := ExtractWithDiagnostics(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), destination, nil)
					var f *Failure
					if !errors.As(err, &f) || f.CheckID != tc.check || f.Summary == nil || f.Summary.Total != tc.count {
						t.Fatalf("%s: %#v / %v", tc.check, f, err)
					}
					if len(destination) != 0 {
						t.Fatal("rejected path reached destination")
					}
					raw, err := canonicaljson.Marshal(f.Summary)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := DecodeValidationSummary(raw); err != nil {
						t.Fatal(err)
					}
					if bytes.Contains(raw, []byte("secret")) {
						t.Fatal("hostile path echoed")
					}
					if expected == nil {
						expected = raw
					} else if !bytes.Equal(expected, raw) {
						t.Fatal("member order changed summary")
					}
				}
			})
		}
	}
	t.Run("format before size", func(t *testing.T) {
		data := bytes.Repeat([]byte("invalid"), 100)
		limits := DefaultArchiveLimits()
		limits.ContainerBytes = 4
		_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), limits, memoryDestination{})
		var f *Failure
		if !errors.As(err, &f) || f.CheckID != "container_format" {
			t.Fatal(err)
		}
	})
	t.Run("late framing before early traversal", func(t *testing.T) {
		data := container(t, "tar", []string{"../secret", "ok"}, []byte("x"))
		data[len(data)-1] = 1
		destination := memoryDestination{}
		_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), destination)
		var f *Failure
		if !errors.As(err, &f) || f.CheckID != "archive_structure" || len(destination) != 0 {
			t.Fatal(err)
		}
	})
	t.Run("zero-body type before and after traversal", func(t *testing.T) {
		for _, reverse := range []bool{false, true} {
			var encoded bytes.Buffer
			writer := tar.NewWriter(&encoded)
			headers := []*tar.Header{{Name: "link", Linkname: "target", Typeflag: tar.TypeSymlink, Mode: 0600, Format: tar.FormatUSTAR}, {Name: "../secret", Typeflag: tar.TypeReg, Mode: 0600, Format: tar.FormatUSTAR}}
			if reverse {
				slices.Reverse(headers)
			}
			for _, header := range headers {
				if err := writer.WriteHeader(header); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			destination := memoryDestination{}
			_, err := Extract(context.Background(), bytes.NewReader(encoded.Bytes()), int64(encoded.Len()), DefaultArchiveLimits(), destination)
			var f *Failure
			if !errors.As(err, &f) || f.CheckID != "archive_path" || len(destination) != 0 {
				t.Fatal(err)
			}
		}
	})
}

type changingArchiveReader struct {
	first, second *bytes.Reader
	headers       int
}

func (r *changingArchiveReader) ReadAt(data []byte, offset int64) (int, error) {
	if offset == 0 && len(data) == 512 {
		r.headers++
	}
	if r.headers >= 3 {
		return r.second.ReadAt(data, offset)
	}
	return r.first.ReadAt(data, offset)
}
func TestArchiveExtractionRejectsChangedAdmittedSource_Unit(t *testing.T) {
	first := container(t, "tar", []string{"safe"}, []byte("payload"))
	changed := container(t, "tar", []string{"../unsafe"}, []byte("payload"))
	destination := memoryDestination{}
	_, err := Extract(context.Background(), &changingArchiveReader{first: bytes.NewReader(first), second: bytes.NewReader(changed)}, int64(len(first)), DefaultArchiveLimits(), destination)
	if !errors.Is(err, ErrStorage) || len(destination) != 0 {
		t.Fatalf("changed source reached extraction: %v", err)
	}
}

func TestArchiveTAROriginalNameAndChecksumAdmission_Unit(t *testing.T) {
	for _, field := range []string{"name", "prefix", "extension checksum", "extension traversal"} {
		t.Run(field, func(t *testing.T) {
			data := container(t, "tar", []string{"safe"}, []byte("payload"))
			switch field {
			case "name":
				data[5] = 'x'
			case "prefix":
				data[346] = 'x'
			case "extension checksum":
				data[156] = tar.TypeXHeader
			case "extension traversal":
				clear(data[:100])
				copy(data[:100], []byte("../secret"))
				data[156] = tar.TypeXHeader
			}
			if field != "extension checksum" {
				for i := 148; i < 156; i++ {
					data[i] = ' '
				}
				var checksum int
				for _, b := range data[:512] {
					checksum += int(b)
				}
				copy(data[148:156], []byte(fmt.Sprintf("%06o\x00 ", checksum)))
			}
			destination := memoryDestination{}
			_, err := Extract(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), destination)
			var f *Failure
			check := "archive_structure"
			if field == "extension traversal" {
				check = "archive_path"
			}
			if !errors.As(err, &f) || f.CheckID != check || len(destination) != 0 {
				t.Fatalf("truncated name or unverified extension: %v", err)
			}
		})
	}
}
