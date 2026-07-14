import { useEffect, useState } from "react";
import {
  type NetworkFlowDiagnostic,
  queryNetworkFlowRejectedRows,
} from "./networkFlowClient";
import { isNetworkFlowAuthorizationLoss } from "./networkFlowErrors";

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
  readonly onError: (message: string | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const [diagnostics, setDiagnostics] = useState<NetworkFlowDiagnostic[]>([]);
  useEffect(() => {
    if (!enabled || activeTableId === null) {
      setDiagnostics([]);
      return;
    }
    const controller = new AbortController();
    void queryNetworkFlowRejectedRows({
      apiBase,
      incidentId,
      tableId: activeTableId,
      signal: controller.signal,
    })
      .then((result) => {
        if (!controller.signal.aborted) {
          setDiagnostics(result.diagnostics);
          onError(null);
        }
      })
      .catch((caught: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        const message =
          caught instanceof Error ? caught.message : "query_failed";
        if (isNetworkFlowAuthorizationLoss(message)) {
          onIncidentAccessLost?.();
        }
        onError(message);
      });
    return () => controller.abort();
  }, [
    activeTableId,
    apiBase,
    enabled,
    incidentId,
    onError,
    onIncidentAccessLost,
  ]);
  return {
    clearDiagnostics: () => setDiagnostics([]),
    diagnostics,
  };
}
