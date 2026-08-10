package revisions

import (
	"context"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func mutationSequencesContiguous(rows []portableChangeSetMutation) bool {
	byChangeSet := map[uuid.UUID][]int{}
	for _, row := range rows {
		byChangeSet[row.ChangeSetID] = append(byChangeSet[row.ChangeSetID], row.SequenceNo)
	}
	for _, sequence := range byChangeSet {
		sort.Ints(sequence)
		for index, value := range sequence {
			if value != index+1 {
				return false
			}
		}
	}
	return true
}

func recordRevisionIdentitiesUnique(rows []portableRecordRevision) bool {
	seenRevisionIDs := map[int64]struct{}{}
	seenRecordVersions := map[string]struct{}{}
	byRecord := map[uuid.UUID][]int64{}
	for _, row := range rows {
		if _, duplicate := seenRevisionIDs[row.RevisionID]; duplicate {
			return false
		}
		seenRevisionIDs[row.RevisionID] = struct{}{}
		identity := row.RecordID.String() + ":" + strconv.FormatInt(row.RowVersion, 10)
		if _, duplicate := seenRecordVersions[identity]; duplicate {
			return false
		}
		seenRecordVersions[identity] = struct{}{}
		byRecord[row.RecordID] = append(byRecord[row.RecordID], row.RowVersion)
	}
	for _, versions := range byRecord {
		sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
		for index := 1; index < len(versions); index++ {
			if versions[index] <= versions[index-1] {
				return false
			}
		}
	}
	return true
}

func recordRevisionChainsCoherent(rows []portableRecordRevision) bool {
	byRecord := map[uuid.UUID][]portableRecordRevision{}
	for _, row := range rows {
		byRecord[row.RecordID] = append(byRecord[row.RecordID], row)
	}
	for _, revisions := range byRecord {
		sort.Slice(revisions, func(i, j int) bool {
			return revisions[i].RowVersion < revisions[j].RowVersion
		})
		for index := 1; index < len(revisions); index++ {
			if revisions[index].RowVersion != revisions[index-1].RowVersion+1 ||
				!sameStoredCanonicalJSON(revisions[index-1].AfterJSON, revisions[index].BeforeJSON) {
				return false
			}
		}
	}
	return true
}

func BeginIncidentBundleImportedRevisionSequenceTx(
	ctx context.Context,
	tx pgx.Tx,
) (int64, error) {
	var originalNext int64
	if err := tx.QueryRow(ctx, `
SELECT public.revisions_incident_bundle_sequence_begin_v1()
`).Scan(&originalNext); err != nil {
		return 0, err
	}
	return originalNext, nil
}

func FinishIncidentBundleImportedRevisionSequenceTx(
	ctx context.Context,
	tx pgx.Tx,
	originalNext int64,
) error {
	var repairedNext int64
	return tx.QueryRow(ctx, `
SELECT public.revisions_incident_bundle_sequence_finish_v1($1)
`, originalNext).Scan(&repairedNext)
}
