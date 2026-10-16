package performancefixture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// EvidenceInspectorApplication constructs the Timeline portion of an isolated
// Evidence inspector fixture. The selected row and every attachment use normal
// create admission, revisions, Links mutations and both owners' projections.
// The caller owns disposal of the whole fixture on any construction failure.
type EvidenceInspectorApplication struct {
	owner     *timeline.PerformanceFixtureContribution
	mutations *timeline.Facade
	actor     authn.UserRecord
}

type EvidenceInspectorTarget struct {
	RecordID   uuid.UUID
	RowVersion int64
	Summary    string
}

func NewEvidenceInspectorApplication(owner *timeline.PerformanceFixtureContribution, mutations *timeline.Facade, actor authn.UserRecord) (*EvidenceInspectorApplication, error) {
	if owner == nil || mutations == nil || actor.ID == uuid.Nil || !actor.IsActive {
		return nil, errors.New("evidence inspector fixture requires Timeline owners and an active actor")
	}
	return &EvidenceInspectorApplication{owner: owner, mutations: mutations, actor: actor}, nil
}

func (a *EvidenceInspectorApplication) Create(ctx context.Context, incidentID uuid.UUID, rowCount int, evidenceIDs []uuid.UUID) (EvidenceInspectorTarget, error) {
	if incidentID == uuid.Nil || rowCount < 2 || rowCount > 100000 || len(evidenceIDs) < 1 || len(evidenceIDs) > 100 {
		return EvidenceInspectorTarget{}, errors.New("evidence inspector fixture population is invalid")
	}
	seen := make(map[uuid.UUID]bool, len(evidenceIDs))
	actions := make([]map[string]string, len(evidenceIDs))
	for i, id := range evidenceIDs {
		if id == uuid.Nil || seen[id] {
			return EvidenceInspectorTarget{}, errors.New("evidence inspector fixture requires distinct Evidence identities")
		}
		seen[id] = true
		actions[i] = map[string]string{"op": "add_record_ref", "linked_record_id": id.String()}
	}
	const summary = "Evidence inspector performance target"
	encoded, err := json.Marshal(map[string]any{
		"client_txn_id":                   uuid.NewString(),
		"timeline.activity_synopsis_text": summary,
		"timeline.attached_evidence_ids":  map[string]any{"kind": "collection_actions_v1", "actions": actions[:min(len(actions), mutationpolicy.MaxCollectionActions)]},
	})
	if err != nil {
		return EvidenceInspectorTarget{}, err
	}
	request, failure := admission.DecodeTimelineCreateRequest(bytes.NewReader(encoded))
	if failure != nil {
		return EvidenceInspectorTarget{}, errors.New("evidence inspector fixture create admission failed")
	}
	rows := make([]timeline.PerformanceFixtureRow, rowCount-1)
	for i := range rows {
		rows[i].Summary = fmt.Sprintf("Evidence fixture Timeline %05d", i+1)
	}
	now := time.Now().UTC()
	createdRows, err := a.owner.CreatePerformanceFixtureRows(ctx, timeline.PerformanceFixtureCommand{Actor: a.actor, IncidentID: incidentID, Rows: rows, Now: now})
	if err != nil {
		return EvidenceInspectorTarget{}, err
	}
	if createdRows.RowCount != len(rows) || createdRows.RelationshipRows != 0 {
		return EvidenceInspectorTarget{}, errors.New("evidence inspector fixture Timeline population diverged")
	}
	target, err := a.mutations.CreateRow(ctx, timeline.CreateRowCommand{
		Actor: a.actor, IncidentID: incidentID, Request: request,
		RequestHash: admission.CreateRequestHash(request), RequestID: uuid.NewString(), Now: now,
	})
	if err != nil {
		return EvidenceInspectorTarget{}, fmt.Errorf("create Evidence inspector target and attachments: %w", err)
	}
	if target.StatusCode != http.StatusCreated || target.Replayed || target.RecordID == uuid.Nil || target.ChangeSetID == uuid.Nil || target.RowVersion != 1 {
		return EvidenceInspectorTarget{}, errors.New("evidence inspector fixture target mutation did not create a fresh record")
	}
	// Keep ordinary collection-action bounds. Additional attachments are real
	// versioned patches, not a privileged oversized fixture mutation.
	for offset := mutationpolicy.MaxCollectionActions; offset < len(actions); offset += mutationpolicy.MaxCollectionActions {
		encoded, err := json.Marshal(map[string]any{
			"view_schema_id": timeline.TimelineViewSchemaID,
			"client_txn_id":  uuid.NewString(), "base_row_version": target.RowVersion,
			"changes": []map[string]any{{"field_key": "timeline.attached_evidence_ids", "action_payload": map[string]any{
				"kind": "collection_actions_v1", "actions": actions[offset:min(len(actions), offset+mutationpolicy.MaxCollectionActions)],
			}}},
		})
		if err != nil {
			return EvidenceInspectorTarget{}, err
		}
		request, failure := admission.DecodeTimelinePatchRequest(bytes.NewReader(encoded))
		if failure != nil {
			return EvidenceInspectorTarget{}, errors.New("evidence inspector attachment patch admission failed")
		}
		patched, err := a.mutations.PatchRow(ctx, timeline.PatchRowCommand{
			Actor: a.actor, RecordID: target.RecordID, Request: request,
			RequestHash: admission.PatchRequestHash(request), RequestID: uuid.NewString(), Now: now,
		})
		if err != nil {
			return EvidenceInspectorTarget{}, fmt.Errorf("attach remaining Evidence inspector records: %w", err)
		}
		if patched.StatusCode != http.StatusOK || patched.Replayed || patched.RecordID != target.RecordID || patched.RowVersion != target.RowVersion+1 || patched.ChangeSetID == uuid.Nil || patched.ChangeSetID == target.ChangeSetID {
			return EvidenceInspectorTarget{}, errors.New("evidence inspector attachment patch did not advance the target")
		}
		target = patched
	}
	if err := a.owner.ValidatePerformanceFixtureRows(ctx, incidentID, timeline.PerformanceFixtureResult{RowCount: rowCount}); err != nil {
		return EvidenceInspectorTarget{}, err
	}
	return EvidenceInspectorTarget{RecordID: target.RecordID, RowVersion: target.RowVersion, Summary: summary}, nil
}
