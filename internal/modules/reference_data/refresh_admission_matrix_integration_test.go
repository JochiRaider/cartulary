package reference_data_test

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestRefreshAdmissionEligibilityMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	barrier := &appsupport.ReferencePackVerificationBarrier{}
	h := startReferencePackServerWithEnv(t, runtime, "reference-refresh-admission-matrix", nil, barrier)
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	refresh := func(txn string, keys []string) *http.Response {
		request := map[string]any{"client_txn_id": txn}
		if keys != nil {
			request["pack_keys"] = keys
		}
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/reference-packs/refresh", request, csrfOptions(admin)...)
	}
	accepted := func(response *http.Response) string {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	}
	type member struct{ Key, Version, Envelope string }
	readMembers := func(job string) []member {
		t.Helper()
		rows, err := h.DB.Query(`SELECT m.ordinal,m.pack_key,m.pack_version,m.envelope_id FROM reference_pack_operation_members m JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1 ORDER BY m.ordinal`, job)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		got := []member{}
		for rows.Next() {
			var ordinal int
			var m member
			if err := rows.Scan(&ordinal, &m.Key, &m.Version, &m.Envelope); err != nil {
				t.Fatal(err)
			}
			if ordinal != len(got)+1 {
				t.Fatal("noncontiguous cohort", ordinal)
			}
			got = append(got, m)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return got
	}
	assertEmpty := func(txn string, keys []string) {
		t.Helper()
		sets := queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_sets`)
		events := queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_events`)
		job := accepted(refresh(txn, keys))
		if done := requireJob(t, h, admin, job); done["status"] != "succeeded" {
			t.Fatal(done)
		}
		if got := readMembers(job); len(got) != 0 {
			t.Fatal("empty selection admitted versions", got)
		}
		if queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_sets`) != sets || queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_events`) != events {
			t.Fatal("empty cohort mutated versions or sets")
		}
	}
	assertEmpty("empty-omitted", nil)
	assertEmpty("empty-explicit", []string{"type_registry.indicator", "enrichment.tor"})
	importHost := func(version string, sequence int) {
		t.Helper()
		bundle := referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackVersion: version, ManifestTransform: func(m map[string]any) { m["pack_release_sequence"] = sequence }})
		job := accepted(postReferencePackUpload(t, h.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"matrix-import-%s"}`, version), bundle, "fixture.zip", reference_data.MediaTypeZip))
		if done := requireJob(t, h, admin, job); done["status"] != "succeeded" {
			t.Fatal(done)
		}
	}
	for i, version := range []string{"2", "10", "failed", "missing", "removed"} {
		importHost(version, i+1)
	}
	importReferencePack(t, h, admin, "type_registry.evidence", "9", "matrix-evidence")
	httptestx.RequireSuccessEnvelope(t, postAction(t, h, admin, "/api/v1/reference-packs/type_registry.host/10/disable", "matrix-disable", ""), http.StatusOK)
	httptestx.RequireSuccessEnvelope(t, postAction(t, h, admin, "/api/v1/reference-packs/type_registry.host/removed/remove", "matrix-remove", "retire fixture"), http.StatusOK)
	importReferencePackWithKind(t, h, admin, "framework.veris", "framework", "1", "matrix-veris")
	httptestx.RequireSuccessEnvelope(t, postAction(t, h, admin, "/api/v1/reference-packs/framework.veris/1/remove", "matrix-remove-only", "retire fixture"), http.StatusOK)
	// Physical corruption/missing bytes enter the real retained reverify path;
	// successful envelopes must survive both resulting health transitions.
	path := storedBundlePath(t, h, "type_registry.host", "failed")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[0] ^= 1
	overwriteStoredBundle(t, h, "type_registry.host", "failed", data)
	requireReverifyFailure(t, h, admin, "type_registry.host", "failed", "matrix-failed", "checksum_mismatch")
	removeStoredBundle(t, h, "type_registry.host", "missing")
	requireReverifyFailure(t, h, admin, "type_registry.host", "missing", "matrix-missing", "payload_missing")
	for _, key := range []string{"type_registry.host", "enrichment.tor"} {
		bundle := referencePackBundle(t, bundleOptions{PackKey: key, PackVersion: "never", BadPayloadSHA: true})
		job := accepted(postReferencePackUpload(t, h.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"matrix-never-%s"}`, key), bundle, "fixture.zip", reference_data.MediaTypeZip))
		if done := requireJob(t, h, admin, job); done["status"] != "failed" {
			t.Fatal(done)
		}
		if n := queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version='never' AND current_envelope_id IS NULL AND health='failed'`, key); n != 1 {
			t.Fatal("never-successful setup", key, n)
		}
	}
	for _, state := range []struct {
		version, health   string
		disabled, removed bool
		reason            string
	}{
		{"2", "verified_available", false, false, ""}, {"10", "verified_available", true, false, ""},
		{"failed", "failed", false, false, ""}, {"missing", "missing", false, false, "storage_loss"}, {"removed", "missing", false, true, "administrative_removal"},
	} {
		var health, reason string
		var disabled, removed bool
		var envelope *string
		if err := h.DB.QueryRow(`SELECT health,administratively_disabled,removed,coalesce(missing_reason,''),current_envelope_id FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version=$1`, state.version).Scan(&health, &disabled, &removed, &reason, &envelope); err != nil {
			t.Fatal(err)
		}
		if health != state.health || disabled != state.disabled || removed != state.removed || reason != state.reason || envelope == nil {
			t.Fatal("incorrect pre-admission state", state, health, reason, disabled, removed, envelope)
		}
	}
	// Hold a real import at preparation: this known key has no eligible version.
	// Pending work must reject explicit refresh before empty-cohort resolution.
	t.Run("pending explicitly selected ineligible key", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		barrier.BlockNext(entered, release)
		defer close(release)
		bundle := referencePackBundle(t, bundleOptions{PackKey: "enrichment.tor", PackVersion: "pending", BadPayloadSHA: true})
		job := accepted(postReferencePackUpload(t, h.Server.HTTP.URL, admin, `{"client_txn_id":"matrix-pending"}`, bundle, "fixture.zip", reference_data.MediaTypeZip))
		select {
		case <-entered:
		case <-time.After(10 * time.Second):
			t.Fatal("pending barrier not reached")
		}
		if n := queryCount(t, h.DB, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key='enrichment.tor' AND current_envelope_id IS NOT NULL`); n != 0 {
			t.Fatal("pending key became eligible", n)
		}
		tables := []string{"jobs", "route_idempotency", "reference_pack_operations", "reference_pack_operation_keys", "reference_pack_operation_members", "reference_pack_operation_repositories"}
		before := make([]int, len(tables))
		for i, table := range tables {
			before[i] = queryCount(t, h.DB, `SELECT count(*) FROM `+table)
		}
		for i, keys := range [][]string{{"enrichment.tor"}, {"type_registry.host", "enrichment.tor"}} {
			requireReasonError(t, refresh(fmt.Sprintf("matrix-pending-refresh-%d", i), keys), http.StatusConflict, "reference_pack_operation_rejected", "verification_pending")
		}
		for i, table := range tables {
			if after := queryCount(t, h.DB, `SELECT count(*) FROM `+table); after != before[i] {
				t.Fatal("rejected admission left partial writes", table, before[i], after)
			}
		}
		httptestx.RequireSuccessEnvelope(t, cancelJob(t, h, admin, job, "matrix-cancel-pending"), http.StatusOK)
		if done := requireJob(t, h, admin, job); done["status"] != "canceled" {
			t.Fatal(done)
		}
	})
	assertEmpty("ineligible-explicit", []string{"enrichment.tor", "framework.veris", "type_registry.indicator"})
	want := []member{{Key: "type_registry.evidence", Version: "9"}, {Key: "type_registry.host", Version: "10"}, {Key: "type_registry.host", Version: "2"}, {Key: "type_registry.host", Version: "failed"}, {Key: "type_registry.host", Version: "missing"}}
	for i := range want {
		if err := h.DB.QueryRow(`SELECT current_envelope_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, want[i].Key, want[i].Version).Scan(&want[i].Envelope); err != nil {
			t.Fatal(err)
		}
	}
	admitted := []string{}
	for _, mode := range []string{"explicit", "omitted"} {
		t.Run(mode, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			barrier.BlockNext(entered, release)
			defer close(release)
			var keys []string
			wantKeys := []string{"type_registry.evidence", "type_registry.host"}
			if mode == "explicit" {
				keys = []string{"type_registry.host", "framework.veris", "type_registry.indicator", "enrichment.tor", "type_registry.evidence", "type_registry.host"}
				wantKeys = []string{"enrichment.tor", "framework.veris", "type_registry.evidence", "type_registry.host", "type_registry.indicator"}
			}
			job := accepted(refresh("matrix-"+mode, keys))
			admitted = append(admitted, job)
			select {
			case <-entered:
			case <-time.After(10 * time.Second):
				t.Fatal("refresh barrier not reached")
			}
			if got := readMembers(job); !reflect.DeepEqual(got, want) {
				t.Fatalf("ordered cohort = %#v; want %#v", got, want)
			}
			rows, err := h.DB.Query(`SELECT k.pack_key FROM reference_pack_operation_keys k JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1 ORDER BY k.pack_key COLLATE "C"`, job)
			if err != nil {
				t.Fatal(err)
			}
			gotKeys := []string{}
			for rows.Next() {
				var key string
				if err := rows.Scan(&key); err != nil {
					t.Fatal(err)
				}
				gotKeys = append(gotKeys, key)
			}
			err = rows.Err()
			rows.Close()
			if err != nil || !reflect.DeepEqual(gotKeys, wantKeys) {
				t.Fatal("selector resolution", gotKeys, wantKeys, err)
			}
			httptestx.RequireSuccessEnvelope(t, cancelJob(t, h, admin, job, "matrix-cancel-"+mode), http.StatusOK)
			if done := requireJob(t, h, admin, job); done["status"] != "canceled" {
				t.Fatal(done)
			}
		})
	}
	importReferencePack(t, h, admin, "type_registry.indicator", "1", "matrix-later-key")
	importHost("later", 6)
	for _, job := range admitted {
		if got := readMembers(job); !reflect.DeepEqual(got, want) {
			t.Fatal("later lifecycle mutation changed frozen membership", got, want)
		}
	}
}
