package evidence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

type importCreateCommand = ownerfacade.ImportOwnerCreateCommand

func newImportCreateFacade(
	facadeID string,
	service *sourceMutationService,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if service == nil {
		return nil, fmt.Errorf("compose Evidence import contribution: source mutations are required")
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: ViewSchemaID,
			FacadeID:           facadeID,
		},
		service.CreateImportRowTx,
	)
}

func (s *sourceMutationService) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command importCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != ViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("evidence import surface %q not mapped", request.TargetViewSchemaID)
	}
	values, err := ownerfacade.IndexImportFieldValues(request.FieldValues)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	params := createParams{Values: evidenceValuesFromImport(values)}
	if err := validateCreateParams(params); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	mutationOrder, err := command.AllocateMutationSequence(1)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	result, err := s.mutations.createTx(ctx, tx, evidenceCreateTxCommand{
		IncidentID:    request.IncidentID,
		ActorUserID:   request.ActorUserID,
		ViewSchemaID:  request.TargetViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		ChangeSetID:   &command.ChangeSetID,
		MutationOrder: mutationOrder,
		Values:        params.Values,
		Now:           command.Now,
	}, params)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:             result.recordID,
		RowVersion:           1,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", result.changeSetID, mutationOrder),
		CreatedOrReused:      "created",
		OwnerResultCode:      "created",
		RowRefresh:           result.row,
	}, nil
}

func evidenceValuesFromImport(values map[string]ownerfacade.ImportScalarValue) map[string]FieldValue {
	result := make(map[string]FieldValue, len(values))
	for field, value := range values {
		converted := FieldValue{}
		if scalar, ok := value.Text(); ok {
			converted.Text = &scalar
		}
		if scalar, ok := value.Timestamp(); ok {
			converted.Timestamp = &scalar
		}
		if scalar, ok := value.UUID(); ok {
			converted.UUID = &scalar
		}
		if scalar, ok := value.Number(); ok {
			converted.Number = &scalar
		}
		if scalar, ok := value.Bool(); ok {
			converted.Bool = &scalar
		}
		result[field] = converted
	}
	return result
}
