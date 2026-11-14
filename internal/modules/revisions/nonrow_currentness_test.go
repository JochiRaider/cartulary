package revisions_test

import (
	"database/sql"
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

func requireNonRowCurrentness(t testing.TB, db *sql.DB, incidentID, sourceID, targetID uuid.UUID, table, key string, cases map[string]map[string]any, rollback func() *http.Response) {
	t.Helper()
	var original []byte
	if err := db.QueryRow(`SELECT to_jsonb(v) FROM `+table+` v WHERE `+key+`=$1`, targetID).Scan(&original); err != nil {
		t.Fatal(err)
	}
	for name, values := range cases {
		func() {
			t.Logf("reject complete current %s drift: %s", table, name)
			fields := strings.Join(slices.Sorted(maps.Keys(values)), ",")
			set := func(payload []byte) {
				t.Helper()
				if _, err := db.Exec(`UPDATE `+table+` SET (`+fields+`)=(SELECT `+fields+` FROM jsonb_populate_record(NULL::`+table+`,$2::jsonb)) WHERE `+key+`=$1`, targetID, payload); err != nil {
					t.Fatal(err)
				}
			}
			payload, err := json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			set(payload)
			defer set(original)
			var changed bool
			if err := db.QueryRow(`SELECT to_jsonb(v)<>$2::jsonb FROM `+table+` v WHERE `+key+`=$1`, targetID, original).Scan(&changed); err != nil || !changed {
				t.Fatalf("%s drift fixture changed=%v error=%v", name, changed, err)
			}
			read := func() string {
				var result string
				if err := db.QueryRow(`SELECT jsonb_build_object(
'target',(SELECT to_jsonb(v) FROM `+table+` v WHERE `+key+`=$3),
'mentions',(SELECT jsonb_agg(to_jsonb(em) ORDER BY entity_mention_id) FROM entity_mentions em JOIN records r ON r.record_id=em.source_record_id WHERE r.incident_id=$1),
'aliases',(SELECT jsonb_agg(to_jsonb(a) ORDER BY entity_alias_id) FROM entity_aliases a WHERE incident_id=$1),
'preserved',(SELECT jsonb_agg(to_jsonb(p) ORDER BY entity_preserved_identifier_id) FROM entity_preserved_identifiers p WHERE incident_id=$1),
'claims',(SELECT jsonb_agg(to_jsonb(c) ORDER BY entity_type,identifier_type,normalized_value) FROM entity_active_identifier_claims c WHERE incident_id=$1),
'records',(SELECT jsonb_agg(to_jsonb(r) ORDER BY record_id) FROM records r WHERE incident_id=$1),
'links',(SELECT jsonb_agg(to_jsonb(l) ORDER BY record_link_id) FROM record_links l WHERE incident_id=$1),
	'tags',(SELECT jsonb_agg(to_jsonb(t) ORDER BY record_tag_id) FROM record_tags t WHERE incident_id=$1),
'hosts',(SELECT jsonb_agg(to_jsonb(p) ORDER BY record_id) FROM host_grid_projection p WHERE incident_id=$1),
'identities',(SELECT jsonb_agg(to_jsonb(p) ORDER BY record_id) FROM identity_grid_projection p WHERE incident_id=$1),
'timeline',(SELECT jsonb_agg(to_jsonb(p) ORDER BY record_id) FROM timeline_grid_projection p WHERE incident_id=$1),
'change_sets',(SELECT count(*) FROM change_sets WHERE incident_id=$1),
'mutations',(SELECT count(*) FROM change_set_mutations m JOIN change_sets cs USING(change_set_id) WHERE cs.incident_id=$1),
'revisions',(SELECT count(*) FROM record_revisions r JOIN records USING(record_id) WHERE records.incident_id=$1),
'idempotency',(SELECT count(*) FROM route_idempotency WHERE scope_key=$2))::text`, incidentID, sourceID.String(), targetID).Scan(&result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			before := read()
			intents := collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID.String()})
			requireRollbackReasonCode(t, rollback(), "stale_target")
			if read() != before || collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}) != intents {
				t.Fatalf("rejected %s drift changed source, companion, records, history, idempotency, projections or publication", name)
			}
		}()
	}
}
