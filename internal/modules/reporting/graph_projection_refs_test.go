package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

type fakeProjectionBindingReader struct {
	bindings map[string]graphprojection.ProjectionBinding
	err      error
}

func (f fakeProjectionBindingReader) LookupProjectionBinding(_ context.Context, runID string) (graphprojection.ProjectionBinding, error) {
	if f.err != nil {
		return graphprojection.ProjectionBinding{}, f.err
	}
	binding, ok := f.bindings[runID]
	if !ok {
		return graphprojection.ProjectionBinding{}, graphprojection.ErrProjectionRunNotFound
	}
	return binding, nil
}

func TestValidateReleaseGraphProjectionRefsOwnsReasonMapping(t *testing.T) {
	digestA := strings.Repeat("a", 64)
	digestB := strings.Repeat("b", 64)
	ref := sourceProjectionRef{ProjectionSchemaID: graphprojection.ProjectionSchemaID, GraphViewID: "gv_valid", SourceSnapshotID: "snapshot", ProjectionRunID: "gpr_valid", ProjectionVersion: "v1", ProjectionConfigDigest: digestA, ProjectionSourceDigest: digestA, ProjectionOutputDigest: digestA}
	binding := graphprojection.ProjectionBinding{ProjectionRunID: ref.ProjectionRunID, GraphViewID: ref.GraphViewID, SourceSnapshotID: ref.SourceSnapshotID, ProjectionVersion: ref.ProjectionVersion, State: graphprojection.RunStateAvailable, ProjectionConfigDigest: digestA, ProjectionSourceDigest: digestA, ProjectionOutputDigest: digestA}
	reader := fakeProjectionBindingReader{bindings: map[string]graphprojection.ProjectionBinding{ref.ProjectionRunID: binding}}
	if err := validateReleaseGraphProjectionRefs(context.Background(), reader, "snapshot", []sourceProjectionRef{ref}); err != nil {
		t.Fatalf("validate completed binding: %v", err)
	}
	cases := []struct {
		name       string
		ref        sourceProjectionRef
		reader     graphprojection.ProjectionBindingReader
		snapshotID string
		wantReason string
	}{
		{name: "not bound", ref: withProjectionRun(ref, "gpr_missing"), reader: reader, snapshotID: "snapshot", wantReason: "graph_projection_not_bound"},
		{name: "not completed", ref: ref, reader: fakeProjectionBindingReader{bindings: map[string]graphprojection.ProjectionBinding{ref.ProjectionRunID: withBindingState(binding, graphprojection.RunStateComputing)}}, snapshotID: "snapshot", wantReason: "graph_projection_not_completed"},
		{name: "stale", ref: ref, reader: reader, snapshotID: "other", wantReason: "graph_projection_stale"},
		{name: "digest mismatch", ref: withProjectionOutputDigest(ref, digestB), reader: reader, snapshotID: "snapshot", wantReason: "graph_projection_digest_mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReleaseGraphProjectionRefs(context.Background(), tc.reader, tc.snapshotID, []sourceProjectionRef{tc.ref})
			var requestErr *InvalidReleaseRequestError
			if !errors.As(err, &requestErr) || requestErr.Field != "graph_projection_refs" || requestErr.ReasonCode != tc.wantReason {
				t.Fatalf("validation error = %#v want %s", err, tc.wantReason)
			}
		})
	}
}

func withProjectionRun(ref sourceProjectionRef, value string) sourceProjectionRef {
	ref.ProjectionRunID = value
	return ref
}
func withProjectionOutputDigest(ref sourceProjectionRef, value string) sourceProjectionRef {
	ref.ProjectionOutputDigest = value
	return ref
}
func withBindingState(binding graphprojection.ProjectionBinding, value graphprojection.RunState) graphprojection.ProjectionBinding {
	binding.State = value
	return binding
}
