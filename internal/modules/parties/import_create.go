package parties

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type ImportRecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
}

type ImportDependencies struct {
	RecordEnvelopes ImportRecordEnvelopeCapability
	Projections     partyprojection.Rows
	Revisions       ownerfacade.LiveRecordRevisionAppender
	Collaboration   collaboration.RecordChangedAppender
}

func (d ImportDependencies) validate() error {
	if nilDependency(d.RecordEnvelopes) {
		return fmt.Errorf("parties import dependencies: Records insert is required")
	}
	if nilDependency(d.Projections) {
		return fmt.Errorf("parties import dependencies: Projection refresh/load is required")
	}
	if nilDependency(d.Revisions) {
		return fmt.Errorf("parties import dependencies: Revision finalization is required")
	}
	if nilDependency(d.Collaboration) {
		return fmt.Errorf("parties import dependencies: Collaboration publication is required")
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
	command ownerfacade.ImportOwnerCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != ViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("party import surface %q not mapped", request.TargetViewSchemaID)
	}
	indexed, err := ownerfacade.IndexImportFieldValues(request.FieldValues)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	values, err := valuesFromImport(indexed)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	now := command.Now.UTC()
	created, err := createOrReusePartyTx(
		ctx,
		tx,
		o.dependencies.RecordEnvelopes,
		request.IncidentID,
		request.ActorUserID,
		values,
		now,
	)
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
	row, err := o.refreshImportRowTx(ctx, tx, request.TargetViewSchemaID, created.recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	operation := "reuse"
	createdOrReused := "reused"
	if created.created {
		operation = "create"
		createdOrReused = "created"
	}
	return ownerfacade.FinalizeLiveRecordTx(ctx, tx, o.dependencies.Revisions, o.dependencies.Collaboration, ownerfacade.FinalizeCommand{
		Request: request, ChangeSetID: command.ChangeSetID, SequenceNo: command.SequenceNo,
		RecordID: created.recordID, Operation: operation, CreatedOrReused: createdOrReused,
		OwnerResultCode: createdOrReused, Row: row, CreatedAt: command.Now,
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
	inputs := make(map[string]createValueInput, len(values))
	for fieldKey, value := range values {
		field, known := policy.LookupField(fieldKey)
		if !known || !value.IsValid() ||
			(value.Kind() != ownerfacade.ImportScalarText && value.Kind() != ownerfacade.ImportScalarNull) {
			guard := "party_field_registry"
			if known {
				guard = field.StringContractID
			}
			return nil, ownerfacade.NewImportOwnerCreateValidationError(
				"invalid_text", fieldKey, guard,
				fmt.Errorf("party import scalar kind %q is not admitted", value.Kind()),
			)
		}
		var text *string
		if scalar, ok := value.Text(); ok {
			text = &scalar
		}
		inputs[fieldKey] = createValueInput{present: true, text: text}
	}
	result, admissionErr := admitCreateValues(inputs)
	if admissionErr == nil {
		return result, nil
	}
	registryField, _ := policy.LookupField(admissionErr.field)
	return nil, ownerfacade.NewImportOwnerCreateValidationError(
		"invalid_text", admissionErr.field, registryField.StringContractID,
		fmt.Errorf("party field admission failed: %s", admissionErr.reasonCode),
	)
}
