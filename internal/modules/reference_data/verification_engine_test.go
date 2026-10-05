package reference_data

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

type engineContainer struct{ *bytes.Reader }

func (engineContainer) Close() error { return nil }

// This isolated verifier fixture starts without persisted logical versions.
// Production supplies the coordinator or an exact historical binding.
type emptyIdentityHistory struct{}

func (emptyIdentityHistory) resolveDependency(context.Context, packformat.Dependency) (packformat.Manifest, bool, error) {
	return packformat.Manifest{}, false, nil
}

func (emptyIdentityHistory) checkRetainedPresence(context.Context) error  { return nil }
func (emptyIdentityHistory) checkRetainedLength(context.Context) error    { return nil }
func (emptyIdentityHistory) checkRetainedIntegrity(context.Context) error { return nil }

func (emptyIdentityHistory) checkReleaseSequence(context.Context, packformat.Manifest, string, string) error {
	return nil
}
func (emptyIdentityHistory) checkLogicalVersion(context.Context, packformat.Manifest, string, string) error {
	return nil
}

type engineStorage struct {
	container    []byte
	closed       bool
	workspaces   int
	closeFaultAt int
	closeFault   error
}

func (s *engineStorage) OpenStaged(context.Context, StagingRef) (ContainerReader, int64, error) {
	if s.container == nil {
		return nil, 0, fs.ErrNotExist
	}
	return engineContainer{bytes.NewReader(s.container)}, int64(len(s.container)), nil
}
func (s *engineStorage) OpenPublished(ctx context.Context, _ StorageRef) (ContainerReader, int64, error) {
	return s.OpenStaged(ctx, StagingRef{})
}
func (s *engineStorage) NewWorkspace(context.Context) (VerificationWorkspace, error) {
	s.closed = false
	s.workspaces++
	return &engineWorkspace{storage: s, members: map[string][]byte{}, ordinal: s.workspaces}, nil
}

type engineWorkspace struct {
	storage *engineStorage
	members map[string][]byte
	ordinal int
}

func (w *engineWorkspace) Write(_ context.Context, name string, write func(io.Writer) error) error {
	var b bytes.Buffer
	if err := write(&b); err != nil {
		return err
	}
	w.members[name] = b.Bytes()
	return nil
}
func (w *engineWorkspace) Open(_ context.Context, name string) (io.ReadCloser, error) {
	b, ok := w.members[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (w *engineWorkspace) Close() error {
	w.storage.closed = true
	if w.ordinal == w.storage.closeFaultAt {
		return w.storage.closeFault
	}
	return nil
}

func (w *engineWorkspace) Remove(_ context.Context, name string) error {
	if _, ok := w.members[name]; !ok {
		return fs.ErrNotExist
	}
	delete(w.members, name)
	return nil
}

func TestCanonicalVerificationEngineIndependentSignedContainer_Unit(t *testing.T) {
	testPortableInputAdmission(t)
	t.Run("complete live program", testCompleteLiveProgram)
	var v struct {
		Bootstrap    string    `json:"bootstrap"`
		Repository   string    `json:"repository_id"`
		Container    string    `json:"container_base64"`
		ContainerSHA string    `json:"container_sha256"`
		ManifestSHA  string    `json:"manifest_sha256"`
		PayloadSHA   string    `json:"payload_sha256"`
		At           time.Time `json:"verification_time"`
		Expiry       time.Time `json:"expected_valid_until"`
		Signers      []string  `json:"expected_signers"`
	}
	data, err := os.ReadFile("../../../contracts/reference-packs/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	container, err := base64.StdEncoding.DecodeString(v.Container)
	if err != nil {
		t.Fatal(err)
	}
	trust, err := packformat.AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	storage := &engineStorage{container: container}
	ref, _ := ParseStagingRef("staged/fixture.zip")
	attempt := verificationAttempt{Staged: &ref, ContainerSHA256: v.ContainerSHA, Start: v.At, ClockTrusted: true, Identity: emptyIdentityHistory{}, Limits: packformat.DefaultArchiveLimits(), Repositories: map[string]packformat.TrustSnapshot{v.Repository: trust}}
	var rows builtinRows
	factory := func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
		return &rows, nil
	}
	content, err := verifyCanonicalContainer(context.Background(), storage, attempt, factory)
	if err != nil {
		t.Fatal(err)
	}
	if content.ManifestSHA256 != v.ManifestSHA || content.PayloadSHA256 != v.PayloadSHA || content.ContainerSHA256 != v.ContainerSHA || !content.Trust.ValidUntil.Equal(v.Expiry) || !reflect.DeepEqual(content.Trust.Signers["targets"], v.Signers) || len(rows) != 1 || storage.closed {
		t.Fatal("signed container disagrees with independent expected result")
	}
	if err := content.Close(); err != nil || !storage.closed {
		t.Fatal("successful workspace ownership lost")
	}
	for _, kind := range []string{"zip", "tar", "gzip"} {
		t.Run("independent equivalent "+kind, func(t *testing.T) {
			var archive struct {
				Bytes string `json:"bytes_base64"`
			}
			raw, err := os.ReadFile("../../../contracts/reference-packs/fixtures/admission/equivalent_" + kind + ".v1.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &archive); err != nil {
				t.Fatal(err)
			}
			data, err := base64.StdEncoding.Strict().DecodeString(archive.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			var expected struct {
				ContainerSHA string `json:"expected_container_sha256"`
			}
			raw, err = os.ReadFile("../../../contracts/reference-packs/fixture-manifests/admission/equivalent_" + kind + ".v1.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &expected); err != nil {
				t.Fatal(err)
			}
			candidate := attempt
			candidate.ContainerSHA256 = expected.ContainerSHA
			s := &engineStorage{container: data}
			got, err := verifyEngineFixture(s, candidate)
			if err != nil {
				t.Fatal(err)
			}
			if got.ManifestSHA256 != v.ManifestSHA || got.PayloadSHA256 != v.PayloadSHA || !got.Trust.ValidUntil.Equal(v.Expiry) || !reflect.DeepEqual(got.Trust.Signers["targets"], v.Signers) {
				t.Fatal("container format changed logical content or trust")
			}
			if err := got.Close(); err != nil || !s.closed {
				t.Fatal("workspace cleanup", err)
			}
		})
	}
	attempt.Start = v.Expiry
	if _, err := verifyCanonicalContainer(context.Background(), storage, attempt, factory); err == nil {
		t.Fatal("queue-expired envelope admitted")
	} else {
		var rejected *ContentRejection
		if !errors.As(err, &rejected) || rejected.Code != "metadata_expired" || !storage.closed {
			t.Fatalf("expiry verdict or cleanup: %v", err)
		}
	}
	attempt.ClockTrusted = false
	if _, err := verifyCanonicalContainer(context.Background(), storage, attempt, factory); err == nil {
		t.Fatal("untrusted clock admitted")
	} else {
		var rejected *OperationRejection
		if !errors.As(err, &rejected) || rejected.Reason != "clock_untrusted" {
			t.Fatalf("clock abort: %v", err)
		}
	}
	attempt.ClockTrusted = true
	attempt.Start = v.At
	storage.container = nil
	if _, err := verifyCanonicalContainer(context.Background(), storage, attempt, factory); err == nil {
		t.Fatal("missing bytes admitted")
	}
}

func canonicalEngineFixture(t *testing.T) (*engineStorage, verificationAttempt) {
	t.Helper()
	var v struct {
		Bootstrap  string    `json:"bootstrap"`
		Repository string    `json:"repository_id"`
		Container  string    `json:"container_base64"`
		At         time.Time `json:"verification_time"`
	}
	data, err := os.ReadFile("../../../contracts/reference-packs/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	container, err := base64.StdEncoding.DecodeString(v.Container)
	if err != nil {
		t.Fatal(err)
	}
	trust, err := packformat.AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	ref, _ := ParseStagingRef("staged/fixture.zip")
	return &engineStorage{container: container}, verificationAttempt{Staged: &ref, Start: v.At, ClockTrusted: true, Identity: emptyIdentityHistory{}, Limits: packformat.DefaultArchiveLimits(), Repositories: map[string]packformat.TrustSnapshot{v.Repository: trust}}
}
func verifyEngineFixture(storage VerificationStorage, attempt verificationAttempt) (*verifiedContent, error) {
	return verifyCanonicalContainer(context.Background(), storage, attempt, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
		return new(builtinRows), nil
	})
}

type retainedEngineHistory struct {
	emptyIdentityHistory
	absent      bool
	wrongLength bool
	hashes      int
}

func (h *retainedEngineHistory) checkRetainedPresence(context.Context) error {
	if h.absent {
		return &ContentRejection{Code: "payload_missing", CheckID: "retained_payload"}
	}
	return nil
}
func (h *retainedEngineHistory) checkRetainedLength(context.Context) error {
	if h.wrongLength {
		return &ContentRejection{Code: "target_length_mismatch", CheckID: "member_length"}
	}
	return nil
}
func (h *retainedEngineHistory) checkRetainedIntegrity(context.Context) error {
	h.hashes++
	return &ContentRejection{Code: "checksum_mismatch", CheckID: "member_hash"}
}

func TestRetainedContentChecksFollowVerificationPrecedence_Unit(t *testing.T) {
	t.Run("cleanup failure supersedes content rejection", func(t *testing.T) {
		var manyPaths bytes.Buffer
		writer := zip.NewWriter(&manyPaths)
		for i := range 5000 {
			header := &zip.FileHeader{Name: fmt.Sprintf("../invalid-%04d/", i), Method: zip.Store}
			header.SetMode(fs.ModeDir | 0700)
			if _, err := writer.CreateHeader(header); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			name      string
			data      []byte
			workspace int
		}{
			{"extraction workspace", []byte("invalid container"), 1},
			{"diagnostic workspace", manyPaths.Bytes(), 2},
		} {
			t.Run(test.name, func(t *testing.T) {
				storage, attempt := canonicalEngineFixture(t)
				fault := errors.New("cleanup failed")
				storage.container, storage.closeFaultAt, storage.closeFault = test.data, test.workspace, fault
				indexed := false
				result, err := verifyCanonicalContainer(context.Background(), storage, attempt, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
					indexed = true
					return new(builtinRows), nil
				})
				var rejected *ContentRejection
				if result != nil || indexed || !errors.Is(err, fault) || errors.As(err, &rejected) || storage.workspaces != test.workspace {
					t.Fatalf("cleanup misclassified: result=%v err=%v workspaces=%d", result, err, storage.workspaces)
				}
			})
		}
	})
	for _, tc := range []struct {
		name, check                                                              string
		absent, malformed, expired, wrongLength, containerHash, containerMissing bool
		issues                                                                   int
		hashes                                                                   int
	}{
		{name: "missing retained member precedes container format", absent: true, malformed: true, check: "retained_payload"},
		{name: "container format precedes retained hash", malformed: true, check: "container_format"},
		{name: "trust expiry precedes retained hash", expired: true, check: "metadata_expiry"},
		{name: "retained length precedes retained hash", wrongLength: true, check: "member_length"},
		{name: "retained hash runs after earlier checks pass", check: "member_hash", hashes: 1},
		{name: "all missing retained members and container", absent: true, containerMissing: true, check: "retained_payload", issues: 2},
		{name: "all retained and container hash failures", containerHash: true, check: "member_hash", hashes: 1, issues: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage, attempt := canonicalEngineFixture(t)
			history := &retainedEngineHistory{absent: tc.absent, wrongLength: tc.wrongLength}
			attempt.Identity = history
			if tc.containerHash {
				attempt.ContainerSHA256 = strings.Repeat("0", 64)
			}
			if tc.containerMissing {
				storage.container = nil
			}
			if tc.malformed {
				storage.container = []byte("invalid container")
			}
			if tc.expired {
				attempt.Start = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
			}
			indexed := false
			_, err := verifyCanonicalContainer(context.Background(), storage, attempt, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
				indexed = true
				return new(builtinRows), nil
			})
			var rejected *ContentRejection
			if !errors.As(err, &rejected) || rejected.CheckID != tc.check || history.hashes != tc.hashes || indexed {
				t.Fatalf("verification order: result=%v hashes=%d indexed=%v", err, history.hashes, indexed)
			}
			if tc.issues > 0 && (rejected.Summary == nil || rejected.Summary.Total != tc.issues) {
				t.Fatal("same-check findings were dropped", rejected.Summary)
			}
		})
	}
}

type engineReadFaultStorage struct {
	*engineStorage
	at int
}

func (s engineReadFaultStorage) OpenStaged(ctx context.Context, ref StagingRef) (ContainerReader, int64, error) {
	r, size, err := s.engineStorage.OpenStaged(ctx, ref)
	if err != nil {
		return nil, 0, err
	}
	return &engineReadFault{ContainerReader: r, at: s.at}, size, nil
}

type engineReadFault struct {
	ContainerReader
	at, reads int
}

func (r *engineReadFault) ReadAt(p []byte, offset int64) (int, error) {
	r.reads++
	if r.reads == r.at {
		return 0, io.ErrClosedPipe
	}
	return r.ContainerReader.ReadAt(p, offset)
}

func TestVerifierNeverTurnsOperationalReadsIntoContentVerdicts_Unit(t *testing.T) {
	for _, at := range []int{1, 2, 4, 20} {
		storage, attempt := canonicalEngineFixture(t)
		faulted := engineReadFaultStorage{storage, at}
		content, err := verifyEngineFixture(faulted, attempt)
		var verdict *ContentRejection
		if content != nil || !errors.Is(err, io.ErrClosedPipe) || errors.As(err, &verdict) || !storage.closed {
			t.Fatalf("read %d was misclassified: %v", at, err)
		}
		identity, err := probeContainerIdentity(context.Background(), faulted, *attempt.Staged, attempt.Limits)
		if !errors.Is(err, io.ErrClosedPipe) || !reflect.DeepEqual(identity, importIdentity{}) || !storage.closed {
			t.Fatalf("probe read %d was misclassified: %v", at, err)
		}
	}
}
func TestVerifierRejectsArchiveLimits_Unit(t *testing.T) {
	// The independently signed container is complete at each configured bound.
	// Tightening one bound by one must reject that same authenticated input,
	// before publishing any usable content.
	for _, field := range []string{"container", "extracted", "members"} {
		t.Run("complete equality/"+field, func(t *testing.T) {
			storage, attempt := canonicalEngineFixture(t)
			archive, err := zip.NewReader(bytes.NewReader(storage.container), int64(len(storage.container)))
			if err != nil {
				t.Fatal(err)
			}
			var extracted int64
			for _, member := range archive.File {
				extracted += int64(member.UncompressedSize64)
			}
			attempt.Limits.ContainerBytes = int64(len(storage.container))
			attempt.Limits.ExtractedBytes = extracted
			attempt.Limits.Members = int64(len(archive.File))
			content, err := verifyEngineFixture(storage, attempt)
			if err != nil {
				t.Fatal("complete at-limit verification", err)
			}
			if err := content.Close(); err != nil {
				t.Fatal(err)
			}
			var code string
			switch field {
			case "container":
				attempt.Limits.ContainerBytes--
				code = "container_bytes_exceeded"
			case "extracted":
				attempt.Limits.ExtractedBytes--
				code = "archive_extracted_bytes_exceeded"
			case "members":
				attempt.Limits.Members--
				code = "archive_member_count_exceeded"
			}
			_, err = verifyEngineFixture(storage, attempt)
			var rejected *ContentRejection
			if !errors.As(err, &rejected) || rejected.Code != code || !storage.closed {
				t.Fatalf("complete one-over verification: %v", err)
			}
		})
	}
	for _, tc := range []struct{ name, code string }{
		{"member-count", "archive_member_count_exceeded"}, {"extracted-bytes", "archive_extracted_bytes_exceeded"}, {"compression-ratio", "archive_compression_ratio_exceeded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage, attempt := canonicalEngineFixture(t)
			switch tc.name {
			case "member-count":
				attempt.Limits.Members = 1
			case "extracted-bytes":
				attempt.Limits.ExtractedBytes = 1
			case "compression-ratio":
				var b bytes.Buffer
				w := zip.NewWriter(&b)
				addZipFile(t, w, "manifest.json", bytes.Repeat([]byte("a"), 65536))
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
				storage.container = b.Bytes()
				attempt.Limits.CompressionRatio = 1
			}
			_, err := verifyEngineFixture(storage, attempt)
			var rejected *ContentRejection
			if !errors.As(err, &rejected) || rejected.Code != tc.code || !storage.closed {
				t.Fatalf("bounded canonical admission: %v", err)
			}
		})
	}
}
func TestVerifierAcceptsLocalBundleAndRejectsFailures_Unit(t *testing.T) {
	storage, attempt := canonicalEngineFixture(t)
	verified, err := verifyEngineFixture(storage, attempt)
	if err != nil {
		t.Fatal(err)
	}
	if err = verified.Close(); err != nil {
		t.Fatal(err)
	}
	for _, signed := range []bool{false, true} {
		storage.container = retiredReferencePackBundle(t, retiredBundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "1", Signed: signed})
		_, err := verifyEngineFixture(storage, attempt)
		var rejected *ContentRejection
		if !errors.As(err, &rejected) || rejected.Code != "bundle_hint_invalid" {
			t.Fatalf("retired checksum-labelled format admitted (signed=%v): %v", signed, err)
		}
	}
	storage, attempt = canonicalEngineFixture(t)
	reader, err := zip.NewReader(bytes.NewReader(storage.container), int64(len(storage.container)))
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	writer := zip.NewWriter(&b)
	for _, f := range reader.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		if f.Name == "metadata/targets.json" {
			var metadata map[string]any
			if err = json.Unmarshal(data, &metadata); err != nil {
				t.Fatal(err)
			}
			metadata["signatures"].([]any)[0].(map[string]any)["sig"] = string(bytes.Repeat([]byte("0"), 128))
			data, err = json.Marshal(metadata)
			if err != nil {
				t.Fatal(err)
			}
		}
		addZipFile(t, writer, f.Name, data)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	storage.container = b.Bytes()
	_, err = verifyEngineFixture(storage, attempt)
	if err == nil {
		t.Fatal("modified signed role admitted")
	}
}

func TestVerifierCombinedFaultPrecedence_Unit(t *testing.T) {
	for _, test := range []struct {
		name, code, check string
		change            func(map[string][]byte, *verificationAttempt)
	}{
		{"missing manifest and invalid hint", "bundle_hint_invalid", "bundle_shape", func(files map[string][]byte, _ *verificationAttempt) {
			delete(files, "manifest.json")
			files["bundle.json"] = []byte("{}")
		}},
		{"oversized manifest and invalid hint", "bundle_hint_invalid", "bundle_shape", func(files map[string][]byte, _ *verificationAttempt) {
			files["manifest.json"] = bytes.Repeat([]byte(" "), 1048577)
			files["bundle.json"] = []byte("{}")
		}},
		{"unknown repository and oversized metadata", "tuf_root_untrusted", "repository_selection", func(files map[string][]byte, attempt *verificationAttempt) {
			files["metadata/targets.json"] = bytes.Repeat([]byte(" "), 2097153)
			attempt.Repositories = nil
		}},
		{"missing role and invalid signature", "tuf_metadata_invalid", "tuf_schema", func(files map[string][]byte, _ *verificationAttempt) {
			delete(files, "metadata/snapshot.json")
			var envelope map[string]any
			if err := json.Unmarshal(files["metadata/targets.json"], &envelope); err != nil {
				t.Fatal(err)
			}
			envelope["signatures"].([]any)[0].(map[string]any)["sig"] = string(bytes.Repeat([]byte("0"), 128))
			files["metadata/targets.json"], _ = json.Marshal(envelope)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			storage, attempt := canonicalEngineFixture(t)
			rewriteEngineContainer(t, storage, func(files map[string][]byte) { test.change(files, &attempt) })
			_, err := verifyEngineFixture(storage, attempt)
			var rejected *ContentRejection
			if !errors.As(err, &rejected) || rejected.Code != test.code || rejected.CheckID != test.check || !storage.closed {
				t.Fatalf("verdict=%v, want %s/%s and closed workspace", err, test.check, test.code)
			}
		})
	}
	for _, test := range []struct {
		path, code string
		maximum    int
	}{
		{"manifest.json", "manifest_schema_invalid", 1048576},
		{"bundle.json", "bundle_hint_invalid", 16384},
		{"metadata/targets.json", "tuf_metadata_invalid", 2097152},
	} {
		t.Run(test.path+" byte guard", func(t *testing.T) {
			for _, delta := range []int{0, 1} {
				payload := bytes.Repeat([]byte("x"), test.maximum+delta)
				workspace := &engineWorkspace{members: map[string][]byte{test.path: payload}}
				inventory := packformat.Inventory{test.path: {Size: int64(len(payload)), SHA256: packformat.Digest(payload)}}
				got, err := readAttemptMember(context.Background(), workspace, inventory, test.path, int64(test.maximum))
				if delta == 0 {
					if err != nil || !bytes.Equal(got, payload) {
						t.Fatalf("at-limit read failed: %v", err)
					}
				} else {
					var failure *packformat.Failure
					if !errors.As(err, &failure) || failure.Code != test.code || got != nil {
						t.Fatalf("limit outcome: %v", err)
					}
				}
			}
		})
	}
}

func rewriteEngineContainer(t *testing.T, storage *engineStorage, change func(map[string][]byte)) {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(storage.container), int64(len(storage.container)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		payload, err := io.ReadAll(stream)
		if closeErr := stream.Close(); err != nil || closeErr != nil {
			t.Fatal(errors.Join(err, closeErr))
		}
		files[file.Name] = payload
	}
	change(files)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, name := range names {
		// Stored bytes keep these schema-limit fixtures below the independent
		// compression-ratio guard. Signatures remain those of the original vector.
		member, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := member.Write(files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	storage.container = buffer.Bytes()
}

func testCompleteLiveProgram(t *testing.T) {
	t.Run("retained trust guard and misplaced verdict", func(t *testing.T) {
		for _, check := range []string{"tuf_schema", "metadata_expiry"} {
			storage, attempt := canonicalEngineFixture(t)
			attempt.ResolveTrust = func(context.Context, string, []int64) (packformat.TrustSnapshot, bool, error) {
				return packformat.TrustSnapshot{}, false, &ContentRejection{CheckID: check, Code: "tuf_metadata_invalid"}
			}
			content, err := verifyEngineFixture(storage, attempt)
			var rejected *ContentRejection
			if content != nil || err == nil {
				t.Fatal("invalid trust guard admitted")
			}
			if check == "tuf_schema" {
				if !errors.As(err, &rejected) || rejected.Summary == nil || rejected.Summary.Total != 1 || rejected.CheckID != check {
					t.Fatal("trust guard escaped registry", err)
				}
			} else if errors.As(err, &rejected) {
				t.Fatal("misplaced verdict remained publishable", err)
			}
		}
	})

	storage, attempt := canonicalEngineFixture(t)
	p := &containerVerificationProgram{attempt: attempt, storage: storage}
	program, err := p.checks()
	if err != nil {
		t.Fatal(err)
	}
	executed := []string{}
	for id, check := range program {
		program[id] = func(ctx context.Context, emit packformat.FindingSink) error {
			executed = append(executed, id)
			return check(ctx, emit)
		}
	}
	summary, err := packformat.RunChecks(context.Background(), nil, program)
	if p.source != nil {
		defer p.source.Close()
	}
	if p.workspace != nil {
		defer p.workspace.Close()
	}
	if err != nil || summary.Result != "succeeded" {
		t.Fatalf("live program: %v %#v", err, summary)
	}
	want := []string{}
	for _, check := range packformat.Checks() {
		want = append(want, check.ID)
	}
	if len(want) != 49 || !slices.Equal(executed, want) {
		t.Fatalf("live check order: %v", executed)
	}
	// A missing implementation fails before any private bytes are opened.
	for _, id := range want {
		fresh := &containerVerificationProgram{attempt: attempt, storage: storage}
		incomplete, err := fresh.checks()
		if err != nil {
			t.Fatal(err)
		}
		delete(incomplete, id)
		if result, err := packformat.RunChecks(context.Background(), nil, incomplete); result != nil || err == nil || fresh.source != nil || fresh.workspace != nil {
			t.Fatalf("omitted %s did not fail before execution", id)
		}
	}
}
