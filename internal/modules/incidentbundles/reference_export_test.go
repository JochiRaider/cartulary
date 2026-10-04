package incidentbundles

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type unreadableContainer struct{ read bool }

func (r *unreadableContainer) Read([]byte) (int, error) {
	r.read = true
	return 0, errors.New("unexpected allocation/read")
}

func testBoundedContainerExport(t *testing.T) {
	t.Helper()
	limits := Limits{Archives: ArchiveLimits{MaxMembers: 5}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 10}}
	files := map[string][]byte{"data/existing": []byte("1234")}
	sink := boundedContainerSink(files, limits)
	denied := &unreadableContainer{}
	if err := sink(context.Background(), "ext/reference_packs/over", 7, denied); err == nil || denied.read {
		t.Fatal("over-budget container read")
	}
	if err := sink(context.Background(), "ext/reference_packs/exact", 6, bytes.NewReader([]byte("123456"))); err != nil {
		t.Fatal("exact remaining budget rejected", err)
	}
	if len(files) != 2 {
		t.Fatal("container omitted")
	}
	if err := sink(context.Background(), "ext/reference_packs/next", 1, denied); err == nil || denied.read {
		t.Fatal("cumulative budget not enforced")
	}
	for _, body := range []string{"12", "1234"} {
		target := map[string][]byte{}
		if err := boundedContainerSink(target, limits)(context.Background(), "ext/reference_packs/length", 3, bytes.NewReader([]byte(body))); err == nil || len(target) != 0 {
			t.Fatal("short/long stream published")
		}
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := boundedContainerSink(map[string][]byte{}, limits)(canceled, "ext/reference_packs/canceled", 1, denied); !errors.Is(err, context.Canceled) || denied.read {
		t.Fatal("canceled stream read", err)
	}
	memberLimits := limits
	memberLimits.Archives.MaxMembers = 2
	if err := boundedContainerSink(map[string][]byte{}, memberLimits)(context.Background(), "ext/reference_packs/member", 1, denied); err == nil || denied.read {
		t.Fatal("outer metadata members not counted")
	}
	// Whole archive construction includes its actual generated manifest and
	// checksum bytes in the export limit, beyond the container copy budget.
	input := manifestInput{BundleID: "22222222-2222-4222-8222-222222222222", IncidentID: "11111111-1111-4111-8111-111111111111", IncidentKey: "LIMIT", ExportedAt: "2026-10-03T00:00:00Z", ReferencePackMode: referencePackModeRefsOnly}
	source := minimalRequiredBundleFiles()
	archive, err := buildBundleArchive(input, source)
	if err != nil {
		t.Fatal(err)
	}
	members, err := readBundleArchive(archive.Bytes, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	total := int64(0)
	for _, member := range members {
		total += int64(len(member))
	}
	boundary := Limits{Archives: ArchiveLimits{MaxMembers: int64(len(members))}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: total}}
	input.Limits = &boundary
	if _, err := buildBundleArchive(input, source); err != nil {
		t.Fatal("complete export equality rejected", err)
	}
	boundary.IncidentBundles.MaxExtractedBytes--
	if _, err := buildBundleArchive(input, source); err == nil {
		t.Fatal("metadata bytes escaped export limit")
	}
	var _ io.Reader = denied
}
