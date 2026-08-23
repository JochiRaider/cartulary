// Package deleterestore implements the Artifacts-owned Revisions source adapter.
package deleterestore

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type Source struct {
	catalog *sourcecatalog.Catalog
}

var _ deleterestorecontract.DeleteRestoreSource = Source{}

func NewSource(catalog *sourcecatalog.Catalog) Source {
	return Source{catalog: catalog}
}

func (Source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object(
           'record', to_jsonb(r),
           'source', to_jsonb(a)
             || COALESCE(to_jsonb(f) - ARRAY['record_id', 'incident_id', 'updated_at']::text[], '{}'::jsonb)
             || COALESCE(to_jsonb(iq) - ARRAY['record_id', 'incident_id', 'created_at', 'created_by_user_id']::text[], '{}'::jsonb)
             || COALESCE(to_jsonb(fk) - ARRAY['record_id', 'incident_id', 'created_at']::text[], '{}'::jsonb)
       )
  FROM records r
  JOIN artifacts a
    ON a.record_id = r.record_id
  LEFT JOIN artifact_findings f
    ON f.incident_id = a.incident_id
   AND f.record_id = a.record_id
  LEFT JOIN artifact_investigative_queries iq
    ON iq.incident_id = a.incident_id
   AND iq.record_id = a.record_id
  LEFT JOIN artifact_forensic_keywords fk
    ON fk.incident_id = a.incident_id
   AND fk.record_id = a.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (Source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (s Source) ViewSchemaID(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var artifactType string
	if err := tx.QueryRow(ctx, `SELECT artifact_type FROM artifacts WHERE record_id = $1`, recordID).Scan(&artifactType); err != nil {
		return "", err
	}
	surface, ok := s.catalog.SurfaceByArtifactType(artifactType)
	if !ok {
		return "", fmt.Errorf("unsupported artifact type %q", artifactType)
	}
	return surface.ViewSchemaID, nil
}

func (Source) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}
