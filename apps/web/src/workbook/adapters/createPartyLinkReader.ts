import { requireViewContract } from "@cartulary/view-contracts";
import { boundedRead } from "../../services/asyncObservation";
import {
  type PartyLinkReadPort,
  type PartyPage,
  partyPairs,
  partyViewId,
} from "../features/parties/partyLinkModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import {
  buildQueryRequest,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createPartyLinkReader(options: {
  apiBase: string | undefined;
  incidentId: string;
  recheckAuthority?: (() => void) | undefined;
}): PartyLinkReadPort {
  const operations = createWorkbookOperationExecutor(options);
  async function page(
    viewSchemaId: string,
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<PartyPage> {
    const contract = requireViewContract(viewSchemaId);
    const result = await operations.execute({
      operationID: "queryWorkbookView",
      pathParameters: {
        incident_id: options.incidentId,
        view_schema_id: viewSchemaId,
      },
      request: {
        ...buildQueryRequest(contract, emptyWorkbookQueryState()),
        limit: 100,
        ...(cursor ? { cursor_token: cursor } : {}),
      },
      signal,
    });
    if (signal.aborted) throw new Error("Read interrupted.");
    if (result.kind === "rejected") {
      if (
        workbookFailureLifecycle(result.failure).kind ===
        "authority_unavailable"
      )
        options.recheckAuthority?.();
      throw new Error("Current records could not be loaded. Retry the read.");
    }
    const { data, meta } = result.value;
    if (
      data.incident_id !== options.incidentId ||
      data.view_schema_id !== viewSchemaId ||
      !meta.paging ||
      data.rows.length > 100 ||
      (meta.paging.has_more
        ? !meta.paging.next_cursor || meta.paging.next_cursor === cursor
        : meta.paging.next_cursor !== null)
    )
      throw new Error("Record paging could not be verified. Reload records.");
    return {
      rows: normalizeWorkbookViewRows(
        contract,
        data.rows,
        "Party workflow records",
      ),
      hasMore: meta.paging.has_more,
      nextCursor: meta.paging.next_cursor,
    };
  }
  return {
    page: (cursor, signal) => page(partyViewId, cursor, signal),
    source(view, recordId, signal) {
      if (
        view !== partyViewId &&
        !partyPairs.some((pair) => pair.viewSchemaId === view)
      )
        return Promise.reject(new Error("Unsupported Party source."));
      return boundedRead(async (currentSignal) => {
        let cursor: string | null = null;
        const visited = new Set<string>();
        do {
          const result = await page(view, cursor, currentSignal);
          const row = result.rows.find((row) => row.record_id === recordId);
          if (row) return row;
          if (!result.hasMore)
            throw new Error(
              "The original record is unavailable in this incident. Its committed history is preserved.",
            );
          cursor = result.nextCursor;
          if (!cursor || visited.has(cursor))
            throw new Error("Record paging changed. Retry the read.");
          visited.add(cursor);
        } while (!currentSignal.aborted);
        throw new Error("Read interrupted.");
      }, signal);
    },
  };
}
