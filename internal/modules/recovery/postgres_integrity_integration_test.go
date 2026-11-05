package recovery_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestRestoreRejectsBypassedReferentialIntegrity_Integration(t *testing.T) {
	newTarget := func(t *testing.T, name string) (recovery.VNextRestoreTarget, postgres.DB) {
		t.Helper()
		fixture := newRestoreProjectionContractFixture(t, t.Context(), "restore-integrity-"+name, uuid.New())
		projection, err := projectionassembly.Build(fixture.Target.Postgres)
		if err != nil {
			t.Fatal(err)
		}
		fixture.Target.Projections = projection.RecoveryPorts().Rebuilder
		admitted := admitCurrentRestore(t, fixture.capturedBackupContractFixture, fixture.Runner, fixture.Target)
		return recovery.AtomicRestoreTargetForTest(admitted.Fence), fixture.Target.Postgres
	}
	target, _ := newTarget(t, "empty")
	if err := target.WithAtomicRestore(t.Context(), nil, func(recovery.VNextRestoreMutation) error { return nil }); err != nil {
		t.Fatal("empty restore", err)
	}
	t.Run("inactive state metadata is preserved without profile admission", func(t *testing.T) {
		for _, version := range []int{4, 5} {
			target, pool := newTarget(t, "inactive")
			err := target.WithAtomicRestore(t.Context(), nil, func(m recovery.VNextRestoreMutation) error {
				_, err := recovery.RestoreMutationTransactionForTest(m).Exec(t.Context(), `INSERT INTO extension_state_metadata(profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at) VALUES ('network_flow_activity','network_flow_activity.state_v1',$1,1,now(),now())`, version)
				return err
			})
			if err != nil {
				t.Fatalf("structural restore version %d: %v", version, err)
			}
			var restored int
			if err := pool.QueryRow(t.Context(), `SELECT state_version FROM extension_state_metadata WHERE profile_id='network_flow_activity'`).Scan(&restored); err != nil || restored != version {
				t.Fatal("inactive metadata was changed", restored, err)
			}
			if _, err := pool.Exec(t.Context(), `DELETE FROM extension_state_metadata WHERE profile_id='network_flow_activity'`); err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("canonical endpoint restore admission", func(t *testing.T) {
		seed, err := os.ReadFile("../../../fixtures/network-flow-validation/v1/restore.sql")
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"valid", "insert", "update"} {
			t.Run(mode, func(t *testing.T) {
				target, pool := newTarget(t, "endpoint-"+mode)
				err := target.WithAtomicRestore(t.Context(), nil, func(m recovery.VNextRestoreMutation) error {
					tx := recovery.RestoreMutationTransactionForTest(m)
					source := string(seed)
					if mode == "insert" {
						source = strings.Replace(source, "'2001:db8::1'", "'2001:0db8::1'", 1)
					}
					if _, err := tx.Exec(t.Context(), source); err != nil {
						return err
					}
					if mode == "update" {
						_, err := tx.Exec(t.Context(), `UPDATE network_flow_rows SET endpoint_a_ip='2001:0db8::1'`)
						return err
					}
					return nil
				})
				if mode == "valid" {
					if err != nil {
						t.Fatal("canonical inactive restore", err)
					}
					var valid bool
					if err := pool.QueryRow(t.Context(), `SELECT public.network_flow_orientation_state_is_valid_v1() AND (SELECT endpoint_a_ip='2001:db8::1' AND bytes_count='42' AND packets_count='2' FROM network_flow_rows)`).Scan(&valid); err != nil || !valid {
						t.Fatal("restored canonical evidence changed", err)
					}
					// Claim admission applies these owner validators after structural
					// restore has committed; restore itself never executes a profile.
					for _, family := range networkflow.ExtensionStateFamilyCounters() {
						if family.Validate != nil {
							if err := family.Validate(t.Context(), pool); err != nil {
								t.Fatalf("restored %s failed claim validation: %v", family.FamilyID, err)
							}
						}
					}
					return
				}
				var failure *pgconn.PgError
				if !errors.As(err, &failure) || failure.Code != "23514" || failure.ConstraintName != "network_flow_rows_endpoints_ck" {
					t.Fatalf("malformed %s restore escaped canonical invariant: %v", mode, err)
				}
				var rows, tables, completed int
				if err := pool.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM network_flow_rows),(SELECT count(*) FROM network_flow_tables),(SELECT count(*) FROM recovery_target_state WHERE state='complete' OR completion_evidence_sha256 IS NOT NULL)`).Scan(&rows, &tables, &completed); err != nil || rows != 0 || tables != 0 || completed != 0 {
					t.Fatal("invalid restore published state", rows, tables, completed, err)
				}
			})
		}
	})

	target, pool := newTarget(t, "foreign-key")
	err := target.WithAtomicRestore(t.Context(), nil, func(m recovery.VNextRestoreMutation) error {
		tx := recovery.RestoreMutationTransactionForTest(m)
		// CHECKs remain enforced, but this actor FK is bypassed despite convalidated.
		_, err := tx.Exec(t.Context(), `INSERT INTO public.incidents
  (id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
  VALUES ('18600000-0000-4000-8000-000000000001','RESTORE','restore','Restore','active',
  '18600000-0000-4000-8000-000000000002','18600000-0000-4000-8000-000000000002')`)
		if err != nil {
			return err
		}
		var validated bool
		if err := tx.QueryRow(t.Context(), `SELECT bool_and(convalidated AND conenforced) FROM pg_catalog.pg_constraint WHERE conrelid='public.incidents'::regclass`).Scan(&validated); err != nil {
			return err
		}
		if !validated {
			t.Fatal("fixture did not retain valid catalog flags")
		}
		return nil
	})
	if !errors.Is(err, recovery.ErrVNextBackup) {
		t.Fatalf("orphan restore accepted: %v", err)
	}
	var count int
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM public.incidents`).Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected restore published rows", count, err)
	}
	// All relational FKs and rebuilt claims are valid; the bypassed JSON support
	// reference must still fail the Indicator owner's restored-source validator.
	target, pool = newTarget(t, "support-reference")
	err = target.WithAtomicRestore(t.Context(), nil, func(m recovery.VNextRestoreMutation) error {
		tx := recovery.RestoreMutationTransactionForTest(m)
		_, err := tx.Exec(t.Context(), `
INSERT INTO public.users(id,email,display_name,password_hash) VALUES('18600000-0000-4000-8000-000000000002','restore@example.test','Restore','unused');
INSERT INTO public.incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES('18600000-0000-4000-8000-000000000001','RESTORE','restore','Restore','active','18600000-0000-4000-8000-000000000002','18600000-0000-4000-8000-000000000002');
INSERT INTO public.records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id)
VALUES('18600000-0000-4000-8000-000000000003','18600000-0000-4000-8000-000000000001','indicator','18600000-0000-4000-8000-000000000002','18600000-0000-4000-8000-000000000002');
INSERT INTO public.indicators(record_id,incident_id,indicator_type,value_kind,display_value,normalized_value,dedupe_key)
VALUES('18600000-0000-4000-8000-000000000003','18600000-0000-4000-8000-000000000001','domain_name','atomic','restore.example','restore.example','domain_name:restore.example');
SELECT public.rebuild_indicator_active_identities();
INSERT INTO public.indicator_state_intervals(incident_id,indicator_record_id,lifecycle_state,valid_from,support_refs,created_by_user_id)
VALUES('18600000-0000-4000-8000-000000000001','18600000-0000-4000-8000-000000000003','active',now(),'["18600000-0000-4000-8000-000000000004"]','18600000-0000-4000-8000-000000000002');`)
		return err
	})
	if !errors.Is(err, recovery.ErrVNextBackup) {
		t.Fatalf("orphan support restore accepted: %v", err)
	}
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM public.records`).Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected support restore published rows", count, err)
	}

}
