import { useEffect, useState } from "react";
import {
  type NetworkFlowRow,
  queryNetworkFlowTable,
} from "./networkFlowClient";
import { isNetworkFlowAuthorizationLoss } from "./networkFlowErrors";

export function useNetworkFlowRowsController({
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
  const [rows, setRows] = useState<NetworkFlowRow[]>([]);
  useEffect(() => {
    if (!enabled || activeTableId === null) {
      setRows([]);
      return;
    }
    const controller = new AbortController();
    void queryNetworkFlowTable({
      apiBase,
      incidentId,
      tableId: activeTableId,
      signal: controller.signal,
    })
      .then((result) => {
        if (!controller.signal.aborted) {
          setRows(result.rows);
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
  return { clearRows: () => setRows([]), rows };
}
