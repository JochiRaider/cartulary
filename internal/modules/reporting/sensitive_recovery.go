package reporting

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateRetainedSensitiveArtifacts verifies private adjuncts independently of
// shareable export inventories. Recovery must retain them even though incident
// portability deliberately does not. It never reconstructs an absent map.
func ValidateRetainedSensitiveArtifacts(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("reporting: missing sensitive recovery database")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var lastID *uuid.UUID
	lastFamily := ""
	for {
		var family, digest string
		var id uuid.UUID
		var raw, private []byte
		var privateDigest *string
		err := tx.QueryRow(ctx, `
SELECT family,id,bundle_manifest_json,bundle_manifest_sha256,canonical_bytes,reveal_map_sha256 FROM (
 SELECT 'release' AS family,b.release_id AS id,b.bundle_manifest_json,b.bundle_manifest_sha256,s.canonical_bytes,s.reveal_map_sha256
 FROM reporting_render_bundles b LEFT JOIN reporting_release_reveal_maps s USING(release_id)
 UNION ALL
 SELECT 'preview',b.preview_attempt_id,b.bundle_manifest_json,b.bundle_manifest_sha256,s.canonical_bytes,s.reveal_map_sha256
 FROM reporting_composition_preview_outputs b LEFT JOIN reporting_preview_reveal_maps s USING(preview_attempt_id)
) retained WHERE ($2::uuid IS NULL OR (family,id)>($1,$2)) ORDER BY family,id LIMIT 1`, lastFamily, lastID).Scan(&family, &id, &raw, &digest, &private, &privateDigest)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		manifest, _, err := decodePortableObject[RenderBundleManifest](raw, digest)
		if err != nil || validateBundleIdentity(manifest) != nil || manifest.SchemaID != RenderBundleManifestSchemaID || manifest.ManifestVersion != RenderBundleManifestVersion {
			return fmt.Errorf("reporting: invalid retained sensitive artifact parent")
		}
		if family == "release" && (manifest.ReleaseID == nil || *manifest.ReleaseID != id.String()) || family == "preview" && (manifest.PreviewAttemptID == nil || *manifest.PreviewAttemptID != id.String()) {
			return fmt.Errorf("reporting: inconsistent retained sensitive artifact identity")
		}
		evidenceFiles := map[string]RenderBundleManifestFile{}
		for _, f := range manifest.Files {
			if !shareableBundleMember(f) {
				return fmt.Errorf("reporting: private member in retained shareable inventory")
			}
			if f.Role == renderBundleRoleTokenManifest || f.Role == renderBundleRoleRedactionManifest {
				if _, exists := evidenceFiles[f.Role]; exists {
					return fmt.Errorf("reporting: duplicate retained redaction evidence")
				}
				evidenceFiles[f.Role] = f
			}
		}
		query := `SELECT role,bundle_path,inline_bytes,file_sha256,size_bytes FROM reporting_render_bundle_files WHERE release_id=$1 AND role IN ('token_manifest','redaction_manifest')`
		if family == "preview" {
			query = `SELECT role,bundle_path,inline_bytes,file_sha256,size_bytes FROM reporting_composition_preview_output_files WHERE preview_attempt_id=$1 AND role IN ('token_manifest','redaction_manifest')`
		}
		rows, err := tx.Query(ctx, query, id)
		if err != nil {
			return err
		}
		evidence := map[string][]byte{}
		for rows.Next() {
			var role, path, fileSHA string
			var content []byte
			var size int64
			if err := rows.Scan(&role, &path, &content, &fileSHA, &size); err != nil {
				rows.Close()
				return err
			}
			file, declared := evidenceFiles[role]
			_, duplicate := evidence[role]
			if !declared || duplicate || path != file.Path || fileSHA != file.SHA256 || size != file.SizeBytes || size != int64(len(content)) || hashHex(content) != fileSHA {
				rows.Close()
				return fmt.Errorf("reporting: invalid retained redaction evidence bytes")
			}
			evidence[role] = content
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		if len(evidenceFiles) != len(evidence) {
			return fmt.Errorf("reporting: missing retained redaction evidence bytes")
		}
		r := RenderedRelease{
			RenderBundle: RenderBundle{Manifest: manifest}, Redaction: RedactionResult{TokenManifestJSON: evidence[renderBundleRoleTokenManifest], RevealMapJSON: private},
			RedactionManifestJSON: evidence[renderBundleRoleRedactionManifest], RedactionManifestSHA256: manifest.RedactionManifestSHA256,
		}
		if privateDigest != nil {
			r.Redaction.RevealMapSHA256 = *privateDigest
		}
		if manifest.TokenManifestSHA256 != nil {
			r.Redaction.TokenManifestSHA256 = *manifest.TokenManifestSHA256
		}
		if err := validateRevealMap(r); err != nil {
			return err
		}
		lastID = &id
		lastFamily = family
	}
	return tx.Commit(ctx)
}
