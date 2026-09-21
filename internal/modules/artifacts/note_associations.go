package artifacts

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteAssociationAccess interface {
	CheckTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

type NoteAssociationLinks interface {
	ListAssociationsTx(context.Context, pgx.Tx, links.AssociationLookup) ([]links.AssociationLink, error)
	LoadAssociationTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (links.AssociationLink, links.LinkType, error)
	EnsureAssociationTx(context.Context, pgx.Tx, links.UpsertLinkCommand) (links.RecordLinkCommandResult, error)
	TombstoneActiveLinkCommandTx(context.Context, pgx.Tx, links.TombstoneActiveLinkCommand) (links.RecordLinkCommandResult, bool, error)
}

// Counterpart rows are contributed by their source/projection owners. Artifacts
// never queries or mutates another owner's source tables.
type NoteAssociationRows interface {
	LabelTx(context.Context, pgx.Tx, records.Envelope) (string, error)
	LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error
}

type NoteAssociations struct {
	mutations *MutationFacade
	access    NoteAssociationAccess
	links     NoteAssociationLinks
	rows      NoteAssociationRows
}

func NewNoteAssociations(mutations *MutationFacade, access NoteAssociationAccess, linkStore NoteAssociationLinks, rows NoteAssociationRows) (*NoteAssociations, error) {
	if mutations == nil || isNilCapability(access) || isNilCapability(linkStore) || isNilCapability(rows) {
		return nil, fmt.Errorf("artifacts: Note association dependencies are required")
	}
	return &NoteAssociations{mutations: mutations, access: access, links: linkStore, rows: rows}, nil
}

type NoteAssociationItem struct {
	ItemRef       string    `json:"item_ref"`
	CounterpartID uuid.UUID `json:"counterpart_record_id"`
	ViewSchemaID  string    `json:"view_schema_id"`
	DisplayLabel  string    `json:"display_label"`
	Direction     string    `json:"direction"`
}

type NoteAssociationRead struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	NoteID      uuid.UUID
	Kind        NoteAssociationKind
	Limit       int
	BeforeTime  *time.Time
	BeforeID    uuid.UUID
}

type NoteAssociationPage struct {
	NoteID     uuid.UUID
	RowVersion int64
	Items      []NoteAssociationItem
	HasMore    bool
	LastTime   time.Time
	LastID     uuid.UUID
}

func associationSemantics(kind NoteAssociationKind) (links.LinkType, []string) {
	switch kind {
	case NoteAssociationSource:
		return links.LinkTypeReferencesArtifact, []string{"timeline_event", "host", "identity", "evidence"}
	case NoteAssociationEvidence:
		return links.LinkTypeAttachedEvidence, []string{"evidence"}
	case NoteAssociationRelatedNote:
		return links.LinkTypeReferencesArtifact, []string{"artifact"}
	default:
		return links.LinkTypeInvalid, nil
	}
}

func associationView(recordType string) string {
	switch recordType {
	case "artifact":
		return NotesViewSchemaID
	case "timeline_event":
		return "cartulary.view.timeline.v2"
	case "host":
		return "cartulary.view.hosts.v1"
	case "identity":
		return "cartulary.view.identities.v1"
	case "evidence":
		return "cartulary.view.evidence.v1"
	default:
		return ""
	}
}

func (o *NoteAssociations) noteTx(ctx context.Context, tx pgx.Tx, incidentID, noteID uuid.UUID) (records.Envelope, error) {
	note, err := o.mutations.recordEnvelopes.LoadEnvelopeTx(ctx, tx, noteID, false)
	if err != nil {
		return records.Envelope{}, err
	}
	if note.IncidentID != incidentID || note.RecordType != "artifact" || note.DeletedAt != nil {
		return records.Envelope{}, pgx.ErrNoRows
	}
	if err := validateArtifactViewRecordTx(ctx, tx, noteID, NotesViewSchemaID); err != nil {
		return records.Envelope{}, err
	}
	return note, nil
}

func (o *NoteAssociations) List(ctx context.Context, request NoteAssociationRead) (NoteAssociationPage, error) {
	linkType, types := associationSemantics(request.Kind)
	if linkType == links.LinkTypeInvalid || request.Limit < 1 || request.Limit > 500 {
		return NoteAssociationPage{}, &ValidationError{Field: "kind", ReasonCode: "invalid_value"}
	}
	tx, err := o.mutations.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return NoteAssociationPage{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := o.access.CheckTx(ctx, tx, request.IncidentID, request.ActorUserID, admission.Requirement{AllowedRoles: admission.RolesMember, Lifecycle: admission.LifecycleAny}); err != nil {
		return NoteAssociationPage{}, err
	}
	note, err := o.noteTx(ctx, tx, request.IncidentID, request.NoteID)
	if err != nil {
		return NoteAssociationPage{}, err
	}
	// Hold the Note version stable while obtaining the corresponding page.
	if _, err := o.mutations.recordEnvelopes.LoadEnvelopeForShareTx(ctx, tx, note.RecordID); err != nil {
		return NoteAssociationPage{}, err
	}
	note, err = o.noteTx(ctx, tx, request.IncidentID, request.NoteID)
	if err != nil {
		return NoteAssociationPage{}, err
	}
	facts, err := o.links.ListAssociationsTx(ctx, tx, links.AssociationLookup{IncidentID: request.IncidentID, RecordID: request.NoteID, LinkType: linkType, Incoming: request.Kind != NoteAssociationEvidence, Outgoing: request.Kind != NoteAssociationSource, Limit: request.Limit + 1, BeforeTime: request.BeforeTime, BeforeID: request.BeforeID})
	if err != nil {
		return NoteAssociationPage{}, err
	}
	page := NoteAssociationPage{NoteID: note.RecordID, RowVersion: note.RowVersion, Items: []NoteAssociationItem{}, HasMore: len(facts) > request.Limit}
	if page.HasMore {
		facts = facts[:request.Limit]
	}
	for _, fact := range facts {
		page.LastTime, page.LastID = fact.CreatedAt, fact.ID
		counterpart, err := o.mutations.recordEnvelopes.LoadEnvelopeTx(ctx, tx, fact.CounterpartID, false)
		if errors.Is(err, records.ErrEnvelopeNotFound) {
			continue
		}
		if err != nil {
			return NoteAssociationPage{}, err
		}
		if counterpart.DeletedAt != nil || counterpart.IncidentID != note.IncidentID || !slices.Contains(types, counterpart.RecordType) {
			continue
		}
		if request.Kind == NoteAssociationRelatedNote {
			if err := validateArtifactViewRecordTx(ctx, tx, counterpart.RecordID, NotesViewSchemaID); errors.Is(err, pgx.ErrNoRows) {
				continue
			} else if err != nil {
				return NoteAssociationPage{}, err
			}
		}
		label, err := o.rows.LabelTx(ctx, tx, counterpart)
		if err != nil {
			return NoteAssociationPage{}, err
		}
		direction := "outgoing"
		if fact.TargetID == note.RecordID {
			direction = "incoming"
		}
		page.Items = append(page.Items, NoteAssociationItem{ItemRef: associationItemRef(fact.ID), CounterpartID: counterpart.RecordID, ViewSchemaID: associationView(counterpart.RecordType), DisplayLabel: label, Direction: direction})
	}
	return page, tx.Commit(ctx)
}

func (o *NoteAssociations) validateCounterpartTx(ctx context.Context, tx pgx.Tx, note records.Envelope, kind NoteAssociationKind, id uuid.UUID) (records.Envelope, error) {
	_, types := associationSemantics(kind)
	other, err := o.mutations.recordEnvelopes.LoadEnvelopeTx(ctx, tx, id, false)
	if err != nil && !errors.Is(err, records.ErrEnvelopeNotFound) {
		return records.Envelope{}, err
	}
	if err != nil || other.IncidentID != note.IncidentID || other.DeletedAt != nil || id == note.RecordID || !slices.Contains(types, other.RecordType) {
		return records.Envelope{}, &ValidationError{Field: "actions", ReasonCode: "invalid_value"}
	}
	if kind == NoteAssociationRelatedNote {
		if err := validateArtifactViewRecordTx(ctx, tx, id, NotesViewSchemaID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return records.Envelope{}, &ValidationError{Field: "actions", ReasonCode: "invalid_value"}
			}
			return records.Envelope{}, err
		}
	}
	return other, nil
}
