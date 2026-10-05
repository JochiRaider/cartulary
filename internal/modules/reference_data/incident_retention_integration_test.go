package reference_data

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIncidentReferencesRetainLocalVersionsAndUnavailableHistory_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-pack-portability")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	storage := &coordinatorMemoryStorage{objects: map[string][]byte{}}
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := ReconcileBaseRelease(ctx, pool, storage, BaseReleaseOptions{Limits: DefaultLimits()}, at); err != nil {
		t.Fatal(err)
	}
	actor := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Portability fixture','hash',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	t.Run("durable destination cohort admission", func(t *testing.T) { testPortablePreparationAdmission(t, pool, storage, actor, at) })
	createIncident := func() uuid.UUID {
		t.Helper()
		id := uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Portability fixture','active',$3,$3)`, id, id.String(), actor); err != nil {
			t.Fatal(err)
		}
		return id
	}
	repository := canonicalRepository{pool: pool, storage: storage}
	current, err := repository.CurrentSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	retained := retention{repository: &repository}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := retained.ExportReferencesTx(ctx, tx, []string{current.ID})
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Rollback(ctx)
	port, err := newTestIncidentReferences(pool, storage)
	if err != nil {
		t.Fatal(err)
	}
	binding := SetBinding{SetID: current.ID, SHA256: current.SHA256, Provenance: []PackProvenance{}}
	for _, member := range current.Members {
		provenance, err := repository.Provenance(ctx, current.ID, member.Key)
		if err != nil {
			t.Fatal(err)
		}
		binding.Provenance = append(binding.Provenance, provenance)
	}
	t.Run("native owner bindings are exact and deduplicated", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		native, err := port.ExportTx(ctx, tx, uuid.New(), []SetBinding{binding, binding})
		if err != nil || !bytes.Equal(native, catalog) {
			t.Fatal("native reference export mismatch", err)
		}
		altered := binding
		altered.SHA256 = strings.Repeat("f", 64)
		if _, err := port.ExportTx(ctx, tx, uuid.New(), []SetBinding{altered}); err == nil {
			t.Fatal("altered native binding admitted")
		}
	})
	incident := createIncident()
	operation := uuid.New()
	catalogRefs, err := DecodeIncidentBundleReferences(catalog)
	if err != nil {
		t.Fatal(err)
	}
	contentManifest := packformat.EmptyPortableContent()
	contentManifest.RequiredMembers = []packformat.PortableRequiredMember{{SetID: catalogRefs.Sets[0].ID, Key: catalogRefs.Sets[0].Members[0].Key}}
	contentBytes, err := packformat.EncodePortableContent(contentManifest, catalogRefs.format())
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := port.PrepareImport(ctx, IncidentReferenceImportRequest{References: catalog, ContentManifest: &contentBytes, IncidentID: incident, OperationID: operation, ActorID: actor, At: at})
	if err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := port.ApplyImportTx(ctx, tx, prepared); err != nil {
		t.Fatal(err)
	}
	// A partial set's version pins must serialize with removal just as full-set
	// pins do. The competing mutation cannot acquire a key until publication.
	competing, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer competing.Rollback(ctx)
	if _, err := competing.Exec(ctx, `SET LOCAL lock_timeout='100ms'`); err != nil {
		t.Fatal(err)
	}
	_, err = competing.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key='type_registry.host' FOR UPDATE`)
	var locked *pgconn.PgError
	if !errors.As(err, &locked) || locked.Code != "55P03" {
		t.Fatal("pin did not guard removal", err)
	}
	_ = competing.Rollback(ctx)
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := repository.CurrentSet(ctx)
	if err != nil || after.ID != current.ID {
		t.Fatal("portability changed activation", err)
	}
	var pins int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_version_pins WHERE owner_id=$1`, incident.String()).Scan(&pins); err != nil || pins != 3 {
		t.Fatal("complete catalog pins", pins, err)
	}
	if err := ValidateRequiredState(ctx, pool, storage, DefaultLimits()); err != nil {
		t.Fatal("retained catalog readiness", err)
	}
	if err := RestoreHistoricalState(ctx, pool, storage, DefaultLimits()); err != nil {
		t.Fatal("retained catalog restore", err)
	}
	// A proven committed identity replays before later revisions are consulted.
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.host'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := port.ApplyImportTx(ctx, tx, prepared); err != nil {
		t.Fatal("retention replay", err)
	}
	exported, err := port.ExportTx(ctx, tx, incident, nil)
	if err != nil || !bytes.Equal(exported, catalog) {
		t.Fatal("reference catalog changed on reexport", err)
	}
	control, err := port.ExportContentTx(ctx, tx, IncidentReferenceExportRequest{IncidentID: incident, References: catalog})
	if err != nil || !bytes.Equal(control, contentBytes) {
		t.Fatal("required membership changed on refs-only reexport", err)
	}
	_ = tx.Rollback(ctx)
	retainedObjects := storage.objects
	storage.objects = map[string][]byte{}
	replayed, replayErr := port.PrepareImport(ctx, IncidentReferenceImportRequest{References: catalog, ContentManifest: &contentBytes, IncidentID: incident, OperationID: operation, ActorID: actor, At: at})
	storage.objects = retainedObjects
	if replayErr != nil || replayed == nil || !replayed.replay {
		t.Fatal("committed replay consulted current storage", replayErr)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := port.ApplyImportTx(ctx, tx, replayed); err != nil {
		t.Fatal("reprepared replay", err)
	}
	_ = tx.Rollback(ctx)
	if _, err := port.PrepareImport(ctx, IncidentReferenceImportRequest{References: catalog, IncidentID: incident, OperationID: operation, ActorID: actor, At: at}); err == nil {
		t.Fatal("replay erased required membership")
	}
	// One absent member cannot erase the other destination-trusted members or
	// the exact unresolved source tuple. It does not create a partial set ID.
	refs, err := DecodeIncidentBundleReferences(catalog)
	if err != nil {
		t.Fatal(err)
	}
	refs.Versions[0].Version = "unavailable.source.version"
	refs.Sets[0].Members[0].Version = refs.Versions[0].Version
	refs.Sets[0], err = buildPackSet(refs.Sets[0].Members)
	if err != nil {
		t.Fatal(err)
	}
	partial, err := EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	partialIncident := createIncident()
	requiredPartial := packformat.EmptyPortableContent()
	requiredPartial.RequiredMembers = []packformat.PortableRequiredMember{{SetID: refs.Sets[0].ID, Key: refs.Versions[0].Key}}
	requiredBytes, err := packformat.EncodePortableContent(requiredPartial, refs.format())
	if err != nil {
		t.Fatal(err)
	}
	requiredOperation := uuid.New()
	_, err = port.PrepareImport(ctx, IncidentReferenceImportRequest{References: partial, ContentManifest: &requiredBytes, IncidentID: partialIncident, OperationID: requiredOperation, ActorID: actor, At: at})
	if invariant, ok := IncidentBundleReferenceInvariant(err); !ok || invariant != IncidentBundleReferenceDegradationInvariant {
		t.Fatal("required absent member became optional", err)
	}
	var admitted bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operations WHERE operation_id=$1)`, uuid.NewSHA1(requiredOperation, []byte("reference_pack:incident_retention"))).Scan(&admitted); err != nil || admitted {
		t.Fatal("rejected required import mutated owner state", err)
	}
	p, err := port.PrepareImport(ctx, IncidentReferenceImportRequest{References: partial, IncidentID: partialIncident, OperationID: uuid.New(), ActorID: actor, At: at})
	if err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := port.ApplyImportTx(ctx, tx, p); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM reference_pack_version_pins WHERE owner_id=$1`, partialIncident.String()).Scan(&pins); err != nil || pins != 2 {
		t.Fatal("partial catalog lost local pins", pins, err)
	}
	exported, err = port.ExportTx(ctx, tx, partialIncident, []SetBinding{binding})
	if err != nil {
		t.Fatal(err)
	}
	merged, err := DecodeIncidentBundleReferences(exported)
	if err != nil || len(merged.Sets) != 2 || len(merged.Versions) != 4 {
		t.Fatal("native and imported references were not combined", err)
	}
	found := false
	for _, v := range merged.Versions {
		if v.Version == "unavailable.source.version" {
			found = true
		}
	}
	if !found {
		t.Fatal("unavailable reference was dropped")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRequiredState(ctx, pool, storage, DefaultLimits()); err != nil {
		t.Fatal("optional absence blocked readiness", err)
	}
	// Same tuple with a different digest is a collision, not optional absence.
	refs, err = DecodeIncidentBundleReferences(catalog)
	if err != nil {
		t.Fatal(err)
	}
	refs.Versions[0].PayloadSHA256 = strings.Repeat("f", 64)
	refs.Sets[0].Members[0].PayloadSHA256 = refs.Versions[0].PayloadSHA256
	refs.Sets[0], err = buildPackSet(refs.Sets[0].Members)
	if err != nil {
		t.Fatal(err)
	}
	collision, err := EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	_, err = port.PrepareImport(ctx, IncidentReferenceImportRequest{References: collision, IncidentID: uuid.New(), OperationID: uuid.New(), ActorID: actor, At: at})
	if invariant, ok := IncidentBundleReferenceInvariant(err); !ok || invariant != IncidentBundleReferenceIdentityInvariant {
		t.Fatal("collision became absence", err)
	}
	staleIncident := createIncident()
	p, err = port.PrepareImport(ctx, IncidentReferenceImportRequest{References: catalog, IncidentID: staleIncident, OperationID: uuid.New(), ActorID: actor, At: at})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.host'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var rejected *OperationRejection
	if err := port.ApplyImportTx(ctx, tx, p); !errors.As(err, &rejected) || rejected.Reason != "stale_admission_state" {
		t.Fatal("stale retention published", err)
	}
	_ = tx.Rollback(ctx)
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_portable_catalogs WHERE incident_id=$1)`, staleIncident).Scan(&exists); err != nil || exists {
		t.Fatal("failed retention left semantic state", err)
	}
	// Tampering with a retained resolution is refused before serving restored
	// content, even though its referenced bytes and ordinary set pins are valid.
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `ALTER TABLE reference_pack_portable_catalogs DISABLE TRIGGER reference_pack_portable_catalogs_immutable`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_portable_catalogs SET canonical_resolution='{}'::bytea WHERE incident_id=$1`, incident); err != nil {
		t.Fatal(err)
	}
	if err := validatePortableCatalogs(ctx, tx); !errors.Is(err, errHistoricalIntegrity) {
		t.Fatal("altered catalog was accepted", err)
	}
}

func newTestIncidentReferences(pool *pgxpool.Pool, storage ArtifactStorage) (IncidentReferences, error) {
	catalog := collaborationsupport.NewJobCatalog()
	transactions := collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, JobTerminalEffects{})
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: time.Now})
	if err != nil {
		return nil, err
	}
	return NewIncidentReferences(IncidentReferenceOptions{Postgres: pool, Storage: storage, Limits: DefaultLimits(), JobExecutions: transactions, JobOperations: manager, Now: time.Now})
}
