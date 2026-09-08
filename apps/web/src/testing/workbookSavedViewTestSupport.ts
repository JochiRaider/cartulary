import { requireViewContract } from "@cartulary/view-contracts";
import { useLayoutEffect, useState } from "react";
import { observeAccountOperation } from "../app/accountOperation";
import type { AuthorizationRecoveryPort } from "../shared/authorizationRecovery";
import { createWorkbookSavedViewAdapter } from "../workbook/adapters/createWorkbookSavedViewAdapter";
import { buildSavedViewLayoutJson } from "../workbook/models/workbookQuery";
import type { SavedViewResource } from "../workbook/models/workbookSavedViews";
import type { WorkbookSavedViewPort } from "../workbook/ports/WorkbookSavedViewPort";
import type {
  SavedViewAuthority,
  SavedViewBinding,
} from "../workbook/savedviews/savedViewOperationModel";
import { WorkbookSavedViewController } from "../workbook/savedviews/WorkbookSavedViewController";

export const savedViewTestAuthority: SavedViewAuthority = {
  incidentId: "incident-1",
  actorId: "user-1",
  lifetime: "session-1",
  role: "admin",
};
export function savedViewTestResource(
  overrides: Partial<SavedViewResource> = {},
): SavedViewResource {
  const schema = overrides.view_schema_id ?? "cartulary.view.timeline.v2";
  return {
    incident_id: "incident-1",
    saved_view_id: "saved-1",
    view_schema_id: schema,
    display_name: "Timeline view",
    scope: "private",
    query_json: { sort: [], filters: [] },
    layout_json: buildSavedViewLayoutJson(requireViewContract(schema)),
    owner_user_id: "user-1",
    saved_view_version: 1,
    created_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-01T00:00:00Z",
    ...overrides,
  };
}
export function createSavedViewTestController(port: WorkbookSavedViewPort) {
  const controller = new WorkbookSavedViewController({
    port: () => port,
    observe: observeAccountOperation,
    isCurrent: () => true,
    recover: async (authority) => ({
      kind: "authorized",
      role: authority.role,
      userId: authority.actorId,
    }),
    lost: () => {},
  });
  controller.setAuthority(savedViewTestAuthority);
  return controller;
}

/** Shell tests supply the same application composition boundary as production. */
export function useSavedViewTestApplication(
  incidentId: string,
  actorId: string,
  recovery: AuthorizationRecoveryPort,
) {
  const [controller] = useState(
    () =>
      new WorkbookSavedViewController({
        port: (a) =>
          createWorkbookSavedViewAdapter({
            incidentId: a.incidentId,
            apiBase: a.apiBase,
          }),
        observe: observeAccountOperation,
        isCurrent: () => true,
        recover: (a, signal) =>
          recovery.recover({ incidentId: a.incidentId, signal }),
        lost: () => {},
      }),
  );
  useLayoutEffect(() => () => controller.retire(), [controller]);
  return {
    savedViewController: controller,
    bindWorkbookSavedViews: (binding: SavedViewBinding | null) => {
      if (binding)
        controller.setAuthority({
          incidentId,
          actorId,
          lifetime: "session",
          role: "admin",
          apiBase: binding.apiBase,
        });
      controller.setBinding(binding);
    },
  };
}
