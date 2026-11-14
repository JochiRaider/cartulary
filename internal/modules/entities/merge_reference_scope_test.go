package entities_test

import (
	"context"
	"crypto/sha256"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
	"github.com/google/uuid"
)

func exerciseMergeReferenceScope(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "merge-reference-scope", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incidentText := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "merge-reference-scope", "incident_key": "IR-MERGE-SCOPE", "title": "Merge dependency scope"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	for _, kind := range []string{"host", "identity"} {
		for _, mode := range []string{"link only", "Mention only"} {
			t.Run(kind+" "+mode, func(t *testing.T) {
				source, survivor, loser, unrelated, mention, link := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
				field, linkType := "timeline.host_refs", "observed_on_host"
				if kind == "identity" {
					field, linkType = "timeline.identity_refs", "observed_as_identity"
				}
				for _, id := range []uuid.UUID{survivor, loser} {
					if kind == "host" {
						entitytest.SeedHostRecord(t, h.DB, incident, actor, id, "Scoped host", id.String(), "", "")
					} else {
						entitytest.SeedIdentityRecord(t, h.DB, incident, actor, id, "Scoped identity", id.String()+"@example.test", id.String()+"@example.test", id.String())
					}
				}
				timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
				timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, unrelated)
				// A second hop is outside the direct dependency set, even though its
				// envelope remains locked throughout the accepted Merge below.
				linktest.SeedRecordLink(t, h.DB, incident, actor, uuid.New(), source, unrelated, "references_record", "manual", nil)
				// Isolate each authoritative dependency contributor. These are source-owner
				// admission fixtures, not complete portable incident histories.
				if mode == "Mention only" {
					entitytest.SeedResolvedMention(t, h.DB, actor, mention, source, loser, field, kind, "Captured target")
					setMentionHistoryCreationFixture(t, h.DB, mention)
				} else {
					linktest.SeedRecordLink(t, h.DB, incident, actor, link, source, loser, linkType, "manual", nil)
				}
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				unrelatedHolder, err := h.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer unrelatedHolder.Rollback(context.Background())
				if _, err := unrelatedHolder.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE`, unrelated); err != nil {
					t.Fatal(err)
				}
				holder, err := h.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer holder.Rollback(context.Background())
				if _, err := holder.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE`, source); err != nil {
					t.Fatal(err)
				}
				body := map[string]any{"loser_record_id": loser.String(), "survivor_base_row_version": 1, "loser_base_row_version": 1, "client_txn_id": "merge-scope-" + source.String()}
				client := *http.DefaultClient
				client.Timeout = 5 * time.Second
				send := func() *http.Response {
					return httptestx.DoJSON(t, &client, http.MethodPost, h.Server.HTTP.URL+"/api/v1/records/"+survivor.String()+"/merge", body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
				}
				snapshot := func() struct {
					source  [32]byte
					effects workbookscenariotest.MutationEffects
				} {
					var content string
					if err := h.DB.QueryRowContext(ctx, `SELECT jsonb_build_object(
 'hosts',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM hosts x WHERE incident_id=$1),
 'identities',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identities x WHERE incident_id=$1),
 'mentions',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY entity_mention_id),'[]') FROM entity_mentions x JOIN records r ON r.record_id=x.source_record_id WHERE r.incident_id=$1),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_link_id),'[]') FROM record_links x WHERE incident_id=$1),
 'timeline',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM timeline_events x WHERE incident_id=$1),
 'timeline_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM timeline_grid_projection x WHERE incident_id=$1),
 'host_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM host_grid_projection x WHERE incident_id=$1),
 'identity_projection',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY record_id),'[]') FROM identity_grid_projection x WHERE incident_id=$1),
 'claims',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY to_jsonb(x)::text),'[]') FROM entity_active_identifier_claims x WHERE incident_id=$1),
 'aliases',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY entity_alias_id),'[]') FROM entity_aliases x WHERE incident_id=$1),
 'identifiers',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY entity_preserved_identifier_id),'[]') FROM entity_preserved_identifiers x WHERE incident_id=$1))::text`, incident).Scan(&content); err != nil {
						t.Fatal(err)
					}
					return struct {
						source  [32]byte
						effects workbookscenariotest.MutationEffects
					}{sha256.Sum256([]byte(content)), workbookscenariotest.SnapshotIncidentMutationEffects(t, h.DB, incidentText)}
				}
				before := snapshot()
				for attempt := 0; attempt < 2; attempt++ {
					result := httptestx.RequireErrorEnvelope(t, send(), http.StatusConflict, "record_locked")["error"].(map[string]any)
					if result["retryable"] != true || !reflect.DeepEqual(result["details"], map[string]any{"record_id": source.String()}) {
						t.Fatalf("dependency lock error=%#v", result)
					}
					if after := snapshot(); after != before {
						t.Fatal("dependency rejection changed retained state")
					}
				}
				if err := holder.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				// Keep the unrelated holder alive: this same previously rejected request
				// must now commit without an incident-wide or transitive lock expansion.
				data := httptestx.RequireSuccessData(t, send(), http.StatusOK)
				var coherent bool
				if err := h.DB.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM records WHERE record_id=ANY($1::uuid[]) AND row_version=2)=2
 AND (SELECT count(*) FROM record_revisions WHERE record_id=ANY($1::uuid[]))=2
 AND (SELECT count(*) FROM records WHERE record_id=ANY($2::uuid[]) AND row_version=1)=2
 AND (SELECT count(*) FROM timeline_events WHERE record_id=ANY($2::uuid[]) AND row_version=1 AND raw_activity_text='record-support-source-row')=2
 AND NOT EXISTS(SELECT 1 FROM record_revisions WHERE record_id=ANY($2::uuid[]))
 AND (SELECT row_version=1 FROM timeline_grid_projection WHERE record_id=$3)`, []uuid.UUID{survivor, loser}, []uuid.UUID{source, unrelated}, source).Scan(&coherent); err != nil {
					t.Fatal(err)
				}
				if !coherent {
					t.Fatal("merge advanced a read-model dependent or lost source/projection/history state")
				}
				rows := workbookscenariotest.QueryViewRows(t, h.Server.HTTP.URL, incidentText, "cartulary.view.timeline.v2", login)
				row := workbookscenariotest.FindRow(t, rows, source.String())
				items := workbookscenariotest.CollectionItems(t, row, field)
				if mode == "Mention only" {
					if len(items) != 1 || items[0]["item_kind"] != "resolved_ref" || items[0]["raw_text"] != "Captured target" || items[0]["resolved_record_id"] != survivor.String() {
						t.Fatalf("dependent projection=%#v", items)
					}
					current := entitytest.LookupMention(t, h.DB, mention)
					if current.ResolvedRecordID == nil || *current.ResolvedRecordID != survivor || current.RowVersion != 2 || current.RawText != "Captured target" {
						t.Fatalf("repointed Mention=%#v", current)
					}
				} else {
					if len(items) != 0 {
						t.Fatalf("link refresh invented Mentions: %#v", items)
					}
					var exact bool
					if err := h.DB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2 AND link_type=$3 AND deleted_at IS NULL)=1 AND (SELECT deleted_at IS NOT NULL FROM record_links WHERE record_link_id=$4)`, source, survivor, linkType, link).Scan(&exact); err != nil {
						t.Fatal(err)
					}
					if !exact {
						t.Fatal("merge did not preserve exact repoint/tombstone effects")
					}
				}
				after := snapshot()
				replay := httptestx.RequireSuccessData(t, send(), http.StatusOK)
				if !reflect.DeepEqual(data, replay) || snapshot() != after {
					t.Fatal("merge replay changed payload or retained effects")
				}
			})
		}
	}
}
