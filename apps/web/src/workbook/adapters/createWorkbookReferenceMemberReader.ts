import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import type { WorkbookReferenceMemberPage } from "../ports/WorkbookReferenceReadPort";
import {
  invalidWorkbookAdapterResult,
  normalizeWorkbookAdapterFailure,
  workbookAdapterCaughtResult,
} from "./workbookAdapterResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

/** Membership paging retains its route's ordering and opaque continuation. */
export function createWorkbookReferenceMemberReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}) {
  const operations = createWorkbookOperationExecutor(options);
  return async (
    cursorToken: string | undefined,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<WorkbookReferenceMemberPage>> => {
    const message = "Incident members could not be loaded. Retry the read.";
    try {
      const outcome = await operations.execute({
        operationID: "listIncidentMemberships",
        pathParameters: { incident_id: options.incidentId },
        query: {
          limit: 100,
          ...(cursorToken ? { cursor_token: cursorToken } : {}),
        },
        signal,
      });
      if (outcome.kind === "rejected")
        return normalizeWorkbookAdapterFailure(outcome, message);
      const { memberships } = outcome.value.data;
      const paging = outcome.value.meta.paging;
      if (
        !paging ||
        paging.limit !== 100 ||
        memberships.length > 100 ||
        paging.has_more !== (paging.next_cursor !== null) ||
        (paging.next_cursor !== null &&
          (!paging.next_cursor || paging.next_cursor === cursorToken)) ||
        memberships.some(
          (member) =>
            member.incident_id !== options.incidentId || !member.user_id,
        ) ||
        new Set(memberships.map((member) => member.user_id)).size !==
          memberships.length
      ) {
        return invalidWorkbookAdapterResult(message);
      }
      return signal.aborted
        ? { kind: "aborted" }
        : {
            kind: "accepted",
            value: {
              members: memberships.map((member) => ({
                userId: member.user_id,
                displayName: member.display_name,
              })),
              paging: {
                limit: paging.limit,
                hasMore: paging.has_more,
                nextCursor: paging.next_cursor,
              },
            },
          };
    } catch (error) {
      return workbookAdapterCaughtResult(error, signal, message);
    }
  };
}
