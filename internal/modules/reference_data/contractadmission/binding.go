package contractadmission

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Binding admission exposes only the owner's immutable contract check. Calling
// owners supply their retained public binding and never inspect storage tables.
// A portable catalog can retain a compatible set whose content is unavailable.
const bindingSetQuery = `SELECT COALESCE(
 (SELECT canonical_set FROM public.reference_pack_sets WHERE pack_set_id=$1),
 (SELECT convert_to(s::text,'UTF8') FROM public.reference_pack_portable_catalogs c,
 jsonb_array_elements(convert_from(c.canonical_references,'UTF8')::jsonb->'sets') s
 WHERE s->>'pack_set_id'=$1 LIMIT 1))`

func ValidateBindingMigration(ctx context.Context, conn SQLReader, binding []byte) error {
	return validateBinding(binding, func(id string) ([]byte, error) {
		var b []byte
		err := conn.QueryRowContext(ctx, bindingSetQuery, id).Scan(&b)
		return b, err
	})
}
func ValidateBinding(ctx context.Context, db Reader, binding []byte) error {
	return validateBinding(binding, func(id string) ([]byte, error) {
		var b []byte
		err := db.QueryRow(ctx, bindingSetQuery, id).Scan(&b)
		return b, err
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
	if validate("set", raw) != nil {
		return incompatible{}
	}
	var set packformat.Set
	if json.Unmarshal(raw, &set) != nil || set.SHA256 != binding.SHA {
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
