package revisions

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

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
