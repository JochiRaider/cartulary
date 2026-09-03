export type WorkbookIncidentIdentity = {
  readonly closed_at?: string | null;
  readonly current_phase: string | null;
  readonly description: string | null;
  readonly incident_id: string;
  readonly incident_key: string;
  readonly incident_version: number;
  readonly primary_external_case_ref: string | null;
  readonly severity: string | null;
  readonly status?: "active" | "closed";
  readonly title: string;
  readonly tlp: string | null;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

export function normalizeIncidentIdentity(
  incidentId: string,
  value: unknown,
): WorkbookIncidentIdentity | null {
  if (!isRecord(value)) {
    return null;
  }
  const record = value;
  const incidentID =
    typeof record.incident_id === "string" ? record.incident_id : incidentId;
  if (
    typeof record.incident_key !== "string" ||
    typeof record.title !== "string"
  ) {
    return null;
  }
  return {
    closed_at: typeof record.closed_at === "string" ? record.closed_at : null,
    current_phase:
      typeof record.current_phase === "string" ? record.current_phase : null,
    description:
      typeof record.description === "string" ? record.description : null,
    incident_id: incidentID,
    incident_key: record.incident_key,
    incident_version:
      typeof record.incident_version === "number" ? record.incident_version : 0,
    primary_external_case_ref:
      typeof record.primary_external_case_ref === "string"
        ? record.primary_external_case_ref
        : null,
    severity: typeof record.severity === "string" ? record.severity : null,
    status: record.status === "closed" ? "closed" : "active",
    title: record.title,
    tlp: typeof record.tlp === "string" ? record.tlp : null,
  };
}
