package reference_data

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

type dependencyRevision struct {
	key      string
	revision int64
}

// Admission follows only exact, usable retained tuples. PostgreSQL deduplicates
// recursive tuples, including cycles, and may spill its working set; Go retains
// at most the sixteen supported profile keys. Missing supported keys receive a
// revision guard too. Unknown profile keys are permanently unavailable under
// the captured application/configuration contract and need no mutable guard.
func admissionDependencyRevisions(ctx context.Context, tx pgx.Tx, a operationAdmission, selected []string) ([]dependencyRevision, error) {
	if a.Kind != "import" && a.Kind != "reverify" && a.Kind != "refresh" {
		return nil, nil
	}
	dependencies := []packformat.Dependency{}
	if a.Import != nil {
		dependencies = append(dependencies, a.Import.Dependencies...)
	}
	encoded, err := json.Marshal(dependencies)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `WITH RECURSIVE seeds(pack_key,pack_version,payload_sha256) AS (
 SELECT pack_key,pack_version,payload_sha256 FROM jsonb_to_recordset($1::jsonb) AS d(pack_key text,pack_version text,payload_sha256 text)
 UNION
 SELECT d->>'pack_key',d->>'pack_version',d->>'payload_sha256'
 FROM reference_pack_candidates c JOIN reference_pack_versions v USING(pack_key,pack_version)
 CROSS JOIN LATERAL jsonb_array_elements(convert_from(v.manifest_bytes,'UTF8')::jsonb->'dependencies') d
 WHERE $2 IN ('reverify','refresh') AND c.pack_key=ANY($3) AND ($4='' OR c.pack_version=$4)
 AND c.distribution_kind='operator_imported' AND NOT c.removed AND c.current_envelope_id IS NOT NULL
), graph(pack_key,pack_version,payload_sha256) AS (
 SELECT * FROM seeds
 UNION
 SELECT d->>'pack_key',d->>'pack_version',d->>'payload_sha256'
 FROM graph g JOIN reference_pack_versions v USING(pack_key,pack_version,payload_sha256)
 JOIN reference_pack_candidates c USING(pack_key,pack_version)
 CROSS JOIN LATERAL jsonb_array_elements(convert_from(v.manifest_bytes,'UTF8')::jsonb->'dependencies') d
 WHERE c.health='verified_available' AND NOT c.removed AND c.current_envelope_id IS NOT NULL
)
SELECT DISTINCT g.pack_key COLLATE "C",coalesce(k.revision,1) FROM graph g
LEFT JOIN reference_pack_key_state k USING(pack_key) WHERE g.pack_key=ANY($5) ORDER BY 1`, encoded, a.Kind, selected, a.Version, packformat.SupportedDependencyKeys())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []dependencyRevision{}
	for rows.Next() {
		var r dependencyRevision
		if err := rows.Scan(&r.key, &r.revision); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func admissionGuardKeys(selected []string, dependencies []dependencyRevision) []string {
	keys := slices.Clone(selected)
	for _, d := range dependencies {
		keys = append(keys, d.key)
	}
	slices.Sort(keys)
	return slices.Compact(keys)
}
