// Package contractadmission owns the read-only native Reference Pack cutover.
// It inspects current tags and contracts; it never decodes or converts a legacy
// representation for consumption, and never changes retained state.
package contractadmission

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Reader grants inspection only; owners cannot mutate through this port.
type Reader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type SQLReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// IsIncompatible recognizes only this owner's closed contract-state rejection.
func IsIncompatible(err error) bool { var target incompatible; return errors.As(err, &target) }

type incompatible struct{}

func (incompatible) Error() string {
	return "Reference Pack v2 requires current retained contracts; incompatible historical evidence cannot enter runtime"
}
func (incompatible) ReasonCode() string { return "reference_pack_contract_state_incompatible" }

type check struct{ table, query, kind string }

var checks = []check{
	{"reference_pack_versions", "SELECT manifest_bytes FROM public.reference_pack_versions", "manifest"},
	{"reference_pack_sets", "SELECT canonical_set FROM public.reference_pack_sets", "set"},
	{"reference_pack_portable_catalogs", "SELECT canonical_references FROM public.reference_pack_portable_catalogs", "references"},
	{"reference_pack_portable_preparations", "SELECT canonical_references FROM public.reference_pack_portable_preparations", "references"},
	{"reference_pack_attempt_members", "SELECT canonical_prepared FROM public.reference_pack_attempt_members WHERE canonical_prepared IS NOT NULL", "prepared"},
	{"reference_pack_operations", "SELECT frozen_input FROM public.reference_pack_operations WHERE terminal_at IS NULL", "pending"},
}

func validate(kind string, b []byte) error {
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return incompatible{}
	}
	m, ok := v.(map[string]any)
	if !ok {
		return incompatible{}
	}
	switch kind {
	case "manifest":
		if m["pack_contract_version"] != packformat.ContractID {
			return incompatible{}
		}
		var manifest packformat.Manifest
		if err = json.Unmarshal(b, &manifest); err == nil {
			err = packformat.ValidateManifestContract(manifest)
		}
	case "set":
		if m["schema_id"] != "cartulary.reference_pack_set.v2" {
			return incompatible{}
		}
		var retained packformat.Set
		if err = json.Unmarshal(b, &retained); err == nil {
			var actual packformat.Set
			actual, err = packformat.BuildSet(retained.Members)
			if err == nil && (actual.ID != retained.ID || actual.SHA256 != retained.SHA256) {
				err = errors.New("set identity mismatch")
			}
		}
	case "references":
		_, err = packformat.DecodePortableReferences(b)
	case "prepared":
		err = packformat.ValidatePrepared(b)
	case "pending":
		if m["schema_id"] != "cartulary.reference_pack_operation_input.v3" && m["schema_id"] != "cartulary.reference_pack_portability_input.v2" {
			return incompatible{}
		}
		if m["schema_id"] == "cartulary.reference_pack_operation_input.v3" && m["pack_contract_version"] != packformat.ContractID {
			return incompatible{}
		}
		err = rejectRetired(v)
	}
	if err != nil {
		return incompatible{}
	}
	return nil
}

func rejectRetired(v any) error {
	switch x := v.(type) {
	case map[string]any:
		if contract, present := x["pack_contract_version"]; present && contract != packformat.ContractID {
			return incompatible{}
		}
		if schema, present := x["schema_id"]; present && (schema == "reference_pack_refs.v1" || schema == "cartulary.reference_pack_set.v1") {
			return incompatible{}
		}
		for _, child := range x {
			if err := rejectRetired(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range x {
			if err := rejectRetired(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateMigration(ctx context.Context, conn SQLReader) error {
	for _, c := range checks {
		var present bool
		if err := conn.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+c.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		rows, err := conn.QueryContext(ctx, c.query)
		if err != nil {
			return err
		}
		for rows.Next() {
			var b []byte
			if err = rows.Scan(&b); err == nil {
				err = validate(c.kind, b)
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
	return nil
}

func Validate(ctx context.Context, db Reader) error {
	if db == nil {
		return errors.New("reference pack admission database unavailable")
	}
	for _, c := range checks {
		var present bool
		if err := db.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+c.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		rows, err := db.Query(ctx, c.query)
		if err != nil {
			return err
		}
		for rows.Next() {
			var b []byte
			if err = rows.Scan(&b); err == nil {
				err = validate(c.kind, b)
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
	return nil
}
