import type { GetCurrentSessionResponse } from "@cartulary/protocol-ts/http";
import {
  incidentLandingTestId,
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackJobStatusTestId,
  referencePackListStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRowTestId,
} from "@cartulary/ui-contracts";
import type { Route } from "@playwright/test";

import { expect, test } from "./fixtures";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";

const jobID = "11111111-1111-4111-8111-111111111111";

test("shows Reference Pack progress and cancel controls without blocking landing interaction", async ({
  page,
}) => {
  let jobReads = 0;
  let cancelRequests = 0;
  let demoted = false;
  let enteredSearchURL = "";
  const appendGate = gate();
  const identityGate = gate();
  const newestGate = gate();
  const staleErrorGate = gate();
  const latestErrorGate = gate();
  const identityRequested = gate();
  const newestRequested = gate();
  const staleErrorRequested = gate();
  const latestErrorRequested = gate();

  await page.route("**/api/v1/auth/session", async (route) => {
    if (!demoted) {
      await route.continue();
      return;
    }
    const response = await route.fetch();
    const envelope = (await response.json()) as GetCurrentSessionResponse;
    await route.fulfill({
      response,
      contentType: "application/json",
      body: JSON.stringify({
        ...envelope,
        data: { ...envelope.data, is_deployment_admin: false },
      }),
    });
  });

  await page.route("**/api/v1/reference-packs?*", async (route) => {
    const requestURL = new URL(route.request().url());
    const search = requestURL.searchParams.get("search");
    const cursor = requestURL.searchParams.get("cursor_token");
    if (cursor === "cursor-a") {
      await appendGate.promise;
      await fulfillPackList(route, [packResource("type_registry.appended")]);
      return;
    }
    if (search === "identity") {
      enteredSearchURL = `${requestURL.pathname}${requestURL.search}`;
      identityRequested.resolve();
      await identityGate.promise;
      await fulfillPackList(route, [packResource("type_registry.identity")]);
      return;
    }
    if (search === "newest") {
      newestRequested.resolve();
      await newestGate.promise;
      await fulfillPackList(route, [packResource("type_registry.newest")]);
      return;
    }
    if (search === "stale-error") {
      staleErrorRequested.resolve();
      await staleErrorGate.promise;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "stale_reference_pack_error" } }),
      });
      return;
    }
    if (search === "latest-error") {
      latestErrorRequested.resolve();
      await latestErrorGate.promise;
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({
          error: { code: "latest_reference_pack_error" },
        }),
      });
      return;
    }
    await fulfillPackList(route, [packResource("type_registry.host")], {
      limit: 100,
      has_more: true,
      next_cursor: "cursor-a",
    });
  });

  await page.route("**/api/v1/reference-packs/refresh", async (route) => {
    expect(route.request().method()).toBe("POST");
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({
        data: jobResource(jobID, "queued", true, 0, 3),
        meta: { request_id: "request-refresh" },
      }),
    });
  });

  await page.route(new RegExp(`/api/v1/jobs/${jobID}$`), async (route) => {
    jobReads += 1;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: jobResource(jobID, "running", true, 1, 3),
        meta: { request_id: "request-job" },
      }),
    });
  });

  await page.route(
    new RegExp(`/api/v1/jobs/${jobID}/cancel$`),
    async (route) => {
      cancelRequests += 1;
      expect(route.request().method()).toBe("POST");
      await route.fulfill({
        status: 202,
        contentType: "application/json",
        body: JSON.stringify({
          data: jobResource(jobID, "cancel_requested", false, 1, 3),
          meta: { request_id: "request-cancel" },
        }),
      });
    },
  );

  await page.goto("/");
  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await expect(page.getByTestId(referencePackAdminPanelTestId())).toBeVisible();

  await expect(
    page.getByTestId(referencePackRowTestId("type_registry.host", "1")),
  ).toBeVisible();
  await page.getByRole("button", { name: "Load more" }).click();
  await page.getByLabel("Search reference packs").fill("identity");
  await page.getByLabel("Search reference packs").press("Enter");
  await identityRequested.promise;
  await expect(page.getByTestId(referencePackListStatusTestId())).toHaveText(
    "Searching reference packs",
  );
  await expect(
    page.getByTestId(referencePackRowTestId("type_registry.host", "1")),
  ).toBeVisible();
  expect(enteredSearchURL).toBe(
    "/api/v1/reference-packs?limit=100&search=identity",
  );

  await page.getByLabel("Search reference packs").fill("newest");
  await page.getByLabel("Search reference packs").press("Enter");
  await newestRequested.promise;
  identityGate.resolve();
  await expect(
    page.getByTestId(referencePackRowTestId("type_registry.identity", "1")),
  ).not.toBeVisible();
  newestGate.resolve();
  await expect(
    page.getByTestId(referencePackRowTestId("type_registry.newest", "1")),
  ).toBeVisible();

  await page.getByLabel("Search reference packs").fill("stale-error");
  await page.getByLabel("Search reference packs").press("Enter");
  await staleErrorRequested.promise;
  await page.getByLabel("Search reference packs").fill("latest-error");
  await page.getByLabel("Search reference packs").press("Enter");
  await latestErrorRequested.promise;
  latestErrorGate.resolve();
  await expect(page.getByTestId(referencePackErrorTestId())).toHaveText(
    "latest_reference_pack_error",
  );
  staleErrorGate.resolve();
  appendGate.resolve();
  await expect(page.getByTestId(referencePackErrorTestId())).toHaveText(
    "latest_reference_pack_error",
  );

  await page.getByTestId(referencePackRefreshAllButtonTestId()).click();
  await expect(page.getByTestId(referencePackJobStatusTestId())).toContainText(
    "running",
    { timeout: 1000 },
  );
  await expect(page.getByTestId(referencePackCancelButtonTestId())).toBeVisible(
    {
      timeout: 1000,
    },
  );

  await new IncidentDirectory(page).open();
  await page.getByTestId(incidentLandingTestId("create-open-button")).click();
  await expect(
    page.getByTestId(incidentLandingTestId("incident-key")),
  ).toBeVisible();
  await page
    .getByTestId(incidentLandingTestId("incident-key"))
    .fill("IR-REFERENCE-PACK");
  await expect(
    page.getByTestId(incidentLandingTestId("incident-key")),
  ).toHaveValue("IR-REFERENCE-PACK");
  await page.getByRole("button", { name: "Close new incident" }).click();

  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await page.getByTestId(referencePackCancelButtonTestId()).click();
  await expect(page.getByTestId(referencePackJobStatusTestId())).toContainText(
    "cancel_requested",
  );
  expect(jobReads).toBeGreaterThanOrEqual(1);
  expect(cancelRequests).toBe(1);

  demoted = true;
  await page.reload();
  await expect(
    page.getByTestId(referencePackRowTestId("type_registry.newest", "1")),
  ).not.toBeVisible();
});

function jobResource(
  id: string,
  status: string,
  cancelable: boolean,
  completed: number,
  total: number,
) {
  return {
    job_id: id,
    status,
    cancelable,
    progress: {
      completed,
      total,
    },
    scope: { kind: "deployment" },
    submitted_by_user_id: "22222222-2222-4222-8222-222222222222",
    submitted_at: "2026-08-04T20:00:00Z",
    updated_at: "2026-08-04T20:00:01Z",
    started_at: status === "queued" ? null : "2026-08-04T20:00:01Z",
    finished_at: null,
    retained_until: null,
    result_summary: null,
    error_summary: null,
    status_route: `/api/v1/jobs/${id}`,
  };
}

function packResource(packKey: string) {
  return {
    activated_at: null,
    activated_by_user_id: null,
    active: false,
    imported_at: "2026-08-04T20:00:00Z",
    imported_by_user_id: null,
    manifest_sha256: "a".repeat(64),
    pack_contract_version: "1",
    pack_key: packKey,
    pack_kind: "type_registry",
    pack_version: "1",
    pack_version_state: "verified_available",
    payload_sha256: "b".repeat(64),
    previous_active_version: null,
    signer_key_id: null,
    source_identifier: null,
    verification_method: "sha256",
    verification_result: "passed",
  };
}

async function fulfillPackList(
  route: Route,
  packVersions: ReturnType<typeof packResource>[],
  paging = { limit: 100, has_more: false, next_cursor: null as string | null },
) {
  await route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify({
      data: { pack_versions: packVersions },
      meta: { request_id: "request-list", paging },
    }),
  });
}

function gate() {
  let release = () => {};
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, resolve: release };
}
