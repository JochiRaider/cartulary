package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestBaseReconciliationAtomicPublication_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	for _, mode := range []string{"prepared", "operation", "connection", "index", "before commit", "lost acknowledgement", "unreadable proof"} {
		t.Run(mode, func(t *testing.T) {
			db := harness.PrepareIsolatedDatabaseT(t, "reference-base-atomic")
			pool, err := pgxpool.New(t.Context(), db.DSN)
			if err != nil {
				t.Fatal(err)
			}
			defer pool.Close()
			faults := &baseFaultDB{DB: pool, mode: mode}
			storage := &coordinatorMemoryStorage{objects: map[string][]byte{}}
			before := baseAtomicState(t, pool)
			var fatal error
			owner, err := extensionstore.NewOwnerMutationFinalizer(func(err error) { fatal = err })
			if err != nil {
				t.Fatal(err)
			}
			options := BaseReleaseOptions{Finalizer: coordinatorTestFinalizer{owner: owner}, ProfileClaimed: true, Limits: DefaultLimits()}
			at := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
			err = ReconcileBaseRelease(t.Context(), faults, storage, options, at)
			if !faults.hit {
				t.Fatal("fault not reached", mode)
			}
			committed := mode == "lost acknowledgement" || mode == "unreadable proof"
			if mode == "lost acknowledgement" {
				if err != nil {
					t.Fatal("exact receipt failed to prove commit", err)
				}
			} else if err == nil {
				t.Fatal("fault admitted readiness")
			}
			if (fatal != nil) != (mode == "unreadable proof") {
				t.Fatal("wrong integrity disposition", fatal, err)
			}
			if !committed && !reflect.DeepEqual(before, baseAtomicState(t, pool)) {
				t.Fatal("aborted publication left database effects")
			}
			var unfinished int
			if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM reference_pack_operations WHERE terminal_at IS NULL`).Scan(&unfinished); err != nil || unfinished != 0 {
				t.Fatal("unfinished reconciliation survived", unfinished, err)
			}
			if committed && len(storage.objects) == 0 {
				t.Fatal("uncertain commit deleted published bytes")
			}
			if err := contractadmission.Validate(t.Context(), pool); err != nil {
				t.Fatal("restart preflight", err)
			}
			options.Finalizer = newBaseTestFinalizer(t)
			options.ProfileClaimed = false
			if err := ReconcileBaseRelease(t.Context(), pool, storage, options, at.Add(time.Minute)); err != nil {
				t.Fatal("fresh startup after interruption", err)
			}
			if err := ValidateRequiredState(t.Context(), pool, storage, DefaultLimits(), testVerificationLifetime(t)); err != nil {
				t.Fatal("restart readiness", err)
			}
			if err := CollectUnreferencedObjects(t.Context(), pool, storage); err != nil {
				t.Fatal("post-recovery collection", err)
			}
			if err := ValidateRequiredState(t.Context(), pool, storage, DefaultLimits(), testVerificationLifetime(t)); err != nil {
				t.Fatal("collection removed retained bytes", err)
			}
		})
	}
	t.Run("existing pinned state", testBaseInterruptionWithPinnedState)
}

func testBaseInterruptionWithPinnedState(t *testing.T) {
	for _, claimed := range []bool{true, false} {
		profile := "unclaimed"
		if claimed {
			profile = "claimed"
		}
		t.Run(profile, func(t *testing.T) {
			for _, mode := range []string{"operation", "connection", "before commit", "lost acknowledgement", "unreadable proof"} {
				t.Run(mode, func(t *testing.T) {
					f := newCanonicalCoordinatorFixture(t, "reference-base-pinned-interruption")
					ctx := t.Context()
					pin := func() PackSet {
						t.Helper()
						tx, err := f.pool.Begin(ctx)
						if err != nil {
							t.Fatal(err)
						}
						defer tx.Rollback(ctx)
						set, err := PinCurrentTx(ctx, tx, "test_snapshot", uuid.NewString(), uuid.New())
						if err != nil {
							t.Fatal(err)
						}
						if err := tx.Commit(ctx); err != nil {
							t.Fatal(err)
						}
						return set
					}
					previous := pin()
					accepted, err := f.coordinator.Import(ctx, f.actor, "pinned-import", bytes.NewReader(f.container))
					if err != nil {
						t.Fatal(err)
					}
					jobID := uuid.MustParse(accepted.Job.JobID)
					execution, leased, err := f.manager.Claim(ctx, jobID)
					if err != nil || !leased {
						t.Fatal("claim", err)
					}
					if err := f.coordinator.Execute(ctx, execution); err != nil {
						t.Fatal(err)
					}
					job, err := f.manager.Get(ctx, jobID)
					if err != nil || job.Status != jobs.StatusSucceeded {
						t.Fatal("import", job.Status, err)
					}
					request, apiErr := decodeActionRequest(strings.NewReader(`{"client_txn_id":"pinned-activate"}`), false)
					if apiErr != nil {
						t.Fatal(apiErr)
					}
					if _, err := f.coordinator.Activate(ctx, ActionParams{ActorUserID: f.actor, PackKey: "type_registry.host", PackVersion: "signed-fixture.1", Request: request, Now: *f.now}); err != nil {
						t.Fatal(err)
					}
					active := pin()
					if active.ID == previous.ID {
						t.Fatal("fixture did not change selection")
					}
					before := baseAtomicState(t, f.pool)
					objects := map[string][]byte{}
					for ref, data := range f.storage.objects {
						objects[ref] = bytes.Clone(data)
					}
					faults := &baseFaultDB{DB: f.pool, mode: mode}
					var fatal error
					owner, err := extensionstore.NewOwnerMutationFinalizer(func(err error) { fatal = err })
					if err != nil {
						t.Fatal(err)
					}
					options := BaseReleaseOptions{Finalizer: coordinatorTestFinalizer{owner: owner}, ProfileClaimed: claimed, Limits: DefaultLimits()}
					err = ReconcileBaseRelease(ctx, faults, f.storage, options, f.now.Add(time.Minute))
					if !faults.hit {
						t.Fatal("fault not reached", mode)
					}
					if mode == "lost acknowledgement" {
						if err != nil {
							t.Fatal("committed receipt not proved", err)
						}
					} else if err == nil {
						t.Fatal("interrupted publication admitted readiness")
					}
					if (fatal != nil) != (mode == "unreadable proof") {
						t.Fatal("integrity disposition", fatal, err)
					}
					committed := mode == "lost acknowledgement" || mode == "unreadable proof"
					after := baseAtomicState(t, f.pool)
					if !committed && !reflect.DeepEqual(before, after) {
						t.Fatal("rollback changed existing pointer, revisions, pins or retained rows")
					}
					for _, table := range []string{"reference_pack_sets", "reference_pack_set_members", "reference_pack_pins", "reference_pack_versions", "reference_pack_envelopes", "reference_pack_index_generations", "reference_pack_indexes", "reference_pack_lookup_keys", "reference_pack_objects", "reference_pack_object_refs", "reference_pack_release_bindings"} {
						if before[table] != after[table] {
							t.Fatal("existing immutable state changed", table)
						}
					}
					if !reflect.DeepEqual(objects, f.storage.objects) {
						t.Fatal("interruption changed retained bytes")
					}
					repository := &canonicalRepository{pool: f.pool, storage: f.storage}
					set, err := repository.CurrentSet(ctx)
					want := active.ID
					if committed && !claimed {
						want = previous.ID
					}
					if err != nil || set.ID != want {
						t.Fatal("wrong authoritative selection", set.ID, want, err)
					}
					if committed && !claimed && before["reference_pack_current_set"] == after["reference_pack_current_set"] {
						t.Fatal("successful unclaimed reconciliation did not change pointer and revision")
					}
					var unfinished int
					if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operations WHERE terminal_at IS NULL`).Scan(&unfinished); err != nil || unfinished != 0 {
						t.Fatal("unfinished operation survived", unfinished, err)
					}
					options.Finalizer = newBaseTestFinalizer(t)
					if err := contractadmission.Validate(ctx, f.pool); err != nil {
						t.Fatal("fresh admission", err)
					}
					if err := ReconcileBaseRelease(ctx, f.pool, f.storage, options, f.now.Add(2*time.Minute)); err != nil {
						t.Fatal("fresh reconciliation", err)
					}
					if err := CollectUnreferencedObjects(ctx, f.pool, f.storage); err != nil {
						t.Fatal(err)
					}
					if err := ValidateRequiredState(ctx, f.pool, f.storage, DefaultLimits(), testVerificationLifetime(t)); err != nil {
						t.Fatal("restart readiness", err)
					}
					final := baseAtomicState(t, f.pool)
					for _, table := range []string{"reference_pack_sets", "reference_pack_set_members", "reference_pack_pins"} {
						if before[table] != final[table] {
							t.Fatal("restart altered pinned history", table)
						}
					}
					for _, pinned := range []PackSet{previous, active} {
						for _, member := range pinned.Members {
							if _, err := repository.Provenance(ctx, pinned.ID, member.Key); err != nil {
								t.Fatal("pinned provenance unavailable", err)
							}
						}
					}
				})
			}
		})
	}
}

// Capture full retained row values, including existing pointer/revision and pin
// provenance. Counts alone cannot detect updates that escape a failed attempt.
func baseAtomicState(t testing.TB, db dbport.DB) map[string]string {
	t.Helper()
	state := map[string]string{}
	for _, table := range []string{"reference_pack_operations", "reference_pack_index_generations", "reference_pack_indexes", "reference_pack_lookup_keys", "reference_pack_versions", "reference_pack_candidates", "reference_pack_envelopes", "reference_pack_release_bindings", "reference_pack_sets", "reference_pack_set_members", "reference_pack_pins", "reference_pack_current_set", "reference_pack_events", "reference_pack_objects", "reference_pack_object_refs", "reference_pack_key_state", "reference_pack_registry_usage"} {
		var value string
		if err := db.QueryRow(t.Context(), `SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM `+pgx.Identifier{table}.Sanitize()+` t`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		state[table] = value
	}
	return state
}

type baseFaultDB struct {
	dbport.DB
	mode           string
	hit, committed bool
}

func (db *baseFaultDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	if options.AccessMode == pgx.ReadOnly {
		return db.DB.BeginTx(ctx, options)
	}
	if db.mode == "prepared" {
		db.hit = true
		return nil, errors.New("injected after byte preparation")
	}
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &baseFaultTx{Tx: tx, db: db}, nil
}
func (db *baseFaultDB) QueryRow(ctx context.Context, q string, args ...any) pgx.Row {
	if db.mode == "unreadable proof" && db.committed && strings.Contains(q, "frozen_input=$2") {
		return baseFaultRow{io.ErrUnexpectedEOF}
	}
	return db.DB.QueryRow(ctx, q, args...)
}

type baseFaultRow struct{ err error }

func (r baseFaultRow) Scan(...any) error { return r.err }

type baseFaultTx struct {
	pgx.Tx
	db *baseFaultDB
}

func (tx *baseFaultTx) Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error) {
	tag, err := tx.Tx.Exec(ctx, q, args...)
	if err != nil {
		return tag, err
	}
	if strings.HasPrefix(q, "INSERT INTO reference_pack_operations") && (tx.db.mode == "operation" || tx.db.mode == "connection") {
		tx.db.hit = true
		if tx.db.mode == "connection" {
			_ = tx.Tx.Conn().Close(context.WithoutCancel(ctx))
		}
		return tag, io.ErrUnexpectedEOF
	}
	if strings.HasPrefix(q, "UPDATE reference_pack_index_generations") && tx.db.mode == "index" {
		tx.db.hit = true
		return tag, io.ErrUnexpectedEOF
	}
	return tag, nil
}
func (tx *baseFaultTx) Commit(ctx context.Context) error {
	if tx.db.mode == "before commit" {
		tx.db.hit = true
		return io.ErrUnexpectedEOF
	}
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	tx.db.committed = true
	if tx.db.mode == "lost acknowledgement" || tx.db.mode == "unreadable proof" {
		tx.db.hit = true
		return io.ErrUnexpectedEOF
	}
	return nil
}

func TestCurrentRetainedStructureAcrossConsumersAndRecovery_Integration(t *testing.T) {
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-current-structure")
	pool, err := pgxpool.New(t.Context(), db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	storage := &coordinatorMemoryStorage{objects: map[string][]byte{}}
	if err := ReconcileBaseRelease(t.Context(), pool, storage, BaseReleaseOptions{Finalizer: newBaseTestFinalizer(t), Limits: DefaultLimits()}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"unknown", "noncanonical", "unordered", "identity"} {
		t.Run(mode, func(t *testing.T) {
			tx, err := pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(t.Context())
			var original, provenance []byte
			var originalID, originalDigest string
			if err := tx.QueryRow(t.Context(), `SELECT canonical_set,canonical_provenance,pack_set_id,pack_set_sha256 FROM reference_pack_sets`).Scan(&original, &provenance, &originalID, &originalDigest); err != nil {
				t.Fatal(err)
			}
			var value map[string]any
			if err := json.Unmarshal(original, &value); err != nil {
				t.Fatal(err)
			}
			if mode == "unknown" {
				value["extra"] = true
			}
			if mode == "unordered" {
				members := value["members"].([]any)
				members[0], members[1] = members[1], members[0]
			}
			raw, err := canonicaljson.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "noncanonical" {
				raw = append(raw, '\n')
			}
			if mode == "identity" {
				other, err := packformat.DecodeSet(original)
				if err != nil {
					t.Fatal(err)
				}
				other.Members[0].Version = "different-valid-set.1"
				other, err = packformat.BuildSet(other.Members)
				if err != nil {
					t.Fatal(err)
				}
				raw, err = canonicaljson.Marshal(other)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := packformat.DecodeSet(raw); err != nil || other.ID == originalID {
					t.Fatal("identity fixture must be an internally valid different set", err)
				}
			}
			if _, err := tx.Exec(t.Context(), `ALTER TABLE reference_pack_sets DISABLE TRIGGER USER`); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `UPDATE reference_pack_sets SET canonical_set=$1`, raw); err != nil {
				t.Fatal(err)
			}
			reader := &canonicalRepository{pool: tx, storage: storage}
			before := baseAtomicState(t, baseTestTxDB{tx})
			for _, read := range []func() error{
				func() error { _, err := reader.CurrentSet(t.Context()); return err },
				func() error { _, err := reader.RetainedSet(t.Context(), originalID); return err },
				func() error { _, err := reader.Provenance(t.Context(), originalID, "type_registry.host"); return err },
				func() error {
					_, err := PinCurrentTx(t.Context(), tx, "test_snapshot", uuid.NewString(), uuid.New())
					return err
				},
				func() error { _, _, err := loadSelectionTx(t.Context(), tx, originalID); return err },
			} {
				var unavailable *ConsumerError
				if err := read(); !errors.As(err, &unavailable) || unavailable.Code != "pack_unavailable" {
					t.Fatal("consumer admitted malformed or misbound retained set", err)
				}
			}
			binding, err := canonicaljson.Marshal(map[string]any{"pack_set_id": originalID, "pack_set_sha256": originalDigest, "provenance": json.RawMessage(provenance)})
			if err != nil {
				t.Fatal(err)
			}
			if err := contractadmission.ValidateBinding(t.Context(), tx, binding); !contractadmission.IsIncompatible(err) {
				t.Fatal("Reporting binding admitted malformed or misbound local set", err)
			}
			for _, check := range []func(context.Context, dbport.DB, ArtifactStorage, Limits, *VerificationLifetime) error{ValidateRequiredState, RestoreHistoricalState} {
				if err := check(t.Context(), baseTestTxDB{tx}, storage, DefaultLimits(), testVerificationLifetime(t)); !contractadmission.IsIncompatible(err) {
					t.Fatal("recovery failed to reject malformed set before rebuild", err)
				}
			}
			if err := ReconcileBaseRelease(t.Context(), baseTestTxDB{tx}, storage, BaseReleaseOptions{Finalizer: newBaseTestFinalizer(t), Limits: DefaultLimits()}, time.Now().UTC()); !contractadmission.IsIncompatible(err) {
				t.Fatal("reconciliation admitted malformed or misbound set", err)
			}
			if !reflect.DeepEqual(before, baseAtomicState(t, baseTestTxDB{tx})) {
				t.Fatal("rejected retained binding mutated rows or rebuilt indexes")
			}
		})
	}
	for _, authority := range []string{"", "framework_reference", "advisory_enrichment"} {
		t.Run("portable-only reporting binding/"+authority, func(t *testing.T) {
			tx, err := pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(t.Context())
			var raw, anchors []byte
			if err := tx.QueryRow(t.Context(), `SELECT canonical_set,canonical_provenance FROM reference_pack_sets`).Scan(&raw, &anchors); err != nil {
				t.Fatal(err)
			}
			set, err := packformat.DecodeSet(raw)
			if err != nil {
				t.Fatal(err)
			}
			for i := range set.Members {
				set.Members[i].Version = "portable-only.1"
			}
			if authority != "" {
				member := set.Members[0]
				member.Key, member.ProfileID, member.ProfileVersion = "zz.unregistered", "cartulary.reference_pack.dataset.v1", "1"
				set.Members = append(set.Members, member)
			}
			set, err = packformat.BuildSet(set.Members)
			if err != nil {
				t.Fatal(err)
			}
			var provenance []map[string]any
			if err := json.Unmarshal(anchors, &provenance); err != nil {
				t.Fatal(err)
			}
			if authority != "" {
				anchor := map[string]any{}
				for key, value := range provenance[0] {
					anchor[key] = value
				}
				anchor["pack_key"], anchor["content_profile_id"], anchor["content_profile_version"], anchor["authority_class"] = "zz.unregistered", "cartulary.reference_pack.dataset.v1", "1", authority
				provenance = append(provenance, anchor)
			}
			versions := []packformat.PortableVersion{}
			for i := range provenance {
				provenance[i]["pack_set_id"] = set.ID
				provenance[i]["pack_version"] = "portable-only.1"
				versions = append(versions, packformat.PortableVersion{SetMember: set.Members[i], DistributionKind: "packaged_builtin", VerificationMethod: "packaged_release_manifest_v1", SourceProfileID: provenance[i]["source_profile_id"].(string), SourceProfileSHA256: provenance[i]["source_profile_sha256"].(string)})
			}
			catalog, err := packformat.EncodePortableReferences([]packformat.Set{set}, versions)
			if err != nil {
				t.Fatal(err)
			}
			catalog, err = canonicaljson.Canonicalize(catalog)
			if err != nil {
				t.Fatal(err)
			}
			id := uuid.New()
			for _, query := range []string{
				`INSERT INTO users(id,email,display_name,password_hash) VALUES($1,'portable-only@example.test','Portable-only fixture','unused')`,
				`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'PORTABLE','portable','Portable-only fixture','active',$1,$1)`,
				`INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,terminal_at,frozen_input,final_outcome) VALUES($1,'reconcile','system',now(),now(),'{}','{}')`,
			} {
				if _, err := tx.Exec(t.Context(), query, id); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := tx.Exec(t.Context(), `INSERT INTO reference_pack_portable_catalogs(incident_id,operation_id,catalog_sha256,canonical_references,canonical_resolution) VALUES($1,$1,$2,$3,'{}')`, id, packformat.Digest(catalog), catalog); err != nil {
				t.Fatal(err)
			}
			binding, err := canonicaljson.Marshal(map[string]any{"pack_set_id": set.ID, "pack_set_sha256": set.SHA256, "provenance": provenance})
			if err != nil {
				t.Fatal(err)
			}
			if err := contractadmission.ValidateBinding(t.Context(), tx, binding); err != nil {
				t.Fatal("valid unavailable portable binding rejected", err)
			}
			if authority != "" {
				provenance[len(provenance)-1]["authority_class"] = "registry_authoritative"
				invalid, err := canonicaljson.Marshal(map[string]any{"pack_set_id": set.ID, "pack_set_sha256": set.SHA256, "provenance": provenance})
				if err != nil {
					t.Fatal(err)
				}
				if err := contractadmission.ValidateBinding(t.Context(), tx, invalid); !contractadmission.IsIncompatible(err) {
					t.Fatal("portable registry escalation", err)
				}
			}
			if _, err := tx.Exec(t.Context(), `ALTER TABLE reference_pack_portable_catalogs DISABLE TRIGGER USER`); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `UPDATE reference_pack_portable_catalogs SET canonical_references=$1,catalog_sha256=encode(sha256($1),'hex')`, append(catalog, '\n')); err != nil {
				t.Fatal(err)
			}
			if err := contractadmission.ValidateBinding(t.Context(), tx, binding); !contractadmission.IsIncompatible(err) {
				t.Fatal("noncanonical portable catalog admitted", err)
			}
		})
	}
}

type baseTestTxDB struct{ pgx.Tx }

func (db baseTestTxDB) BeginTx(ctx context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	return db.Begin(ctx)
}
