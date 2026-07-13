import { useEffect, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  handleWorkbookLoadFailure,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import {
  normalizeIncidentIdentity,
  type WorkbookIncidentIdentity,
} from "../models/workbookIncidentIdentity";

type IncidentIdentityEnvelope = {
  data: WorkbookIncidentIdentity;
};

export function useWorkbookIncidentIdentity({
  apiBase,
  incidentId,
  initialIncidentIdentity,
  onIncidentAccessLost,
  onIncidentSnapshot,
}: {
  readonly apiBase?: string | undefined;
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
    let cancelled = false;
    const loadIncidentIdentity = async () => {
      setIncidentIdentityError(null);
      const result = await fetchWorkbookJSON<IncidentIdentityEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}`),
      );
      if (cancelled) {
        return;
      }
      if (!result.ok) {
        const message = handleWorkbookLoadFailure(
          parseErrorMessage(result.payload),
          "Incident identity load failed.",
          onIncidentAccessLost,
        );
        setIncidentIdentityError(message);
        return;
      }
      const envelope = readEnvelope<IncidentIdentityEnvelope>(result.payload);
      const normalized = normalizeIncidentIdentity(incidentId, envelope.data);
      if (normalized === null) {
        setIncidentIdentityError("Incident identity load failed.");
        return;
      }
      setIncidentIdentity(normalized);
      onIncidentSnapshot?.(normalized);
    };
    void loadIncidentIdentity();
    return () => {
      cancelled = true;
    };
  }, [
    apiBase,
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
