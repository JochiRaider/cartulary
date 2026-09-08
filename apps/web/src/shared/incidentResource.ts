import type { IncidentResource } from "../app/api/incidentResourceContracts";
import { normalizeAuditInstant } from "./auditReadValues";

export type { IncidentResource } from "../app/api/incidentResourceContracts";

/** Generated decoding checks shape; this boundary checks resource semantics. */
export function validIncidentResource(
  resource: IncidentResource,
  incidentId: string,
): boolean {
  return (
    !!resource &&
    typeof resource === "object" &&
    resource.incident_id === incidentId &&
    Number.isSafeInteger(resource.incident_version) &&
    resource.incident_version > 0 &&
    (resource.status === "active"
      ? resource.closed_at === null
      : resource.status === "closed" &&
        typeof resource.closed_at === "string" &&
        normalizeAuditInstant(resource.closed_at) !== null)
  );
}

export function incidentResourceOrder(
  previous: IncidentResource | null,
  next: IncidentResource,
  incidentId: string,
): "new" | "same" | "stale" | "invalid" {
  if (!validIncidentResource(next, incidentId)) return "invalid";
  if (!previous || previous.incident_id !== next.incident_id) return "new";
  if (next.incident_version < previous.incident_version) return "stale";
  if (next.incident_version > previous.incident_version) return "new";
  return (Object.keys(previous) as (keyof IncidentResource)[]).every(
    (key) => previous[key] === next[key],
  )
    ? "same"
    : "invalid";
}
