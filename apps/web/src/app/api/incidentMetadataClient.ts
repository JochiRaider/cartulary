import type {
  GetIncidentResponse,
  PatchIncidentRequest,
  PatchIncidentResponse,
} from "@cartulary/protocol-ts/http";
import { type APIError, fetchHTTPOperation } from "../../services/browserApi";
import {
  type IncidentMetadataResource,
  type MetadataAuthority,
  type MetadataProblem,
  metadataFields,
  metadataProblemCodes,
} from "../incidentMetadataModel";

export type MetadataResult =
  | { readonly ok: true; readonly resource: IncidentMetadataResource }
  | {
      readonly ok: false;
      readonly status: number;
      readonly problem: MetadataProblem;
    };
const invalid = (): MetadataResult => ({
  ok: false,
  status: 502,
  problem: { code: "invalid_public_contract_response" },
});
function failure(status: number, error?: APIError): MetadataResult {
  return {
    ok: false,
    status,
    problem: {
      code:
        metadataProblemCodes.find((code) => code === error?.code) ??
        "unknown_public_error",
      field: metadataFields.find((field) => field === error?.details?.field),
      reason: (
        [
          "field_too_long",
          "control_character_not_allowed",
          "invalid_value",
        ] as const
      ).find((reason) => reason === error?.details?.reason_code),
    },
  };
}
export async function readIncidentMetadata(options: {
  authority: MetadataAuthority;
  signal: AbortSignal;
}): Promise<MetadataResult> {
  const result = await fetchHTTPOperation<GetIncidentResponse>({
    apiBase: options.authority.apiBase,
    operationID: "getIncident",
    pathParameters: { incident_id: options.authority.incidentId },
    init: { signal: options.signal },
  });
  if (!result.ok) return failure(result.status, result.payload.error);
  const resource = result.payload.data;
  return result.status === 200 &&
    resource.incident_id === options.authority.incidentId &&
    Number.isSafeInteger(resource.incident_version) &&
    resource.incident_version > 0
    ? { ok: true, resource }
    : invalid();
}
export async function patchIncidentMetadata(options: {
  authority: MetadataAuthority;
  payload: Readonly<PatchIncidentRequest>;
  signal: AbortSignal;
}): Promise<MetadataResult> {
  const {
    base_incident_version,
    description,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
  } = options.payload;
  const payload = {
    base_incident_version,
    ...(description === undefined ? {} : { description }),
    ...(severity === undefined ? {} : { severity }),
    ...(tlp === undefined ? {} : { tlp }),
    ...(current_phase === undefined ? {} : { current_phase }),
    ...(primary_external_case_ref === undefined
      ? {}
      : { primary_external_case_ref }),
  } satisfies PatchIncidentRequest;
  const result = await fetchHTTPOperation<PatchIncidentResponse>({
    apiBase: options.authority.apiBase,
    operationID: "patchIncident",
    pathParameters: { incident_id: options.authority.incidentId },
    init: {
      method: "PATCH",
      body: JSON.stringify(payload),
      signal: options.signal,
    },
  });
  if (!result.ok) return failure(result.status, result.payload.error);
  const resource = result.payload.data;
  return result.status === 200 &&
    resource.incident_id === options.authority.incidentId &&
    Number.isSafeInteger(resource.incident_version) &&
    resource.incident_version >= base_incident_version
    ? { ok: true, resource }
    : invalid();
}
