package workbook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type CreateAdmission struct {
	ClientTxnID string
	normalized  any
}

type CreateCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	Admission    CreateAdmission
	RequestHash  []byte
	RequestID    string
	Now          time.Time
}

type CreateProvider interface {
	DecodeCreate(io.Reader) (CreateAdmission, *httpapi.APIError)
	Create(context.Context, CreateCommand) (MutationResult, error)
}

type PatchAdmission struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	normalized     any
}

type PatchCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	Admission               PatchAdmission
	RequestHash             []byte
	RequestID               string
	Now                     time.Time
}

type PatchProvider interface {
	DecodePatch(io.Reader) (PatchAdmission, *httpapi.APIError)
	Patch(context.Context, PatchCommand) (MutationResult, error)
}

type ConflictAdmission struct {
	ClientTxnID string
	normalized  any
}

type ConflictCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	Claims                  conflicttokens.ConflictTokenClaims
	Admission               ConflictAdmission
	RequestHash             []byte
	RequestID               string
	Now                     time.Time
}

type ConflictProvider interface {
	DecodeConflict(
		io.Reader,
		string,
		conflicttokens.ConflictTokenClaims,
	) (ConflictAdmission, *httpapi.APIError)
	ResolveConflict(context.Context, ConflictCommand) (MutationResult, error)
}

type publicMutationError struct {
	apiError *httpapi.APIError
}

func (e *publicMutationError) Error() string {
	if e == nil || e.apiError == nil {
		return "workbook: public mutation error"
	}
	return e.apiError.Code
}

type createProvider struct {
	decode func(io.Reader) (CreateAdmission, *httpapi.APIError)
	create func(context.Context, CreateCommand) (MutationResult, error)
}

func (p createProvider) DecodeCreate(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
	return p.decode(reader)
}

func (p createProvider) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	return p.create(ctx, command)
}

type patchProvider struct {
	decode func(io.Reader) (PatchAdmission, *httpapi.APIError)
	patch  func(context.Context, PatchCommand) (MutationResult, error)
}

type conflictProvider struct {
	decode func(
		io.Reader,
		string,
		conflicttokens.ConflictTokenClaims,
	) (ConflictAdmission, *httpapi.APIError)
	resolve func(context.Context, ConflictCommand) (MutationResult, error)
}

func (p patchProvider) DecodePatch(reader io.Reader) (PatchAdmission, *httpapi.APIError) {
	return p.decode(reader)
}

func (p patchProvider) Patch(ctx context.Context, command PatchCommand) (MutationResult, error) {
	return p.patch(ctx, command)
}

func (p conflictProvider) DecodeConflict(
	reader io.Reader,
	token string,
	claims conflicttokens.ConflictTokenClaims,
) (ConflictAdmission, *httpapi.APIError) {
	return p.decode(reader, token, claims)
}

func (p conflictProvider) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	return p.resolve(ctx, command)
}

func NewTimelineCreateProvider(owner *timeline.Facade) CreateProvider {
	return createProvider{
		decode: func(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
			request, apiErr := timelineadmission.DecodeTimelineCreateRequest(reader)
			return CreateAdmission{ClientTxnID: request.ClientTxnID, normalized: request}, apiErr
		},
		create: func(ctx context.Context, command CreateCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(timeline.CreateRequest)
			if !ok || command.ViewSchemaID != timeline.TimelineViewSchemaID {
				return MutationResult{}, mutationValidationError("view_schema_id", "invalid_view_schema_id")
			}
			result, err := owner.CreateRow(ctx, timeline.CreateRowCommand{
				Actor:       command.Actor,
				IncidentID:  command.IncidentID,
				Request:     request,
				RequestHash: requestHash(command.RequestHash, timelineadmission.CreateRequestHash(request)),
				RequestID:   command.RequestID,
				Now:         command.Now,
			})
			return mutationResultFromTimeline(result, timeline.TimelineViewSchemaID), adaptTimelineMutationError(err, request.ClientTxnID, "", "")
		},
	}
}

func NewHostCreateProvider(owner *hostidentity.Store) CreateProvider {
	return newEntityCreateProvider(hostidentity.HostsViewSchemaID, owner.CreateHostRow)
}

func NewIdentityCreateProvider(owner *hostidentity.Store) CreateProvider {
	return newEntityCreateProvider(hostidentity.IdentitiesViewSchemaID, owner.CreateIdentityRow)
}

type entityCreateFunc func(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	hostidentity.CreateRequest,
	[]byte,
	string,
	time.Time,
) (hostidentity.MutationResult, error)

func newEntityCreateProvider(viewSchemaID string, create entityCreateFunc) CreateProvider {
	return createProvider{
		decode: func(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
			request, apiErr := hostidentity.DecodeCreateRequest(viewSchemaID, reader)
			return CreateAdmission{ClientTxnID: request.ClientTxnID, normalized: request}, apiErr
		},
		create: func(ctx context.Context, command CreateCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(hostidentity.CreateRequest)
			if !ok || command.ViewSchemaID != viewSchemaID {
				return MutationResult{}, mutationValidationError("view_schema_id", "invalid_view_schema_id")
			}
			result, err := create(
				ctx,
				command.Actor,
				command.IncidentID,
				request,
				requestHash(command.RequestHash, hostidentity.CreateRequestHash(viewSchemaID, request)),
				command.RequestID,
				command.Now,
			)
			if errors.Is(err, hostidentity.ErrInvalidCreateRequest) {
				err = mutationValidationError("payload", "at_least_one_value_required")
			}
			var conflict *hostidentity.ExactMatchConflictError
			if errors.As(err, &conflict) {
				err = &publicMutationError{apiError: entityMatchConflictError(
					conflict.EntityType,
					conflict.IdentifierClass,
					conflict.CandidateRecords,
				)}
			}
			return mutationResultFromEntityCreate(result, viewSchemaID, command.IncidentID, request.ClientTxnID), err
		},
	}
}

func NewIndicatorCreateProvider(owner *indicators.Store) CreateProvider {
	return createProvider{
		decode: func(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
			command, apiErr := decodeIndicatorCreate(reader)
			return CreateAdmission{ClientTxnID: command.ClientTxnID, normalized: command}, apiErr
		},
		create: func(ctx context.Context, command CreateCommand) (MutationResult, error) {
			indicatorCommand, ok := command.Admission.normalized.(indicators.CreateCommand)
			if !ok || command.ViewSchemaID != indicators.ViewSchemaID {
				return MutationResult{}, mutationValidationError("view_schema_id", "invalid_view_schema_id")
			}
			result, err := owner.CreateIndicatorRow(
				ctx,
				command.Actor,
				command.IncidentID,
				indicatorCommand,
				requestHash(command.RequestHash, indicatorCreateCommandHash(indicatorCommand)),
				command.RequestID,
				command.Now,
			)
			if errors.Is(err, indicators.ErrInvalidCreateRequest) {
				err = mutationValidationError("payload", "at_least_one_value_required")
			}
			var validation *indicators.IndicatorCreateValidationError
			if errors.As(err, &validation) {
				err = mutationValidationError(validation.Field, validation.ReasonCode)
			}
			return mutationResultFromSimpleCreate(
				result.Payload,
				result.StatusCode,
				result.Replayed,
				result.RecordID,
				result.ChangeSetID,
				result.RowVersion,
				command.IncidentID,
				indicatorCommand.ClientTxnID,
				indicators.ViewSchemaID,
			), err
		},
	}
}

func NewAssessmentCreateProvider(owner *assessments.Facade) CreateProvider {
	return createProvider{
		decode: func(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
			request, apiErr := DecodeCreateRequest(assessments.AssessmentsViewSchemaID, reader)
			return CreateAdmission{ClientTxnID: request.ClientTxnID, normalized: request}, apiErr
		},
		create: func(ctx context.Context, command CreateCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(CreateRequest)
			if !ok || command.ViewSchemaID != assessments.AssessmentsViewSchemaID {
				return MutationResult{}, mutationValidationError("view_schema_id", "invalid_view_schema_id")
			}
			input, err := assessmentCreateInputFromWorkbook(request)
			if err != nil {
				return MutationResult{}, err
			}
			const routeKey = "assessments.rows.create"
			result, err := owner.Create(ctx, assessments.CreateCommand{
				ActorUserID: command.Actor.ID,
				IncidentID:  command.IncidentID,
				Input:       input,
				Idempotency: assessments.CreateIdempotencyKey{
					RouteKey:    routeKey,
					ActorUserID: command.Actor.ID,
					ScopeKey:    command.IncidentID.String() + ":" + assessments.AssessmentsViewSchemaID,
					ClientTxnID: input.ClientTxnID,
					RequestHash: requestHash(command.RequestHash, assessmentCreateRequestHash(request)),
				},
				RequestID: command.RequestID,
				Now:       command.Now,
			})
			var validation *assessments.CreateValidationError
			if errors.As(err, &validation) {
				err = mutationValidationError(validation.Field, validation.ReasonCode)
			}
			if errors.Is(err, assessments.ErrClientTxnConflict) {
				err = authn.ErrClientTxnConflict
			}
			statusCode := 0
			if err == nil {
				statusCode = http.StatusCreated
				if result.Outcome == assessments.CreateOutcomeReplayed {
					statusCode = http.StatusOK
				}
			}
			payload := map[string]any(nil)
			if err == nil {
				payload = BuildMutationPayload(
					assessments.AssessmentsViewSchemaID,
					result.ChangeSetID,
					result.CanonicalRow,
				)
			}
			return mutationResultFromSimpleCreate(
				payload,
				statusCode,
				result.Outcome == assessments.CreateOutcomeReplayed,
				result.RecordID,
				result.ChangeSetID,
				result.RowVersion,
				command.IncidentID,
				request.ClientTxnID,
				assessments.AssessmentsViewSchemaID,
			), err
		},
	}
}

func assessmentCreateInputFromWorkbook(request CreateRequest) (assessments.CreateInput, error) {
	input := assessments.CreateInput{ClientTxnID: request.ClientTxnID}
	if value, ok := request.Values["assessment.subject_ref"]; ok {
		if value.Kind != "uuid" || value.UUID == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.subject_ref", "invalid_value")
		}
		input.SubjectRef = *value.UUID
	}
	if value, ok := request.Values["assessment.subject_type"]; ok {
		if value.Kind != "text" || value.Text == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.subject_type", "invalid_value")
		}
		input.SubjectType = *value.Text
	}
	if value, ok := request.Values["assessment.assessment_state"]; ok {
		if value.Kind != "text" || value.Text == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.assessment_state", "invalid_value")
		}
		input.AssessmentState = *value.Text
	}
	if value, ok := request.Values["assessment.confidence_score"]; ok {
		switch {
		case value.Kind == "null":
		case value.Kind == "number" && value.Number != nil:
			score := int(*value.Number)
			input.ConfidenceScore = &score
		default:
			return assessments.CreateInput{}, mutationValidationError("assessment.confidence_score", "invalid_value")
		}
	}
	if value, ok := request.Values["assessment.rationale"]; ok {
		if value.Kind != "text" || value.Text == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.rationale", "invalid_value")
		}
		input.Rationale = *value.Text
	}
	if value, ok := request.Values["assessment.assessor"]; ok {
		if value.Kind != "uuid" || value.UUID == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.assessor", "invalid_value")
		}
		assessor := *value.UUID
		input.Assessor = &assessor
	}
	if value, ok := request.Values["assessment.assessed_at"]; ok {
		if value.Kind != "timestamp" || value.Timestamp == nil {
			return assessments.CreateInput{}, mutationValidationError("assessment.assessed_at", "invalid_value")
		}
		assessedAt := value.Timestamp.UTC()
		input.AssessedAt = &assessedAt
	}
	if collection, ok := request.Collections["assessment.support_refs"]; ok {
		input.SupportRefs = make([]uuid.UUID, 0, len(collection.Actions))
		for _, action := range collection.Actions {
			if action.Op != "add_record_ref" || action.LinkedRecordID == nil {
				return assessments.CreateInput{}, mutationValidationError("assessment.support_refs", "invalid_value")
			}
			input.SupportRefs = append(input.SupportRefs, *action.LinkedRecordID)
		}
	}
	return input, nil
}

func assessmentCreateRequestHash(request CreateRequest) []byte {
	input, err := assessmentCreateInputFromWorkbook(request)
	if err != nil {
		return CreateRequestHash(request)
	}
	payload := map[string]any{
		"view_schema_id": assessments.AssessmentsViewSchemaID,
		"client_txn_id":  input.ClientTxnID,
	}
	if input.SubjectRef != uuid.Nil {
		payload["assessment.subject_ref"] = input.SubjectRef.String()
	}
	if input.SubjectType != "" {
		payload["assessment.subject_type"] = input.SubjectType
	}
	if input.AssessmentState != "" {
		payload["assessment.assessment_state"] = input.AssessmentState
	}
	if input.ConfidenceScore != nil {
		payload["assessment.confidence_score"] = *input.ConfidenceScore
	}
	if input.Rationale != "" {
		payload["assessment.rationale"] = input.Rationale
	}
	if input.Assessor != nil {
		payload["assessment.assessor"] = input.Assessor.String()
	}
	if input.AssessedAt != nil {
		payload["assessment.assessed_at"] = input.AssessedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(input.SupportRefs) > 0 {
		refs := make([]string, 0, len(input.SupportRefs))
		for _, ref := range input.SupportRefs {
			refs = append(refs, ref.String())
		}
		slices.Sort(refs)
		payload["assessment.support_refs"] = refs
	}
	return hashRequestPayload(payload)
}

func NewArtifactCreateProvider(viewSchemaID string, owner *artifacts.WorkbookFacade) CreateProvider {
	return newGenericCreateProvider(viewSchemaID, func(ctx context.Context, command CreateCommand, request CreateRequest) (MutationResult, error) {
		result, err := owner.Create(ctx, artifacts.WorkbookCreateCommand{
			Actor:       command.Actor,
			IncidentID:  command.IncidentID,
			Request:     artifactCreateRequestFromWorkbook(request),
			RequestHash: requestHash(command.RequestHash, CreateRequestHash(request)),
			RequestID:   command.RequestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         command.Now,
		})
		return mutationResultFromArtifactWorkbook(result), adaptArtifactWorkbookOwnerError(err)
	})
}

func NewEvidenceCreateProvider(owner evidence.WorkbookContribution) CreateProvider {
	return newGenericCreateProvider(EvidenceViewSchemaID, func(ctx context.Context, command CreateCommand, request CreateRequest) (MutationResult, error) {
		result, err := owner.Create(ctx, evidence.WorkbookCreateCommand{
			Actor:       command.Actor,
			IncidentID:  command.IncidentID,
			Request:     evidenceCreateRequestFromWorkbook(request),
			RequestHash: requestHash(command.RequestHash, CreateRequestHash(request)),
			RequestID:   command.RequestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         command.Now,
		})
		return mutationResultFromEvidenceWorkbook(result), adaptEvidenceWorkbookOwnerError(err)
	})
}

func NewPartyCreateProvider(owner *parties.WorkbookFacade) CreateProvider {
	return newGenericCreateProvider(PartiesViewSchemaID, func(ctx context.Context, command CreateCommand, request CreateRequest) (MutationResult, error) {
		result, err := owner.Create(ctx, parties.WorkbookCreateCommand{
			Actor:       command.Actor,
			IncidentID:  command.IncidentID,
			Request:     partyCreateRequestFromWorkbook(request),
			RequestHash: requestHash(command.RequestHash, CreateRequestHash(request)),
			RequestID:   command.RequestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         command.Now,
		})
		return mutationResultFromPartyWorkbook(result), adaptPartyWorkbookOwnerError(err)
	})
}

type taskDecisionCreateOwner interface {
	Create(context.Context, tasksdecisions.WorkbookCreateCommand) (tasksdecisions.WorkbookMutationResult, error)
}

func NewTaskDecisionCreateProvider(viewSchemaID string, owner taskDecisionCreateOwner) CreateProvider {
	return newGenericCreateProvider(viewSchemaID, func(ctx context.Context, command CreateCommand, request CreateRequest) (MutationResult, error) {
		result, err := owner.Create(ctx, tasksdecisions.WorkbookCreateCommand{
			ActorUserID: command.Actor.ID,
			IncidentID:  command.IncidentID,
			Request:     taskDecisionCreateRequestFromWorkbook(request),
			RequestHash: requestHash(command.RequestHash, CreateRequestHash(request)),
			RequestID:   command.RequestID,
			RouteKey:    workbookCreateRouteKey,
			Now:         command.Now,
		})
		return mutationResultFromTaskDecisionWorkbook(result), adaptTaskDecisionWorkbookOwnerError(err)
	})
}

type genericCreateFunc func(context.Context, CreateCommand, CreateRequest) (MutationResult, error)

func newGenericCreateProvider(viewSchemaID string, create genericCreateFunc) CreateProvider {
	return createProvider{
		decode: func(reader io.Reader) (CreateAdmission, *httpapi.APIError) {
			request, apiErr := DecodeCreateRequest(viewSchemaID, reader)
			return CreateAdmission{ClientTxnID: request.ClientTxnID, normalized: request}, apiErr
		},
		create: func(ctx context.Context, command CreateCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(CreateRequest)
			if !ok || request.ViewSchemaID != viewSchemaID || command.ViewSchemaID != viewSchemaID {
				return MutationResult{}, mutationValidationError("view_schema_id", "invalid_view_schema_id")
			}
			return create(ctx, command, request)
		},
	}
}

func NewTimelinePatchProvider(owner *timeline.Facade) PatchProvider {
	return patchProvider{
		decode: func(reader io.Reader) (PatchAdmission, *httpapi.APIError) {
			request, apiErr := timelineadmission.DecodeTimelinePatchRequest(reader)
			return PatchAdmission{
				ViewSchemaID:   timeline.TimelineViewSchemaID,
				BaseRowVersion: request.BaseRowVersion,
				ClientTxnID:    request.ClientTxnID,
				normalized:     request,
			}, apiErr
		},
		patch: func(ctx context.Context, command PatchCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(timeline.PatchRequest)
			if !ok || command.AuthoritativeRecordType != "timeline_event" {
				return MutationResult{}, pgx.ErrNoRows
			}
			result, err := owner.PatchRow(ctx, timeline.PatchRowCommand{
				Actor:       command.Actor,
				RecordID:    command.RecordID,
				Request:     request,
				RequestHash: requestHash(command.RequestHash, timelineadmission.PatchRequestHash(request)),
				RequestID:   command.RequestID,
				Now:         command.Now,
			})
			return mutationResultFromTimeline(result, timeline.TimelineViewSchemaID), adaptTimelineMutationError(
				err,
				request.ClientTxnID,
				"superseded_terminal",
				"changes",
			)
		},
	}
}

func NewHostPatchProvider(owner *hostidentity.Store) PatchProvider {
	return newEntityPatchProvider("host", hostidentity.HostsViewSchemaID, owner)
}

func NewIdentityPatchProvider(owner *hostidentity.Store) PatchProvider {
	return newEntityPatchProvider("identity", hostidentity.IdentitiesViewSchemaID, owner)
}

func newEntityPatchProvider(recordType string, viewSchemaID string, owner *hostidentity.Store) PatchProvider {
	return patchProvider{
		decode: func(reader io.Reader) (PatchAdmission, *httpapi.APIError) {
			request, apiErr := hostidentity.DecodePatchRequest(reader)
			return PatchAdmission{
				ViewSchemaID:   request.ViewSchemaID,
				BaseRowVersion: request.BaseRowVersion,
				ClientTxnID:    request.ClientTxnID,
				normalized:     request,
			}, apiErr
		},
		patch: func(ctx context.Context, command PatchCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(hostidentity.PatchRequest)
			if !ok || command.AuthoritativeRecordType != recordType || request.ViewSchemaID != viewSchemaID {
				return MutationResult{}, pgx.ErrNoRows
			}
			result, err := owner.PatchEntityRow(
				ctx,
				command.Actor,
				command.RecordID,
				request,
				requestHash(command.RequestHash, hostidentity.PatchRequestHash(request)),
				command.RequestID,
				command.Now,
				workbookPatchRouteKey,
			)
			return mutationResultFromEntityPatch(result, viewSchemaID), adaptEntityPatchOwnerError(err)
		},
	}
}

func NewArtifactPatchProvider(owner *artifacts.WorkbookFacade) PatchProvider {
	return newGenericPatchProvider("artifact", artifactViewSchemaIDs(), func(ctx context.Context, command PatchCommand, request PatchRequest) (MutationResult, error) {
		result, err := owner.Patch(ctx, artifacts.WorkbookPatchCommand{
			Actor:            command.Actor,
			RecordID:         command.RecordID,
			Request:          artifactPatchRequestFromWorkbook(request),
			RequestHash:      requestHash(command.RequestHash, PatchRequestHash(request)),
			RequestID:        command.RequestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              command.Now,
		})
		return mutationResultFromArtifactWorkbook(result), adaptArtifactWorkbookOwnerError(err)
	})
}

func NewEvidencePatchProvider(owner evidence.WorkbookContribution) PatchProvider {
	return newGenericPatchProvider("evidence", []string{EvidenceViewSchemaID}, func(ctx context.Context, command PatchCommand, request PatchRequest) (MutationResult, error) {
		result, err := owner.Patch(ctx, evidence.WorkbookPatchCommand{
			Actor:            command.Actor,
			RecordID:         command.RecordID,
			Request:          evidencePatchRequestFromWorkbook(request),
			RequestHash:      requestHash(command.RequestHash, PatchRequestHash(request)),
			RequestID:        command.RequestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              command.Now,
		})
		return mutationResultFromEvidenceWorkbook(result), adaptEvidenceWorkbookOwnerError(err)
	})
}

func NewPartyPatchProvider(owner *parties.WorkbookFacade) PatchProvider {
	return newGenericPatchProvider("party", []string{PartiesViewSchemaID}, func(ctx context.Context, command PatchCommand, request PatchRequest) (MutationResult, error) {
		result, err := owner.Patch(ctx, parties.WorkbookPatchCommand{
			Actor:            command.Actor,
			RecordID:         command.RecordID,
			Request:          partyPatchRequestFromWorkbook(request),
			RequestHash:      requestHash(command.RequestHash, PatchRequestHash(request)),
			RequestID:        command.RequestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              command.Now,
		})
		return mutationResultFromPartyWorkbook(result), adaptPartyWorkbookOwnerError(err)
	})
}

type taskDecisionPatchOwner interface {
	Patch(context.Context, tasksdecisions.WorkbookPatchCommand) (tasksdecisions.WorkbookMutationResult, error)
}

func NewTaskDecisionPatchProvider(recordType string, viewSchemaID string, owner taskDecisionPatchOwner) PatchProvider {
	return newGenericPatchProvider(recordType, []string{viewSchemaID}, func(ctx context.Context, command PatchCommand, request PatchRequest) (MutationResult, error) {
		result, err := owner.Patch(ctx, tasksdecisions.WorkbookPatchCommand{
			ActorUserID:      command.Actor.ID,
			RecordID:         command.RecordID,
			Request:          taskDecisionPatchRequestFromWorkbook(request),
			RequestHash:      requestHash(command.RequestHash, PatchRequestHash(request)),
			RequestID:        command.RequestID,
			RouteKey:         workbookPatchRouteKey,
			ConflictRouteKey: workbookConflictResolveRouteKey,
			Now:              command.Now,
		})
		return mutationResultFromTaskDecisionWorkbook(result), adaptTaskDecisionWorkbookOwnerError(err)
	})
}

type genericPatchFunc func(context.Context, PatchCommand, PatchRequest) (MutationResult, error)

func newGenericPatchProvider(
	recordType string,
	viewSchemaIDs []string,
	patch genericPatchFunc,
) PatchProvider {
	allowedViews := make(map[string]struct{}, len(viewSchemaIDs))
	for _, viewSchemaID := range viewSchemaIDs {
		allowedViews[viewSchemaID] = struct{}{}
	}
	return patchProvider{
		decode: func(reader io.Reader) (PatchAdmission, *httpapi.APIError) {
			request, apiErr := DecodePatchRequest(reader)
			return PatchAdmission{
				ViewSchemaID:   request.ViewSchemaID,
				BaseRowVersion: request.BaseRowVersion,
				ClientTxnID:    request.ClientTxnID,
				normalized:     request,
			}, apiErr
		},
		patch: func(ctx context.Context, command PatchCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(PatchRequest)
			_, allowed := allowedViews[request.ViewSchemaID]
			if !ok || !allowed || command.AuthoritativeRecordType != recordType {
				return MutationResult{}, pgx.ErrNoRows
			}
			return patch(ctx, command, request)
		},
	}
}

func NewTimelineConflictProvider(owner *timeline.Facade) ConflictProvider {
	return conflictProvider{
		decode: func(
			reader io.Reader,
			token string,
			claims conflicttokens.ConflictTokenClaims,
		) (ConflictAdmission, *httpapi.APIError) {
			if claims.RouteKey != timeline.ConflictResolveRouteKey ||
				claims.ViewSchemaID != timeline.TimelineViewSchemaID {
				return ConflictAdmission{}, invalidMutationPayload("conflict_token", "invalid_value")
			}
			request, apiErr := timelineadmission.DecodeTimelineConflictResolveRequest(
				reader,
				token,
				claims,
			)
			return ConflictAdmission{
				ClientTxnID: request.ClientTxnID,
				normalized:  request,
			}, apiErr
		},
		resolve: func(ctx context.Context, command ConflictCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(timeline.ConflictResolveRequest)
			if !ok || command.AuthoritativeRecordType != "timeline_event" {
				return MutationResult{}, pgx.ErrNoRows
			}
			result, err := owner.ResolveConflict(ctx, timeline.ConflictResolveCommand{
				Actor:       command.Actor,
				RecordID:    command.RecordID,
				Claims:      command.Claims,
				Request:     request,
				RequestHash: requestHash(command.RequestHash, timelineadmission.ConflictResolveRequestHash(command.Claims, request)),
				RequestID:   command.RequestID,
				Now:         command.Now,
			})
			return mutationResultFromTimeline(result, timeline.TimelineViewSchemaID),
				adaptTimelineMutationError(err, request.ClientTxnID, "", "")
		},
	}
}

func NewHostConflictProvider(owner *hostidentity.Store) ConflictProvider {
	return newEntityConflictProvider("host", hostidentity.HostsViewSchemaID, owner)
}

func NewIdentityConflictProvider(owner *hostidentity.Store) ConflictProvider {
	return newEntityConflictProvider("identity", hostidentity.IdentitiesViewSchemaID, owner)
}

func newEntityConflictProvider(
	recordType string,
	viewSchemaID string,
	owner *hostidentity.Store,
) ConflictProvider {
	return newGenericConflictProvider(
		recordType,
		[]string{viewSchemaID},
		func(
			ctx context.Context,
			command ConflictCommand,
			request ConflictResolveRequest,
			patch *PatchRequest,
		) (MutationResult, error) {
			var entityPatch *hostidentity.PatchRequest
			if patch != nil {
				converted, err := entityConflictPatchRequest(*patch)
				if err != nil {
					return MutationResult{}, err
				}
				entityPatch = &converted
			}
			result, err := owner.ResolveWorkbookConflict(ctx, hostidentity.WorkbookConflictCommand{
				Mechanics:      conflictMechanics(command, request.ClientTxnID),
				Actor:          command.Actor,
				ResolutionKind: request.ResolutionKind,
				Patch:          entityPatch,
				Now:            command.Now,
			})
			return mutationResultFromEntityPatch(result, viewSchemaID), adaptEntityPatchOwnerError(err)
		},
	)
}

func NewArtifactConflictProvider(owner *artifacts.WorkbookFacade) ConflictProvider {
	return newGenericConflictProvider(
		"artifact",
		artifactViewSchemaIDs(),
		func(
			ctx context.Context,
			command ConflictCommand,
			request ConflictResolveRequest,
			patch *PatchRequest,
		) (MutationResult, error) {
			var ownerPatch *artifacts.WorkbookPatchRequest
			if patch != nil {
				converted := artifactPatchRequestFromWorkbook(*patch)
				ownerPatch = &converted
			}
			result, err := owner.ResolveConflict(ctx, artifacts.WorkbookConflictCommand{
				Mechanics:      conflictMechanics(command, request.ClientTxnID),
				Actor:          command.Actor,
				ResolutionKind: request.ResolutionKind,
				Patch:          ownerPatch,
				Now:            command.Now,
			})
			return mutationResultFromArtifactWorkbook(result), adaptArtifactWorkbookOwnerError(err)
		},
	)
}

func NewEvidenceConflictProvider(owner evidence.WorkbookContribution) ConflictProvider {
	return newGenericConflictProvider(
		"evidence",
		[]string{EvidenceViewSchemaID},
		func(
			ctx context.Context,
			command ConflictCommand,
			request ConflictResolveRequest,
			patch *PatchRequest,
		) (MutationResult, error) {
			var ownerPatch *evidence.WorkbookPatchRequest
			if patch != nil {
				converted := evidencePatchRequestFromWorkbook(*patch)
				ownerPatch = &converted
			}
			result, err := owner.ResolveConflict(ctx, evidence.WorkbookConflictCommand{
				Mechanics:      conflictMechanics(command, request.ClientTxnID),
				Actor:          command.Actor,
				ResolutionKind: request.ResolutionKind,
				Patch:          ownerPatch,
				Now:            command.Now,
			})
			return mutationResultFromEvidenceWorkbook(result), adaptEvidenceWorkbookOwnerError(err)
		},
	)
}

func NewPartyConflictProvider(owner *parties.WorkbookFacade) ConflictProvider {
	return newGenericConflictProvider(
		"party",
		[]string{PartiesViewSchemaID},
		func(
			ctx context.Context,
			command ConflictCommand,
			request ConflictResolveRequest,
			patch *PatchRequest,
		) (MutationResult, error) {
			var ownerPatch *parties.WorkbookPatchRequest
			if patch != nil {
				converted := partyPatchRequestFromWorkbook(*patch)
				ownerPatch = &converted
			}
			result, err := owner.ResolveConflict(ctx, parties.WorkbookConflictCommand{
				Mechanics:      conflictMechanics(command, request.ClientTxnID),
				Actor:          command.Actor,
				ResolutionKind: request.ResolutionKind,
				Patch:          ownerPatch,
				Now:            command.Now,
			})
			return mutationResultFromPartyWorkbook(result), adaptPartyWorkbookOwnerError(err)
		},
	)
}

func NewTaskDecisionConflictProvider(
	recordType string,
	viewSchemaID string,
	owner taskDecisionConflictOwner,
) ConflictProvider {
	return newGenericConflictProvider(
		recordType,
		[]string{viewSchemaID},
		func(
			ctx context.Context,
			command ConflictCommand,
			request ConflictResolveRequest,
			patch *PatchRequest,
		) (MutationResult, error) {
			var ownerPatch *tasksdecisions.WorkbookPatchRequest
			if patch != nil {
				converted := taskDecisionPatchRequestFromWorkbook(*patch)
				ownerPatch = &converted
			}
			result, err := owner.ResolveConflict(ctx, tasksdecisions.WorkbookConflictCommand{
				Mechanics:      conflictMechanics(command, request.ClientTxnID),
				ResolutionKind: request.ResolutionKind,
				Patch:          ownerPatch,
				Now:            command.Now,
			})
			return mutationResultFromTaskDecisionWorkbook(result), adaptTaskDecisionWorkbookOwnerError(err)
		},
	)
}

type taskDecisionConflictOwner interface {
	ResolveConflict(context.Context, tasksdecisions.WorkbookConflictCommand) (tasksdecisions.WorkbookMutationResult, error)
}

type genericConflictFunc func(
	context.Context,
	ConflictCommand,
	ConflictResolveRequest,
	*PatchRequest,
) (MutationResult, error)

func newGenericConflictProvider(
	recordType string,
	viewSchemaIDs []string,
	resolve genericConflictFunc,
) ConflictProvider {
	allowedViews := make(map[string]struct{}, len(viewSchemaIDs))
	for _, viewSchemaID := range viewSchemaIDs {
		allowedViews[viewSchemaID] = struct{}{}
	}
	return conflictProvider{
		decode: func(
			reader io.Reader,
			token string,
			claims conflicttokens.ConflictTokenClaims,
		) (ConflictAdmission, *httpapi.APIError) {
			if claims.RouteKey != workbookConflictResolveRouteKey {
				return ConflictAdmission{}, invalidMutationPayload("conflict_token", "invalid_value")
			}
			if _, ok := allowedViews[claims.ViewSchemaID]; !ok {
				return ConflictAdmission{}, invalidMutationPayload("conflict_token", "invalid_value")
			}
			request, apiErr := DecodeConflictResolveRequest(reader, token, claims)
			return ConflictAdmission{
				ClientTxnID: request.ClientTxnID,
				normalized:  request,
			}, apiErr
		},
		resolve: func(ctx context.Context, command ConflictCommand) (MutationResult, error) {
			request, ok := command.Admission.normalized.(ConflictResolveRequest)
			if !ok || command.AuthoritativeRecordType != recordType {
				return MutationResult{}, pgx.ErrNoRows
			}
			command.RequestHash = requestHash(
				command.RequestHash,
				ConflictResolveRequestHash(command.Claims, request),
			)
			var patch *PatchRequest
			if request.ResolutionKind != "keep_saved" {
				if request.ResolvedChange == nil {
					return MutationResult{}, mutationValidationError("resolved_value", "missing_required_field")
				}
				patch = &PatchRequest{
					ViewSchemaID:   command.Claims.ViewSchemaID,
					BaseRowVersion: command.Claims.CurrentRowVersion,
					ClientTxnID:    request.ClientTxnID,
					Changes:        []PatchChange{*request.ResolvedChange},
				}
			}
			result, err := resolve(ctx, command, request, patch)
			if errors.Is(err, conflicttokens.ErrClientTxnConflict) {
				return MutationResult{}, authn.ErrClientTxnConflict
			}
			return result, err
		},
	}
}

func conflictMechanics(
	command ConflictCommand,
	clientTxnID string,
) conflicttokens.Command {
	return conflicttokens.Command{
		ActorUserID: command.Actor.ID,
		RecordID:    command.RecordID,
		Claims:      command.Claims,
		ClientTxnID: clientTxnID,
		RequestHash: command.RequestHash,
		RequestID:   command.RequestID,
		RouteKey:    command.Claims.RouteKey,
	}
}

func entityConflictPatchRequest(request PatchRequest) (hostidentity.PatchRequest, error) {
	changes := make([]hostidentity.PatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		converted := hostidentity.PatchChange{FieldKey: change.FieldKey}
		if change.Collection != nil {
			converted.CollectionActions = make(
				[]hostidentity.CollectionAction,
				0,
				len(change.Collection.Actions),
			)
			for _, action := range change.Collection.Actions {
				converted.CollectionActions = append(
					converted.CollectionActions,
					hostidentity.CollectionAction{
						Op:             action.Op,
						RawText:        action.RawText,
						NormalizedText: action.NormalizedText,
						ItemRef:        action.ItemRef,
					},
				)
			}
		} else if change.Value != nil {
			switch change.Value.Kind {
			case "text":
				converted.Value = change.Value.Text
			case "null":
				converted.Value = nil
			default:
				return hostidentity.PatchRequest{}, mutationValidationError(
					change.FieldKey,
					"invalid_value",
				)
			}
		}
		changes = append(changes, converted)
	}
	return hostidentity.PatchRequest{
		ViewSchemaID:   request.ViewSchemaID,
		BaseRowVersion: request.BaseRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        changes,
	}, nil
}

func mutationResultFromTimeline(result timeline.MutationResult, viewSchemaID string) MutationResult {
	return MutationResult{
		Payload:          result.Payload,
		StatusCode:       result.StatusCode,
		Replayed:         result.Replayed,
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		RowVersion:       result.RowVersion,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: append([]string(nil), result.ChangedFieldKeys...),
	}
}

func mutationResultFromEntityCreate(
	result hostidentity.MutationResult,
	viewSchemaID string,
	incidentID uuid.UUID,
	clientTxnID string,
) MutationResult {
	return mutationResultFromSimpleCreate(
		result.Payload,
		result.StatusCode,
		result.Replayed,
		result.RecordID,
		result.ChangeSetID,
		result.RowVersion,
		incidentID,
		clientTxnID,
		viewSchemaID,
	)
}

func mutationResultFromSimpleCreate(
	payload map[string]any,
	statusCode int,
	replayed bool,
	recordID uuid.UUID,
	changeSetID uuid.UUID,
	rowVersion int64,
	incidentID uuid.UUID,
	clientTxnID string,
	viewSchemaID string,
) MutationResult {
	return MutationResult{
		Payload:      payload,
		StatusCode:   statusCode,
		Replayed:     replayed,
		IncidentID:   incidentID,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		ClientTxnID:  clientTxnID,
		RowVersion:   rowVersion,
		ViewSchemaID: viewSchemaID,
	}
}

func adaptTimelineMutationError(
	err error,
	clientTxnID string,
	illegalTransitionReason string,
	noEffectiveChangeField string,
) error {
	if err == nil {
		return nil
	}
	if apiErr, ok := timelineadmission.ClassifyMutationAPIError(err, timelineadmission.MutationAPIErrorContext{
		ClientTxnID:                 clientTxnID,
		IllegalTransitionReasonCode: illegalTransitionReason,
		NoEffectiveChangeField:      noEffectiveChangeField,
	}); ok {
		return &publicMutationError{apiError: apiErr}
	}
	var mentionTransition *mentions.MentionTransitionError
	if errors.As(err, &mentionTransition) {
		return &LifecycleValidationError{
			FromStatus:     mentionTransition.FromStatus,
			ToStatus:       mentionTransition.ToStatus,
			ViolatedGuards: append([]string(nil), mentionTransition.ViolatedGuards...),
		}
	}
	var entityConflict *hostidentity.ExactMatchConflictError
	if errors.As(err, &entityConflict) {
		return &publicMutationError{apiError: entityMatchConflictError(
			entityConflict.EntityType,
			entityConflict.IdentifierClass,
			entityConflict.CandidateRecords,
		)}
	}
	if isTimelineMentionMutationError(err) {
		return mutationValidationError("action_payload", "invalid_value")
	}
	return err
}

func requestHash(provided []byte, derived []byte) []byte {
	if len(provided) > 0 {
		return provided
	}
	return derived
}

func artifactViewSchemaIDs() []string {
	return []string{
		NotesViewSchemaID,
		CommLogViewSchemaID,
		HandoffViewSchemaID,
		StatusReviewViewSchemaID,
		LessonViewSchemaID,
		FindingsViewSchemaID,
		InvestigativeQueriesViewSchemaID,
		ForensicKeywordsViewSchemaID,
	}
}
