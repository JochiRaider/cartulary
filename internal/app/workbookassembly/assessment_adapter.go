package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentadmission "github.com/JochiRaider/cartulary/internal/modules/assessments/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
)

const assessmentCreateRouteKey = "assessments.rows.create"

func newAssessmentCreateProvider(owner *assessments.Facade) (workbook.CreateProvider, error) {
	if owner == nil {
		return nil, fmt.Errorf("compose Assessment Workbook adapter: owner is required")
	}
	return workbook.NewCreateProvider(
		func(reader io.Reader) (assessments.CreateInput, bool, *workbook.MutationFailure, error) {
			input, apiErr := assessmentadmission.DecodeCreateRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return assessments.CreateInput{}, false, failure, err
			}
			return input, true, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand, input assessments.CreateInput) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != assessments.AssessmentsViewSchemaID {
				return workbook.RejectedMutation(
					workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id"),
				), nil
			}
			result, err := owner.Create(ctx, assessments.CreateCommand{
				ActorUserID: command.Actor.ID,
				IncidentID:  command.IncidentID,
				Input:       input,
				Idempotency: assessments.CreateIdempotencyKey{
					RouteKey:    assessmentCreateRouteKey,
					ActorUserID: command.Actor.ID,
					ScopeKey:    command.IncidentID.String() + ":" + assessments.AssessmentsViewSchemaID,
					ClientTxnID: input.ClientTxnID,
					RequestHash: assessmentadmission.CreateRequestHash(input),
				},
				RequestID: command.RequestID,
				Now:       command.Now,
			})
			if failure, safe := assessmentCreateFailure(err, input.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(assessmentMutationResult(
				result,
				command.IncidentID,
				input.ClientTxnID,
			)), nil
		},
	)
}

func assessmentCreateFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, assessments.ErrClientTxnConflict) {
		return workbook.ClientTxnConflictFailure(clientTxnID), true
	}
	var validation *assessments.CreateValidationError
	if errors.As(err, &validation) {
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	return nil, false
}

func assessmentMutationResult(
	result assessments.CreateResult,
	incidentID uuid.UUID,
	clientTxnID string,
) workbook.MutationResult {
	statusCode := http.StatusCreated
	if result.Outcome == assessments.CreateOutcomeReplayed {
		statusCode = http.StatusOK
	}
	return workbook.MutationResult{
		Payload: workbook.BuildMutationPayload(
			assessments.AssessmentsViewSchemaID,
			result.ChangeSetID,
			result.CanonicalRow,
		),
		StatusCode:   statusCode,
		Replayed:     result.Outcome == assessments.CreateOutcomeReplayed,
		IncidentID:   incidentID,
		RecordID:     result.RecordID,
		ChangeSetID:  result.ChangeSetID,
		ClientTxnID:  clientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: assessments.AssessmentsViewSchemaID,
	}
}
