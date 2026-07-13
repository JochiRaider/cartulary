package graphprojection

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AdmitRetainedProjection is the persistence-adapter seam for the retained
// lifecycle. It does not expose a transport admission API.
func AdmitRetainedProjection(data []byte, nonce string, acceptedAt time.Time) (ProjectionRun, error) {
	if strings.TrimSpace(nonce) == "" {
		return ProjectionRun{}, fmt.Errorf("graphprojection: retained projection nonce is required")
	}
	return admitProjectionInput(data, admitOptions{ProjectionRunNonce: nonce, AcceptedAt: acceptedAt})
}

// DeriveGraphViewID is the stable identity helper used by Graph Projection
// adapters and conformance fixtures. It does not admit or project input.
func DeriveGraphViewID(graphViewKey string) (string, error) {
	return deriveGraphViewID(graphViewKey)
}

// ProjectAdmittedRetainedProjection applies deterministic graph derivation to
// an admitted retained run.
func ProjectAdmittedRetainedProjection(ctx context.Context, run ProjectionRun, generatedAt time.Time, previousProjectionRunID *string) (ProjectionRun, error) {
	return projectAdmittedRunWithContext(ctx, run, projectOptions{GeneratedAt: generatedAt, PreviousProjectionRunID: previousProjectionRunID})
}

// FailAdmittedRetainedProjection constructs the one-issue terminal result used
// by persistence adapters when work fails after admission but before ordinary
// validation or publication can complete.
func FailAdmittedRetainedProjection(run ProjectionRun, completedAt time.Time, reasonCode string) ProjectionRun {
	issue := run.issue("fatal", "projection_computation_failed", "graph_view", run.GraphViewID, nil, map[string]any{"reason_code": reasonCode})
	if issue.Code != "projection_computation_failed" {
		issue = run.computationFailureIssue("implementation_invariant_failed")
		reasonCode = "implementation_invariant_failed"
	}
	completedAt = completedAt.UTC()
	run.State = RunStateFailed
	run.GraphView = nil
	run.ProjectionOutputDigest = ""
	run.GeneratedAt = nil
	run.CompletedAt = &completedAt
	run.FailureReason = truncateScalars(reasonCode, graphProjectionLimits.MaxFailureReasonLength)
	run.ValidationSummary = validationSummary(run, []ValidationIssue{issue})
	return run
}

// CanonicalJSON is restricted by boundary guards to Graph Projection adapters
// and fixture verification. It is not a consumer-facing serialization API.
func CanonicalJSON(value any) ([]byte, error) {
	return canonicalJSON(value)
}

func FormatLifecycleTimestamp(value time.Time) string {
	return formatLifecycleTimestamp(value)
}

func SortVertices(vertices []Vertex) {
	sortVertices(vertices)
}

func SortEdges(edges []Edge) {
	sortEdges(edges)
}

// ProjectionDigestTranscripts returns the exact canonical tuple bytes used by
// the two input digests. Conformance fixtures compare these before hashes.
func ProjectionDigestTranscripts(data []byte) ([]byte, []byte, error) {
	run, err := AdmitRetainedProjection(data, "fixture-transcript", time.Unix(0, 0).UTC())
	if err != nil {
		return nil, nil, err
	}
	config, err := projectionConfigDigestTranscript(run.Request)
	if err != nil {
		return nil, nil, err
	}
	source, err := projectionSourceDigestTranscript(run.Request)
	if err != nil {
		return nil, nil, err
	}
	return config, source, nil
}
