package records

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func ExportIncidentBundleFiles(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]incidentportability.File, error) {
	file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, "data/records.ndjson", `SELECT to_jsonb(t) FROM records t WHERE incident_id = $1 ORDER BY record_id`)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{file}, nil
}

func ImportIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files map[string][]byte, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	return incidentportability.ImportFixedBundleFileNDJSON(ctx, tx, incidentportability.FixedImportSpec{
		LogicalBundlePath: "data/records.ndjson",
		AttributionTable:  "records",
		StableIdentity:    []string{"record_id"},
		RequiredColumns:   []string{"record_id", "incident_id", "record_type", "row_version"},
		InsertSQL:         `INSERT INTO records SELECT * FROM jsonb_populate_record(NULL::records, $1::jsonb)`,
	}, files, actorUserID, attributions)
}
