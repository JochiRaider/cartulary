package workbookassembly

import (
	"context"
	"fmt"
	"io"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
)

type timelineProviderSet struct {
	create   workbook.CreateProvider
	patch    workbook.PatchProvider
	conflict workbook.ConflictProvider
}

func newTimelineProviderSet(owner *timeline.Facade) (timelineProviderSet, error) {
	if owner == nil {
		return timelineProviderSet{}, fmt.Errorf("compose Timeline Workbook adapters: owner is required")
	}
	create, err := newTimelineCreateProvider(owner)
	if err != nil {
		return timelineProviderSet{}, err
	}
	patch, err := newTimelinePatchProvider(owner)
	if err != nil {
		return timelineProviderSet{}, err
	}
	conflict, err := newTimelineConflictProvider(owner)
	if err != nil {
		return timelineProviderSet{}, err
	}
	return timelineProviderSet{create: create, patch: patch, conflict: conflict}, nil
}

type timelineCreateAdmission struct{ request timeline.CreateRequest }

func (value timelineCreateAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newTimelineCreateProvider(owner *timeline.Facade) (workbook.CreateProvider, error) {
	return workbook.NewCreateProvider(
		func(reader io.Reader) (workbook.CreateAdmission, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeTimelineCreateRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return timelineCreateAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(timelineCreateAdmission)
			if !ok || command.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id")), nil
			}
			result, err := owner.CreateRow(ctx, timeline.CreateRowCommand{
				Actor: command.Actor, IncidentID: command.IncidentID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, timelineadmission.CreateRequestHash(admitted.request)),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, admitted.request.ClientTxnID, "", ""); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

type timelinePatchAdmission struct{ request timeline.PatchRequest }

func (value timelinePatchAdmission) ClientTransactionID() string { return value.request.ClientTxnID }
func (value timelinePatchAdmission) AdmittedViewSchemaID() string {
	return timeline.TimelineViewSchemaID
}
func (value timelinePatchAdmission) AdmittedBaseRowVersion() int64 {
	return value.request.BaseRowVersion
}

func newTimelinePatchProvider(owner *timeline.Facade) (workbook.PatchProvider, error) {
	return workbook.NewPatchProvider(
		func(reader io.Reader) (workbook.PatchAdmission, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeTimelinePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return timelinePatchAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.PatchCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(timelinePatchAdmission)
			if !ok || command.AuthoritativeRecordType != "timeline_event" {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.PatchRow(ctx, timeline.PatchRowCommand{
				Actor: command.Actor, RecordID: command.RecordID, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, timelineadmission.PatchRequestHash(admitted.request)),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, admitted.request.ClientTxnID, "changes", "superseded_terminal"); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

type timelineConflictAdmission struct {
	request timeline.ConflictResolveRequest
}

func (value timelineConflictAdmission) ClientTransactionID() string { return value.request.ClientTxnID }

func newTimelineConflictProvider(owner *timeline.Facade) (workbook.ConflictProvider, error) {
	return workbook.NewConflictProvider(
		func(reader io.Reader, token string, claims workbook.ConflictClaims) (workbook.ConflictAdmission, *workbook.MutationFailure, error) {
			if claims.RouteKey != timeline.ConflictResolveRouteKey || claims.ViewSchemaID != timeline.TimelineViewSchemaID {
				return nil, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			request, apiErr := timelineadmission.DecodeTimelineConflictResolveRequest(reader, token, timelineConflictTokenClaims(claims))
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return nil, failure, err
			}
			return timelineConflictAdmission{request: request}, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand) (workbook.MutationOutcome, error) {
			admitted, ok := command.Admission.(timelineConflictAdmission)
			if !ok || command.AuthoritativeRecordType != "timeline_event" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != timeline.ConflictResolveRouteKey || command.Claims.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			claims := timelineConflictTokenClaims(command.Claims)
			result, err := owner.ResolveConflict(ctx, timeline.ConflictResolveCommand{
				Actor: command.Actor, RecordID: command.RecordID, Claims: claims, Request: admitted.request,
				RequestHash: preferredRequestHash(command.RequestHash, timelineadmission.ConflictResolveRequestHash(claims, admitted.request)),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, admitted.request.ClientTxnID, "", ""); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

func timelineConflictTokenClaims(claims workbook.ConflictClaims) conflicttokens.ConflictTokenClaims {
	return conflicttokens.ConflictTokenClaims{
		Version: claims.Version, RecordID: claims.RecordID.String(), ViewSchemaID: claims.ViewSchemaID,
		RouteKey: claims.RouteKey, FieldKey: claims.FieldKey,
		ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion:          claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}
}
