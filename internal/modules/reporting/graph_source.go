package reporting

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

// GraphSourceProvider is implemented by each authoritative source owner. The
// provider validates owner state through the caller's transaction and never
// asks Reporting to interpret a mutable declaration.
type GraphSourceProvider interface {
	SourceOwnerID() string
	GraphStateReady() bool
	ReadSelectedResultTx(context.Context, pgx.Tx, uuid.UUID, graphprojection.ResultBindingV2) (graphsourcecontract.Result, error)
}

type GraphSourceRegistry struct {
	providers map[string]GraphSourceProvider
}

type resolvedGraphResult struct {
	Ref             sourceProjectionRef                 `json:"projection_ref"`
	Result          graphprojection.CompletedResultV2   `json:"result"`
	LabelCandidates graphsourcecontract.LabelCandidates `json:"label_candidates"`
}

func NewGraphSourceRegistry(providers ...GraphSourceProvider) (*GraphSourceRegistry, error) {
	registry := &GraphSourceRegistry{providers: make(map[string]GraphSourceProvider, len(providers))}
	for _, provider := range providers {
		if provider == nil || !validReportingIdentifier(provider.SourceOwnerID()) {
			return nil, errors.New("reporting graph source provider identity is invalid")
		}
		if _, exists := registry.providers[provider.SourceOwnerID()]; exists {
			return nil, fmt.Errorf("reporting graph source provider %q is duplicated", provider.SourceOwnerID())
		}
		registry.providers[provider.SourceOwnerID()] = provider
	}
	return registry, nil
}

func (registry *GraphSourceRegistry) readSelectedTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, refs []sourceProjectionRef) ([]resolvedGraphResult, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	if registry == nil || tx == nil || incidentID == uuid.Nil {
		return nil, invalidGraphProjectionRef("graph_projection_not_bound")
	}
	results := make([]resolvedGraphResult, 0, len(refs))
	for _, ref := range refs {
		provider := registry.providers[ref.SourceOwnerID]
		if provider == nil || !provider.GraphStateReady() {
			return nil, invalidGraphProjectionRef("graph_projection_not_bound")
		}
		contribution, err := provider.ReadSelectedResultTx(ctx, tx, incidentID, ref.binding())
		if err != nil {
			return nil, mapGraphSourceError(err)
		}
		if contribution.Projection.Binding != ref.binding() || !validGraphLabelCandidates(contribution) {
			return nil, invalidGraphProjectionRef("graph_projection_digest_mismatch")
		}
		results = append(results, resolvedGraphResult{Ref: ref, Result: contribution.Projection, LabelCandidates: contribution.LabelCandidates})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Ref.GraphViewID < results[j].Ref.GraphViewID })
	return results, nil
}

func validGraphLabelCandidates(result graphsourcecontract.Result) bool {
	candidates := result.LabelCandidates
	if candidates.SchemaID != graphsourcecontract.SchemaID ||
		candidates.SourceProjectionRef != result.Projection.Binding ||
		len(candidates.VertexLabelCandidates) != len(result.Projection.Vertices) ||
		len(candidates.EdgeLabelCandidates) != len(result.Projection.Edges) {
		return false
	}
	for index, candidate := range candidates.VertexLabelCandidates {
		if candidate.ProjectedVertexID != result.Projection.Vertices[index].VertexID ||
			!validGraphLabelComponent(candidate.Endpoint, "endpoint_value") {
			return false
		}
	}
	for index, candidate := range candidates.EdgeLabelCandidates {
		if candidate.ProjectedEdgeID != result.Projection.Edges[index].EdgeID ||
			!validGraphLabelComponent(candidate.Protocol, "protocol") ||
			!validGraphLabelComponent(candidate.DestinationPort, "destination_port") {
			return false
		}
		switch candidate.Kind {
		case "default_flow_edge_v1":
			if candidate.BucketStartUTC != nil || candidate.BucketEndUTC != nil {
				return false
			}
		case "time_bucket_v1":
			if candidate.BucketStartUTC == nil || candidate.BucketEndUTC == nil ||
				!validGraphLabelComponent(*candidate.BucketStartUTC, "bucket_start_utc") ||
				!validGraphLabelComponent(*candidate.BucketEndUTC, "bucket_end_utc") {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func validGraphLabelComponent(component graphsourcecontract.LabelComponent, kind string) bool {
	if component.ComponentKind != kind || component.FieldPath == "" || component.SourceObjectRef == "" ||
		component.Classification != graphsourcecontract.ClassificationDerivedAnalytic ||
		len(component.DisclosurePartitions) != 1 || component.DisclosurePartitions[0] != graphsourcecontract.DisclosurePartitionInternal ||
		component.RedactionBehavior != graphsourcecontract.RedactionInternalOnly {
		return false
	}
	switch component.State {
	case graphsourcecontract.ComponentStatePresent:
		if component.ValueKind == graphsourcecontract.ValueKindString {
			return component.StringValue != ""
		}
		return component.ValueKind == graphsourcecontract.ValueKindInteger
	case graphsourcecontract.ComponentStateAbsent, graphsourcecontract.ComponentStateRemoved, graphsourcecontract.ComponentStateMissing:
		return component.StringValue == "" && component.IntegerValue == 0
	default:
		return false
	}
}

func mapGraphSourceError(err error) error {
	switch {
	case errors.Is(err, graphsourcecontract.ErrNotSelected):
		return invalidGraphProjectionRef("graph_projection_not_completed")
	case errors.Is(err, graphsourcecontract.ErrSourceStale):
		return invalidGraphProjectionRef("graph_projection_stale")
	case errors.Is(err, graphprojection.ErrResultV2BindingMismatch), errors.Is(err, graphprojection.ErrResultV2IdentityConflict):
		return invalidGraphProjectionRef("graph_projection_digest_mismatch")
	case errors.Is(err, graphprojection.ErrResultV2NotFound):
		return invalidGraphProjectionRef("graph_projection_not_bound")
	default:
		return err
	}
}
