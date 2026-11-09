package recovery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	referencehttp "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport/fixturehttp"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/jackc/pgx/v5"
)

func TestATTACKProductionPackRestoresPinnedHistory_Integration(t *testing.T) {
	// This selector verifies retained hashes and MITRE notices before import. It
	// neither signs a substitute nor overrides the application's trusted clock.
	f := referencetest.ATTACKRelease(t, "../../..")
	runtime := appsupport.StartRuntime(t)
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrap, f.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	options := appsupport.ServerOptions{Prefix: "attack-production-recovery", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}}
	source := runtime.StartServer(t, options)
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	referencehttp.ImportPackFixture(t, source.Server, admin, f.PortablePackFixture, "attack-recovery-import")
	action := func(s *appsupport.ServerHarness, action, transaction string) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, s.Server.HTTP.URL+"/api/v1/reference-packs/"+f.Key+"/"+f.Version+"/"+action, map[string]any{"client_txn_id": transaction}, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader("X-CSRF-Token", admin.CSRF.Value))
	}
	httptestx.RequireSuccessEnvelope(t, action(source, "activate", "attack-recovery-activate"), http.StatusOK)
	incident := appsupport.CreateIncident(t, source.Server, admin, map[string]any{"client_txn_id": "attack-recovery-incident", "incident_key": "ATTACK-RECOVERY", "title": "ATT&CK retained production recovery"})
	response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, source.Server.HTTP.URL+"/api/v1/snapshots", map[string]any{"incident_id": incident["incident_id"], "client_txn_id": "attack-recovery-pin"}, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader("X-CSRF-Token", admin.CSRF.Value))
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	deadline := time.Now().Add(120 * time.Second)
	for {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, source.Server.HTTP.URL+"/api/v1/jobs/"+job, nil, httptestx.WithCookies(admin.Session))
		state := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if state["status"] == "succeeded" {
			break
		}
		if state["status"] == "failed" || state["status"] == "canceled" || time.Now().After(deadline) {
			t.Fatal("reporting pin failed", state["error_summary"])
		}
		time.Sleep(25 * time.Millisecond)
	}
	httptestx.RequireSuccessEnvelope(t, action(source, "disable", "attack-recovery-disable"), http.StatusOK)
	newConsumer := func(h *appsupport.ServerHarness) reference_data.Consumer {
		t.Helper()
		storage, err := referenceassembly.NewRootStorage(h.Server.Config.Roots.TemporaryWork.Path, h.Server.Config.Roots.ReferencePackStorage.Path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		finalizer, err := extensionassembly.NewReferencePackMutationFinalizer(func(err error) { t.Error(err) })
		if err != nil {
			t.Fatal(err)
		}
		consumer, err := reference_data.NewConsumer(h.Pool, storage, pagination.NewCodec([32]byte{1}), func() time.Time { return time.Now().UTC() }, reference_data.IntegrityOptions{Finalizer: finalizer, Limits: reference_data.DefaultLimits()})
		if err != nil {
			t.Fatal(err)
		}
		return consumer
	}
	request := reference_data.GetPackDescriptorRequest{PackSetID: "rpset_" + f.SetSHA256, PackKey: f.Key}
	before := newConsumer(source).GetPackDescriptor(t.Context(), request)
	if before.Error != nil {
		t.Fatal(before.Error)
	}
	snapshot := func(h *appsupport.ServerHarness) map[string]string {
		t.Helper()
		result := map[string]string{}
		for _, table := range []string{"reference_pack_versions", "reference_pack_sets", "reference_pack_pins", "reference_pack_object_refs", "reference_pack_envelopes"} {
			var value string
			filter := ""
			if table == "reference_pack_object_refs" {
				filter = " WHERE owner_kind IN ('version','envelope','release')"
			}
			if err := h.DB.QueryRowContext(t.Context(), `SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]')::text FROM `+table+` t`+filter).Scan(&value); err != nil {
				t.Fatal(err)
			}
			result[table] = value
		}
		var binding string
		if err := h.DB.QueryRowContext(t.Context(), `SELECT (snapshot_admission_json->'reference_packs')::text FROM reporting_job_payloads WHERE job_id=$1`, job).Scan(&binding); err != nil {
			t.Fatal(err)
		}
		result["reporting_binding"] = binding
		return result
	}
	original := snapshot(source)
	source.Server.Close()
	var probes []appsupport.CurrentRestoreProbe
	for _, defect := range []string{"missing_member", "damaged_member", "wrong_authority"} {
		probes = append(probes, appsupport.CurrentRestoreProbe{Name: defect, Run: func(t testing.TB, params appsupport.CurrentRestoreParams) {
			fault := &attackRestoreFault{ReferencePackStorage: params.Target.ReferencePacks, defect: defect, manifest: f.ManifestSHA256, set: "rpset_" + f.SetSHA256}
			params.Target.ReferencePacks = fault
			_, err := appsupport.RestoreCurrentBackup(t.Context(), params)
			if err == nil || !fault.injected {
				t.Fatal("restore fault not rejected", defect, err)
			}
			if defect == "wrong_authority" && !contractadmission.IsIncompatible(err) {
				t.Fatal("wrong authority lost contract rejection", err)
			}
			for _, table := range []string{"reference_pack_versions", "reference_pack_sets", "reference_pack_pins", "reference_pack_indexes"} {
				var count int
				if err := params.Target.Postgres.QueryRow(t.Context(), `SELECT count(*) FROM `+table).Scan(&count); err != nil || count != 0 {
					t.Fatal("partial restored state committed", defect, table, count, err)
				}
			}
			var state string
			if err := params.Target.Postgres.QueryRow(t.Context(), `SELECT state FROM recovery_target_state WHERE singleton`).Scan(&state); err != nil || state != "restoring" {
				t.Fatal("failed restore escaped quarantine", state, err)
			}
			if defect != "wrong_authority" {
				if stream, err := fault.OpenRecoveryObject(t.Context(), fault.member); err == nil {
					stream.Close()
					t.Fatal("failed member published")
				}
			}
		}})
	}
	restored := runtime.CaptureRestoreServer(t, source, appsupport.ServerOptions{Prefix: "attack-production-restored", TestRouteMode: httptestx.TestRouteModeDisabled, Env: options.Env}, probes...)
	for name, value := range snapshot(restored) {
		if original[name] != value {
			t.Fatal("restored logical content, anchors, inventory or pin changed", name)
		}
	}
	consumer := newConsumer(restored)
	after := consumer.GetPackDescriptor(t.Context(), request)
	if after.Error != nil || !reflect.DeepEqual(before.Value, after.Value) {
		t.Fatal("restored exact descriptor/provenance", after.Error)
	}
	current := consumer.ResolveCurrentPackSet(t.Context())
	if current.Error != nil {
		t.Fatal(current.Error)
	}
	unavailable := consumer.GetPackDescriptor(t.Context(), reference_data.GetPackDescriptorRequest{PackSetID: current.Value.ID, PackKey: f.Key})
	if unavailable.Error == nil || unavailable.Error.Code != "pack_unavailable" {
		t.Fatal("restore reactivated disabled content")
	}
	var objects, relationships int
	if err := restored.DB.QueryRowContext(t.Context(), `SELECT count(*) FILTER(WHERE entry_kind='object'),count(*) FILTER(WHERE entry_kind='relationship') FROM reference_pack_indexes i JOIN reference_pack_candidates c ON c.current_index_id=i.index_id WHERE c.pack_key=$1 AND c.pack_version=$2`, f.Key, f.Version).Scan(&objects, &relationships); err != nil || objects != 6221 || relationships != 34632 {
		t.Fatal("restored corpus", objects, relationships, err)
	}
	var containerRef string
	if err := restored.DB.QueryRowContext(t.Context(), `SELECT e.container_ref FROM reference_pack_candidates c JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE c.pack_key=$1 AND c.pack_version=$2`, f.Key, f.Version).Scan(&containerRef); err != nil {
		t.Fatal(err)
	}
	container, err := os.ReadFile(filepath.Join(restored.Server.Config.Roots.ReferencePackStorage.Path, containerRef))
	if err != nil || !bytes.Equal(container, f.Container) {
		t.Fatal("exact archive, notices and payload changed", err)
	}
	// A fresh authenticated administrator exercises the real removal guard on
	// restored pins. Authentication state is deliberately not reused across restore.
	flowtest.SeedLocalUserRecord(t, restored.DB, "restored-admin@example.test", "Restored fixture admin", "FixturePassword1!", false, true, true)
	login := flowtest.LoginLocalUser(t, http.DefaultClient, restored.Server.HTTP.URL, "restored-admin@example.test", "FixturePassword1!", nil)
	remove := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, restored.Server.HTTP.URL+"/api/v1/reference-packs/"+f.Key+"/"+f.Version+"/remove", map[string]any{"client_txn_id": "attack-restored-remove", "reason": "qualification"}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader("X-CSRF-Token", login.CSRF.Value))
	rejection := httptestx.RequireErrorEnvelope(t, remove, http.StatusConflict, "reference_pack_operation_rejected")
	raw, _ := json.Marshal(rejection)
	if !bytes.Contains(raw, []byte(`"reason_code":"pinned"`)) {
		t.Fatal("restored pin did not block removal")
	}
}

type attackRestoreFault struct {
	recovery.ReferencePackStorage
	defect, manifest, set, member string
	injected                      bool
}

func (f *attackRestoreFault) RestoreMember(ctx context.Context, key, digest string, size int64, reader io.Reader) error {
	if digest == f.manifest && f.defect != "wrong_authority" {
		f.injected = true
		f.member = key
		if f.defect == "missing_member" {
			_, err := io.Copy(io.Discard, reader)
			return err
		}
		return f.ReferencePackStorage.RestoreMember(ctx, key, digest, size, strings.NewReader("damaged retained manifest"))
	}
	return f.ReferencePackStorage.RestoreMember(ctx, key, digest, size, reader)
}
func (f *attackRestoreFault) AdmitRestoredState(ctx context.Context, tx pgx.Tx) error {
	if f.defect == "wrong_authority" {
		var raw []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, f.set).Scan(&raw); err != nil {
			return err
		}
		changed := bytes.Replace(raw, []byte(`"authority_class":"framework_reference"`), []byte(`"authority_class":"advisory_enrichment"`), 1)
		if bytes.Equal(raw, changed) {
			return fmt.Errorf("authority fault has no subject")
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_sets SET canonical_provenance=$2 WHERE pack_set_id=$1`, f.set, changed); err != nil {
			return err
		}
		f.injected = true
	}
	return f.ReferencePackStorage.AdmitRestoredState(ctx, tx)
}
