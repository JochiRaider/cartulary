package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
	"testing"
	"time"
)

func assertMaterializationPayloadAdmission(t *testing.T) {
	incident := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	valid := `{"schema_id":"cartulary.network_flow.graph_view_materialization_payload.v1","incident_id":"` + incident.String() + `","graph_view_id":"nfgv_` + strings.Repeat("a", 32) + `","materialization_generation":1,"source_snapshot_id":"nfsnap_` + strings.Repeat("b", 64) + `"}`
	if payload, err := decodeGraphViewMaterializationPayload([]byte(valid), incident); err != nil || !payload.valid() {
		t.Fatalf("valid payload: %v", err)
	}
	invalid := []string{`null`, `[]`, `{`, valid + ` {}`, strings.Replace(valid, `"schema_id":`, `"unknown":true,"schema_id":`, 1), strings.Replace(valid, `"materialization_generation":1`, `"materialization_generation":1,"materialization_generation":2`, 1), strings.Replace(valid, `"materialization_generation":1`, `"materialization_generation":0`, 1), strings.Replace(valid, `"materialization_generation":1`, `"materialization_generation":1.5`, 1), strings.Replace(valid, `"materialization_generation":1`, `"materialization_generation":null`, 1), strings.Replace(valid, `"materialization_generation":1,`, "", 1), strings.Replace(valid, "nfsnap_", "invalid_", 1), strings.Replace(valid, incident.String(), "22222222-2222-4222-8222-222222222222", 1), strings.Replace(valid, "nfgv_", "invalid_", 1)}
	assertStartupPayloadCorpus(t, incident, valid, invalid)
	for _, raw := range invalid {
		tx := &payloadRestoreTx{rows: &payloadRestoreRows{incident: incident, raw: json.RawMessage(raw)}}
		if count, err := ReconcileGraphRestoreJobsTx(context.Background(), tx); err == nil || count != 0 {
			t.Fatalf("restore accepted malformed payload: %q, count=%d err=%v", raw, count, err)
		}
		payload, err := decodeGraphViewMaterializationPayload([]byte(raw), incident)
		if err == nil || payload != (graphViewMaterializationPayload{}) {
			t.Fatalf("rejected payload returned identity: %#v %v", payload, err)
		}
		finalizer := &deadlineGraphViewFinalizer{}
		module := &Module{store: &store{}, limits: EffectiveLimits{GraphMaterializationTimeoutSeconds: 10}, now: time.Now, graphProjection: newGraphProjectionAdapter(), jobManager: &payloadGraphManager{incident: incident, raw: json.RawMessage(raw)}, jobFinalizer: finalizer}
		if err := module.handleGraphViewMaterialization(context.Background(), jobs.Execution{}); err != nil {
			t.Fatal(err)
		}
		if !finalizer.failureCalled || finalizer.successCalled || finalizer.failure.Completion.ErrorSummary.Details["reason_code"] != "source_invalid" {
			t.Fatal("worker did not safely reject payload")
		}
		if err := finalizer.failure.Mutate(context.Background(), nil); err != nil {
			t.Fatal("rejected payload attempted declaration mutation", err)
		}
	}
}

type payloadGraphManager struct {
	deadlineGraphViewJobManager
	incident uuid.UUID
	raw      json.RawMessage
}

func (m payloadGraphManager) ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error) {
	return jobs.Resource{Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &m.incident}}, nil
}
func (m payloadGraphManager) HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error) {
	return m.raw, nil
}

// Exercise byte-level admission before PostgreSQL JSONB normalization can erase
// duplicate members. Real service tests separately prove transactional rollback.
type payloadRestoreTx struct {
	pgx.Tx
	rows *payloadRestoreRows
}

func (tx *payloadRestoreTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return tx.rows, nil
}

type payloadRestoreRows struct {
	pgx.Rows
	incident uuid.UUID
	raw      json.RawMessage
	read     bool
}

func (r *payloadRestoreRows) Next() bool {
	if r.read {
		return false
	}
	r.read = true
	return true
}
func (r *payloadRestoreRows) Close()     {}
func (r *payloadRestoreRows) Err() error { return nil }
func (r *payloadRestoreRows) Scan(dest ...any) error {
	*dest[0].(*uuid.UUID) = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	*dest[1].(**uuid.UUID) = &r.incident
	*dest[2].(*json.RawMessage) = r.raw
	return nil
}

type payloadAdmissionReader struct {
	pgx.Tx
	reads int
}

func (r *payloadAdmissionReader) QueryRow(context.Context, string, ...any) pgx.Row {
	r.reads++
	return payloadMissingProof{}
}

type payloadMissingProof struct{}

func (payloadMissingProof) Scan(...any) error { return pgx.ErrNoRows }

func assertStartupPayloadCorpus(t *testing.T, incident uuid.UUID, valid string, invalid []string) {
	t.Helper()
	jobID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	actor := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	key := authn.RouteIdempotencyKey{RouteKey: routeKeyGraphViewsCreate, ScopeKey: incident.String() + ":/graph-views", ActorUserID: actor, ClientTxnID: "payload-corpus"}
	identity, err := json.Marshal(map[string]any{"schema_id": "cartulary.route_scoped_idempotency_identity.v1", "actor_user_id": actor.String(), "route_identity": key.RouteKey + ":" + key.ScopeKey, "client_txn_id": key.ClientTxnID, "scope_kind": jobs.ScopeKindIncident, "scope_id": incident.String()})
	if err != nil {
		t.Fatal(err)
	}
	definition := jobs.Definition{JobKind: GraphViewMaterializationJobKind, HandlerName: graphViewWorkerKind, ProgressUnitID: "rows", Extension: &jobs.ExtensionPolicy{OwnerProfileID: ProfileID}}
	hash := sha256.Sum256([]byte("payload-corpus"))
	payload, err := decodeGraphViewMaterializationPayload([]byte(valid), incident)
	if err != nil {
		t.Fatal(err)
	}
	graph := graphViewDeclaration{IncidentID: incident, GraphViewID: payload.GraphViewID, MaterializationGeneration: payload.MaterializationGeneration, DesiredSourceSnapshotID: payload.SourceSnapshotID, LatestJobID: &jobID}
	retained := jobs.RetainedExtensionJob{JobKind: definition.JobKind, HandlerName: definition.HandlerName, ProgressUnitID: definition.ProgressUnitID, OwnerProfileID: ProfileID, RouteKey: key.RouteKey, ScopeKey: key.ScopeKey, RequestSHA256: hex.EncodeToString(hash[:]), IdempotencyIdentity: identity, Resource: jobs.Resource{JobID: jobID.String(), Status: jobs.StatusQueued, AuthPolicy: jobs.AuthPolicyIncidentMembership, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incident}, SubmittedByUserID: actor.String()}}
	reader := &payloadAdmissionReader{}
	retained.Payload = json.RawMessage(valid)
	if err := validateSavedGraphJobFacts(context.Background(), reader, definition, retained, key, hash[:], graph); err != nil || reader.reads != 1 {
		t.Fatalf("startup valid control: reads=%d err=%v", reader.reads, err)
	}
	for _, raw := range invalid {
		reader := &payloadAdmissionReader{}
		retained.Payload = json.RawMessage(raw)
		if err := validateSavedGraphJobFacts(context.Background(), reader, definition, retained, key, hash[:], graph); err == nil || reader.reads != 0 {
			t.Fatalf("startup accepted malformed identity or read proof: %q reads=%d err=%v", raw, reader.reads, err)
		}
	}
}
