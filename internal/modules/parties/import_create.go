package parties

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

type ImportRecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
}

type ImportDependencies struct {
	RecordEnvelopes ImportRecordEnvelopeCapability
	Projections     partyprojection.Rows
	Revisions       ownerfacade.RecordRevisionAndIntentAppender
}

func (d ImportDependencies) validate() error {
	if d.RecordEnvelopes == nil {
		return fmt.Errorf("parties import dependencies: Records insert is required")
	}
	if d.Projections == nil {
		return fmt.Errorf("parties import dependencies: Projection refresh/load is required")
	}
	if d.Revisions == nil {
		return fmt.Errorf("parties import dependencies: Revision finalization is required")
	}
	return nil
}

type importOwner struct {
	dependencies ImportDependencies
}

func NewImportContribution(
	targetViewSchemaID string,
	facadeID string,
	dependencies ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("party import surface %q not mapped", targetViewSchemaID)
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	owner := &importOwner{dependencies: dependencies}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		owner.CreateImportRowTx,
	)
}

func (o *importOwner) CreateImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ImportCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != ViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("party import surface %q not mapped", request.TargetViewSchemaID)
	}
	values, err := valuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	params := partysource.CreateParams{Values: values}
	if err := validateCreateParams(params); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	now := command.Now.UTC()
	if recordID, found, err := partysource.FindReusablePartyTx(ctx, tx, request.IncidentID, params); err != nil || found {
		if err != nil {
			var matchConflict *PartyMatchConflictError
			if errors.As(err, &matchConflict) {
				return ownerfacade.ImportOwnerCreateResponse{}, ownerfacade.NewPartyMatchConflictError(
					matchConflict.ReasonCode,
					matchConflict.ConflictingFieldKeys,
					err,
				)
			}
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		row, err := o.refreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return ownerfacade.FinalizeRecordRevisionAndIntentTx(ctx, tx, o.dependencies.Revisions, ownerfacade.FinalizeCommand{
			Request: request, ChangeSetID: command.ChangeSetID, SequenceNo: command.SequenceNo,
			RecordID: recordID, Operation: "reuse", CreatedOrReused: "reused",
			OwnerResultCode: "reused", Row: row,
		})
	}
	recordID, err := o.dependencies.RecordEnvelopes.InsertTx(ctx, tx, records.InsertParams{
		IncidentID: request.IncidentID, RecordType: "party",
		CreatedByUserID: request.ActorUserID, CreatedAt: now,
		UpdatedByUserID: request.ActorUserID, UpdatedAt: now, RowVersion: 1,
	})
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := partysource.InsertPartyTx(ctx, tx, recordID, request.IncidentID, params, now); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := o.refreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeRecordRevisionAndIntentTx(ctx, tx, o.dependencies.Revisions, ownerfacade.FinalizeCommand{
		Request: request, ChangeSetID: command.ChangeSetID, SequenceNo: command.SequenceNo,
		RecordID: recordID, Operation: "create", CreatedOrReused: "created",
		OwnerResultCode: "created", Row: row,
	})
}

func (o *importOwner) refreshImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	viewSchemaID string,
	recordID uuid.UUID,
) (map[string]any, error) {
	if viewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("party import projection surface %q not mapped", viewSchemaID)
	}
	if err := o.dependencies.Projections.RefreshPartyTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return o.dependencies.Projections.LoadPartyTx(ctx, tx, recordID)
}

func valuesFromImport(values map[string]ownerfacade.ImportScalarValue) (map[string]policy.Value, error) {
	result := make(map[string]policy.Value, len(values))
	for field, value := range values {
		normalized, admissionErr := policy.Admit(field, value.Text)
		if admissionErr != nil {
			registryField, _ := policy.LookupField(field)
			return nil, ownerfacade.NewImportOwnerCreateValidationError(
				"invalid_text", field, registryField.StringContractID,
				fmt.Errorf("party field admission failed: %w", admissionErr),
			)
		}
		result[field] = normalized
	}
	return result, nil
}
