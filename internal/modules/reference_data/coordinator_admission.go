package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Frozen inputs contain immutable references, not a copied in-memory refresh
// cohort. Members and repository revisions live in ordered, immutable rows so
// admission and execution need not allocate in proportion to retained versions.
type frozenOperation struct {
	SchemaID            string    `json:"schema_id"`
	Kind                string    `json:"kind"`
	AdmittedAt          time.Time `json:"admitted_at"`
	PackSetID           *string   `json:"pack_set_id"`
	SetRevision         int64     `json:"set_revision"`
	ConfigurationSHA256 string    `json:"configuration_sha256"`
	ClockTrusted        bool      `json:"clock_trusted"`
	TimeoutSeconds      int64     `json:"timeout_seconds"`
	ContainerReference  *string   `json:"container_reference"`
	ContainerSHA256     *string   `json:"container_sha256"`
	ContainerBytes      *int64    `json:"container_bytes"`
}

// previousSetID is used only for optional historical attestation linkage.
func (f frozenOperation) previousSetID() string {
	if f.PackSetID == nil {
		return ""
	}
	return *f.PackSetID
}

type importIdentity struct {
	Key, Version, Repository string
	Dependencies             []packformat.Dependency
	Manifest                 *packformat.Manifest
	ManifestSHA256           string
}

// ProbeContainerIdentity is bounded lexical attribution only. Its result is
// never evidence of authenticity or success. Invalid content is examined again
// by the queued attempt so it receives the ordered verification verdict.
func probeContainerIdentity(ctx context.Context, storage VerificationStorage, ref StagingRef, limits packformat.ArchiveLimits) (identity importIdentity, resultErr error) {
	source, size, err := storage.OpenStaged(ctx, ref)
	if err != nil {
		return identity, err
	}
	checked := &checkedContainerReader{ContainerReader: source}
	source = checked
	defer func() {
		if operationalErr := errors.Join(checked.fault, source.Close()); operationalErr != nil {
			identity = importIdentity{}
			resultErr = operationalErr
		}
	}()
	workspace, err := storage.NewWorkspace(ctx)
	if err != nil {
		return identity, err
	}
	defer func() { resultErr = errors.Join(resultErr, workspace.Close()) }()
	inventory, err := packformat.Extract(ctx, source, size, limits, workspace)
	if err != nil {
		var invalid *packformat.Failure
		if errors.As(err, &invalid) {
			return identity, nil
		}
		return identity, err
	}
	data, err := readAttemptMember(ctx, workspace, inventory, "manifest.json", 1048576)
	if err == nil {
		identity.Key, identity.Version = packformat.SafeManifestIdentity(data)
		if manifest, err := packformat.DecodeManifestSchema(data, true); err == nil {
			identity.Dependencies = manifest.Dependencies
			identity.Manifest = &manifest
			identity.ManifestSHA256 = packformat.Digest(data)
		}
	} else {
		var invalid *packformat.Failure
		if !errors.As(err, &invalid) {
			return identity, err
		}
	}
	data, err = readAttemptMember(ctx, workspace, inventory, "bundle.json", 16384)
	if err == nil {
		identity.Repository, _ = packformat.DecodeHint(data)
	} else {
		var invalid *packformat.Failure
		if !errors.As(err, &invalid) {
			return identity, err
		}
	}
	return identity, nil
}

type operationAdmission struct {
	ID             uuid.UUID
	JobID          *uuid.UUID
	Kind           string
	ActorKind      string
	Actor          *uuid.UUID
	At             time.Time
	Keys           []string
	Version        string
	Import         *importIdentity
	ContainerRef   *StorageRef
	InputObject    *preparedObject
	ContainerSHA   *string
	ContainerBytes *int64
	ClockTrusted   bool
	TimeoutSeconds int64
}

// admitOperationTx is called by the coordinator in the same transaction as
// Jobs admission and its idempotency receipt. Locks are short and acquired in
// repository/key/current-set/usage order; no verification runs under them.
func admitOperationTx(ctx context.Context, tx pgx.Tx, a operationAdmission) (frozenOperation, error) {
	frozen := frozenOperation{SchemaID: "cartulary.reference_pack_operation_input.v2", Kind: a.Kind, AdmittedAt: a.At.UTC(), ClockTrusted: a.ClockTrusted, TimeoutSeconds: a.TimeoutSeconds, ContainerSHA256: a.ContainerSHA, ContainerBytes: a.ContainerBytes}
	if a.ID == uuid.Nil || a.At.IsZero() || a.TimeoutSeconds < 60 || a.TimeoutSeconds > 86400 {
		return frozen, errors.New("reference pack: invalid operation admission")
	}
	if a.InputObject != nil {
		if err := lockByteRetentionTx(ctx, tx, false); err != nil {
			return frozen, err
		}
	}
	if a.ContainerRef != nil {
		ref := a.ContainerRef.String()
		frozen.ContainerReference = &ref
	}
	keys := slices.Clone(a.Keys)
	slices.Sort(keys)
	keys = slices.Compact(keys)
	if a.Kind == "import" && a.Import != nil && a.Import.Key != "" {
		keys = []string{a.Import.Key}
	}
	// Freeze repository rows first. A refresh reads repositories through its
	// admitted envelope inventory; a later relevant key change is detected by
	// the short key lock below rather than silently changing the cohort.
	repositories, err := admissionRepositoryIDs(ctx, tx, a, keys)
	if err != nil {
		return frozen, err
	}
	type rootRevision struct {
		id                string
		revision, version int64
	}
	rootRevisions := []rootRevision{}
	for _, id := range repositories {
		root := rootRevision{id: id}
		err := tx.QueryRow(ctx, `SELECT revision,root_version FROM reference_pack_repositories WHERE repository_id=$1 FOR UPDATE`, id).Scan(&root.revision, &root.version)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return frozen, err
		}
		rootRevisions = append(rootRevisions, root)
	}
	dependencyRevisions, err := admissionDependencyRevisions(ctx, tx, a, keys)
	if err != nil {
		return frozen, err
	}
	for _, key := range admissionGuardKeys(keys, dependencyRevisions) {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_key_state(pack_key) VALUES($1) ON CONFLICT DO NOTHING`, key); err != nil {
			return frozen, err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, key); err != nil {
			return frozen, err
		}
	}
	for _, d := range dependencyRevisions {
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, d.key).Scan(&revision); err != nil {
			return frozen, err
		}
		if revision != d.revision {
			return frozen, &OperationRejection{Reason: "stale_admission_state"}
		}
	}
	confirmedDependencies, err := admissionDependencyRevisions(ctx, tx, a, keys)
	if err != nil {
		return frozen, err
	}
	if !slices.Equal(dependencyRevisions, confirmedDependencies) {
		return frozen, &OperationRejection{Reason: "stale_admission_state"}
	}
	// Pending work wins over eligibility, including work on another version.
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operation_keys k JOIN reference_pack_operations o USING(operation_id) WHERE k.pack_key=ANY($1) AND o.terminal_at IS NULL)`, keys).Scan(&pending); err != nil {
		return frozen, err
	}
	if pending && a.Kind != "integrity" {
		return frozen, &OperationRejection{Reason: "verification_pending"}
	}
	currentRepositories, err := admissionRepositoryIDs(ctx, tx, a, keys)
	if err != nil {
		return frozen, err
	}
	if !slices.Equal(repositories, currentRepositories) {
		return frozen, &OperationRejection{Reason: "stale_admission_state"}
	}
	if a.Kind == "reverify" {
		if len(keys) != 1 {
			return frozen, errors.New("reference pack: invalid reverify selection")
		}
		var envelope *string
		var distribution string
		var removed bool
		if err := tx.QueryRow(ctx, `SELECT current_envelope_id,distribution_kind,removed FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, keys[0], a.Version).Scan(&envelope, &distribution, &removed); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return frozen, ErrNotFound
			}
			return frozen, err
		}
		if distribution == "packaged_builtin" {
			return frozen, &OperationRejection{Reason: "packaged_builtin"}
		}
		if removed {
			return frozen, &OperationRejection{Reason: "removed"}
		}
		if envelope == nil {
			return frozen, &OperationRejection{Reason: "no_successful_verification"}
		}
	}
	if err := tx.QueryRow(ctx, `SELECT pack_set_id,revision,configuration_sha256 FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&frozen.PackSetID, &frozen.SetRevision, &frozen.ConfigurationSHA256); err != nil {
		return frozen, err
	}
	if frozen.PackSetID == nil && a.Kind != "integrity" {
		return frozen, &OperationRejection{Reason: "required_registry_gap"}
	}
	// Only normal activation consults record assignments for replacement
	// compatibility. Verification and emergency fallback preserve assignments;
	// unrelated record creation must not invalidate their frozen inputs.
	if a.Kind == "activate" {
		for _, key := range keys {
			if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_registry_usage WHERE pack_key=$1 FOR UPDATE`, key); err != nil {
				return frozen, err
			}
		}
	}
	encoded, err := canonicaljson.Marshal(frozen)
	if err != nil {
		return frozen, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,job_id,kind,actor_kind,actor_user_id,admitted_at,frozen_input) VALUES($1,$2,$3,$4,$5,$6,$7)`, a.ID, a.JobID, a.Kind, a.ActorKind, a.Actor, a.At, encoded); err != nil {
		return frozen, err
	}
	if a.Kind == "import" {
		if a.InputObject == nil || a.ContainerRef == nil || a.ContainerSHA == nil || a.ContainerBytes == nil {
			return frozen, errors.New("reference pack: unretained import input")
		}
		object := a.InputObject
		if object.Reference != *a.ContainerRef || object.Digest != *a.ContainerSHA || object.Size != *a.ContainerBytes {
			return frozen, errors.New("reference pack: inconsistent frozen container")
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_objects(object_id,sha256,storage_ref,size_bytes,generation) VALUES($1,$2,$3,$4,1)`, object.ID, object.Digest, object.Reference.String(), object.Size); err != nil {
			return frozen, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_object_refs(owner_kind,owner_id,logical_path,object_id) VALUES('operation',$1,'input/container',$2)`, a.ID.String(), object.ID); err != nil {
			return frozen, err
		}
	}
	for _, key := range keys {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_keys(operation_id,pack_key,admitted_revision,usage_revision) SELECT $1,k.pack_key,k.revision,CASE WHEN $3='activate' THEN u.revision END FROM reference_pack_key_state k LEFT JOIN reference_pack_registry_usage u USING(pack_key) WHERE k.pack_key=$2`, a.ID, key, a.Kind); err != nil {
			return frozen, err
		}
	}
	for _, d := range dependencyRevisions {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_dependency_keys(operation_id,pack_key,admitted_revision) VALUES($1,$2,$3)`, a.ID, d.key, d.revision); err != nil {
			return frozen, err
		}
	}
	if a.Kind == "import" && a.Import != nil && a.Import.Key != "" {
		if err := insertCandidateTx(ctx, tx, a.Import.Key, a.Import.Version, "operator_imported", a.Actor, a.At); err != nil {
			return frozen, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_members(operation_id,ordinal,pack_key,pack_version,envelope_id) SELECT $1,1,pack_key,pack_version,current_envelope_id FROM reference_pack_candidates WHERE pack_key=$2 AND pack_version=$3`, a.ID, a.Import.Key, a.Import.Version); err != nil {
			return frozen, err
		}
	} else if a.Kind == "reverify" || a.Kind == "refresh" {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_members(operation_id,ordinal,pack_key,pack_version,envelope_id) SELECT $1,row_number() OVER(ORDER BY pack_key COLLATE "C",pack_version COLLATE "C"),pack_key,pack_version,current_envelope_id FROM reference_pack_candidates WHERE pack_key=ANY($2) AND ($3='' OR pack_version=$3) AND distribution_kind='operator_imported' AND NOT removed AND current_envelope_id IS NOT NULL`, a.ID, keys, a.Version); err != nil {
			return frozen, err
		}
	}
	for _, root := range rootRevisions {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_repositories(operation_id,repository_id,admitted_revision,root_version) VALUES($1,$2,$3,$4)`, a.ID, root.id, root.revision, root.version); err != nil {
			return frozen, err
		}
	}
	return frozen, nil
}

func admissionRepositoryIDs(ctx context.Context, tx pgx.Tx, a operationAdmission, keys []string) ([]string, error) {
	if a.Import != nil {
		if a.Import.Repository != "" {
			return []string{a.Import.Repository}, nil
		}
		return []string{}, nil
	}
	rows, err := tx.Query(ctx, `SELECT DISTINCT e.repository_id COLLATE "C" FROM reference_pack_candidates c JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE c.pack_key=ANY($1) AND ($2='' OR c.pack_version=$2) AND e.repository_id IS NOT NULL AND NOT c.removed AND c.distribution_kind='operator_imported' ORDER BY 1`, keys, a.Version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func decodeFrozenOperation(data []byte) (frozenOperation, error) {
	var value frozenOperation
	decoded, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return value, errors.New("reference pack: invalid retained operation input")
	}
	object, ok := decoded.(map[string]any)
	if !ok || len(object) != 11 {
		return value, errors.New("reference pack: invalid retained operation input")
	}
	for _, name := range []string{"schema_id", "kind", "admitted_at", "set_revision", "configuration_sha256", "clock_trusted", "timeout_seconds"} {
		if object[name] == nil {
			return value, errors.New("reference pack: incomplete retained operation input")
		}
	}
	for _, name := range []string{"pack_set_id", "container_reference", "container_sha256", "container_bytes"} {
		if _, ok := object[name]; !ok {
			return value, errors.New("reference pack: omitted nullable operation input")
		}
	}
	canonical, err := canonicaljson.Marshal(decoded)
	if err != nil || !bytes.Equal(canonical, data) {
		return value, errors.New("reference pack: noncanonical operation input")
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	if value.SchemaID != "cartulary.reference_pack_operation_input.v2" || value.AdmittedAt.IsZero() || value.TimeoutSeconds < 60 || value.TimeoutSeconds > 86400 || !slices.Contains([]string{"import", "reverify", "refresh", "activate", "disable", "remove", "integrity"}, value.Kind) || (value.PackSetID == nil && value.Kind != "integrity" || value.PackSetID != nil && !setIDPattern.MatchString(*value.PackSetID)) || value.SetRevision < 1 || !isLowerSHA256(value.ConfigurationSHA256) {
		return value, errors.New("reference pack: retired or invalid operation input")
	}
	if value.Kind == "import" {
		if value.ContainerReference == nil || value.ContainerSHA256 == nil || value.ContainerBytes == nil || *value.ContainerBytes < 0 || !isLowerSHA256(*value.ContainerSHA256) {
			return value, errors.New("reference pack: incomplete retained import input")
		}
		if _, err := ParseStorageRef(*value.ContainerReference); err != nil {
			return value, errors.New("reference pack: invalid retained input reference")
		}
	} else if value.ContainerReference != nil || value.ContainerSHA256 != nil || value.ContainerBytes != nil {
		return value, errors.New("reference pack: unexpected retained upload input")
	}
	return value, nil
}

func isLowerSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
