import { requireViewContract } from "@cartulary/view-contracts";
import { decisionViewId } from "../features/coordination/decisionSupersessionModel";
import type { DecisionSupersessionReadPort } from "../features/coordination/decisionSupersessionOperation";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import {
  buildQueryRequest,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createDecisionCandidateReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope?: WorkbookReadScopeSource;
}): DecisionSupersessionReadPort {
  const operations = createWorkbookOperationExecutor(options);
  const contract = requireViewContract(decisionViewId);
  return {
    async page(cursor, signal) {
      const scope = options.readScope?.() ?? null;
      try {
        const outcome = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: decisionViewId,
          },
          request: {
            ...buildQueryRequest(contract, emptyWorkbookQueryState()),
            limit: 100,
            ...(cursor === null ? {} : { cursor_token: cursor }),
          },
          signal,
        });
        if (
          signal.aborted ||
          (options.readScope &&
            !sameWorkbookReadScope(scope, options.readScope()))
        )
          return { kind: "aborted" };
        if (outcome.kind === "rejected") return outcome;
        const { data, meta } = outcome.value;
        if (
          data.incident_id !== options.incidentId ||
          data.view_schema_id !== decisionViewId ||
          !meta.paging ||
          (!meta.paging.has_more && meta.paging.next_cursor !== null) ||
          (meta.paging.has_more &&
            (!meta.paging.next_cursor || meta.paging.next_cursor === cursor))
        )
          return {
            kind: "rejected",
            failure: {
              kind: "invalid_contract",
              message:
                "Decision candidates could not be verified. Retry the read.",
            },
          };
        return {
          kind: "accepted",
          value: {
            rows: normalizeWorkbookViewRows(
              contract,
              data.rows,
              "Decision candidates",
            ).map((row) => acceptWorkbookRowObservation(row, scope)),
            hasMore: meta.paging.has_more,
            nextCursor: meta.paging.next_cursor,
          },
        };
      } catch {
        return signal.aborted
          ? { kind: "aborted" }
          : {
              kind: "rejected",
              failure: {
                kind: "retryable",
                message:
                  "Decision candidates could not be loaded. Retry the read.",
              },
            };
      }
    },
  };
}
