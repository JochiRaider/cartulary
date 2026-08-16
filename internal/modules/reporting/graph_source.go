package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const reportingGraphLeaseDuration = 5 * time.Minute
const reportingGraphLeasePurpose = "render"

// GraphSourceProvider is implemented by each authoritative source owner. The
// provider validates owner state through the caller's transaction and never
// asks Reporting to interpret a mutable declaration.
type GraphSourceProvider interface {
	SourceOwnerID() string
	ValidateAndLeaseResultTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, graphprojection.ResultBindingV2, time.Time, time.Time) (graphprojection.ResultLeaseV2, error)
	ReadAndRenewLeasedResult(context.Context, uuid.UUID, graphprojection.ResultBindingV2, time.Time, time.Time) (graphsourcecontract.Result, error)
	ReleaseJobLeasesTx(context.Context, pgx.Tx, uuid.UUID) error
	ReleaseJobLeases(context.Context, uuid.UUID) error
}

type GraphSourceRegistry struct {
	providers map[string]GraphSourceProvider
}

type resolvedGraphResult struct {
	Ref             sourceProjectionRef
	Result          graphprojection.CompletedResultV2
	LabelCandidates graphsourcecontract.LabelCandidates
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

func (registry *GraphSourceRegistry) ValidateAndLeaseTx(ctx context.Context, tx pgx.Tx, incidentID, jobID uuid.UUID, refs []sourceProjectionRef, observedAt time.Time) error {
	if len(refs) == 0 {
		return nil
	}
	if registry == nil || tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil || observedAt.IsZero() {
		return invalidGraphProjectionRef("graph_projection_not_bound")
	}
	for _, ref := range refs {
		provider := registry.providers[ref.SourceOwnerID]
		if provider == nil {
			return invalidGraphProjectionRef("graph_projection_not_bound")
		}
		lease, err := provider.ValidateAndLeaseResultTx(ctx, tx, incidentID, jobID, ref.binding(), observedAt.UTC(), observedAt.UTC().Add(reportingGraphLeaseDuration))
		if err != nil {
			return mapGraphSourceError(err)
		}
		if lease.ProjectionResultID != ref.ProjectionResultID || lease.LeaseOwnerResourceID != jobID.String() {
			return invalidGraphProjectionRef("graph_projection_digest_mismatch")
		}
	}
	return nil
}

func (registry *GraphSourceRegistry) ReadAndRenew(ctx context.Context, jobID uuid.UUID, refs []sourceProjectionRef, observedAt time.Time) ([]resolvedGraphResult, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	if registry == nil || jobID == uuid.Nil || observedAt.IsZero() {
		return nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
	}
	results := make([]resolvedGraphResult, 0, len(refs))
	for _, ref := range refs {
		provider := registry.providers[ref.SourceOwnerID]
		if provider == nil {
			return nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
		}
		sourceResult, err := provider.ReadAndRenewLeasedResult(ctx, jobID, ref.binding(), observedAt.UTC(), observedAt.UTC().Add(reportingGraphLeaseDuration))
		if err != nil {
			return nil, graphSourceRenderError(err)
		}
		if sourceResult.Projection.Binding != ref.binding() || !validGraphLabelCandidates(sourceResult) {
			return nil, newRenderValidationError("graph_projection_unavailable", "graph_projection_digest_mismatch")
		}
		results = append(results, resolvedGraphResult{
			Ref: ref, Result: sourceResult.Projection, LabelCandidates: sourceResult.LabelCandidates,
		})
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

func (registry *GraphSourceRegistry) ReleaseJobLeasesTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) error {
	if registry == nil || jobID == uuid.Nil {
		return nil
	}
	owners := make([]string, 0, len(registry.providers))
	for ownerID := range registry.providers {
		owners = append(owners, ownerID)
	}
	sort.Strings(owners)
	for _, ownerID := range owners {
		if err := registry.providers[ownerID].ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
			return err
		}
	}
	return nil
}

func (registry *GraphSourceRegistry) ReleaseJobLeases(ctx context.Context, jobID uuid.UUID) error {
	if registry == nil || jobID == uuid.Nil {
		return nil
	}
	owners := make([]string, 0, len(registry.providers))
	for ownerID := range registry.providers {
		owners = append(owners, ownerID)
	}
	sort.Strings(owners)
	for _, ownerID := range owners {
		if err := registry.providers[ownerID].ReleaseJobLeases(ctx, jobID); err != nil {
			return err
		}
	}
	return nil
}

func mapGraphSourceError(err error) error {
	switch {
	case errors.Is(err, graphprojection.ErrResultV2NotSelected):
		return invalidGraphProjectionRef("graph_projection_not_completed")
	case errors.Is(err, graphprojection.ErrResultV2SourceStale):
		return invalidGraphProjectionRef("graph_projection_stale")
	case errors.Is(err, graphprojection.ErrResultV2BindingMismatch), errors.Is(err, graphprojection.ErrResultV2IdentityConflict):
		return invalidGraphProjectionRef("graph_projection_digest_mismatch")
	case errors.Is(err, graphprojection.ErrResultV2NotFound):
		return invalidGraphProjectionRef("graph_projection_not_bound")
	default:
		return err
	}
}

func graphSourceRenderError(err error) error {
	mapped := mapGraphSourceError(err)
	var requestErr *InvalidReleaseRequestError
	if errors.As(mapped, &requestErr) {
		return newRenderValidationError("graph_projection_unavailable", requestErr.ReasonCode)
	}
	if errors.Is(err, graphprojection.ErrResultV2LeaseNotFound) {
		return newRenderValidationError("graph_projection_unavailable", "graph_projection_not_completed")
	}
	return mapped
}

func validReportingIdentifier(value string) bool {
	return value != "" && len(value) <= 255
}

// ReconcileGraphProjectionRestoreTx recreates leases for restored nonterminal
// Reporting release jobs. Durable payloads remain Reporting-owned; callers
// provide only the exact immutable bindings reproduced by Graph Recovery.
func ReconcileGraphProjectionRestoreTx(
	ctx context.Context,
	tx pgx.Tx,
	rebuiltBindings []graphprojection.ResultBindingV2,
	observedAt time.Time,
) (int, int, error) {
	if ctx == nil || tx == nil || observedAt.IsZero() {
		return 0, 0, errors.New("reporting graph restore reconciliation requires transaction and time")
	}
	rebuilt := make(map[graphprojection.ResultBindingV2]struct{}, len(rebuiltBindings))
	for _, binding := range rebuiltBindings {
		if binding.ProjectionResultID == "" {
			return 0, 0, errors.New("reporting graph restore reconciliation received an invalid binding")
		}
		rebuilt[binding] = struct{}{}
	}
	type restoredReleaseJob struct {
		jobID       uuid.UUID
		requestJSON []byte
	}
	rows, err := tx.Query(ctx, `
SELECT job.job_id, payload.request_json
  FROM jobs job
  JOIN reporting_job_payloads payload ON payload.job_id = job.job_id
 WHERE job.status IN ('queued', 'running', 'cancel_requested')
   AND payload.job_kind = 'release_create'
 ORDER BY job.job_id
`)
	if err != nil {
		return 0, 0, fmt.Errorf("enumerate restored Reporting Graph jobs: %w", err)
	}
	restoredJobs := make([]restoredReleaseJob, 0)
	for rows.Next() {
		var restored restoredReleaseJob
		if err := rows.Scan(&restored.jobID, &restored.requestJSON); err != nil {
			rows.Close()
			return 0, 0, fmt.Errorf("scan restored Reporting Graph job: %w", err)
		}
		restoredJobs = append(restoredJobs, restored)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, 0, fmt.Errorf("iterate restored Reporting Graph jobs: %w", err)
	}
	rows.Close()

	leaseWriter, err := postgresresult.NewLeaseWriter(tx)
	if err != nil {
		return 0, 0, err
	}
	reconciledJobs, reconciledLeases := 0, 0
	observedAt = observedAt.UTC()
	for _, restored := range restoredJobs {
		var payload struct {
			GraphProjectionRefs json.RawMessage `json:"graph_projection_refs"`
		}
		if err := json.Unmarshal(restored.requestJSON, &payload); err != nil {
			return 0, 0, fmt.Errorf("decode restored Reporting Graph job %s: %w", restored.jobID, err)
		}
		if len(payload.GraphProjectionRefs) == 0 {
			continue
		}
		if apiErr := validateSourceProjectionRefs(payload.GraphProjectionRefs, "invalid_release_request"); apiErr != nil {
			return 0, 0, fmt.Errorf("restored Reporting Graph job %s has invalid references", restored.jobID)
		}
		refs, err := decodeSourceProjectionRefs(payload.GraphProjectionRefs)
		if err != nil {
			return 0, 0, fmt.Errorf("decode restored Reporting Graph references for job %s: %w", restored.jobID, err)
		}
		if len(refs) == 0 {
			continue
		}
		if err := jobs.ReconcileRestoredNonterminalTx(ctx, tx, restored.jobID, ReleaseCreateJobKind); err != nil {
			return 0, 0, err
		}
		for _, ref := range refs {
			binding := ref.binding()
			if _, present := rebuilt[binding]; !present {
				return 0, 0, fmt.Errorf("restored Reporting Graph job %s references an unavailable exact result", restored.jobID)
			}
			if _, err := leaseWriter.AcquireLease(ctx, graphprojection.ResultLeaseV2{
				LeaseID:              uuid.NewSHA1(restored.jobID, []byte(binding.ProjectionResultID+":"+reportingGraphLeasePurpose)).String(),
				ProjectionResultID:   binding.ProjectionResultID,
				LeaseOwnerID:         "snapshot_reporting",
				LeaseOwnerResourceID: restored.jobID.String(),
				LeasePurpose:         reportingGraphLeasePurpose,
				LeasedUntil:          observedAt.Add(reportingGraphLeaseDuration),
				CreatedAt:            observedAt,
				RenewedAt:            observedAt,
			}); err != nil {
				return 0, 0, err
			}
			reconciledLeases++
		}
		reconciledJobs++
	}
	return reconciledJobs, reconciledLeases, nil
}
