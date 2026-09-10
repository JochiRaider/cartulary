import { useCallback, useMemo } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import { queryNetworkFlowTable } from "./networkFlowClient";
import {
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import {
  acceptedContinuationRequest,
  acceptedInitialRequest,
  type NetworkFlowAcceptedPageRequest,
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowRows,
} from "./networkFlowQueryModel";
import { useNetworkFlowPagedQuery } from "./useNetworkFlowPagedQuery";

export function useNetworkFlowRowsController({
  availability,
  activeTableId,
  apiBase,
  enabled,
  incidentId,
  onError,
  onIncidentAccessLost,
  query,
  revision,
  onQueryResult,
}: {
  readonly query: NetworkFlowAcceptedQuery;
  readonly revision: number;
  readonly onQueryResult: (
    revision: number,
    error: NetworkFlowRequestError | null,
  ) => void;
  readonly availability: ExtensionAvailabilityController;
  readonly activeTableId: string | null;
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const initialRequest = useMemo(() => acceptedInitialRequest(query), [query]);
  const fetchPage = useCallback(
    async (request: NetworkFlowAcceptedPageRequest, signal: AbortSignal) => {
      if (activeTableId === null) {
        throw new Error("network_flow_table_not_selected");
      }
      try {
        const result = await queryNetworkFlowTable({
          availability,
          apiBase,
          incidentId,
          tableId: activeTableId,
          request,
          signal,
        });
        if (!signal.aborted && !("cursor_token" in request))
          onQueryResult(revision, null);
        return { items: result.rows, paging: result.paging };
      } catch (error) {
        if (!signal.aborted && !("cursor_token" in request))
          onQueryResult(
            revision,
            networkFlowErrorFromUnknown(error, "Network Flow query failed."),
          );
        throw error;
      }
    },
    [activeTableId, apiBase, availability, incidentId, onQueryResult, revision],
  );
  const paged = useNetworkFlowPagedQuery({
    enabled: enabled && activeTableId !== null,
    fetchPage,
    initialRequest,
    isContinuation: (request) =>
      request.schema_id ===
      "cartulary.network_flow.table_query_continuation.v1",
    makeContinuation: acceptedContinuationRequest,
    onError,
    onIncidentAccessLost,
    queryKey: `${revision}:${incidentId}:${activeTableId ?? "none"}:${JSON.stringify(initialRequest)}`,
    reconcile: reconcileNetworkFlowRows,
  });
  return {
    ...paged,
    clearRows: paged.clear,
    query,
    rows: paged.items,
  };
}
