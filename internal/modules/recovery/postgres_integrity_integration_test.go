package recovery

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreRejectsBypassedReferentialIntegrity_Integration(t *testing.T) {
	targetDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "restore-integrity")
	dsn, err := targetDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	target := &vNextRestoreTarget{target: RestoreTarget{Postgres: pool}}
	if err := target.WithAtomicRestore(t.Context(), nil, func(VNextRestoreMutation) error { return nil }); err != nil {
		t.Fatal("empty restore", err)
	}
	t.Run("owner state admission before restore commit", func(t *testing.T) {
		bound := &vNextRestoreTarget{target: RestoreTarget{Postgres: pool}, extensionBackups: &ExtensionBackupCatalog{bindings: []ExtensionBackupBinding{{ProfileID: "network_flow_activity", PostRestoreValidationAlgorithm: "network_flow_activity.validate_state_v5", ValidateRestoredState: func(ctx context.Context, tx pgx.Tx) error {
			var version int
			if err := tx.QueryRow(ctx, `SELECT state_version FROM extension_state_metadata WHERE profile_id='network_flow_activity'`).Scan(&version); err != nil {
				return err
			}
			if version != 5 {
				return errors.New("owner rejected restored state")
			}
			return nil
		}}}}}
		for _, version := range []int{4, 5} {
			err := bound.WithAtomicRestore(t.Context(), nil, func(m VNextRestoreMutation) error {
				_, err := m.(*vNextRestoreMutation).tx.Exec(t.Context(), `INSERT INTO extension_state_metadata(profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at) VALUES ('network_flow_activity','network_flow_activity.state_v1',$1,1,now(),now())`, version)
				return err
			})
			if (err == nil) != (version == 5) {
				t.Fatalf("restore version %d outcome: %v", version, err)
			}
			var count int
			if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM extension_state_metadata WHERE profile_id='network_flow_activity'`).Scan(&count); err != nil || count != map[int]int{4: 0, 5: 1}[version] {
				t.Fatal("restore admission transaction outcome", count, err)
			}
		}
		if _, err := pool.Exec(t.Context(), `DELETE FROM extension_state_metadata WHERE profile_id='network_flow_activity'`); err != nil {
			t.Fatal(err)
		}
	})

	err = target.WithAtomicRestore(t.Context(), nil, func(m VNextRestoreMutation) error {
		tx := m.(*vNextRestoreMutation).tx
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
	if !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("orphan restore accepted: %v", err)
	}
	var count int
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM public.incidents`).Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected restore published rows", count, err)
	}
	// All relational FKs and rebuilt claims are valid; the bypassed JSON support
	// reference must still fail the Indicator owner's restored-source validator.
	err = target.WithAtomicRestore(t.Context(), nil, func(m VNextRestoreMutation) error {
		tx := m.(*vNextRestoreMutation).tx
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
	if !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("orphan support restore accepted: %v", err)
	}
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM public.records`).Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected support restore published rows", count, err)
	}

}
