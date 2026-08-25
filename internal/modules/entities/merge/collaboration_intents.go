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
	if result.RecordType == "host" {
		viewSchemaID = entitycontract.HostsViewSchemaID
	}
	survivorKeys := mergePublicFieldKeys(result.RecordType, true)
	loserKeys := mergePublicFieldKeys(result.RecordType, false)

	changes := make([]collaboration.RecordChangeIntentInput, 0, 2+len(result.TimelineInvalidations))
	changes = append(changes,
		collaboration.RecordChangeIntentInput{
			IncidentID:      result.IncidentID,
			RecordID:        result.SurvivorRecordID,
			RowVersion:      result.SurvivorRowVersion,
			ChangeSetID:     result.ChangeSetID,
			ClientTxnID:     clientTxnID,
			ActorUserID:     actorUserID,
			PublicFieldKeys: survivorKeys,
			AffectedViews:   []collaboration.AffectedViewChange{{ViewSchemaID: viewSchemaID, RecordID: result.SurvivorRecordID, RowVersion: result.SurvivorRowVersion, ChangeKind: "invalidate"}},
		},
		collaboration.RecordChangeIntentInput{
			IncidentID:      result.IncidentID,
			RecordID:        result.LoserRecordID,
			RowVersion:      result.LoserRowVersion,
			ChangeSetID:     result.ChangeSetID,
			ClientTxnID:     clientTxnID,
			ActorUserID:     actorUserID,
			MutationOrdinal: 1,
			PublicFieldKeys: loserKeys,
			AffectedViews:   []collaboration.AffectedViewChange{{ViewSchemaID: viewSchemaID, RecordID: result.LoserRecordID, RowVersion: result.LoserRowVersion, ChangeKind: "remove"}},
		},
	)
	for ordinal, invalidation := range result.TimelineInvalidations {
		changes = append(changes, collaboration.RecordChangeIntentInput{
			IncidentID:      result.IncidentID,
			RecordID:        invalidation.RecordID,
			RowVersion:      invalidation.RowVersion,
			ChangeSetID:     result.ChangeSetID,
			ClientTxnID:     clientTxnID,
			ActorUserID:     actorUserID,
			MutationOrdinal: ordinal + 2,
			PublicFieldKeys: invalidation.ChangedFieldKeys,
			AffectedViews:   []collaboration.AffectedViewChange{{ViewSchemaID: timelineViewSchemaID, RecordID: invalidation.RecordID, RowVersion: invalidation.RowVersion, ChangeKind: "invalidate"}},
		})
	}
	for index := range changes {
		changes[index].CreatedAt = createdAt
		if err := s.ports.collaboration.AppendRecordChangedTx(ctx, tx, changes[index]); err != nil {
			return err
		}
	}
	return nil
}

func mergePublicFieldKeys(recordType string, survivor bool) []string {
	var keys []string
	switch recordType {
	case "host":
		if survivor {
			keys = []string{
				"host.aad_device_id", "host.aliases", "host.edited_at", "host.fqdn",
				"host.host_state", "host.hostname", "host.reusable_identifiers",
			}
		} else {
			keys = []string{"host.edited_at", "host.host_state"}
		}
	case "identity":
		if survivor {
			keys = []string{
				"identity.aad_object_id", "identity.aliases", "identity.edited_at", "identity.email",
				"identity.identity_state", "identity.reusable_identifiers", "identity.sam_account_name",
				"identity.sid", "identity.upn",
			}
		} else {
			keys = []string{"identity.edited_at", "identity.identity_state"}
		}
	}
	slices.Sort(keys)
	return keys
}
