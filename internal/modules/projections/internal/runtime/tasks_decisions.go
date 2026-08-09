package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

type TaskRequestSource = taskdecisionprojection.TaskRequestSourceReader
type DecisionSource = taskdecisionprojection.DecisionSourceReader

func (s *Store) refreshTaskRequestTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, source TaskRequestSource) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("task-request projection source is required")
	}
	input, found, err := source.LoadTaskRequestProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if err := s.physical.DeleteTaskRequestRowTx(ctx, tx, recordID); err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.physical.InsertTaskRequestTx(ctx, tx, input)
}

func (s *Store) refreshDecisionTxCore(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, source DecisionSource) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("decision projection source is required")
	}
	input, found, err := source.LoadDecisionProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if err := s.physical.DeleteDecisionRowTx(ctx, tx, recordID); err != nil {
		return err
	}
	if !found {
		return nil
	}
	return s.physical.InsertDecisionTx(ctx, tx, input)
}

func (s *Store) RebuildIncidentTaskRequestsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, taskRequestsViewSchemaID, incidentID)
}

func (s *Store) RebuildTaskRequests(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildTaskDecisionProvider(ctx, incidentID, taskRequestsViewSchemaID, "task-request")
}

func (s *Store) rebuildIncidentTaskRequestsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source TaskRequestSource) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("task-request projection source is required")
	}
	if err := s.physical.DeleteTaskRequestIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListTaskRequestProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.InsertTaskRequestTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}

func (s *Store) RebuildIncidentDecisionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, decisionsViewSchemaID, incidentID)
}

func (s *Store) RebuildDecisions(ctx context.Context, incidentID uuid.UUID) error {
	return s.rebuildTaskDecisionProvider(ctx, incidentID, decisionsViewSchemaID, "decision")
}

func (s *Store) rebuildTaskDecisionProvider(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	providerName string,
) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin %s projection rebuild: %w", providerName, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.rebuildProjectionIncidentTx(ctx, tx, viewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s projection rebuild: %w", providerName, err)
	}
	return nil
}

func (s *Store) rebuildIncidentDecisionsTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, source DecisionSource) error {
	if s == nil || s.physical == nil {
		return errors.New("projection storage is required")
	}
	if source == nil {
		return errors.New("decision projection source is required")
	}
	if err := s.physical.DeleteDecisionIncidentTx(ctx, tx, incidentID); err != nil {
		return err
	}
	var afterRecordID *uuid.UUID
	for {
		page, err := source.ListDecisionProjectionInputsTx(ctx, tx, incidentID, afterRecordID, 500)
		if err != nil {
			return err
		}
		for _, input := range page.Inputs {
			if err := s.physical.InsertDecisionTx(ctx, tx, input); err != nil {
				return err
			}
		}
		if page.NextRecordID == nil {
			return nil
		}
		afterRecordID = page.NextRecordID
	}
}
