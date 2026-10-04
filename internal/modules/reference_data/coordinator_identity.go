package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Imports verify their admitted new container. Reverify and refresh must also
// verify the exact established logical content; the retained container is not
// permission to repair a missing or altered published member.
type operationVerificationIdentity struct {
	*Coordinator
	retained    *frozenMember
	operationID uuid.UUID
}

func (i operationVerificationIdentity) checkRetainedPresence(ctx context.Context) error {
	if i.retained == nil {
		return nil
	}
	return inspectRetainedVersion(ctx, i.pool, i.storage, i.retained.Key, i.retained.Version, retainedPresence)
}

func (i operationVerificationIdentity) checkRetainedLength(ctx context.Context) error {
	if i.retained == nil {
		return nil
	}
	return inspectRetainedVersion(ctx, i.pool, i.storage, i.retained.Key, i.retained.Version, retainedLengths)
}

func (i operationVerificationIdentity) checkRetainedIntegrity(ctx context.Context) error {
	if i.retained == nil {
		return nil
	}
	return inspectRetainedVersion(ctx, i.pool, i.storage, i.retained.Key, i.retained.Version, retainedHashes)
}

// Verification classifies immutable-identity failures before proposing any
// publication. Captured key revisions protect this read until finalization;
// publication repeats its constraints and rejects stale admission first.
func (c *Coordinator) checkReleaseSequence(ctx context.Context, m packformat.Manifest, manifestSHA, payloadSHA string) error {
	var highest int64
	var exactSequence, conflictingSequence bool
	if err := c.pool.QueryRow(ctx, `SELECT coalesce(max(pack_release_sequence),0),coalesce(bool_or(pack_release_sequence=$3 AND manifest_sha256=$4 AND payload_sha256=$5),false),coalesce(bool_or(pack_release_sequence=$3 AND (manifest_sha256<>$4 OR payload_sha256<>$5)),false) FROM reference_pack_versions WHERE repository_id IS NOT DISTINCT FROM $1 AND pack_key=$2`, m.Repository, m.Key, m.Sequence, manifestSHA, payloadSHA).Scan(&highest, &exactSequence, &conflictingSequence); err != nil {
		return err
	}
	if m.Sequence < highest && !exactSequence {
		return &ContentRejection{Code: "pack_release_sequence_rollback", CheckID: "sequence_rollback", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	if conflictingSequence {
		return &ContentRejection{Code: "pack_release_sequence_collision", CheckID: "sequence_collision", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	return nil
}

func (c *Coordinator) checkLogicalVersion(ctx context.Context, m packformat.Manifest, manifestSHA, payloadSHA string) error {
	var distribution string
	err := c.pool.QueryRow(ctx, `SELECT distribution_kind FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, m.Key, m.Version).Scan(&distribution)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil && distribution != "operator_imported" {
		return &ContentRejection{Code: "pack_version_collision", CheckID: "logical_collision", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	var manifest, payload string
	err = c.pool.QueryRow(ctx, `SELECT manifest_sha256,payload_sha256 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, m.Key, m.Version).Scan(&manifest, &payload)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil && (manifest != manifestSHA || payload != payloadSHA) {
		return &ContentRejection{Code: "pack_version_collision", CheckID: "logical_collision", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	return nil
}

func (i operationVerificationIdentity) resolveDependency(ctx context.Context, d packformat.Dependency) (packformat.Manifest, bool, error) {
	var revision, admitted int64
	err := i.pool.QueryRow(ctx, `SELECT k.revision,f.admitted_revision FROM reference_pack_key_state k JOIN reference_pack_operation_dependency_keys f USING(pack_key) WHERE f.operation_id=$1 AND k.pack_key=$2`, i.operationID, d.Key).Scan(&revision, &admitted)
	if errors.Is(err, pgx.ErrNoRows) || err == nil && revision != admitted {
		return packformat.Manifest{}, false, &OperationRejection{Reason: "stale_admission_state"}
	}
	if err != nil {
		return packformat.Manifest{}, false, err
	}
	return resolveRetainedDependency(ctx, i.pool, i.storage, d, false)
}

func resolveRetainedDependency(ctx context.Context, db retainedContentQuery, storage VerificationStorage, d packformat.Dependency, historical bool) (packformat.Manifest, bool, error) {
	var raw []byte
	var digest, distribution, health string
	var removed, successful bool
	err := db.QueryRow(ctx, `SELECT v.manifest_bytes,v.payload_sha256,c.distribution_kind,c.health,c.removed,c.current_envelope_id IS NOT NULL FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version) WHERE v.pack_key=$1 AND v.pack_version=$2`, d.Key, d.Version).Scan(&raw, &digest, &distribution, &health, &removed, &successful)
	if errors.Is(err, pgx.ErrNoRows) {
		return packformat.Manifest{}, false, nil
	}
	if err != nil {
		return packformat.Manifest{}, false, err
	}
	if digest != d.SHA256 || !successful || !historical && (removed || health != "verified_available") {
		return packformat.Manifest{}, false, nil
	}
	manifest, err := packformat.DecodeManifest(raw, distribution == "operator_imported")
	if err != nil {
		return packformat.Manifest{}, false, errHistoricalIntegrity
	}
	if err := checkRetainedVersion(ctx, db, storage, d.Key, d.Version); err != nil {
		var rejection *ContentRejection
		if errors.As(err, &rejection) {
			return packformat.Manifest{}, false, nil
		}
		return packformat.Manifest{}, false, err
	}
	return manifest, true, nil
}
