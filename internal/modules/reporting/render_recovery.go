package reporting

import "github.com/jackc/pgx/v5"

// Callers choose the owner's fixed query. This shared decoder retains every
// stored file for the canonical bundle validator, including undeclared members.
func loadRetainedRenderFiles(rows pgx.Rows, result *RenderedRelease) error {
	defer rows.Close()
	declared := map[string]RenderBundleManifestFile{}
	for _, f := range result.RenderBundle.Manifest.Files {
		declared[f.Path] = f
	}
	for rows.Next() {
		var f RenderBundleFile
		if err := rows.Scan(&f.Path, &f.Role, &f.MediaType, &f.SHA256, &f.SizeBytes, &f.StorageKind, &f.Bytes); err != nil {
			return err
		}
		f.RequiredForRelease = declared[f.Path].RequiredForRelease
		result.RenderBundle.Files = append(result.RenderBundle.Files, f)
		if f.Role == renderBundleRoleTokenManifest {
			result.Redaction.TokenManifestJSON = f.Bytes
		}
	}
	if result.RenderBundle.Manifest.TokenManifestSHA256 != nil {
		result.Redaction.TokenManifestSHA256 = *result.RenderBundle.Manifest.TokenManifestSHA256
	}
	return rows.Err()
}
