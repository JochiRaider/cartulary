package ownerfacade

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type FinalizeCommand struct {
	Request         ImportOwnerCreateRequest
	ChangeSetID     uuid.UUID
	SequenceNo      int
	RecordID        uuid.UUID
	Operation       string
	CreatedOrReused string
	OwnerResultCode string
	BeforeVersionID *string
	BeforeValue     map[string]any
	BeforeSnapshot  *revisions.RecordSnapshot
	Row             map[string]any
	CreatedAt       time.Time
}

type HistoricalRecordRevisionAppender interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendHistoricalRevisionTx(context.Context, pgx.Tx, revisions.HistoricalRevisionInput) error
}

type LiveRecordRevisionAppender interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
}

type recordFinalizationAppender interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
}

func FinalizeHistoricalRecordTx(ctx context.Context, tx pgx.Tx, revisionAppender HistoricalRecordRevisionAppender, command FinalizeCommand) (ImportOwnerCreateResponse, error) {
	if revisionAppender == nil {
		return ImportOwnerCreateResponse{}, fmt.Errorf("finalize import owner row: revision appender is required")
	}
	return finalizeHistoricalRecordTx(ctx, tx, revisionAppender, command)
}

func FinalizeLiveRecordTx(ctx context.Context, tx pgx.Tx, revisionAppender LiveRecordRevisionAppender, publications collaboration.RecordChangedAppender, command FinalizeCommand) (ImportOwnerCreateResponse, error) {
	if revisionAppender == nil {
		return ImportOwnerCreateResponse{}, fmt.Errorf("finalize import owner row: revision and intent appender is required")
	}
	if publications == nil {
		return ImportOwnerCreateResponse{}, fmt.Errorf("finalize import owner row: Collaboration publication appender is required")
	}
	return finalizeLiveRecordTx(ctx, tx, revisionAppender, publications, command)
}

func prepareRecordFinalizationTx(ctx context.Context, tx pgx.Tx, revisionAppender recordFinalizationAppender, command FinalizeCommand) (ImportOwnerCreateResponse, revisions.RecordSnapshot, int64, bool, error) {
	rowVersion, err := RowVersionFromRow(command.Row)
	if err != nil {
		return ImportOwnerCreateResponse{}, revisions.RecordSnapshot{}, 0, false, err
	}
	operation := command.Operation
	if operation == "" {
		operation = "create"
	}
	createdOrReused := command.CreatedOrReused
	if createdOrReused == "" {
		createdOrReused = "created"
	}
	resultCode := command.OwnerResultCode
	if resultCode == "" {
		resultCode = createdOrReused
	}
	afterSnapshot, err := revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return ImportOwnerCreateResponse{}, revisions.RecordSnapshot{}, 0, false, err
	}
	afterVersionID := VersionID(command.RecordID, rowVersion)
	if err := revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "record",
		RecordID:        command.RecordID,
		OperationKind:   operation,
		BeforeVersionID: command.BeforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  command.BeforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return ImportOwnerCreateResponse{}, revisions.RecordSnapshot{}, 0, false, err
	}
	result := ImportOwnerCreateResponse{
		RecordID:             command.RecordID,
		RowVersion:           rowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           command.Row,
	}
	return result, afterSnapshot, rowVersion, operation == "create" || (command.BeforeSnapshot != nil && operation != "reuse"), nil
}

func finalizeHistoricalRecordTx(ctx context.Context, tx pgx.Tx, revisionAppender HistoricalRecordRevisionAppender, command FinalizeCommand) (ImportOwnerCreateResponse, error) {
	result, afterSnapshot, rowVersion, appendRevision, err := prepareRecordFinalizationTx(ctx, tx, revisionAppender, command)
	if err != nil || !appendRevision {
		return result, err
	}
	err = revisionAppender.AppendHistoricalRevisionTx(ctx, tx, revisions.HistoricalRevisionInput{
		ChangeSetID: command.ChangeSetID, RecordID: command.RecordID, RowVersion: rowVersion,
		BeforeSnapshot: command.BeforeSnapshot, AfterSnapshot: &afterSnapshot,
	})
	return result, err
}

func finalizeLiveRecordTx(ctx context.Context, tx pgx.Tx, revisionAppender LiveRecordRevisionAppender, publications collaboration.RecordChangedAppender, command FinalizeCommand) (ImportOwnerCreateResponse, error) {
	result, afterSnapshot, rowVersion, appendRevision, err := prepareRecordFinalizationTx(ctx, tx, revisionAppender, command)
	if err != nil || !appendRevision {
		return result, err
	}
	fieldKeys, facts := importRecordFacts(command.BeforeValue, command.Row)
	if err := revisionAppender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID: command.ChangeSetID, RecordID: command.RecordID, RowVersion: rowVersion,
		BeforeSnapshot: command.BeforeSnapshot, AfterSnapshot: &afterSnapshot, ConflictFacts: facts,
	}); err != nil {
		return ImportOwnerCreateResponse{}, err
	}
	patch := collabprotocol.BuildViewRowPatch(command.Row, fieldKeys)
	changeKind := "invalidate"
	if patch != nil {
		changeKind = "patch"
	}
	if err := publications.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID: command.Request.IncidentID, RecordID: command.RecordID, ChangeSetID: command.ChangeSetID,
		ActorUserID: command.Request.ActorUserID, RowVersion: rowVersion, ClientTxnID: command.Request.ClientTxnID,
		MutationOrdinal: max(command.SequenceNo-1, 0), CreatedAt: command.CreatedAt.UTC(), PublicFieldKeys: fieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{ViewSchemaID: command.Request.TargetViewSchemaID, RecordID: command.RecordID, RowVersion: rowVersion, ChangeKind: changeKind, PatchCells: patch}},
	}); err != nil {
		return ImportOwnerCreateResponse{}, err
	}
	return result, nil
}

func importRecordFacts(beforeRow map[string]any, afterRow map[string]any) ([]string, []revisions.RevisionConflictFact) {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	keys := make([]string, 0, len(beforeCells)+len(afterCells))
	seen := map[string]struct{}{}
	for key := range beforeCells {
		seen[key] = struct{}{}
	}
	for key := range afterCells {
		seen[key] = struct{}{}
	}
	for key := range seen {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	facts := make([]revisions.RevisionConflictFact, 0, len(keys))
	for _, key := range keys {
		beforeValue, beforePresent := beforeCells[key]
		afterValue, afterPresent := afterCells[key]
		facts = append(facts, revisions.RevisionConflictFact{FieldKey: key, BeforePresent: beforePresent, BeforeValue: beforeValue, AfterPresent: afterPresent, AfterValue: afterValue})
	}
	return keys, facts
}

func RowVersionFromRow(row map[string]any) (int64, error) {
	switch value := row["row_version"].(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("import row has unexpected row_version type %T", value)
	}
}

func VersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}
