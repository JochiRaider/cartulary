import { requireViewContract } from "@cartulary/view-contracts";
import {
  type NoteCreateReader,
  noteCreateFeature,
  noteCreateView,
  noteFeature,
} from "../features/notes/noteCreateModel";
import { createWorkbookAuthoringReader } from "./createWorkbookAuthoringReader";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createNoteCreateReader(
  options: Parameters<typeof createWorkbookAuthoringReader>[0],
): NoteCreateReader {
  const neutral = createWorkbookAuthoringReader(options);
  const operations = createWorkbookOperationExecutor(options);
  return {
    page: neutral.page,
    availableViews: neutral.availableViews,
    async verifyNote(draft, signal) {
      const target = requireViewContract(noteCreateView);
      const result = await operations.execute({
        operationID: "getViewSchema",
        pathParameters: { view_schema_id: noteCreateView },
        signal,
      });
      if (result.kind !== "accepted")
        throw new Error("Note creation could not be verified.");
      const schema = result.value.data;
      if (
        schema.view_schema_id !== noteCreateView ||
        !schema.create_capable ||
        schema.inline_create.permits_zero_field_create ||
        schema.create_inputs.length ||
        JSON.stringify(schema.inline_create.minimum_create_field_sets) !==
          JSON.stringify(target.minimumCreateFieldSets) ||
        schema.fields.length !== target.fields.length ||
        schema.fields.some((field) => {
          const expected = target.fieldMap[field.field_key];
          return (
            !expected ||
            field.create_writable !== expected.createWritable ||
            field.read_kind !== expected.readKind ||
            field.write_kind !== expected.writeKind ||
            field.clearable !== expected.clearable ||
            field.string_contract_id !== expected.stringContractId ||
            field.direct_scalar_contract_id !==
              expected.directScalarContractId ||
            field.direct_reference_contract_id !==
              expected.directReferenceContractId ||
            JSON.stringify(field.enum_values) !==
              JSON.stringify(expected.enumValues)
          );
        })
      )
        throw new Error(
          "Note creation capability changed. Review the retained draft.",
        );
      if (!draft.source) return;
      const expected = noteFeature(draft.source.viewSchemaId);
      const source = await operations.execute({
        operationID: "getViewSchema",
        pathParameters: { view_schema_id: draft.source.viewSchemaId },
        signal,
      });
      if (
        !expected ||
        source.kind !== "accepted" ||
        source.value.data.view_schema_id !== draft.source.viewSchemaId
      )
        throw new Error("Note source capability could not be verified.");
      const feature = source.value.data.inspector_config.feature_groups.find(
        (item) => item.feature_group_key === noteCreateFeature,
      );
      const route = feature?.route_binding;
      if (
        !feature ||
        route?.kind !== "record_action" ||
        route.owner !== "record_linked_note_create_route" ||
        route.action_key !== noteCreateFeature ||
        route.target_view_schema_id !== undefined ||
        feature.panel_id !== "workflow" ||
        feature.minimum_incident_role !== expected.minimumIncidentRole ||
        feature.requires_confirmation !== false ||
        feature.mutates !== true ||
        feature.seed_bindings.length ||
        JSON.stringify(feature.disabled_when) !==
          JSON.stringify(expected.disabledWhen) ||
        feature.success_result_behavior !== expected.successResultBehavior ||
        feature.failure_result_behavior !== expected.failureResultBehavior
      )
        throw new Error(
          "Note source capability changed. Review the retained draft.",
        );
    },
  };
}
