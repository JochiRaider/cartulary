package revisions

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func prepareRevisionsImport(ctx context.Context, bundle incidentportability.MemberReader, importContext sourceport.ImportContext) (preparedRevisionsImport, error) {
	if bundle == nil || importContext.IncidentID == uuid.Nil || importContext.ActorUserID == uuid.Nil {
		return preparedRevisionsImport{}, revisionsFailure(revisionsReferencesInvariant)
	}
	members, err := incidentportability.BindMembers(bundle, []string{changeSetsBundlePath, changeSetMutationsPath, recordRevisionsBundlePath})
	if err != nil {
		return preparedRevisionsImport{}, revisionsFailure(revisionsReferencesInvariant)
	}
	prepared := preparedRevisionsImport{incidentID: importContext.IncidentID, runtimeActor: importContext.ActorUserID, actors: importContext.Actors, members: members}
	previous := ""
	if err := prepared.walkChangeSets(ctx, func(row portableChangeSet) error {
		id := row.ChangeSetID.String()
		if id <= previous || row.IncidentID != prepared.incidentID {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if _, admitted := prepared.actors.Lookup(row.PortableActorID.String()); !admitted {
			return revisionsFailure(revisionsActorsInvariant)
		}
		previous = id
		return nil
	}); err != nil {
		return preparedRevisionsImport{}, err
	}
	previous = ""
	sequence := 0
	if err := prepared.walkMutations(ctx, func(row portableChangeSetMutation) error {
		id := row.ChangeSetID.String()
		if id != previous {
			if id < previous {
				prepared.sequenceInvalid = true
			}
			previous = id
			sequence = 0
		}
		if row.SequenceNo != sequence+1 {
			prepared.sequenceInvalid = true
		}
		sequence = row.SequenceNo
		for _, snapshot := range []any{row.BeforeValue, row.AfterValue} {
			if !walkSnapshotActors(snapshot, func(actorID string) bool { _, admitted := prepared.actors.Lookup(actorID); return admitted }) {
				return revisionsFailure(revisionsActorsInvariant)
			}
		}
		return nil
	}); err != nil {
		return preparedRevisionsImport{}, err
	}
	var revisionID int64
	if err := prepared.walkRevisions(ctx, func(row portableRecordRevision) error {
		if row.RevisionID <= revisionID {
			prepared.revisionIdentityInvalid = true
		}
		revisionID = row.RevisionID
		for _, snapshot := range []map[string]any{row.BeforeJSON, row.AfterJSON} {
			if !snapshotActorsAdmitted(snapshot, prepared.actors) {
				return revisionsFailure(revisionsActorsInvariant)
			}
		}
		return nil
	}); err != nil {
		return preparedRevisionsImport{}, err
	}
	return prepared, nil
}
func (p preparedRevisionsImport) walkChangeSets(ctx context.Context, visit func(portableChangeSet) error) error {
	return incidentportability.WalkPath(ctx, p.members, changeSetsBundlePath, func(raw map[string]any) error {
		row, failure := parsePortableChangeSet(raw, sourceport.ImportContext{ActorUserID: p.runtimeActor, IncidentID: p.incidentID, Actors: p.actors})
		if failure != nil {
			return revisionsFailure(failure.invariant)
		}
		return visit(row)
	})
}
func (p preparedRevisionsImport) walkMutations(ctx context.Context, visit func(portableChangeSetMutation) error) error {
	return incidentportability.WalkPath(ctx, p.members, changeSetMutationsPath, func(raw map[string]any) error {
		row, failure := parsePortableMutation(raw)
		if failure != nil {
			return revisionsFailure(failure.invariant)
		}
		return visit(row)
	})
}
func (p preparedRevisionsImport) walkRevisions(ctx context.Context, visit func(portableRecordRevision) error) error {
	return incidentportability.WalkPath(ctx, p.members, recordRevisionsBundlePath, func(raw map[string]any) error {
		row, failure := parsePortableRecordRevision(raw)
		if failure != nil {
			return revisionsFailure(failure.invariant)
		}
		return visit(row)
	})
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
