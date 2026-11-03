package reporting

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

type fakeGraphSourceProvider struct {
	calls       int
	unavailable bool
	ownerID     string
	result      graphsourcecontract.Result
	err         error
}

func (provider *fakeGraphSourceProvider) SourceOwnerID() string { return provider.ownerID }
func (provider *fakeGraphSourceProvider) GraphStateReady() bool { return !provider.unavailable }

func (provider *fakeGraphSourceProvider) ReadSelectedResultTx(context.Context, pgx.Tx, uuid.UUID, graphprojection.ResultBindingV2) (graphsourcecontract.Result, error) {
	provider.calls++
	return provider.result, provider.err
}

func TestGraphSourceRegistryReadsExactResultAndOwnsReasonMapping(t *testing.T) {
	t.Run("frozen source retention", testFrozenGraphSources)
	t.Run("inactive provider is never invoked", func(t *testing.T) {
		ref := testSourceProjectionRef("gv_valid", "gpr_valid")
		provider := &fakeGraphSourceProvider{ownerID: ref.SourceOwnerID, unavailable: true}
		registry, err := NewGraphSourceRegistry(provider)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := registry.readSelectedTx(t.Context(), borrowedReportingTx{}, uuid.New(), []sourceProjectionRef{ref}); err == nil {
			t.Fatal("inactive admission succeeded")
		}
		if provider.calls != 0 {
			t.Fatal("inactive source callback executed")
		}
	})
	ref := testSourceProjectionRef("gv_valid", "gpr_valid")
	result := graphprojection.CompletedResultV2{Binding: ref.binding()}
	registry, err := NewGraphSourceRegistry(&fakeGraphSourceProvider{ownerID: ref.SourceOwnerID, result: graphsourcecontract.Result{
		Projection: result, LabelCandidates: graphLabelCandidatesForResult(result),
	}})
	if err != nil {
		t.Fatalf("new graph source registry: %v", err)
	}
	resolved, err := registry.readSelectedTx(context.Background(), borrowedReportingTx{}, uuid.New(), []sourceProjectionRef{ref})
	if err != nil {
		t.Fatalf("read exact selected result: %v", err)
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
	for _, owner := range []string{"", " owner", "owner\n", "owner/child", "owner#child", "owner\\child", "https://owner", string([]byte{0xff})} {
		if _, err := NewGraphSourceRegistry(&fakeGraphSourceProvider{ownerID: owner}); err == nil {
			t.Fatal("invalid source owner identity accepted")
		}
	}
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
	_, err = registry.readSelectedTx(context.Background(), borrowedReportingTx{}, uuid.New(), []sourceProjectionRef{ref})
	var renderErr *InvalidReleaseRequestError
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
