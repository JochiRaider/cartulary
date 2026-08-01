import { sheetRefsEqual } from "../../shared/sheetRef";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import {
  invalidWorkbookAdapterResult,
  normalizeWorkbookAdapterFailure,
  workbookAdapterCaughtResult,
} from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createWorkbookPreferenceAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookPreferencePort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async setDefaultSheet(input) {
      const message = "Workbook default preference update failed.";
      try {
        const outcome = await operations.execute({
          operationID: "putIncidentDefaultWorkbookPreferences",
          pathParameters: { incident_id: options.incidentId },
          request: { default_sheet_ref: input.sheetRef },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        return outcome.value.data.incident_id === options.incidentId &&
          outcome.value.data.default_sheet_ref !== null &&
          sheetRefsEqual(outcome.value.data.default_sheet_ref, input.sheetRef)
          ? { kind: "accepted", value: undefined }
          : invalidWorkbookAdapterResult(message);
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
    async setHomeSheet(input) {
      const message = "Workbook home preference update failed.";
      try {
        const outcome = await operations.execute({
          operationID: "putCurrentUserWorkbookPreferences",
          pathParameters: { incident_id: options.incidentId },
          request: { home_sheet_ref: input.sheetRef },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        return outcome.value.data.incident_id === options.incidentId &&
          outcome.value.data.home_sheet_ref !== null &&
          sheetRefsEqual(outcome.value.data.home_sheet_ref, input.sheetRef)
          ? { kind: "accepted", value: undefined }
          : invalidWorkbookAdapterResult(message);
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
  };
}
