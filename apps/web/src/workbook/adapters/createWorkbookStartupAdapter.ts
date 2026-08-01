import { normalizeWorkbookStartupSelection } from "../models/workbookStartup";
import type { WorkbookStartupPort } from "../startup/WorkbookStartupPort";
import {
  invalidWorkbookAdapterResult,
  normalizeWorkbookAdapterFailure,
  workbookAdapterCaughtResult,
} from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createWorkbookStartupAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookStartupPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async load(input) {
      const message = "Workbook startup load failed.";
      try {
        const outcome = await operations.execute({
          operationID: "getIncidentWorkbookStartup",
          pathParameters: { incident_id: options.incidentId },
          query: {
            ...(input.query.extensionProfileId === undefined
              ? {}
              : { extension_profile_id: input.query.extensionProfileId }),
            ...(input.query.sheetRefId === undefined
              ? {}
              : { sheet_ref_id: input.query.sheetRefId }),
            ...(input.query.sheetRefKind === undefined
              ? {}
              : { sheet_ref_kind: input.query.sheetRefKind }),
            ...(input.query.viewSchemaId === undefined
              ? {}
              : { view_schema_id: input.query.viewSchemaId }),
          },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        const data = outcome.value.data;
        if (
          data.incident_id !== options.incidentId ||
          data.extension_workspace_availability.incident_id !==
            options.incidentId
        ) {
          return invalidWorkbookAdapterResult(message);
        }
        const selection = normalizeWorkbookStartupSelection(data);
        if (selection === null) {
          return invalidWorkbookAdapterResult(message);
        }
        return {
          kind: "accepted",
          value: {
            availability: {
              workspaces: data.extension_workspace_availability.workspaces.map(
                (workspace) => ({
                  extensionProfileId: workspace.extension_profile_id,
                  workspaceKey: workspace.workspace_key,
                }),
              ),
            },
            selection,
          },
        };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
  };
}
