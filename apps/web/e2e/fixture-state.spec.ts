import type {
  HTTPOperationRequest,
  HTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { type Page, request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  csrfHeaders,
  loginLocalAPIContext,
} from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import {
  loadUser,
  loginFixtureActorContext,
  readCurrentSession,
  verifySessionUnauthorized,
} from "./support/auth/sessions";
import { authHeadersForStorageState } from "./support/auth/storageState";
import { loadWorkerAdminManifest } from "./support/auth/workerAdmin";
import {
  createAndUploadObjectBlob,
  verifyUploadedObjectBlob,
} from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { createSavedView, readSavedView } from "./support/workbook/savedViews";

test.use({ locale: "fr-FR", timezoneId: "Pacific/Honolulu" });

test.beforeEach(async ({ page, browserScope, sessionTracker }) => {
  // Real unrelated records precede both orderings, including their first case.
  // They remain durable until the cohort reset; their live resources are owned.
  const password = "UnrelatedFixture1!";
  const actor = await createDeploymentUser(page, {
    email: uniqueEmail("unrelated-fixture"),
    initial_password: password,
    display_name: "Unrelated fixture actor",
    is_deployment_admin: true,
    mfa_required: false,
  });
  const context = await browserScope.newContext();
  const unrelated = await context.newPage();
  await sessionTracker.loginTrackedUser(unrelated, {
    createdBy: "fixture-state",
    email: actor.email,
    password,
    purpose: "unrelated durable residue",
    userId: actor.user_id,
  });
  const incidentId = await createIncident(
    unrelated,
    uniqueIncidentKey("UNRELATED"),
    "Unrelated durable incident",
  );
  const blob = await createAndUploadObjectBlob(unrelated, {
    incidentId,
    clientTxnId: uniqueTxn("unrelated-upload"),
    filename: "unrelated.txt",
    contentType: "text/plain",
    body: Buffer.from("Unrelated job and audit residue"),
  });
  await verifyUploadedObjectBlob(
    unrelated,
    blob.object_blob_id,
    uniqueTxn("unrelated-verify"),
  );
  await browserScope.closeContext(context);
});

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
  // Deployment administrators may see unrelated durable incidents. A fresh
  // actor's empty membership and preferences, rather than global cardinality,
  // establish its initial state.

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
    ((await after.json()) as VisibleIncidentsResponse).data.incidents
      .filter((entry) => entry.incident_id === incidentID)
      .map((entry) => entry.incident_id),
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
  return incidentID;
}

test("browser fixture A owns its state before B", async ({
  page,
  testActor,
  testActorRequest,
  fixtureProvisioner,
  browserScope,
}) => {
  const incidentId = await assertOwnedState(page, testActor.user_id, "A");
  const secondary = await browserScope.newContext({
    storageState: testActor.storageState,
  });
  const secondaryPage = await secondary.newPage();
  await secondaryPage.goto("/");
  expect(secondaryPage.viewportSize()).toEqual(page.viewportSize());
  expect(
    await secondaryPage.evaluate(() => ({
      language: navigator.language,
      zone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    })),
  ).toEqual({ language: "fr-FR", zone: "Pacific/Honolulu" });
  const overridden = await browserScope.newContext({
    viewport: { width: 700, height: 500 },
  });
  expect((await overridden.newPage()).viewportSize()).toEqual({
    width: 700,
    height: 500,
  });

  // A successful login followed by failure to capture its snapshot is still owned.
  const email = `uncaptured-${testActor.user_id}@example.test`;
  const password = "UncapturedFixture1!";
  await createDeploymentUser(page, {
    email,
    initial_password: password,
    display_name: "Uncaptured actor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const uncapturedClient = await browserScope.ownRequestContext(() =>
    request.newContext({ baseURL: apiBase }),
  );
  await expect(
    (async () => {
      const login = await loginLocalAPIContext(uncapturedClient, {
        email,
        password,
      });
      expect(login.ok).toBe(true);
      throw new Error("Injected failure before session snapshot registration");
    })(),
  ).rejects.toThrow("Injected failure before session snapshot registration");
  const uncapturedState = await uncapturedClient.storageState();

  // This additional session belongs to the existing worker control-plane actor.
  // The worker fixture remains its owner, independently of the test scope.
  const worker = loadWorkerAdminManifest().worker_admins[0];
  if (!worker) throw new Error("Missing control-plane actor");
  const borrowed = await loginFixtureActorContext(worker);

  const actorBefore = await loadUser(fixtureProvisioner, testActor.user_id);
  const path = `/api/v1/users/${testActor.user_id}`;
  const held = await holdBrowserRequest(page, { method: "PATCH", path });
  const headers = await csrfHeaders(page);
  const mutation = page.evaluate(
    async ({ url, headers, version }) => {
      try {
        await fetch(url, {
          method: "PATCH",
          credentials: "include",
          headers: { ...headers, "content-type": "application/json" },
          body: JSON.stringify({
            base_user_version: version,
            display_name: "This delayed mutation must not be dispatched",
          }),
        });
        return "dispatched";
      } catch {
        return "cancelled";
      }
    },
    { url: `${apiBase}${path}`, headers, version: actorBefore.user_version },
  );
  await held.waitForHit;
  const cancellation = held.dispose();
  expect(held.dispose()).toBe(cancellation);
  await cancellation;
  expect(await mutation).toBe("cancelled");
  expect(await loadUser(fixtureProvisioner, testActor.user_id)).toEqual(
    actorBefore,
  );
  expect((await testActorRequest.get("/api/v1/auth/session")).status()).toBe(
    200,
  );

  // A real Job remains owned even if the scenario fails after admission. Early
  // context closure must settle it while its authority and request client live.
  const blob = await createAndUploadObjectBlob(secondaryPage, {
    incidentId,
    clientTxnId: uniqueTxn("owned-job-slot"),
    filename: "owned-job.txt",
    contentType: "text/plain",
    body: Buffer.from("Owned feature job lifecycle witness"),
  });
  const admittedJob = await secondaryPage.evaluate(
    async ({ url, headers, clientTxnId }) => {
      const response = await fetch(url, {
        method: "POST",
        credentials: "include",
        headers: { ...headers, "content-type": "application/json" },
        body: JSON.stringify({ client_txn_id: clientTxnId }),
      });
      if (response.status !== 202)
        throw new Error("Expected real verification admission");
      return (await response.json()).data.job_id as string;
    },
    {
      url: `${apiBase}/api/v1/object-blobs/${blob.object_blob_id}/verify`,
      headers,
      clientTxnId: uniqueTxn("owned-job-verify"),
    },
  );
  await expect(
    (async () => {
      throw new Error("Injected failure after job admission");
    })(),
  ).rejects.toThrow("Injected failure after job admission");
  await browserScope.closeContext(secondary);
  expect(secondaryPage.isClosed()).toBe(true);
  const terminalJob = await testActorRequest.get(`/api/v1/jobs/${admittedJob}`);
  expect(terminalJob.status()).toBe(200);
  expect((await terminalJob.json()).data.status).toBe("succeeded");

  const completion = browserScope.dispose();
  expect(browserScope.dispose()).toBe(completion);
  await completion;
  expect(page.isClosed()).toBe(true);
  expect(secondaryPage.isClosed()).toBe(true);
  await verifySessionUnauthorized(
    testActor.storageState,
    "owned actor",
    fixtureProvisioner,
  );
  await verifySessionUnauthorized(
    uncapturedState,
    "snapshot capture failure",
    fixtureProvisioner,
  );
  expect((await fixtureProvisioner.get("/api/v1/auth/session")).status()).toBe(
    200,
  );
  expect(
    (
      await fixtureProvisioner.get("/api/v1/auth/session", {
        headers: authHeadersForStorageState(borrowed.storageState),
      })
    ).status(),
  ).toBe(200);
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
