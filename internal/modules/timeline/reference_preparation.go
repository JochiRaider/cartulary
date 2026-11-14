package timeline

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errReferencePreparationChanged = errors.New("timeline: reference preparation changed")

type referencePreparation struct{ auto preparedAutoResolution }
type referenceRow struct {
	recordID uuid.UUID
	changes  []PatchChange
}

func (s *store) prepareCreateReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest, auto bool) (preparedAutoResolution, error) {
	rows := []referenceRow{{changes: []PatchChange{
		{FieldKey: "timeline.host_refs", ActionPayload: request.HostRefs},
		{FieldKey: "timeline.identity_refs", ActionPayload: request.IdentityRefs},
		{FieldKey: "timeline.attached_evidence_ids", ActionPayload: request.AttachedEvidence},
	}}}
	plan := &referencePreparation{}
	if err := s.prepareReferencesTx(ctx, tx, incidentID, rows, auto, plan); err != nil {
		return nil, err
	}
	return plan.auto, nil
}

// Discovery is provisional until all source envelopes have been protected and
// their versions still match. The caller rolls back an obsolete attempt in full.
func (s *store) prepareReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, rows []referenceRow, auto bool, plan *referencePreparation) error {
	sourceIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row.recordID != uuid.Nil {
			sourceIDs = append(sourceIDs, row.recordID)
		}
	}
	before, err := s.recordStore.LoadEnvelopesTx(ctx, tx, sourceIDs, false)
	if err != nil {
		return err
	}
	for _, id := range sourceIDs {
		envelope, present := before[id]
		if !present || envelope.RecordType != "timeline_event" || envelope.DeletedAt != nil {
			return ErrRecordNotFound
		}
		if incidentID == uuid.Nil {
			incidentID = envelope.IncidentID
		}
		if envelope.IncidentID != incidentID {
			return ErrRecordNotFound
		}
	}
	initial := plan.auto == nil
	if initial {
		plan.auto = make(preparedAutoResolution)
	}
	ids := append([]uuid.UUID(nil), sourceIDs...)
	for _, row := range rows {
		for _, change := range row.changes {
			if change.ActionPayload == nil {
				continue
			}
			if isTimelineMentionCollection(change.FieldKey) {
				refs := make(map[uuid.UUID]bool)
				for _, action := range change.ActionPayload.Actions {
					switch action.Op {
					case "add_resolved_ref", "resolve_item":
						if action.ResolvedRecord != nil {
							ids = append(ids, *action.ResolvedRecord)
						}
					case "add_token":
						if auto {
							key := autoResolutionKey{change.FieldKey, action.RawText}
							match, present := plan.auto[key]
							if initial && !present {
								match, err = s.initialAutoResolutionMatchTx(ctx, tx, incidentID, change.FieldKey, action.RawText)
								if err != nil {
									return err
								}
								plan.auto[key] = match
							}
							if match != nil {
								ids = append(ids, match.RecordID)
							}
						}
					}
					switch action.Op {
					case "resolve_item", "dismiss_item", "revert_to_unresolved":
						if id, err := mentionIDFromItemRef(action.ItemRef); err == nil {
							refs[id] = true
						}
					}
				}
				if row.recordID != uuid.Nil && len(refs) > 0 {
					facts, err := s.mentionStore.LoadCollectionSourceFactsTx(ctx, tx, row.recordID, change.FieldKey)
					if err != nil {
						return err
					}
					for _, fact := range facts {
						if refs[fact.MentionID] && fact.ResolvedRecordID != nil {
							ids = append(ids, *fact.ResolvedRecordID)
						}
					}
				}
			} else if isTimelineAttachedEvidenceCollection(change.FieldKey) {
				for _, action := range change.ActionPayload.Actions {
					switch action.Op {
					case "add_record_ref":
						if action.LinkedRecordID != nil {
							ids = append(ids, *action.LinkedRecordID)
						}
					case "remove_record_ref":
						if id, err := parseRecordRefItemRef(action.ItemRef); err == nil {
							ids = append(ids, id)
						}
					}
				}
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	protected, err := s.recordStore.LoadEnvelopesForNoKeyUpdateTx(ctx, tx, ids)
	if err != nil {
		return err
	}
	for _, id := range sourceIDs {
		current, present := protected[id]
		if !present || current.RecordType != "timeline_event" || current.IncidentID != incidentID || current.DeletedAt != nil {
			return ErrRecordNotFound
		}
		if current.RowVersion != before[id].RowVersion {
			return errReferencePreparationChanged
		}
	}
	return nil
}
