package contractadmission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Binding admission exposes only the owner's immutable contract check. Calling
// owners supply their retained public binding and never inspect storage tables.
// A portable catalog can retain a compatible set whose content is unavailable.
const bindingSetQuery = `SELECT canonical_set,pack_set_sha256 FROM public.reference_pack_sets WHERE pack_set_id=$1`

func ValidateBinding(ctx context.Context, db Reader, binding []byte) error {
	return validateBinding(binding, func(id string) ([]byte, error) {
		var b []byte
		var digest string
		err := db.QueryRow(ctx, bindingSetQuery, id).Scan(&b, &digest)
		if err == nil {
			if _, decodeErr := packformat.DecodeBoundSet(b, id, digest); decodeErr != nil {
				return nil, incompatible{}
			}
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return b, err
		}
		// Inspect retained canonical bytes, never JSONB's reserialized set text.
		rows, err := db.Query(ctx, `SELECT canonical_references FROM public.reference_pack_portable_catalogs ORDER BY incident_id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&b); err != nil {
				return nil, err
			}
			refs, err := packformat.DecodeRetainedPortableReferences(b)
			if err != nil {
				return nil, incompatible{}
			}
			for _, set := range refs.Sets {
				if set.ID == id {
					return canonicaljson.Marshal(set)
				}
			}
		}
		return nil, rows.Err()
	})
}
func validateBinding(b []byte, load func(string) ([]byte, error)) error {
	var binding struct {
		ID         string          `json:"pack_set_id"`
		SHA        string          `json:"pack_set_sha256"`
		Provenance json.RawMessage `json:"provenance"`
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&binding) != nil || binding.ID != "rpset_"+binding.SHA {
		return incompatible{}
	}
	raw, err := load(binding.ID)
	if err != nil {
		return err
	}
	set, err := packformat.DecodeBoundSet(raw, binding.ID, binding.SHA)
	if err != nil {
		return incompatible{}
	}
	v, err := canonicaljson.DecodeStrict(binding.Provenance)
	if err != nil {
		return incompatible{}
	}
	canonical, err := canonicaljson.Marshal(v)
	if err != nil || packformat.ValidateProvenance(canonical, set) != nil {
		return incompatible{}
	}
	return nil
}
