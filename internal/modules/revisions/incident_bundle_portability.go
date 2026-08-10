package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const (
	changeSetsBundlePath      = "data/change_sets.ndjson"
	changeSetMutationsPath    = "data/change_set_mutations.ndjson"
	recordRevisionsBundlePath = "data/record_revisions.ndjson"

	revisionsReferencesInvariant     = "revisions.references_complete"
	revisionsActorsInvariant         = "revisions.actor_references_complete"
	revisionsSequenceInvariant       = "revisions.mutation_sequence_contiguous"
	revisionsRecordVersionInvariant  = "revisions.record_version_unique"
	revisionsHistoryInvariant        = "revisions.history_reconstruction"
	revisionsSequenceRepairInvariant = "revisions.sequence_repair_after_validation"
)

var revisionsPositiveInteger = regexp.MustCompile(`^[1-9][0-9]*$`)

type portableChangeSet struct {
	ChangeSetID     uuid.UUID
	IncidentID      uuid.UUID
	PortableActorID uuid.UUID
	RuntimeActorID  uuid.UUID
	Source          string
	Reason          *string
	ClientTxnID     *string
	RequestID       *string
	CreatedAt       time.Time
}

type portableChangeSetMutation struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type portableRecordRevision struct {
	RevisionID  int64
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeJSON  map[string]any
	AfterJSON   map[string]any
	CreatedAt   time.Time
}

type preparedRevisionsImport struct {
	incidentID   uuid.UUID
	runtimeActor uuid.UUID
	actors       sourceport.ActorCatalog
	changeSets   []portableChangeSet
	mutations    []portableChangeSetMutation
	revisions    []portableRecordRevision
}

type revisionsParseFailure struct {
	invariant string
	path      string
	identity  string
}

func exportIncidentBundleFiles(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]incidentportability.File, error) {
	changeSets, err := exportPortableChangeSets(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	mutations, err := exportPortableMutations(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	revisions, err := exportPortableRecordRevisions(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{
		{Path: changeSetsBundlePath, Payload: changeSets},
		{Path: changeSetMutationsPath, Payload: mutations},
		{Path: recordRevisionsBundlePath, Payload: revisions},
	}, nil
}

func exportPortableChangeSets(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]byte, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT change_set_id, incident_id, actor_user_id, source, reason,
       client_txn_id, request_id, created_at
  FROM change_sets
 WHERE incident_id = $1
 ORDER BY change_set_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changeSets := make([]portableChangeSet, 0)
	rowIDs := make([]string, 0)
	for rows.Next() {
		var row portableChangeSet
		var reason, clientTxnID, requestID pgtype.Text
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.IncidentID,
			&row.RuntimeActorID,
			&row.Source,
			&reason,
			&clientTxnID,
			&requestID,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		if row.ChangeSetID == uuid.Nil || row.IncidentID != exportContext.IncidentID ||
			row.RuntimeActorID == uuid.Nil || !validRequiredPortableString(row.Source) {
			return nil, errors.New("revisions export change set is invalid")
		}
		row.Reason = textPointer(reason)
		row.ClientTxnID = textPointer(clientTxnID)
		row.RequestID = textPointer(requestID)
		if !validNullablePortableString(row.Reason) ||
			!validNullablePortableString(row.ClientTxnID) ||
			!validNullablePortableString(row.RequestID) {
			return nil, errors.New("revisions export change set text is invalid")
		}
		changeSets = append(changeSets, row)
		rowIDs = append(rowIDs, row.ChangeSetID.String())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	portableActors := map[string]string{}
	if exportContext.PortableAttributions != nil && len(rowIDs) > 0 {
		portableActors, err = exportContext.PortableAttributions.ResolvePortableSourceActors(
			ctx,
			exportContext.Query,
			exportContext.IncidentID,
			"change_sets",
			"actor_user_id",
			rowIDs,
		)
		if err != nil {
			return nil, err
		}
	}
	var payload bytes.Buffer
	for _, row := range changeSets {
		actorID := row.RuntimeActorID.String()
		if sourceActorID := portableActors[row.ChangeSetID.String()]; sourceActorID != "" {
			parsed, parseErr := uuid.Parse(sourceActorID)
			if parseErr != nil || parsed.String() != sourceActorID {
				return nil, errors.New("revisions export attribution is invalid")
			}
			actorID = sourceActorID
		}
		encoded, err := incidentportability.CanonicalJSONString(map[string]any{
			"change_set_id": row.ChangeSetID.String(),
			"incident_id":   row.IncidentID.String(),
			"actor_user_id": actorID,
			"source":        row.Source,
			"reason":        nullableStringValue(row.Reason),
			"client_txn_id": nullableStringValue(row.ClientTxnID),
			"request_id":    nullableStringValue(row.RequestID),
			"created_at":    row.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, err
		}
		payload.Write(encoded)
	}
	return payload.Bytes(), nil
}

func exportPortableMutations(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]byte, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT mutation.change_set_id, mutation.sequence_no, mutation.target_kind,
       mutation.target_id, mutation.operation_kind, mutation.before_version_id,
       mutation.after_version_id, mutation.before_value, mutation.after_value
  FROM change_set_mutations mutation
  JOIN change_sets change_set
    ON change_set.change_set_id = mutation.change_set_id
 WHERE change_set.incident_id = $1
 ORDER BY mutation.change_set_id, mutation.sequence_no
`, exportContext.IncidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payload bytes.Buffer
	for rows.Next() {
		var row portableChangeSetMutation
		var beforeVersion, afterVersion pgtype.Text
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.SequenceNo,
			&row.TargetKind,
			&row.TargetID,
			&row.OperationKind,
			&beforeVersion,
			&afterVersion,
			&beforeRaw,
			&afterRaw,
		); err != nil {
			return nil, err
		}
		row.BeforeVersionID = textPointer(beforeVersion)
		row.AfterVersionID = textPointer(afterVersion)
		row.BeforeValue, err = decodeDatabaseJSON(beforeRaw)
		if err != nil {
			return nil, err
		}
		row.AfterValue, err = decodeDatabaseJSON(afterRaw)
		if err != nil {
			return nil, err
		}
		encoded, err := incidentportability.CanonicalJSONString(map[string]any{
			"change_set_id":     row.ChangeSetID.String(),
			"sequence_no":       row.SequenceNo,
			"target_kind":       row.TargetKind,
			"target_id":         row.TargetID,
			"operation_kind":    row.OperationKind,
			"before_version_id": nullableStringValue(row.BeforeVersionID),
			"after_version_id":  nullableStringValue(row.AfterVersionID),
			"before_value":      row.BeforeValue,
			"after_value":       row.AfterValue,
		})
		if err != nil {
			return nil, err
		}
		payload.Write(encoded)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload.Bytes(), nil
}

func exportPortableRecordRevisions(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]byte, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT revision.revision_id, revision.change_set_id, revision.record_id,
       revision.row_version, revision.before_json, revision.after_json,
       revision.created_at
  FROM record_revisions revision
  JOIN change_sets change_set
    ON change_set.change_set_id = revision.change_set_id
 WHERE change_set.incident_id = $1
 ORDER BY revision.revision_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payload bytes.Buffer
	for rows.Next() {
		var row portableRecordRevision
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(
			&row.RevisionID,
			&row.ChangeSetID,
			&row.RecordID,
			&row.RowVersion,
			&beforeRaw,
			&afterRaw,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		before, err := decodeDatabaseJSONObject(beforeRaw)
		if err != nil {
			return nil, err
		}
		after, err := decodeDatabaseJSONObject(afterRaw)
		if err != nil {
			return nil, err
		}
		encoded, err := incidentportability.CanonicalJSONString(map[string]any{
			"revision_id":   row.RevisionID,
			"change_set_id": row.ChangeSetID.String(),
			"record_id":     row.RecordID.String(),
			"row_version":   row.RowVersion,
			"before_json":   before,
			"after_json":    after,
			"created_at":    row.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, err
		}
		payload.Write(encoded)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload.Bytes(), nil
}

func prepareRevisionsImport(
	bundle sourceport.Bundle,
	importContext sourceport.ImportContext,
) (preparedRevisionsImport, error) {
	prepared := preparedRevisionsImport{
		incidentID:   importContext.IncidentID,
		runtimeActor: importContext.ActorUserID,
		actors:       importContext.Actors,
	}
	if bundle == nil || importContext.IncidentID == uuid.Nil ||
		importContext.ActorUserID == uuid.Nil {
		return preparedRevisionsImport{}, revisionsFailure(revisionsReferencesInvariant)
	}
	var failures []revisionsParseFailure
	changeSetRows, err := decodeRevisionsRows(bundle, changeSetsBundlePath)
	if err != nil {
		failures = append(failures, parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, ""))
	}
	mutationRows, mutationErr := decodeRevisionsRows(bundle, changeSetMutationsPath)
	if mutationErr != nil {
		failures = append(failures, parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, ""))
	}
	revisionRows, revisionErr := decodeRevisionsRows(bundle, recordRevisionsBundlePath)
	if revisionErr != nil {
		failures = append(failures, parseFailure(revisionsReferencesInvariant, recordRevisionsBundlePath, ""))
	}
	if len(failures) > 0 {
		return preparedRevisionsImport{}, selectedRevisionsFailure(failures)
	}

	for _, raw := range changeSetRows {
		row, failure := parsePortableChangeSet(raw, importContext)
		if failure != nil {
			failures = append(failures, *failure)
			continue
		}
		prepared.changeSets = append(prepared.changeSets, row)
	}
	for _, raw := range mutationRows {
		row, failure := parsePortableMutation(raw)
		if failure != nil {
			failures = append(failures, *failure)
			continue
		}
		prepared.mutations = append(prepared.mutations, row)
	}
	for _, raw := range revisionRows {
		row, failure := parsePortableRecordRevision(raw)
		if failure != nil {
			failures = append(failures, *failure)
			continue
		}
		prepared.revisions = append(prepared.revisions, row)
	}
	if len(failures) > 0 {
		return preparedRevisionsImport{}, selectedRevisionsFailure(failures)
	}

	sort.Slice(prepared.changeSets, func(i, j int) bool {
		return prepared.changeSets[i].ChangeSetID.String() < prepared.changeSets[j].ChangeSetID.String()
	})
	sort.Slice(prepared.mutations, func(i, j int) bool {
		left, right := prepared.mutations[i], prepared.mutations[j]
		if left.ChangeSetID != right.ChangeSetID {
			return left.ChangeSetID.String() < right.ChangeSetID.String()
		}
		return left.SequenceNo < right.SequenceNo
	})
	sort.Slice(prepared.revisions, func(i, j int) bool {
		return prepared.revisions[i].RevisionID < prepared.revisions[j].RevisionID
	})

	changeSets := make(map[uuid.UUID]struct{}, len(prepared.changeSets))
	for _, row := range prepared.changeSets {
		if row.IncidentID != importContext.IncidentID {
			failures = append(failures, parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, row.ChangeSetID.String()))
		}
		if _, duplicate := changeSets[row.ChangeSetID]; duplicate {
			failures = append(failures, parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, row.ChangeSetID.String()))
		}
		changeSets[row.ChangeSetID] = struct{}{}
		if _, admitted := importContext.Actors.Lookup(row.PortableActorID.String()); !admitted {
			failures = append(failures, parseFailure(revisionsActorsInvariant, changeSetsBundlePath, row.ChangeSetID.String()))
		}
	}
	for _, row := range prepared.mutations {
		if _, exists := changeSets[row.ChangeSetID]; !exists {
			failures = append(failures, parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, mutationIdentity(row)))
		}
	}
	for _, row := range prepared.revisions {
		if _, exists := changeSets[row.ChangeSetID]; !exists {
			failures = append(failures, parseFailure(revisionsReferencesInvariant, recordRevisionsBundlePath, strconv.FormatInt(row.RevisionID, 10)))
		}
		for _, snapshot := range []map[string]any{row.BeforeJSON, row.AfterJSON} {
			if !snapshotActorsAdmitted(snapshot, importContext.Actors) {
				failures = append(failures, parseFailure(revisionsActorsInvariant, recordRevisionsBundlePath, strconv.FormatInt(row.RevisionID, 10)))
			}
		}
	}
	if len(failures) > 0 {
		return preparedRevisionsImport{}, selectedRevisionsFailure(failures)
	}
	return prepared, nil
}

func decodeRevisionsRows(bundle sourceport.Bundle, path string) ([]map[string]any, error) {
	payload, ok := bundle.File(path)
	if !ok {
		return nil, errors.New("required revisions bundle member is absent")
	}
	return incidentportability.DecodeStrictNDJSONObjects(payload, path)
}

func parsePortableChangeSet(
	raw map[string]any,
	importContext sourceport.ImportContext,
) (portableChangeSet, *revisionsParseFailure) {
	identity := portableIdentity(raw["change_set_id"])
	required := []string{
		"change_set_id", "incident_id", "actor_user_id", "source", "reason",
		"client_txn_id", "request_id", "created_at",
	}
	if !exactRevisionsMembers(raw, required) {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, identity)
		return portableChangeSet{}, &failure
	}
	changeSetID, ok := canonicalRevisionsUUID(raw["change_set_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, identity)
		return portableChangeSet{}, &failure
	}
	incidentID, ok := canonicalRevisionsUUID(raw["incident_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	actorID, ok := canonicalRevisionsUUID(raw["actor_user_id"])
	if !ok {
		failure := parseFailure(revisionsActorsInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	source, ok := requiredRevisionsString(raw["source"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	reason, ok := nullableRevisionsString(raw["reason"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	clientTxnID, ok := nullableRevisionsString(raw["client_txn_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	requestID, ok := nullableRevisionsString(raw["request_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	createdAt, ok := canonicalRevisionsTime(raw["created_at"])
	if !ok {
		failure := parseFailure(revisionsHistoryInvariant, changeSetsBundlePath, changeSetID.String())
		return portableChangeSet{}, &failure
	}
	return portableChangeSet{
		ChangeSetID: changeSetID, IncidentID: incidentID,
		PortableActorID: actorID, RuntimeActorID: importContext.ActorUserID,
		Source: source, Reason: reason, ClientTxnID: clientTxnID,
		RequestID: requestID, CreatedAt: createdAt,
	}, nil
}

func parsePortableMutation(raw map[string]any) (portableChangeSetMutation, *revisionsParseFailure) {
	required := []string{
		"change_set_id", "sequence_no", "target_kind", "target_id",
		"operation_kind", "before_version_id", "after_version_id",
		"before_value", "after_value",
	}
	identity := portableIdentity(raw["change_set_id"]) + ":" + portableIdentity(raw["sequence_no"])
	if !exactRevisionsMembers(raw, required) {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	changeSetID, ok := canonicalRevisionsUUID(raw["change_set_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	sequenceNo, ok := positiveRevisionsInt(raw["sequence_no"])
	if !ok || sequenceNo > int64(^uint(0)>>1) {
		failure := parseFailure(revisionsSequenceInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	targetKind, ok := requiredRevisionsString(raw["target_kind"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	targetID, ok := requiredRevisionsString(raw["target_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	operationKind, ok := requiredRevisionsString(raw["operation_kind"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	beforeVersionID, ok := nullableNonemptyRevisionsString(raw["before_version_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	afterVersionID, ok := nullableNonemptyRevisionsString(raw["after_version_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, changeSetMutationsPath, identity)
		return portableChangeSetMutation{}, &failure
	}
	return portableChangeSetMutation{
		ChangeSetID: changeSetID, SequenceNo: int(sequenceNo),
		TargetKind: targetKind, TargetID: targetID, OperationKind: operationKind,
		BeforeVersionID: beforeVersionID, AfterVersionID: afterVersionID,
		BeforeValue: raw["before_value"], AfterValue: raw["after_value"],
	}, nil
}

func parsePortableRecordRevision(raw map[string]any) (portableRecordRevision, *revisionsParseFailure) {
	required := []string{
		"revision_id", "change_set_id", "record_id", "row_version",
		"before_json", "after_json", "created_at",
	}
	identity := portableIdentity(raw["revision_id"])
	if !exactRevisionsMembers(raw, required) {
		failure := parseFailure(revisionsReferencesInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	revisionID, ok := positiveRevisionsInt(raw["revision_id"])
	if !ok {
		failure := parseFailure(revisionsRecordVersionInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	changeSetID, ok := canonicalRevisionsUUID(raw["change_set_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	recordID, ok := canonicalRevisionsUUID(raw["record_id"])
	if !ok {
		failure := parseFailure(revisionsReferencesInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	rowVersion, ok := positiveRevisionsInt(raw["row_version"])
	if !ok {
		failure := parseFailure(revisionsRecordVersionInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	beforeJSON, ok := nullableRevisionsObject(raw["before_json"])
	if !ok {
		failure := parseFailure(revisionsHistoryInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	afterJSON, ok := nullableRevisionsObject(raw["after_json"])
	if !ok {
		failure := parseFailure(revisionsHistoryInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	createdAt, ok := canonicalRevisionsTime(raw["created_at"])
	if !ok {
		failure := parseFailure(revisionsHistoryInvariant, recordRevisionsBundlePath, identity)
		return portableRecordRevision{}, &failure
	}
	return portableRecordRevision{
		RevisionID: revisionID, ChangeSetID: changeSetID, RecordID: recordID,
		RowVersion: rowVersion, BeforeJSON: beforeJSON, AfterJSON: afterJSON,
		CreatedAt: createdAt,
	}, nil
}

func applyPreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	importContext sourceport.ImportContext,
	validation *IncidentBundleValidationCatalog,
) error {
	if importContext.Attributions == nil || importContext.ActorUserID == uuid.Nil ||
		importContext.IncidentID != prepared.incidentID || validation == nil {
		return revisionsFailure(revisionsReferencesInvariant)
	}
	if err := validatePreparedRevisionsBeforeWriteTx(ctx, tx, prepared, validation); err != nil {
		return err
	}
	for _, row := range prepared.changeSets {
		tag, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id, incident_id, actor_user_id, source, reason,
    client_txn_id, request_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, row.ChangeSetID, row.IncidentID, importContext.ActorUserID, row.Source,
			row.Reason, row.ClientTxnID, row.RequestID, row.CreatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if err := importContext.Attributions.RecordImportedAttribution(
			"change_sets",
			row.ChangeSetID.String(),
			"actor_user_id",
			row.PortableActorID.String(),
		); err != nil {
			return revisionsFailure(revisionsActorsInvariant)
		}
	}
	for _, row := range prepared.mutations {
		history, err := validation.targets.DescribeValues(row.TargetKind, row.TargetID, row.BeforeValue, row.AfterValue)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_version_id, after_version_id, before_value, after_value,
    history_record_ids, history_entry_record_ids
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, row.ChangeSetID, row.SequenceNo, row.TargetKind, row.TargetID,
			row.OperationKind, row.BeforeVersionID, row.AfterVersionID,
			jsonOrNil(row.BeforeValue), jsonOrNil(row.AfterValue),
			history.HistoryRecordIDs, history.HistoryEntryRecordIDs)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsSequenceInvariant)
		}
	}
	for _, row := range prepared.revisions {
		tag, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    revision_id, change_set_id, record_id, row_version,
    before_json, after_json, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
`, row.RevisionID, row.ChangeSetID, row.RecordID, row.RowVersion,
			jsonOrNil(row.BeforeJSON), jsonOrNil(row.AfterJSON), row.CreatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsRecordVersionInvariant)
		}
	}
	return nil
}

func validatePreparedRevisionsBeforeWriteTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	validation *IncidentBundleValidationCatalog,
) error {
	for _, mutation := range prepared.mutations {
		if !validation.resolvesTargetKind(mutation.TargetKind) ||
			!canonicalPortableTargetID(mutation.TargetKind, mutation.TargetID) {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if _, err := validation.targets.DescribeValues(
			mutation.TargetKind,
			mutation.TargetID,
			mutation.BeforeValue,
			mutation.AfterValue,
		); err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		dispatch, err := validation.targets.dispatchClass(mutation.TargetKind)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if dispatch == RollbackDispatchRow {
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
	validation *IncidentBundleValidationCatalog,
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
	switch invariantID {
	case revisionsReferencesInvariant,
		revisionsActorsInvariant,
		revisionsSequenceInvariant,
		revisionsRecordVersionInvariant,
		revisionsHistoryInvariant,
		revisionsSequenceRepairInvariant:
		return &sourceport.Failure{FamilyID: "revisions", InvariantID: invariantID}
	default:
		return errors.New("revisions portability invariant is undeclared")
	}
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

func parseFailure(invariantID, path, identity string) revisionsParseFailure {
	return revisionsParseFailure{invariant: invariantID, path: path, identity: identity}
}

func exactRevisionsMembers(row map[string]any, required []string) bool {
	if len(row) != len(required) {
		return false
	}
	for _, member := range required {
		if _, ok := row[member]; !ok {
			return false
		}
	}
	return true
}

func canonicalRevisionsUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func positiveRevisionsInt(value any) (int64, bool) {
	number, ok := value.(json.Number)
	if !ok || !revisionsPositiveInteger.MatchString(number.String()) {
		return 0, false
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	return parsed, err == nil && parsed >= 1
}

func canonicalRevisionsTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || parsed.Nanosecond()%1_000 != 0 ||
		parsed.UTC().Format(time.RFC3339Nano) != text {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func requiredRevisionsString(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok && validRequiredPortableString(text)
}

func nullableRevisionsString(value any) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := value.(string)
	if !ok || strings.ContainsRune(text, '\x00') {
		return nil, false
	}
	return &text, true
}

func nullableNonemptyRevisionsString(value any) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := requiredRevisionsString(value)
	if !ok {
		return nil, false
	}
	return &text, true
}

func nullableRevisionsObject(value any) (map[string]any, bool) {
	if value == nil {
		return nil, true
	}
	object, ok := value.(map[string]any)
	return object, ok
}

func validRequiredPortableString(value string) bool {
	return value != "" && !strings.ContainsRune(value, '\x00')
}

func validNullablePortableString(value *string) bool {
	return value == nil || !strings.ContainsRune(*value, '\x00')
}

func canonicalPortableTargetID(targetKind, targetID string) bool {
	parsed, err := uuid.Parse(targetID)
	if err == nil && parsed.String() == targetID {
		return true
	}
	prefix := targetKind + ":"
	if !strings.HasPrefix(targetID, prefix) {
		return false
	}
	parsed, err = uuid.Parse(strings.TrimPrefix(targetID, prefix))
	return err == nil && prefix+parsed.String() == targetID
}

func portableIdentity(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func mutationIdentity(row portableChangeSetMutation) string {
	return row.ChangeSetID.String() + ":" + strconv.Itoa(row.SequenceNo)
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func decodeDatabaseJSON(raw []byte) (any, error) {
	if raw == nil {
		return nil, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func decodeDatabaseJSONObject(raw []byte) (map[string]any, error) {
	value, err := decodeDatabaseJSON(raw)
	if err != nil || value == nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("revisions history snapshot is not an object")
	}
	return object, nil
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
