package deleterestore

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Provider struct {
	recordsdeleterestore.TableProvider
}

func NewProvider() Provider {
	return Provider{TableProvider: recordsdeleterestore.TableProvider{
		SourceTable:     "artifacts",
		SourceRecordCol: "record_id",
	}}
}

func (p Provider) ViewSchemaID(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
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
