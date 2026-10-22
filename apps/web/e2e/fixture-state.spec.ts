import type {
  HTTPOperationRequest,
  HTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { readCurrentSession } from "./support/auth/sessions";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { createSavedView, readSavedView } from "./support/workbook/savedViews";

type AccountPreferencesResponse =
  HTTPOperationResponse<"getCurrentAccountPreferences">;
type VisibleIncidentsResponse = HTTPOperationResponse<"listVisibleIncidents">;

// Each case is independently selectable. The two pairs exercise both orders
// within one worker without a reset between cases. Persisted state deliberately
// remains so the next case must obtain isolation from its actor and context.
async function assertOwnedState(page: Page, userID: string, actor: "A" | "B") {
  const session = await readCurrentSession(page);
  expect(session.user_id).toBe(userID);
  expect(session.memberships).toEqual([]);
  const before = await page.request.get(`${apiBase}/api/v1/incidents`);
  expect(before.ok()).toBeTruthy();
  expect(
    ((await before.json()) as VisibleIncidentsResponse).data.incidents,
  ).toEqual([]);

  const preferencesResponse = await page.request.get(
    `${apiBase}/api/v1/account/preferences`,
  );
  expect(preferencesResponse.ok()).toBeTruthy();
  const preferences = (
    (await preferencesResponse.json()) as AccountPreferencesResponse
  ).data;
  expect(preferences.user_id).toBe(userID);
  expect(preferences.density_mode).toBeNull();
  const body: HTTPOperationRequest<"putCurrentAccountPreferences"> = {
    base_preferences_version: preferences.preferences_version,
    client_txn_id: uniqueTxn("fixture-state-preferences"),
    density_mode: actor === "A" ? "compact" : "comfortable",
  };
  const updated = await page.request.put(
    `${apiBase}/api/v1/account/preferences`,
    {
      headers: await csrfHeaders(page),
      data: body,
    },
  );
  expect(updated.ok()).toBeTruthy();
  expect(
    ((await updated.json()) as AccountPreferencesResponse).data,
  ).toMatchObject({
    user_id: userID,
    density_mode: body.density_mode,
  });

  const incidentID = await createIncident(
    page,
    uniqueIncidentKey("FIXTURESTATE"),
    `Fixture actor ${actor}`,
  );
  const row = await createViewRow(page, incidentID, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("fixture-state-row"),
    "timeline.activity_synopsis_text": `Owned by ${actor}`,
  });
  const savedView = await createSavedView(page, incidentID, {
    display_name: "Fixture state boundary",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
  });
  expect(
    await readSavedView(page, incidentID, savedView.saved_view_id),
  ).toEqual(savedView);
  expect(
    (await queryViewRows(page, incidentID, timelineViewSchemaId)).map(
      (entry) => entry.record_id,
    ),
  ).toEqual([row.record_id]);
  const after = await page.request.get(`${apiBase}/api/v1/incidents`);
  expect(after.ok()).toBeTruthy();
  expect(
    ((await after.json()) as VisibleIncidentsResponse).data.incidents.map(
      (entry) => entry.incident_id,
    ),
  ).toEqual([incidentID]);
  expect(
    (await readCurrentSession(page)).memberships.map(
      (entry) => entry.incident_id,
    ),
  ).toEqual([incidentID]);

  await page.goto(`/?incident_id=${incidentID}`);
  expect(
    await page.evaluate(() => localStorage.getItem("fixture-state-owner")),
  ).toBeNull();
  await page.evaluate(
    (value) => localStorage.setItem("fixture-state-owner", value),
    actor,
  );
  expect(
    await page.evaluate(() => localStorage.getItem("fixture-state-owner")),
  ).toBe(actor);
}

test("browser fixture A owns its state before B", async ({
  page,
  testActor,
}) => {
  await assertOwnedState(page, testActor.user_id, "A");
});

test("browser fixture B owns its state after A", async ({
  page,
  testActor,
}) => {
  await assertOwnedState(page, testActor.user_id, "B");
});

test("browser fixture B owns its state before A", async ({
  page,
  testActor,
}) => {
  await assertOwnedState(page, testActor.user_id, "B");
});

test("browser fixture A owns its state after B", async ({
  page,
  testActor,
}) => {
  await assertOwnedState(page, testActor.user_id, "A");
});
