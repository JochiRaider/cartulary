package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

type fakeGraphSourceProvider struct {
	ownerID string
	result  graphprojection.CompletedResultV2
	err     error
}

func (provider *fakeGraphSourceProvider) SourceOwnerID() string { return provider.ownerID }

func (provider *fakeGraphSourceProvider) ValidateAndLeaseResultTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, graphprojection.ResultBindingV2, time.Time, time.Time) (graphprojection.ResultLeaseV2, error) {
	return graphprojection.ResultLeaseV2{}, provider.err
}

func (provider *fakeGraphSourceProvider) ReadAndRenewLeasedResult(context.Context, uuid.UUID, graphprojection.ResultBindingV2, time.Time, time.Time) (graphprojection.CompletedResultV2, error) {
	return provider.result, provider.err
}

func (*fakeGraphSourceProvider) ReleaseJobLeasesTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (*fakeGraphSourceProvider) ReleaseJobLeases(context.Context, uuid.UUID) error { return nil }

func TestGraphSourceRegistryReadsExactResultAndOwnsReasonMapping(t *testing.T) {
	ref := testSourceProjectionRef("gv_valid", "gpr_valid")
	result := graphprojection.CompletedResultV2{Binding: ref.binding()}
	registry, err := NewGraphSourceRegistry(&fakeGraphSourceProvider{ownerID: ref.SourceOwnerID, result: result})
	if err != nil {
		t.Fatalf("new graph source registry: %v", err)
	}
	resolved, err := registry.ReadAndRenew(context.Background(), uuid.New(), []sourceProjectionRef{ref}, time.Now().UTC())
	if err != nil {
		t.Fatalf("read exact leased result: %v", err)
	}
	if len(resolved) != 1 || resolved[0].Ref != ref || resolved[0].Result.Binding != ref.binding() {
		t.Fatalf("resolved exact result = %#v", resolved)
	}

	cases := []struct {
		name       string
		err        error
		wantReason string
	}{
		{name: "not bound", err: graphprojection.ErrResultV2NotFound, wantReason: "graph_projection_not_bound"},
		{name: "not completed", err: graphprojection.ErrResultV2NotSelected, wantReason: "graph_projection_not_completed"},
		{name: "stale", err: graphprojection.ErrResultV2SourceStale, wantReason: "graph_projection_stale"},
		{name: "binding mismatch", err: graphprojection.ErrResultV2BindingMismatch, wantReason: "graph_projection_digest_mismatch"},
		{name: "identity conflict", err: graphprojection.ErrResultV2IdentityConflict, wantReason: "graph_projection_digest_mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := mapGraphSourceError(tc.err)
			var requestErr *InvalidReleaseRequestError
			if !errors.As(mapped, &requestErr) || requestErr.Field != "graph_projection_refs" || requestErr.ReasonCode != tc.wantReason {
				t.Fatalf("validation error = %#v want %s", mapped, tc.wantReason)
			}
		})
	}
}

func TestGraphSourceRegistryRejectsDuplicateAndUnknownOwners(t *testing.T) {
	provider := &fakeGraphSourceProvider{ownerID: "network_flow_activity"}
	if _, err := NewGraphSourceRegistry(provider, provider); err == nil {
		t.Fatal("duplicate source owner provider must fail")
	}
	registry, err := NewGraphSourceRegistry(provider)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	ref := testSourceProjectionRef("gv_valid", "gpr_valid")
	ref.SourceOwnerID = "unknown"
	_, err = registry.ReadAndRenew(context.Background(), uuid.New(), []sourceProjectionRef{ref}, time.Now().UTC())
	var renderErr *renderValidationError
	if !errors.As(err, &renderErr) || renderErr.ReasonCode != "graph_projection_not_bound" {
		t.Fatalf("unknown owner error = %#v", err)
	}
}

func testSourceProjectionRef(graphViewID, resultID string) sourceProjectionRef {
	digest := strings.Repeat("a", 64)
	return sourceProjectionRef{
		SourceOwnerID:                 "network_flow_activity",
		GraphViewID:                   graphViewID,
		ProjectionResultID:            resultID,
		SourceSnapshotID:              "snapshot",
		ProjectionSchemaID:            graphprojection.ProjectionSchemaIDV2,
		ProjectionVersion:             "v2",
		NormalizedConfigurationSHA256: digest,
		NormalizedSourceSHA256:        digest,
		CanonicalOutputSHA256:         digest,
	}
}
