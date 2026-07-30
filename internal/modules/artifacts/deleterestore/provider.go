package deleterestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Source struct{}

var _ deleterestorecontract.DeleteRestoreSource = Source{}

func NewSource() Source {
	return Source{}
}

func (Source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(a))
  FROM records r
  JOIN artifacts a
    ON a.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (Source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (Source) ViewSchemaID(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var artifactType string
	if err := tx.QueryRow(ctx, `SELECT artifact_type FROM artifacts WHERE record_id = $1`, recordID).Scan(&artifactType); err != nil {
		return "", err
	}
	variant, ok := viewschema.LookupArtifactVariantByArtifactType(artifactType)
	if !ok {
		switch artifactType {
		case "investigative_query":
			return "cartulary.view.investigative_queries.v1", nil
		case "forensic_keyword":
			return "cartulary.view.forensic_keywords.v1", nil
		default:
			return "", fmt.Errorf("unsupported artifact type %q", artifactType)
		}
	}
	return variant.PublicSurfaceRef, nil
}

func (Source) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
}
