package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func validatePreparedRevisionsBeforeWriteTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
) error {
	for _, mutation := range prepared.mutations {
		if !validation.resolvesTargetKind(mutation.TargetKind) ||
			(validation.targets.requiresGenericPortableTargetID(mutation.TargetKind) &&
				!canonicalPortableTargetID(mutation.TargetKind, mutation.TargetID)) {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if _, err := validation.targets.DescribeValues(
			mutation.TargetKind,
			mutation.TargetID,
			mutation.OperationKind,
			mutation.BeforeValue,
			mutation.AfterValue,
		); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		dispatch, err := validation.targets.dispatchClass(mutation.TargetKind)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if dispatch == rollbackcontract.DispatchRow {
			recordID, err := uuid.Parse(mutation.TargetID)
			if err != nil {
				return revisionsFailure(revisionsHistoryInvariant)
			}
			recordType, err := validation.envelopes.RecordTypeTx(ctx, tx, prepared.incidentID, recordID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return revisionsFailure(revisionsReferencesInvariant)
				}
				return err
			}
			if err := validation.validateSnapshot(recordID, recordType, mutation.BeforeValue); err != nil {
				return revisionsFailure(revisionsHistoryInvariant)
			}
			if err := validation.validateSnapshot(recordID, recordType, mutation.AfterValue); err != nil {
				return revisionsFailure(revisionsHistoryInvariant)
			}
		}
	}
	for _, revision := range prepared.revisions {
		recordType, err := validation.envelopes.RecordTypeTx(
			ctx,
			tx,
			prepared.incidentID,
			revision.RecordID,
		)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if err := validation.validateSnapshot(revision.RecordID, recordType, revision.BeforeJSON); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if err := validation.validateSnapshot(revision.RecordID, recordType, revision.AfterJSON); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
	}
	if !mutationSequencesContiguous(prepared.mutations) {
		return revisionsFailure(revisionsSequenceInvariant)
	}
	if !recordRevisionIdentitiesUnique(prepared.revisions) {
		return revisionsFailure(revisionsRecordVersionInvariant)
	}
	for _, revision := range prepared.revisions {
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_revisions
     WHERE revision_id = $1
        OR (record_id = $2 AND row_version = $3)
)
`, revision.RevisionID, revision.RecordID, revision.RowVersion).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return revisionsFailure(revisionsRecordVersionInvariant)
		}
	}
	if !recordRevisionChainsCoherent(prepared.revisions) {
		return revisionsFailure(revisionsHistoryInvariant)
	}
	return nil
}

func validatePreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
) error {
	terminal := map[uuid.UUID]portableRecordRevision{}
	for _, revision := range prepared.revisions {
		prior, exists := terminal[revision.RecordID]
		if !exists || revision.RowVersion > prior.RowVersion {
			terminal[revision.RecordID] = revision
		}
	}
	recordIDs := make([]uuid.UUID, 0, len(terminal))
	for recordID := range terminal {
		recordIDs = append(recordIDs, recordID)
	}
	sort.Slice(recordIDs, func(i, j int) bool {
		return recordIDs[i].String() < recordIDs[j].String()
	})
	for _, recordID := range recordIDs {
		expected := normalizePortableSnapshotActors(
			terminal[recordID].AfterJSON,
			prepared.actors,
			prepared.runtimeActor,
		)
		matches, err := validation.currentHistoryRowMatchesTx(
			ctx,
			tx,
			prepared.incidentID,
			recordID,
			expected,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return revisionsFailure(revisionsReferencesInvariant)
			}
			return err
		}
		if !matches {
			return revisionsFailure(revisionsHistoryInvariant)
		}
	}
	return nil
}

func selectedRevisionsFailure(failures []revisionsParseFailure) error {
	sort.Slice(failures, func(i, j int) bool {
		left, right := failures[i], failures[j]
		if revisionsInvariantRank(left.invariant) != revisionsInvariantRank(right.invariant) {
			return revisionsInvariantRank(left.invariant) < revisionsInvariantRank(right.invariant)
		}
		if left.path != right.path {
			return left.path < right.path
		}
		return left.identity < right.identity
	})
	return revisionsFailure(failures[0].invariant)
}

func revisionsFailure(invariantID string) error {
	return revisionsIncidentBundleDescriptor().DeclaredFailure(invariantID)
}

func revisionsInvariantRank(invariantID string) int {
	switch invariantID {
	case revisionsReferencesInvariant:
		return 1
	case revisionsActorsInvariant:
		return 2
	case revisionsSequenceInvariant:
		return 3
	case revisionsRecordVersionInvariant:
		return 4
	case revisionsHistoryInvariant:
		return 5
	case revisionsSequenceRepairInvariant:
		return 6
	default:
		return 100
	}
}

func sameCanonicalJSON(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func snapshotActorsAdmitted(snapshot map[string]any, actors sourceport.ActorCatalog) bool {
	return walkSnapshotActors(snapshot, func(actorID string) bool {
		parsed, err := uuid.Parse(actorID)
		if err != nil || parsed.String() != actorID {
			return false
		}
		_, admitted := actors.Lookup(actorID)
		return admitted
	})
}

func walkSnapshotActors(value any, validate func(string) bool) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			member := typed[key]
			if strings.HasSuffix(key, "_user_id") && member != nil {
				actorID, ok := member.(string)
				if !ok || !validate(actorID) {
					return false
				}
			}
			if !walkSnapshotActors(member, validate) {
				return false
			}
		}
	case []any:
		for _, item := range typed {
			if !walkSnapshotActors(item, validate) {
				return false
			}
		}
	}
	return true
}

func normalizePortableSnapshotActors(
	snapshot map[string]any,
	actors sourceport.ActorCatalog,
	runtimeActor uuid.UUID,
) map[string]any {
	if snapshot == nil {
		return nil
	}
	return normalizePortableActorValue(snapshot, actors, runtimeActor).(map[string]any)
}

func normalizePortableActorValue(
	value any,
	actors sourceport.ActorCatalog,
	runtimeActor uuid.UUID,
) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, member := range typed {
			if strings.HasSuffix(key, "_user_id") {
				if actorID, ok := member.(string); ok {
					if _, admitted := actors.Lookup(actorID); admitted {
						result[key] = runtimeActor.String()
						continue
					}
				}
			}
			result[key] = normalizePortableActorValue(member, actors, runtimeActor)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = normalizePortableActorValue(item, actors, runtimeActor)
		}
		return result
	default:
		return value
	}
}
