package incidentbundles

import (
	"bytes"
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"io"
	"os"
	"sort"
	"testing"
)

type verificationBytesInput struct {
	Bundle []byte
	Limits Limits
}

func verifyBundleBytes(t testing.TB, input verificationBytesInput) (verifiedBundle, error) {
	t.Helper()
	members := newArchiveTestMembers(t, input.Limits)
	ctx := context.Background()
	if err := stageBundleArchive(ctx, bytes.NewReader(input.Bundle), int64(len(input.Bundle)), members); err != nil {
		return verifiedBundle{}, err
	}
	members.Seal()
	return verifyBundle(ctx, members)
}
func archiveBytes(t testing.TB, archive bundleArchive) []byte {
	t.Helper()
	var output bytes.Buffer
	if err := archive.Write(&output); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

type testArchiveSource struct{ *bytes.Reader }

func (testArchiveSource) Close() error { return nil }

func buildFixtureBundleArchive(t testing.TB, input manifestInput, files map[string][]byte) (bundleArchive, error) {
	t.Helper()
	limits := Limits{}
	if input.Limits != nil {
		limits = *input.Limits
	}
	members := newArchiveTestMembers(t, limits)
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if _, err := members.WriteMember(context.Background(), path, bytes.NewReader(files[path])); err != nil {
			return bundleArchive{}, err
		}
	}
	return buildBundleArchive(context.Background(), input, members)
}
func fixtureMembers(t testing.TB, files map[string][]byte) *memberStore {
	t.Helper()
	members := newArchiveTestMembers(t, Limits{})
	for path, data := range files {
		if _, err := members.WriteMember(context.Background(), path, bytes.NewReader(data)); err != nil {
			t.Fatal(err)
		}
	}
	return members
}

// The worker uses the same disk-backed fake through the full staged-object port.
// The fake owns only test-directory bytes; production lifecycle has its own
// real repository and janitor qualification.
type workerArchiveStaging struct{ disk *archiveTestStaging }

func newWorkerArchiveStaging(t testing.TB) *workerArchiveStaging {
	t.Helper()
	return &workerArchiveStaging{disk: &archiveTestStaging{root: t.TempDir(), refs: map[string]string{}}}
}
func (s *workerArchiveStaging) AllocateStream(ctx context.Context, operation, profile string, source io.Reader, maximum int64) (stagedobjects.SealedReference, error) {
	if operation == "" || profile == "" {
		return stagedobjects.SealedReference{}, stagedobjects.ErrScopeDenied
	}
	return s.disk.AllocateStream(ctx, "archive-test", source, maximum)
}
func (s *workerArchiveStaging) OpenSealed(ctx context.Context, operation, profile string, ref stagedobjects.SealedReference) (io.ReadCloser, error) {
	if operation == "" || profile == "" {
		return nil, stagedobjects.ErrScopeDenied
	}
	return s.disk.OpenSealed(ctx, "archive-test", ref)
}
func (s *workerArchiveStaging) Allocate(ctx context.Context, operation, profile string, payload []byte) (stagedobjects.Reference, error) {
	ref, err := s.AllocateStream(ctx, operation, profile, bytes.NewReader(payload), 64<<20)
	return ref.Reference, err
}
func (s *workerArchiveStaging) Abandon(_ context.Context, ref stagedobjects.Reference) error {
	if name := s.disk.refs[ref.StagingID]; name != "" {
		delete(s.disk.refs, ref.StagingID)
		return os.Remove(name)
	}
	return nil
}
