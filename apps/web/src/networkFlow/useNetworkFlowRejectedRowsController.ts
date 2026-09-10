import { useCallback, useMemo } from "react";
import {
  type ExtensionAvailabilityController,
  ExtensionAvailabilityUnavailableError,
} from "../extensions/extensionAvailability";
import { validateNetworkFlowPageContinuation } from "../services/networkFlowContractAdapter";
import { queryNetworkFlowRejectedRows } from "./networkFlowClient";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
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
  readIdentity,
  isCurrentRead,
  onProtectedStateLoss,
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
  readonly readIdentity: string | null;
  readonly isCurrentRead: () => boolean;
  readonly onProtectedStateLoss: (error: NetworkFlowRequestError) => void;
}) {
  const initialRequest = useMemo(() => rejectedInitialRequest(query), [query]);
  const fetchPage = useCallback(
    async (request: NetworkFlowRejectedPageRequest, signal: AbortSignal) => {
      if (activeTableId === null)
        throw new Error("network_flow_table_not_selected");
      const result = await queryNetworkFlowRejectedRows({
        availability,
        apiBase,
        incidentId,
        tableId: activeTableId,
        request,
        signal,
        authorizeDispatch: () => {
          if (!isCurrentRead())
            throw new ExtensionAvailabilityUnavailableError();
        },
      });
      return {
        items: result.diagnostics,
        paging: result.paging,
        metadata: result.metadata,
      };
    },
    [activeTableId, apiBase, availability, incidentId, isCurrentRead],
  );
  const paged = useNetworkFlowPagedQuery({
    enabled: enabled && activeTableId !== null && readIdentity !== null,
    fetchPage,
    initialRequest,
    isContinuation: (request) => "cursor_token" in request,
    makeContinuation: rejectedContinuationRequest,
    onError,
    onIncidentAccessLost,
    onProtectedStateLoss,
    onQueryResult: (error) => onQueryResult(revision, error),
    readIdentity,
    isCurrent: isCurrentRead,
    queryKey: JSON.stringify([
      apiBase,
      revision,
      incidentId,
      activeTableId,
      initialRequest,
    ]),
    validatePage: (page, previous, request) => {
      if ("cursor_token" in request && previous)
        validateNetworkFlowPageContinuation(
          page.metadata,
          previous.metadata,
          page.paging.limit,
          previous.paging.limit,
        );
    },
    reconcile: reconcileNetworkFlowDiagnostics,
  });
  return {
    ...paged,
    clearDiagnostics: paged.clear,
    query,
    diagnostics: paged.items,
  };
}
