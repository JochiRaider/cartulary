package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// CheckReadyState checks committed availability without re-verifying historical
// signatures on every readiness request. Startup validates retained bytes and
// trust; consumer detection commits definitive loss through the coordinator.
// Administrative disablement and expired metadata do not invalidate historical
// pins. Current selections must satisfy the stronger live eligibility rule.
func CheckReadyState(ctx context.Context, db postgres.DB) error {
	if db == nil {
		return errors.New("reference pack: readiness persistence unavailable")
	}
	var ready bool
	err := db.QueryRow(ctx, `SELECT
 EXISTS(SELECT 1 FROM reference_pack_current_set s WHERE s.singleton AND s.pack_set_id IS NOT NULL
 AND 3=(SELECT count(*) FROM reference_pack_set_members m WHERE m.pack_set_id=s.pack_set_id
 AND m.pack_key IN ('type_registry.host','type_registry.evidence','type_registry.indicator')))
 AND NOT EXISTS(SELECT 1 FROM reference_pack_candidates c
 JOIN reference_pack_versions v USING(pack_key,pack_version)
 LEFT JOIN reference_pack_index_generations i ON i.index_id=c.current_index_id
 WHERE (EXISTS(SELECT 1 FROM reference_pack_set_members m WHERE m.pack_key=c.pack_key AND m.pack_version=c.pack_version
 AND (EXISTS(SELECT 1 FROM reference_pack_current_set s WHERE s.pack_set_id=m.pack_set_id)
 OR EXISTS(SELECT 1 FROM reference_pack_pins p WHERE p.pack_set_id=m.pack_set_id)))
 OR EXISTS(SELECT 1 FROM reference_pack_version_pins p WHERE p.pack_key=c.pack_key AND p.pack_version=c.pack_version))
 AND (c.removed OR c.current_envelope_id IS NULL OR i.index_id IS NULL OR NOT i.complete
 OR i.manifest_sha256<>v.manifest_sha256 OR i.payload_sha256<>v.payload_sha256
 OR c.health='missing' OR c.health='staged' OR (c.health='failed' AND c.last_failure_code IS DISTINCT FROM 'metadata_expired')))
 AND NOT EXISTS(SELECT 1 FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id)
 JOIN reference_pack_candidates c USING(pack_key,pack_version)
 WHERE c.health<>'verified_available' OR c.administratively_disabled OR c.removed)`).Scan(&ready)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("reference pack: required content unavailable")
	}
	return nil
}
