package workbook

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type NoteAssociationQuery struct {
	ActorID    uuid.UUID
	Target     RecordTarget
	Kind       string
	Limit      int
	BeforeTime *time.Time
	BeforeID   uuid.UUID
}

type NoteAssociationPage struct {
	NoteID     uuid.UUID
	RowVersion int64
	Items      []NoteAssociationItem
	HasMore    bool
	LastTime   time.Time
	LastID     uuid.UUID
}

type NoteAssociationItem struct {
	ItemRef       string    `json:"item_ref"`
	CounterpartID uuid.UUID `json:"counterpart_record_id"`
	ViewSchemaID  string    `json:"view_schema_id"`
	DisplayLabel  string    `json:"display_label"`
	Direction     string    `json:"direction"`
}

type NoteAssociationCommand struct {
	Actor     authn.UserRecord
	Target    RecordTarget
	RequestID string
	Now       time.Time
}

// Notes contributes its own admission and semantics; Workbook owns transport.
type NoteAssociationProvider interface {
	List(context.Context, NoteAssociationQuery) (NoteAssociationPage, *MutationFailure, error)
	Apply(context.Context, NoteAssociationCommand, io.Reader) (MutationOutcome, error)
}

// admittedNoteTarget resolves the admitted note_record_id to its record target
// and applies the incident's visibility and the route's role gate
// (REQ-01-704), before any query or body is read.
func (s *service) admittedNoteTarget(r *http.Request, roles admission.RoleSet, requiredRole string) (httpauth.Principal, RecordTarget, *httpapi.APIError) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		return httpauth.Principal{}, RecordTarget{}, apiErr
	}
	noteID, apiErr := admittedPathUUID(r, "note_record_id")
	if apiErr != nil {
		return httpauth.Principal{}, RecordTarget{}, apiErr
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), noteID)
	if isRecordTargetNotFound(err) {
		return httpauth.Principal{}, RecordTarget{}, httpapi.NotFound("incident_not_found")
	}
	if err != nil {
		return httpauth.Principal{}, RecordTarget{}, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, roles, requiredRole); apiErr != nil {
		return httpauth.Principal{}, RecordTarget{}, apiErr
	}
	return principal, target, nil
}

func (s *service) handleMutateNoteAssociations(w http.ResponseWriter, r *http.Request) {
	principal, target, apiErr := s.admittedNoteTarget(r, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := s.noteAssociations.Apply(r.Context(), NoteAssociationCommand{Actor: principal.User, Target: target, RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now()}, r.Body)
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}

func (s *service) handleListNoteAssociations(w http.ResponseWriter, r *http.Request) {
	principal, target, apiErr := s.admittedNoteTarget(r, admission.RolesMember, "")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeNoteAssociationListRequest(r.URL.RawQuery, principal.User.ID, target, s.cursorCodec)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	query, binding := request.query, request.binding
	page, failure, err := s.noteAssociations.List(r.Context(), query)
	if failure != nil {
		writeAPIError(w, r, mutationFailureAPIError(failure))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var next *string
	if page.HasMore {
		token, err := s.cursorCodec.Encode(pagination.Cursor{Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: map[string]string{"created_at": page.LastTime.UTC().Format(time.RFC3339Nano), "link_id": page.LastID.String()}})
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		next = &token
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccessWithMeta(w, r, http.StatusOK, map[string]any{"note_record_id": page.NoteID.String(), "row_version": page.RowVersion, "kind": query.Kind, "items": page.Items, "next_cursor_token": next}, httpapi.EnvelopeMeta{RequestID: httpapi.RequestIDFromContext(r.Context()), Paging: &httpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next}})
}
