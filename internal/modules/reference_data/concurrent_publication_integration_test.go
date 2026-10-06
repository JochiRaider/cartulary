package reference_data_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"
)

func TestConcurrentRootPublicationRejectsFrozenOtherKey_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	barrier := &appsupport.ReferencePackVerificationBarrier{}
	harness := startReferencePackServerWithEnv(t, runtime, "reference-pack-shared-root-race", nil, barrier)
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	started, release := make(chan struct{}), make(chan struct{})
	released := false
	barrier.BlockNext(started, release)
	defer func() {
		if !released {
			close(release)
		}
	}()
	post := func(key, request string, root int) map[string]any {
		t.Helper()
		body := referencePackBundle(t, bundleOptions{PackKey: key, PackKind: "type_registry", PackVersion: "6", RootVersion: root})
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"`+request+`"}`, body, "pack.zip", reference_data.MediaTypeZip)
		return requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	}
	first := post("type_registry.host", "frozen-root-one", 0)["job_id"].(string)
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("first import did not reach the verification boundary")
	}
	var active string
	if err := harness.DB.QueryRow(`SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&active); err != nil {
		t.Fatal(err)
	}
	second := post("type_registry.evidence", "advance-shared-root", 2)["job_id"].(string)
	if terminal := requireJob(t, harness, admin, second); terminal["status"] != "succeeded" {
		t.Fatal("other key could not advance shared root", terminal)
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_repositories WHERE repository_id='integration.repo' AND root_version=2`) != 1 {
		t.Fatal("concurrent operation did not advance trust")
	}
	close(release)
	released = true
	terminal := requireJob(t, harness, admin, first)
	if terminal["status"] != "failed" {
		t.Fatal("stale import did not fail", terminal)
	}
	failure := terminal["error_summary"].(map[string]any)
	if failure["code"] != "reference_pack_operation_rejected" || failure["details"].(map[string]any)["reason_code"] != "stale_admission_state" {
		t.Fatal("stale trust was rebased or condemned content", failure)
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_versions WHERE pack_key='type_registry.host' AND pack_version='6'`) != 0 ||
		queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='6' AND health='failed' AND last_failure_code IS NULL AND current_envelope_id IS NULL`) != 1 ||
		queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1 AND m.verdict='succeeded' AND a.outcome='stale_state'`, first) != 1 {
		t.Fatal("stale publication changed successful state or skipped frozen-input verification")
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set WHERE singleton AND pack_set_id=$1`, active) != 1 {
		t.Fatal("concurrent inactive imports changed activation")
	}
	if replay := post("type_registry.host", "frozen-root-one", 0); replay["job_id"] != first {
		t.Fatal("stale-result replay admitted new work")
	}
	third := post("type_registry.host", "explicit-new-root-two", 0)["job_id"].(string)
	if terminal := requireJob(t, harness, admin, third); terminal["status"] != "succeeded" {
		t.Fatal("explicit new import did not use the retained root", terminal)
	}
	t.Run("simultaneous activation preserves the winning set", func(t *testing.T) {
		ctx := context.Background()
		guard, err := harness.DB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer guard.Rollback()
		if _, err := guard.ExecContext(ctx, `SELECT 1 FROM reference_pack_current_set WHERE singleton FOR UPDATE`); err != nil {
			t.Fatal(err)
		}
		type result struct {
			key      string
			response *http.Response
		}
		results := make(chan result, 2)
		for _, key := range []string{"type_registry.host", "type_registry.evidence"} {
			go func() {
				results <- result{key, postAction(t, harness, admin, "/api/v1/reference-packs/"+key+"/6/activate", "simultaneous-"+key, "")}
			}()
		}
		waitReferenceMutationLocks(t, harness, 2)
		if err := guard.Commit(); err != nil {
			t.Fatal(err)
		}
		winner, loser := "", ""
		for range 2 {
			r := <-results
			if r.response.StatusCode == http.StatusOK {
				requireSuccessEnvelope(t, r.response, http.StatusOK)
				if winner != "" {
					t.Fatal("simultaneous admission silently rebased the set")
				}
				winner = r.key
			} else {
				requireReasonError(t, r.response, http.StatusConflict, "reference_pack_operation_rejected", "stale_admission_state")
				loser = r.key
			}
		}
		if winner == "" || loser == "" || queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set c JOIN reference_pack_set_members m USING(pack_set_id) WHERE c.singleton AND m.pack_version='6'`) != 1 {
			t.Fatal("activation race lost or combined mutations")
		}
		requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+loser+"/6/activate", "explicit-after-stale", ""), http.StatusOK)
		if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set c JOIN reference_pack_set_members m USING(pack_set_id) WHERE c.singleton AND m.pack_version='6'`) != 2 {
			t.Fatal("explicit activation overwrote the earlier winner")
		}
	})
	t.Run("removal waits for historical pin acquisition", func(t *testing.T) {
		ctx := context.Background()
		var historical string
		if err := harness.DB.QueryRow(`SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&historical); err != nil {
			t.Fatal(err)
		}
		requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/6/disable", "pin-race-disable", ""), http.StatusOK)
		storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
		if err != nil {
			t.Fatal(err)
		}
		defer storage.Close()
		retention, err := reference_data.NewRetention(harness.Pool, storage, time.Now, consumerIntegrityOptions(t))
		if err != nil {
			t.Fatal(err)
		}
		pin, err := harness.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer pin.Rollback(ctx)
		if _, err := retention.CaptureRetainedTx(ctx, pin, historical, "portable_snapshot", uuid.NewString(), uuid.New()); err != nil {
			t.Fatal(err)
		}
		result := make(chan *http.Response, 1)
		go func() {
			result <- postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/6/remove", "pin-race-remove", "fixture removal")
		}()
		waitReferenceMutationLocks(t, harness, 1)
		if err := pin.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		requireReasonError(t, <-result, http.StatusConflict, "reference_pack_operation_rejected", "pinned")
		retained := readReferencePack(t, harness, admin, "type_registry.host", "6")
		if retained["removed"] != false || retained["administratively_disabled"] != true || retained["reproducibility_pinned"] != true {
			t.Fatal("removal crossed committed pin boundary", retained)
		}
	})
}

func waitReferenceMutationLocks(t testing.TB, harness *appsupport.ServerHarness, count int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var blocked int
		if err := harness.DB.QueryRow(`SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND cardinality(pg_blocking_pids(pid))>0 AND (query LIKE '%reference_pack_repositories%FOR UPDATE%' OR query LIKE '%reference_pack_key_state%FOR UPDATE%' OR query LIKE '%reference_pack_current_set%FOR UPDATE%')`).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked >= count {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("mutations did not reach the guarded publication boundary")
}
