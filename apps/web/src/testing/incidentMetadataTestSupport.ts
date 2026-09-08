import type { GetIncidentResponse } from "@cartulary/protocol-ts/http";

export const metadataIncidentId = "00000000-0000-4000-8000-000000001001";
export const metadataActorId = "00000000-0000-4000-8000-000000000001";
export type MetadataIncident = GetIncidentResponse["data"];
export const metadataIncident = (
  overrides: Partial<MetadataIncident> = {},
): MetadataIncident => ({
  incident_id: metadataIncidentId,
  incident_key: "IR-METADATA",
  title: "Metadata investigation",
  description: "Original description",
  severity: "high",
  tlp: "TLP:AMBER",
  current_phase: "triage",
  primary_external_case_ref: "CASE-1",
  status: "active",
  incident_version: 1,
  created_at: "2026-08-01T00:00:00Z",
  created_by_user_id: metadataActorId,
  updated_at: "2026-08-01T00:00:00Z",
  updated_by_user_id: metadataActorId,
  closed_at: null,
  ...overrides,
});
export const metadataEnvelope = (data = metadataIncident()) => ({
  data,
  meta: { request_id: "metadata-test" },
});
export const metadataJSON = (payload: unknown, status = 200) =>
  new Response(JSON.stringify(payload), {
    status,
    headers: { "Content-Type": "application/json" },
  });
export function metadataDeferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
