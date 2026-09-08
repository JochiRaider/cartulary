import type {
  CloseIncidentResponse,
  GetIncidentResponse,
} from "@cartulary/protocol-ts/http";
import { type APIError, fetchHTTPOperation } from "../../services/browserApi";
import {
  type IncidentResource,
  validIncidentResource,
} from "../../shared/incidentResource";
import {
  type LifecycleAttempt,
  type LifecycleAuthority,
  type LifecycleProblem,
  lifecycleProblemCodes,
  lifecycleReasons,
} from "../incidentLifecycleModel";

export type LifecycleResult =
  | { readonly ok: true; readonly resource: IncidentResource }
  | {
      readonly ok: false;
      readonly status: number;
      readonly problem: LifecycleProblem;
    };
const invalid = (): LifecycleResult => ({
  ok: false,
  status: 502,
  problem: { code: "invalid_public_contract_response" },
});
function failure(status: number, error?: APIError): LifecycleResult {
  return {
    ok: false,
    status,
    problem: {
      code:
        lifecycleProblemCodes.find((code) => code === error?.code) ??
        "unknown_public_error",
      field: (
        ["reason", "base_incident_version", "client_txn_id"] as const
      ).find((field) => field === error?.details?.field),
      reason: lifecycleReasons.find(
        (reason) => reason === error?.details?.reason_code,
      ),
    },
  };
}
export async function readLifecycleIncident(options: {
  authority: LifecycleAuthority;
  signal: AbortSignal;
}): Promise<LifecycleResult> {
  const result = await fetchHTTPOperation<GetIncidentResponse>({
    apiBase: options.authority.apiBase,
    operationID: "getIncident",
    pathParameters: { incident_id: options.authority.incidentId },
    init: { signal: options.signal },
  });
  if (!result.ok) return failure(result.status, result.payload.error);
  return result.status === 200 &&
    validIncidentResource(result.payload.data, options.authority.incidentId)
    ? { ok: true, resource: result.payload.data }
    : invalid();
}
export async function mutateIncidentLifecycle(options: {
  attempt: LifecycleAttempt;
  signal: AbortSignal;
}): Promise<LifecycleResult> {
  const { attempt } = options;
  const { base_incident_version, client_txn_id, reason } = attempt.payload;
  const result = await fetchHTTPOperation<CloseIncidentResponse>({
    apiBase: attempt.authority.apiBase,
    operationID:
      attempt.action === "close" ? "closeIncident" : "reopenIncident",
    pathParameters: { incident_id: attempt.authority.incidentId },
    init: {
      method: "POST",
      signal: options.signal,
      body: JSON.stringify({ base_incident_version, client_txn_id, reason }),
    },
  });
  if (!result.ok) return failure(result.status, result.payload.error);
  const resource = result.payload.data;
  return result.status === 200 &&
    validIncidentResource(resource, attempt.authority.incidentId) &&
    resource.incident_version === base_incident_version + 1 &&
    resource.status === (attempt.action === "close" ? "closed" : "active")
    ? { ok: true, resource }
    : invalid();
}
