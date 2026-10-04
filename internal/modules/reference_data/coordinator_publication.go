package reference_data

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// publishAttemptTx is invoked only within the leased Extensions finalizer.
// The retained cohort is walked one row at a time, so memory is bounded by a
// single container's admitted metadata rather than the number of versions.
func (c *Coordinator) publishAttemptTx(ctx context.Context, tx pgx.Tx, a executionAttempt, completed time.Time) error {
	if err := lockPublicationTx(ctx, tx, a.OperationID, a.Frozen); err != nil {
		return err
	}
	var count, failures int64
	if err := tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE verdict='content_rejected') FROM reference_pack_attempt_members WHERE attempt_id=$1`, a.ID).Scan(&count, &failures); err != nil {
		return err
	}
	if count != a.Count {
		return errors.New("reference pack: incomplete prepared cohort")
	}
	// Merge all root proposals before publishing any member. Conflicting exact
	// root bytes abort this transaction, including every member's verdict.
	exactReimport := false
	for ordinal := int64(1); ordinal <= a.Count; ordinal++ {
		var data []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, ordinal).Scan(&data); err != nil {
			return err
		}
		if data == nil {
			continue
		}
		p, err := decodePrepared(data)
		if err != nil {
			return err
		}
		if err := publishRootsTx(ctx, tx, *p.Content.Manifest.Repository, *p.Envelope.TrustProposal); err != nil {
			return err
		}
	}
	for ordinal := int64(1); ordinal <= a.Count; ordinal++ {
		var data []byte
		var key, version, code, check *string
		if err := tx.QueryRow(ctx, `SELECT pack_key,pack_version,canonical_prepared,failure_code,check_id FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, ordinal).Scan(&key, &version, &data, &code, &check); err != nil {
			return err
		}
		if data != nil {
			p, err := decodePrepared(data)
			if err != nil {
				return err
			}
			if key == nil || version == nil || p.Envelope.PackKey != *key || p.Envelope.PackVersion != *version ||
				p.Envelope.OperationID != a.OperationID || !p.Envelope.VerifiedAt.Equal(a.Start) {
				return errors.New("reference pack: prepared attempt binding mismatch")
			}
			if a.Frozen.Kind == "import" {
				if err := tx.QueryRow(ctx, `SELECT removed FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, *key, *version).Scan(&exactReimport); err != nil {
					return err
				}
			}
			if _, err := publishVersionTx(ctx, tx, p, a.OperationID, a.Actor); err != nil {
				return err
			}
			if err := publishMetadataTx(ctx, tx, *p.Content.Manifest.Repository, *p.Envelope.TrustProposal, *key, *version); err != nil {
				return err
			}
		} else if key != nil {
			// This is the first committed content loss in the current health
			// cycle, not merely another failed verification. Record the fact in
			// this transaction so its attestation can name the complete final set.
			if a.Frozen.Kind != "import" && slices.Contains([]string{"payload_missing", "target_length_mismatch", "checksum_mismatch", "content_schema_invalid", "content_semantic_invalid", "disallowed_content", "contract_incompatible"}, *code) {
				if _, err := tx.Exec(ctx, `UPDATE reference_pack_attempt_members m SET invalidated_content=true FROM reference_pack_candidates c WHERE m.attempt_id=$1 AND m.ordinal=$2 AND c.pack_key=m.pack_key AND c.pack_version=m.pack_version AND c.current_envelope_id IS NOT NULL AND NOT c.removed AND (c.health='verified_available' OR (c.health='failed' AND c.last_failure_code='metadata_expired'))`, a.ID, ordinal); err != nil {
					return err
				}
			}
			health := "failed"
			var missing *string
			if *code == "payload_missing" {
				health = "missing"
				reason := "storage_loss"
				if a.Frozen.Kind == "import" {
					reason = "staging_loss"
				}
				missing = &reason
			}
			// Import into an established logical version is a renewal. Its failed
			// attempt does not change health, disablement, removal or trust.
			changed, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health=$3,last_failure_code=$4,missing_reason=$5 WHERE pack_key=$1 AND pack_version=$2 AND NOT removed AND ($6 <> 'import' OR current_envelope_id IS NULL)`, *key, *version, health, *code, missing, a.Frozen.Kind)
			if err != nil {
				return err
			}
			if changed.RowsAffected() > 0 {
				if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, *key); err != nil {
					return err
				}
			}
		}
	}
	if err := publishFallbackTx(ctx, tx, a, completed); err != nil {
		return err
	}
	var setID *string
	if err := tx.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&setID); err != nil {
		return err
	}
	for ordinal := int64(1); ordinal <= a.Count; ordinal++ {
		var key, version, code, check *string
		var summaryID *string
		var prepared []byte
		var invalidated bool
		if err := tx.QueryRow(ctx, `SELECT pack_key,pack_version,failure_code,check_id,canonical_prepared,invalidated_content,validation_summary_id FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, a.ID, ordinal).Scan(&key, &version, &code, &check, &prepared, &invalidated, &summaryID); err != nil {
			return err
		}
		if code != nil {
			auditKind := ""
			switch *code {
			case "tuf_root_untrusted", "tuf_root_rotation_invalid":
				auditKind = "trust_root_rejection"
			case "pack_release_sequence_rollback", "pack_release_sequence_collision":
				auditKind = "sequence_rejection"
			}
			if auditKind != "" {
				k, v := "", ""
				if key != nil && version != nil {
					k, v = *key, *version
				}
				if err := appendPackAuditTx(ctx, tx, a.OperationID, auditKind, "rejected", completed, k, v, code); err != nil {
					return err
				}
			}
		}
		if key == nil {
			continue
		}
		verdict := "succeeded"
		if code != nil {
			verdict = "failed"
		}
		kind := map[string]string{"import": "import_verification", "reverify": "reverification", "refresh": "refresh_verification"}[a.Frozen.Kind]
		container := a.Frozen.ContainerSHA256
		if a.Frozen.Kind != "import" && (check == nil || *check != "retained_payload") {
			if err := tx.QueryRow(ctx, `SELECT e.container_sha256 FROM reference_pack_operation_members m JOIN reference_pack_envelopes e ON e.envelope_id=m.envelope_id WHERE m.operation_id=$1 AND m.ordinal=$2`, a.OperationID, ordinal).Scan(&container); err != nil {
				return err
			}
		}
		if check != nil && *check == "retained_payload" {
			container = nil
		}
		if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: a.OperationID, Key: *key, Version: *version, Kind: kind, Result: verdict, At: completed, PreviousSet: a.Frozen.previousSetID(), ResultingSet: setID, Container: container, Verification: true, Summary: summaryID}); err != nil {
			return err
		}
		if invalidated {
			if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: a.OperationID, Key: *key, Version: *version, Kind: "payload_invalidation", Result: "failed", At: completed, PreviousSet: a.Frozen.previousSetID(), ResultingSet: setID}); err != nil {
				return err
			}
		}
		if exactReimport {
			if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: a.OperationID, Key: *key, Version: *version, Kind: "exact_reimport", Result: "succeeded", At: completed, PreviousSet: a.Frozen.previousSetID(), ResultingSet: setID}); err != nil {
				return err
			}
		}
		if prepared != nil {
			p, err := decodePrepared(prepared)
			if err != nil {
				return err
			}
			if len(p.Envelope.TrustProposal.RootTransitions) > 0 {
				if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: a.OperationID, Key: *key, Version: *version, Kind: "trust_root_update", Result: "succeeded", At: completed, PreviousSet: a.Frozen.previousSetID(), ResultingSet: setID}); err != nil {
					return err
				}
			}
		}
	}
	outcome := "succeeded"
	if failures > 0 {
		outcome = "content_rejected"
	}
	return finishAttemptTx(ctx, tx, a, completed, outcome, failures)
}

func loadSelectionTx(ctx context.Context, tx pgx.Tx, setID string) ([]packstate.Version, []packstate.Version, error) {
	var data []byte
	if err := tx.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets WHERE pack_set_id=$1`, setID).Scan(&data); err != nil {
		return nil, nil, err
	}
	set, err := decodeRetainedSet(data)
	if err != nil {
		return nil, nil, err
	}
	selected := []packstate.Version{}
	for _, m := range set.Members {
		v, err := loadStateVersionTx(ctx, tx, m.Key, m.Version)
		if err != nil {
			return nil, nil, err
		}
		selected = append(selected, v)
	}
	rows, err := tx.Query(ctx, `SELECT b.pack_key,b.pack_version FROM reference_pack_release_bindings b JOIN reference_pack_current_set c USING(application_release_id) WHERE c.singleton ORDER BY b.pack_key COLLATE "C"`)
	if err != nil {
		return nil, nil, err
	}
	keys := [][2]string{}
	for rows.Next() {
		var pair [2]string
		if err := rows.Scan(&pair[0], &pair[1]); err != nil {
			rows.Close()
			return nil, nil, err
		}
		keys = append(keys, pair)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	base := []packstate.Version{}
	for _, pair := range keys {
		v, err := loadStateVersionTx(ctx, tx, pair[0], pair[1])
		if err != nil {
			return nil, nil, err
		}
		base = append(base, v)
	}
	return selected, base, nil
}
func publishFallbackTx(ctx context.Context, tx pgx.Tx, a executionAttempt, at time.Time) error {
	if a.Frozen.PackSetID == nil {
		return nil
	}
	selected, base, err := loadSelectionTx(ctx, tx, a.Frozen.previousSetID())
	if err != nil {
		return err
	}
	effective, _, err := packstate.ResolveEffectiveSet(selected, base)
	if errors.Is(err, packstate.ErrRequiredRegistryUnavailable) {
		if a.Frozen.Kind != "integrity" && a.Frozen.Kind != "reverify" && a.Frozen.Kind != "refresh" {
			return &OperationRejection{Reason: "required_registry_gap"}
		}
		// Keep immutable history and the definitive loss. No incomplete set is
		// valid, and no ready deployment can use the previous selection.
		_, err = tx.Exec(ctx, `UPDATE reference_pack_current_set SET pack_set_id=NULL,revision=revision+1 WHERE singleton AND pack_set_id IS NOT NULL`)
		return err
	}
	if err != nil {
		return err
	}
	members := make([]PackSetMember, 0, len(effective))
	for _, v := range effective {
		members = append(members, v.Member)
	}
	next, err := publishSetTx(ctx, tx, members, a.OperationID)
	if err != nil {
		return err
	}
	return publishSelectionConsequencesTx(ctx, tx, a.OperationID, a.Frozen.previousSetID(), next.ID, selected, effective, at, true)
}

func publishSelectionConsequencesTx(ctx context.Context, tx pgx.Tx, operation uuid.UUID, previous, next string, selected, effective []packstate.Version, at time.Time, safety bool) error {
	for _, old := range selected {
		index := slices.IndexFunc(effective, func(v packstate.Version) bool { return v.Member.Key == old.Member.Key })
		kind := ""
		member := old.Member
		if index < 0 && old.Health == packstate.Available && !old.Disabled && !old.Removed {
			kind = "dependency_invalidation"
		}
		if safety && index >= 0 && effective[index].Builtin && effective[index].Member != old.Member {
			kind = "safety_fallback"
			member = effective[index].Member
			if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET fallback_from_version=$3 WHERE pack_key=$1 AND pack_version=$2`, member.Key, member.Version, old.Member.Version); err != nil {
				return err
			}
		}
		if kind == "" {
			continue
		}
		if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operation, Key: member.Key, Version: member.Version, Kind: kind, Result: "succeeded", At: at, PreviousSet: previous, ResultingSet: &next}); err != nil {
			return err
		}
	}
	return nil
}

func finishAttemptTx(ctx context.Context, tx pgx.Tx, a executionAttempt, completed time.Time, outcome string, failures int64) error {
	data, err := packformat.EncodeAttemptResult(outcome, a.Count, failures)
	if err != nil {
		return err
	}
	updated, err := tx.Exec(ctx, `UPDATE reference_pack_attempts SET completed_at=$2,outcome=$3,canonical_result=$4 WHERE attempt_id=$1 AND completed_at IS NULL`, a.ID, completed, outcome, data)
	if err != nil {
		return err
	}
	if updated.RowsAffected() != 1 {
		return errors.New("reference pack: obsolete execution attempt")
	}
	updated, err = tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1 AND terminal_at IS NULL`, a.OperationID, completed, data)
	if err != nil {
		return err
	}
	if updated.RowsAffected() != 1 {
		return errors.New("reference pack: operation already terminal")
	}
	return appendPackAuditTx(ctx, tx, a.OperationID, "verification_completed", outcome, completed, "", "", nil)
}

func abortAttemptTx(ctx context.Context, tx pgx.Tx, a executionAttempt, completed time.Time, outcome string) error {
	if a.Frozen.Kind == "import" {
		if err := closeUnverifiedCandidatesTx(ctx, tx, a.OperationID); err != nil {
			return err
		}
	}
	return finishAttemptTx(ctx, tx, a, completed, outcome, 0)
}
