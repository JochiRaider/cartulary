package timeline_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

type timelineMutationState struct {
	Effects workbookscenariotest.MutationEffects
	Source  [32]byte
}

func snapshotTimelineMutationState(t testing.TB, db *sql.DB, incidentID, recordID string) timelineMutationState {
	t.Helper()
	var content string
	if err := db.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'source',(SELECT to_jsonb(e) FROM timeline_events e WHERE record_id::text=$2),
 'projection',(SELECT to_jsonb(p) FROM timeline_grid_projection p WHERE record_id::text=$2),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]') FROM record_links l WHERE incident_id::text=$1),
 'mentions',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY to_jsonb(m)::text),'[]') FROM entity_mentions m WHERE source_record_id::text=$2)
 )::text`, incidentID, recordID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return timelineMutationState{Effects: workbookscenariotest.SnapshotIncidentMutationEffects(t, db, incidentID), Source: sha256.Sum256([]byte(content))}
}
