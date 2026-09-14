import { useCallback, useEffect, useRef, useState } from "react";
import {
  type IncidentResource,
  validIncidentResource,
} from "../../shared/incidentResource";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";

export function useWorkbookIncidentIdentity({
  incidentPort,
  incidentId,
  initialIncidentIdentity,
  acceptedIncidentResource,
  onIncidentResourceObserved,
  onAuthorityUncertain,
}: {
  readonly incidentPort: WorkbookIncidentPort;
  readonly acceptedIncidentResource?: IncidentResource | null | undefined;
  readonly onIncidentResourceObserved?:
    | ((resource: IncidentResource) => void)
    | undefined;
  readonly incidentId: string;
  readonly initialIncidentIdentity?: WorkbookIncidentIdentity | undefined;
  readonly onAuthorityUncertain?: (() => void) | undefined;
}) {
  const [incidentIdentity, setIncidentIdentity] =
    useState<WorkbookIncidentIdentity | null>(
      () => initialIncidentIdentity ?? null,
    );
  const [incidentIdentityError, setIncidentIdentityError] = useState<
    string | null
  >(null);

  const observed = useRef(onIncidentResourceObserved);
  observed.current = onIncidentResourceObserved;
  const currentIncident = useRef(incidentId);
  currentIncident.current = incidentId;
  const acceptIncidentResource = useCallback(
    (next: WorkbookIncidentIdentity) => {
      if (next.incident_id !== currentIncident.current) return;
      setIncidentIdentity((previous) =>
        previous?.incident_id === next.incident_id &&
        previous.incident_version >= next.incident_version
          ? previous
          : next,
      );
    },
    [],
  );

  useEffect(() => {
    if (
      acceptedIncidentResource &&
      validIncidentResource(acceptedIncidentResource, incidentId)
    )
      acceptIncidentResource(acceptedIncidentResource);
  }, [acceptedIncidentResource, incidentId, acceptIncidentResource]);

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
      if (
        controller.signal.aborted ||
        currentIncident.current !== incidentId ||
        result.kind === "aborted"
      ) {
        return;
      }
      if (result.kind === "rejected") {
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        ) {
          onAuthorityUncertain?.();
        }
        setIncidentIdentityError(result.failure.message);
        return;
      }
      acceptIncidentResource(result.value);
      if (result.value.resource) observed.current?.(result.value.resource);
    };
    void loadIncidentIdentity();
    return () => {
      controller.abort();
    };
  }, [
    incidentPort,
    incidentId,
    initialIncidentIdentity,
    onAuthorityUncertain,
    acceptIncidentResource,
  ]);

  return {
    incidentIdentity,
    acceptIncidentResource,
    incidentIdentityError,
  };
}
