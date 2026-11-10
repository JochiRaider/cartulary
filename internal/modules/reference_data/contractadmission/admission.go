// Package contractadmission validates retained Reference Pack state against
// current contracts without changing it.
package contractadmission

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

// Reader grants inspection only; owners cannot mutate through this port.
type Reader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// IsIncompatible recognizes only this owner's closed contract-state rejection.
func IsIncompatible(err error) bool { var target incompatible; return errors.As(err, &target) }

type incompatible struct{}

func (incompatible) Error() string {
	return "Reference Pack retained state does not satisfy the current contract"
}
func (incompatible) ReasonCode() string { return "reference_pack_contract_state_incompatible" }

type check struct {
	query sqlshape.Fragment
	kind  string
}

var checks = []check{
	{sqlshape.Literal("SELECT v.manifest_bytes,c.distribution_kind FROM public.reference_pack_versions v JOIN public.reference_pack_candidates c USING(pack_key,pack_version)"), "manifest"},
	{sqlshape.Literal("SELECT canonical_set,pack_set_id,pack_set_sha256 FROM public.reference_pack_sets"), "set"},
	{sqlshape.Literal("SELECT canonical_references FROM public.reference_pack_portable_catalogs"), "references"},
	{sqlshape.Literal("SELECT canonical_references FROM public.reference_pack_portable_preparations"), "references"},
	{sqlshape.Literal("SELECT canonical_prepared FROM public.reference_pack_attempt_members WHERE canonical_prepared IS NOT NULL"), "prepared"},
	{sqlshape.Literal("SELECT frozen_input FROM public.reference_pack_operations"), "pending"},
}

func validate(kind string, b []byte) error {
	var err error
	switch kind {
	case "manifest:packaged_builtin", "manifest:operator_imported":
		_, err = packformat.DecodeManifest(b, kind == "manifest:operator_imported")
	case "set":
		_, err = packformat.DecodeSet(b)
	case "references":
		_, err = packformat.DecodeRetainedPortableReferences(b)
	case "prepared":
		err = packformat.ValidatePrepared(b)
	case "pending":
		err = packformat.ValidateRetainedOperationInput(b)
	default:
		return incompatible{}
	}
	if err != nil {
		return incompatible{}
	}
	return nil
}

func Validate(ctx context.Context, db Reader) error {
	if db == nil {
		return errors.New("reference pack admission database unavailable")
	}
	for _, c := range checks {
		rows, err := db.Query(ctx, sqlshape.Text(c.query))
		if err != nil {
			return err
		}
		for rows.Next() {
			var b []byte
			kind := c.kind
			if kind == "set" {
				var id, digest string
				err = rows.Scan(&b, &id, &digest)
				if err == nil {
					if _, decodeErr := packformat.DecodeBoundSet(b, id, digest); decodeErr != nil {
						err = incompatible{}
					}
				}
			} else if kind == "manifest" {
				var distribution string
				err = rows.Scan(&b, &distribution)
				kind += ":" + distribution
			} else {
				err = rows.Scan(&b)
			}
			if err == nil && kind != "set" {
				err = validate(kind, b)
			}
			if err != nil {
				rows.Close()
				return err
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return validateLocalSets(ctx, db)
}
