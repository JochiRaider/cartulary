package exportprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// FieldQuery is an owner-authored query returning source ID, source family,
// record type, content class and JSON value, in that order. The prefix defines
// only the selector namespace; it cannot supply source identity or type.
type FieldQuery struct {
	Prefix                       string
	SQL                          string
	DisclosurePartitionRefPrefix string
}

func CollectQueryProviderOutputTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, providerKey string, supportRefs map[string][]string, queries []FieldQuery) (ProviderOutput, error) {
	fields := []BoundField{}
	for _, query := range queries {
		rows, err := tx.Query(ctx, query.SQL, incidentID)
		if err != nil {
			return ProviderOutput{}, err
		}
		for rows.Next() {
			var id, family, recordType, contentClass string
			var raw []byte
			if err := rows.Scan(&id, &family, &recordType, &contentClass, &raw); err != nil {
				rows.Close()
				return ProviderOutput{}, err
			}
			// Keep exact integer tokens. A float64 decode can round source values
			// before Reporting's canonical-byte validation can detect the loss.
			value, err := canonicaljson.DecodeStrict(raw)
			if err != nil {
				rows.Close()
				return ProviderOutput{}, err
			}
			field := BoundField{
				RecordID: id, RecordType: recordType,
				FieldFact: FieldFact{
					SchemaID: FieldFactSchemaID, Path: fmt.Sprintf("/%s/%s", query.Prefix, id),
					ContentClass: contentClass, SourceFamily: family, Value: value,
					SupportRefs: CloneStrings(supportRefs[id]),
				},
			}
			if query.DisclosurePartitionRefPrefix != "" {
				field.DisclosurePartitionRefs = []string{query.DisclosurePartitionRefPrefix + id}
			}
			fields = append(fields, field)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return ProviderOutput{}, err
		}
	}
	return NewProviderOutput(providerKey, fields, nil)
}
