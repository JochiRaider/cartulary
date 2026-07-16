import { useCallback, useMemo, useState } from "react";
import { queryNetworkFlowRejectedRows } from "./networkFlowClient";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  emptyNetworkFlowRejectedQuery,
  type NetworkFlowRejectedPageRequest,
  type NetworkFlowRejectedQuery,
  reconcileNetworkFlowDiagnostics,
  rejectedContinuationRequest,
  rejectedInitialRequest,
} from "./networkFlowQueryModel";
import { useNetworkFlowPagedQuery } from "./useNetworkFlowPagedQuery";

export function useNetworkFlowRejectedRowsController({
  activeTableId,
  apiBase,
  enabled,
  incidentId,
  onError,
  onIncidentAccessLost,
}: {
  readonly activeTableId: string | null;
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const [query, setQuery] = useState<NetworkFlowRejectedQuery>(
    emptyNetworkFlowRejectedQuery,
  );
  const initialRequest = useMemo(() => rejectedInitialRequest(query), [query]);
  const fetchPage = useCallback(
    async (request: NetworkFlowRejectedPageRequest, signal: AbortSignal) => {
      if (activeTableId === null) {
        throw new Error("network_flow_table_not_selected");
      }
      const result = await queryNetworkFlowRejectedRows({
        apiBase,
        incidentId,
        tableId: activeTableId,
        request,
        signal,
      });
      return { items: result.diagnostics, paging: result.paging };
    },
    [activeTableId, apiBase, incidentId],
  );
  const paged = useNetworkFlowPagedQuery({
    enabled: enabled && activeTableId !== null,
    fetchPage,
    initialRequest,
    isContinuation: (request) =>
      request.schema_id ===
      "cartulary.network_flow.rejected_rows_query_continuation.v1",
    makeContinuation: rejectedContinuationRequest,
    onError,
    onIncidentAccessLost,
    queryKey: `${activeTableId ?? "none"}:${JSON.stringify(initialRequest)}`,
    reconcile: reconcileNetworkFlowDiagnostics,
  });
  return {
    ...paged,
    clearDiagnostics: paged.clear,
    diagnostics: paged.items,
    query,
    setQuery,
  };
}
