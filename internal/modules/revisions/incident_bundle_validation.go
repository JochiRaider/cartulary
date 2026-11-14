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
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func validatePreparedRevisionsBeforeWriteTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
	attributions incidentportability.AttributionRecorder,
) error {
	if err := validatePreparedRevisionsChangeSetMembership(ctx, prepared); err != nil {
		return err
	}
	// Resolve every record reference before admitting any history shape. A later
	// missing envelope outranks an earlier malformed snapshot across both files.
	if err := validatePreparedRevisionsRecordReferencesTx(ctx, tx, prepared, validation); err != nil {
		return err
	}
	if attributions == nil {
		return revisionsFailure(revisionsActorsInvariant)
	}
	if prepared.sequenceInvalid {
		return revisionsFailure(revisionsSequenceInvariant)
	}
	if prepared.revisionIdentityInvalid {
		return revisionsFailure(revisionsRecordVersionInvariant)
	}
	if err := validatePreparedRevisionsRecordVersions(ctx, prepared); err != nil {
		return err
	}

	if err := prepared.walkMutations(ctx, func(mutation portableChangeSetMutation) error {
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
			if mutation.AfterValue == nil {
				return revisionsFailure(revisionsHistoryInvariant)
			}
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
			if err := validation.validateSnapshot(prepared.incidentID, recordID, recordType, mutation.BeforeValue); err != nil {
				return revisionsFailure(revisionsHistoryInvariant)
			}
			if err := validation.validateSnapshot(prepared.incidentID, recordID, recordType, mutation.AfterValue); err != nil {
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
		if err := validation.validateSnapshot(prepared.incidentID, revision.RecordID, recordType, revision.BeforeJSON); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if err := validation.validateSnapshot(prepared.incidentID, revision.RecordID, recordType, revision.AfterJSON); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if !revisionSnapshotPairMatches(revision.BeforeJSON, revision.AfterJSON, revision.RowVersion) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Check each record's versions in revision-ID order without retaining all
// record identities or historical snapshots. A batch keeps only last versions.
func validatePreparedRevisionsRecordVersions(ctx context.Context, prepared preparedRevisionsImport) error {
	const batchLimit = 128
	versions := make(map[uuid.UUID]int64, batchLimit)
	flush := func() error {
		if len(versions) == 0 {
			return nil
		}
		if err := prepared.walkRevisions(ctx, func(row portableRecordRevision) error {
			previous, tracked := versions[row.RecordID]
			if !tracked {
				return nil
			}
			if row.RowVersion <= previous {
				return revisionsFailure(revisionsRecordVersionInvariant)
			}
			versions[row.RecordID] = row.RowVersion
			return nil
		}); err != nil {
			return err
		}
		clear(versions)
		return nil
	}
	if err := prepared.walkRevisions(ctx, func(row portableRecordRevision) error {
		versions[row.RecordID] = 0
		if len(versions) == batchLimit {
			return flush()
		}
		return nil
	}); err != nil {
		return err
	}
	return flush()
}

// Membership batches retain only bounded scalar identities. Every change-set
// scan reaches EOF so the admitted member's byte verification remains effective.
func validatePreparedRevisionsChangeSetMembership(ctx context.Context, prepared preparedRevisionsImport) error {
	const batchLimit = 128
	pending := make(map[uuid.UUID]struct{}, batchLimit)
	flush := func() error {
		if len(pending) == 0 {
			return nil
		}
		if err := prepared.walkChangeSets(ctx, func(row portableChangeSet) error {
			delete(pending, row.ChangeSetID)
			return nil
		}); err != nil {
			return err
		}
		if len(pending) != 0 {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		return nil
	}
	admit := func(changeSetID uuid.UUID) error {
		pending[changeSetID] = struct{}{}
		if len(pending) == batchLimit {
			return flush()
		}
		return nil
	}
	if err := prepared.walkMutations(ctx, func(row portableChangeSetMutation) error {
		return admit(row.ChangeSetID)
	}); err != nil {
		return err
	}
	if err := prepared.walkRevisions(ctx, func(row portableRecordRevision) error {
		return admit(row.ChangeSetID)
	}); err != nil {
		return err
	}
	return flush()
}

func validatePreparedRevisionsRecordReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
) error {
	checkRecord := func(recordID uuid.UUID) (string, error) {
		recordType, err := validation.envelopes.RecordTypeTx(ctx, tx, prepared.incidentID, recordID)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", revisionsFailure(revisionsReferencesInvariant)
		}
		if err != nil {
			return "", err
		}
		// Source owners have already applied in the catalog's dependency order.
		// An envelope alone does not establish an authoritative row reference.
		if _, err := validation.currentSourceRowTx(ctx, tx, recordID, recordType); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrUnexpectedHistoryTargetProvider) {
				return "", revisionsFailure(revisionsReferencesInvariant)
			}
			return "", err
		}
		return recordType, nil
	}
	if err := prepared.walkMutations(ctx, func(mutation portableChangeSetMutation) error {
		if !validation.resolvesTargetKind(mutation.TargetKind) ||
			(validation.targets.requiresGenericPortableTargetID(mutation.TargetKind) &&
				!canonicalPortableTargetID(mutation.TargetKind, mutation.TargetID)) {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		dispatch, err := validation.targets.dispatchClass(mutation.TargetKind)
		if err != nil {
			return err
		}
		if dispatch != rollbackcontract.DispatchRow {
			references, err := validation.targets.historyRecordReferences(mutation, prepared.incidentID)
			if err != nil {
				return revisionsFailure(revisionsReferencesInvariant)
			}
			for _, reference := range references {
				if reference.IncidentID != prepared.incidentID || reference.RecordID == uuid.Nil {
					return revisionsFailure(revisionsReferencesInvariant)
				}
				if _, err := checkRecord(reference.RecordID); err != nil {
					return err
				}
			}
			return nil
		}
		recordID, err := uuid.Parse(mutation.TargetID)
		if err != nil {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		recordType, err := checkRecord(recordID)
		if err != nil {
			return err
		}
		if _, err := validation.targets.rowProvider(mutation.TargetKind, recordType); err != nil {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		return nil
	}); err != nil {
		return err
	}
	return prepared.walkRevisions(ctx, func(revision portableRecordRevision) error {
		_, err := checkRecord(revision.RecordID)
		return err
	})
}

func validatePreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
	attributions incidentportability.AttributionRecorder,
) error {
	if err := validatePreparedRevisionsRecordReferencesTx(ctx, tx, prepared, validation); err != nil {
		return err
	}
	if err := validatePreparedRevisionsSourceStateTx(ctx, tx, prepared, validation); err != nil {
		return err
	}
	if err := prepared.validateAttributions(ctx, attributions); err != nil {
		return err
	}
	if err := validation.envelopes.VisitIncidentRecordIDsTx(ctx, tx, prepared.incidentID, func(recordID uuid.UUID) error {
		var retained bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM record_revisions WHERE record_id=$1)`, recordID).Scan(&retained); err != nil {
			return err
		}
		if !retained {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.validateActorContextPaths(ctx); err != nil {
		return err
	}
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

func snapshotActorsAdmitted(snapshot any, actors sourceport.ActorCatalog) bool {
	return walkSnapshotActors(snapshot, func(actorID string) bool {
		if _, ok := canonicalRevisionsUUID(actorID); !ok {
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

// The database provides record/version ordering. Only adjacent history records
// are retained, so a long revision chain cannot expand preparation memory.
func validateImportedRevisionChainsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	rows, err := tx.Query(ctx, `
 SELECT revision.record_id,revision.row_version,revision.before_json,revision.after_json,changes.actor_import_contexts,changes.history_order
 FROM record_revisions revision JOIN change_sets changes USING(change_set_id)
 WHERE changes.incident_id=$1
 ORDER BY revision.record_id,revision.row_version
 `, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var previousID uuid.UUID
	var previousVersion, previousOrder int64
	var previousAfter map[string]any
	var previousPath []revisionActorImportContext
	for rows.Next() {
		var recordID uuid.UUID
		var version, historyOrder int64
		var beforeRaw, afterRaw, contextsRaw []byte
		if err := rows.Scan(&recordID, &version, &beforeRaw, &afterRaw, &contextsRaw, &historyOrder); err != nil {
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
		path, err := decodeRevisionActorContexts(contextsRaw)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if !revisionSnapshotPairMatches(before, after, version) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if recordID != previousID {
			if version != 1 {
				return revisionsFailure(revisionsHistoryInvariant)
			}
		} else if historyOrder < previousOrder || version != previousVersion+1 || !revisionSnapshotsContinue(previousAfter, before, previousPath, path) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		previousID, previousVersion, previousOrder, previousAfter, previousPath = recordID, version, historyOrder, after, path
	}
	return rows.Err()
}
