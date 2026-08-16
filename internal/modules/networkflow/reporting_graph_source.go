package networkflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const reportingGraphLeaseOwnerID = "snapshot_reporting"
const reportingGraphLeasePurpose = "render"

// ReportingGraphSource is Network Flow's typed immutable-result contribution.
// It owns declaration interpretation while Graph Projection owns exact result
// and lease mechanics.
type ReportingGraphSource struct {
	db    postgres.DB
	store reportingGraphDeclarationReader
}

type reportingGraphDeclarationReader interface {
	GetGraphViewDeclarationTx(context.Context, pgx.Tx, uuid.UUID, string, bool) (GraphViewDeclaration, error)
}

func NewReportingGraphSource(db postgres.DB, declarations reportingGraphDeclarationReader) (*ReportingGraphSource, error) {
	if db == nil || declarations == nil {
		return nil, errors.New("network flow Reporting graph source requires persistence")
	}
	return &ReportingGraphSource{db: db, store: declarations}, nil
}

func (m *Module) ReportingGraphSource() *ReportingGraphSource {
	if m == nil || m.store == nil {
		return nil
	}
	source, err := NewReportingGraphSource(m.store.pool, m.store)
	if err != nil {
		return nil
	}
	return source
}

func (*ReportingGraphSource) SourceOwnerID() string { return ProfileID }

func (source *ReportingGraphSource) ValidateAndLeaseResultTx(ctx context.Context, tx pgx.Tx, incidentID, jobID uuid.UUID, binding graphprojection.ResultBindingV2, observedAt, leasedUntil time.Time) (graphprojection.ResultLeaseV2, error) {
	if source == nil || source.store == nil || tx == nil || binding.SourceOwnerID != ProfileID || binding.ProjectionSchemaID != graphprojection.ProjectionSchemaIDV2 {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	declaration, err := source.store.GetGraphViewDeclarationTx(ctx, tx, incidentID, binding.GraphViewID, false)
	if errors.Is(err, ErrGraphViewDeclarationNotFound) {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2NotFound
	}
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive || declaration.SelectedResult == nil {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2NotSelected
	}
	selected := graphViewResultBinding(declaration)
	if selected.SourceSnapshotID != binding.SourceSnapshotID {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2SourceStale
	}
	if selected != binding {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	reader, err := postgresresult.NewReader(tx)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	if _, err := reader.ReadExactResult(ctx, binding); err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	writer, err := postgresresult.NewLeaseWriter(tx)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	return writer.AcquireLease(ctx, graphprojection.ResultLeaseV2{
		LeaseID:              uuid.NewSHA1(jobID, []byte(binding.ProjectionResultID+":"+reportingGraphLeasePurpose)).String(),
		ProjectionResultID:   binding.ProjectionResultID,
		LeaseOwnerID:         reportingGraphLeaseOwnerID,
		LeaseOwnerResourceID: jobID.String(),
		LeasePurpose:         reportingGraphLeasePurpose,
		LeasedUntil:          leasedUntil.UTC(),
		CreatedAt:            observedAt.UTC(),
		RenewedAt:            observedAt.UTC(),
	})
}

func (source *ReportingGraphSource) ReadAndRenewLeasedResult(ctx context.Context, jobID uuid.UUID, binding graphprojection.ResultBindingV2, observedAt, leasedUntil time.Time) (graphprojection.CompletedResultV2, error) {
	if source == nil || source.db == nil || jobID == uuid.Nil || binding.SourceOwnerID != ProfileID {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	var leaseID string
	err := source.db.QueryRow(ctx, `
SELECT lease_id
  FROM graph_projection_result_leases
 WHERE projection_result_id = $1
   AND lease_owner_id = $2
   AND lease_owner_resource_id = $3
   AND lease_purpose = $4
   AND leased_until > $5
`, binding.ProjectionResultID, reportingGraphLeaseOwnerID, jobID.String(), reportingGraphLeasePurpose, observedAt.UTC()).Scan(&leaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2LeaseNotFound
	}
	if err != nil {
		return graphprojection.CompletedResultV2{}, fmt.Errorf("read Reporting graph result lease: %w", err)
	}
	writer, err := postgresresult.NewLeaseWriter(source.db)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	if _, err := writer.RenewLease(ctx, leaseID, observedAt.UTC(), leasedUntil.UTC()); err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	reader, err := postgresresult.NewReader(source.db)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	return reader.ReadExactResult(ctx, binding)
}

func (source *ReportingGraphSource) ReleaseJobLeasesTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) error {
	if source == nil || tx == nil || jobID == uuid.Nil {
		return nil
	}
	_, err := tx.Exec(ctx, `
DELETE FROM graph_projection_result_leases lease
 USING graph_projection_results result
 WHERE lease.projection_result_id = result.projection_result_id
   AND result.source_owner_id = $1
   AND lease.lease_owner_id = $2
   AND lease.lease_owner_resource_id = $3
   AND lease.lease_purpose = $4
`, ProfileID, reportingGraphLeaseOwnerID, jobID.String(), reportingGraphLeasePurpose)
	return err
}

func (source *ReportingGraphSource) ReleaseJobLeases(ctx context.Context, jobID uuid.UUID) error {
	if source == nil || source.db == nil || jobID == uuid.Nil {
		return nil
	}
	tx, err := source.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := source.ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
