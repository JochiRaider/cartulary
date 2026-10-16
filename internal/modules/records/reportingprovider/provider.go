package reportingprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func CollectFactsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string) (exportprovider.ProviderOutput, error) {
	output, err := exportprovider.CollectQueryProviderOutputTx(ctx, tx, incidentID, "records", supportRefs, []exportprovider.FieldQuery{{
		Prefix: "record_envelopes",
		SQL: `SELECT r.record_id::text, 'record_envelope'::text, r.record_type::text, 'derived_analytic'::text, jsonb_build_object(
       'record_id', r.record_id,
       'record_type', r.record_type,
       'created_by_user_id', r.created_by_user_id,
       'created_at', r.created_at,
       'updated_by_user_id', r.updated_by_user_id,
       'updated_at', r.updated_at,
       'row_version', r.row_version
   )
  FROM records r
 WHERE r.incident_id = $1 AND r.deleted_at IS NULL`,
	}})
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	rows, err := tx.Query(ctx, `SELECT record_id::text, record_type::text FROM records WHERE incident_id = $1 AND deleted_at IS NOT NULL ORDER BY record_id`, incidentID)
	if err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var fact exportprovider.DeletedRecordFact
		if err := rows.Scan(&fact.RecordID, &fact.RecordType); err != nil {
			return exportprovider.ProviderOutput{}, err
		}
		output.SourceFacts.DeletedRecords = append(output.SourceFacts.DeletedRecords, fact)
	}
	if err := rows.Err(); err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	if err := output.SourceFacts.Validate(); err != nil {
		return exportprovider.ProviderOutput{}, err
	}
	return output, nil
}
