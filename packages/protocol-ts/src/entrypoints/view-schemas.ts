import { viewInspectorRegistry as generatedViewInspectorRegistry } from "../generated/view-inspector-registry.js";
import { viewSchemaRegistry as generatedViewSchemaRegistry } from "../generated/view-schema-registry.js";
import type { ViewSchemaSourceDocument } from "../generated/view-schema-source-types.js";
import { validateCartularyViewSchemaSourceV1 } from "../generated/view-schema-source-validator.js";
import { viewSchemaArtifacts } from "../generated/view-schemas-artifacts.js";
import { createDecoder } from "../internal/decoder.js";

export type * from "../generated/view-schema-source-types.js";
export { viewSchemaArtifacts };

export const viewSchemaSourceDocumentDecoder =
  createDecoder<ViewSchemaSourceDocument>(
    "cartulary.view_schema_source.v1",
    validateCartularyViewSchemaSourceV1,
  );

export const viewInspectorRegistry = generatedViewInspectorRegistry;
export type ViewInspectorRegistry = typeof viewInspectorRegistry;
export type ViewInspectorPanelID =
  ViewInspectorRegistry["vocabularies"]["panels"][number];
export type ViewInspectorRouteKind =
  ViewInspectorRegistry["vocabularies"]["route_kinds"][number];
export type ViewInspectorRouteOwner =
  ViewInspectorRegistry["vocabularies"]["route_owners"][number];
export type ViewInspectorDisabledCondition =
  ViewInspectorRegistry["vocabularies"]["disabled_conditions"][number];
export type ViewInspectorSuccessResultBehavior =
  ViewInspectorRegistry["vocabularies"]["success_result_behaviors"][number];
export type ViewInspectorFailureResultBehavior =
  ViewInspectorRegistry["vocabularies"]["failure_result_behaviors"][number];
export type ViewInspectorSeedSourceKind =
  ViewInspectorRegistry["vocabularies"]["seed_source_kinds"][number];
export type ViewInspectorIncidentRole =
  ViewInspectorRegistry["vocabularies"]["incident_roles"][number];
export type ViewInspectorSpecializedActionKey =
  ViewInspectorRegistry["specialized_features"][number]["action_key"];

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
