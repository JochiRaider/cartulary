package reference_data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type coordinatorMemoryStorage struct {
	engineStorage
	objects map[string][]byte
	staged  map[string][]byte
}

func (s *coordinatorMemoryStorage) PublishStream(_ context.Context, digest string, size int64, r io.Reader) (StorageRef, io.Closer, error) {
	data, err := io.ReadAll(io.LimitReader(r, size+1))
	if err != nil {
		return StorageRef{}, nil, err
	}
	if int64(len(data)) != size || packformat.Digest(data) != digest {
		return StorageRef{}, nil, errors.New("fixture byte mismatch")
	}
	ref, _ := ParseStorageRef("objects/" + uuid.NewString())
	s.objects[ref.String()] = data
	return ref, io.NopCloser(bytes.NewReader(nil)), nil
}
func (s *coordinatorMemoryStorage) OpenPublished(_ context.Context, ref StorageRef) (ContainerReader, int64, error) {
	data, ok := s.objects[ref.String()]
	if !ok {
		return nil, 0, fs.ErrNotExist
	}
	return engineContainer{bytes.NewReader(data)}, int64(len(data)), nil
}
func (s *coordinatorMemoryStorage) StageStream(_ context.Context, r io.Reader, limit int64) (StagingRef, string, int64, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return StagingRef{}, "", 0, err
	}
	if int64(len(data)) > limit {
		return StagingRef{}, "", 0, errors.New("fixture staging limit")
	}
	if s.staged == nil {
		s.staged = map[string][]byte{}
	}
	ref, _ := ParseStagingRef("staged/" + uuid.NewString())
	s.staged[ref.String()] = data
	return ref, packformat.Digest(data), int64(len(data)), nil
}
func (s *coordinatorMemoryStorage) OpenStaged(ctx context.Context, ref StagingRef) (ContainerReader, int64, error) {
	if s.staged == nil {
		return s.engineStorage.OpenStaged(ctx, ref)
	}
	data, ok := s.staged[ref.String()]
	if !ok {
		return nil, 0, fs.ErrNotExist
	}
	return engineContainer{bytes.NewReader(data)}, int64(len(data)), nil
}
func (s *coordinatorMemoryStorage) RemoveStaged(ref StagingRef) error {
	delete(s.staged, ref.String())
	return nil
}
func (s *coordinatorMemoryStorage) RemovePublished(ref StorageRef) error {
	delete(s.objects, ref.String())
	return nil
}

func (s *coordinatorMemoryStorage) WithPublishedCollection(ctx context.Context, run func(PublishedObjects) error) error {
	return run(memoryPublishedObjects{s})
}

type memoryPublishedObjects struct{ storage *coordinatorMemoryStorage }

func (p memoryPublishedObjects) Visit(ctx context.Context, visit func(StorageRef) error) error {
	for key := range p.storage.objects {
		ref, err := ParseStorageRef(key)
		if err != nil {
			return err
		}
		if err := visit(ref); err != nil {
			return err
		}
	}
	return ctx.Err()
}
func (p memoryPublishedObjects) Remove(ref StorageRef) error { return p.storage.RemovePublished(ref) }

func TestCanonicalAdmissionFreezesCohortAndRejectsRelevantStaleState_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-pack-admission")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	storage := &coordinatorMemoryStorage{objects: map[string][]byte{}}
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	options := BaseReleaseOptions{ProfileClaimed: true, Limits: DefaultLimits()}
	if err := ReconcileBaseRelease(ctx, pool, storage, options, at); err != nil {
		t.Fatal(err)
	}
	op := operationAdmission{ID: uuid.New(), Kind: "refresh", ActorKind: "system", At: at, Keys: []string{"type_registry.host"}, TimeoutSeconds: 1800}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := admitOperationTx(ctx, tx, op)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_members WHERE operation_id=$1`, op.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("Base entered refresh cohort", err)
	}
	// A never-successful version added after admission cannot join the cohort.
	if _, err := pool.Exec(ctx, `INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at) VALUES('type_registry.host','later','operator_imported','failed',$1)`, at); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = admitOperationTx(ctx, tx, operationAdmission{ID: uuid.New(), Kind: "reverify", ActorKind: "system", At: at, Keys: op.Keys, Version: "later", TimeoutSeconds: 1800})
	_ = tx.Rollback(ctx)
	var rejected *OperationRejection
	if !errors.As(err, &rejected) || rejected.Reason != "verification_pending" {
		t.Fatalf("pending did not precede absent success: %v", err)
	}
	// Changes outside captured keys do not reject publication.
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.evidence'`); err != nil {
		t.Fatal(err)
	}
	// Existing-record usage is not an input to verification or safety fallback.
	// Even assignments under this selected key may change without a stale abort.
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_registry_usage SET revision=revision+1 WHERE pack_key='type_registry.host'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = lockPublicationTx(ctx, tx, op.ID, frozen)
	_ = tx.Rollback(ctx)
	if err != nil {
		t.Fatalf("unrelated mutation rejected: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.host'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = lockPublicationTx(ctx, tx, op.ID, frozen)
	_ = tx.Rollback(ctx)
	if !errors.As(err, &rejected) || rejected.Reason != "stale_admission_state" {
		t.Fatalf("relevant revision silently rebased: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_operation_keys SET admitted_revision=admitted_revision+1 WHERE operation_id=$1`, op.ID); err == nil {
		t.Fatal("frozen admission revision was mutable")
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_members WHERE operation_id=$1`, op.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("later import joined frozen cohort", err)
	}
	// Normal replacement does read assignment usage and must detect a change.
	var builtinVersion string
	if err := pool.QueryRow(ctx, `SELECT pack_version FROM reference_pack_candidates WHERE pack_key='type_registry.evidence' AND distribution_kind='packaged_builtin'`).Scan(&builtinVersion); err != nil {
		t.Fatal(err)
	}
	activation := operationAdmission{ID: uuid.New(), Kind: "activate", ActorKind: "system", At: at, Keys: []string{"type_registry.evidence"}, Version: builtinVersion, TimeoutSeconds: 1800}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	activationInput, err := admitOperationTx(ctx, tx, activation)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_registry_usage SET revision=revision+1 WHERE pack_key='type_registry.evidence'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = lockPublicationTx(ctx, tx, activation.ID, activationInput)
	_ = tx.Rollback(ctx)
	if !errors.As(err, &rejected) || rejected.Reason != "stale_admission_state" {
		t.Fatalf("replacement ignored changed record usage: %v", err)
	}
	// A dependency is a read guard, not selected pending work. The host key
	// already has nonterminal refresh work, but an unrelated optional import
	// may capture it and must reject publication if its revision changes.
	var baseVersion, baseDigest string
	if err := pool.QueryRow(ctx, `SELECT pack_version,payload_sha256 FROM reference_pack_versions WHERE pack_key='type_registry.host' LIMIT 1`).Scan(&baseVersion, &baseDigest); err != nil {
		t.Fatal(err)
	}
	object := preparedObject{ID: uuid.New(), Digest: packformat.Digest(nil), Size: 0}
	object.Reference, object.publication, err = storage.PublishStream(ctx, object.Digest, 0, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	dependencyOp := operationAdmission{ID: uuid.New(), Kind: "import", ActorKind: "system", At: at, TimeoutSeconds: 1800, Import: &importIdentity{Key: "enrichment.lolbas", Version: "dependency-capture", Dependencies: []packformat.Dependency{{Key: "type_registry.host", Version: baseVersion, SHA256: baseDigest}}}, InputObject: &object, ContainerRef: &object.Reference, ContainerSHA: &object.Digest, ContainerBytes: &object.Size}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dependencyInput, err := admitOperationTx(ctx, tx, dependencyOp)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_dependency_keys WHERE operation_id=$1 AND pack_key='type_registry.host'`, dependencyOp.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("dependency was not captured", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.indicator'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = lockPublicationTx(ctx, tx, dependencyOp.ID, dependencyInput)
	_ = tx.Rollback(ctx)
	if err != nil {
		t.Fatal("unrelated key rejected dependency publication", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key='type_registry.host'`); err != nil {
		t.Fatal(err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = lockPublicationTx(ctx, tx, dependencyOp.ID, dependencyInput)
	_ = tx.Rollback(ctx)
	if !errors.As(err, &rejected) || rejected.Reason != "stale_admission_state" {
		t.Fatal("changed dependency did not reject publication", err)
	}
	// Startup must not revise configuration simply because it ran again.
	var before, after int64
	if err := pool.QueryRow(ctx, `SELECT revision FROM reference_pack_current_set WHERE singleton`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileBaseRelease(ctx, pool, storage, options, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT revision FROM reference_pack_current_set WHERE singleton`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("unchanged startup invalidated frozen configuration")
	}
	options.ClockTrusted = true
	if err := ReconcileBaseRelease(ctx, pool, storage, options, at.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT revision FROM reference_pack_current_set WHERE singleton`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatal("clock policy change did not revise configuration")
	}
}

func TestCanonicalCoordinatorPublishesSignedContentAndAtomicInvalidation_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-pack-coordinator")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var vector struct {
		Bootstrap  string    `json:"bootstrap"`
		Repository string    `json:"repository_id"`
		Container  string    `json:"container_base64"`
		At         time.Time `json:"verification_time"`
		Expiry     time.Time `json:"expected_valid_until"`
	}
	data, err := os.ReadFile("../../../contracts/reference-packs/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &vector); err != nil {
		t.Fatal(err)
	}
	container, err := base64.StdEncoding.DecodeString(vector.Container)
	if err != nil {
		t.Fatal(err)
	}
	root, err := packformat.AdmitBootstrap([]byte(vector.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	if err := ReconcileTrustBootstrap(ctx, pool, TrustBootstrap{repositories: map[string]packformat.TrustSnapshot{vector.Repository: root}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	storage := &coordinatorMemoryStorage{engineStorage: engineStorage{container: container}, objects: map[string][]byte{}}
	if err := ReconcileBaseRelease(ctx, pool, storage, BaseReleaseOptions{ProfileClaimed: true, ClockTrusted: true, Limits: DefaultLimits()}, vector.At); err != nil {
		t.Fatal(err)
	}
	c := &verificationService{referenceDependencies: &referenceDependencies{pool: pool, storage: storage, limits: DefaultLimits(), configuration: Configuration{ClockTrusted: true}, now: func() time.Time { return vector.At }}}
	staged, _ := ParseStagingRef("staged/fixture.zip")
	identity, err := probeContainerIdentity(ctx, storage, staged, packformat.DefaultArchiveLimits())
	if err != nil {
		t.Fatal(err)
	}
	admit := func(kind string, at time.Time) executionAttempt {
		t.Helper()
		op := operationAdmission{ID: uuid.New(), Kind: kind, ActorKind: "system", At: at, Keys: []string{identity.Key}, Version: identity.Version, ClockTrusted: true, TimeoutSeconds: 1800}
		if kind == "import" {
			_, input, err := c.prepareImportInput(ctx, bytes.NewReader(container))
			if err != nil {
				t.Fatal(err)
			}
			op.Import = &identity
			op.ContainerRef = &input.Reference
			op.ContainerSHA = &input.Digest
			op.ContainerBytes = &input.Size
			op.InputObject = &input
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := admitOperationTx(ctx, tx, op); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		tx, err = pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		a, err := beginAttemptTx(ctx, tx, op.ID, at)
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		return a
	}
	prepare := func(a executionAttempt) {
		t.Helper()
		for n := int64(1); n <= a.Count; n++ {
			m, err := c.frozenMember(ctx, a, n)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.prepareMember(ctx, a, m); err != nil {
				t.Fatal(err)
			}
		}
	}
	publish := func(a executionAttempt) {
		t.Helper()
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := publishAttemptTx(ctx, tx, a, a.Start.Add(time.Second)); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	first := admit("import", vector.At)
	prepare(first)
	var health string
	var envelope *string
	if err := pool.QueryRow(ctx, `SELECT health,current_envelope_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version).Scan(&health, &envelope); err != nil || health != "staged" || envelope != nil {
		t.Fatal("preparation published content", health, envelope, err)
	}
	var preparedBytes []byte
	if err := pool.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1`, first.ID).Scan(&preparedBytes); err != nil {
		t.Fatal(err)
	}
	prepared, err := decodePrepared(preparedBytes)
	if err != nil {
		t.Fatal(err)
	}
	for _, existing := range []bool{false, true} {
		name := "unpublished object"
		if existing {
			name = "object without operation ownership"
		}
		t.Run(name, func(t *testing.T) {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			altered := prepared
			altered.Objects = slices.Clone(prepared.Objects)
			// Container sorts first; substitute a logical member with the same
			// declared bytes but without the preparation's retention proof.
			object := &altered.Objects[1]
			object.ID = uuid.New()
			object.Reference, _ = ParseStorageRef("objects/" + uuid.NewString())
			if existing {
				if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_objects(object_id,sha256,storage_ref,size_bytes,generation) VALUES($1,$2,$3,$4,1)`, object.ID, object.Digest, object.Reference.String(), object.Size); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := publishVersionTx(ctx, tx, altered, first.OperationID, nil); err == nil {
				t.Fatal("publication manufactured or borrowed an unowned object")
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM reference_pack_objects WHERE object_id=$1`, object.ID).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 && !existing {
				t.Fatal("publication inserted an unprepared object")
			}
		})
	}
	publish(first)
	if err := pool.QueryRow(ctx, `SELECT health,current_envelope_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version).Scan(&health, &envelope); err != nil || health != "verified_available" || envelope == nil {
		t.Fatal("signed import unavailable", health, err)
	}
	firstEnvelope := *envelope
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_candidates SET administratively_disabled=true WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version); err != nil {
		t.Fatal(err)
	}
	a := admit("reverify", vector.At.Add(time.Minute))
	prepare(a)
	publish(a)
	var disabled bool
	if err := pool.QueryRow(ctx, `SELECT administratively_disabled,current_envelope_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version).Scan(&disabled, &envelope); err != nil || !disabled || *envelope == firstEnvelope {
		t.Fatal("verification changed disablement or failed to renew envelope", err)
	}
	// Activate the verified tuple to exercise invalidation and mandatory fallback.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET administratively_disabled=false WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version); err != nil {
		t.Fatal(err)
	}
	var setData []byte
	if err := tx.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets JOIN reference_pack_current_set USING(pack_set_id)`).Scan(&setData); err != nil {
		t.Fatal(err)
	}
	base, err := decodeRetainedSet(setData)
	if err != nil {
		t.Fatal(err)
	}
	v, err := loadStateVersionTx(ctx, tx, identity.Key, identity.Version)
	if err != nil {
		t.Fatal(err)
	}
	for i := range base.Members {
		if base.Members[i].Key == identity.Key {
			base.Members[i] = PackSetMember(v.Member)
		}
	}
	active, err := publishSetTx(ctx, tx, base.Members, a.OperationID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_current_set SET pack_set_id=$1,revision=revision+1`, active.ID); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	testPortableContainerExport(t, ctx, pool, storage, active.ID, container)
	// Expiry on renewal preserves the old success and the active set.
	renewal := admit("import", vector.Expiry)
	prepare(renewal)
	publish(renewal)
	var current string
	if err := pool.QueryRow(ctx, `SELECT health,(SELECT pack_set_id FROM reference_pack_current_set) FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version).Scan(&health, &current); err != nil || health != "verified_available" || current != active.ID {
		t.Fatal("failed renewal condemned established content", health, current, err)
	}
	// Expiry on reverify condemns the active version and restores Base in the
	// same transaction, while retaining the historical set and its provenance.
	// A present but altered retained member must not mask the earlier trust
	// expiry check. Presence and content hashing have different registry ranks.
	var retainedManifest string
	if err := pool.QueryRow(ctx, `SELECT o.storage_ref FROM reference_pack_object_refs r JOIN reference_pack_objects o USING(object_id) WHERE r.owner_kind='version' AND r.owner_id=$1 AND r.logical_path='manifest.json'`, versionObjectID(identity.Key, identity.Version)).Scan(&retainedManifest); err != nil {
		t.Fatal(err)
	}
	originalManifest := storage.objects[retainedManifest]
	storage.objects[retainedManifest] = []byte("altered retained manifest")
	expired := admit("reverify", vector.Expiry)
	prepare(expired)
	var check string
	if err := pool.QueryRow(ctx, `SELECT check_id FROM reference_pack_attempt_members WHERE attempt_id=$1`, expired.ID).Scan(&check); err != nil || check != "metadata_expiry" {
		t.Fatal("retained hash masked trust expiry", check, err)
	}
	storage.objects[retainedManifest] = originalManifest
	publish(expired)
	if err := pool.QueryRow(ctx, `SELECT health,(SELECT pack_set_id FROM reference_pack_current_set) FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, identity.Key, identity.Version).Scan(&health, &current); err != nil || health != "failed" || current == active.ID {
		t.Fatal("invalidation did not atomically fall back", health, current, err)
	}
	var fallbackVersion string
	if err := pool.QueryRow(ctx, `SELECT pack_version FROM reference_pack_current_set JOIN reference_pack_set_members USING(pack_set_id) WHERE pack_key=$1`, identity.Key).Scan(&fallbackVersion); err != nil {
		t.Fatal(err)
	}
	assertFallback := func(want *string) {
		t.Helper()
		resource, err := c.GetVersion(ctx, identity.Key, fallbackVersion)
		if err != nil || !resource.Active || (resource.FallbackFromVersion == nil) != (want == nil) || (want != nil && *resource.FallbackFromVersion != *want) {
			t.Fatal("fallback attribution", resource, err)
		}
	}
	assertFallback(&identity.Version)
	if err := ReconcileBaseRelease(ctx, pool, storage, BaseReleaseOptions{ProfileClaimed: true, ClockTrusted: true, Limits: DefaultLimits()}, vector.Expiry.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	assertFallback(&identity.Version)
	if err := ReconcileBaseRelease(ctx, pool, storage, BaseReleaseOptions{Limits: DefaultLimits()}, vector.Expiry.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	assertFallback(nil)
	var anchors int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_sets WHERE pack_set_id=$1`, active.ID).Scan(&anchors); err != nil || anchors != 1 {
		t.Fatal("historical provenance lost", err)
	}
	// A later path's absence wins over an earlier path's altered bytes in both
	// direct consumer integrity checks and live cohort preparation.
	var otherRef string
	if err := pool.QueryRow(ctx, `SELECT o.storage_ref FROM reference_pack_object_refs r JOIN reference_pack_objects o USING(object_id) WHERE r.owner_kind='version' AND r.owner_id=$1 AND r.logical_path<>'manifest.json' ORDER BY r.logical_path COLLATE "C" DESC LIMIT 1`, versionObjectID(identity.Key, identity.Version)).Scan(&otherRef); err != nil {
		t.Fatal(err)
	}
	originalOther := storage.objects[otherRef]
	// All members at a winning retained check contribute findings. This
	// inventory comes from a successful immutable manifest, not input names.
	storage.objects[retainedManifest] = append(bytes.Clone(originalManifest), 'x')
	storage.objects[otherRef] = append(bytes.Clone(originalOther), 'x')
	var complete *ContentRejection
	if err := inspectRetainedVersion(ctx, pool, storage, identity.Key, identity.Version, retainedLengths); !errors.As(err, &complete) || complete.CheckID != "member_length" || complete.Summary == nil || complete.Summary.Total != 2 {
		t.Fatalf("incomplete retained lengths: %#v / %v", complete, err)
	}
	storage.objects[retainedManifest] = bytes.Clone(originalManifest)
	storage.objects[retainedManifest][0] ^= 1
	storage.objects[otherRef] = bytes.Clone(originalOther)
	storage.objects[otherRef][0] ^= 1
	if err := inspectRetainedVersion(ctx, pool, storage, identity.Key, identity.Version, retainedHashes); !errors.As(err, &complete) || complete.CheckID != "member_hash" || complete.Summary == nil || complete.Summary.Total != 2 {
		t.Fatalf("incomplete retained hashes: %#v / %v", complete, err)
	}
	storage.objects[retainedManifest] = []byte("altered retained manifest")
	delete(storage.objects, otherRef)
	var rejected *ContentRejection
	if err := c.checkRetainedMembers(ctx, frozenMember{Key: identity.Key, Version: identity.Version}); !errors.As(err, &rejected) || rejected.CheckID != "retained_payload" {
		t.Fatal("consumer hash masked retained loss", err)
	}
	delete(storage.objects, retainedManifest)
	if err := c.checkRetainedMembers(ctx, frozenMember{Key: identity.Key, Version: identity.Version}); !errors.As(err, &rejected) || rejected.CheckID != "retained_payload" || rejected.Summary == nil || rejected.Summary.Total != 2 {
		t.Fatalf("incomplete retained presence: %#v / %v", rejected, err)
	}
	lost := admit("reverify", vector.Expiry)
	prepare(lost)
	if err := pool.QueryRow(ctx, `SELECT check_id FROM reference_pack_attempt_members WHERE attempt_id=$1`, lost.ID).Scan(&check); err != nil || check != "retained_payload" {
		t.Fatal("retained loss did not precede trust expiry", check, err)
	}
	var retainedSummary []byte
	if err := pool.QueryRow(ctx, `SELECT canonical_validation_summary FROM reference_pack_attempt_members WHERE attempt_id=$1`, lost.ID).Scan(&retainedSummary); err != nil {
		t.Fatal(err)
	}
	if summary, err := packformat.DecodeValidationSummary(retainedSummary); err != nil || summary.Total != 2 {
		t.Fatal("prepared attempt discarded retained findings", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_attempt_members SET invalidated_content=true,canonical_validation_summary=canonical_validation_summary||' '::bytea WHERE attempt_id=$1`, lost.ID); err == nil {
		t.Fatal("invalidation guard allowed diagnostic mutation")
	}
	publish(lost)
	storage.objects[retainedManifest] = originalManifest
	storage.objects[otherRef] = originalOther
}
