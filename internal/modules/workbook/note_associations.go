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

func (s *service) handleNoteAssociations(w http.ResponseWriter, r *http.Request) {
	noteID, ok := pathUUID(w, r, "note_record_id")
	if !ok {
		return
	}
	mutating := r.Method == http.MethodPost
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: mutating})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), noteID)
	if isRecordTargetNotFound(err) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if mutating {
		_, apiErr = s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin")
	} else {
		_, apiErr = s.requireIncidentMembership(r.Context(), target.IncidentID, principal.User.ID)
	}
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if mutating {
		outcome, err := s.noteAssociations.Apply(r.Context(), NoteAssociationCommand{Actor: principal.User, Target: target, RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now()}, r.Body)
		result, outcomeErr := resolveMutationOutcome(outcome, err)
		writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
		return
	}
	values := r.URL.Query()
	for key, entries := range values {
		if len(entries) != 1 || (key != "kind" && key != "limit" && key != "cursor_token") {
			writeAPIError(w, r, invalidViewQuery(key, "invalid_value"))
			return
		}
	}
	kind := values.Get("kind")
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(values, "workbook.note-associations", principal.User.ID.String(), map[string]string{"incident": target.IncidentID.String(), "note": noteID.String(), "kind": kind, "order": "created_at_desc_link_id_desc"})
	if reason != "" {
		writeAPIError(w, r, invalidViewQuery("", reason))
		return
	}
	query := NoteAssociationQuery{ActorID: principal.User.ID, Target: target, Kind: kind, Limit: binding.Limit}
	if cursor != nil {
		stamp, timeErr := time.Parse(time.RFC3339Nano, cursor.Position["created_at"])
		id, idErr := uuid.Parse(cursor.Position["link_id"])
		if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 2 || timeErr != nil || idErr != nil {
			writeAPIError(w, r, invalidViewQuery("", pagination.ReasonInvalidCursorToken))
			return
		}
		query.BeforeTime = &stamp
		query.BeforeID = id
	}
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccessWithMeta(w, r, http.StatusOK, map[string]any{"note_record_id": page.NoteID.String(), "row_version": page.RowVersion, "kind": kind, "items": page.Items, "next_cursor_token": next}, httpapi.EnvelopeMeta{RequestID: httpapi.RequestIDFromContext(r.Context()), Paging: &httpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next}})
}
