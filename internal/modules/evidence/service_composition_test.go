package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	evidencecontract "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func newTestStore(t testing.TB, prefix string) postgres.DB {
	t.Helper()
	harness := pgtest.Start(t)
	if pgtest.ExplicitPostgresFixturePolicyT(t) == pgtest.PostgresFixturePolicyTemplateClone {
		testDB := harness.PrepareIsolatedDatabaseT(t, prefix)
		pool, err := pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatalf("open template-clone postgres pool: %v", err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	return harness.BeginRollbackDBT(t, prefix)
}

func createTestIncident(
	t testing.TB,
	database postgres.DB,
	actor authn.UserRecord,
	clientTxnID string,
	incidentKey string,
	title string,
) incidents.IncidentRecord {
	t.Helper()
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres: database, PreferenceBootstrap: workbookstartuppostgres.NewWriter(), Now: time.Now,
	})
	if err != nil {
		t.Fatalf("compose Incidents test application: %v", err)
	}
	return incidentstoretest.CreateIncidentInStore(t, application, actor, clientTxnID, incidentKey, title).Incident
}

func newTestBlobLifecycleService(t testing.TB, database postgres.DB) *blobLifecycleService {
	t.Helper()
	projectionContribution, err := NewProjectionContribution()
	if err != nil {
		t.Fatalf("compose Evidence projection contribution: %v", err)
	}
	service, err := newBlobLifecycleService(blobLifecycleDependencies{
		Postgres:        database,
		Revisions:       testRevisionAppendPort{},
		Projections:     testEvidenceProjectionRows{source: projectionContribution.Source()},
		SupportEffects:  testEvidenceAssociationEffects{},
		Collaboration:   collaborationsupport.NewRecordChangedAppender(),
		IncidentState:   admission.NewChecker(database),
		RecordEnvelopes: records.NewStore(database),
		Idempotency:     testLifecycleIdempotency{store: authn.NewStore(database)},
	})
	if err != nil {
		t.Fatalf("compose Evidence lifecycle service: %v", err)
	}
	return service
}

func newTestRouteOperations(t testing.TB, database postgres.DB) *routeOperations {
	t.Helper()
	blobs := newTestBlobLifecycleService(t, database)
	access, err := newAccessHandleService(database)
	if err != nil {
		t.Fatalf("compose Evidence access-handle service: %v", err)
	}
	operations, err := newRouteOperations(blobs, access)
	if err != nil {
		t.Fatalf("compose Evidence route operations: %v", err)
	}
	return operations
}

type testEvidenceProjectionRows struct {
	source evidencecontract.SourceReader
}

func (rows testEvidenceProjectionRows) RefreshEvidenceTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	_, _, err := rows.source.LoadProjectionInputTx(ctx, tx, recordID)
	return err
}

func (rows testEvidenceProjectionRows) LoadEvidenceTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	input, found, err := rows.source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrEvidenceNotFound
	}
	return evidencecontract.ViewRow(input), nil
}

type testEvidenceAssociationEffects struct{}

func (testEvidenceAssociationEffects) RefreshEvidenceAssociationEffects(
	context.Context,
	pgx.Tx,
	evidenceprojection.EvidenceAssociationEffectsInput,
) (evidenceprojection.EvidenceAssociationEffectsResult, error) {
	return evidenceprojection.EvidenceAssociationEffectsResult{}, nil
}

type testRevisionAppendPort struct{}

func (testRevisionAppendPort) CaptureRecordSnapshotTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
) (revisions.RecordSnapshot, error) {
	return revisions.RecordSnapshot{}, nil
}

func (testRevisionAppendPort) AppendChangeSetTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendChangeSetParams,
) (uuid.UUID, error) {
	changeSetID := uuid.New()
	if params.ChangeSetID != nil {
		changeSetID = *params.ChangeSetID
	}
	_, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id, incident_id, actor_user_id, source, reason,
    client_txn_id, request_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, changeSetID, params.IncidentID, params.ActorUserID, params.Source, params.Reason,
		params.ClientTxnID, params.RequestID, params.CreatedAt.UTC())
	return changeSetID, err
}

func (testRevisionAppendPort) AppendRecordMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendRecordMutationParams,
) error {
	_, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_version_id, after_version_id, history_record_ids, history_entry_record_ids
)
VALUES ($1, $2, $3, $4, $5, $6, $7, ARRAY[$8::uuid], ARRAY[$8::uuid])
`, params.ChangeSetID, params.SequenceNo, params.TargetKind, params.RecordID.String(),
		params.OperationKind, params.BeforeVersionID, params.AfterVersionID, params.RecordID)
	return err
}

func (testRevisionAppendPort) AppendLiveRevisionTx(
	ctx context.Context,
	tx pgx.Tx,
	input revisions.LiveRevisionInput,
) error {
	_, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    change_set_id, record_id, row_version, before_json, after_json
)
VALUES ($1, $2, $3, NULL, '{}'::jsonb)
`, input.ChangeSetID, input.RecordID, input.RowVersion)
	return err
}

type testLifecycleIdempotency struct {
	store *authn.Store
}

func (adapter testLifecycleIdempotency) Get(
	ctx context.Context,
	key LifecycleIdempotencyKey,
	requestHash []byte,
) (map[string]any, bool, error) {
	record, err := adapter.store.GetRouteIdempotency(ctx, testLifecycleAuthKey(key))
	return decodeTestLifecycleReplay(record, err, requestHash)
}

func (testLifecycleIdempotency) GetTx(
	ctx context.Context,
	tx pgx.Tx,
	key LifecycleIdempotencyKey,
	requestHash []byte,
) (map[string]any, bool, error) {
	record, err := authn.GetRouteIdempotencyTx(ctx, tx, testLifecycleAuthKey(key))
	return decodeTestLifecycleReplay(record, err, requestHash)
}

func (testLifecycleIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key LifecycleIdempotencyKey,
	requestHash []byte,
	payload map[string]any,
) error {
	status := http.StatusOK
	if key.OperationID == LifecycleOperationBlobCreate {
		status = http.StatusCreated
	}
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, testLifecycleAuthKey(key), nil, requestHash, status, payload)
	if authn.IsUniqueViolation(err) {
		return ErrClientTxnConflict
	}
	return err
}

func testLifecycleAuthKey(key LifecycleIdempotencyKey) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}
}

func decodeTestLifecycleReplay(
	record authn.RouteIdempotencyRecord,
	err error,
	requestHash []byte,
) (map[string]any, bool, error) {
	if errors.Is(err, authn.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return nil, false, ErrClientTxnConflict
	}
	var payload map[string]any
	if err := json.Unmarshal(record.ResponseJSON, &payload); err != nil {
		return nil, false, err
	}
	return payload, true, nil
}
