package incidentbundle

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

type portableTaskRequest struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	Title               string
	Status              string
	PortableOwnerUserID *uuid.UUID
	Priority            string
	TaskKind            string
	Workstream          *string
	DueAt               *time.Time
	RequesterPartyText  *string
	RequesterPartyID    *uuid.UUID
	BlockedReason       *string
	CompletedAt         *time.Time
	ExternalTicketRef   *string
	ClosureSummary      *string
	DecisionRecordID    *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type portableDecision struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	Summary             string
	Status              string
	PortableOwnerUserID uuid.UUID
	DecisionType        string
	DecidedAt           time.Time
	Rationale           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func exactPortableMembers(row map[string]any, required []string) bool {
	if len(row) != len(required) {
		return false
	}
	for _, member := range required {
		if _, ok := row[member]; !ok {
			return false
		}
	}
	return true
}

func canonicalPortableUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func nullableCanonicalPortableUUID(value any) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableUUID(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func admittedPortableActor(value any, importContext sourceport.ImportContext) (uuid.UUID, bool) {
	actorID, ok := canonicalPortableUUID(value)
	if !ok {
		return uuid.Nil, false
	}
	_, admitted := importContext.Actors.Lookup(actorID.String())
	return actorID, admitted
}

func nullableAdmittedPortableActor(value any, importContext sourceport.ImportContext) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	actorID, ok := admittedPortableActor(value, importContext)
	if !ok {
		return nil, false
	}
	return &actorID, true
}

func portableString(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok
}

func nullablePortableString(value any) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := portableString(value)
	if !ok {
		return nil, false
	}
	return &text, true
}

func canonicalPortableTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, false
	}
	_, offset := parsed.Zone()
	canonicalText := text
	if strings.HasSuffix(canonicalText, "+00:00") {
		canonicalText = strings.TrimSuffix(canonicalText, "+00:00") + "Z"
	}
	if offset != 0 || parsed.UTC().Format(time.RFC3339Nano) != canonicalText {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func nullableCanonicalPortableTime(value any) (*time.Time, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableTime(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}
