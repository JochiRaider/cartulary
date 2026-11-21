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
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// validatePreparedRevisionsBeforeWriteTx reports, in the owner's precedence,
// the verdicts preparation recorded and the database-side reference checks.
// It opens no bundle member (Core 01 REQ-01-641).
func validatePreparedRevisionsBeforeWriteTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	store incidentportability.MemberReader,
	validation *incidentBundleValidationCatalog,
	attributions incidentportability.AttributionRecorder,
) error {
	if err := prepared.verdicts.At(rankMembership); err != nil {
		return err
	}
	// Resolve every record reference before admitting any history shape. A later
	// missing envelope outranks an earlier malformed snapshot across both files.
	if err := prepared.verdicts.At(rankBundleReferences); err != nil {
		return err
	}
	outcome, err := prepared.checkReferencesTx(ctx, tx, store, validation, nil)
	if err != nil {
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
	if err := prepared.verdicts.At(rankRecordVersions); err != nil {
		return err
	}
	if err := prepared.verdicts.At(rankHistoryShapes); err != nil || outcome.typeMismatch {
		return revisionsFailure(revisionsHistoryInvariant)
	}
	return nil
}

func validatePreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *incidentBundleValidationCatalog,
	input sourceport.ImportContext,
) error {
	store := input.Preparation
	if store == nil || prepared.verdicts == nil || validation == nil {
		return errors.New("revisions: incident bundle preparation is required")
	}
	// References rank first, bundle-only verdicts included. One pass resolves
	// each distinct referenced record and compares its terminal revision with
	// the current row; that verdict keeps its own rank.
	if err := prepared.verdicts.At(rankBundleReferences); err != nil {
		return err
	}
	terminal := &terminalRevisionCursor{ctx: ctx, tx: tx, incidentID: prepared.incidentID}
	outcome, err := prepared.checkReferencesTx(ctx, tx, store, validation, terminal)
	if err != nil {
		return err
	}
	if err := validatePreparedRevisionsSourceStateTx(ctx, tx, prepared, store); err != nil {
		return err
	}
	// Attribution is flushed in step 9 and checked by counts there.
	if input.Attributions == nil {
		return revisionsFailure(revisionsActorsInvariant)
	}
	if err := validateIncidentHistoryCoverageTx(ctx, tx, validation, prepared.incidentID); err != nil {
		return err
	}
	if err := prepared.verdicts.At(rankActorContexts); err != nil {
		return err
	}
	if err := validateImportedRevisionChainsTx(ctx, tx, prepared.incidentID); err != nil {
		return err
	}
	if outcome.terminalMismatch {
		return revisionsFailure(revisionsHistoryInvariant)
	}
	return nil
}

// validateIncidentHistoryCoverageTx requires retained history for every
// incident record with one aggregate read per page of record identities.
func validateIncidentHistoryCoverageTx(ctx context.Context, tx pgx.Tx, validation *incidentBundleValidationCatalog, incidentID uuid.UUID) error {
	const pageSize = 128
	page := make([]uuid.UUID, 0, pageSize)
	flush := func() error {
		if len(page) == 0 {
			return nil
		}
		var covered int
		if err := tx.QueryRow(ctx, `SELECT count(DISTINCT record_id) FROM record_revisions WHERE record_id = ANY($1::uuid[])`, page).Scan(&covered); err != nil {
			return err
		}
		if covered != len(page) {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		page = page[:0]
		return nil
	}
	if err := validation.envelopes.VisitIncidentRecordIDsTx(ctx, tx, incidentID, func(recordID uuid.UUID) error {
		page = append(page, recordID)
		if len(page) == pageSize {
			return flush()
		}
		return nil
	}); err != nil {
		return err
	}
	return flush()
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
		var version int64
		var storedOrder pgtype.Int8
		var beforeRaw, afterRaw, contextsRaw []byte
		if err := rows.Scan(&recordID, &version, &beforeRaw, &afterRaw, &contextsRaw, &storedOrder); err != nil {
			return err
		}
		// Imported history has explicit orders; native history appended inside
		// the import transaction has none and cannot reconstruct the bundle.
		if !storedOrder.Valid {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		historyOrder := storedOrder.Int64
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
