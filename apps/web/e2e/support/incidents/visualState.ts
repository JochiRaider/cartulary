import type { Page } from "@playwright/test";
import { expect } from "../../fixtures";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
export async function closeIncidentForVisual(page: Page, incidentId: string) {
  const incidentResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}`,
  );
  expect(incidentResponse.ok()).toBeTruthy();
  const incident = (await incidentResponse.json()) as {
    data: { incident_version: number };
  };
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/close`,
    {
      data: {
        base_incident_version: incident.data.incident_version,
        client_txn_id: uniqueTxn("visual-grid-state-close-incident"),
        reason: "Visual read-only state evidence",
      },
      headers: await csrfHeaders(page),
    },
  );
  expect(response.ok()).toBeTruthy();
}
