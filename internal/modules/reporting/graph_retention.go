package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const frozenGraphSourcesSchema = "cartulary.reporting_graph_input.v1"

// This private admission object retains source-owner contributions, not public
// presentation bytes. Never put it in a shareable bundle. The independent Job
// row digest prevents replacing these bytes by rehashing an output bundle.
type frozenGraphSources struct {
	SchemaID string                `json:"schema_id"`
	Results  []resolvedGraphResult `json:"results"`
}

func freezeGraphSources(results []resolvedGraphResult) (frozenGraphSources, error) {
	if results == nil {
		results = []resolvedGraphResult{}
	}
	// Own every nested slice, raw JSON value and label pointer. Source-owner
	// operational publication time is not a render input.
	raw, err := json.Marshal(frozenGraphSources{SchemaID: frozenGraphSourcesSchema, Results: results})
	if err != nil {
		return frozenGraphSources{}, err
	}
	var frozen frozenGraphSources
	if err := json.Unmarshal(raw, &frozen); err != nil {
		return frozenGraphSources{}, err
	}
	for i := range frozen.Results {
		frozen.Results[i].Result.PublishedAt = time.Time{}
	}
	return frozen, nil
}

func (registry *GraphSourceRegistry) FreezeTx(ctx context.Context, tx pgx.Tx, incidentID, jobID uuid.UUID, refs []sourceProjectionRef, observedAt time.Time) (frozenGraphSources, error) {
	if err := registry.ValidateAndLeaseTx(ctx, tx, incidentID, jobID, refs, observedAt); err != nil {
		return frozenGraphSources{}, err
	}
	results, err := registry.ReadAndRenewTx(ctx, tx, jobID, refs, observedAt)
	if err != nil {
		return frozenGraphSources{}, err
	}
	frozen, err := freezeGraphSources(results)
	if err != nil {
		return frozenGraphSources{}, err
	}
	if err := validateFrozenGraphSources(frozen, refs); err != nil {
		return frozenGraphSources{}, err
	}
	return frozen, nil
}

func validateFrozenGraphSources(frozen frozenGraphSources, refs []sourceProjectionRef) error {
	invalid := errors.New("reporting: invalid retained graph input")
	if frozen.SchemaID != frozenGraphSourcesSchema || frozen.Results == nil || len(frozen.Results) != len(refs) {
		return invalid
	}
	byView := make(map[string]sourceProjectionRef, len(refs))
	for _, ref := range refs {
		if _, exists := byView[ref.GraphViewID]; exists {
			return invalid
		}
		byView[ref.GraphViewID] = ref
	}
	for i, result := range frozen.Results {
		ref, exists := byView[result.Ref.GraphViewID]
		if !exists || result.Ref != ref || result.Result.Binding != ref.binding() || !result.Result.PublishedAt.IsZero() ||
			(i > 0 && frozen.Results[i-1].Ref.GraphViewID >= result.Ref.GraphViewID) ||
			!validGraphLabelCandidates(graphsourcecontract.Result{Projection: result.Result, LabelCandidates: result.LabelCandidates}) ||
			len(result.Result.Vertices) > graphprojection.MaximumResultVerticesV2 || len(result.Result.Edges) > graphprojection.MaximumResultEdgesV2 {
			return invalid
		}
		delete(byView, result.Ref.GraphViewID)
	}
	return nil
}

func graphSourcesDigest(frozen frozenGraphSources) (string, error) {
	data, err := canonicalJSON(frozen)
	if err != nil {
		return "", err
	}
	return reportingObjectDigest(frozenGraphSourcesSchema, data), nil
}

func validateAdmittedGraphSources(input renderInput, admittedSHA string) error {
	refs, err := decodeSourceProjectionRefs(input.GraphProjectionRefs)
	if err != nil {
		return err
	}
	if err := validateFrozenGraphSources(input.GraphSources, refs); err != nil {
		return err
	}
	digest, err := graphSourcesDigest(input.GraphSources)
	if err != nil || !sha256HexPattern.MatchString(admittedSHA) || digest != admittedSHA {
		return errors.New("reporting: graph input differs from admission")
	}
	return nil
}

func sameFrozenGraphSources(left, right frozenGraphSources) bool {
	a, err := canonicalJSON(left)
	if err != nil {
		return false
	}
	b, err := canonicalJSON(right)
	return err == nil && bytes.Equal(a, b)
}
