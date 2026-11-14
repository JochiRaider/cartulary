package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
)

// Disk-backed test staging exercises the archive boundary without hiding an
// archive-sized buffer in a fake. Staged-object lifecycle has its own tests.
type archiveTestStaging struct {
	root string
	next int
	refs map[string]string
}

func newArchiveTestMembers(t testing.TB, limits Limits) *memberStore {
	t.Helper()
	staging := &archiveTestStaging{root: t.TempDir(), refs: map[string]string{}}
	members, err := newMemberStore("archive-test", staging, limits)
	if err != nil {
		t.Fatal(err)
	}
	return members
}
func (s *archiveTestStaging) AllocateStream(ctx context.Context, operation string, source io.Reader, maximum int64) (stagedobjects.SealedReference, error) {
	if operation != "archive-test" {
		return stagedobjects.SealedReference{}, stagedobjects.ErrScopeDenied
	}
	s.next++
	id := fmt.Sprint(s.next)
	name := filepath.Join(s.root, id)
	file, err := os.Create(name)
	if err != nil {
		return stagedobjects.SealedReference{}, err
	}
	digest := sha256.New()
	n, err := io.CopyBuffer(io.MultiWriter(file, digest), io.LimitReader(source, maximum+1), make([]byte, 32*1024))
	closeErr := file.Close()
	if err != nil {
		return stagedobjects.SealedReference{}, err
	}
	if closeErr != nil {
		return stagedobjects.SealedReference{}, closeErr
	}
	if n > maximum {
		return stagedobjects.SealedReference{}, stagedobjects.NewFailure(stagedobjects.FailureIntegrity, "member_bytes_exceeded", nil)
	}
	s.refs[id] = name
	return stagedobjects.SealedReference{Reference: stagedobjects.Reference{StagingID: id}, ContentBinding: stagedobjects.ContentBinding{ByteSize: n, SHA256: hex.EncodeToString(digest.Sum(nil))}}, ctx.Err()
}
func (s *archiveTestStaging) OpenSealed(_ context.Context, operation string, ref stagedobjects.SealedReference) (io.ReadCloser, error) {
	if operation != "archive-test" || s.refs[ref.StagingID] == "" {
		return nil, stagedobjects.ErrScopeDenied
	}
	return os.Open(s.refs[ref.StagingID])
}

type archiveTestMember struct {
	name, value string
	link        bool
	compressed  bool
}

func testStreamContainer(t testing.TB, kind string, entries []archiveTestMember) []byte {
	t.Helper()
	var out bytes.Buffer
	if kind == "zip" {
		writer := zip.NewWriter(&out)
		for _, entry := range entries {
			header := &zip.FileHeader{Name: entry.name, Method: zip.Store}
			if entry.compressed {
				header.Method = zip.Deflate
			}
			if entry.link {
				header.SetMode(os.ModeSymlink | 0600)
			}
			destination, err := writer.CreateHeader(header)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(destination, entry.value); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		var destination io.Writer = &out
		var gz *gzip.Writer
		if kind == "tar.gz" {
			gz = gzip.NewWriter(&out)
			destination = gz
		}
		writer := tar.NewWriter(destination)
		for _, entry := range entries {
			header := &tar.Header{Name: entry.name, Size: int64(len(entry.value)), Mode: 0600, Typeflag: tar.TypeReg}
			if entry.link {
				header.Typeflag = tar.TypeSymlink
				header.Size = 0
				header.Linkname = "target"
			}
			if err := writer.WriteHeader(header); err != nil {
				t.Fatal(err)
			}
			if !entry.link {
				if _, err := io.WriteString(writer, entry.value); err != nil {
					t.Fatal(err)
				}
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if gz != nil {
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	return out.Bytes()
}

func TestArchiveStream_Unit_AllContainersAndRejections(t *testing.T) {
	t.Run("preparation shares ceilings and confines profile paths", func(t *testing.T) {
		budget := &memberBudget{bytes: 5, members: 2}
		a, b := newArchiveTestMembers(t, Limits{}), newArchiveTestMembers(t, Limits{})
		a.sharedBudget, b.sharedBudget = budget, budget
		a.pathPrefix, b.pathPrefix = "prepared/a/", "prepared/b/"
		if _, err := a.WriteMember(t.Context(), "prepared/b/foreign", bytes.NewReader(nil)); !errors.Is(err, stagedobjects.ErrScopeDenied) {
			t.Fatal("foreign preparation path admitted")
		}
		if _, err := a.WriteMember(t.Context(), "prepared/a/first", bytes.NewBufferString("12345")); err != nil {
			t.Fatal(err)
		}
		if _, err := b.WriteMember(t.Context(), "prepared/b/overflow", bytes.NewBufferString("6")); err == nil {
			t.Fatal("aggregate preparation byte limit bypassed")
		}
		if _, err := b.WriteMember(t.Context(), "prepared/b/empty", bytes.NewReader(nil)); err != nil {
			t.Fatal(err)
		}
		if _, err := a.WriteMember(t.Context(), "prepared/a/third", bytes.NewReader(nil)); err == nil {
			t.Fatal("aggregate preparation member limit bypassed")
		}
		if budget.bytes != 0 || budget.members != 0 {
			t.Fatal("preparation accounting changed")
		}
		a.Seal()
		if _, err := a.WriteMember(t.Context(), "prepared/a/after", bytes.NewReader(nil)); !errors.Is(err, stagedobjects.ErrScopeDenied) {
			t.Fatal("sealed preparation accepted writes")
		}
	})
	base, err := buildFixtureBundleArchive(t, manifestInput{BundleID: "33333333-3333-4333-8333-333333333333", IncidentID: "11111111-1111-4111-8111-111111111111", IncidentKey: "CONTAINERS", ExportedAt: "2026-10-03T00:00:00Z", ReferencePackMode: referencePackModeRefsOnly}, minimalRequiredBundleFiles())
	if err != nil {
		t.Fatal(err)
	}
	logical := zipFilesMap(t, archiveBytes(t, base))
	for _, kind := range []string{"zip", "tar", "tar.gz"} {
		t.Run(kind, func(t *testing.T) {
			for _, variant := range []struct{ name, reason string }{{"valid_bundle", ""}, {"missing_member", "missing_required_file"}, {"tampered_member", "checksum_mismatch"}, {"undeclared_member", "malformed_manifest"}} {
				t.Run(variant.name, func(t *testing.T) {
					paths := make([]string, 0, len(logical))
					for name := range logical {
						paths = append(paths, name)
					}
					sort.Strings(paths)
					entries := make([]archiveTestMember, 0, len(paths)+1)
					for _, name := range paths {
						if variant.name == "missing_member" && name == "data/records.ndjson" {
							continue
						}
						value := string(logical[name])
						if variant.name == "tampered_member" && name == "data/records.ndjson" {
							value += "{}\n"
						}
						entries = append(entries, archiveTestMember{name: name, value: value})
					}
					if variant.name == "undeclared_member" {
						entries = append(entries, archiveTestMember{name: "data/undeclared.ndjson", value: "{}\n"})
					}
					_, err := verifyBundleBytes(t, verificationBytesInput{Bundle: testStreamContainer(t, kind, entries)})
					if variant.reason == "" {
						if err != nil {
							t.Fatal(err)
						}
					} else if !isVerificationReason(err, variant.reason) {
						t.Fatalf("logical member rejection: %v, want %s", err, variant.reason)
					}
				})
			}
			t.Run("compression_limit", func(t *testing.T) {
				raw := testStreamContainer(t, kind, []archiveTestMember{{name: "data/large", value: strings.Repeat("z", 128*1024), compressed: true}})
				members := newArchiveTestMembers(t, Limits{Archives: ArchiveLimits{MaxCompressionRatio: 2}})
				err := stageBundleArchive(t.Context(), bytes.NewReader(raw), int64(len(raw)), members)
				if kind == "tar" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if !isVerificationReason(err, "archive_compression_ratio_exceeded") {
					t.Fatalf("compression admission: %v", err)
				}
			})
			for _, tc := range []struct {
				name    string
				entries []archiveTestMember
				limits  Limits
				reason  string
			}{
				{"valid", []archiveTestMember{{name: "data/value.ndjson", value: "{\"value\":1}\n"}}, Limits{}, ""},
				{"traversal", []archiveTestMember{{name: "../escape", value: "x"}}, Limits{}, "invalid_member_path"},
				{"link", []archiveTestMember{{name: "data/link", link: true}}, Limits{}, "unsupported_member_type"},
				{"duplicate", []archiveTestMember{{name: "data/value", value: "a"}, {name: "data/value", value: "b"}}, Limits{}, "malformed_manifest"},
				{"extracted bytes", []archiveTestMember{{name: "data/value", value: "123456"}}, Limits{IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 5}}, "archive_extracted_bytes_exceeded"},
				{"member count", []archiveTestMember{{name: "a", value: "a"}, {name: "b", value: "b"}}, Limits{Archives: ArchiveLimits{MaxMembers: 1}}, "archive_member_count_exceeded"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					raw := testStreamContainer(t, kind, tc.entries)
					members := newArchiveTestMembers(t, tc.limits)
					err := stageBundleArchive(t.Context(), bytes.NewReader(raw), int64(len(raw)), members)
					if tc.reason != "" {
						var failure *verificationError
						if !errors.As(err, &failure) || failure.ReasonCode != tc.reason {
							t.Fatalf("rejection: %v, want %s", err, tc.reason)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					members.Seal()
					got, err := readControlMember(t.Context(), members, "data/value.ndjson", 1024)
					if err != nil || string(got) != tc.entries[0].value {
						t.Fatalf("retained bytes %q: %v", got, err)
					}
					if _, err := members.WriteMember(t.Context(), "later", bytes.NewReader(nil)); err == nil {
						t.Fatal("sealed catalog was mutable")
					}
				})
			}
		})
	}
}

type boundedZeroSource struct {
	remaining   int64
	maximumRead int
}

func (s *boundedZeroSource) Read(p []byte) (int, error) {
	if len(p) > s.maximumRead {
		s.maximumRead = len(p)
	}
	if s.remaining == 0 {
		return 0, io.EOF
	}
	n := min(int64(len(p)), s.remaining)
	clear(p[:n])
	s.remaining -= n
	return int(n), nil
}

func TestArchiveStream_Unit_BulkExceedsControlCeiling(t *testing.T) {
	const size = 65 * 1024 * 1024
	original := newArchiveTestMembers(t, Limits{})
	source := &boundedZeroSource{remaining: size}
	member, err := original.WriteMember(t.Context(), "blobs/bulk", source)
	if err != nil {
		t.Fatal(err)
	}
	original.Seal()
	archive, err := os.CreateTemp(t.TempDir(), "bundle-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	if err := writeZIPMembers(t.Context(), archive, original); err != nil {
		t.Fatal(err)
	}
	stat, err := archive.Stat()
	if err != nil {
		t.Fatal(err)
	}
	imported := newArchiveTestMembers(t, Limits{})
	if err := stageBundleArchive(t.Context(), archive, stat.Size(), imported); err != nil {
		t.Fatal(err)
	}
	// Independent SHA-256 of exactly 65 MiB of zero bytes.
	wantMember := incidentportability.Member{Path: "blobs/bulk", ByteSize: size, SHA256: "25631f11bd18756ec0029380ec886af0c8824dc6b2706bbdb1d9451c7cf45f42"}
	if member != wantMember || !slices.Equal(imported.Members(), []incidentportability.Member{wantMember}) {
		t.Fatalf("original=%+v imported=%+v want=%+v", member, imported.Members(), wantMember)
	}
	reader, opened, err := imported.OpenMember(t.Context(), "blobs/bulk")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	actualDigest := sha256.New()
	if n, err := io.CopyBuffer(actualDigest, reader, make([]byte, 32<<10)); err != nil || n != size || opened != wantMember || hex.EncodeToString(actualDigest.Sum(nil)) != wantMember.SHA256 {
		t.Fatalf("imported bytes=%d member=%+v digest=%x error=%v", n, opened, actualDigest.Sum(nil), err)
	}
	if source.maximumRead > 32*1024 {
		t.Fatal("bulk stream read an unbounded buffer")
	}
	if _, err := readControlMember(t.Context(), imported, "blobs/bulk", PortabilityParticipantByteLimit); err == nil {
		t.Fatal("bulk member admitted as control metadata")
	}
}

func TestArchiveStream_Unit_AdmittedUploadDigest(t *testing.T) {
	data := []byte("immutable admitted archive bytes")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	for _, tc := range []struct {
		name   string
		data   []byte
		size   int64
		digest string
		valid  bool
	}{
		{"exact", data, int64(len(data)), digest, true},
		{"replacement", []byte("otherwise valid replacement"), 27, digest, false},
		{"truncated", data[:len(data)-1], int64(len(data)), digest, false},
		{"missing digest", data, int64(len(data)), "", false},
		{"extended size", data, int64(len(data) + 1), digest, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyStagedArchiveDigest(context.Background(), bytes.NewReader(tc.data), tc.size, tc.digest)
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *verificationError
			if !errors.As(err, &failure) || failure.ReasonCode != "checksum_mismatch" {
				t.Fatalf("invalid upload returned %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := verifyStagedArchiveDigest(ctx, bytes.NewReader(data), int64(len(data)), digest); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
