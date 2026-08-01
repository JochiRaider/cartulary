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
        if (outcome.value.data.incident_id !== options.incidentId) {
          return invalidWorkbookAdapterResult(message);
        }
        const identity = normalizeIncidentIdentity(
          options.incidentId,
          outcome.value.data,
        );
        return identity === null
          ? invalidWorkbookAdapterResult(message)
          : { kind: "accepted", value: identity };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
    async listMembers(input) {
      const message = "Incident member references are unavailable.";
      try {
        const outcome = await operations.execute({
          operationID: "listIncidentMemberships",
          pathParameters: { incident_id: options.incidentId },
          signal: input.signal,
        });
        if (outcome.kind === "rejected") {
          return normalizeWorkbookAdapterFailure(outcome, message);
        }
        if (
          outcome.value.data.memberships.some(
            (membership) => membership.incident_id !== options.incidentId,
          )
        ) {
          return invalidWorkbookAdapterResult(message);
        }
        return {
          kind: "accepted",
          value: {
            members: outcome.value.data.memberships.map((membership) => ({
              displayName: membership.display_name,
              userId: membership.user_id,
            })),
          },
        };
      } catch (error) {
        return workbookAdapterCaughtResult(error, input.signal, message);
      }
    },
  };
}
