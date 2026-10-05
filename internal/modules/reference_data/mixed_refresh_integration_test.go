package reference_data_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestMixedRefreshCommitsCohortFallbackAndReplay_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	barrier := &appsupport.ReferencePackVerificationBarrier{}
	harness := startReferencePackServerWithEnv(t, runtime, "extension_profile-reference-pack-mixed-refresh", nil, barrier)
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	capture := testsupport.StartCapture()
	defer capture.Close(context.Background())
	importReferencePack(t, harness, admin, "type_registry.host", "1", "mixed-import-1")
	importReferencePack(t, harness, admin, "type_registry.host", "2", "mixed-import-2")
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/1/disable", "mixed-disable", ""), http.StatusOK)
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/2/activate", "mixed-activate", ""), http.StatusOK)

	readEnvelope := func(version string) string {
		t.Helper()
		var id string
		if err := harness.DB.QueryRow(`SELECT current_envelope_id FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	priorEnvelope := readEnvelope("1")
	missingEnvelope := readEnvelope("2")
	var priorSet string
	var priorProvenance []byte
	if err := harness.DB.QueryRow(`SELECT pack_set_id,canonical_provenance FROM reference_pack_current_set JOIN reference_pack_sets USING(pack_set_id) WHERE singleton`).Scan(&priorSet, &priorProvenance); err != nil {
		t.Fatal(err)
	}
	removeStoredBundle(t, harness, "type_registry.host", "2")

	started, release := make(chan struct{}), make(chan struct{})
	released := false
	barrier.BlockNext(started, release)
	defer func() {
		if !released {
			close(release)
		}
	}()
	request := map[string]any{"client_txn_id": "mixed-refresh", "pack_keys": []string{"type_registry.host"}}
	post := func() map[string]any {
		t.Helper()
		response := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", request, csrfOptions(admin)...)
		return requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	}
	jobID := post()["job_id"].(string)
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("refresh did not reach its execution boundary")
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_operation_members m JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID); got != 2 {
		t.Fatalf("frozen cohort includes built-ins or omits a retained version: %d", got)
	}
	if readEnvelope("1") != priorEnvelope || readEnvelope("2") != missingEnvelope {
		t.Fatal("refresh published an envelope before finalization")
	}
	requireReferencePackResource(t, readReferencePack(t, harness, admin, "type_registry.host", "2"), "type_registry.host", "2", reference_data.ConditionVerifiedAvailable, true)
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set WHERE singleton AND pack_set_id=$1`, priorSet); got != 1 {
		t.Fatal("refresh changed the active set before finalization")
	}
	close(release)
	released = true
	done := requireJob(t, harness, admin, jobID)
	if done["status"] != "failed" || done["error_summary"].(map[string]any)["code"] != "reference_pack_verification_failed" {
		t.Fatalf("mixed cohort terminal result: %#v", done)
	}
	if readEnvelope("1") == priorEnvelope || readEnvelope("2") != missingEnvelope {
		t.Fatal("mixed refresh must renew the healthy envelope and retain the missing member's last success")
	}
	healthy := readReferencePack(t, harness, admin, "type_registry.host", "1")
	requireReferencePackResource(t, healthy, "type_registry.host", "1", reference_data.ConditionDisabled, false)
	if healthy["health"] != "verified_available" || healthy["administratively_disabled"] != true {
		t.Fatal("verification changed administrative disablement")
	}
	missing := readReferencePack(t, harness, admin, "type_registry.host", "2")
	requireReferencePackResource(t, missing, "type_registry.host", "2", reference_data.ConditionMissing, false)
	if missing["missing_reason"] != "storage_loss" {
		t.Fatalf("missing member reason: %#v", missing)
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set c JOIN reference_pack_set_members m USING(pack_set_id) JOIN reference_pack_candidates v USING(pack_key,pack_version) WHERE c.singleton AND m.pack_key='type_registry.host' AND v.distribution_kind='packaged_builtin'`); got != 1 {
		t.Fatal("mixed refresh did not publish mandatory Base fallback")
	}
	var baseVersion string
	if err := harness.DB.QueryRow(`SELECT m.pack_version FROM reference_pack_current_set c JOIN reference_pack_set_members m USING(pack_set_id) WHERE c.singleton AND m.pack_key='type_registry.host'`).Scan(&baseVersion); err != nil {
		t.Fatal(err)
	}
	base := readReferencePack(t, harness, admin, "type_registry.host", baseVersion)
	if base["fallback_from_version"] != "2" || missing["fallback_from_version"] != nil {
		t.Fatal("fallback attribution missing from the active Base projection")
	}
	var retainedProvenance []byte
	if err := harness.DB.QueryRow(`SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, priorSet).Scan(&retainedProvenance); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retainedProvenance, priorProvenance) {
		t.Fatal("refresh rewrote historical set provenance")
	}
	for _, verdict := range []string{"succeeded", "content_rejected"} {
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1 AND a.outcome='content_rejected' AND m.verdict=$2`, jobID, verdict); got != 1 {
			t.Fatalf("missing committed %s member evidence: %d", verdict, got)
		}
	}
	if post()["job_id"] != jobID {
		t.Fatal("exact replay returned a new Job")
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events JOIN reference_pack_operations USING(operation_id) WHERE job_id=$1 AND event_kind='refresh_verification'`, jobID); got != 2 {
		t.Fatalf("replay changed the cohort attestation count: %d", got)
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_attempts JOIN reference_pack_operations USING(operation_id) WHERE job_id=$1`, jobID); got != 1 {
		t.Fatalf("replay started another execution: %d", got)
	}
	for action, want := range map[string]int{"refresh_admitted": 1, "refresh_verification": 2, "payload_invalidation": 1, "safety_fallback": 1, "verification_completed": 1} {
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM administrative_audit_projections p JOIN reference_pack_operations o ON p.target_id=o.operation_id::text WHERE o.job_id=$1 AND p.action_code=$2 AND p.actor_kind='user' AND p.actor_user_id=o.actor_user_id AND p.source='api'`, jobID, "reference_pack_"+action); got != want {
			t.Fatalf("committed/replayed audit %s: got %d want %d", action, got, want)
		}
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM deployment_admin_audit_events r JOIN administrative_audit_projections p ON p.audit_event_id=r.id JOIN reference_pack_operations o ON p.target_id=o.operation_id::text WHERE o.job_id=$1 AND (r.after_json::text LIKE '%type_registry.host%' OR p.changes::text LIKE '%type_registry.host%')`, jobID); got != 0 {
		t.Fatal("audit exposed source identifiers")
	}
	refreshSpans := 0
	for _, span := range capture.EndedSpans() {
		if !strings.HasPrefix(span.Name, "reference_pack.") {
			continue
		}
		if len(span.Attributes) != 2 || span.Attributes["cartulary.operation"] != span.Name {
			t.Fatal("unsafe Reference Pack telemetry", span)
		}
		if span.Name == "reference_pack.refresh" {
			refreshSpans++
			if span.Attributes["cartulary.result"] != "rejected" {
				t.Fatal("mixed cohort reported telemetry success", span)
			}
		}
	}
	if refreshSpans != 1 {
		t.Fatal("refresh execution observation missing or replayed", refreshSpans)
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1 AND m.invalidated_content AND m.pack_version='2'`, jobID); got != 1 {
		t.Fatal("first invalidation was not retained with its committed attempt", got)
	}
	response := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.host/2/reverify", map[string]any{"client_txn_id": "repeat-missing-content"}, csrfOptions(admin)...)
	repeatedID := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	if repeated := requireJob(t, harness, admin, repeatedID); repeated["status"] != "failed" {
		t.Fatal("missing retained content unexpectedly recovered", repeated)
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE pack_key='type_registry.host' AND pack_version='2' AND event_kind='payload_invalidation'`); got != 1 {
		t.Fatal("reverification replaced or duplicated first detection evidence", got)
	}
	for _, mutation := range []string{
		`UPDATE reference_pack_attempt_members SET invalidated_content=false WHERE invalidated_content`,
		`UPDATE reference_pack_attempt_members SET failure_code='metadata_expired' WHERE invalidated_content`,
		`UPDATE reference_pack_attempt_members SET invalidated_content=true WHERE verdict='content_rejected' AND NOT invalidated_content`,
		`DELETE FROM reference_pack_attempt_members WHERE invalidated_content`,
	} {
		if _, err := harness.DB.Exec(mutation); err == nil {
			t.Fatal("terminal attempt evidence admitted mutation")
		}
	}
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/1/activate", "replace-fallback", ""), http.StatusOK)
	if resource := readReferencePack(t, harness, admin, "type_registry.host", baseVersion); resource["active"] != false || resource["fallback_from_version"] != nil {
		t.Fatal("inactive Base retained current fallback projection")
	}
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/"+baseVersion+"/activate", "explicit-base", ""), http.StatusOK)
	if resource := readReferencePack(t, harness, admin, "type_registry.host", baseVersion); resource["active"] != true || resource["fallback_from_version"] != nil {
		t.Fatal("explicit Base activation revived obsolete fallback attribution")
	}
}
