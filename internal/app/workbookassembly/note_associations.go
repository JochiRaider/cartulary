package workbookassembly

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteAssociationProjectionRows interface {
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
}

type noteAssociationRows struct{ projections NoteAssociationProjectionRows }

func (a noteAssociationRows) LoadEvidenceTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (map[string]any, error) {
	return a.projections.LoadRowTx(ctx, tx, "cartulary.view.evidence.v1", id)
}
func (a noteAssociationRows) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return a.projections.RefreshRowTx(ctx, tx, "cartulary.view.evidence.v1", id)
}
func (a noteAssociationRows) LabelTx(ctx context.Context, tx pgx.Tx, record records.Envelope) (string, error) {
	switch record.RecordType {
	case "host", "identity":
		return hostidentity.ReadRecordLabelTx(ctx, tx, record.RecordType, record.RecordID)
	}
	schema, field, fallback := "cartulary.view.notes.v1", "note.title", "Note"
	if record.RecordType == "timeline_event" {
		schema, field, fallback = "cartulary.view.timeline.v2", "timeline.activity_synopsis_text", "Timeline record"
	}
	if record.RecordType == "evidence" {
		schema, field, fallback = "cartulary.view.evidence.v1", "evidence.title", "Evidence"
	}
	row, err := a.projections.LoadRowTx(ctx, tx, schema, record.RecordID)
	if err != nil {
		return "", err
	}
	cells, _ := row["cells"].(map[string]any)
	cell, _ := cells[field].(map[string]any)
	label, _ := cell["value"].(string)
	if strings.TrimSpace(label) == "" {
		label = fallback + " " + record.RecordID.String()
	}
	return label, nil
}

type noteAssociationProvider struct{ owner *artifacts.NoteAssociations }

func NewNoteAssociationProvider(pool postgres.DB, mutations *artifacts.MutationFacade, rows NoteAssociationProjectionRows) (workbook.NoteAssociationProvider, error) {
	if isNilDependency(pool) || isNilDependency(rows) {
		return nil, fmt.Errorf("Note association projection dependencies are required")
	}
	owner, err := artifacts.NewNoteAssociations(mutations, admission.NewChecker(pool), links.NewStore(), noteAssociationRows{projections: rows})
	if err != nil {
		return nil, err
	}
	return &noteAssociationProvider{owner: owner}, nil
}

func (p *noteAssociationProvider) List(ctx context.Context, q workbook.NoteAssociationQuery) (workbook.NoteAssociationPage, *workbook.MutationFailure, error) {
	kind, valid := artifacts.ParseNoteAssociationKind(q.Kind)
	if !valid {
		return workbook.NoteAssociationPage{}, workbook.InvalidPayloadFailure("kind", "invalid_value"), nil
	}
	result, err := p.owner.List(ctx, artifacts.NoteAssociationRead{ActorUserID: q.ActorID, IncidentID: q.Target.IncidentID, NoteID: q.Target.RecordID, Kind: kind, Limit: q.Limit, BeforeTime: q.BeforeTime, BeforeID: q.BeforeID})
	if failure, safe := noteAssociationFailure(err, ""); safe {
		return workbook.NoteAssociationPage{}, failure, nil
	}
	if err != nil {
		return workbook.NoteAssociationPage{}, nil, err
	}
	items := make([]workbook.NoteAssociationItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, workbook.NoteAssociationItem{ItemRef: item.ItemRef, CounterpartID: item.CounterpartID, ViewSchemaID: item.ViewSchemaID, DisplayLabel: item.DisplayLabel, Direction: item.Direction})
	}
	return workbook.NoteAssociationPage{NoteID: result.NoteID, RowVersion: result.RowVersion, Items: items, HasMore: result.HasMore, LastTime: result.LastTime, LastID: result.LastID}, nil, nil
}

func (p *noteAssociationProvider) Apply(ctx context.Context, c workbook.NoteAssociationCommand, reader io.Reader) (workbook.MutationOutcome, error) {
	request, err := artifacts.AdmitNoteAssociations(reader)
	if err != nil {
		return workbook.RejectedMutation(artifactAdmissionFailure(err)), nil
	}
	result, applyErr := p.owner.Apply(ctx, artifacts.NoteAssociationCommand{ActorUserID: c.Actor.ID, IncidentID: c.Target.IncidentID, NoteID: c.Target.RecordID, Admission: request, RequestID: c.RequestID, Now: c.Now})
	if failure, safe := noteAssociationFailure(applyErr, request.ClientTxnID()); safe {
		return workbook.RejectedMutation(failure), nil
	}
	if applyErr != nil {
		return workbook.MutationOutcome{}, applyErr
	}
	return workbook.SuccessfulRowMutation(artifactMutationResult(result)), nil
}

func noteAssociationFailure(err error, txn string) (*workbook.MutationFailure, bool) {
	if admission.IsDenied(err, admission.DenialInsufficientRole) {
		return workbook.AuthorizationDeniedFailure(), true
	}
	if admission.IsDenied(err, admission.DenialNotVisible) {
		return workbook.TargetNotFoundFailure(), true
	}
	return artifactMutationFailure(err, txn)
}
