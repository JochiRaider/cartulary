package merge

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func identifierMutationsToMergeMutations(records []mergeIdentifierMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_preserved_identifier",
			TargetID:      identifierMutationTargetID(record.Before, record.After),
			OperationKind: "create",
			BeforeValue:   record.Before,
			AfterValue:    record.After,
		})
	}
	return result
}

func aliasMutationsToMergeMutations(records []mergeAliasMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(records))
	for _, record := range records {
		result = append(result, mergeMutation{
			TargetKind:    "entity_alias",
			TargetID:      aliasMutationTargetID(record.Before, record.After),
			OperationKind: "create",
			BeforeValue:   record.Before,
			AfterValue:    record.After,
		})
	}
	return result
}

func identifierMutationTargetID(before map[string]any, after map[string]any) string {
	value := after
	if value == nil {
		value = before
	}
	if value == nil {
		return ""
	}
	return strings.Join([]string{
		"entity_preserved_identifier",
		targetIDComponent(stringMapValue(value, "record_id")),
		targetIDComponent(stringMapValue(value, "entity_type")),
		targetIDComponent(stringMapValue(value, "identifier_type")),
		targetIDComponent(stringMapValue(value, "normalized_value")),
		targetIDComponent(stringMapValue(value, "classification")),
	}, ":")
}

func aliasMutationTargetID(before map[string]any, after map[string]any) string {
	value := after
	if value == nil {
		value = before
	}
	if value == nil {
		return ""
	}
	if aliasID := stringMapValue(value, "entity_alias_id"); aliasID != "" {
		return "entity_alias:" + aliasID
	}
	return strings.Join([]string{
		"entity_alias",
		targetIDComponent(stringMapValue(value, "record_id")),
		targetIDComponent(stringMapValue(value, "entity_type")),
		targetIDComponent(stringMapValue(value, "normalized_text")),
	}, ":")
}

func stringMapValue(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func targetIDComponent(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
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
