package postgresrestore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const truncateGraphProjectionTablesSQL = `TRUNCATE TABLE
  graph_projection_result_edges,
  graph_projection_result_leases,
  graph_projection_result_vertices,
  graph_projection_results
RESTRICT`

const committedProofTimeout = 30 * time.Second

// DerivedStateReconciler is supplied by application assembly. It lets the
// Recovery participant recreate owner-specific jobs and Reporting leases in
// the same transaction without importing those subsystems into Graph.
type DerivedStateReconciler interface {
	ReconcileGraphProjectionDerivedState(context.Context, pgx.Tx, []graphprojection.ResultBindingV2) (int, int, error)
}

type Writer struct {
	db         postgres.DB
	reconciler DerivedStateReconciler
}

var _ graphprojection.RestorePublisher = (*Writer)(nil)

func New(db postgres.DB) (*Writer, error) { return NewWithReconciler(db, nil) }

func NewWithReconciler(db postgres.DB, reconciler DerivedStateReconciler) (*Writer, error) {
	if db == nil {
		return nil, graphprojection.NewRestoreError(graphprojection.RestoreErrorInvalidRequest)
	}
	return &Writer{db: db, reconciler: reconciler}, nil
}

func (writer *Writer) ReplaceAll(ctx context.Context, plan graphprojection.RestorePublicationPlan) (graphprojection.RestorePublicationProof, error) {
	if writer == nil || writer.db == nil || ctx == nil || ctx.Err() != nil || !validPlan(plan) {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Cause: context.Canceled}
	}
	tx, err := writer.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Cause: err}
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, truncateGraphProjectionTablesSQL); err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	bindings := make([]graphprojection.ResultBindingV2, 0, len(plan.Projections))
	for _, staged := range plan.Projections {
		if err := publisher.PublishResult(ctx, staged.Result); err != nil {
			return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
		}
		bindings = append(bindings, staged.Result.Binding)
	}
	jobs, leases := 0, 0
	if writer.reconciler != nil {
		jobs, leases, err = writer.reconciler.ReconcileGraphProjectionDerivedState(ctx, tx, bindings)
		if err != nil || jobs < 0 || leases < 0 {
			return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
		}
	}
	if err := verifyPublishedState(ctx, tx, plan, leases); err != nil {
		return graphprojection.RestorePublicationProof{}, publicationError(ctx, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Indeterminate: true, Cause: err}
	}
	proofContext, cancelProof := context.WithTimeout(context.WithoutCancel(ctx), committedProofTimeout)
	defer cancelProof()
	if err := verifyPublishedState(proofContext, writer.db, plan, leases); err != nil {
		return graphprojection.RestorePublicationProof{}, &graphprojection.RestorePublicationError{Indeterminate: true, Cause: err}
	}
	return graphprojection.RestorePublicationProof{
		RebuiltViews:                  append([]graphprojection.RestoreRebuiltView{}, plan.RebuiltViews...),
		ReconciledNonterminalJobCount: jobs, ReconciledLeaseCount: leases,
		PostconditionSHA256: plan.PostconditionSHA256,
	}, nil
}

func publicationError(ctx context.Context, cause error) error {
	return &graphprojection.RestorePublicationError{Indeterminate: ctx != nil && ctx.Err() != nil, Cause: cause}
}

func validPlan(plan graphprojection.RestorePublicationPlan) bool {
	if plan.RestoreOperationID.String() == "00000000-0000-0000-0000-000000000000" ||
		plan.TargetGenerationID.String() == "00000000-0000-0000-0000-000000000000" ||
		len(plan.ClearedTableIDs) != len(graphprojection.RestoreGraphTableIDs()) ||
		len(plan.Projections) != len(plan.RebuiltViews) || len(plan.PostconditionSHA256) != 64 {
		return false
	}
	for index, tableID := range graphprojection.RestoreGraphTableIDs() {
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
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func verifyPublishedState(ctx context.Context, queryer restoreQueryer, plan graphprojection.RestorePublicationPlan, wantLeases int) error {
	var results, vertices, edges, leases int
	if err := queryer.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM graph_projection_results),
    (SELECT COUNT(*) FROM graph_projection_result_vertices),
    (SELECT COUNT(*) FROM graph_projection_result_edges),
    (SELECT COUNT(*) FROM graph_projection_result_leases)
`).Scan(&results, &vertices, &edges, &leases); err != nil {
		return fmt.Errorf("verify Graph restore aggregate state: %w", err)
	}
	wantVertices, wantEdges := 0, 0
	for _, rebuilt := range plan.RebuiltViews {
		wantVertices += rebuilt.VertexCount
		wantEdges += rebuilt.EdgeCount
	}
	if results != len(plan.Projections) || vertices != wantVertices || edges != wantEdges || leases != wantLeases {
		return fmt.Errorf("graph restore aggregate postcondition mismatch")
	}
	reader, err := postgresresult.NewReader(queryer)
	if err != nil {
		return err
	}
	for _, staged := range plan.Projections {
		loaded, err := reader.ReadExactResult(ctx, staged.Result.Binding)
		if err != nil || len(loaded.Vertices) != len(staged.Result.Vertices) || len(loaded.Edges) != len(staged.Result.Edges) {
			return fmt.Errorf("verify Graph restore exact result: %w", err)
		}
	}
	return nil
}
