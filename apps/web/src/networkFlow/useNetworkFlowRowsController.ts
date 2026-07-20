import { useCallback, useMemo, useState } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import { queryNetworkFlowTable } from "./networkFlowClient";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  acceptedContinuationRequest,
  acceptedInitialRequest,
  emptyNetworkFlowAcceptedQuery,
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
}: {
  readonly availability: ExtensionAvailabilityController;
  readonly activeTableId: string | null;
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const [query, setQuery] = useState<NetworkFlowAcceptedQuery>(
    emptyNetworkFlowAcceptedQuery,
  );
  const initialRequest = useMemo(() => acceptedInitialRequest(query), [query]);
  const fetchPage = useCallback(
    async (request: NetworkFlowAcceptedPageRequest, signal: AbortSignal) => {
      if (activeTableId === null) {
        throw new Error("network_flow_table_not_selected");
      }
      const result = await queryNetworkFlowTable({
        availability,
        apiBase,
        incidentId,
        tableId: activeTableId,
        request,
        signal,
      });
      return { items: result.rows, paging: result.paging };
    },
    [activeTableId, apiBase, availability, incidentId],
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
    queryKey: `${activeTableId ?? "none"}:${JSON.stringify(initialRequest)}`,
    reconcile: reconcileNetworkFlowRows,
  });
  return {
    ...paged,
    clearRows: paged.clear,
    query,
    rows: paged.items,
    setQuery,
  };
}
