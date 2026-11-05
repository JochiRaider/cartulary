package networkflow_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	reportingmigration "github.com/JochiRaider/cartulary/internal/modules/reporting/migrationadmission"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
)

// Historical fixture bytes remain unchanged. Current admission rejects every
// old interpretation before any migration can rewrite its derived identities.
func TestNetworkFlowOrientationMigrationAdmission(t *testing.T) {
	harness := pgtest.Start(t)
	for _, name := range []string{"canonical_migration", "canonical_preserved_migration", "canonical_historical_alias_migration", "canonical_invalid_precision", "canonical_filter_order_migration", "canonical_duplicate_filters", "canonical_invalid_utc_below", "canonical_invalid_utc_above"} {
		t.Run(name, func(t *testing.T) {
			database := harness.MigrationDatabaseThroughT(t, 33)
			db := database.SQL()
			seedHistoricalOrientationState(t, db, name, false)
			before := orientationAdmissionSnapshot(t, db)
			conn, err := db.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			var blocked interface{ ReasonCode() string }
			admissionErr := ValidateMigrationAdmission(t.Context(), conn)
			conn.Close()
			if !errors.As(admissionErr, &blocked) || blocked.ReasonCode() != "network_flow_orientation_state_incompatible" {
				t.Fatalf("historical state admitted: %v", admissionErr)
			}
			if !reflect.DeepEqual(before, orientationAdmissionSnapshot(t, db)) {
				t.Fatal("admission changed historical state")
			}
		})
	}
	t.Run("pending imports without observations", func(t *testing.T) {
		database := harness.MigrationDatabaseThroughT(t, 33)
		db := database.SQL()
		seedHistoricalOrientationState(t, db, "canonical_migration", true)
		before := orientationAdmissionSnapshot(t, db)
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		var blocked interface{ ReasonCode() string }
		admissionErr := ValidateMigrationAdmission(t.Context(), conn)
		conn.Close()
		if !errors.As(admissionErr, &blocked) || blocked.ReasonCode() != "network_flow_orientation_state_incompatible" {
			t.Fatalf("historical approval admitted: %v", admissionErr)
		}
		if !reflect.DeepEqual(before, orientationAdmissionSnapshot(t, db)) {
			t.Fatal("admission changed pending import")
		}
	})
	t.Run("historical Reporting frozen input is unchanged", func(t *testing.T) {
		database := harness.MigrationDatabaseThroughT(t, 37)
		db := database.SQL()
		seedHistoricalOrientationState(t, db, "canonical_migration", true)
		const actor = "11111111-1111-4111-8111-111111111111"
		const incident = "22222222-2222-4222-8222-222222222222"
		const job = "99999999-9999-4999-8999-999999999999"
		if _, err := db.ExecContext(t.Context(), `INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id) VALUES($1,'incident',$2,'queued',true,$3,now(),now(),0,'incident_membership','reporting.release.create_v1','reporting.release.create_v1','reporting.render.phases.v1')`, job, incident, actor); err != nil {
			t.Fatal(err)
		}
		request := map[string]any{"schema_id": "cartulary.reporting_render_admission.v1", "preparation_algorithm": "materialize_retained_reporting_render_v1", "source_capture_sha256": strings.Repeat("a", 64), "authored_redaction_profile_sha256": strings.Repeat("b", 64), "authored_redaction_profile": map[string]any{"schema_id": "cartulary.redaction_profile.v3", "version": 5, "allowed_release_scopes": []string{"internal_review"}}, "release_scope": "internal_review", "template_contract": map[string]any{"schema_id": "cartulary.reporting_template_pack_manifest.v3"}, "template_sha256": strings.Repeat("c", 64), "graph_sources_sha256": strings.Repeat("d", 64), "graph_sources": map[string]any{"schema_id": "cartulary.reporting_graph_input.v1", "results": []any{}}, "identity": map[string]any{"release_id": "historical-release", "preview_attempt_id": nil}}
		raw, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(t.Context(), `INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,request_json,admitted_profile_sha256,admitted_graph_sources_sha256,admitted_template_manifest_sha256,render_admission_sha256) VALUES($1,'release_create',$2,$3,$4,repeat('b',64),repeat('d',64),repeat('c',64),repeat('e',64))`, job, incident, actor, string(raw)); err != nil {
			t.Fatal(err)
		}
		before := orientationAdmissionSnapshot(t, db)
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		admissionErr := reportingmigration.Validate(t.Context(), conn)
		conn.Close()
		var blocked interface{ ReasonCode() string }
		if !errors.As(admissionErr, &blocked) || blocked.ReasonCode() != "reporting_direction_state_incompatible" {
			t.Fatalf("historical frozen input admitted: %v", admissionErr)
		}
		if !reflect.DeepEqual(before, orientationAdmissionSnapshot(t, db)) {
			t.Fatal("Reporting admission changed historical state")
		}
	})
	for _, kind := range []string{"receipt", "job", "expired_job", "proof"} {
		t.Run("retained "+kind+" without observations", func(t *testing.T) {
			database := harness.MigrationDatabaseThroughT(t, 37)
			db := database.SQL()
			seedHistoricalOrientationState(t, db, "canonical_migration", true)
			for _, query := range []string{"DELETE FROM import_units", "DELETE FROM import_sessions"} {
				if _, err := db.ExecContext(t.Context(), query); err != nil {
					t.Fatal(err)
				}
			}
			const actor = "11111111-1111-4111-8111-111111111111"
			const incident = "22222222-2222-4222-8222-222222222222"
			const job = "99999999-9999-4999-8999-999999999999"
			switch kind {
			case "receipt":
				if _, err := db.ExecContext(t.Context(), `INSERT INTO route_idempotency(route_key,scope_key,client_txn_id,actor_user_id,request_hash,status_code,response_json) VALUES('nf.tables.patch',$1,'historical',$2,decode(repeat('a',64),'hex'),200,'{}')`, incident, actor); err != nil {
					t.Fatal(err)
				}
			case "job", "expired_job":
				if _, err := db.ExecContext(t.Context(), `INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id) VALUES($1,'incident',$2,'queued',true,$3,now(),now(),0,'incident_membership','network_flow_activity.graph_view_worker_v1','network_flow_activity.graph_view_materialize_v1','network_flow_activity.graph_view_materialize.projection_result.v1')`, job, incident, actor); err != nil {
					t.Fatal(err)
				}
				if kind == "expired_job" {
					if _, err := db.ExecContext(t.Context(), `UPDATE jobs SET status='canceled',cancelable=false,finished_at=now(),retained_until=now(),expired_at=now() WHERE job_id=$1`, job); err != nil {
						t.Fatal(err)
					}
				}
			case "proof":
				if _, err := db.ExecContext(t.Context(), `INSERT INTO jobs(job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,submitted_at,updated_at,progress_completed,auth_policy,handler_name,job_kind,progress_unit_id) VALUES($1,'incident',$2,'queued',true,$3,now(),now(),0,'incident_membership','reporting.release.create_v1','reporting.release.create_v1','reporting.render.phases.v1')`, job, incident, actor); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(t.Context(), `INSERT INTO extension_job_commit_proofs(job_id,owner_profile_id,operation_kind,final_commit_id,normalized_request_sha256,terminal_result,terminal_result_sha256,resource_refs,committed_at) VALUES($1,'network_flow_activity','materialize','historical',repeat('a',64),'{}',repeat('b',64),'[]',now())`, job); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := db.ExecContext(t.Context(), "INSERT INTO application_crypto_format(singleton,format_id) VALUES(true,'cartulary.application_crypto_format.v1') ON CONFLICT DO NOTHING"); err != nil {
				t.Fatal(err)
			}
			before := orientationAdmissionSnapshot(t, db)
			conn, err := db.Conn(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			admissionErr := ValidateMigrationAdmission(t.Context(), conn)
			conn.Close()
			var blocked interface{ ReasonCode() string }
			if !errors.As(admissionErr, &blocked) || blocked.ReasonCode() != "network_flow_orientation_state_incompatible" {
				t.Fatalf("historical %s admitted: %v", kind, admissionErr)
			}
			if kind == "receipt" {
				target, err := database.RuntimeDatabase()
				if err != nil {
					t.Fatal(err)
				}
				guard, err := pgtest.OpenPurposeDatabase(t.Context(), target.DSN, postgres.PurposeMigration)
				if err != nil {
					t.Fatal(err)
				}
				defer guard.Close()
				source, err := dbmigrations.Source()
				if err != nil {
					t.Fatal(err)
				}
				admissionErr = migrations.ApplyWithCryptoAdmission(t.Context(), db, guard, source, func(context.Context) error { return nil }, ValidateMigrationAdmission, reportingmigration.Validate)
				if !errors.As(admissionErr, &blocked) || blocked.ReasonCode() != "network_flow_orientation_state_incompatible" {
					t.Fatalf("migration wrote before owner admission: %v", admissionErr)
				}
			}
			if !reflect.DeepEqual(before, orientationAdmissionSnapshot(t, db)) {
				t.Fatal("admission changed retained state")
			}
		})
	}
	t.Run("fresh state initializes", func(t *testing.T) {
		database := harness.MigrationDatabaseThroughT(t, 37)
		db := database.SQL()
		conn, err := db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		err = ValidateMigrationAdmission(t.Context(), conn)
		conn.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err := database.ApplyThrough(t.Context(), 38); err != nil {
			t.Fatal(err)
		}
		conn, err = db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		if err := ValidateMigrationAdmission(t.Context(), conn); err != nil {
			t.Fatal(err)
		}
	})
}

func seedHistoricalOrientationState(t *testing.T, db *sql.DB, name string, importsOnly bool) {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Seed []struct {
			Table string                     `json:"table"`
			Row   map[string]json.RawMessage `json:"row"`
		} `json:"seed"`
	}
	if err := json.Unmarshal(raw, &vector); err != nil {
		t.Fatal(err)
	}
	for _, entry := range vector.Seed {
		if importsOnly && entry.Table != "users" && entry.Table != "incidents" && entry.Table != "incident_memberships" && entry.Table != "import_sessions" && entry.Table != "import_units" {
			continue
		}
		if entry.Table == "users" {
			flowtest.SeedActor(t, db, flowtest.ActorAttributes{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Email: "orientation-history@example.invalid", DisplayName: "Historical fixture"})
			continue
		}
		columns := make([]string, 0, len(entry.Row))
		for key := range entry.Row {
			columns = append(columns, `"`+strings.ReplaceAll(key, `"`, `""`)+`"`)
		}
		sort.Strings(columns)
		table := `public."` + strings.ReplaceAll(entry.Table, `"`, `""`) + `"`
		names := strings.Join(columns, ",")
		row, err := json.Marshal(entry.Row)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(t.Context(), "INSERT INTO "+table+" ("+names+") SELECT "+names+" FROM jsonb_populate_record(NULL::"+table+",$1::jsonb)", string(row)); err != nil {
			t.Fatal(err)
		}
	}
}

func orientationAdmissionSnapshot(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	state := map[string]string{}
	for _, table := range []string{"goose_db_version", "extension_state_metadata", "network_flow_tables", "network_flow_rows", "network_flow_rejected_row_diagnostics", "network_flow_indicator_bindings", "network_flow_graph_views", "import_sessions", "import_units", "import_apply_unit_plans", "import_unit_apply_outcomes", "import_apply_journal", "route_idempotency", "jobs", "extension_job_commit_proofs", "deployment_admin_audit_events", "reporting_job_payloads"} {
		var value string
		if err := db.QueryRowContext(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(value) ORDER BY to_jsonb(value)::text),'[]'::jsonb)::text FROM "+table+" value").Scan(&value); err != nil {
			t.Fatal(err)
		}
		state[table] = value
	}
	return state
}
