package incidentbundles_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestNetworkFlowRetainedStateBlocksBundleExport_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-network-flow-block")
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-network-flow-block",
		"incident_key":  "BUNDLE-NF-BLOCK",
		"title":         "Incident bundle Network Flow block",
	})
	incidentID := incident["incident_id"].(string)
	tableID := seedIncidentBundleNetworkFlowTable(t, harness.DB, incidentID, adminID)

	assertBlocked := func(clientTxnID string) {
		job := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, admin, map[string]any{
			"incident_id": incidentID, "client_txn_id": clientTxnID,
		}), http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, harness.Server, admin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "extension_state_not_portable")
		details := terminal["error_summary"].(map[string]any)["details"].(map[string]any)
		if len(details) != 2 || details["profile_id"] != "network_flow_activity" {
			t.Fatalf("blocked export details mismatch: %#v", details)
		}
		if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, job["job_id"].(string)) != 0 {
			t.Fatal("blocked Network Flow export published a bundle descriptor")
		}
	}
	assertBlocked("txn-export-network-flow-active")
	if _, err := harness.DB.Exec(`
UPDATE network_flow_tables
   SET table_status = 'soft_deleted',
       deleted_at = now(),
       updated_at = now()
 WHERE network_flow_table_id = $1
`, tableID); err != nil {
		t.Fatalf("soft delete Network Flow table: %v", err)
	}
	assertBlocked("txn-export-network-flow-soft-deleted")

	t.Run("state commit serializes before the publication guard", func(t *testing.T) {
		raceIncident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
			"client_txn_id": "txn-incident-bundle-network-flow-publication-race",
			"incident_key":  "BUNDLE-NF-RACE",
			"title":         "Incident bundle Network Flow publication race",
		})
		raceIncidentID := raceIncident["incident_id"].(string)
		stateTx, err := harness.Pool.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin publication race transaction: %v", err)
		}
		defer func() { _ = stateTx.Rollback(context.Background()) }()
		lock := incidents.NewTransactionParticipant()
		found, err := lock.LockIncidentTx(context.Background(), stateTx, uuid.MustParse(raceIncidentID))
		if err != nil || !found {
			t.Fatalf("lock publication race incident: found=%t err=%v", found, err)
		}
		seedIncidentBundleNetworkFlowTablePGX(t, stateTx, raceIncidentID, adminID)
		type publicationResult struct {
			present bool
			err     error
		}
		publication := make(chan publicationResult, 1)
		go func() {
			publicationTx, beginErr := harness.Pool.BeginTx(context.Background(), pgx.TxOptions{})
			if beginErr != nil {
				publication <- publicationResult{err: beginErr}
				return
			}
			defer func() { _ = publicationTx.Rollback(context.Background()) }()
			locked, lockErr := lock.LockIncidentTx(context.Background(), publicationTx, uuid.MustParse(raceIncidentID))
			if lockErr != nil || !locked {
				publication <- publicationResult{err: fmt.Errorf("publication lock: found=%t: %w", locked, lockErr)}
				return
			}
			present, presenceErr := networkflow.NewPortabilityStateBinding().RetainedAuthoritativeStatePresentTx(
				context.Background(), publicationTx, uuid.MustParse(raceIncidentID), []string{
					networkflow.ExtensionFamilyGraphViews,
					networkflow.ExtensionFamilyIndicatorBindings,
					networkflow.ExtensionFamilyRejectedRowDiagnostics,
					networkflow.ExtensionFamilyRows,
					networkflow.ExtensionFamilyTables,
				},
			)
			publication <- publicationResult{present: present, err: presenceErr}
		}()
		select {
		case result := <-publication:
			t.Fatalf("publication crossed the incident boundary before state commit: %#v", result)
		case <-time.After(100 * time.Millisecond):
		}
		if err := stateTx.Commit(context.Background()); err != nil {
			t.Fatalf("commit publication race state: %v", err)
		}
		select {
		case result := <-publication:
			if result.err != nil || !result.present {
				t.Fatalf("publication guard did not observe committed blocking state: %#v", result)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("publication did not cross the incident boundary after state commit")
		}
	})
}

type incidentBundleSQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func seedIncidentBundleNetworkFlowTable(t testing.TB, db incidentBundleSQLExecutor, incidentID, actorID string) string {
	return seedIncidentBundleNetworkFlowTableWithExec(t, incidentID, actorID, func(query string, args ...any) error {
		_, err := db.ExecContext(context.Background(), query, args...)
		return err
	})
}

func seedIncidentBundleNetworkFlowTablePGX(t testing.TB, tx pgx.Tx, incidentID, actorID string) string {
	return seedIncidentBundleNetworkFlowTableWithExec(t, incidentID, actorID, func(query string, args ...any) error {
		_, err := tx.Exec(context.Background(), query, args...)
		return err
	})
}

func seedIncidentBundleNetworkFlowTableWithExec(t testing.TB, incidentID, actorID string, exec func(string, ...any) error) string {
	t.Helper()
	sessionID := uuid.New()
	unitID := uuid.New()
	tableID := "nft_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	digest := strings.Repeat("1", 64)
	if err := exec(`
INSERT INTO import_sessions (
    import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
    source_file_kind, original_filename, source_content_sha256, source_media_type, source_byte_size,
    parser_profile_id, parser_version, session_status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'network_flow_test', 'csv', 'flows.csv', $5, 'text/csv', 12,
    'network_flow.rfc4180_headered_csv.v1', 'test', 'ready_to_apply', now(), now()
)
`, sessionID, incidentID, actorID, "txn-import-"+unitID.String(), digest); err != nil {
		t.Fatalf("seed Network Flow import session: %v", err)
	}
	if err := exec(`
INSERT INTO import_units (
    import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
    header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
    warning_codes, mapping_fingerprint, approved_mapping_json, columns_json, source_rows_json,
    preview_rows_json, approved_target_kind, approved_extension_profile_id, discovery_sequence,
    created_at, updated_at
) VALUES (
    $1, $2, 'ready', 'csv', 'unit-1', 'A1:Z2', 1, 2, 1, 9,
    '{}', $3, '{}'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,
    'network_flow_table', 'network_flow_activity', 1, now(), now()
)
`, unitID, sessionID, digest); err != nil {
		t.Fatalf("seed Network Flow import unit: %v", err)
	}
	if err := exec(`
INSERT INTO network_flow_tables (
    network_flow_table_id, incident_id, display_name, table_status,
    source_import_session_id, source_import_unit_id, source_content_sha256,
    source_filename_display, source_filename_digest, source_filename_digest_key_id,
    mapping_fingerprint, source_profile_id, parser_profile_id,
    row_count_accepted, row_count_rejected, created_by_user_id
) VALUES (
    $1, $2, 'Retained flows', 'active', $3, $4, $5,
    'flows.csv', $5, 'nf-test-key', $5,
    'network_flow.cisco_sna_netflow_csv.v1', 'network_flow.rfc4180_headered_csv.v1',
    1, 0, $6
)
`, tableID, incidentID, sessionID, unitID, digest, actorID); err != nil {
		t.Fatalf("seed Network Flow table: %v", err)
	}
	return tableID
}
