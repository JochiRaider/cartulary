package reference_data

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type successfulEnvelope struct {
	SchemaID         string                    `json:"schema_id"`
	OperationID      uuid.UUID                 `json:"operation_id"`
	PackKey          string                    `json:"pack_key"`
	PackVersion      string                    `json:"pack_version"`
	DistributionKind string                    `json:"distribution_kind"`
	ManifestSHA256   string                    `json:"manifest_sha256"`
	PayloadSHA256    string                    `json:"payload_sha256"`
	ContainerSHA256  *string                   `json:"container_sha256"`
	ContainerRef     *string                   `json:"container_ref"`
	VerifiedAt       time.Time                 `json:"verified_at"`
	TrustValidUntil  *time.Time                `json:"trust_valid_until"`
	TrustSnapshot    *packformat.TrustSnapshot `json:"trust_snapshot"`
	TrustProposal    *packformat.TrustProposal `json:"trust_proposal"`
}
type preparedObject struct {
	ID           uuid.UUID
	Path, Digest string
	Size         int64
	Reference    StorageRef
	publication  io.Closer
}

// releasePublication is idempotent through the underlying lease and is safe
// after an uncertain commit: the transaction guard outlives that uncertainty.
func (o preparedObject) releasePublication() error {
	if o.publication == nil {
		return nil
	}
	return o.publication.Close()
}

type preparedVersion struct {
	Content  *VerifiedContent
	Objects  []preparedObject
	Envelope successfulEnvelope
	Rows     []packformat.ContentRow
	IndexID  uuid.UUID
}

func insertCandidateTx(ctx context.Context, tx pgx.Tx, key, version, distribution string, actor *uuid.UUID, at time.Time) error {
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_key_state(pack_key) VALUES($1) ON CONFLICT DO NOTHING`, key); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO reference_pack_candidates(pack_key,pack_version,distribution_kind,health,admitted_at,admitted_by_user_id) VALUES($1,$2,$3,'staged',$4,$5) ON CONFLICT DO NOTHING`, key, version, distribution, at, actor)
	return err
}

func publishVersionTx(ctx context.Context, tx pgx.Tx, p preparedVersion, operationID uuid.UUID, actor *uuid.UUID) (string, error) {
	m := p.Content.Manifest
	if err := insertCandidateTx(ctx, tx, m.Key, m.Version, p.Envelope.DistributionKind, actor, p.Envelope.VerifiedAt); err != nil {
		return "", err
	}
	var distribution string
	if err := tx.QueryRow(ctx, `SELECT distribution_kind FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2 FOR UPDATE`, m.Key, m.Version).Scan(&distribution); err != nil {
		return "", err
	}
	if distribution != p.Envelope.DistributionKind {
		return "", &ContentRejection{Code: "pack_version_collision", CheckID: "logical_collision"}
	}
	var highest int64
	var exactSequence, conflictingSequence bool
	if err := tx.QueryRow(ctx, `SELECT coalesce(max(pack_release_sequence),0),coalesce(bool_or(pack_release_sequence=$3 AND manifest_sha256=$4 AND payload_sha256=$5),false),coalesce(bool_or(pack_release_sequence=$3 AND (manifest_sha256<>$4 OR payload_sha256<>$5)),false) FROM reference_pack_versions WHERE repository_id IS NOT DISTINCT FROM $1 AND pack_key=$2`, m.Repository, m.Key, m.Sequence, p.Content.ManifestSHA256, p.Content.PayloadSHA256).Scan(&highest, &exactSequence, &conflictingSequence); err != nil {
		return "", err
	}
	if m.Sequence < highest && !exactSequence {
		return "", &ContentRejection{Code: "pack_release_sequence_rollback", CheckID: "sequence_rollback"}
	}
	if conflictingSequence {
		return "", &ContentRejection{Code: "pack_release_sequence_collision", CheckID: "sequence_collision"}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_versions(pack_key,pack_version,pack_release_sequence,manifest_sha256,payload_sha256,manifest_bytes,repository_id) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (pack_key,pack_version) DO NOTHING`, m.Key, m.Version, m.Sequence, p.Content.ManifestSHA256, p.Content.PayloadSHA256, p.Content.ManifestBytes, m.Repository); err != nil {
		return "", err
	}
	var retainedManifest, retainedPayload string
	if err := tx.QueryRow(ctx, `SELECT manifest_sha256,payload_sha256 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, m.Key, m.Version).Scan(&retainedManifest, &retainedPayload); err != nil {
		return "", err
	}
	if retainedManifest != p.Content.ManifestSHA256 || retainedPayload != p.Content.PayloadSHA256 {
		return "", &ContentRejection{Code: "pack_version_collision", CheckID: "logical_collision"}
	}
	for _, object := range p.Objects {
		// Operator objects must have been completed and retained during
		// preparation. Publication cannot manufacture their storage records.
		// Base objects are created by the separately verified release boundary.
		if p.Envelope.DistributionKind == "packaged_builtin" {
			if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_objects(object_id,sha256,storage_ref,size_bytes,generation) VALUES($1,$2,$3,$4,1) ON CONFLICT(object_id) DO NOTHING`, object.ID, object.Digest, object.Reference.String(), object.Size); err != nil {
				return "", err
			}
		}
		var matches bool
		if err := tx.QueryRow(ctx, `SELECT sha256=$2 AND storage_ref=$3 AND size_bytes=$4 AND available AND ($7='packaged_builtin' OR EXISTS (
 SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=$1 AND
 ((r.owner_kind='operation' AND r.owner_id=$5::text) OR
 ($6='container' AND r.owner_kind='envelope' AND EXISTS (
  SELECT 1 FROM reference_pack_operation_members m WHERE m.operation_id=$5::text::uuid AND m.envelope_id=r.owner_id
 ))))) FROM reference_pack_objects WHERE object_id=$1`, object.ID, object.Digest, object.Reference.String(), object.Size, operationID.String(), object.Path, p.Envelope.DistributionKind).Scan(&matches); err != nil {
			return "", err
		}
		if !matches {
			return "", errors.New("reference pack: prepared object identity collision")
		}
		if object.Path == "container" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_object_refs(owner_kind,owner_id,logical_path,object_id) VALUES('version',$1,$2,$3) ON CONFLICT(owner_kind,owner_id,logical_path) DO UPDATE SET object_id=excluded.object_id`, versionObjectID(m.Key, m.Version), object.Path, object.ID); err != nil {
			return "", err
		}
	}
	var complete bool
	var indexManifest, indexPayload string
	if err := tx.QueryRow(ctx, `SELECT complete,manifest_sha256,payload_sha256 FROM reference_pack_index_generations WHERE index_id=$1 AND pack_key=$2 AND pack_version=$3 AND operation_id=$4`, p.IndexID, m.Key, m.Version, operationID).Scan(&complete, &indexManifest, &indexPayload); err != nil {
		return "", err
	}
	if !complete || indexManifest != p.Content.ManifestSHA256 || indexPayload != p.Content.PayloadSHA256 {
		return "", errors.New("reference pack: unpublished or inconsistent index")
	}
	p.Envelope.OperationID = operationID
	encoded, err := canonicaljson.Marshal(p.Envelope)
	if err != nil {
		return "", err
	}
	if _, err := decodeSuccessfulEnvelope(encoded); err != nil {
		return "", err
	}
	id := "rpenv_" + packformat.Digest(encoded)
	var repository *string
	var rootVersion *int64
	if p.Envelope.TrustProposal != nil {
		repository = m.Repository
		version, err := packformat.RootVersion(p.Envelope.TrustProposal.Root)
		if err != nil {
			return "", err
		}
		rootVersion = &version
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_envelopes(envelope_id,operation_id,pack_key,pack_version,verified_at,trust_valid_until,container_sha256,container_ref,repository_id,root_version,canonical_envelope) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, id, operationID, m.Key, m.Version, p.Envelope.VerifiedAt, p.Envelope.TrustValidUntil, p.Envelope.ContainerSHA256, p.Envelope.ContainerRef, repository, rootVersion, encoded); err != nil {
		return "", err
	}
	for _, object := range p.Objects {
		if object.Path != "container" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_object_refs(owner_kind,owner_id,logical_path,object_id) VALUES('envelope',$1,'container',$2) ON CONFLICT DO NOTHING`, id, object.ID); err != nil {
			return "", err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health='verified_available',last_failure_code=NULL,missing_reason=NULL,removed=false,current_envelope_id=$3,current_index_id=$4 WHERE pack_key=$1 AND pack_version=$2`, m.Key, m.Version, id, p.IndexID); err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, m.Key)
	return id, err
}

func publishSetTx(ctx context.Context, tx pgx.Tx, members []PackSetMember, operationID uuid.UUID) (PackSet, error) {
	// Activation eligibility is separate from historical set retention. A
	// portable artifact may retain a healthy disabled version, but cannot make
	// that version active as a side effect of retaining its set.
	for _, member := range members {
		var eligible bool
		if err := tx.QueryRow(ctx, `SELECT health='verified_available' AND NOT administratively_disabled AND NOT removed FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, member.Key, member.Version).Scan(&eligible); err != nil {
			return PackSet{}, err
		}
		if !eligible {
			return PackSet{}, consumerError("pack_unavailable")
		}
	}
	set, err := retainSetTx(ctx, tx, members, operationID)
	if err != nil {
		return PackSet{}, err
	}
	// Preserve attribution for unchanged members, including no-op restart
	// reconciliation. A newly selected member starts without a fallback cause;
	// the safety transition records that cause in this same transaction.
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates c SET fallback_from_version=NULL
 WHERE c.fallback_from_version IS NOT NULL
 AND EXISTS(SELECT 1 FROM reference_pack_set_members m WHERE m.pack_set_id=$1 AND m.pack_key=c.pack_key AND m.pack_version=c.pack_version)
 AND NOT EXISTS(SELECT 1 FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id) WHERE s.singleton AND m.pack_key=c.pack_key AND m.pack_version=c.pack_version)`, set.ID); err != nil {
		return PackSet{}, err
	}
	_, err = tx.Exec(ctx, `UPDATE reference_pack_current_set SET pack_set_id=$1,revision=revision+1 WHERE singleton AND pack_set_id IS DISTINCT FROM $1`, set.ID)
	return set, err
}

// retainSetTx creates an immutable destination provenance anchor only on first
// success. It never changes health, administrative disablement or activation.
// The caller owns the pack-key publication guards and validates retained bytes.
func retainSetTx(ctx context.Context, tx pgx.Tx, members []PackSetMember, operationID uuid.UUID) (PackSet, error) {
	set, err := packformat.BuildSet(members)
	if err != nil {
		return PackSet{}, err
	}
	encoded, err := canonicaljson.Marshal(set)
	if err != nil {
		return PackSet{}, err
	}
	anchors := make([]PackProvenance, 0, len(set.Members))
	envelopes := make(map[string]string, len(set.Members))
	for _, member := range set.Members {
		var manifestBytes, envelopeBytes []byte
		var envelopeID string
		err := tx.QueryRow(ctx, `SELECT v.manifest_bytes,e.envelope_id,e.canonical_envelope FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version) JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE v.pack_key=$1 AND v.pack_version=$2 AND c.health='verified_available' AND NOT c.removed`, member.Key, member.Version).Scan(&manifestBytes, &envelopeID, &envelopeBytes)
		if err != nil {
			return PackSet{}, err
		}
		envelope, err := decodeSuccessfulEnvelope(envelopeBytes)
		if err != nil {
			return PackSet{}, err
		}
		manifest, err := packformat.DecodeManifest(manifestBytes, envelope.DistributionKind == "operator_imported")
		if err != nil {
			return PackSet{}, err
		}
		if member.ManifestSHA256 != envelope.ManifestSHA256 || member.PayloadSHA256 != envelope.PayloadSHA256 {
			return PackSet{}, errors.New("reference pack: inconsistent set publication")
		}
		anchors = append(anchors, provenanceFor(set.ID, manifest, envelope))
		envelopes[member.Key] = envelopeID
	}
	provenance, err := canonicaljson.Marshal(anchors)
	if err != nil {
		return PackSet{}, err
	}
	if err := packformat.ValidateProvenance(provenance, set); err != nil {
		return PackSet{}, err
	}
	result, err := tx.Exec(ctx, `INSERT INTO reference_pack_sets(pack_set_id,pack_set_sha256,canonical_set,canonical_provenance,first_operation_id) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, set.ID, set.SHA256, encoded, provenance, operationID)
	if err != nil {
		return PackSet{}, err
	}
	if result.RowsAffected() == 1 {
		for _, member := range set.Members {
			if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_set_members(pack_set_id,pack_key,pack_version,manifest_sha256,payload_sha256,provenance_envelope_id) VALUES($1,$2,$3,$4,$5,$6)`, set.ID, member.Key, member.Version, member.ManifestSHA256, member.PayloadSHA256, envelopes[member.Key]); err != nil {
				return PackSet{}, err
			}
		}
	} else {
		var existing []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets WHERE pack_set_id=$1`, set.ID).Scan(&existing); err != nil {
			return PackSet{}, err
		}
		if !bytes.Equal(existing, encoded) {
			return PackSet{}, errors.New("reference pack: pack-set digest collision")
		}
	}
	return set, nil
}

func provenanceFor(setID string, m packformat.Manifest, envelope successfulEnvelope) PackProvenance {
	method := "packaged_release_manifest_v1"
	signers := []string{}
	if envelope.TrustProposal != nil {
		method = "tuf_1_0_35_offline_bundle_v1"
		signers = append(signers, envelope.TrustProposal.Signers["targets"]...)
	}
	authority := ""
	for _, profile := range packformat.Profiles() {
		if profile.Key == m.Key {
			authority = profile.Authority
			break
		}
	}
	return PackProvenance{PackSetID: setID, PackKey: m.Key, PackVersion: m.Version, ManifestSHA256: envelope.ManifestSHA256, PayloadSHA256: envelope.PayloadSHA256, PackContractVersion: m.Contract, ContentProfileID: m.ProfileID, ContentProfileVersion: m.ProfileVersion, AuthorityClass: authority, SourceProfileID: m.SourceProfileID, SourceProfileSHA256: m.SourceProfileSHA256, SourceIdentifier: m.SourceIdentifier, SourceVersion: m.SourceVersion, SourceAsOf: m.SourceAsOf, SourceArtifacts: m.Artifacts, License: m.License, VerificationMethod: method, LastVerifiedAt: envelope.VerifiedAt, TrustValidUntil: envelope.TrustValidUntil, VerifiedSignerKeyIDs: signers}
}

func publishRootsTx(ctx context.Context, tx pgx.Tx, repository string, proposal packformat.TrustProposal) error {
	versions := make([]int64, 0, len(proposal.RootHistory))
	for v := range proposal.RootHistory {
		versions = append(versions, v)
	}
	slices.Sort(versions)
	for _, v := range versions {
		data := proposal.RootHistory[v]
		var predecessor *int64
		var evidence any = map[string]any{}
		for _, transition := range proposal.RootTransitions {
			if transition.Version == v {
				prev := v - 1
				predecessor = &prev
				evidence = transition
				break
			}
		}
		raw, err := canonicaljson.Marshal(evidence)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_roots(repository_id,root_version,canonical_bytes,sha256,predecessor_version,transition_evidence) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, repository, v, data, packformat.Digest(data), predecessor, raw); err != nil {
			return err
		}
		var retained []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_bytes FROM reference_pack_roots WHERE repository_id=$1 AND root_version=$2`, repository, v).Scan(&retained); err != nil {
			return err
		}
		if !bytes.Equal(retained, data) {
			return &OperationRejection{Reason: "stale_admission_state"}
		}
	}
	current, err := packformat.RootVersion(proposal.Root)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_repositories SET root_version=$2,revision=revision+1 WHERE repository_id=$1 AND root_version<$2`, repository, current); err != nil {
		return err
	}
	return nil
}

func publishMetadataTx(ctx context.Context, tx pgx.Tx, repository string, proposal packformat.TrustProposal, key, version string) error {
	for _, role := range []string{"timestamp", "snapshot", "targets"} {
		metadata := proposal.Metadata[role]
		var retainedVersion int64
		var retainedBytes []byte
		err := tx.QueryRow(ctx, `SELECT metadata_version,canonical_bytes FROM reference_pack_metadata_versions WHERE repository_id=$1 AND pack_key=$2 AND pack_version=$3 AND role=$4`, repository, key, version, role).Scan(&retainedVersion, &retainedBytes)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil && (metadata.Version < retainedVersion || metadata.Version == retainedVersion && !bytes.Equal(metadata.Bytes, retainedBytes)) {
			return &OperationRejection{Reason: "stale_admission_state"}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_metadata_versions(repository_id,pack_key,pack_version,role,metadata_version,canonical_bytes) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(repository_id,pack_key,pack_version,role) DO UPDATE SET metadata_version=excluded.metadata_version,canonical_bytes=excluded.canonical_bytes WHERE reference_pack_metadata_versions.metadata_version<excluded.metadata_version`, repository, key, version, role, metadata.Version, metadata.Bytes); err != nil {
			return err
		}
	}
	return nil
}
