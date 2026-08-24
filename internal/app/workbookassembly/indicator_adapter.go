package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatoradmission "github.com/JochiRaider/cartulary/internal/modules/indicators/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type indicatorCreateAdmission struct {
	command indicators.CreateCommand
}

func (value indicatorCreateAdmission) ClientTransactionID() string {
	return value.command.ClientTxnID
}

func newIndicatorCreateProvider(owner *indicators.Application) (workbook.CreateProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Indicator Workbook adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			command, validation := indicatoradmission.DecodeCreateRequest(reader)
			if validation != nil {
				return nil, indicatorValidationFailure(validation), nil
			}
			return indicatorCreateAdmission{command: command}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(indicatorCreateAdmission)
			if !ok || command.ViewSchemaID != indicators.ViewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.CreateIndicatorRow(
				ctx,
				command.Actor.ID,
				command.IncidentID,
				admitted.command,
				command.RequestID,
			)
			if failure, safe := indicatorCreateFailure(err, admitted.command.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(indicatorMutationResult(
				result,
				command.IncidentID,
				admitted.command.ClientTxnID,
			)), nil
		},
	)
}

func indicatorCreateFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return workbook.ClientTxnConflictFailure(clientTxnID), true
	}
	if admission.IsDenied(err, admission.DenialIncidentClosed) {
		return workbook.IncidentClosedFailure(), true
	}
	if errors.Is(err, indicators.ErrInvalidCreateRequest) {
		return workbook.InvalidPayloadFailure("payload", "at_least_one_value_required"), true
	}
	var validation *indicators.IndicatorCreateValidationError
	if errors.As(err, &validation) {
		return indicatorValidationFailure(validation), true
	}
	return nil, false
}

func indicatorValidationFailure(validation *indicators.IndicatorCreateValidationError) *workbook.MutationFailure {
	if validation == nil {
		return nil
	}
	return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode)
}

func indicatorMutationResult(
	result indicators.CreateResult,
	incidentID uuid.UUID,
	clientTxnID string,
) workbook.MutationResult {
	statusCode := http.StatusOK
	if result.Created {
		statusCode = http.StatusCreated
	}
	return workbook.MutationResult{
		Payload: workbook.BuildMutationPayload(
			indicators.ViewSchemaID,
			result.ChangeSetID,
			result.CanonicalRow,
		),
		StatusCode:   statusCode,
		Replayed:     result.Replayed,
		IncidentID:   incidentID,
		RecordID:     result.RecordID,
		ChangeSetID:  result.ChangeSetID,
		ClientTxnID:  clientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: indicators.ViewSchemaID,
	}
}

func preferredRequestHash(provided []byte, derived []byte) []byte {
	if len(provided) > 0 {
		return append([]byte(nil), provided...)
	}
	return append([]byte(nil), derived...)
}
