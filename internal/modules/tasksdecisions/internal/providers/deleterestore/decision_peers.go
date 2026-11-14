package deleterestore

import (
	"context"
	"errors"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p DecisionSource) validateDecisionPeersAfterDeleteTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if p.linkFacts == nil {
		return deleterestorecontract.StateTransitionPreparation{}, errors.New("tasks/decisions: Links facts are required")
	}
	facts, err := p.linkFacts.LoadRecordLinkFactsTx(ctx, tx, request.IncidentID, request.RecordID)
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	peers := map[uuid.UUID]struct{}{}
	for _, fact := range facts {
		if fact.LinkType != "supersedes" {
			continue
		}
		if fact.SourceRecordID == request.RecordID && fact.DestinationRecordID != request.RecordID {
			peers[fact.DestinationRecordID] = struct{}{}
		}
		if fact.DestinationRecordID == request.RecordID && fact.SourceRecordID != request.RecordID {
			peers[fact.SourceRecordID] = struct{}{}
		}
	}
	ordered := make([]uuid.UUID, 0, len(peers))
	for id := range peers {
		ordered = append(ordered, id)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].String() < ordered[j].String() })
	blocked := deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "decision_supersession_dependency"}}
	for _, id := range ordered {
		// The coordinator already protects all directly related envelopes.
		// Lock typed owner state in the same order before computing its survivor.
		state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, p.linkFacts, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return blocked, nil
		}
		if err != nil {
			return deleterestorecontract.StateTransitionPreparation{}, err
		}
		if state.IncidentID != request.IncidentID {
			return blocked, nil
		}
		for _, fact := range facts {
			if fact.LinkType != "supersedes" {
				continue
			}
			if fact.SourceRecordID == request.RecordID && fact.DestinationRecordID == id {
				state.IncomingSupersedes--
			}
			if fact.DestinationRecordID == request.RecordID && fact.SourceRecordID == id {
				state.OutgoingSupersedes--
			}
		}
		if state.IncomingSupersedes < 0 || state.OutgoingSupersedes < 0 || policy.ValidateDecisionMachineState(state) != nil {
			return blocked, nil
		}
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}
