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
	if err := prepared.walkMutations(ctx, func(mutation portableChangeSetMutation) error {
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
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.walkRevisions(ctx, func(revision portableRecordRevision) error {
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
		return nil
	}); err != nil {
		return err
	}
	if prepared.sequenceInvalid {
		return revisionsFailure(revisionsSequenceInvariant)
	}
	if prepared.revisionIdentityInvalid {
		return revisionsFailure(revisionsRecordVersionInvariant)
	}
	return nil
}
func validatePreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
) error {

	if err := validateImportedRevisionChainsTx(ctx, tx, prepared.incidentID); err != nil {
		return err
	}
	return prepared.walkRevisions(ctx, func(revision portableRecordRevision) error {
		var terminal bool
		if err := tx.QueryRow(ctx, `SELECT NOT EXISTS (SELECT 1 FROM record_revisions WHERE record_id=$1 AND row_version>$2)`, revision.RecordID, revision.RowVersion).Scan(&terminal); err != nil {
			return err
		}
		if !terminal {
			return nil
		}
		recordID := revision.RecordID

		expected := normalizePortableSnapshotActors(
			revision.AfterJSON,
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
		return nil
	})
}

func revisionsFailure(invariantID string) error {
	return revisionsIncidentBundleDescriptor().DeclaredFailure(invariantID)
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

func validateImportedChangeSetTx(ctx context.Context, tx pgx.Tx, incidentID, changeSetID uuid.UUID) error {
	var present bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM change_sets WHERE incident_id=$1 AND change_set_id=$2)`, incidentID, changeSetID).Scan(&present); err != nil {
		return err
	}
	if !present {
		return revisionsFailure(revisionsReferencesInvariant)
	}
	return nil
}

// The database provides record/version ordering. Only adjacent history records
// are retained, so a long revision chain cannot expand preparation memory.
func validateImportedRevisionChainsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	rows, err := tx.Query(ctx, `
 SELECT revision.record_id,revision.row_version,revision.before_json,revision.after_json
 FROM record_revisions revision JOIN change_sets changes USING(change_set_id)
 WHERE changes.incident_id=$1
 ORDER BY revision.record_id,revision.row_version
 `, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var previousID uuid.UUID
	var previousVersion int64
	var previousAfter map[string]any
	for rows.Next() {
		var recordID uuid.UUID
		var version int64
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(&recordID, &version, &beforeRaw, &afterRaw); err != nil {
			return err
		}
		before, err := decodeDatabaseJSONObject(beforeRaw)
		if err != nil {
			return err
		}
		after, err := decodeDatabaseJSONObject(afterRaw)
		if err != nil {
			return err
		}
		if recordID == previousID && (version != previousVersion+1 || !sameStoredCanonicalJSON(previousAfter, before)) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		previousID, previousVersion, previousAfter = recordID, version, after
	}
	return rows.Err()
}
