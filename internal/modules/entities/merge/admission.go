package merge

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
)

func classifyMissingMergeTargetTx(ctx context.Context, tx pgx.Tx, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) error {
	if _, err := loadMergeTargetMetaTx(ctx, tx, survivorRecordID); err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return ErrMergeTargetNotFound
		}
		return err
	}
	if _, err := loadMergeTargetMetaTx(ctx, tx, loserRecordID); err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return ErrMergeTargetNotFound
		}
		return err
	}
	return &MergePreconditionError{ReasonCode: "target_not_found"}
}

func (s *Store) planMergeProtectedRecordIDsTx(ctx context.Context, tx pgx.Tx, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, now time.Time) ([]uuid.UUID, error) {
	recordIDs := []uuid.UUID{survivorRecordID, loserRecordID}
	survivorMeta, err := loadMergeTargetMetaTx(ctx, tx, survivorRecordID)
	if err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return nil, ErrMergeTargetNotFound
		}
		return nil, err
	}
	loserMeta, err := loadMergeTargetMetaTx(ctx, tx, loserRecordID)
	if err != nil {
		if errors.Is(err, ErrMergeTargetNotFound) {
			return nil, ErrMergeTargetNotFound
		}
		return nil, err
	}
	if loserMeta.IncidentID != survivorMeta.IncidentID {
		return nil, ErrMergeTargetNotFound
	}
	if survivorRecordID == loserRecordID ||
		(survivorMeta.RecordType != "host" && survivorMeta.RecordType != "identity") ||
		loserMeta.RecordType != survivorMeta.RecordType {
		return recordIDs, nil
	}
	assessmentRecordIDs, err := s.ports.assessments.LoadProtectedRecordIDsTx(ctx, tx, AssessmentProtectedSetCommand{
		IncidentID:       survivorMeta.IncidentID,
		RecordType:       survivorMeta.RecordType,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    loserRecordID,
		Now:              now,
	})
	if err != nil {
		return nil, err
	}
	return append(recordIDs, assessmentRecordIDs...), nil
}

func uuidSet(recordIDs []uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(recordIDs))
	for _, recordID := range recordIDs {
		result[recordID] = struct{}{}
	}
	return result
}

func sameUUIDSet(left []uuid.UUID, right []uuid.UUID) bool {
	leftSet := uuidSet(left)
	rightSet := uuidSet(right)
	if len(leftSet) != len(rightSet) {
		return false
	}
	for recordID := range leftSet {
		if _, ok := rightSet[recordID]; !ok {
			return false
		}
	}
	return true
}

func mergeIdentifierConflict(entityType string, conflict *hostidentity.ActiveIdentifierTransitionConflict) error {
	return &MergePreconditionError{
		ReasonCode: "carry_forward_identifier_collision",
		Details: map[string]any{
			"record_type":        entityType,
			"identifier_class":   conflict.IdentifierClass,
			"normalized_value":   conflict.NormalizedValue,
			"blocking_record_id": conflict.BlockingRecordID.String(),
		},
	}
}

func sortedUUIDSet(recordIDs map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(recordIDs))
	for recordID := range recordIDs {
		result = append(result, recordID)
	}
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	return result
}

func loadMergeTargetMetaTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (mergeTargetMeta, error) {
	row := tx.QueryRow(ctx, `
SELECT incident_id, record_type
  FROM records
 WHERE record_id = $1
`, recordID)
	var meta mergeTargetMeta
	meta.RecordID = recordID
	if err := row.Scan(&meta.IncidentID, &meta.RecordType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mergeTargetMeta{}, ErrMergeTargetNotFound
		}
		return mergeTargetMeta{}, fmt.Errorf("load merge target meta: %w", err)
	}
	return meta, nil
}

func (s *Store) loadHostByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (hostidentity.HostRecord, error) {
	record, err := s.hostIdentity.LoadHostTx(ctx, tx, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return hostidentity.HostRecord{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return hostidentity.HostRecord{}, fmt.Errorf("load host merge target: %w", err)
	}
	return record, nil
}

func (s *Store) loadIdentityByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (hostidentity.IdentityRecord, error) {
	record, err := s.hostIdentity.LoadIdentityTx(ctx, tx, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return hostidentity.IdentityRecord{}, ErrMergeTargetNotFound
	}
	if err != nil {
		return hostidentity.IdentityRecord{}, fmt.Errorf("load identity merge target: %w", err)
	}
	return record, nil
}

func validateHostMergePair(survivor hostidentity.HostRecord, loser hostidentity.HostRecord) error {
	if survivor.IncidentID != loser.IncidentID {
		return &MergePreconditionError{ReasonCode: "cross_incident_pair"}
	}
	if survivor.HostState != "stub" && survivor.HostState != "canonical" {
		return &MergePreconditionError{ReasonCode: "survivor_not_active"}
	}
	if loser.HostState != "stub" && loser.HostState != "canonical" {
		return &MergePreconditionError{ReasonCode: "loser_not_active"}
	}
	return nil
}

func validateIdentityMergePair(survivor hostidentity.IdentityRecord, loser hostidentity.IdentityRecord) error {
	if survivor.IncidentID != loser.IncidentID {
		return &MergePreconditionError{ReasonCode: "cross_incident_pair"}
	}
	if survivor.IdentityState != "stub" && survivor.IdentityState != "canonical" {
		return &MergePreconditionError{ReasonCode: "survivor_not_active"}
	}
	if loser.IdentityState != "stub" && loser.IdentityState != "canonical" {
		return &MergePreconditionError{ReasonCode: "loser_not_active"}
	}
	return nil
}
