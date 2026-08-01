import { useEffect, useState } from "react";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";

export function useWorkbookIncidentIdentity({
  incidentPort,
  incidentId,
  initialIncidentIdentity,
  onIncidentAccessLost,
  onIncidentSnapshot,
}: {
  readonly incidentPort: WorkbookIncidentPort;
  readonly incidentId: string;
  readonly initialIncidentIdentity?: WorkbookIncidentIdentity | undefined;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly onIncidentSnapshot?:
    | ((incident: WorkbookIncidentIdentity) => void)
    | undefined;
}) {
  const [incidentIdentity, setIncidentIdentity] =
    useState<WorkbookIncidentIdentity | null>(
      () => initialIncidentIdentity ?? null,
    );
  const [incidentIdentityError, setIncidentIdentityError] = useState<
    string | null
  >(null);

  useEffect(() => {
    if (initialIncidentIdentity?.incident_id === incidentId) {
      setIncidentIdentity(initialIncidentIdentity);
      setIncidentIdentityError(null);
      onIncidentSnapshot?.(initialIncidentIdentity);
      return;
    }
    const controller = new AbortController();
    const loadIncidentIdentity = async () => {
      setIncidentIdentityError(null);
      const result = await incidentPort.getIdentity({
        signal: controller.signal,
      });
      if (controller.signal.aborted || result.kind === "aborted") {
        return;
      }
      if (result.kind === "rejected") {
        if (workbookOperationFailureIsAccessLoss(result.failure)) {
          onIncidentAccessLost?.();
        }
        setIncidentIdentityError(result.failure.message);
        return;
      }
      setIncidentIdentity(result.value);
      onIncidentSnapshot?.(result.value);
    };
    void loadIncidentIdentity();
    return () => {
      controller.abort();
    };
  }, [
    incidentPort,
    incidentId,
    initialIncidentIdentity,
    onIncidentAccessLost,
    onIncidentSnapshot,
  ]);

  return {
    incidentIdentity,
    incidentIdentityError,
  };
}
