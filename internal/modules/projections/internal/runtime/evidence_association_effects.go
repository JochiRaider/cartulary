package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
)

type EvidenceAssociationEffects struct {
	store *Store
}

func NewEvidenceAssociationEffectsFromStore(store *Store) *EvidenceAssociationEffects {
	return &EvidenceAssociationEffects{store: store}
}

func (effects *EvidenceAssociationEffects) RefreshEvidenceAssociationEffects(
	ctx context.Context,
	tx pgx.Tx,
	input evidenceprojection.EvidenceAssociationEffectsInput,
) (evidenceprojection.EvidenceAssociationEffectsResult, error) {
	result := evidenceprojection.EvidenceAssociationEffectsResult{
		Changes: make([]evidenceprojection.EvidenceSupportRowChange, 0, len(input.Subjects)),
	}
	if effects == nil || effects.store == nil {
		return result, errors.New("projection store is required")
	}
	if tx == nil {
		return result, errors.New("projection transaction is required")
	}
	if input.IncidentID == uuid.Nil {
		return result, errors.New("evidence association incident_id is required")
	}
	if err := validateEvidenceAssociationSubjects(input.Subjects); err != nil {
		return result, err
	}
	registry, err := effects.store.providerRegistry()
	if err != nil {
		return result, err
	}
	for _, subject := range input.Subjects {
		providers := registry.evidenceAssociationProviders(subject.RecordType)
		if len(providers) == 0 {
			continue
		}
		change := evidenceprojection.EvidenceSupportRowChange{
			RecordID:      subject.RecordID,
			AffectedViews: make([]evidenceprojection.EvidenceAffectedViewChange, 0, len(providers)),
		}
		for _, provider := range providers {
			viewSchemaID := provider.descriptor.ViewSchemaIDs[0]
			if _, err := provider.loadEvidenceAssociationStateTx(ctx, effects.store, tx, subject.RecordID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return result, fmt.Errorf("load support projection %s before refresh: %w", viewSchemaID, err)
			}
			if err := effects.store.refreshProjectionRowTx(ctx, tx, viewSchemaID, subject.RecordID); err != nil {
				return result, fmt.Errorf("refresh support projection %s: %w", viewSchemaID, err)
			}
			after, err := provider.loadEvidenceAssociationStateTx(ctx, effects.store, tx, subject.RecordID)
			if err != nil {
				return result, fmt.Errorf("load support projection %s after refresh: %w", viewSchemaID, err)
			}
			rowVersion, err := projectionRowVersion(after["row_version"])
			if err != nil {
				return result, fmt.Errorf("load support projection %s row version: %w", viewSchemaID, err)
			}
			if change.RowVersion == 0 {
				change.RowVersion = rowVersion
			} else if change.RowVersion != rowVersion {
				return result, fmt.Errorf("support projection row versions disagree for record %s", subject.RecordID)
			}
			fieldKeys := append([]string(nil), provider.evidenceAssociationEffectFields...)
			slices.Sort(fieldKeys)
			fieldKeys = slices.Compact(fieldKeys)
			change.AffectedViews = append(change.AffectedViews, evidenceprojection.EvidenceAffectedViewChange{
				ViewSchemaID:     viewSchemaID,
				ChangeKind:       evidenceprojection.SupportChangeInvalidate,
				ChangedFieldKeys: fieldKeys,
			})
		}
		result.Changes = append(result.Changes, change)
	}
	return result, nil
}

func loadHostEvidenceAssociationStateTx(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if store == nil || store.physical == nil {
		return nil, errors.New("projection storage is required")
	}
	rowVersion, evidenceCount, err := store.physical.LoadHostEvidenceAssociationStateTx(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	return evidenceCountAssociationState(recordID, rowVersion, "host.evidence_count", evidenceCount), nil
}

func loadIdentityEvidenceAssociationStateTx(ctx context.Context, store *Store, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if store == nil || store.physical == nil {
		return nil, errors.New("projection storage is required")
	}
	rowVersion, evidenceCount, err := store.physical.LoadIdentityEvidenceAssociationStateTx(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	return evidenceCountAssociationState(recordID, rowVersion, "identity.evidence_count", evidenceCount), nil
}

func evidenceCountAssociationState(recordID uuid.UUID, rowVersion int64, effectFieldKey string, evidenceCount int64) map[string]any {
	return map[string]any{
		"record_id":   recordID.String(),
		"row_version": rowVersion,
		"cells": map[string]any{
			effectFieldKey: evidenceCount,
		},
	}
}

func validateEvidenceAssociationSubjects(subjects []evidenceprojection.EvidenceAssociationSubject) error {
	for index, subject := range subjects {
		if subject.RecordID == uuid.Nil || strings.TrimSpace(subject.RecordType) == "" {
			return fmt.Errorf("evidence association subject %d is incomplete", index)
		}
		if index == 0 {
			continue
		}
		previous := subjects[index-1]
		comparison := strings.Compare(previous.RecordID.String(), subject.RecordID.String())
		if comparison > 0 || (comparison == 0 && strings.Compare(previous.RecordType, subject.RecordType) >= 0) {
			return errors.New("evidence association subjects must be unique and ordered by record_id then record_type")
		}
	}
	return nil
}

func projectionRowVersion(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case int:
		return int64(typed), nil
	default:
		return 0, fmt.Errorf("row_version was %T", value)
	}
}

var _ evidenceprojection.AssociationEffects = (*EvidenceAssociationEffects)(nil)
