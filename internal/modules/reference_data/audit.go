package reference_data

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Audit is published in the same owner transaction as its semantic effect.
// Only registered tokens and opaque identities enter either audit representation;
// user reasons, source identifiers, payloads and storage locations are excluded.
func appendPackAuditTx(ctx context.Context, tx pgx.Tx, operation uuid.UUID, kind, result string, at time.Time, key, version string, reason *string) error {
	if !slices.Contains([]string{"admitted", "succeeded", "failed", "rejected", "content_rejected", "execution_failed", "canceled", "timed_out", "stale_state"}, result) {
		return errors.New("reference pack: invalid audit outcome")
	}
	if reason != nil {
		known := false
		for _, check := range packformat.Checks() {
			known = known || check.Code == *reason
		}
		if !known {
			return errors.New("reference pack: unsafe audit reason")
		}
	}
	var actorKind string
	var actor *uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT actor_kind,actor_user_id FROM reference_pack_operations WHERE operation_id=$1`, operation).Scan(&actorKind, &actor); err != nil {
		return err
	}
	source := administrativeaudit.SourceSystem
	switch actorKind {
	case "user":
		source = administrativeaudit.SourceAPI
	case "local_operator":
		actorKind = administrativeaudit.ActorOperator
		source = administrativeaudit.SourceOperator
	case "application_release":
		actorKind = administrativeaudit.ActorSystem
		source = administrativeaudit.SourceStartup
	case "system":
		if kind == "root_import" {
			source = administrativeaudit.SourceStartup
		}
	default:
		return errors.New("reference pack: invalid audit actor")
	}
	var reference *string
	if key != "" && version != "" {
		value := versionObjectID(key, version)
		reference = &value
	}
	target := operation.String()
	safe := map[string]any{"operation_id": target, "event_kind": kind, "result": result, "pack_reference": reference}
	_, err := administrativeaudit.AppendTx(ctx, tx,
		administrativeaudit.RawEvent{ActorUserID: actor, EventSource: "reference_pack", EventKind: kind, ReasonCode: reason, ClientTxnID: &target, After: safe, OccurredAt: at},
		administrativeaudit.Event{ScopeKind: administrativeaudit.ScopeDeployment, OccurredAt: at, ActorKind: actorKind, ActorUserID: actor, Source: source, ActionCode: "reference_pack_" + kind, TargetKind: "reference_pack_operation", TargetID: &target, ReasonCode: reason, Changes: []administrativeaudit.Change{
			administrativeaudit.Visible("event_kind", nil, kind), administrativeaudit.Visible("result", nil, result), administrativeaudit.Visible("pack_reference", nil, reference),
		}})
	return err
}
