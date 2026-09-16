import {
  gridGroupingSelectTestId,
  incidentAdministrationTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  savedViewSetHomeButtonTestId,
  surfaceTabTestId,
  workbookPreferenceTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import {
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
} from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  createSavedView,
  openSavedViewActionMenu,
} from "./support/workbook/savedViews";

const timeline = { kind: "view_schema", id: timelineViewSchemaId };
const hosts = { kind: "view_schema", id: hostsViewSchemaId };
const field = (kind: "home" | "default") =>
  kind === "home" ? "home_sheet_ref" : "default_sheet_ref";
const routePath = (id: string, kind: "home" | "default") =>
  `/api/v1/incidents/${id}/workbook-preferences/${kind === "home" ? "me" : "default"}`;
async function read(page: Page, id: string, kind: "home" | "default") {
  const response = await page.request.get(`${apiBase}${routePath(id, kind)}`);
  expect(response.status()).toBe(200);
  return (await response.json()).data;
}
async function actPreference(
  page: Page,
  kind: "home" | "default",
  action: "set" | "clear",
) {
  const button = page.getByTestId(workbookPreferenceTestId(kind, action));
  await expect(button).toHaveAttribute("aria-disabled", "false");
  await button.click();
  await expect(
    page.getByTestId(workbookPreferenceTestId(kind, "outcome")),
  ).toHaveText(
    `${kind === "home" ? "Home" : "Incident default"} ${action === "clear" ? "clear" : "update"} confirmed.`,
  );
  await expect(
    page.getByTestId(workbookPreferenceTestId(kind, "refresh")),
  ).toHaveAttribute("aria-disabled", "false");
}
async function close(page: Page) {
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
}

test("preferences persist exact base and saved identities and explicit clears through subsequent startup fallback", async ({
  workerAdminPage: page,
}, testInfo) => {
  const id = await createIncident(
    page,
    uniqueIncidentKey("WP-FALLBACK"),
    "Preference fallback",
  );
  const saved = await createSavedView(page, id, {
    display_name: "Stored home",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
  });
  await openIncidentFromLanding(page, id);
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await openIncidentControls(page);
  await actPreference(page, "default", "set");
  expect((await read(page, id, "default")).default_sheet_ref).toEqual(hosts);
  await close(page);
  await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
  await page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)).click();
  await page
    .getByTestId(
      savedViewOptionTestId(timelineViewSchemaId, saved.saved_view_id),
    )
    .click();
  await page
    .getByTestId(gridGroupingSelectTestId(timelineViewSchemaId))
    .selectOption("timeline.capture_state");
  const url = page.url();
  const sideEffects: string[] = [];
  page.on("request", (request) => {
    if (
      request.url().includes("workbook-startup") ||
      (request.method() === "POST" && /saved-views|\/query/.test(request.url()))
    )
      sideEffects.push(request.url());
  });
  await openIncidentControls(page);
  await actPreference(page, "home", "set");
  const initial = await read(page, id, "home");
  expect(initial).toMatchObject({
    incident_id: id,
    home_sheet_ref: { kind: "saved_view", id: saved.saved_view_id },
    created_at: expect.any(String),
    updated_at: expect.any(String),
    user_id: expect.any(String),
  });
  await actPreference(page, "home", "set");
  expect(await read(page, id, "home")).toEqual(initial);
  expect(page.url()).toBe(url);
  expect(sideEffects).toEqual([]);
  await close(page);
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toHaveValue("timeline.capture_state");
  await openIncidentControls(page);
  await actPreference(page, "home", "clear");
  const cleared = await read(page, id, "home");
  expect(cleared.home_sheet_ref).toBeNull();
  await actPreference(page, "home", "clear");
  expect(await read(page, id, "home")).toEqual(cleared);
  expect(page.url()).toBe(url);
  await page.goto(`/?incident_id=${id}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toHaveAttribute(
    "data-active-view-schema-id",
    hostsViewSchemaId,
  );
  await openIncidentControls(page);
  await actPreference(page, "default", "clear");
  const defaultCleared = await read(page, id, "default");
  await actPreference(page, "default", "clear");
  expect(await read(page, id, "default")).toEqual(defaultCleared);
  expect(defaultCleared.default_sheet_ref).toBeNull();
  await page.goto(`/?incident_id=${id}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toHaveAttribute(
    "data-active-view-schema-id",
    timelineViewSchemaId,
  );
  await testInfo.attach("preference-resources", {
    body: JSON.stringify({ initial, cleared, defaultCleared }),
    contentType: "application/json",
  });
});

test("preferences retain uncertain committed writes until explicit observed-value review and separate confirmed refresh failure", async ({
  workerAdminPage: page,
}) => {
  const id = await createIncident(
    page,
    uniqueIncidentKey("WP-RECOVERY"),
    "Preference recovery",
  );
  await openIncidentFromLanding(page, id);
  await openIncidentControls(page);
  let writes = 0;
  let lose = true;
  let failRead = false;
  await page.route(`**${routePath(id, "home")}`, async (route) => {
    if (route.request().method() === "PUT") {
      ++writes;
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      if (lose) await route.abort("failed");
      else {
        failRead = true;
        await route.fulfill({ response });
      }
    } else if (failRead)
      await route.fulfill({
        status: 503,
        json: {
          error: { code: "internal_error" },
          meta: { request_id: "read-failure" },
        },
      });
    else await route.continue();
  });
  await page.getByTestId(workbookPreferenceTestId("home", "set")).click();
  const recovery = page.getByTestId(
    workbookPreferenceTestId("home", "recovery"),
  );
  await expect(recovery).toContainText(
    "Captured target: View schema: Timeline",
  );
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "keep-observed")),
  ).toHaveAttribute("aria-disabled", "false");
  expect(writes).toBe(1);
  expect((await read(page, id, "home")).home_sheet_ref).toEqual(timeline);
  await close(page);
  await page
    .getByRole("button", { name: "Account and application navigation" })
    .click();
  await page.getByRole("menuitem", { name: "Incidents", exact: true }).click();
  const departure = page.getByRole("dialog", {
    name: "Leave workbook preference recovery?",
  });
  await expect(departure).toBeVisible();
  await departure.getByRole("button", { name: "Stay", exact: true }).click();
  await expect(departure).toHaveCount(0);
  await openIncidentControls(page);
  await expect(recovery).toBeVisible();
  const intervening = await page.request.put(
    `${apiBase}${routePath(id, "home")}`,
    { headers: await csrfHeaders(page), data: { home_sheet_ref: hosts } },
  );
  expect(intervening.status()).toBe(200);
  await page.getByTestId(workbookPreferenceTestId("home", "refresh")).click();
  await expect(recovery).toContainText(
    "Current observation: View schema: Hosts",
  );
  await page
    .getByTestId(workbookPreferenceTestId("home", "keep-observed"))
    .click();
  await expect(recovery).toHaveCount(0);
  expect(writes).toBe(1);
  await page.getByTestId(workbookPreferenceTestId("home", "set")).click();
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "write-captured")),
  ).toHaveAttribute("aria-disabled", "false");
  lose = false;
  await page
    .getByTestId(workbookPreferenceTestId("home", "write-captured"))
    .click();
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "outcome")),
  ).toHaveText("Home update confirmed.");
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "read")),
  ).toContainText("may be stale");
  expect(writes).toBe(3);
  failRead = false;
  await page.getByTestId(workbookPreferenceTestId("home", "refresh")).click();
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "read")),
  ).toHaveText("");
  expect(writes).toBe(3);
});

test("preferences independently publish delayed reads and guard shortcut admission across changed surfaces", async ({
  workerAdminPage: page,
}) => {
  const id = await createIncident(
    page,
    uniqueIncidentKey("WP-TIMING"),
    "Independent preference reads",
  );
  await openIncidentFromLanding(page, id);
  let releaseDefault!: () => void;
  const defaultGate = new Promise<void>((resolve) => {
    releaseDefault = resolve;
  });
  await page.route(`**${routePath(id, "default")}`, async (route) => {
    await defaultGate;
    await route.continue();
  });
  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-title")),
  ).toHaveText("Independent preference reads");
  await expect(
    page.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref")),
  ).toHaveText("Unset");
  await expect(
    page.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref")),
  ).toHaveText("Loading…");
  await expect(
    page.getByRole("textbox", { name: "Reason", exact: true }),
  ).toBeEnabled();
  releaseDefault();
  let releaseWrite!: () => void;
  const writeGate = new Promise<void>((resolve) => {
    releaseWrite = resolve;
  });
  const requests: unknown[] = [];
  await page.route(`**${routePath(id, "home")}`, async (route) => {
    if (route.request().method() === "PUT") {
      requests.push(route.request().postDataJSON());
      const response = await route.fetch();
      await writeGate;
      await route.fulfill({ response });
    } else await route.continue();
  });
  await page.getByTestId(workbookPreferenceTestId("home", "set")).click();
  await expect.poll(() => requests.length).toBe(1);
  await close(page);
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await openSavedViewActionMenu(page, hostsViewSchemaId);
  const shortcut = page.getByTestId(
    savedViewSetHomeButtonTestId(hostsViewSchemaId),
  );
  await expect(shortcut).toHaveAttribute("aria-disabled", "true");
  await shortcut.dispatchEvent("click");
  await page.keyboard.press("Escape");
  releaseWrite();
  await openSavedViewActionMenu(page, hostsViewSchemaId);
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "shortcut-outcome")),
  ).toHaveText("Home update confirmed.");
  expect(requests).toEqual([{ home_sheet_ref: timeline }]);
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(workbookShellReadyTestId())).toHaveAttribute(
    "data-active-view-schema-id",
    hostsViewSchemaId,
  );
  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref")),
  ).toContainText(timelineViewSchemaId);
});

test("preferences preserve personal isolation and viewer boundaries on a closed incident", async ({
  workerAdminPage: page,
  browser,
  sessionTracker,
}) => {
  const id = await createIncident(
    page,
    uniqueIncidentKey("WP-AUTH"),
    "Closed preference permissions",
  );
  const viewer = await createIncidentMemberUser(page, id, {
    email: uniqueEmail("wp-viewer"),
    display_name: "Preference viewer",
    initial_password: "PreferenceViewer1!",
    role: "viewer",
    mfa_required: false,
    is_deployment_admin: false,
  });
  const other = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "preference isolation",
    purpose: "viewer preference authority",
    email: viewer.email,
    password: viewer.initial_password,
    incidentId: id,
    userId: viewer.user_id,
  });
  try {
    const current = await currentLifecycle(page, id);
    expect(
      (
        await lifecycleAction(page, id, "closeIncident", {
          base_incident_version: current.incident_version,
          client_txn_id: uniqueTxn("wp-close"),
          reason: "Preference authorization on closed incident",
        })
      ).ok,
    ).toBe(true);
    await openIncidentFromLanding(page, id);
    await openIncidentControls(page);
    await actPreference(page, "default", "set");
    await actPreference(page, "default", "clear");
    await other.reload();
    await openIncidentControls(other);
    await actPreference(other, "home", "set");
    expect((await read(other, id, "home")).user_id).toBe(viewer.user_id);
    expect((await read(page, id, "home")).home_sheet_ref).toBeNull();
    await expect(
      other.getByTestId(workbookPreferenceTestId("default", "clear")),
    ).toHaveAttribute("aria-disabled", "true");
    const denied = await other.request.put(
      `${apiBase}${routePath(id, "default")}`,
      { headers: await csrfHeaders(other), data: { default_sheet_ref: null } },
    );
    expect(denied.status()).toBe(403);
    await actPreference(other, "home", "clear");
    expect((await read(other, id, "home"))[field("home")]).toBeNull();
  } finally {
    await other.context().close();
  }
});
