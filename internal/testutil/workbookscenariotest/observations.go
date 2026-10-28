package workbookscenariotest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

// MutationEffects observes stored content, not just row counts. Source-owner
// values and projected rows are asserted by the caller. Hashes keep retained
// payloads out of failure diagnostics; each selection is incident scoped.
type MutationEffects struct {
	History            [32]byte
	PublicationIntents [32]byte
}

func SnapshotIncidentMutationEffects(t testing.TB, db *sql.DB, incidentID string) MutationEffects {
	t.Helper()
	var content string
	// Workbook mutation receipt scopes start with the native incident or record
	// UUID, optionally followed by a view/operation component. Include rejected
	// transaction identities too by selecting the scope, not known change sets.
	if err := db.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'records',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY record_id),'[]') FROM records r WHERE incident_id::text=$1),
 'changes',(SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY change_set_id),'[]') FROM change_sets c WHERE incident_id::text=$1),
 'mutations',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY m.change_set_id,sequence_no),'[]') FROM change_set_mutations m JOIN change_sets c USING(change_set_id) WHERE c.incident_id::text=$1),
 'revisions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM record_revisions v JOIN records r USING(record_id) WHERE r.incident_id::text=$1),
 'receipts',(SELECT coalesce(jsonb_agg(to_jsonb(i) ORDER BY id),'[]') FROM route_idempotency i WHERE split_part(scope_key,':',1)=$1 OR split_part(scope_key,':',1) IN (SELECT record_id::text FROM records WHERE incident_id::text=$1))
 )::text`, incidentID).Scan(&content); err != nil {
		t.Fatalf("snapshot mutation effects: %v", err)
	}
	id, err := uuid.Parse(incidentID)
	if err != nil {
		t.Fatal(err)
	}
	return MutationEffects{History: sha256.Sum256([]byte(content)), PublicationIntents: collaborationsupport.SnapshotIntentContent(t, db, id)}
}

func RequireChangeSetAttribution(t testing.TB, db *sql.DB, changeSetID string, actorUserID string, source string, clientTxnID string) {
	t.Helper()
	var changeSet struct {
		ActorUserID string
		Source      string
		ClientTxnID string
		RequestID   string
		CreatedAt   time.Time
	}
	if err := db.QueryRowContext(context.Background(), `
SELECT actor_user_id::text, source, client_txn_id, request_id, created_at
FROM change_sets
WHERE change_set_id::text = $1
`, changeSetID).Scan(
		&changeSet.ActorUserID,
		&changeSet.Source,
		&changeSet.ClientTxnID,
		&changeSet.RequestID,
		&changeSet.CreatedAt,
	); err != nil {
		t.Fatalf("lookup change set: %v", err)
	}
	auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, actorUserID, source, clientTxnID)
}
