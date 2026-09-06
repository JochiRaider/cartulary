import type {
  CreateIncidentRequest,
  CreateIncidentResponse,
  ListVisibleIncidentsResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";

export function listVisibleIncidents(
  options: {
    apiBase?: string | undefined;
    search?: string | undefined;
    status?:
      | ListVisibleIncidentsResponse["data"]["incidents"][number]["status"]
      | undefined;
    limit?: number | undefined;
    cursorToken?: string | null | undefined;
    signal?: AbortSignal | undefined;
  } = {},
) {
  return fetchHTTPOperation<ListVisibleIncidentsResponse>({
    apiBase: options.apiBase,
    operationID: "listVisibleIncidents",
    init: options.signal === undefined ? undefined : { signal: options.signal },
    query: {
      limit: options.limit ?? 100,
      ...(options.search === undefined || options.search === ""
        ? {}
        : { search: options.search }),
      ...(options.status === undefined ? {} : { status: options.status }),
      ...(options.cursorToken == null
        ? {}
        : { cursor_token: options.cursorToken }),
    },
  });
}

export async function createIncident(options: {
  request: Readonly<CreateIncidentRequest>;
  signal: AbortSignal;
}) {
  const {
    client_txn_id,
    incident_key,
    title,
    description,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
  } = options.request;
  const request = {
    client_txn_id,
    incident_key,
    title,
    ...(description === undefined ? {} : { description }),
    ...(severity === undefined ? {} : { severity }),
    ...(tlp === undefined ? {} : { tlp }),
    ...(current_phase === undefined ? {} : { current_phase }),
    ...(primary_external_case_ref === undefined
      ? {}
      : { primary_external_case_ref }),
  } satisfies CreateIncidentRequest;
  const result = await fetchHTTPOperation<CreateIncidentResponse>({
    operationID: "createIncident",
    init: {
      method: "POST",
      body: JSON.stringify(request),
      signal: options.signal,
    },
  });
  if (result.ok && result.status !== 200 && result.status !== 201) {
    return {
      ok: false as const,
      status: 502,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          status: 502,
          retryable: true,
        },
      },
    };
  }
  return result;
}
