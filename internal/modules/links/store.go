package links

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/vocabulary"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
)

type Store struct {
	projections ProjectionRows
}

// ProjectionRows publishes relationship-derived cells in the mutation's
// transaction. Source owners retain responsibility for their record revisions.
type ProjectionRows interface {
	RefreshLinkedRecordsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) error
}

// NewMutationStore composes ordinary relationship writes with their projection
// effects. NewStore is the storage capability used by coordinators that publish
// their own complete projection scope (import, rollback and merge).
func NewMutationStore(rows ProjectionRows) (*Store, error) {
	if rows == nil {
		return nil, errors.New("links: projection rows are required")
	}
	return &Store{projections: rows}, nil
}

var errRecordLinkNotFound = errors.New("links: record link not found")
var errFieldReferenceNotFound = errors.New("links: field reference not found")
var errTagNotFound = errors.New("links: tag not found")
var errInvalidTag = errors.New("links: invalid tag")
var errInvalidRecordLink = errors.New("links: invalid record link")

const (
	LinkTypeInvalid             = vocabulary.LinkTypeInvalid
	LinkTypeObservedOnHost      = vocabulary.LinkTypeObservedOnHost
	LinkTypeObservedAsIdentity  = vocabulary.LinkTypeObservedAsIdentity
	LinkTypeReferencesIndicator = vocabulary.LinkTypeReferencesIndicator
	LinkTypeAttachedEvidence    = vocabulary.LinkTypeAttachedEvidence
	LinkTypeReferencesArtifact  = vocabulary.LinkTypeReferencesArtifact
	LinkTypeDerivedFrom         = vocabulary.LinkTypeDerivedFrom
	LinkTypeMergedInto          = vocabulary.LinkTypeMergedInto
	LinkTypeSupportedBy         = vocabulary.LinkTypeSupportedBy
	LinkTypeReferencesRecord    = vocabulary.LinkTypeReferencesRecord
	LinkTypeSupersedes          = vocabulary.LinkTypeSupersedes

	LinkProvenanceInvalid   = vocabulary.LinkProvenanceInvalid
	LinkProvenanceManual    = vocabulary.LinkProvenanceManual
	LinkProvenanceAutoMatch = vocabulary.LinkProvenanceAutoMatch
	LinkProvenanceImport    = vocabulary.LinkProvenanceImport
	LinkProvenanceRollback  = vocabulary.LinkProvenanceRollback
	LinkProvenanceSystem    = vocabulary.LinkProvenanceSystem
)

func NewStore() *Store {
	return &Store{}
}

func (s *Store) HasActiveIncomingSupersedesLinkForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (bool, error) {
	var linkID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, recordID).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query active incoming supersedes link: %w", err)
	}
	return true, nil
}

func validateRecordLinkCommand(linkType LinkType, provenance LinkProvenance, confidence *int, srcRecordID uuid.UUID, dstRecordID uuid.UUID) error {
	if srcRecordID == uuid.Nil || dstRecordID == uuid.Nil || srcRecordID == dstRecordID {
		return fmt.Errorf("%w: invalid endpoints", errInvalidRecordLink)
	}
	if linkType.String() == "" {
		return fmt.Errorf("%w: unsupported link type", errInvalidRecordLink)
	}
	switch provenance {
	case LinkProvenanceManual:
		if confidence != nil {
			return fmt.Errorf("%w: manual links must not carry confidence", errInvalidRecordLink)
		}
	case LinkProvenanceAutoMatch:
		if linkType != LinkTypeObservedOnHost && linkType != LinkTypeObservedAsIdentity {
			return fmt.Errorf("%w: auto_match provenance is only valid for entity observation links", errInvalidRecordLink)
		}
		if confidence == nil || *confidence != 100 {
			return fmt.Errorf("%w: auto_match confidence must be exactly 100", errInvalidRecordLink)
		}
	default:
		return fmt.Errorf("%w: unsupported provenance %q", errInvalidRecordLink, provenance.String())
	}
	return nil
}

func validateActiveLinkEndpointsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID) error {
	envelopes, err := records.NewStore().LoadEnvelopesTx(ctx, tx, []uuid.UUID{srcRecordID, dstRecordID}, false)
	if err != nil {
		return fmt.Errorf("validate link endpoints: %w", err)
	}
	if len(envelopes) != 2 || !activeIncidentEnvelope(envelopes[srcRecordID], incidentID) || !activeIncidentEnvelope(envelopes[dstRecordID], incidentID) {
		return fmt.Errorf("%w: endpoints must be active same-incident records", errInvalidRecordLink)
	}
	return nil
}

func activeIncidentEnvelope(envelope recordenvelope.Envelope, incidentID uuid.UUID) bool {
	return envelope.RecordID != uuid.Nil && envelope.IncidentID == incidentID && envelope.DeletedAt == nil
}

func validateSupersedesEndpointsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, replacementRecordID uuid.UUID, supersededRecordID uuid.UUID) error {
	envelopes, err := records.NewStore().LoadEnvelopesTx(ctx, tx, []uuid.UUID{replacementRecordID, supersededRecordID}, false)
	if err != nil {
		return fmt.Errorf("validate supersedes endpoints: %w", err)
	}
	src, dst := envelopes[replacementRecordID], envelopes[supersededRecordID]
	if !activeIncidentEnvelope(src, incidentID) || !activeIncidentEnvelope(dst, incidentID) {
		return fmt.Errorf("%w: supersedes endpoints must be active same-incident records", errInvalidRecordLink)
	}
	if src.RecordType != dst.RecordType || (src.RecordType != "timeline_event" && src.RecordType != "decision") {
		return fmt.Errorf("%w: supersedes endpoints must both be timeline events or both be decisions", errInvalidRecordLink)
	}
	return nil
}

func intPointersEqual(left *int, right *int) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}
