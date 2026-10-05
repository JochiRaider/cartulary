package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RegistryUsageReader is implemented by participating source owners. Reads
// occur under the registry usage guard; writers must acquire that same guard
// before changing assignments. Reference Data never reads their tables.
type RegistryUsageReader interface {
	ReferencedRegistryEntriesTx(context.Context, pgx.Tx, string) ([]string, error)
}

func (c *Coordinator) Activate(ctx context.Context, p ActionParams) (ActionResult, error) {
	return c.applyAdministrativeAction(ctx, p, "activate")
}
func (c *Coordinator) Disable(ctx context.Context, p ActionParams) (ActionResult, error) {
	return c.applyAdministrativeAction(ctx, p, "disable")
}
func (c *Coordinator) Remove(ctx context.Context, p ActionParams) (ActionResult, error) {
	if p.Request.Reason == nil || *p.Request.Reason == "" {
		return ActionResult{}, &RequestRejection{Field: "reason", Reason: "missing_required_field"}
	}
	return c.applyAdministrativeAction(ctx, p, "remove")
}

func (c *Coordinator) applyAdministrativeAction(ctx context.Context, p ActionParams, kind string) (result ActionResult, resultErr error) {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack."+kind)
	defer func() { end(referenceOutcome(resultErr)) }()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.limits.ReferencePacks.MaxVerificationSeconds)*time.Second)
	defer cancel()
	at := c.now().UTC()
	var activationKeyRevision, activationSetRevision int64
	var activationIntegrity error
	if kind == "activate" {
		// Hash immutable retained members outside every publication lock. The
		// relevant revisions are rechecked after acquiring the short boundary.
		activationIntegrity = c.pool.QueryRow(ctx, `SELECT k.revision,s.revision FROM reference_pack_candidates c JOIN reference_pack_key_state k USING(pack_key) CROSS JOIN reference_pack_current_set s WHERE c.pack_key=$1 AND c.pack_version=$2 AND c.current_envelope_id IS NOT NULL AND s.singleton`, p.PackKey, p.PackVersion).Scan(&activationKeyRevision, &activationSetRevision)
		if activationIntegrity == nil {
			activationIntegrity = c.checkRetainedMembers(ctx, frozenMember{Key: p.PackKey, Version: p.PackVersion})
		}
	}
	if p.ActorUserID == uuid.Nil || p.Request.ClientTxnID == "" {
		return ActionResult{}, errors.New("reference pack: invalid action attribution")
	}
	// The receipt binds the exact path and normalized request, including reason.
	raw, err := canonicaljson.Marshal(map[string]any{"kind": kind, "pack_key": p.PackKey, "pack_version": p.PackVersion, "request": json.RawMessage(p.Request.Normalized)})
	if err != nil {
		return ActionResult{}, err
	}
	key := receiptKey{RouteKey: "reference_packs." + kind, ActorUserID: p.ActorUserID, ScopeKey: p.PackKey + ":" + p.PackVersion, ClientTxnID: p.Request.ClientTxnID}
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ActionResult{}, err
	}
	defer tx.Rollback(ctx)
	if err := lockReceiptTx(ctx, tx, key); err != nil {
		return ActionResult{}, err
	}
	if receipt, err := readReceipt(ctx, tx, key); err == nil {
		if !bytes.Equal(receipt.RequestHash, hashBytes(raw)) {
			return ActionResult{}, ErrClientTxnConflict
		}
		var payload actionReceipt
		if err := json.Unmarshal(receipt.ResponseJSON, &payload); err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Version: payload.Version, Replayed: true}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return ActionResult{}, err
	}
	operation := uuid.New()
	frozen, err := admitOperationTx(ctx, tx, operationAdmission{ID: operation, Kind: kind, ActorKind: "user", Actor: &p.ActorUserID, At: at, Keys: []string{p.PackKey}, Version: p.PackVersion, ClockTrusted: c.configuration.ClockTrusted, TimeoutSeconds: c.limits.ReferencePacks.MaxVerificationSeconds})
	if err != nil {
		return ActionResult{}, err
	}
	v, err := scanAdministrativeVersion(tx.QueryRow(ctx, administrativeVersionSelect+` WHERE c.pack_key=$1 AND c.pack_version=$2`, p.PackKey, p.PackVersion))
	if err != nil {
		return ActionResult{}, err
	}
	if v.Removed {
		return ActionResult{}, &OperationRejection{Reason: "removed"}
	}
	if (kind == "disable" || kind == "remove") && v.DistributionKind == "packaged_builtin" {
		return ActionResult{}, &OperationRejection{Reason: "packaged_builtin"}
	}
	var next PackSet
	activationKind := "activation"
	switch kind {
	case "activate":
		if v.Active {
			return ActionResult{}, &ActivationRejection{Reason: "already_active"}
		}
		if v.Health != "verified_available" || v.LastVerifiedAt == nil {
			return ActionResult{}, &ActivationRejection{Reason: "not_verified_available"}
		}
		if activationIntegrity != nil {
			return ActionResult{}, &ActivationRejection{Reason: "not_verified_available"}
		}
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, p.PackKey).Scan(&revision); err != nil {
			return ActionResult{}, err
		}
		if revision != activationKeyRevision || frozen.SetRevision != activationSetRevision {
			return ActionResult{}, &OperationRejection{Reason: "stale_admission_state"}
		}
		if v.DistributionKind == "operator_imported" {
			if !c.configuration.ClockTrusted {
				return ActionResult{}, &OperationRejection{Reason: "clock_untrusted"}
			}
			if v.TrustValidUntil == nil || !at.Before(*v.TrustValidUntil) {
				return ActionResult{}, &ActivationRejection{Reason: "metadata_expired"}
			}
		}
		target, err := loadStateVersionTx(ctx, tx, p.PackKey, p.PackVersion)
		if err != nil {
			return ActionResult{}, err
		}
		target.Disabled = false
		selected, base, err := loadSelectionTx(ctx, tx, frozen.previousSetID())
		if err != nil {
			return ActionResult{}, err
		}
		previousSelection := slices.Clone(selected)
		var previouslyActivated bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_events WHERE pack_key=$1 AND pack_version=$2 AND event_kind IN ('activation','rollback_activation') AND convert_from(canonical_attestation,'UTF8')::jsonb->>'result'='succeeded')`, p.PackKey, p.PackVersion).Scan(&previouslyActivated); err != nil {
			return ActionResult{}, err
		}
		if previouslyActivated {
			activationKind = "rollback_activation"
		}
		for _, old := range selected {
			if !target.Builtin && target.Manifest.Repository != nil && old.Manifest.Repository != nil && *old.Manifest.Repository == *target.Manifest.Repository && old.Member.Key == target.Member.Key && old.Manifest.Sequence > target.Manifest.Sequence {
				activationKind = "rollback_activation"
			}
		}
		if err := c.registryReplacementTx(ctx, tx, target, selected, base); err != nil {
			return ActionResult{}, err
		}
		for _, d := range target.Manifest.Dependencies {
			if !slices.ContainsFunc(selected, func(v packstate.Version) bool {
				return v.Member.Key == d.Key && v.Member.Version == d.Version && v.Member.PayloadSHA256 == d.SHA256
			}) {
				return ActionResult{}, &OperationRejection{Reason: "dependency_unsatisfied"}
			}
		}
		selected = slices.DeleteFunc(selected, func(v packstate.Version) bool { return v.Member.Key == p.PackKey })
		selected = append(selected, target)
		effective, _, err := packstate.ResolveEffectiveSet(selected, base)
		if errors.Is(err, packstate.ErrRequiredRegistryUnavailable) {
			return ActionResult{}, &OperationRejection{Reason: "required_registry_gap"}
		}
		if err != nil {
			return ActionResult{}, actionStateError(err)
		}
		if !slices.ContainsFunc(effective, func(v packstate.Version) bool { return v.Member == target.Member }) {
			return ActionResult{}, &OperationRejection{Reason: "dependency_unsatisfied"}
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET administratively_disabled=false WHERE pack_key=$1 AND pack_version=$2`, p.PackKey, p.PackVersion); err != nil {
			return ActionResult{}, err
		}
		members := make([]PackSetMember, 0, len(effective))
		for _, m := range effective {
			members = append(members, PackSetMember(m.Member))
		}
		next, err = publishSetTx(ctx, tx, members, operation)
		if err != nil {
			return ActionResult{}, err
		}
		if err := publishSelectionConsequencesTx(ctx, tx, operation, frozen.previousSetID(), next.ID, previousSelection, effective, at, false); err != nil {
			return ActionResult{}, err
		}
	case "disable":
		if v.Health != "verified_available" || v.AdministrativelyDisabled {
			return ActionResult{}, &OperationRejection{Reason: "not_disableable"}
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET administratively_disabled=true WHERE pack_key=$1 AND pack_version=$2`, p.PackKey, p.PackVersion); err != nil {
			return ActionResult{}, err
		}
		if err := publishFallbackTx(ctx, tx, executionAttempt{OperationID: operation, Frozen: frozen}, at); err != nil {
			return ActionResult{}, err
		}
	case "remove":
		if v.Active {
			return ActionResult{}, &OperationRejection{Reason: "active"}
		}
		var pinned bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_pins p JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key=$1 AND m.pack_version=$2) OR EXISTS(SELECT 1 FROM reference_pack_version_pins WHERE pack_key=$1 AND pack_version=$2)`, p.PackKey, p.PackVersion).Scan(&pinned); err != nil {
			return ActionResult{}, err
		}
		if pinned {
			return ActionResult{}, &OperationRejection{Reason: "pinned"}
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health='missing',removed=true,missing_reason='administrative_removal',current_index_id=NULL WHERE pack_key=$1 AND pack_version=$2`, p.PackKey, p.PackVersion); err != nil {
			return ActionResult{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, p.PackKey); err != nil {
		return ActionResult{}, err
	}
	if next.ID == "" {
		if err := tx.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&next.ID); err != nil {
			return ActionResult{}, err
		}
	}
	eventKind := map[string]string{"activate": activationKind, "disable": "disablement", "remove": "removal"}[kind]
	event, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operation, Key: p.PackKey, Version: p.PackVersion, Kind: eventKind, Result: "succeeded", At: at, PreviousSet: frozen.previousSetID(), ResultingSet: &next.ID})
	if err != nil {
		return ActionResult{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1`, operation, at, event); err != nil {
		return ActionResult{}, err
	}
	v, err = scanAdministrativeVersion(tx.QueryRow(ctx, administrativeVersionSelect+` WHERE c.pack_key=$1 AND c.pack_version=$2`, p.PackKey, p.PackVersion))
	if err != nil {
		return ActionResult{}, err
	}
	payload := actionReceipt{Version: v}
	if err := writeReceiptTx(ctx, tx, key, hashBytes(raw), receiptCompleted, payload); err != nil {
		return ActionResult{}, err
	}
	proof := func(proofCtx context.Context) (bool, error) {
		var terminal bool
		if err := c.pool.QueryRow(proofCtx, `SELECT terminal_at IS NOT NULL AND final_outcome=$2 FROM reference_pack_operations WHERE operation_id=$1`, operation, event).Scan(&terminal); err != nil {
			return false, err
		}
		if !terminal {
			return false, nil
		}
		receipt, err := readReceipt(proofCtx, c.pool, key)
		if err != nil {
			return false, err
		}
		if receipt.Outcome != receiptCompleted || !bytes.Equal(receipt.RequestHash, hashBytes(raw)) {
			return false, nil
		}
		expected, err := canonicaljson.Marshal(payload)
		if err != nil {
			return false, err
		}
		actual, err := canonicaljson.Canonicalize(receipt.ResponseJSON)
		if err != nil {
			return false, err
		}
		return bytes.Equal(expected, actual), nil
	}
	if err := c.actionFinalizer.FinalizeReferencePackAction(ctx, tx, proof); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{Version: payload.Version}, nil
}

func actionStateError(err error) error {
	var r *packstate.Rejection
	if errors.As(err, &r) {
		return &OperationRejection{Reason: r.Reason}
	}
	return err
}

func (c *Coordinator) registryReplacementTx(ctx context.Context, tx pgx.Tx, target packstate.Version, selected, base []packstate.Version) error {
	if !strings.HasPrefix(target.Member.Key, "type_registry.") {
		return nil
	}
	used, err := c.registryUsage.ReferencedRegistryEntriesTx(ctx, tx, target.Member.Key)
	if err != nil {
		return err
	}
	var targetIndex uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT current_index_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, target.Member.Key, target.Member.Version).Scan(&targetIndex); err != nil {
		return err
	}
	var missing bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM unnest($2::text[]) u(id) WHERE NOT EXISTS(SELECT 1 FROM reference_pack_indexes i WHERE i.index_id=$1 AND i.entry_kind='entry' AND i.entry_id=u.id))`, targetIndex, used).Scan(&missing); err != nil {
		return err
	}
	if missing {
		return &OperationRejection{Reason: "type_registry_incompatible"}
	}
	// Required Base entries and entries retained by pinned predecessor sets
	// must still resolve. Existing identities keep their semantic descriptors.
	for _, previous := range append(slices.Clone(selected), base...) {
		if previous.Member.Key != target.Member.Key {
			continue
		}
		var index uuid.UUID
		var pinned bool
		if err := tx.QueryRow(ctx, `SELECT current_index_id,EXISTS(SELECT 1 FROM reference_pack_pins p JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key=c.pack_key AND m.pack_version=c.pack_version) FROM reference_pack_candidates c WHERE c.pack_key=$1 AND c.pack_version=$2`, previous.Member.Key, previous.Member.Version).Scan(&index, &pinned); err != nil {
			return err
		}
		var incompatible bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_indexes old LEFT JOIN reference_pack_indexes new ON new.index_id=$1 AND new.entry_kind='entry' AND new.entry_id=old.entry_id WHERE old.index_id=$2 AND old.entry_kind='entry' AND ((new.entry_id IS NULL AND $3) OR (new.entry_id IS NOT NULL AND (
convert_from(old.canonical_item,'UTF8')::jsonb->'category' IS DISTINCT FROM convert_from(new.canonical_item,'UTF8')::jsonb->'category' OR
(convert_from(old.canonical_item,'UTF8')::jsonb->>'replacement_entry_id' IS NOT NULL AND convert_from(old.canonical_item,'UTF8')::jsonb->'replacement_entry_id' IS DISTINCT FROM convert_from(new.canonical_item,'UTF8')::jsonb->'replacement_entry_id') OR
NOT (coalesce(convert_from(new.canonical_item,'UTF8')::jsonb->'allowed_value_kinds','[]'::jsonb) @> coalesce(convert_from(old.canonical_item,'UTF8')::jsonb->'allowed_value_kinds','[]'::jsonb)) OR
(convert_from(old.canonical_item,'UTF8')::jsonb - ARRAY['entry_id','display_label','description','category','aliases','deprecated','replacement_entry_id','allowed_value_kinds','stix_mapping','icon_key','extensions']) IS DISTINCT FROM (convert_from(new.canonical_item,'UTF8')::jsonb - ARRAY['entry_id','display_label','description','category','aliases','deprecated','replacement_entry_id','allowed_value_kinds','stix_mapping','icon_key','extensions'])
))))`, targetIndex, index, previous.Builtin || pinned).Scan(&incompatible); err != nil {
			return err
		}
		if incompatible {
			return &OperationRejection{Reason: "type_registry_incompatible"}
		}
	}
	return nil
}
