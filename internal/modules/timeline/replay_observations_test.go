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
 'hosts',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM hosts x WHERE incident_id::text=$1),
 'identities',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identities x WHERE incident_id::text=$1),
 'host_projections',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM host_grid_projection x WHERE incident_id::text=$1),
 'identity_projections',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identity_grid_projection x WHERE incident_id::text=$1),
 'evidence_projections',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY record_id),'[]') FROM evidence_grid_projection p WHERE incident_id::text=$1),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]') FROM record_links l WHERE incident_id::text=$1),
 'tags',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_tag_id),'[]') FROM record_tags x WHERE incident_id::text=$1),
 'mentions',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY m.entity_mention_id),'[]') FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id WHERE r.incident_id::text=$1)
 )::text`, incidentID, recordID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return timelineMutationState{Effects: workbookscenariotest.SnapshotIncidentMutationEffects(t, db, incidentID), Source: sha256.Sum256([]byte(content))}
}
