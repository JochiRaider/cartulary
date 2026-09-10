import { useCallback, useMemo } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import { queryNetworkFlowRejectedRows } from "./networkFlowClient";
import {
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import {
  type NetworkFlowRejectedPageRequest,
  type NetworkFlowRejectedQuery,
  reconcileNetworkFlowDiagnostics,
  rejectedContinuationRequest,
  rejectedInitialRequest,
} from "./networkFlowQueryModel";
import { useNetworkFlowPagedQuery } from "./useNetworkFlowPagedQuery";

export function useNetworkFlowRejectedRowsController({
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
  readonly query: NetworkFlowRejectedQuery;
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
  const initialRequest = useMemo(() => rejectedInitialRequest(query), [query]);
  const fetchPage = useCallback(
    async (request: NetworkFlowRejectedPageRequest, signal: AbortSignal) => {
      if (activeTableId === null) {
        throw new Error("network_flow_table_not_selected");
      }
      try {
        const result = await queryNetworkFlowRejectedRows({
          availability,
          apiBase,
          incidentId,
          tableId: activeTableId,
          request,
          signal,
        });
        if (!signal.aborted && !("cursor_token" in request))
          onQueryResult(revision, null);
        return { items: result.diagnostics, paging: result.paging };
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
      "cartulary.network_flow.rejected_rows_query_continuation.v1",
    makeContinuation: rejectedContinuationRequest,
    onError,
    onIncidentAccessLost,
    queryKey: `${revision}:${incidentId}:${activeTableId ?? "none"}:${JSON.stringify(initialRequest)}`,
    reconcile: reconcileNetworkFlowDiagnostics,
  });
  return {
    ...paged,
    clearDiagnostics: paged.clear,
    diagnostics: paged.items,
    query,
  };
}
