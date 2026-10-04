package reference_data

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

// The parent incident transaction is the sole publication boundary. No active
// selection is changed; only successful destination envelopes, private-candidate
// verdicts, compatible trust proposals and their attributable evidence publish.
func (r *incidentReferences) publishPortableAttemptTx(ctx context.Context, tx pgx.Tx, p *PreparedReferenceImport, cohort *portableVerificationCohort, a executionAttempt, completed time.Time) error {
	current, failures, err := lockPortableAttemptTx(ctx, tx, cohort, a)
	if err != nil {
		return err
	}

	type publication struct {
		ordinal                  int64
		repository, key, version string
		sequence                 int64
	}
	successes := []publication{}
	// Walk one prepared result at a time; metadata allocation is per-container.
	// Merge roots first so any incompatible proposal rolls back every effect.
	for ordinal := int64(1); ordinal <= a.Count; ordinal++ {
		var data []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, ordinal).Scan(&data); err != nil {
			return err
		}
		if data == nil {
			continue
		}
		prepared, err := decodePrepared(data)
		if err != nil {
			return err
		}
		m := prepared.Content.Manifest
		expected := p.versions[cohort.order[ordinal-1]].reference
		if prepared.Envelope.OperationID != cohort.operation || !prepared.Envelope.VerifiedAt.Equal(a.Start) || m.Key != expected.Key || m.Version != expected.Version || prepared.Content.ManifestSHA256 != expected.ManifestSHA256 || prepared.Content.PayloadSHA256 != expected.PayloadSHA256 {
			return errors.New("reference pack: portable prepared binding mismatch")
		}
		if err := publishRootsTx(ctx, tx, *m.Repository, *prepared.Envelope.TrustProposal); err != nil {
			return err
		}
		successes = append(successes, publication{ordinal, *m.Repository, m.Key, m.Version, m.Sequence})
	}
	// Several historical versions of one key may be imported together. Apply
	// sequence checks against the frozen high-water mark, then publish ascending
	// sequences so a newer cohort member cannot retroactively reject an older one.
	slices.SortFunc(successes, func(a, b publication) int {
		if c := strings.Compare(a.repository, b.repository); c != 0 {
			return c
		}
		if c := strings.Compare(a.key, b.key); c != 0 {
			return c
		}
		if a.sequence < b.sequence {
			return -1
		}
		if a.sequence > b.sequence {
			return 1
		}
		return strings.Compare(a.version, b.version)
	})
	for _, success := range successes {
		var data []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, success.ordinal).Scan(&data); err != nil {
			return err
		}
		prepared, err := decodePrepared(data)
		if err != nil {
			return err
		}
		var removed bool
		if err := tx.QueryRow(ctx, `SELECT removed FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, success.key, success.version).Scan(&removed); err != nil {
			return err
		}
		envelope, err := publishVersionTx(ctx, tx, prepared, cohort.operation, &p.actor)
		if err != nil {
			return err
		}
		if err := publishMetadataTx(ctx, tx, success.repository, *prepared.Envelope.TrustProposal, success.key, success.version); err != nil {
			return err
		}
		v := &p.versions[cohort.order[success.ordinal-1]]
		v.envelope = envelope
		v.available = true
		v.reason = nil
		previous := ""
		if current != nil {
			previous = *current
		}
		for _, kind := range []string{"import_verification", "exact_reimport", "trust_root_update"} {
			if kind == "exact_reimport" && !removed || kind == "trust_root_update" && len(prepared.Envelope.TrustProposal.RootTransitions) == 0 {
				continue
			}
			var container *string
			if kind == "import_verification" {
				container = prepared.Envelope.ContainerSHA256
			}
			if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: cohort.operation, Key: success.key, Version: success.version, Kind: kind, Result: "succeeded", At: completed, PreviousSet: previous, ResultingSet: current, Container: container, Verification: kind == "import_verification"}); err != nil {
				return err
			}
		}
	}
	if err := publishPortableRejectionsTx(ctx, tx, p, cohort, a, completed, current); err != nil {
		return err
	}
	outcome := "succeeded"
	if failures > 0 {
		outcome = "content_rejected"
	}
	result, err := packformat.EncodeAttemptResult(outcome, a.Count, failures)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_attempts SET completed_at=$2,outcome=$3,canonical_result=$4 WHERE attempt_id=$1`, a.ID, completed, outcome, result); err != nil {
		return err
	}
	return appendPackAuditTx(ctx, tx, cohort.operation, "verification_completed", outcome, completed, "", "", nil)
}

// Rejections may publish without a catalog when required content prevents the
// parent incident import. Successful siblings remain private in that case.
func publishPortableRejectionsTx(ctx context.Context, tx pgx.Tx, p *PreparedReferenceImport, cohort *portableVerificationCohort, a executionAttempt, completed time.Time, current *string) error {
	for ordinal := int64(1); ordinal <= a.Count; ordinal++ {
		var code, summary *string
		if err := tx.QueryRow(ctx, `SELECT failure_code,validation_summary_id FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, ordinal).Scan(&code, &summary); err != nil {
			return err
		}
		if code == nil {
			continue
		}
		sourceIndex := cohort.order[ordinal-1]
		v := &p.versions[sourceIndex]
		health := "failed"
		var missing *string
		if *code == "payload_missing" {
			health = "missing"
			reason := "staging_loss"
			missing = &reason
		}
		changed, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health=$3,last_failure_code=$4,missing_reason=$5 WHERE pack_key=$1 AND pack_version=$2 AND current_envelope_id IS NULL AND NOT removed`, v.reference.Key, v.reference.Version, health, *code, missing)
		if err != nil {
			return err
		}
		if changed.RowsAffected() > 0 {
			if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, v.reference.Key); err != nil {
				return err
			}
		}
		reason := "not_usable"
		v.available = false
		v.reason = &reason
		previous := ""
		if current != nil {
			previous = *current
		}
		container := cohort.inputs[sourceIndex].object.Digest
		if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: cohort.operation, Key: v.reference.Key, Version: v.reference.Version, Kind: "import_verification", Result: "failed", At: completed, PreviousSet: previous, ResultingSet: current, Container: &container, Verification: true, Summary: summary}); err != nil {
			return err
		}
		auditKind := ""
		switch *code {
		case "tuf_root_untrusted", "tuf_root_rotation_invalid":
			auditKind = "trust_root_rejection"
		case "pack_release_sequence_rollback", "pack_release_sequence_collision":
			auditKind = "sequence_rejection"
		}
		if auditKind != "" {
			if err := appendPackAuditTx(ctx, tx, cohort.operation, auditKind, "rejected", completed, v.reference.Key, v.reference.Version, code); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *incidentReferences) publishRequiredPortableRejectionTx(ctx context.Context, tx pgx.Tx, p *PreparedReferenceImport, completed time.Time) error {
	cohort, a := p.cohort, *p.attempt
	current, failures, err := lockPortableAttemptTx(ctx, tx, cohort, a)
	if err != nil {
		return err
	}
	if failures == 0 {
		return errors.New("reference pack: required rejection lacks a rejected member")
	}
	// Close unpublished successful candidates without assigning a content
	// verdict, then retain only the rejected members' attributable findings.
	if err := closeUnverifiedCandidatesTx(ctx, tx, cohort.operation); err != nil {
		return err
	}
	copy := *p
	copy.versions = slices.Clone(p.versions)
	if err := publishPortableRejectionsTx(ctx, tx, &copy, cohort, a, completed, current); err != nil {
		return err
	}
	return finishAttemptTx(ctx, tx, a, completed, "content_rejected", failures)
}

func lockPortableAttemptTx(ctx context.Context, tx pgx.Tx, cohort *portableVerificationCohort, a executionAttempt) (*string, int64, error) {
	if a.OperationID != cohort.operation || a.Count != int64(len(cohort.order)) {
		return nil, 0, errors.New("reference pack: foreign portable attempt")
	}
	if _, err := lockPublicationDependenciesTx(ctx, tx, cohort.operation); err != nil {
		return nil, 0, err
	}
	var configuration string
	var current *string
	if err := tx.QueryRow(ctx, `SELECT configuration_sha256,pack_set_id FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&configuration, &current); err != nil {
		return nil, 0, err
	}
	if configuration != cohort.context.ConfigurationSHA256 {
		return nil, 0, &OperationRejection{Reason: "stale_admission_state"}
	}
	var live bool
	if err := tx.QueryRow(ctx, `SELECT completed_at IS NULL AND operation_id=$2 FROM reference_pack_attempts WHERE attempt_id=$1 FOR UPDATE`, a.ID, a.OperationID).Scan(&live); err != nil {
		return nil, 0, err
	}
	if !live {
		return nil, 0, errors.New("reference pack: superseded portable attempt")
	}
	var count, failures int64
	if err := tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE verdict='content_rejected') FROM reference_pack_attempt_members WHERE attempt_id=$1`, a.ID).Scan(&count, &failures); err != nil {
		return nil, 0, err
	}
	if count != a.Count {
		return nil, 0, errors.New("reference pack: incomplete portable cohort")
	}
	return current, failures, nil
}
