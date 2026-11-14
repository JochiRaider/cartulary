package deleterestore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/linkfacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/referencefacts"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

type TaskRequestSource struct{ referenceTargets referencefacts.Capability }

var _ deleterestorecontract.DeleteRestoreSource = TaskRequestSource{}

func NewTaskRequestSource(targets referencefacts.Capability) TaskRequestSource {
	return TaskRequestSource{referenceTargets: targets}
}

func (TaskRequestSource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(t))
  FROM records r
  JOIN task_requests t
    ON t.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (TaskRequestSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (TaskRequestSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.task_requests.v1", nil
}

func (TaskRequestSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

type DecisionSource struct {
	linkFacts    linkfacts.Capability
	supersession linkfacts.SupersessionReader
}

var _ deleterestorecontract.DeleteRestoreSource = DecisionSource{}

func NewDecisionSource(facts linkfacts.Capability, supersession linkfacts.SupersessionReader) DecisionSource {
	return DecisionSource{linkFacts: facts, supersession: supersession}
}

func (DecisionSource) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(d))
  FROM records r
  JOIN decisions d
    ON d.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (DecisionSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (DecisionSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.decisions.v1", nil
}

func (DecisionSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (p TaskRequestSource) ValidateStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if request.Kind != deleterestorecontract.StateTransitionRestore {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	var party, decision *uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT requester_party_id,decision_record_id FROM task_requests WHERE incident_id=$1 AND record_id=$2`, request.IncidentID, request.RecordID).Scan(&party, &decision); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	type reference struct {
		id         uuid.UUID
		recordType string
	}
	refs := make([]reference, 0, 2)
	if party != nil {
		refs = append(refs, reference{*party, "party"})
	}
	if decision != nil {
		refs = append(refs, reference{*decision, "decision"})
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].id.String() < refs[j].id.String() })
	for _, ref := range refs {
		valid, err := tasksource.LockReferenceTargetTx(ctx, tx, p.referenceTargets, request.IncidentID, ref.id, ref.recordType, true)
		if err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "55P03" {
				return deleterestorecontract.StateTransitionPreparation{}, &revisions.RecordLockedError{RecordID: request.RecordID}
			}
			return deleterestorecontract.StateTransitionPreparation{}, err
		}
		if !valid {
			return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "inactive_task_reference"}}, nil
		}
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (p DecisionSource) ValidateStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if request.Kind == deleterestorecontract.StateTransitionRestore {
		return p.validateDecisionRestoreTx(ctx, tx, request)
	}
	if request.Kind != deleterestorecontract.StateTransitionDelete {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	// Revisions holds the Decision envelope lock here. Task reference writes
	// retain a shared target lock, so this observes references committed during
	// any wait before delete admission. Completed Tasks remain live references.
	var referenced bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
        SELECT 1 FROM task_requests t
        JOIN records r ON r.incident_id=t.incident_id AND r.record_id=t.record_id
        WHERE t.incident_id=$1 AND t.decision_record_id=$2 AND r.deleted_at IS NULL
    )`, request.IncidentID, request.RecordID).Scan(&referenced); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	if referenced {
		return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "active_incoming_task_decision_reference"}}, nil
	}
	return p.validateDecisionPeersAfterDeleteTx(ctx, tx, request)
}
