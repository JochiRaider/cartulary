import type { GetIncidentResponse } from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import { validIncidentResource } from "../../shared/incidentResource";
/** module.incidents retrieval; acceptance belongs to the neutral incident resource owner. */
export async function readIncidentSummary(
  incidentId: string,
  apiBase: string | undefined,
  signal: AbortSignal,
) {
  const result = await fetchHTTPOperation<GetIncidentResponse>({
    operationID: "getIncident",
    apiBase,
    pathParameters: { incident_id: incidentId },
    init: { signal },
  });
  if (!result.ok)
    return {
      ok: false as const,
      status: result.status,
      code:
        result.payload.error?.code === "incident_not_found"
          ? "incident_not_found"
          : result.payload.error?.code === "authorization_denied"
            ? "authorization_denied"
            : "incident_summary_unavailable",
    };
  return result.status === 200 &&
    validIncidentResource(result.payload.data, incidentId)
    ? { ok: true as const, resource: result.payload.data }
    : {
        ok: false as const,
        status: 502,
        code: "invalid_public_contract_response",
      };
}
