package mutationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type MutationOwner string

const (
	MutationOwnerIncidentResource   MutationOwner = "incident resource mutation"
	MutationOwnerIncidentMembership MutationOwner = "incident membership mutation"
)

type AuditEventRecord struct {
	EventKind   string
	ActorUserID string
	EventSource string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
	Before      map[string]any
	After       map[string]any
}

type rows interface {
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Database struct {
	query func(context.Context, string, ...any) (rows, error)
	close func(rows)
}

func SQLDatabase(db *sql.DB) Database {
	return Database{
		query: func(ctx context.Context, query string, args ...any) (rows, error) {
			return db.QueryContext(ctx, query, args...)
		},
		close: func(result rows) { _ = result.(*sql.Rows).Close() },
	}
}

func PostgresDatabase(db postgres.DB) Database {
	return Database{
		query: func(ctx context.Context, query string, args ...any) (rows, error) {
			return db.Query(ctx, query, args...)
		},
		close: func(result rows) { result.(pgx.Rows).Close() },
	}
}

func decodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	if len(payload) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode JSON payload: %v", err)
	}
	return decoded
}

type MutationSelector struct {
	IncidentID  string
	ClientTxnID string
}

func LookupOwnerMutations(t testing.TB, db Database, selector MutationSelector, owner MutationOwner) []AuditEventRecord {
	t.Helper()

	events := lookupAuditEventsBySelector(t, db, selector)
	filtered := make([]AuditEventRecord, 0, len(events))
	for _, event := range events {
		if mutationOwnedBy(event, owner) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func CountMutationArtifacts(t testing.TB, db Database, selector MutationSelector, owners ...MutationOwner) int {
	t.Helper()

	if len(owners) == 0 {
		owners = []MutationOwner{
			MutationOwnerIncidentResource,
			MutationOwnerIncidentMembership,
		}
	}

	total := 0
	for _, owner := range owners {
		total += len(LookupOwnerMutations(t, db, selector, owner))
	}
	return total
}

func RequireNoMutationArtifacts(t testing.TB, db Database, selector MutationSelector, owners ...MutationOwner) {
	t.Helper()

	if got := CountMutationArtifacts(t, db, selector, owners...); got != 0 {
		t.Fatalf("expected no surviving incident mutation artifacts for %+v, got %d", selector, got)
	}
}

func RequireOwnerMutationEvent(
	t testing.TB,
	events []AuditEventRecord,
	owner MutationOwner,
	eventKind string,
	actorUserID string,
	targetUserID string,
) AuditEventRecord {
	t.Helper()

	for _, event := range events {
		if event.EventKind != eventKind || event.ActorUserID != actorUserID || !mutationOwnedBy(event, owner) {
			continue
		}
		if targetUserID == "" || event.After["user_id"] == targetUserID || event.Before["user_id"] == targetUserID {
			return event
		}
	}
	t.Fatalf("missing %s event %q for actor %s target %s in %#v", owner, eventKind, actorUserID, targetUserID, events)
	return AuditEventRecord{}
}

func lookupAuditEventsBySelector(t testing.TB, db Database, selector MutationSelector) []AuditEventRecord {
	t.Helper()

	query := `
SELECT event_kind,
       actor_user_id::text,
       event_source,
       COALESCE(client_txn_id, ''),
       COALESCE(request_id, ''),
       created_at,
       before_json,
       after_json
  FROM deployment_admin_audit_events`
	args := make([]any, 0, 2)
	clauses := make([]string, 0, 2)
	if selector.IncidentID != "" {
		args = append(args, selector.IncidentID)
		clauses = append(clauses, fmt.Sprintf("incident_id::text = $%d", len(args)))
	}
	if selector.ClientTxnID != "" {
		args = append(args, selector.ClientTxnID)
		clauses = append(clauses, fmt.Sprintf("client_txn_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query += "\n WHERE " + strings.Join(clauses, " AND ")
	}
	query += "\n ORDER BY created_at ASC"

	rows, err := db.query(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("query incident mutation artifacts: %v", err)
	}
	defer db.close(rows)

	events := make([]AuditEventRecord, 0, 4)
	for rows.Next() {
		var (
			record        AuditEventRecord
			beforePayload []byte
			afterPayload  []byte
		)
		if err := rows.Scan(&record.EventKind, &record.ActorUserID, &record.EventSource, &record.ClientTxnID, &record.RequestID, &record.CreatedAt, &beforePayload, &afterPayload); err != nil {
			t.Fatalf("scan incident mutation artifact: %v", err)
		}
		record.Before = decodeJSONMap(t, beforePayload)
		record.After = decodeJSONMap(t, afterPayload)
		events = append(events, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate incident mutation artifacts: %v", err)
	}
	return events
}

func mutationOwnedBy(event AuditEventRecord, owner MutationOwner) bool {
	for _, eventKind := range ownerEventKinds(owner) {
		if event.EventKind == eventKind {
			return true
		}
	}
	return false
}

func ownerEventKinds(owner MutationOwner) []string {
	switch owner {
	case MutationOwnerIncidentResource:
		return []string{"incident_created", "incident_updated"}
	case MutationOwnerIncidentMembership:
		return []string{
			"incident_membership_created",
			"incident_membership_updated",
			"incident_membership_deleted",
		}
	default:
		return nil
	}
}
