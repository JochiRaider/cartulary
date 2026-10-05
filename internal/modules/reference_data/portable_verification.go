package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// A portable cohort reuses the live verifier. Its only extra identity source is
// a predecessor that has already passed every check in this same attempt.
type portableVerificationIdentity struct {
	operationVerificationIdentity
	attempt uuid.UUID
}

func (i portableVerificationIdentity) resolveDependency(ctx context.Context, d packformat.Dependency) (packformat.Manifest, bool, error) {
	var prepared []byte
	err := i.pool.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1 AND pack_key=$2 AND pack_version=$3`, i.attempt, d.Key, d.Version).Scan(&prepared)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return packformat.Manifest{}, false, err
	}
	if err == nil && prepared != nil {
		p, err := decodePrepared(prepared)
		if err != nil {
			return packformat.Manifest{}, false, err
		}
		if p.Content.PayloadSHA256 == d.SHA256 {
			return p.Content.Manifest, true, nil
		}
	}
	// A selected candidate with no successful predecessor cannot obtain a
	// verdict from source attestations or from its unverified lexical manifest.
	var selected bool
	if err := i.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operation_members WHERE operation_id=$1 AND pack_key=$2 AND pack_version=$3)`, i.operationID, d.Key, d.Version).Scan(&selected); err != nil {
		return packformat.Manifest{}, false, err
	}
	if selected {
		return packformat.Manifest{}, false, nil
	}
	return i.operationVerificationIdentity.resolveDependency(ctx, d)
}

func (i portableVerificationIdentity) checkReleaseSequence(ctx context.Context, m packformat.Manifest, manifestSHA, payloadSHA string) error {
	if err := i.referenceDependencies.checkReleaseSequence(ctx, m, manifestSHA, payloadSHA); err != nil {
		return err
	}
	// The frozen repository high-water mark applies to the whole cohort. Earlier
	// predecessors do not raise it, but cannot assign one sequence two identities.
	var collision bool
	if err := i.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_attempt_members a
 CROSS JOIN LATERAL (SELECT convert_from(decode(convert_from(a.canonical_prepared,'UTF8')::jsonb->>'manifest','base64'),'UTF8')::jsonb AS manifest) p
 WHERE a.attempt_id=$1 AND a.verdict='succeeded' AND a.pack_key=$2
 AND p.manifest->>'trust_repository_id'=$3 AND (p.manifest->>'pack_release_sequence')::bigint=$4
 AND (convert_from(a.canonical_prepared,'UTF8')::jsonb->'envelope'->>'manifest_sha256'<>$5 OR convert_from(a.canonical_prepared,'UTF8')::jsonb->'envelope'->>'payload_sha256'<>$6))`, i.attempt, m.Key, m.Repository, m.Sequence, manifestSHA, payloadSHA).Scan(&collision); err != nil {
		return err
	}
	if collision {
		return &ContentRejection{Code: "pack_release_sequence_collision", CheckID: "sequence_collision", CandidateKey: m.Key, CandidateVersion: m.Version}
	}
	return nil
}

func (r *incidentReferences) verifyPortablePreparation(ctx context.Context, request IncidentReferenceImportRequest, p *PreparedReferenceImport, cohort *portableVerificationCohort) (executionAttempt, error) {
	a := executionAttempt{ID: uuid.New(), OperationID: cohort.operation, Actor: &p.actor, Start: r.verifier.now().UTC(), Count: int64(len(cohort.order)), Frozen: frozenOperation{Kind: "import", ClockTrusted: cohort.context.ClockTrusted, TimeoutSeconds: cohort.context.TimeoutSeconds, ConfigurationSHA256: cohort.context.ConfigurationSHA256}}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return a, err
	}
	defer tx.Rollback(ctx)
	if err := r.executions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
		return a, err
	}
	var terminal bool
	if err := tx.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE operation_id=$1 FOR UPDATE`, a.OperationID).Scan(&terminal); err != nil {
		return a, err
	}
	if terminal {
		return a, errors.New("reference pack: terminal portable operation cannot start verification")
	}
	if err := retainAttemptStartTx(ctx, tx, a); err != nil {
		return a, err
	}
	if err := tx.Commit(ctx); err != nil {
		return a, err
	}
	var configuration string
	if err := r.pool.QueryRow(ctx, `SELECT configuration_sha256 FROM reference_pack_current_set WHERE singleton`).Scan(&configuration); err != nil {
		return a, err
	}
	if configuration != cohort.context.ConfigurationSHA256 {
		return a, &OperationRejection{Reason: "stale_admission_state"}
	}
	identity := portableVerificationIdentity{operationVerificationIdentity: operationVerificationIdentity{referenceDependencies: r.verifier.referenceDependencies, operationID: a.OperationID}, attempt: a.ID}
	for index, sourceIndex := range cohort.order {
		m, err := r.verifier.frozenMember(ctx, a, int64(index+1))
		if err != nil {
			return a, err
		}
		source := cohort.inputs[sourceIndex]
		input := verificationAttempt{Observer: r.verifier.observer, Start: a.Start, ClockTrusted: cohort.context.ClockTrusted, Limits: r.verifier.limits.verificationArchiveLimits(), Identity: identity, Retained: &source.object.Reference, ContainerSHA256: source.object.Digest}
		input.ResolveTrust = func(ctx context.Context, id string, versions []int64) (packformat.TrustSnapshot, bool, error) {
			return r.verifier.frozenTrust(ctx, a, m, id, versions)
		}
		input.IdentityAdmitted = func(_ context.Context, key, version string) error {
			if key != m.Key || version != m.Version {
				return portableIdentityError()
			}
			return nil
		}
		if err := r.verifier.prepareVerificationMember(ctx, a, m, input); err != nil {
			return a, err
		}
		var data []byte
		if err := r.pool.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, m.Ordinal).Scan(&data); err != nil {
			return a, err
		}
		if data == nil {
			continue
		}
		prepared, err := decodePrepared(data)
		if err != nil {
			return a, err
		}
		reference := p.versions[sourceIndex].reference
		if prepared.Content.ManifestSHA256 != reference.ManifestSHA256 || prepared.Content.PayloadSHA256 != reference.PayloadSHA256 || !portableEmbeddingAllowed(prepared.Content.Manifest) {
			return a, portableIdentityError()
		}
	}
	return a, nil
}
