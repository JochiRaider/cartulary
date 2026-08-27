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

type TimelineOperations interface {
	CreateRow(context.Context, timeline.CreateRowCommand) (timeline.MutationResult, error)
	PatchRow(context.Context, timeline.PatchRowCommand) (timeline.MutationResult, error)
	ResolveConflict(context.Context, timeline.ConflictResolveCommand) (timeline.MutationResult, error)
	ApplyClipboardPaste(context.Context, timeline.ClipboardPasteCommand) (timeline.BatchMutationResult, error)
	ApplyFillDown(context.Context, timeline.FillDownCommand) (timeline.BatchMutationResult, error)
	ApplyMultiRowTagAssignment(context.Context, timeline.MultiRowTagAssignmentCommand) (timeline.BatchMutationResult, error)
	SupersedeRow(context.Context, timeline.SupersedeCommand) (timeline.MutationResult, error)
}

func newTimelineProviderSet(owner TimelineOperations) (timelineProviderSet, error) {
	if isNilDependency(owner) {
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

func newTimelineCreateProvider(owner TimelineOperations) (workbook.CreateProvider, error) {
	return workbook.NewCreateProvider(
		func(reader io.Reader) (timeline.CreateRequest, bool, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeTimelineCreateRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return timeline.CreateRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(ctx context.Context, command workbook.CreateCommand, request timeline.CreateRequest) (workbook.MutationOutcome, error) {
			if command.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.InvalidPayloadFailure("view_schema_id", "invalid_view_schema_id")), nil
			}
			result, err := owner.CreateRow(ctx, timeline.CreateRowCommand{
				Actor: command.Actor, IncidentID: command.IncidentID, Request: request,
				RequestHash: timelineadmission.CreateRequestHash(request),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "", ""); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

func newTimelinePatchProvider(owner TimelineOperations) (workbook.PatchProvider, error) {
	return workbook.NewPatchProvider(
		func(reader io.Reader) (timeline.PatchRequest, bool, *workbook.MutationFailure, error) {
			request, apiErr := timelineadmission.DecodeTimelinePatchRequest(reader)
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return timeline.PatchRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(timeline.PatchRequest) string { return timeline.TimelineViewSchemaID },
		func(ctx context.Context, command workbook.PatchCommand, request timeline.PatchRequest) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "timeline_event" {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			result, err := owner.PatchRow(ctx, timeline.PatchRowCommand{
				Actor: command.Actor, RecordID: command.RecordID, Request: request,
				RequestHash: timelineadmission.PatchRequestHash(request),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "changes", "superseded_terminal"); safe {
				return workbook.RejectedMutation(failure), nil
			}
			if err != nil {
				return workbook.MutationOutcome{}, err
			}
			return workbook.SuccessfulRowMutation(timelineMutationResult(result, timeline.TimelineViewSchemaID)), nil
		},
	)
}

func newTimelineConflictProvider(owner TimelineOperations) (workbook.ConflictProvider, error) {
	return workbook.NewConflictProvider(
		func(reader io.Reader, token string, claims workbook.ConflictClaims) (timeline.ConflictResolveRequest, bool, *workbook.MutationFailure, error) {
			if claims.RouteKey != timeline.ConflictResolveRouteKey || claims.ViewSchemaID != timeline.TimelineViewSchemaID {
				return timeline.ConflictResolveRequest{}, false, workbook.InvalidPayloadFailure("conflict_token", "invalid_value"), nil
			}
			request, apiErr := timelineadmission.DecodeTimelineConflictResolveRequest(reader, token, timelineConflictTokenClaims(claims))
			if apiErr != nil {
				failure, err := workbook.DecodeMutationFailure(apiErr)
				return timeline.ConflictResolveRequest{}, false, failure, err
			}
			return request, true, nil, nil
		},
		func(ctx context.Context, command workbook.ConflictCommand, request timeline.ConflictResolveRequest) (workbook.MutationOutcome, error) {
			if command.AuthoritativeRecordType != "timeline_event" || command.RecordID != command.Claims.RecordID ||
				command.Claims.RouteKey != timeline.ConflictResolveRouteKey || command.Claims.ViewSchemaID != timeline.TimelineViewSchemaID {
				return workbook.RejectedMutation(workbook.TargetNotFoundFailure()), nil
			}
			claims := timelineConflictTokenClaims(command.Claims)
			result, err := owner.ResolveConflict(ctx, timeline.ConflictResolveCommand{
				Actor: command.Actor, RecordID: command.RecordID, Claims: claims, Request: request,
				RequestHash: timelineadmission.ConflictResolveRequestHash(claims, request),
				RequestID:   command.RequestID, Now: command.Now,
			})
			if failure, safe := timelineActionFailure(err, request.ClientTxnID, "", ""); safe {
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
