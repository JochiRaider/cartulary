package merge

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func (s *Store) appendMergeIntentsTx(
	ctx context.Context,
	tx pgx.Tx,
	actorUserID uuid.UUID,
	clientTxnID string,
	result MergeResult,
	createdAt time.Time,
) error {
	if s == nil || s.ports.collaboration == nil {
		return errors.New("entity merge collaboration intent port is not configured")
	}

	viewSchemaID := entitycontract.IdentitiesViewSchemaID
	survivorKeys := []string{
		"identity.aad_object_id",
		"identity.aliases",
		"identity.edited_at",
		"identity.email",
		"identity.identity_state",
		"identity.reusable_identifiers",
		"identity.sam_account_name",
		"identity.sid",
		"identity.upn",
	}
	loserKeys := []string{"identity.edited_at", "identity.identity_state"}
	if result.RecordType == "host" {
		viewSchemaID = entitycontract.HostsViewSchemaID
		survivorKeys = []string{
			"host.aad_device_id",
			"host.aliases",
			"host.edited_at",
			"host.fqdn",
			"host.host_state",
			"host.hostname",
			"host.reusable_identifiers",
		}
		loserKeys = []string{"host.edited_at", "host.host_state"}
	}
	slices.Sort(survivorKeys)
	slices.Sort(loserKeys)

	changes := make([]collaboration.RecordChange, 0, 2+len(result.TimelineInvalidations))
	changes = append(changes,
		collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         result.SurvivorRecordID,
			RowVersion:       result.SurvivorRowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: survivorKeys,
			ViewSchemaID:     viewSchemaID,
		},
		collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         result.LoserRecordID,
			RowVersion:       result.LoserRowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: loserKeys,
			ViewSchemaID:     viewSchemaID,
			ChangeKind:       "remove",
		},
	)
	for _, invalidation := range result.TimelineInvalidations {
		changes = append(changes, collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
			ViewSchemaID:     timelineViewSchemaID,
		})
	}
	for ordinal, change := range changes {
		intent, err := collaboration.NewRecordChangeIntent(change, ordinal, createdAt)
		if err != nil {
			return err
		}
		if err := s.ports.collaboration.AppendIntentTx(ctx, tx, intent); err != nil {
			return err
		}
	}
	return nil
}
