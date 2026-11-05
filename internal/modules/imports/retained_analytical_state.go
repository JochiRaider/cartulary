package imports

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/jackc/pgx/v5"
)

// RetainedAnalyticalStatePresent supports a source owner's fresh-state boundary
// before migrations, including installations predating analytical import tables.
func RetainedAnalyticalStatePresent(ctx context.Context, conn *sql.Conn, profile string) (bool, error) {
	for _, entry := range []struct{ table, column, query string }{
		{"import_units", "approved_extension_profile_id", "SELECT EXISTS (SELECT 1 FROM public.import_units WHERE approved_extension_profile_id=$1)"},
		{"import_apply_unit_plans", "extension_profile_id", "SELECT EXISTS (SELECT 1 FROM public.import_apply_unit_plans WHERE extension_profile_id=$1)"},
		{"import_unit_apply_outcomes", "extension_profile_id", "SELECT EXISTS (SELECT 1 FROM public.import_unit_apply_outcomes WHERE extension_profile_id=$1)"},
	} {
		var present bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 AND column_name=$2)`, entry.table, entry.column).Scan(&present); err != nil {
			return false, err
		}
		if !present {
			continue
		}
		var retained bool
		if err := conn.QueryRowContext(ctx, entry.query, profile).Scan(&retained); err != nil {
			return false, err
		}
		if retained {
			return true, nil
		}
	}
	return false, nil
}

// AnalyticalStateReader is borrowed from the caller's read-only admission
// snapshot. Imports owns physical storage; source owners validate only payloads.
type AnalyticalStateReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type RetainedAnalyticalPayload struct {
	Kind     string // mapping or applied_result
	SchemaID string
	Payload  json.RawMessage
}

// ValidateRetainedAnalyticalState streams retained units, plans and outcomes.
// It never expires receipts, repairs a plan or loads source bytes. The visitor
// must not issue queries on the borrowed reader while its stream is open.
func ValidateRetainedAnalyticalState(ctx context.Context, reader AnalyticalStateReader, profile string, visit func(RetainedAnalyticalPayload) error) error {
	invalid := errors.New("retained analytical import contract is incompatible")
	rows, err := reader.Query(ctx, `
SELECT 'mapping', approved_target_kind, '', '', approved_mapping_json
  FROM import_units WHERE approved_extension_profile_id=$1
UNION ALL
SELECT 'plan', target_kind, owner_binding_id, target_registry_sha256, approved_mapping_json
  FROM import_apply_unit_plans WHERE extension_profile_id=$1
UNION ALL
SELECT 'result', target_kind, owner_binding_id, outcome_status, owner_result_json
  FROM import_unit_apply_outcomes WHERE extension_profile_id=$1`, profile)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var kind, targetKind, binding, version string
		var raw json.RawMessage
		if err := rows.Scan(&kind, &targetKind, &binding, &version, &raw); err != nil {
			return err
		}
		target, ok := analyticalImportTargets[analyticalImportTargetKey{TargetKind: targetKind, ExtensionProfileID: profile}]
		if !ok || kind != "mapping" && binding != target.FacadeBindingID {
			return invalid
		}
		if kind == "plan" && version != importtargetregistry.RegistrySHA256 {
			return invalid
		}
		if kind == "result" {
			if version != "applied" {
				continue
			}
			if err := visit(RetainedAnalyticalPayload{Kind: "applied_result", SchemaID: target.ApplyResultID, Payload: raw}); err != nil {
				return err
			}
			continue
		}
		var mapping approvedMapping
		if json.Unmarshal(raw, &mapping) != nil || mapping.TargetKind != targetKind || mapping.ExtensionProfileID != profile || mapping.OwnerMappingSchemaID != target.MappingSchemaID {
			return invalid
		}
		if err := visit(RetainedAnalyticalPayload{Kind: "mapping", SchemaID: mapping.OwnerMappingSchemaID, Payload: mapping.OwnerMapping}); err != nil {
			return err
		}
	}
	return rows.Err()
}
