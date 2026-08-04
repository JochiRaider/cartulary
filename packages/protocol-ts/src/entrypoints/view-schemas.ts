import { viewSchemaRegistry as generatedViewSchemaRegistry } from "../generated/view-schema-registry.js";
import { viewSchemaArtifacts } from "../generated/view-schemas-artifacts.js";

export type * from "../generated/view-schema-source-types.js";
export { viewSchemaArtifacts };

export type ViewSchemaRegistryEntry = {
  readonly artifact_path: string;
  readonly required_reference_pack_keys: readonly string[];
  readonly source_record_types: readonly string[];
  readonly surface_kind: "built_in_sheet" | "system_view";
  readonly surface_status:
    | "required_built_in_sheet"
    | "required_system_view"
    | "standardized_optional_workbook_surface";
  readonly title: string;
  readonly view_schema_id: string;
};
type ViewSchemaRegistryProjection = {
  readonly $schema: string;
  readonly note: string;
  readonly registry_id: "cartulary.view_schemas.base.v1";
  readonly view_schemas: readonly ViewSchemaRegistryEntry[];
};
export const viewSchemaRegistry: ViewSchemaRegistryProjection =
  generatedViewSchemaRegistry;

export function listViewSchemaRegistryEntries(): readonly ViewSchemaRegistryEntry[] {
  return viewSchemaRegistry.view_schemas;
}
