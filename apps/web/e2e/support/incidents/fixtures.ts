import type { CreateIncidentRequest } from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

export async function createIncident(
  page: Page,
  incidentKey: string,
  title: string,
) {
  const body: CreateIncidentRequest = {
    client_txn_id: uniqueTxn("incident"),
    incident_key: incidentKey,
    title,
  };
  const response = await publicHttpOperation({
    body,
    headers: await csrfHeaders(page),
    operationID: "createIncident",
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(`create incident failed with HTTP ${response.status}`);
  }
  return response.payload.data.incident_id;
}
