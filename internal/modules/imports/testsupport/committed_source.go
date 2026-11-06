// Package testsupport supplies Imports-owned seeds for focused collaborator
// storage tests. End-to-end import tests must use the public import lifecycle.
package testsupport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const SourceBytes = "flow\nfixture\n"

// RetainUnitSource retains a small, actual source for an already seeded unit.
// It never creates users, incidents, workflows, target rows or services.
func RetainUnitSource(ctx context.Context, db postgres.DB, unitID uuid.UUID) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	sum := sha256.Sum256([]byte(SourceBytes))
	digest := hex.EncodeToString(sum[:])
	if _, err = tx.Exec(ctx, `INSERT INTO import_committed_source_blobs (content_sha256,byte_size)
VALUES ($1,$2) ON CONFLICT DO NOTHING`, digest, len(SourceBytes)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO import_committed_source_chunks (content_sha256,chunk_ordinal,source_bytes)
VALUES ($1,0,$2) ON CONFLICT DO NOTHING`, digest, []byte(SourceBytes)); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO import_committed_sources (
committed_source_id,incident_id,content_sha256,source_file_kind,source_media_type,original_filename,
parser_profile_id,parser_version,locator_kind,locator,source_rect_a1,header_row_ref,data_start_row_ref,approved_mapping_json,mapping_fingerprint,
target_kind,target_view_schema_id,extension_profile_id,origin_import_session_id,origin_import_unit_id,created_by_user_id,committed_at)
SELECT $1,s.incident_id,$2,s.source_file_kind,s.source_media_type,s.original_filename,s.parser_profile_id,s.parser_version,
u.locator_kind,u.locator,u.source_rect_a1,u.header_row_ref,u.data_start_row_ref,u.approved_mapping_json,u.mapping_fingerprint,'network_flow_table','',
'network_flow_activity',s.import_session_id,u.import_unit_id,s.created_by_user_id,s.created_at
FROM import_units u JOIN import_sessions s USING (import_session_id)
WHERE u.import_unit_id=$3 AND s.source_content_sha256=$2`, imports.CommittedSourceID(unitID), digest, unitID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("fixture source and seeded unit disagree")
	}
	return tx.Commit(ctx)
}
