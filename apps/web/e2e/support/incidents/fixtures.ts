import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { atJsonOrigin, requestPublicJson } from "../transport/publicJsonClient";

export async function createIncident(
  page: Page,
  incidentKey: string,
  title: string,
) {
  const response = await requestPublicJson({
    body: {
      client_txn_id: uniqueTxn("incident"),
      incident_key: incidentKey,
      title,
    },
    headers: await csrfHeaders(page),
    method: "POST",
    path: "/api/v1/incidents",
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
  const body = response.body as { data: { incident_id: string } };
  return body.data.incident_id;
}
