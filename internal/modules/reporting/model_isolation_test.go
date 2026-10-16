package reporting

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
)

func testReportingRenderIsolation(t *testing.T) {
	source := fixtureSnapshotSource([]ExportField{{
		SourceFamily: "timeline_event", SourceRecordID: "event", SourceRecordType: "timeline_event",
		Path: "/timeline/event", ContentClass: ContentClassSourceEvidence,
		Value:                   map[string]any{"a_values": []any{"source", true, 9007199254740991}, "capture_state": "rough"},
		DisclosurePartitionRefs: []string{"public"}, SupportRefs: []string{"/timeline/event"},
	}})
	stamp := "2026-05-02T00:00:00.000001Z"
	source.Timeline[0].ActivitySortTS = &stamp
	source.Timeline[0].ActivityParseState, source.Timeline[0].ActivityPrecisionRank, source.Timeline[0].TimeBucket = "parsed", 6, "parsed_time"
	snapshot, _, err := BuildSnapshotModel(IncidentMetadataSnapshot{ID: "00000000-0000-0000-0000-000000000401", Title: "Frozen", Status: "active"}, fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.ReferencePacks = reference_data.SetBinding{Provenance: []reference_data.PackProvenance{{PackKey: "pack"}}}
	frozen, err := canonicalJSON(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	identity := fixtureRenderIdentity()
	identityBefore := *identity.ReleaseID
	partitions := []string{"public"}
	first, firstDigest, err := bindRenderModel(snapshot, identity, ReleaseScopeInternalReview, partitions)
	if err != nil {
		t.Fatal(err)
	}
	previewID := "00000000-0000-0000-0000-000000000404"
	previewIdentity := RenderIdentity{PreviewAttemptID: &previewID, RenderAdmittedAt: identity.RenderAdmittedAt}
	sibling, _, err := bindRenderModel(snapshot, previewIdentity, ReleaseScopeInternalDraft, nil)
	if err != nil {
		t.Fatal(err)
	}
	siblingBefore, _ := canonicalJSON(sibling)
	firstBefore, _ := canonicalJSON(first)
	if firstDigest != reportingObjectDigest(ExportModelSchemaID, firstBefore) {
		t.Fatal("copy changed admitted bytes")
	}
	first.Records[0].Fields[0].Value.([]any)[0] = "changed value"
	first.Records[0].Fields[0].SourceRefs[0] = "/changed/source"
	*first.Records[0].SourceRecordRef.SourceRefID = "/changed/binding"
	first.Sections[0].Blocks[0].Fields[0].Value = "changed section"
	first.TimelineEvents[0].SupportRefs[0] = "/changed/support"
	*first.TimelineEvents[0].ActivitySortTS = "2020-01-01T00:00:00.000000Z"
	first.ReferencePacks.Provenance[0].PackKey = "changed pack"
	*first.ReleaseID = "00000000-0000-0000-0000-000000000499"
	first.RecipientPartitionRefs[0] = "changed partition"
	after, _ := canonicalJSON(snapshot)
	other, _ := canonicalJSON(sibling)
	if !bytes.Equal(frozen, after) || !bytes.Equal(siblingBefore, other) || *identity.ReleaseID != identityBefore || partitions[0] != "public" {
		t.Fatal("render mutation escaped its lifecycle")
	}
	*previewIdentity.PreviewAttemptID = "00000000-0000-0000-0000-000000000498"
	other, _ = canonicalJSON(sibling)
	if !bytes.Equal(siblingBefore, other) {
		t.Fatal("preview identity aliases admission input")
	}
	// Mutating the source after binding cannot alter an already admitted render.
	snapshot.Records[0].Fields[0].Value.([]any)[0] = "new source value"
	other, _ = canonicalJSON(sibling)
	if !bytes.Equal(siblingBefore, other) {
		t.Fatal("source mutation changed a sibling render")
	}

	t.Run("participant cannot mutate caller payload", func(t *testing.T) {
		payload := renderInput{SourceFacts: snapshot.SourceFacts.Clone(), SourceFactsSHA256: fixtureSourceFactsDigest(sibling.SnapshotContent), ActorUserID: "actor", IncidentID: sibling.IncidentID, SnapshotID: sibling.SnapshotID, ExportModel: sibling, ReleaseScope: sibling.ReleaseScope, RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion}
		payload.RedactionProfile, payload.RedactionProfileSHA256, err = ResolveRedactionProfile(payload.RedactionProfileID, payload.RedactionProfileVersion, nil, sibling.SnapshotContent)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := canonicalJSON(payload.ExportModel)
		payload.ExportModelSHA256 = reportingObjectDigest(ExportModelSchemaID, raw)
		before, _ := canonicalJSON(payload)
		worker := reportingJobWorker{renderExport: mutatingRenderExportInvoker{}, timeout: time.Minute}
		if _, err := worker.invokeRenderExport(context.Background(), payload); !errors.Is(err, errMutatingParticipant) {
			t.Fatal("participant failure lost", err)
		}
		after, _ := canonicalJSON(payload)
		if !bytes.Equal(before, after) {
			t.Fatal("participant corrupted its caller on failure")
		}
	})
}

var errMutatingParticipant = errors.New("participant failure after mutation")

type mutatingRenderExportInvoker struct{}

func (mutatingRenderExportInvoker) Invoke(_ context.Context, invocation RenderExportInvocation) (RenderExportResult, error) {
	invocation.ImmutableModel.Records[0].Fields[0].Value.([]any)[0] = "participant mutation"
	*invocation.ImmutableModel.PreviewAttemptID = "00000000-0000-0000-0000-000000000497"
	invocation.ImmutableModel.ReferencePacks.Provenance[0].PackKey = "participant pack"
	return RenderExportResult{}, errMutatingParticipant
}
