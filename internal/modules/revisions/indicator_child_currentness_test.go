package revisions_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

// The route must reject even well-formed same-version source drift. Each case
// restores the original source before the existing successful inverse control.
func requireIndicatorChildCurrentness(t *testing.T, db *sql.DB, incidentID, addressedID, childID uuid.UUID, table, key string, cases map[string]any, rollback func() *http.Response) {
	t.Helper()
	readSource := func() string {
		var result string
		if err := db.QueryRow(`SELECT to_jsonb(c)::text FROM `+table+` c WHERE `+key+`=$1`, childID).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	original := readSource()
	for field, value := range cases {
		t.Run("current "+field+" drift", func(t *testing.T) {
			set := func(payload []byte) {
				t.Helper()
				if _, err := db.Exec(`UPDATE `+table+` SET `+field+`=(SELECT `+field+` FROM jsonb_populate_record(NULL::`+table+`,$2::jsonb)) WHERE `+key+`=$1`, childID, payload); err != nil {
					t.Fatal(err)
				}
			}
			payload, err := json.Marshal(map[string]any{field: value})
			if err != nil {
				t.Fatal(err)
			}
			set(payload)
			t.Cleanup(func() { set([]byte(original)) })
			current := readSource()
			if current == original {
				t.Fatal("drift fixture did not change the source")
			}
			readEffects := func() string {
				var result string
				if err := db.QueryRow(`SELECT jsonb_build_object(
'records',(SELECT jsonb_agg(to_jsonb(r) ORDER BY record_id) FROM records r WHERE incident_id=$1),
'projection',(SELECT jsonb_agg(to_jsonb(p) ORDER BY record_id) FROM indicator_grid_projection p WHERE incident_id=$1),
'change_sets',(SELECT count(*) FROM change_sets WHERE incident_id=$1),
'mutations',(SELECT count(*) FROM change_set_mutations m JOIN change_sets cs USING(change_set_id) WHERE cs.incident_id=$1),
'revisions',(SELECT count(*) FROM record_revisions r JOIN records USING(record_id) WHERE records.incident_id=$1),
'idempotency',(SELECT count(*) FROM route_idempotency WHERE scope_key=$2))::text`, incidentID, addressedID.String()).Scan(&result); err != nil {
					t.Fatal(err)
				}
				return result
			}
			before := readEffects()
			intents := collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID.String()})
			requireRollbackReasonCode(t, rollback(), "stale_target")
			if readSource() != current || readEffects() != before || collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}) != intents {
				t.Fatal("rejected inverse changed source, envelopes, history, idempotency, projection or publication intents")
			}
		})
	}
}
