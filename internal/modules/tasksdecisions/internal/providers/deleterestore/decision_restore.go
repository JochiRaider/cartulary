package deleterestore

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"sort"
)

func (p DecisionSource) validateDecisionRestoreTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if p.supersession == nil {
		return deleterestorecontract.StateTransitionPreparation{}, errors.New("tasks/decisions: retained supersession reader is required")
	}
	facts, err := p.supersession.LoadRetainedSupersessionTx(ctx, tx, request.IncidentID, request.RecordID)
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	blocked := deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "decision_supersession_dependency"}}
	ids := map[uuid.UUID]struct{}{request.RecordID: {}}
	for _, fact := range facts {
		if fact.SourceRecordID == fact.DestinationRecordID || (fact.SourceRecordID != request.RecordID && fact.DestinationRecordID != request.RecordID) {
			return blocked, nil
		}
		ids[fact.SourceRecordID] = struct{}{}
		ids[fact.DestinationRecordID] = struct{}{}
	}
	ordered := make([]uuid.UUID, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].String() < ordered[j].String() })
	states := make(map[uuid.UUID]policy.DecisionMachineState, len(ids))
	deleted := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ordered {
		state, isDeleted, err := tasksource.LoadDecisionRestoreStateTx(ctx, tx, p.linkFacts, id, request.IncidentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return blocked, nil
		}
		if err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "55P03" {
				return deleterestorecontract.StateTransitionPreparation{}, &revisions.RecordLockedError{RecordID: request.RecordID}
			}
			return deleterestorecontract.StateTransitionPreparation{}, err
		}
		states[id] = state
		deleted[id] = isDeleted && id != request.RecordID
	}
	for _, fact := range facts {
		if deleted[fact.SourceRecordID] || deleted[fact.DestinationRecordID] {
			continue
		}
		source, target := states[fact.SourceRecordID], states[fact.DestinationRecordID]
		source.OutgoingSupersedes++
		target.IncomingSupersedes++
		states[fact.SourceRecordID], states[fact.DestinationRecordID] = source, target
	}
	for _, id := range ordered {
		if !deleted[id] && policy.ValidateDecisionMachineState(states[id]) != nil {
			return blocked, nil
		}
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}
