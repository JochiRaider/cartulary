package workbookassembly

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type timelineClipboardValue struct {
	request timelineadmission.ClipboardPasteRequest
	plan    tabularingest.TabularRowPlanV1
}

func newTimelineClipboardProvider(owner TimelineOperations) (workbook.ClipboardProvider, error) {
	return workbook.NewClipboardProvider(
		func(reader io.Reader) (timelineClipboardValue, bool, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeClipboardPasteRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return timelineClipboardValue{}, false, failure, err
			}
			plan, err := timelineadmission.BuildClipboardPlan(request)
			if err != nil {
				return timelineClipboardValue{}, false, workbook.InvalidPayloadFailure("clipboard_text", "invalid_value"), nil
			}
			return timelineClipboardValue{request: request, plan: plan}, true, nil, nil
		},
		func(ctx context.Context, command workbook.ClipboardCommand, admitted timelineClipboardValue) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != timeline.TimelineViewSchemaID || admitted.request.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id")), nil
			}
			request := admitted.request
			result, err := owner.ApplyClipboardPaste(ctx, timeline.ClipboardPasteCommand{
				Actor: command.Actor, IncidentID: command.IncidentID, ClientTxnID: request.ClientTxnID,
				Plan: admitted.plan, Targets: request.Targets,
				RequestHash: timelineadmission.ClipboardPasteRequestHash(request),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "clipboard_text", ""); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulBatchMutation(timelineBatchResult(result)), nil
		},
	)
}

func newTimelineBulkProvider(owner TimelineOperations) (workbook.BulkProvider, error) {
	return workbook.NewBulkProvider(
		func(reader io.Reader) (timelineadmission.BulkMutationRequest, bool, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeBulkMutationRequest(reader, timeline.TimelineViewSchemaID)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return timelineadmission.BulkMutationRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(ctx context.Context, command workbook.BulkCommand, request timelineadmission.BulkMutationRequest) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != timeline.TimelineViewSchemaID || request.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id")), nil
			}
			var result timeline.BatchMutationResult
			var err error
			switch request.Kind {
			case timeline.OwnerBatchOperationFillDownV1:
				result, err = owner.ApplyFillDown(ctx, timeline.FillDownCommand{
					Actor: command.Actor, IncidentID: command.IncidentID, ClientTxnID: request.ClientTxnID,
					FieldKey: request.FieldKey, Value: request.Value, Targets: request.Targets,
					RequestHash: timelineadmission.BulkMutationRequestHash(request), RequestID: command.RequestID, Now: command.Now,
				})
			case timeline.OwnerBatchOperationMultiRowTagAssignmentV1:
				result, err = owner.ApplyMultiRowTagAssignment(ctx, timeline.MultiRowTagAssignmentCommand{
					Actor: command.Actor, IncidentID: command.IncidentID, ClientTxnID: request.ClientTxnID,
					TagName: request.TagName, NormalizedTag: request.NormalizedTag, Targets: request.Targets,
					RequestHash: timelineadmission.BulkMutationRequestHash(request), RequestID: command.RequestID, Now: command.Now,
				})
			default:
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("kind", "invalid_value")), nil
			}
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "value", ""); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulBatchMutation(timelineBatchResult(result)), nil
		},
	)
}

func newLinkedNoteProvider(owner *artifacts.MutationFacade) (workbook.LinkedNoteProvider, error) {
	return workbook.NewLinkedNoteProvider(
		func(reader io.Reader) (artifacts.ContextualNoteAdmission, bool, *workbook.MutationFailure, error) {
			admission, admissionErr := artifacts.AdmitContextualNote(reader)
			if admissionErr != nil {
				return artifacts.ContextualNoteAdmission{}, false, artifactAdmissionFailure(admissionErr), nil
			}
			return admission, true, nil, nil
		},
		func(ctx context.Context, command workbook.LinkedNoteCommand, admission artifacts.ContextualNoteAdmission) (workbook.MutationOutcome, error) {
			result, err := owner.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{
				ActorUserID: command.Actor.ID, SourceRecordID: command.Target.RecordID,
				Admission: admission, RequestID: command.RequestID,
				Now: command.Now,
			})
			if failure, safe := artifactMutationFailure(err, admission.ClientTxnID()); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulLinkedNoteMutation(artifactMutationResult(result)), nil
		},
	)
}

func decodeTimelineSupersede(reader io.Reader) (timeline.SupersedeRequest, bool, *workbook.MutationFailure, error) {
	request, apiErr := timelineadmission.DecodeTimelineSupersedeRequest(reader)
	if apiErr != nil {
		failure, err := workbook.DecodeMutationFailure(apiErr)
		return timeline.SupersedeRequest{}, false, failure, err
	}
	return request, true, nil, nil
}

func newTimelineSupersedeProvider(owner TimelineOperations) (workbook.SupersedeProvider, error) {
	return workbook.NewSupersedeProvider(
		decodeTimelineSupersede,
		func(ctx context.Context, command workbook.SupersedeCommand, request timeline.SupersedeRequest) (workbook.MutationOutcome, error) {
			if command.Target.RecordType != "timeline_event" {
				return unsupportedSupersedeOutcome(), nil
			}
			result, err := owner.SupersedeRow(ctx, timeline.SupersedeCommand{
				Actor: command.Actor, RecordID: command.Target.RecordID, Request: request,
				RequestHash: timelineadmission.ActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "changes", "supersede_not_allowed"); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulSupersedeMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

func decodeDecisionSupersede(reader io.Reader) (tasksdecisions.SupersedeRequest, bool, *workbook.MutationFailure, error) {
	request, admissionFailure := tasksdecisions.AdmitSupersedeJSON(reader)
	if admissionFailure != nil {
		return tasksdecisions.SupersedeRequest{}, false, taskDecisionAdmissionFailure(admissionFailure), nil
	}
	return request, true, nil, nil
}

func newDecisionSupersedeProvider(owner *tasksdecisions.MutationFacade) (workbook.SupersedeProvider, error) {
	return workbook.NewSupersedeProvider(
		decodeDecisionSupersede,
		func(ctx context.Context, command workbook.SupersedeCommand, request tasksdecisions.SupersedeRequest) (workbook.MutationOutcome, error) {
			if command.Target.RecordType != "decision" {
				return unsupportedSupersedeOutcome(), nil
			}
			result, err := owner.SupersedeDecision(ctx, tasksdecisions.SupersedeCommand{
				ActorUserID: command.Actor.ID, TargetRecordID: command.Target.RecordID,
				Request:     request,
				RequestHash: tasksdecisions.SupersedeRequestHash(request),
				RequestID:   command.RequestID, RouteKey: "workbook.records.supersede", Now: command.Now,
			})
			if failure, safe := taskDecisionActionFailure(err, request.ClientTxnID); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulSupersedeMutation(decisionSupersedeResult(result)), nil
		},
	)
}

func unsupportedSupersedeOutcome() workbook.MutationOutcome {
	return workbook.RejectedMutation(workbook.IllegalTransitionFailure(
		"", "", "supersede_not_allowed", []string{"unsupported_record_type"},
	))
}

func timelineBatchResult(result timeline.BatchMutationResult) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, ChangeSetID: result.ChangeSetID, ClientTxnID: result.ClientTxnID,
	}
}

func timelineMutationResult(result timeline.MutationResult, viewSchemaID string) workbook.MutationResult {
	return workbook.MutationResult{
		Payload: result.Payload, StatusCode: result.StatusCode, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: viewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func artifactMutationResult(result artifacts.MutationResult) workbook.MutationResult {
	payload := map[string]any{"view_schema_id": result.ViewSchemaID, "row": result.Row}
	changeSetID := uuid.Nil
	if result.ChangeSetID != nil {
		changeSetID = *result.ChangeSetID
		payload["change_set_id"] = result.ChangeSetID.String()
	}
	if result.ContextualLink != nil {
		payload["source_record_id"] = result.ContextualLink.SourceRecordID.String()
		payload["link_type"] = result.ContextualLink.LinkType
	}
	statusCode := 200
	if result.Outcome == artifacts.MutationOutcomeCreated {
		statusCode = 201
	}
	return workbook.MutationResult{
		Payload: payload, StatusCode: statusCode, Replayed: result.Outcome == artifacts.MutationOutcomeReplayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: changeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func decisionSupersedeResult(result tasksdecisions.SupersedeMutationResult) workbook.MutationResult {
	additional := make([]workbook.MutationResult, 0, len(result.AdditionalRecordChanges))
	for _, change := range result.AdditionalRecordChanges {
		additional = append(additional, decisionSupersedeResult(change))
	}
	payload := map[string]any{"row": result.Row}
	if result.Row == nil {
		payload = map[string]any{
			"view_schema_id": tasksdecisions.DecisionsViewSchemaID,
			"change_set_id":  result.ChangeSetID.String(), "target_record_id": result.Facts.TargetRecordID.String(),
			"superseding_record_id": result.Facts.SupersedingRecordID.String(),
			"target_row_version":    result.Facts.TargetRowVersion, "superseding_row_version": result.Facts.SupersedingRowVersion,
			"target_status": result.Facts.TargetStatus, "reason": result.Facts.Reason,
		}
	}
	return workbook.MutationResult{
		Payload: payload, StatusCode: 200, Replayed: result.Replayed,
		IncidentID: result.IncidentID, RecordID: result.RecordID, ChangeSetID: result.ChangeSetID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...), AdditionalRecordChanges: additional,
	}
}

func timelineActionFailure(err error, clientTxnID string, invalidField string, transitionReason string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return workbook.ClientTxnConflictFailure(clientTxnID), true
	}
	if errors.Is(err, timeline.ErrIncidentClosed) {
		return workbook.IncidentClosedFailure(), true
	}
	if errors.Is(err, timeline.ErrRecordNotFound) {
		return workbook.TargetNotFoundFailure(), true
	}
	if errors.Is(err, timeline.ErrRecordDeleted) {
		return workbook.RecordDeletedFailure(), true
	}
	var rowConflict *timeline.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion), true
	}
	var sameConflict *timeline.SameFieldConflictError
	if errors.As(err, &sameConflict) {
		conflict := sameConflict.Conflict
		failure, conversionErr := workbook.SameFieldConflictFailure(workbook.SameFieldConflictInput{
			ConflictToken: conflict.ConflictToken, RecordID: conflict.RecordID,
			FieldKey:                conflict.FieldKey,
			ConflictResolutionClass: workbook.SameFieldConflictClass(conflict.ConflictResolutionClass),
			BaseRowVersion:          conflict.BaseRowVersion, CurrentRowVersion: conflict.CurrentRowVersion,
			ClientValue: conflict.ClientValue, ServerValue: conflict.ServerValue,
			BaseValue: workbook.OptionalConflictValue{
				Present: conflict.BaseValue.Present, Value: conflict.BaseValue.Value,
			},
			ServerUpdatedBy: conflict.ServerUpdatedBy, ServerUpdatedAt: conflict.ServerUpdatedAt,
			SuggestedMergedValue: workbook.OptionalConflictValue{
				Present: conflict.SuggestedMergedValue.Present, Value: conflict.SuggestedMergedValue.Value,
			},
		})
		return failure, conversionErr == nil
	}
	var transition *timeline.IllegalTransitionError
	if errors.As(err, &transition) {
		reason := transition.ReasonCode
		if reason == "" {
			reason = transitionReason
		}
		return workbook.IllegalTransitionFailure(transition.FromStatus, transition.ToStatus, reason, transition.ViolatedGuards), true
	}
	if errors.Is(err, timeline.ErrIllegalTransition) {
		return workbook.IllegalTransitionFailure("", "", transitionReason, nil), true
	}
	if errors.Is(err, timeline.ErrNoEffectiveChange) {
		if invalidField == "" {
			invalidField = "changes"
		}
		return workbook.NoEffectiveChangeFailure(invalidField), true
	}
	var mentionTransition *mentions.MentionTransitionError
	if errors.As(err, &mentionTransition) {
		return workbook.IllegalTransitionFailure(
			mentionTransition.FromStatus, mentionTransition.ToStatus, "", mentionTransition.ViolatedGuards,
		), true
	}
	var entityConflict *hostidentity.ExactMatchConflictError
	if errors.As(err, &entityConflict) {
		return workbook.EntityMatchConflictFailure(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords), true
	}
	var mentionTarget *mentions.MentionTargetValidationError
	if errors.As(err, &mentionTarget) || errors.Is(err, mentions.ErrEntityMentionNotFound) ||
		errors.Is(err, mentions.ErrResolvedRecordNotFound) || errors.Is(err, mentions.ErrInvalidMentionResolution) {
		return workbook.InvalidPayloadFailure(invalidField, "invalid_value"), true
	}
	return nil, false
}

func taskDecisionActionFailure(err error, clientTxnID string) (*workbook.MutationFailure, bool) {
	if err == nil {
		return nil, false
	}
	if errors.Is(err, tasksdecisions.ErrClientTxnConflict) || errors.Is(err, authn.ErrClientTxnConflict) {
		return workbook.ClientTxnConflictFailure(clientTxnID), true
	}
	if admission.IsDenied(err, admission.DenialIncidentClosed) {
		return workbook.IncidentClosedFailure(), true
	}
	if errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
		return workbook.RecordDeletedFailure(), true
	}
	var rowConflict *tasksdecisions.RowVersionConflictError
	if errors.As(err, &rowConflict) {
		return workbook.RowVersionConflictFailure(rowConflict.RecordID, rowConflict.BaseRowVersion, rowConflict.CurrentRowVersion), true
	}
	var validation *tasksdecisions.ValidationError
	if errors.As(err, &validation) {
		return workbook.InvalidPayloadFailure(validation.Field, validation.ReasonCode), true
	}
	var lifecycle *tasksdecisions.LifecycleValidationError
	if errors.As(err, &lifecycle) {
		return workbook.IllegalTransitionFailure(lifecycle.FromStatus, lifecycle.ToStatus, lifecycle.ReasonCode, lifecycle.ViolatedGuards), true
	}
	return nil, false
}
