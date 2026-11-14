package merge

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func identifierMutationsToMergeMutations(records []mergeIdentifierMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_preserved_identifier",
			TargetID:      record.Value.TargetID(),
			OperationKind: "create",
			AfterValue:    record.Value.Map(),
		})
	}
	return result
}

func aliasMutationsToMergeMutations(records []mergeAliasMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_alias",
			TargetID:      record.Value.TargetID(),
			OperationKind: "create",
			AfterValue:    record.Value.Map(),
		})
	}
	return result
}

func scanMergeAliasRecord(scanner interface{ Scan(dest ...any) error }) (mergeAliasRecord, error) {
	var (
		record    mergeAliasRecord
		deletedAt pgtype.Timestamptz
	)
	if err := scanner.Scan(
		&record.EntityAliasID,
		&record.IncidentID,
		&record.RecordID,
		&record.EntityType,
		&record.RawText,
		&record.NormalizedText,
		&record.Classification,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return mergeAliasRecord{}, fmt.Errorf("scan merge alias: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}

func scanMergePreservedIdentifierRecord(scanner interface{ Scan(dest ...any) error }) (mergePreservedIdentifierRecord, error) {
	var (
		record    mergePreservedIdentifierRecord
		deletedAt pgtype.Timestamptz
	)
	if err := scanner.Scan(
		&record.EntityPreservedIdentifierID,
		&record.IncidentID,
		&record.RecordID,
		&record.EntityType,
		&record.IdentifierType,
		&record.RawValue,
		&record.NormalizedValue,
		&record.Classification,
		&record.CreatedAt,
		&deletedAt,
	); err != nil {
		return mergePreservedIdentifierRecord{}, fmt.Errorf("scan merge preserved identifier: %w", err)
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.CreatedAt = record.CreatedAt.UTC()
	return record, nil
}
