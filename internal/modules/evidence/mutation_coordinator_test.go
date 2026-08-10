package evidence

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestEvidenceMutationCoordinatorEffectOrderAndFaultBoundary(t *testing.T) {
	expectedFailure := errors.New("injected source-row failure")
	events := make([]string, 0)
	recordID := uuid.MustParse("d8ba17f5-bf78-49c3-b113-5e8e645c3c02")
	changeSetID := uuid.MustParse("14ec06dd-f7ae-4915-b447-38ec75df8fa8")
	rows := &coordinatorSourceRows{events: &events, failInsert: expectedFailure}
	coordinator := evidenceMutationCoordinator{
		incidents: coordinatorIncidentAdmission{events: &events},
		source: evidenceSourceKernel{
			records:     coordinatorRecords{events: &events, recordID: recordID},
			rows:        rows,
			projections: coordinatorProjectionRows{events: &events, recordID: recordID},
		},
		revisions:     coordinatorRevisions{events: &events, changeSetID: changeSetID},
		collaboration: coordinatorCollaboration{events: &events},
	}
	command := coordinatorCreateCommand()

	if _, err := coordinator.createTx(context.Background(), nil, command, WorkbookCreateParams{Values: command.Request.Values}); !errors.Is(err, expectedFailure) {
		t.Fatalf("createTx() error = %v, want injected source-row failure", err)
	}
	if want := []string{"incident", "record", "source-row"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("fault events = %#v, want %#v", events, want)
	}

	events = events[:0]
	rows.failInsert = nil
	result, err := coordinator.createTx(context.Background(), nil, command, WorkbookCreateParams{Values: command.Request.Values})
	if err != nil {
		t.Fatalf("createTx() success error = %v", err)
	}
	wantEvents := []string{
		"incident",
		"record",
		"source-row",
		"projection-refresh",
		"projection-load",
		"snapshot-capture",
		"change-set",
		"mutation",
		"record-revision",
		"collaboration-intent",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("success events = %#v, want %#v", events, wantEvents)
	}
	if result.recordID != recordID || result.changeSetID != changeSetID {
		t.Fatalf("result identities = (%s, %s), want (%s, %s)", result.recordID, result.changeSetID, recordID, changeSetID)
	}
}

func coordinatorCreateCommand() WorkbookCreateCommand {
	title := "Disk image"
	return WorkbookCreateCommand{
		Actor:      authn.UserRecord{ID: uuid.MustParse("8d8a27c9-d070-4bb0-baf8-b7f8f48d47c1")},
		IncidentID: uuid.MustParse("c3b6b590-bf8f-489e-bfdd-5244074cd45e"),
		Request: WorkbookCreateRequest{
			ViewSchemaID: ViewSchemaID,
			ClientTxnID:  "txn-evidence-coordinator",
			Values: map[string]WorkbookFieldValue{
				"evidence.title": {Text: &title},
			},
		},
		RequestID: "request-evidence-coordinator",
		RouteKey:  "workbook.rows.create",
		Now:       time.Date(2026, time.July, 30, 18, 0, 0, 0, time.UTC),
	}
}

type coordinatorIncidentAdmission struct {
	events *[]string
}

func (admission coordinatorIncidentAdmission) EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error {
	*admission.events = append(*admission.events, "incident")
	return nil
}

type coordinatorRecords struct {
	events   *[]string
	recordID uuid.UUID
}

func (port coordinatorRecords) InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error) {
	*port.events = append(*port.events, "record")
	return port.recordID, nil
}

func (coordinatorRecords) AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	return 0, errors.New("unexpected AdvanceVersionTx")
}

type coordinatorSourceRows struct {
	events     *[]string
	failInsert error
}

func (port *coordinatorSourceRows) InsertWorkbookRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, WorkbookCreateParams, time.Time) error {
	*port.events = append(*port.events, "source-row")
	return port.failInsert
}

func (*coordinatorSourceRows) ValidateWorkbookLifecyclePatchTx(context.Context, pgx.Tx, uuid.UUID, []WorkbookLifecyclePatchChange) error {
	return errors.New("unexpected ValidateWorkbookLifecyclePatchTx")
}

func (*coordinatorSourceRows) ApplyWorkbookDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, WorkbookFieldValue, time.Time) (bool, error) {
	return false, errors.New("unexpected ApplyWorkbookDirectChangeTx")
}

func (*coordinatorSourceRows) TouchWorkbookRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error {
	return errors.New("unexpected TouchWorkbookRowTx")
}

type coordinatorProjectionRows struct {
	events   *[]string
	recordID uuid.UUID
}

func (port coordinatorProjectionRows) RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error {
	*port.events = append(*port.events, "projection-refresh")
	return nil
}

func (port coordinatorProjectionRows) LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	*port.events = append(*port.events, "projection-load")
	return map[string]any{
		"record_id": port.recordID.String(),
		"cells": map[string]any{
			"evidence.title": map[string]any{"value": "Disk image"},
		},
	}, nil
}

func (coordinatorProjectionRows) RefreshEvidenceSupportTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errors.New("unexpected RefreshEvidenceSupportTx")
}

func (coordinatorProjectionRows) RebuildEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errors.New("unexpected RebuildEvidenceTx")
}

type coordinatorRevisions struct {
	events      *[]string
	changeSetID uuid.UUID
}

func (port coordinatorRevisions) CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.CapturedRecordSnapshot, error) {
	*port.events = append(*port.events, "snapshot-capture")
	return revisions.CapturedRecordSnapshot{}, nil
}

func (port coordinatorRevisions) AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error) {
	*port.events = append(*port.events, "change-set")
	return port.changeSetID, nil
}

func (port coordinatorRevisions) AppendCapturedRecordMutationTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordMutationParams) error {
	*port.events = append(*port.events, "mutation")
	return nil
}

func (port coordinatorRevisions) AppendCapturedRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendCapturedRecordRevisionParams) error {
	*port.events = append(*port.events, "record-revision")
	return nil
}

type coordinatorCollaboration struct {
	events *[]string
}

func (port coordinatorCollaboration) AppendIntentTx(context.Context, pgx.Tx, collaboration.EventIntent) error {
	*port.events = append(*port.events, "collaboration-intent")
	return nil
}
