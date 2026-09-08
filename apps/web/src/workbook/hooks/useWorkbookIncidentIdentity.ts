import { useCallback, useEffect, useRef, useState } from "react";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";

export function useWorkbookIncidentIdentity({
  incidentPort,
  incidentId,
  initialIncidentIdentity,
  onIncidentAccessLost,
}: {
  readonly incidentPort: WorkbookIncidentPort;
  readonly incidentId: string;
  readonly initialIncidentIdentity?: WorkbookIncidentIdentity | undefined;
  readonly onIncidentAccessLost?: (() => void) | undefined;
}) {
  const [incidentIdentity, setIncidentIdentity] =
    useState<WorkbookIncidentIdentity | null>(
      () => initialIncidentIdentity ?? null,
    );
  const [incidentIdentityError, setIncidentIdentityError] = useState<
    string | null
  >(null);

  const currentIncident = useRef(incidentId);
  currentIncident.current = incidentId;
  const acceptIncidentResource = useCallback(
    (next: WorkbookIncidentIdentity) => {
      if (next.incident_id !== currentIncident.current) return;
      setIncidentIdentity((previous) =>
        previous?.incident_id === next.incident_id &&
        previous.incident_version > next.incident_version
          ? previous
          : next,
      );
    },
    [],
  );

  useEffect(() => {
    if (initialIncidentIdentity?.incident_id === incidentId) {
      acceptIncidentResource(initialIncidentIdentity);
      setIncidentIdentityError(null);
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
      acceptIncidentResource(result.value);
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
    acceptIncidentResource,
  ]);

  return {
    incidentIdentity,
    acceptIncidentResource,
    incidentIdentityError,
  };
}
