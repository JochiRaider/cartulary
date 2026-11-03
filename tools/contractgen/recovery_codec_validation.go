package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Roles describe the formats required by capture/restore. Versions and paths
// come from the authored registry and must resolve to current schema resources.
var recoveryCodecRoles = []string{
	"backup_artifact_envelope", "backup_integrity_manifest", "extension_postgres_binding_proof",
	"graph_projection_restore_implementation_binding", "graph_projection_restore_source_registry",
	"object_store_backup_manifest", "object_store_backup_summary", "operator_recovery_journal_payload",
	"postgres_snapshot_artifact", "postgres_snapshot_unit",
}

func requiredRecoveryCodecSchemas(registry map[string]any) ([]string, error) {
	roles, err := asObject(registry["codec_schemas"], "Recovery codec schemas")
	if err != nil {
		return nil, err
	}
	if err := requireAllowedKeys(roles, stringSet(recoveryCodecRoles...), "Recovery codec schemas"); err != nil {
		return nil, err
	}
	current, err := stringArray(registry["current_schema_ids"], "Recovery current schemas", true)
	if err != nil {
		return nil, err
	}
	retired, err := stringArray(registry["retired_schema_ids"], "Recovery retired schemas", true)
	if err != nil {
		return nil, err
	}
	for _, id := range current {
		if slices.Contains(retired, id) {
			return nil, fmt.Errorf("recovery schema %s is both current and retired", id)
		}
	}
	ids := make([]string, 0, len(recoveryCodecRoles))
	for _, role := range recoveryCodecRoles {
		id, err := requiredString(roles, role, "Recovery codec schemas")
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(id, "cartulary."+role+".v") || !slices.Contains(current, id) || slices.Contains(retired, id) {
			return nil, fmt.Errorf("recovery codec role %s must name its current schema", role)
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if err := requireSortedUniqueStrings(ids, "Recovery required codec schemas"); err != nil {
		return nil, err
	}
	return ids, nil
}

func resolveRecoveryCodecSchemas(base string) ([]string, error) {
	registry, err := readRecoveryObject(base, "index.json")
	if err != nil {
		return nil, err
	}
	ids, err := requiredRecoveryCodecSchemas(registry)
	if err != nil {
		return nil, err
	}
	resources, err := recoverySchemaResources(base, registry)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		if resources[id] == nil {
			return nil, fmt.Errorf("recovery codec %s has no packaged schema", id)
		}
	}
	return ids, nil
}

func recoverySchemaResources(base string, registry map[string]any) (map[string]map[string]any, error) {
	paths, err := stringArray(registry["schemas"], "Recovery schemas", true)
	if err != nil {
		return nil, err
	}
	resources := make(map[string]map[string]any, len(paths))
	for _, path := range paths {
		schema, err := readRecoveryObject(base, path)
		if err != nil {
			return nil, err
		}
		id, err := requiredString(schema, "$id", path)
		if err != nil {
			return nil, err
		}
		if resources[id] != nil {
			return nil, fmt.Errorf("duplicate Recovery schema %s", id)
		}
		resources[id] = schema
	}
	return resources, nil
}

func validateRecoverySnapshotSchemaBounds(base string, registry map[string]any) error {
	limits, err := asObject(registry["limits"], "Recovery limits")
	if err != nil {
		return err
	}
	maximum, err := requiredInt(limits, "maximum_tables", "Recovery limits")
	if err != nil {
		return err
	}
	roles, err := asObject(registry["codec_schemas"], "Recovery codec schemas")
	if err != nil {
		return err
	}
	id, err := requiredString(roles, "postgres_snapshot_artifact", "Recovery codec schemas")
	if err != nil {
		return err
	}
	resources, err := recoverySchemaResources(base, registry)
	if err != nil {
		return err
	}
	schema := resources[id]
	if schema == nil {
		return fmt.Errorf("recovery snapshot codec has no packaged schema")
	}
	properties, err := asObject(schema["properties"], "snapshot properties")
	if err != nil {
		return err
	}
	units, err := asObject(properties["units"], "snapshot units")
	if err != nil {
		return err
	}
	minimum, err := requiredInt(units, "minItems", "snapshot units")
	if err != nil {
		return err
	}
	bound, err := requiredInt(units, "maxItems", "snapshot units")
	if err != nil {
		return err
	}
	if minimum != 0 || maximum <= minimum || bound != maximum {
		return fmt.Errorf("snapshot units must use bounded structural limits; exact membership comes from the frozen catalog")
	}
	return nil
}
