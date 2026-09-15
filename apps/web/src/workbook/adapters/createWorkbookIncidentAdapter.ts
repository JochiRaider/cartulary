import { validIncidentResource } from "../../shared/incidentResource";
import { normalizeIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import {
  invalidWorkbookAdapterResult,
  normalizeWorkbookAdapterFailure,
  workbookAdapterCaughtResult,
} from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createWorkbookIncidentAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookIncidentPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async getIdentity(input) {
      const message = "Incident identity load failed.";
      try {
        const outcome = await operations.execute({
          operationID: "getIncident",
          pathParameters: { incident_id: options.incidentId },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        if (!validIncidentResource(outcome.value.data, options.incidentId)) {
          return invalidWorkbookAdapterResult(message);
        }
        const identity = normalizeIncidentIdentity(
          options.incidentId,
          outcome.value.data,
        );
        return identity === null
          ? invalidWorkbookAdapterResult(message)
          : {
              kind: "accepted",
              value: { ...identity, resource: outcome.value.data },
            };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
  };
}
