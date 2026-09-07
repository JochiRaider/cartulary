import { Buffer } from "node:buffer";
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
import {
  installReferencePackPresentation,
  openReferencePacks,
  referencePackBarrier,
  referencePackBundle,
} from "./support/referencePacks";

test("imports and replays real pack bytes then activates disables reverifies and refreshes exact scope by keyboard", async ({
  workerAdminPage: page,
}, testInfo) => {
  const bundle = referencePackBundle();
  const admissions: { id: string; metadata: string; bytes: Buffer }[] = [];
  const admitted = referencePackBarrier();
  const release = referencePackBarrier();
  let failedRead = false;
  await page.route("**/api/v1/reference-packs/import", async (route) => {
    const response = await route.fetch();
    expect(response.status()).toBe(202);
    const envelope = await response.json();
    const request = route.request();
    const form = await new Response(
      new Uint8Array(request.postDataBuffer() ?? []),
      { headers: { "content-type": request.headers()["content-type"] ?? "" } },
    ).formData();
    const file = form.get("file");
    const metadata = form.get("metadata");
    if (!(file instanceof Blob) || !(metadata instanceof Blob))
      throw new Error("Expected complete multipart capture");
    admissions.push({
      id: envelope.data.job_id,
      bytes: Buffer.from(await file.arrayBuffer()),
      metadata: await metadata.text(),
    });
    expect(JSON.parse(await metadata.text()).activation_policy).toBe(
      "staged_only",
    );
    if (admissions.length === 1) {
      admitted.release();
      await release.promise;
      await route.abort("failed");
    } else await route.fulfill({ response });
  });
  await page.route("**/api/v1/jobs/*", async (route) => {
    if (!failedRead) {
      failedRead = true;
      await route.abort("failed");
    } else await route.continue();
  });
  const panel = await openReferencePacks(page);
  const file = panel.getByLabel("Reference pack bundle");
  await file.focus();
  const chooser = page.waitForEvent("filechooser");
  await page.keyboard.press("Space");
  await (await chooser).setFiles(bundle.upload);
  await new IncidentDirectory(page).open();
  await expect(file).toHaveCount(0);
  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await expect(
    panel.getByText(`Selected file: ${bundle.upload.name}`),
  ).toBeVisible();
  await expect(file).toHaveValue("");
  await expect(file).toBeEnabled();
  const start = panel.getByRole("button", { name: "Import", exact: true });
  await start.focus();
  await page.keyboard.press("Enter");
  await admitted.promise;
  await page.keyboard.press("Enter");
  expect(admissions).toHaveLength(1);
  await expect(file).toBeDisabled();
  release.release();
  const replay = panel.getByRole("button", { name: "Retry exact request" });
  await expect(replay).toBeVisible();
  await new DeploymentAdministration(page).selectPanel("deployment-users");
  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await expect(
    panel.getByText("Checking current access and reconciling reference packs."),
  ).toHaveCount(0);
  await replay.focus();
  await page.keyboard.press("Enter");
  const retryRead = panel.getByRole("button", { name: "Retry observation" });
  await expect(retryRead).toBeVisible();
  expect(admissions[1]).toEqual(admissions[0]);
  expect(admissions[0]?.bytes).toEqual(bundle.upload.buffer);
  await retryRead.focus();
  await page.keyboard.press("Enter");
  await expect(
    panel.getByText("Committed. Catalog data may still need reloading.", {
      exact: true,
    }),
  ).toBeVisible({ timeout: 30_000 });
  const row = panel.getByTestId(
    referencePackRowTestId(bundle.key, bundle.version),
  );
  await expect(row).toContainText("Verified, available");
  await expect(row.getByRole("cell").nth(3)).toHaveText("No", {
    useInnerText: true,
  });
  const actions: { path: string; status: number }[] = [];
  await page.route(
    `**/api/v1/reference-packs/${bundle.key}/${bundle.version}/*`,
    async (route) => {
      const response = await route.fetch();
      actions.push({
        path: new URL(route.request().url()).pathname,
        status: response.status(),
      });
      await route.fulfill({ response });
    },
  );
  const invoke = async (label: string, text: string, column = 1) => {
    await row.getByRole("button", { name: label, exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect(row.getByRole("cell").nth(column)).toContainText(text, {
      timeout: 30_000,
    });
    await expect(
      panel.getByText("Committed. Catalog data may still need reloading.", {
        exact: true,
      }),
    ).toBeVisible({ timeout: 30_000 });
  };
  await invoke("Activate", "Yes", 3);
  await expect(
    row.getByRole("button", { name: "Activate", exact: true }),
  ).toBeDisabled();
  await invoke("Disable", "Disabled");
  await invoke("Reverify", "Verified, available");
  const refreshes: Record<string, unknown>[] = [];
  await page.route("**/api/v1/reference-packs/refresh", async (route) => {
    refreshes.push(route.request().postDataJSON());
    await route.continue();
  });
  await row.getByRole("checkbox").focus();
  await page.keyboard.press("Space");
  await panel
    .getByRole("button", { name: "Refresh selected", exact: true })
    .focus();
  await page.keyboard.press("Enter");
  await expect.poll(() => refreshes.length).toBe(1);
  await expect(
    panel.getByText("Committed. Catalog data may still need reloading.", {
      exact: true,
    }),
  ).toBeVisible({ timeout: 30_000 });
  expect(refreshes[0]?.pack_keys).toEqual([bundle.key]);
  await panel.getByRole("button", { name: "Refresh all", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect.poll(() => refreshes.length).toBe(2);
  await expect(
    panel.getByText("Committed. Catalog data may still need reloading.", {
      exact: true,
    }),
  ).toBeVisible({ timeout: 30_000 });
  expect(refreshes[1]).not.toHaveProperty("pack_keys");
  expect(actions.map(({ status }) => status)).toEqual([200, 200, 202]);
  await testInfo.attach("reference-pack-real-lifecycle", {
    body: JSON.stringify({
      key: bundle.key,
      version: bundle.version,
      admissionJob: admissions[0]?.id,
      actions,
      refreshes,
    }),
    contentType: "application/json",
  });
});

test("supports asynchronous exact actions and retained selection with independent observation recovery", async ({
  workerAdminPage: page,
  workerAdmin,
}) => {
  const fixture = await installReferencePackPresentation(
    page,
    workerAdmin.user_id,
  );
  const panel = await openReferencePacks(page);
  const rows = panel.getByRole("row");
  await rows.nth(1).getByRole("checkbox").check();
  await expect(rows.nth(2).getByRole("checkbox")).toBeChecked();
  await panel.getByLabel("Search reference packs").fill("process");
  await panel.getByLabel("Search reference packs").press("Enter");
  await expect(
    panel.getByText(/selected.*outside|outside.*loaded/i).first(),
  ).toBeVisible();
  await panel.getByRole("button", { name: "Clear filters" }).click();
  await expect(rows.nth(2).getByRole("checkbox")).toBeChecked();
  for (const label of ["Activate", "Disable"] as const) {
    fixture.failReads(true);
    await rows.nth(1).getByRole("button", { name: label, exact: true }).click();
    const retry = panel.getByRole("button", { name: "Retry observation" });
    await expect(retry).toBeVisible();
    await expect(
      panel.getByText(
        "Accepted for processing. Completion has not yet been confirmed.",
        { exact: true },
      ),
    ).toBeVisible();
    fixture.setStatus("succeeded");
    fixture.failReads(false);
    await retry.focus();
    await page.keyboard.press("Enter");
    await expect(
      panel.getByText("Committed. Catalog data may still need reloading.", {
        exact: true,
      }),
    ).toBeVisible();
  }
  await expect(
    panel.getByRole("button", { name: "Dismiss operation" }),
  ).toHaveCount(2);
  await panel
    .getByRole("button", { name: "Refresh selected", exact: true })
    .click();
  await expect(
    panel.getByText("Queued", { exact: false }).last(),
  ).toBeVisible();
  expect(JSON.parse(fixture.admissions.at(-1) ?? "{}").pack_keys).toEqual([
    fixture.packs[0]?.pack_key,
  ]);
});

const jobID = "11111111-1111-4111-8111-111111111111";

test("shows Reference Pack progress and cancel controls without blocking landing interaction", async ({
  workerAdminPage: page,
  workerAdmin,
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
        data: jobResource(jobID, "queued", true, 0, 3, workerAdmin.user_id),
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
        data: jobResource(
          jobID,
          cancelRequests ? "cancel_requested" : "running",
          !cancelRequests,
          1,
          3,
          workerAdmin.user_id,
        ),
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
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: jobResource(
            jobID,
            "cancel_requested",
            false,
            1,
            3,
            workerAdmin.user_id,
          ),
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
    /request was rejected/,
  );
  staleErrorGate.resolve();
  appendGate.resolve();
  await expect(page.getByTestId(referencePackErrorTestId())).toHaveText(
    /request was rejected/,
  );

  await page.getByTestId(referencePackRefreshAllButtonTestId()).click();
  await expect(page.getByTestId(referencePackJobStatusTestId())).toContainText(
    "Running",
    { timeout: 5000 },
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
    "Cancellation requested",
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
  actorId: string,
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
    submitted_by_user_id: actorId,
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
    pack_contract_version: "cartulary.reference_pack.v1",
    pack_key: packKey,
    pack_kind: "type_registry",
    pack_version: "1",
    pack_version_state: "verified_available",
    payload_sha256: "b".repeat(64),
    previous_active_version: null,
    signer_key_id: null,
    source_identifier: null,
    verification_method: "manifest_sha256_v1",
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
