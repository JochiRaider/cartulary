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
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteAssociationCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	NoteID      uuid.UUID
	Admission   NoteAssociationAdmission
	RequestID   string
	Now         time.Time
}

func (o *NoteAssociations) Apply(ctx context.Context, command NoteAssociationCommand) (MutationResult, error) {
	f, a := o.mutations, command.Admission
	if !a.valid() {
		return MutationResult{}, &ValidationError{Field: "payload", ReasonCode: "invalid_value"}
	}
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := o.access.CheckTx(ctx, tx, command.IncidentID, command.ActorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorReviewerAdmin, Lifecycle: admission.LifecycleAny}); err != nil {
		return MutationResult{}, err
	}
	key := IdempotencyKey{OperationID: OperationNoteAssociations, ActorUserID: command.ActorUserID, ScopeKey: command.NoteID.String(), ClientTxnID: a.clientTxnID}
	// Serialize identical keys before reading the retained receipt; a concurrent
	// exact replay must not turn into a stale-version failure or a second write.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, string(key.OperationID)+":"+key.ActorUserID.String()+":"+key.ScopeKey+":"+key.ClientTxnID); err != nil {
		return MutationResult{}, err
	}
	existing, exists, lookupErr := f.idempotency.GetTx(ctx, tx, key, a.requestHash())
	stored, found, err := validateStoredMutation(existing, exists, lookupErr, "note associations", storedMutationExpectation{kind: StoredMutationNoteAssociations, viewSchemaID: NotesViewSchemaID, recordID: &command.NoteID})
	if err != nil {
		return MutationResult{}, err
	}
	if found {
		return MutationResult{Row: stored.Row, Outcome: MutationOutcomeReplayed, IncidentID: stored.IncidentID, RecordID: stored.RecordID, RowVersion: stored.RowVersion, ChangeSetID: stored.ChangeSetID, ViewSchemaID: stored.ViewSchemaID, ClientTxnID: a.clientTxnID}, nil
	}
	if _, err := o.access.CheckTx(ctx, tx, command.IncidentID, command.ActorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorReviewerAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return MutationResult{}, err
	}
	note, err := o.noteTx(ctx, tx, command.IncidentID, command.NoteID)
	if err != nil {
		return MutationResult{}, err
	}
	linkType, _ := associationSemantics(a.kind)
	actions := append([]noteAssociationAction(nil), a.actions...)
	ids := []uuid.UUID{command.NoteID}
	for i := range actions {
		action := &actions[i]
		if action.Op == "remove" {
			fact, kind, err := o.links.LoadAssociationTx(ctx, tx, command.IncidentID, action.LinkID)
			if errors.Is(err, pgx.ErrNoRows) {
				return MutationResult{}, invalidAssociationAction()
			}
			if err != nil {
				return MutationResult{}, err
			}
			if kind != linkType || (a.kind == NoteAssociationSource && fact.TargetID != note.RecordID) || (a.kind != NoteAssociationSource && fact.SourceID != note.RecordID) {
				return MutationResult{}, invalidAssociationAction()
			}
			action.CounterpartID = fact.TargetID
			if a.kind == NoteAssociationSource {
				action.CounterpartID = fact.SourceID
			}
		}
		if _, err := o.validateCounterpartTx(ctx, tx, note, a.kind, action.CounterpartID); err != nil {
			return MutationResult{}, err
		}
		ids = append(ids, action.CounterpartID)
	}
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	ids = slices.Compact(ids)
	envelopes := make(map[uuid.UUID]records.Envelope, len(ids))
	for _, id := range ids {
		envelope, err := f.recordEnvelopes.LoadEnvelopeTx(ctx, tx, id, true)
		if err != nil {
			return MutationResult{}, err
		}
		envelopes[id] = envelope
	}
	note, err = o.noteTx(ctx, tx, command.IncidentID, command.NoteID)
	if err != nil {
		return MutationResult{}, err
	}
	if note.RowVersion != a.baseVersion {
		return MutationResult{}, &RowVersionConflictError{RecordID: note.RecordID, BaseRowVersion: a.baseVersion, CurrentRowVersion: note.RowVersion}
	}
	beforeRows := map[uuid.UUID]map[string]any{}
	beforeSnapshots := map[uuid.UUID]revisions.RecordSnapshot{}
	for _, id := range ids {
		envelope := envelopes[id]
		if id != note.RecordID {
			if _, err := o.validateCounterpartTx(ctx, tx, note, a.kind, id); err != nil {
				return MutationResult{}, err
			}
		}
		if envelope.RecordType != "artifact" && envelope.RecordType != "evidence" {
			continue
		}
		row, err := o.loadAffectedRowTx(ctx, tx, envelope)
		if err != nil {
			return MutationResult{}, err
		}
		beforeRows[id] = row
		snapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, id)
		if err != nil {
			return MutationResult{}, err
		}
		beforeSnapshots[id] = snapshot
	}
	mutations := []links.Mutation{}
	changed := map[uuid.UUID]bool{}
	for _, action := range actions {
		src, dst := note.RecordID, action.CounterpartID
		if a.kind == NoteAssociationSource {
			src, dst = dst, src
		}
		var result links.RecordLinkCommandResult
		if action.Op == "add" {
			result, err = o.links.EnsureAssociationTx(ctx, tx, links.UpsertLinkCommand{IncidentID: note.IncidentID, SrcRecordID: src, DstRecordID: dst, LinkType: linkType, Provenance: links.LinkProvenanceManual, OwnerUserID: command.ActorUserID, Now: command.Now})
		} else {
			fact, kind, loadErr := o.links.LoadAssociationTx(ctx, tx, note.IncidentID, action.LinkID)
			if errors.Is(loadErr, pgx.ErrNoRows) {
				return MutationResult{}, invalidAssociationAction()
			}
			if loadErr != nil {
				return MutationResult{}, loadErr
			}
			if fact.SourceID != src || fact.TargetID != dst || kind != linkType {
				return MutationResult{}, invalidAssociationAction()
			}
			var removed bool
			result, removed, err = o.links.TombstoneActiveLinkCommandTx(ctx, tx, links.TombstoneActiveLinkCommand{IncidentID: note.IncidentID, SrcRecordID: src, DstRecordID: dst, LinkType: linkType, ActorUserID: command.ActorUserID, Now: command.Now})
			if err == nil && !removed {
				return MutationResult{}, invalidAssociationAction()
			}
		}
		if err != nil {
			return MutationResult{}, err
		}
		if mutation, ok := result.Mutation(); ok {
			mutations = append(mutations, mutation)
			changed[note.RecordID] = true
			changed[action.CounterpartID] = true
		}
	}
	var changeSetID *uuid.UUID
	afterNote := beforeRows[note.RecordID]
	if len(mutations) > 0 {
		id, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: note.IncidentID, ActorUserID: command.ActorUserID, Source: string(OperationNoteAssociations), ClientTxnID: &a.clientTxnID, RequestID: &command.RequestID, CreatedAt: command.Now.UTC()})
		if err != nil {
			return MutationResult{}, err
		}
		changeSetID = &id
		sequence := 1
		for _, recordID := range ids {
			if !changed[recordID] || beforeRows[recordID] == nil {
				continue
			}
			envelope := envelopes[recordID]
			version, err := f.recordEnvelopes.AdvanceVersionTx(ctx, tx, recordID, command.ActorUserID, command.Now)
			if err != nil {
				return MutationResult{}, err
			}
			if envelope.RecordType == "artifact" {
				err = f.source.rows.touchRowTx(ctx, tx, recordID, command.Now)
				if err == nil {
					err = f.source.projections.RefreshArtifactTx(ctx, tx, recordID)
				}
			} else {
				err = o.rows.RefreshEvidenceTx(ctx, tx, recordID)
			}
			if err != nil {
				return MutationResult{}, err
			}
			after, err := o.loadAffectedRowTx(ctx, tx, envelope)
			if err != nil {
				return MutationResult{}, err
			}
			afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
			if err != nil {
				return MutationResult{}, err
			}
			beforeSnapshot := beforeSnapshots[recordID]
			beforeVersion, afterVersion := workbookVersionID(recordID, envelope.RowVersion), workbookVersionID(recordID, version)
			if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{ChangeSetID: id, SequenceNo: sequence, TargetKind: "record", RecordID: recordID, OperationKind: "patch", BeforeVersionID: &beforeVersion, AfterVersionID: &afterVersion, BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot}); err != nil {
				return MutationResult{}, err
			}
			sequence++
			fields := changedFieldKeys(beforeRows[recordID], after)
			if err := f.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{ChangeSetID: id, RecordID: recordID, RowVersion: version, BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot, ConflictFacts: artifactRevisionFacts(beforeRows[recordID], after, fields)}); err != nil {
				return MutationResult{}, err
			}
			if err := f.appendRecordChangedTx(ctx, tx, note.IncidentID, command.ActorUserID, a.clientTxnID, id, recordID, version, sequence-2, command.Now, associationView(envelope.RecordType), after, fields); err != nil {
				return MutationResult{}, err
			}
			if recordID == note.RecordID {
				afterNote = after
				note.RowVersion = version
			}
		}
		if _, err := f.appendCollectionMutationsTx(ctx, tx, id, sequence, mutations); err != nil {
			return MutationResult{}, err
		}
	}
	result := MutationResult{Row: afterNote, Outcome: MutationOutcomeUpdated, IncidentID: note.IncidentID, RecordID: note.RecordID, RowVersion: note.RowVersion, ChangeSetID: changeSetID, ClientTxnID: a.clientTxnID, ViewSchemaID: NotesViewSchemaID}
	receipt := NewStoredNoteAssociationResult(StoredMutationPayload{Row: afterNote, IncidentID: note.IncidentID, RecordID: note.RecordID, RowVersion: note.RowVersion, ChangeSetID: changeSetID, ViewSchemaID: NotesViewSchemaID})
	if err := f.idempotency.PutTx(ctx, tx, key, a.requestHash(), receipt); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit Note associations: %w", err)
	}
	return result, nil
}

func invalidAssociationAction() error {
	return &ValidationError{Field: "actions", ReasonCode: "invalid_value"}
}
func (o *NoteAssociations) loadAffectedRowTx(ctx context.Context, tx pgx.Tx, record records.Envelope) (map[string]any, error) {
	if record.RecordType == "evidence" {
		return o.rows.LoadEvidenceTx(ctx, tx, record.RecordID)
	}
	return o.mutations.source.projections.LoadArtifactTx(ctx, tx, NotesViewSchemaID, record.RecordID)
}
