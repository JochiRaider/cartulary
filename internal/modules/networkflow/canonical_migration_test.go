package networkflow

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Authored vectors use independent NUL framing and SHA-256, including the
// historical serializer. Neither expected IDs nor hashes come from runtime code.
func TestNetworkFlowCanonicalIdentityMigration(t *testing.T) {
	harness := pgtest.Start(t)
	for _, name := range []string{"canonical_migration", "canonical_preserved_migration", "canonical_historical_alias_migration", "canonical_invalid_precision", "canonical_filter_order_migration", "canonical_duplicate_filters", "canonical_invalid_utc_below", "canonical_invalid_utc_above"} {
		t.Run(name, func(t *testing.T) { assertCanonicalIdentityMigration(t, harness, "testdata/"+name+".json") })
	}
}

func assertCanonicalIdentityMigration(t *testing.T, harness *pgtest.Harness, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		HistoricalAliasSource string `json:"historical_alias_source_utf8"`
		ExpectedFailure       bool   `json:"expected_failure"`
		FailureScenario       string `json:"failure_scenario"`
		Seed                  []struct {
			Table string                     `json:"table"`
			Row   map[string]json.RawMessage `json:"row"`
		} `json:"seed"`
		Expected struct {
			Fingerprint                string            `json:"mapping_fingerprint"`
			Digest                     string            `json:"normalized_digest"`
			UnmappedRaw                json.RawMessage   `json:"unmapped_raw"`
			RowIDs                     []string          `json:"row_ids"`
			DiagnosticID               string            `json:"diagnostic_id"`
			Filters                    json.RawMessage   `json:"filters"`
			SourceRefRowIDs            json.RawMessage   `json:"source_ref_row_ids"`
			SourceRefs                 json.RawMessage   `json:"source_refs"`
			ReceiptHash                string            `json:"receipt_hash"`
			SemanticHash               string            `json:"semantic_hash"`
			JobHashes                  map[string]string `json:"job_hashes"`
			CurrentApprovalFingerprint string            `json:"current_approval_fingerprint"`
		} `json:"expected"`
		Legacy struct {
			Fingerprint string   `json:"mapping_fingerprint"`
			RowIDs      []string `json:"row_ids"`
			ReceiptHash string   `json:"receipt_hash"`
		} `json:"legacy"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.HistoricalAliasSource != "" {
		var historical struct {
			OwnerMapping json.RawMessage `json:"owner_mapping"`
		}
		for _, entry := range fixture.Seed {
			if entry.Table != "import_units" {
				continue
			}
			var status string
			if err := json.Unmarshal(entry.Row["unit_status"], &status); err != nil {
				t.Fatal(err)
			}
			if status == "applied" {
				if err := json.Unmarshal(entry.Row["approved_mapping_json"], &historical); err != nil {
					t.Fatal(err)
				}
			}
		}
		var mapping approvedMapping
		if err := json.Unmarshal(historical.OwnerMapping, &mapping); err != nil {
			t.Fatal(err)
		}
		sourceSHA := sha256Hex([]byte(fixture.HistoricalAliasSource))
		fingerprint, err := mappingFingerprint(mapping, sourceSHA)
		if err != nil || fingerprint != fixture.Expected.Fingerprint {
			t.Fatalf("historical descriptors were reinterpreted: %s, %v", fingerprint, err)
		}
		var candidate map[string]json.RawMessage
		if err := json.Unmarshal(historical.OwnerMapping, &candidate); err != nil {
			t.Fatal(err)
		}
		delete(candidate, "source_columns")
		request, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		facade := &importFacade{sourceStore: previewSource{csv: fixture.HistoricalAliasSource}, limits: defaultLimits()}
		prepared, err := facade.PrepareImportUnitMapping(t.Context(), imports.ExtensionImportMappingRequest{
			TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID,
			OwnerMappingSchemaID: mappingCandidateSchemaID, OwnerMapping: request,
		})
		if err != nil || facade.ValidateImportUnitMappingResult(prepared) != nil {
			t.Fatalf("new source approval failed: %v", err)
		}
		if prepared.MappingFingerprint != fixture.Expected.CurrentApprovalFingerprint || prepared.MappingFingerprint == fixture.Expected.Fingerprint {
			t.Fatalf("new approval did not use independently expected NFC descriptors: %s", prepared.MappingFingerprint)
		}
	}
	scenarios := []string{"repair", "missing_mapping", "tampered_receipt", "tampered_diagnostic", "tampered_binding", "missing_binding_fingerprint", "wrong_binding_fingerprint", "wrong_binding_locator", "wrong_binding_identity_order", "extra_binding_member", "duplicate_binding_ref", "tampered_raw", "invalid_utc", "tampered_proof", "missing_proof", "inflight_apply", "core_inflight_apply"}
	if len(fixture.Expected.Filters) != 0 {
		scenarios = []string{"repair", "tampered_receipt", "tampered_proof"}
	}
	if fixture.ExpectedFailure {
		scenarios = []string{"invalid_precision"}
		if fixture.FailureScenario != "" {
			scenarios = []string{fixture.FailureScenario}
		}
	}
	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			database := harness.MigrationDatabaseThroughT(t, 33)
			db := database.SQL()
			exec := func(query string, args ...any) {
				t.Helper()
				if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
					t.Fatal(err)
				}
			}
			for _, entry := range fixture.Seed {
				columns := make([]string, 0, len(entry.Row))
				for column := range entry.Row {
					columns = append(columns, `"`+strings.ReplaceAll(column, `"`, `""`)+`"`)
				}
				sort.Strings(columns)
				row, err := json.Marshal(entry.Row)
				if err != nil {
					t.Fatal(err)
				}
				table := `public."` + strings.ReplaceAll(entry.Table, `"`, `""`) + `"`
				names := strings.Join(columns, ",")
				exec("INSERT INTO "+table+" ("+names+") SELECT "+names+" FROM jsonb_populate_record(NULL::"+table+",$1::jsonb)", string(row))
			}
			stableBefore := canonicalMigrationStableState(t, db)
			switch scenario {
			case "missing_mapping":
				exec("UPDATE import_units SET approved_mapping_json=NULL")
			case "tampered_binding":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,network_flow_row_id}',to_jsonb('nfr_' || repeat('f',64))), source_row_ref_row_ids=ARRAY['nfr_' || repeat('f',64)]")
			case "missing_binding_fingerprint":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=source_row_refs #- '{0,mapping_fingerprint}'")
			case "wrong_binding_fingerprint":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,mapping_fingerprint}',to_jsonb(repeat('f',64)))")
			case "wrong_binding_locator":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,source_row_number}','99'::jsonb)")
			case "wrong_binding_identity_order":
				exec("UPDATE network_flow_indicator_bindings SET source_row_ref_row_ids=ARRAY[source_row_ref_row_ids[2],source_row_ref_row_ids[1]]")
			case "extra_binding_member":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,extra}','true'::jsonb)")
			case "duplicate_binding_ref":
				exec("UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_build_array(source_row_refs->0,source_row_refs->0), source_row_ref_row_ids=ARRAY[source_row_ref_row_ids[1],source_row_ref_row_ids[1]]")
			case "tampered_diagnostic":
				exec("ALTER TABLE network_flow_rejected_row_diagnostics DISABLE TRIGGER network_flow_rejected_row_diagnostics_immutable_update")
				exec("UPDATE network_flow_rejected_row_diagnostics SET diagnostic_id='nfd_' || repeat('f',64)")
				exec("ALTER TABLE network_flow_rejected_row_diagnostics ENABLE TRIGGER network_flow_rejected_row_diagnostics_immutable_update")
			case "invalid_utc":
				exec("ALTER TABLE network_flow_rows DISABLE TRIGGER network_flow_rows_immutable_update")
				exec("UPDATE network_flow_rows SET flow_start_utc='0001-12-31 23:59:00+00 BC'::timestamptz")
				exec("ALTER TABLE network_flow_rows ENABLE TRIGGER network_flow_rows_immutable_update")
			case "tampered_raw":
				exec("ALTER TABLE network_flow_rows DISABLE TRIGGER network_flow_rows_immutable_update")
				exec(`UPDATE network_flow_rows SET unmapped_raw=jsonb_set(unmapped_raw,'{2,raw_header_text}','"wrong"'::jsonb)`)
				exec("ALTER TABLE network_flow_rows ENABLE TRIGGER network_flow_rows_immutable_update")
			case "tampered_receipt":
				exec("UPDATE route_idempotency SET request_hash=decode(repeat('f',64),'hex')")
			case "tampered_proof":
				exec("UPDATE extension_job_commit_proofs SET normalized_request_sha256=repeat('f',64)")
			case "missing_proof":
				exec("DELETE FROM extension_job_commit_proofs")
			case "core_inflight_apply":
				// Registry identity covers ordinary apply jobs without Network Flow plans.
				exec(`INSERT INTO jobs (job_id,scope_kind,incident_id,status,cancelable,submitted_by_user_id,
      submitted_at,updated_at,progress_completed,progress_total,auth_policy,handler_name,handler_payload_json,job_kind,progress_unit_id)
      SELECT '99999999-9999-4999-8999-999999999999','incident',incident_id,'queued',true,created_by_user_id,
       created_at,created_at,0,1,'incident_membership','imports.apply_v1',jsonb_build_object('import_session_id',import_session_id),
       'import.apply_v1','import.apply.import_unit.v1' FROM import_sessions ORDER BY import_session_id LIMIT 1`)
			case "inflight_apply":
				exec("UPDATE import_units SET unit_status='applying'")
			}
			rawState := func() []byte {
				var value []byte
				if err := db.QueryRowContext(t.Context(), "SELECT jsonb_agg(unmapped_raw ORDER BY source_row_number) FROM network_flow_rows").Scan(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			rawBefore := rawState()
			rollbackBefore := canonicalMigrationRollbackState(t, db)
			err := database.ApplyThrough(t.Context(), 34)
			if scenario == "core_inflight_apply" && (err == nil || !strings.Contains(err.Error(), "drain Network Flow and all import apply jobs")) {
				t.Fatalf("registry-wide quiescence guard did not reject before rewriting: %v", err)
			}
			if scenario != "repair" {
				if err == nil {
					t.Fatal("inconsistent source state passed migration")
				}
				if err := compareFixtureJSON(rollbackBefore, canonicalMigrationRollbackState(t, db)); err != nil {
					t.Fatal("failed migration changed durable state:", err)
				}
				if err := compareFixtureJSON(rawBefore, rawState()); err != nil {
					t.Fatal("rollback changed raw provenance:", err)
				}
				canonicalMigrationScalar(t, db, "SELECT mapping_fingerprint FROM network_flow_tables", fixture.Legacy.Fingerprint)
				canonicalMigrationScalar(t, db, "SELECT table_version::text FROM network_flow_tables", "7")
				canonicalMigrationScalar(t, db, "SELECT network_flow_row_id FROM network_flow_rows WHERE source_row_number=2", fixture.Legacy.RowIDs[0])
				canonicalMigrationScalar(t, db, "SELECT count(*)::text FROM goose_db_version WHERE version_id=34 AND is_applied", "0")
			} else {
				if err != nil {
					// The lifecycle port deliberately exposes only safe error codes.
					// Diagnose this authored fixture's SQL inside a rolled-back
					// transaction; this is never an alternate acceptance path.
					source, readErr := os.ReadFile(filepath.Join(RepoRoot(t), "db", "migrations", "00034_network_flow_canonical_identity.sql"))
					if readErr != nil {
						t.Fatal(readErr)
					}
					tx, beginErr := db.BeginTx(t.Context(), nil)
					if beginErr != nil {
						t.Fatal(beginErr)
					}
					_, sqlErr := tx.ExecContext(t.Context(), strings.SplitN(string(source), "-- +goose Down", 2)[0])
					rollbackErr := tx.Rollback()
					t.Fatalf("migration: %v; rolled-back SQL diagnosis: %v; rollback: %v", err, sqlErr, rollbackErr)
				}
				if err := compareFixtureJSON(stableBefore, canonicalMigrationStableState(t, db)); err != nil {
					t.Fatal("migration rewrote unrelated retained facts:", err)
				}
				wantRaw, err := json.Marshal([]json.RawMessage{fixture.Expected.UnmappedRaw, fixture.Expected.UnmappedRaw})
				if err != nil {
					t.Fatal(err)
				}
				if err := compareFixtureJSON(wantRaw, rawState()); err != nil {
					t.Fatal("retained provenance differs after migration:", err)
				}
				canonicalMigrationScalar(t, db, "SELECT mapping_fingerprint FROM network_flow_tables", fixture.Expected.Fingerprint)
				canonicalMigrationScalar(t, db, "SELECT mapping_fingerprint FROM import_units WHERE unit_status='applied'", fixture.Legacy.Fingerprint)
				canonicalMigrationScalar(t, db, "SELECT table_version::text FROM network_flow_tables", "8")
				canonicalMigrationScalar(t, db, "SELECT unit_status FROM import_units WHERE import_unit_id='99999999-9999-4999-8999-999999999999'", "selected")
				canonicalMigrationScalar(t, db, "SELECT (approved_mapping_json IS NULL AND mapping_fingerprint IS NULL AND approved_target_kind IS NULL)::text FROM import_units WHERE import_unit_id='99999999-9999-4999-8999-999999999999'", "true")
				canonicalMigrationScalar(t, db, "SELECT session_status FROM import_sessions WHERE import_session_id='88888888-8888-4888-8888-888888888888'", "discovered")
				canonicalMigrationScalar(t, db, "SELECT count(*)::text FROM network_flow_rows WHERE normalized_row_digest_sha256='"+fixture.Expected.Digest+"'", "2")
				canonicalMigrationScalar(t, db, "SELECT count(*)::text FROM network_flow_rows WHERE observation_source_ref->>'mapping_fingerprint'='"+fixture.Expected.Fingerprint+"'", "2")
				for i, id := range fixture.Expected.RowIDs {
					canonicalMigrationScalar(t, db, fmt.Sprintf("SELECT network_flow_row_id FROM network_flow_rows WHERE source_row_number=%d", i+2), id)
				}
				canonicalMigrationScalar(t, db, "SELECT diagnostic_id FROM network_flow_rejected_row_diagnostics", fixture.Expected.DiagnosticID)
				var refs []byte
				if err := db.QueryRowContext(t.Context(), "SELECT source_row_refs FROM network_flow_indicator_bindings").Scan(&refs); err != nil {
					t.Fatal(err)
				}
				if err := compareFixtureJSON(fixture.Expected.SourceRefs, refs); err != nil {
					t.Fatal(err)
				}
				var identityIDs []byte
				if err := db.QueryRowContext(t.Context(), "SELECT to_jsonb(source_row_ref_row_ids) FROM network_flow_indicator_bindings").Scan(&identityIDs); err != nil {
					t.Fatal(err)
				}
				if err := compareFixtureJSON(fixture.Expected.SourceRefRowIDs, identityIDs); err != nil {
					t.Fatal("binding identity projection:", err)
				}
				canonicalMigrationScalar(t, db, "SELECT encode(request_hash,'hex') FROM route_idempotency WHERE route_key='nf.graph_views.patch'", fixture.Expected.ReceiptHash)
				canonicalMigrationScalar(t, db, "SELECT response_json->'graph_view'->>'semantic_query_sha256' FROM route_idempotency WHERE route_key='nf.graph_views.patch'", fixture.Expected.SemanticHash)
				// The corrected receipt must also match the serving replay policy.
				hash, err := mutationRequestHash(routeKeyGraphViewsPatch, "graph_view_id:nfgv_"+strings.Repeat("3", 32), map[string]any{"base_graph_view_version": 1, "display_name": "A&B\u2028Z"})
				if err != nil || fmt.Sprintf("%x", hash) != fixture.Expected.ReceiptHash {
					t.Fatalf("replay identity: %x, %v", hash, err)
				}
				// Original source identity and response bytes are historical facts.
				canonicalMigrationScalar(t, db, "SELECT count(*)::text FROM route_idempotency WHERE response_json->'graph_view'->>'display_name'='A&B' || chr(8232) || 'Z'", "1")
				for jobID, digest := range fixture.Expected.JobHashes {
					canonicalMigrationScalar(t, db, "SELECT normalized_request_sha256 FROM extension_job_commit_proofs WHERE job_id='"+jobID+"'", digest)
					canonicalMigrationScalar(t, db, "SELECT coalesce(extension_normalized_request_sha256,'expired') FROM jobs WHERE job_id='"+jobID+"'", func() string {
						if strings.HasPrefix(jobID, "7777") {
							return "expired"
						}
						return digest
					}())
				}
				definition := jobs.Definition{JobKind: GraphViewMaterializationJobKind, HandlerName: graphViewWorkerKind, ProgressUnitID: graphViewProgressUnitID,
					Extension: &jobs.ExtensionPolicy{OwnerProfileID: ProfileID, OperationKind: "network_flow_activity.graph_view_materialize", MaxProofBytes: 1048576,
						ResourceRefs: []jobs.ExtensionResourceRefContract{{Kind: graphViewResultResourceKind, MaxRefs: 1}}}}
				connection, err := db.Conn(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				admissionErr := connection.Raw(func(driver any) error {
					reader := driver.(*stdlib.Conn).Conn()
					adapter := savedGraphReceiptAdapter{reader: canonicalMigrationReceiptReader{reader}}
					for _, entry := range fixture.Seed {
						if entry.Table != "route_idempotency" {
							continue
						}
						var retained struct {
							Route    string          `json:"route_key"`
							Actor    uuid.UUID       `json:"actor_user_id"`
							Scope    string          `json:"scope_key"`
							Client   string          `json:"client_txn_id"`
							Status   int             `json:"status_code"`
							Response json.RawMessage `json:"response_json"`
						}
						encoded, err := json.Marshal(entry.Row)
						if err != nil {
							return err
						}
						if err = json.Unmarshal(encoded, &retained); err != nil {
							return err
						}
						var expected map[string]any
						decoder := json.NewDecoder(bytes.NewReader(retained.Response))
						decoder.UseNumber()
						if err = decoder.Decode(&expected); err != nil {
							return err
						}
						expected["graph_view"].(map[string]any)["semantic_query_sha256"] = fixture.Expected.SemanticHash
						if len(fixture.Expected.Filters) != 0 {
							expected["graph_view"].(map[string]any)["semantic_query"].(map[string]any)["filters"] = fixture.Expected.Filters
						}
						wantHash := fixture.Expected.ReceiptHash
						if job, ok := expected["job"].(map[string]any); ok {
							wantHash = fixture.Expected.JobHashes[job["job_id"].(string)]
						}
						hash, err := hex.DecodeString(wantHash)
						if err != nil || len(hash) != 32 {
							return fmt.Errorf("invalid independent replay vector: %v", err)
						}
						key := authn.RouteIdempotencyKey{RouteKey: retained.Route, ActorUserID: retained.Actor, ScopeKey: retained.Scope, ClientTxnID: retained.Client}
						outcome, found, err := adapter.replay(t.Context(), key, hash)
						if err != nil || !found || savedGraphOutcomeStatus(outcome.kind) != retained.Status {
							return fmt.Errorf("migrated serving replay: found=%v error=%v", found, err)
						}
						want, err := json.Marshal(expected)
						if err != nil {
							return err
						}
						got, err := json.Marshal(savedGraphOutcomePayload(outcome))
						if err != nil {
							return err
						}
						if err = compareFixtureJSON(want, got); err != nil {
							return fmt.Errorf("migrated complete historical response: %w", err)
						}
						hash[0] ^= 1
						if _, found, err = adapter.replay(t.Context(), key, hash); !found || !errors.Is(err, authn.ErrClientTxnConflict) {
							return fmt.Errorf("changed migrated replay did not conflict: %v", err)
						}
					}

					if err := validatePersistedGraphViewFamily(t.Context(), reader); err != nil {
						return err
					}
					var refs []networkFlowRowRef
					if err := json.Unmarshal(fixture.Expected.SourceRefs, &refs); err != nil {
						return err
					}
					linker := indicatorLinkApplication{store: newStore(reader, defaultEffectiveLimits())}
					for _, ref := range refs {
						row, failure := linker.getAcceptedRowForLink(t.Context(), uuid.MustParse("22222222-2222-4222-8222-222222222222"), ref.NetworkFlowTableID, ref.NetworkFlowRowID, &ref)
						if failure != nil || row.MappingFingerprint != ref.MappingFingerprint {
							return fmt.Errorf("migrated reference cannot be reused: %v", failure)
						}
					}
					if err := ValidateRetainedExtensionState(t.Context(), reader); err != nil {
						return err
					}
					if err := assertMigratedBindingAdmission(t, reader); err != nil {
						return err
					}
					return ValidateSavedGraphAdmission(t.Context(), reader, definition)
				})
				closeErr := connection.Close()
				if admissionErr != nil || closeErr != nil {
					t.Fatalf("post-migration owner admission: %v; connection release: %v", admissionErr, closeErr)
				}
			}
			if scenario == "repair" {
				if err := compareFixtureJSON(stableBefore, canonicalMigrationStableState(t, db)); err != nil {
					t.Fatal("replay changed retained facts:", err)
				}
			}
			// Both success and rollback restore immutability and leave no helper API.
			if _, err := db.ExecContext(t.Context(), "UPDATE network_flow_rows SET mapping_fingerprint=repeat('0',64)"); err == nil {
				t.Fatal("row immutability was not restored")
			}
			if _, err := db.ExecContext(t.Context(), "UPDATE network_flow_rejected_row_diagnostics SET reason_code='other'"); err == nil {
				t.Fatal("diagnostic immutability was not restored")
			}
			canonicalMigrationScalar(t, db, "SELECT count(*)::text FROM pg_proc WHERE proname LIKE 'nf34_%'", "0")
		})
	}
}

func canonicalMigrationScalar(t *testing.T, db *sql.DB, query, expected string) {
	t.Helper()
	var actual string
	if err := db.QueryRowContext(t.Context(), query).Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("%s: got %q, want %q", query, actual, expected)
	}
}

// Compare all retained facts except the explicitly authorized derived identities.
// This is a preservation assertion, not the oracle for newly computed values.
func canonicalMigrationStableState(t *testing.T, db *sql.DB) []byte {
	t.Helper()
	var data []byte
	err := db.QueryRowContext(t.Context(), `SELECT jsonb_build_object(
        'tables',(SELECT jsonb_agg(to_jsonb(t)-'mapping_fingerprint'-'table_version' ORDER BY network_flow_table_id) FROM network_flow_tables t),
        'rows',(SELECT jsonb_agg((to_jsonb(r)-'network_flow_row_id'-'mapping_fingerprint'-'normalized_row_digest_sha256'-'unmapped_raw') #- '{observation_source_ref,mapping_fingerprint}' ORDER BY source_row_number) FROM network_flow_rows r),
        'diagnostics',(SELECT jsonb_agg(to_jsonb(d)-'diagnostic_id' ORDER BY source_row_number) FROM network_flow_rejected_row_diagnostics d),
        'units',(SELECT jsonb_agg(to_jsonb(u) ORDER BY import_unit_id) FROM import_units u WHERE unit_status='applied'),
        'graphs',(SELECT jsonb_agg(jsonb_set(to_jsonb(g)-'semantic_query_sha256','{semantic_query_json,filters}',(SELECT coalesce(jsonb_agg(f ORDER BY f::text COLLATE "C"),'[]'::jsonb) FROM jsonb_array_elements(g.semantic_query_json->'filters') f)) ORDER BY graph_view_id) FROM network_flow_graph_views g),
        'receipts',(SELECT jsonb_agg(jsonb_set((to_jsonb(r)-'request_hash') #- '{response_json,graph_view,semantic_query_sha256}','{response_json,graph_view,semantic_query,filters}',(SELECT coalesce(jsonb_agg(f ORDER BY f::text COLLATE "C"),'[]'::jsonb) FROM jsonb_array_elements(r.response_json->'graph_view'->'semantic_query'->'filters') f)) ORDER BY scope_key,client_txn_id) FROM route_idempotency r),
        'jobs',(SELECT jsonb_agg(to_jsonb(j)-'extension_normalized_request_sha256' ORDER BY job_id) FROM jobs j),
        'proofs',(SELECT jsonb_agg(to_jsonb(p)-'normalized_request_sha256' ORDER BY job_id) FROM extension_job_commit_proofs p)
    )`).Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Keep the receipt owner storage boundary in the migration replay test.
type canonicalMigrationReceiptReader struct{ reader authn.RouteIdempotencyQuerier }

func (r canonicalMigrationReceiptReader) GetRouteIdempotency(ctx context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	return authn.GetRouteIdempotencyTx(ctx, r.reader, key)
}

func canonicalMigrationRollbackState(t *testing.T, db *sql.DB) []byte {
	t.Helper()
	state := map[string]json.RawMessage{}
	for _, table := range []string{"network_flow_tables", "network_flow_rows", "network_flow_rejected_row_diagnostics", "network_flow_indicator_bindings", "network_flow_graph_views", "import_sessions", "import_units", "import_apply_unit_plans", "import_unit_apply_outcomes", "import_apply_journal", "route_idempotency", "jobs", "extension_job_commit_proofs", "deployment_admin_audit_events", "goose_db_version"} {
		var data []byte
		if err := db.QueryRowContext(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(value) ORDER BY to_jsonb(value)::text),'[]'::jsonb) FROM "+table+" value").Scan(&data); err != nil {
			t.Fatal(err)
		}
		state[table] = data
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Corrupt only live state in rollback-only transactions. Historical receipts
// retain their own shape/identity admission and never join repaired row IDs.
func assertMigratedBindingAdmission(t *testing.T, reader *pgx.Conn) error {
	t.Helper()
	for _, query := range []string{
		"UPDATE network_flow_indicator_bindings SET source_row_refs=source_row_refs #- '{0,mapping_fingerprint}'",
		"UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,mapping_fingerprint}',to_jsonb(repeat('f',64)))",
		"UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,source_row_number}','99'::jsonb)",
		"UPDATE network_flow_indicator_bindings SET source_row_refs=jsonb_set(source_row_refs,'{0,extra}','true'::jsonb)",
		"UPDATE network_flow_indicator_bindings SET source_row_ref_row_ids=ARRAY[source_row_ref_row_ids[2],source_row_ref_row_ids[1]]",
	} {
		tx, err := reader.Begin(t.Context())
		if err != nil {
			return err
		}
		_, execErr := tx.Exec(t.Context(), query)
		admissionErr := ValidateRetainedExtensionState(t.Context(), tx)
		rollbackErr := tx.Rollback(t.Context())
		if execErr != nil || admissionErr == nil || rollbackErr != nil {
			return fmt.Errorf("live corruption admission: mutation=%v admission=%v rollback=%v", execErr, admissionErr, rollbackErr)
		}
	}
	tx, err := reader.Begin(t.Context())
	if err != nil {
		return err
	}
	_, execErr := tx.Exec(t.Context(), "UPDATE network_flow_tables SET table_status='soft_deleted',deleted_at=created_at")
	admissionErr := ValidateRetainedExtensionState(t.Context(), tx)
	rollbackErr := tx.Rollback(t.Context())
	if execErr != nil || admissionErr != nil || rollbackErr != nil {
		return fmt.Errorf("retained soft-deleted lineage: mutation=%v admission=%v rollback=%v", execErr, admissionErr, rollbackErr)
	}
	return nil
}
