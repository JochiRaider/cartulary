import { requireViewContract } from "@cartulary/view-contracts";
import {
  type NoteAssociationReader,
  noteAssociationView,
} from "../features/notes/noteAssociationOperation";
import { createWorkbookAuthoringReader } from "./createWorkbookAuthoringReader";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createNoteAssociationReader(
  options: Parameters<typeof createWorkbookAuthoringReader>[0],
): NoteAssociationReader {
  const neutral = createWorkbookAuthoringReader(options),
    operations = createWorkbookOperationExecutor(options);
  return {
    page: neutral.page,
    availableViews: neutral.availableViews,
    async verify(kind, signal) {
      const expected = requireViewContract(
        noteAssociationView,
      ).inspectorConfig.featureGroups.find(
        (feature) =>
          feature.mutates &&
          feature.routeBinding.owner === "note_associations_route" &&
          feature.routeBinding.actionKey === kind,
      );
      const result = await operations.execute({
        operationID: "getViewSchema",
        pathParameters: { view_schema_id: noteAssociationView },
        signal,
      });
      if (
        !expected ||
        result.kind !== "accepted" ||
        result.value.data.view_schema_id !== noteAssociationView
      )
        throw new Error("Note association capability could not be verified.");
      const actual = result.value.data.inspector_config.feature_groups.find(
        (feature) => feature.feature_group_key === expected.featureGroupKey,
      );
      if (
        !actual ||
        actual.route_binding.kind !== "note_associations" ||
        actual.route_binding.owner !== "note_associations_route" ||
        actual.route_binding.action_key !== kind ||
        actual.route_binding.target_view_schema_id !== undefined ||
        actual.panel_id !== expected.panelId ||
        actual.minimum_incident_role !== expected.minimumIncidentRole ||
        actual.mutates !== true ||
        actual.requires_confirmation !== expected.requiresConfirmation ||
        actual.seed_bindings.length !== 0 ||
        JSON.stringify(actual.disabled_when) !==
          JSON.stringify(expected.disabledWhen) ||
        actual.success_result_behavior !== expected.successResultBehavior ||
        actual.failure_result_behavior !== expected.failureResultBehavior
      )
        throw new Error(
          "Note association capability changed. Refresh and review this Note.",
        );
    },
  };
}
