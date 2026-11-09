package contractadmission

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

// ValidateLocalProvenance authenticates authority against exact retained
// manifest bytes. allowAbsent is only for portable source history, whose
// missing content cannot be treated as available destination content.
func ValidateLocalProvenance(ctx context.Context, db Reader, data []byte, set packformat.Set, allowAbsent bool) error {
	if packformat.ValidateProvenance(data, set) != nil {
		return incompatible{}
	}
	for _, member := range set.Members {
		var raw []byte
		var distribution string
		err := db.QueryRow(ctx, `SELECT v.manifest_bytes,c.distribution_kind FROM public.reference_pack_versions v JOIN public.reference_pack_candidates c USING(pack_key,pack_version) WHERE v.pack_key=$1 AND v.pack_version=$2`, member.Key, member.Version).Scan(&raw, &distribution)
		if errors.Is(err, pgx.ErrNoRows) && allowAbsent {
			continue
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return incompatible{}
		}
		if err != nil {
			return err
		}
		if distribution != "operator_imported" && distribution != "packaged_builtin" || packformat.ValidateManifestProvenance(data, member, raw, distribution == "operator_imported") != nil {
			return incompatible{}
		}
	}
	return nil
}

func validateLocalSets(ctx context.Context, db Reader) error {
	after := ""
	for {
		var id, digest string
		var raw, provenance []byte
		err := db.QueryRow(ctx, `SELECT pack_set_id,pack_set_sha256,canonical_set,canonical_provenance FROM public.reference_pack_sets WHERE pack_set_id>$1 ORDER BY pack_set_id LIMIT 1`, after).Scan(&id, &digest, &raw, &provenance)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		set, err := packformat.DecodeBoundSet(raw, id, digest)
		if err != nil {
			return incompatible{}
		}
		if err := ValidateLocalProvenance(ctx, db, provenance, set, false); err != nil {
			return err
		}
		after = id
	}
}
