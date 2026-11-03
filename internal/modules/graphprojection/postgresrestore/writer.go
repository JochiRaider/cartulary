package postgresrestore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

// Post-claim reconstruction uses the ordinary runtime role's existing DELETE
// grants. It needs no Recovery pool or table-wide TRUNCATE privilege.
const clearGraphProjectionTablesSQL = `DELETE FROM graph_projection_result_edges;
DELETE FROM graph_projection_result_vertices;
DELETE FROM graph_projection_results;`

const committedProofTimeout = 30 * time.Second

// DerivedStateReconciler is supplied by application assembly. It lets the
// claimed reconstruction reconcile Network Flow materialization jobs in
// the same transaction without importing those subsystems into Graph.
type DerivedStateReconciler interface {
	ReconcileGraphProjectionDerivedState(context.Context, pgx.Tx) (int, error)
}

type Writer struct {
	cleared         bool
	tx              pgx.Tx
	committedReader restoreQueryer
	reconciler      DerivedStateReconciler
}

var _ graphrestore.RestorePublisher = (*Writer)(nil)

// New binds publication to the transaction already used for source enumeration.
// ReplaceAll commits that transaction only after exact owner reconciliation and
// postconditions. The caller rolls it back if preparation never reaches ReplaceAll.
func New(tx pgx.Tx, committedReader restoreQueryer, reconciler DerivedStateReconciler) (*Writer, error) {
	if tx == nil || committedReader == nil || reconciler == nil {
		return nil, graphrestore.NewRestoreError(graphrestore.RestoreErrorInvalidRequest)
	}
	return &Writer{tx: tx, committedReader: committedReader, reconciler: reconciler}, nil
}

// Clear deletes only the derived result envelopes, vertices and edges after
// metadata admission and before source enumeration. The caller rolls back if
// preparation fails; ReplaceAll commits only after publication, reconciliation
// and postcondition verification.
func (writer *Writer) Clear(ctx context.Context) error {
	if writer == nil || writer.tx == nil || writer.cleared || ctx == nil {
		return publicationError(graphprojection.ErrResultV2Invalid)
	}
	if _, err := writer.tx.Exec(ctx, clearGraphProjectionTablesSQL); err != nil {
		return publicationError(err)
	}
	writer.cleared = true
	return nil
}

func (writer *Writer) ReplaceAll(ctx context.Context, plan graphrestore.RestorePublicationPlan) (graphrestore.RestorePublicationProof, error) {
	if writer == nil || writer.tx == nil || writer.committedReader == nil || writer.reconciler == nil || !writer.cleared || ctx == nil || !validPlan(plan) {
		return graphrestore.RestorePublicationProof{}, &graphrestore.RestorePublicationError{Cause: graphprojection.ErrResultV2Invalid}
	}
	if err := ctx.Err(); err != nil {
		return graphrestore.RestorePublicationProof{}, &graphrestore.RestorePublicationError{Cause: err}
	}
	tx := writer.tx
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		return graphrestore.RestorePublicationProof{}, publicationError(err)
	}
	for _, staged := range plan.Projections {
		if err := publisher.PublishResult(ctx, staged.Result); err != nil {
			return graphrestore.RestorePublicationProof{}, publicationError(err)
		}
	}
	jobs, err := writer.reconciler.ReconcileGraphProjectionDerivedState(ctx, tx)
	if err != nil {
		return graphrestore.RestorePublicationProof{}, publicationError(err)
	}
	if jobs < 0 {
		return graphrestore.RestorePublicationProof{}, publicationError(graphprojection.ErrResultV2Invalid)
	}
	if err := verifyPublishedState(ctx, tx, plan); err != nil {
		return graphrestore.RestorePublicationProof{}, publicationError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return graphrestore.RestorePublicationProof{}, &graphrestore.RestorePublicationError{Indeterminate: !errors.Is(err, pgx.ErrTxCommitRollback), Cause: err}
	}
	proofContext, cancelProof := context.WithTimeout(context.WithoutCancel(ctx), committedProofTimeout)
	defer cancelProof()
	if err := verifyPublishedState(proofContext, writer.committedReader, plan); err != nil {
		return graphrestore.RestorePublicationProof{}, &graphrestore.RestorePublicationError{Indeterminate: true, Cause: err}
	}
	return graphrestore.RestorePublicationProof{
		RebuiltViews:                  append([]graphrestore.RestoreRebuiltView{}, plan.RebuiltViews...),
		ReconciledNonterminalJobCount: jobs,
		PostconditionSHA256:           plan.PostconditionSHA256,
	}, nil
}

func publicationError(cause error) error {
	return &graphrestore.RestorePublicationError{Cause: cause}
}

func validPlan(plan graphrestore.RestorePublicationPlan) bool {
	if plan.RestoreOperationID.String() == "00000000-0000-0000-0000-000000000000" ||
		plan.TargetGenerationID.String() == "00000000-0000-0000-0000-000000000000" ||
		len(plan.ClearedTableIDs) != len(graphrestore.RestoreGraphTableIDs()) ||
		len(plan.Projections) != len(plan.RebuiltViews) || len(plan.PostconditionSHA256) != 64 {
		return false
	}
	for index, tableID := range graphrestore.RestoreGraphTableIDs() {
		if plan.ClearedTableIDs[index] != tableID {
			return false
		}
	}
	for index := range plan.Projections {
		staged, rebuilt := plan.Projections[index], plan.RebuiltViews[index]
		if staged.Result.Binding.ProjectionResultID == "" || staged.SourceRegistrationID != rebuilt.SourceRegistrationID ||
			staged.CandidateID != rebuilt.CandidateID || staged.Result.Binding.ProjectionResultID != rebuilt.ProjectionResultID {
			return false
		}
	}
	return true
}

type restoreQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func verifyPublishedState(ctx context.Context, queryer restoreQueryer, plan graphrestore.RestorePublicationPlan) error {
	var results, vertices, edges int
	if err := queryer.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM graph_projection_results),
    (SELECT COUNT(*) FROM graph_projection_result_vertices),
    (SELECT COUNT(*) FROM graph_projection_result_edges)
`).Scan(&results, &vertices, &edges); err != nil {
		return fmt.Errorf("verify Graph restore aggregate state: %w", err)
	}
	wantVertices, wantEdges := 0, 0
	for _, rebuilt := range plan.RebuiltViews {
		wantVertices += rebuilt.VertexCount
		wantEdges += rebuilt.EdgeCount
	}
	if results != len(plan.Projections) || vertices != wantVertices || edges != wantEdges {
		return fmt.Errorf("graph restore aggregate postcondition mismatch")
	}
	reader, err := postgresresult.NewReader(queryer)
	if err != nil {
		return err
	}
	for _, staged := range plan.Projections {
		loaded, err := reader.ReadExactResult(ctx, staged.Result.Binding)
		if err != nil {
			return fmt.Errorf("verify Graph restore exact result: %w", err)
		}
		if !sameCompletedResultBytes(loaded, staged.Result) {
			return fmt.Errorf("verify Graph restore exact result: byte mismatch")
		}
	}
	return nil
}

func sameCompletedResultBytes(left, right graphprojection.CompletedResultV2) bool {
	if left.Binding != right.Binding || !bytes.Equal(left.ResultJSON, right.ResultJSON) ||
		len(left.Vertices) != len(right.Vertices) || len(left.Edges) != len(right.Edges) {
		return false
	}
	for index := range left.Vertices {
		leftVertex, rightVertex := left.Vertices[index], right.Vertices[index]
		if leftVertex.VertexID != rightVertex.VertexID || leftVertex.VertexKind != rightVertex.VertexKind ||
			leftVertex.SortKey != rightVertex.SortKey || !bytes.Equal(leftVertex.JSON, rightVertex.JSON) {
			return false
		}
	}
	for index := range left.Edges {
		leftEdge, rightEdge := left.Edges[index], right.Edges[index]
		if leftEdge.EdgeID != rightEdge.EdgeID || leftEdge.EdgeKind != rightEdge.EdgeKind ||
			leftEdge.SrcVertexID != rightEdge.SrcVertexID || leftEdge.DstVertexID != rightEdge.DstVertexID ||
			leftEdge.Direction != rightEdge.Direction || leftEdge.SortKey != rightEdge.SortKey ||
			!bytes.Equal(leftEdge.JSON, rightEdge.JSON) {
			return false
		}
	}
	return true
}
