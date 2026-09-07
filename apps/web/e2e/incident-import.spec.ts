import type { GetJobResponse } from "@cartulary/protocol-ts/http";
import {
  incidentAdministrationTestId,
  incidentImportTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import {
  DeploymentAdministration,
  openIncidentControls,
} from "./pages/deploymentAdministration";
import { responseBarrier } from "./support/incidents/creation";
import {
  importBundleFixture,
  installImportObservationFixture,
  openImportPresentation,
} from "./support/incidents/import";

test("imports real active and closed bundles with exact recovery and explicit workbook launch", async ({
  workerAdminPage: page,
}, testInfo) => {
  for (const lifecycle of ["active", "closed"] as const) {
    const bundle = importBundleFixture(lifecycle);
    const form = await openImportPresentation(page);
    const admissions: { id: string; metadata: string; file: Buffer }[] = [];
    const admitted = responseBarrier();
    const release = responseBarrier();
    let failedRead = false;
    let jobReads = 0;
    await page.route("**/api/v1/incident-bundles/import", async (route) => {
      const response = await route.fetch();
      expect(response.status()).toBe(202);
      const envelope: GetJobResponse = await response.json();
      if (admissions.length === 0)
        expect(["queued", "running"]).toContain(envelope.data.status);
      const request = route.request();
      const multipart = await new Response(
        new Uint8Array(request.postDataBuffer() ?? []),
        {
          headers: { "content-type": request.headers()["content-type"] ?? "" },
        },
      ).formData();
      const file = multipart.get("file");
      const metadata = multipart.get("metadata");
      if (!(file instanceof Blob) || !(metadata instanceof Blob))
        throw new Error("Expected multipart file and JSON metadata");
      admissions.push({
        id: envelope.data.job_id,
        metadata: await metadata.text(),
        file: Buffer.from(await file.arrayBuffer()),
      });
      if (admissions.length === 1) {
        admitted.release();
        await release.promise;
        await route.abort("failed");
      } else await route.fulfill({ response });
    });
    await page.route("**/api/v1/jobs/*", async (route) => {
      if (route.request().method() === "GET") ++jobReads;
      if (route.request().method() === "GET" && !failedRead) {
        failedRead = true;
        await route.abort("failed");
      } else await route.continue();
    });
    await form.getByLabel("Incident bundle file").setInputFiles(bundle.upload);
    const start = form.getByRole("button", { name: "Start import" });
    await start.focus();
    await page.keyboard.down("Enter");
    await page.keyboard.down("Enter");
    await page.keyboard.up("Enter");
    await admitted.promise;
    await expect(start).toBeDisabled();
    await expect(form.getByLabel("Incident bundle file")).toBeDisabled();
    expect(admissions).toHaveLength(1);
    release.release();
    const retry = page.getByRole("button", { name: "Retry admission" });
    await expect(retry).toBeVisible();
    await new DeploymentAdministration(page).selectPanel("deployment-users");
    await new DeploymentAdministration(page).selectPanel("incident-import");
    await retry.focus();
    await page.keyboard.press("Enter");
    const retryRead = page.getByRole("button", { name: "Retry observation" });
    await expect(retryRead).toBeVisible();
    expect(admissions).toHaveLength(2);
    expect(admissions[1]).toEqual(admissions[0]);
    expect(admissions[0]?.file).toEqual(bundle.upload.buffer);
    await retryRead.focus();
    await page.keyboard.press("Enter");
    const open = page.getByRole("button", {
      name: "Open imported incident",
      exact: true,
    });
    await expect(open).toBeEnabled({ timeout: 30_000 });
    await expect(open).toHaveCount(1);
    expect(new URL(page.url()).pathname).toBe("/deployment-administration");
    await expect(page.getByTestId(workbookShellReadyTestId())).toHaveCount(0);
    // Session refresh must happen before the ordinary startup, without sheet_ref.
    const session = page.waitForResponse(
      (response) => new URL(response.url()).pathname === "/api/v1/auth/session",
    );
    const startup = page.waitForRequest((request) =>
      new URL(request.url()).pathname.includes(
        `/incidents/${bundle.incidentId}/workbook`,
      ),
    );
    const beforeOpen = jobReads;
    await open.focus();
    await page.keyboard.press("Enter");
    const membership = (await (await session).json()).data.memberships;
    expect(jobReads).toBeGreaterThan(beforeOpen);
    expect(membership).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          incident_id: bundle.incidentId,
          role: "admin",
        }),
      ]),
    );
    expect(new URL((await startup).url()).searchParams.has("sheet_ref")).toBe(
      false,
    );
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    expect(new URL(page.url()).searchParams.get("incident_id")).toBe(
      bundle.incidentId,
    );
    if (lifecycle === "closed") {
      await openIncidentControls(page, "summary");
      await expect(
        page.getByTestId(incidentAdministrationTestId("summary-status")),
      ).toHaveText("Closed, read-only");
    }
    expect(admissions).toHaveLength(2);
    await page.unroute("**/api/v1/incident-bundles/import");
    await page.unroute("**/api/v1/jobs/*");
    await testInfo.attach(`real-import-replay-evidence-${lifecycle}`, {
      body: JSON.stringify({
        lifecycle,
        jobCount: new Set(admissions.map((entry) => entry.id)).size,
        identicalMultipartValues: true,
        workbookOpenedExplicitly: true,
      }),
      contentType: "application/json",
    });
  }
});

test("recovers import observation and cancellation without confusing terminal outcomes", async ({
  workerAdminPage: page,
  workerAdmin,
}) => {
  const fixture = await installImportObservationFixture(
    page,
    workerAdmin.user_id,
  );
  const form = await openImportPresentation(page);
  await form.getByLabel("Incident bundle file").setInputFiles({
    name: "Investigation.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("controlled presentation fixture"),
  });
  await form.getByRole("button", { name: "Start import" }).click();
  const detail = page.getByTestId(incidentImportTestId("detail"));
  await expect(detail.getByRole("heading")).toHaveText("Queued");
  await page.getByRole("button", { name: "Pause updates" }).click();
  await expect(
    page.getByTestId(incidentImportTestId("progress")),
  ).not.toHaveAttribute("value");
  const refresh = async () => {
    await page
      .getByRole("button", { name: /^(Refresh job status|Retry observation)$/ })
      .click();
    await expect(
      page.getByRole("button", {
        name: /^(Refresh job status|Retry observation)$/,
      }),
    ).toBeEnabled();
  };
  fixture.setJob(
    fixture.importJob("running", { progress: { completed: 2, total: 8 } }),
  );
  await refresh();
  await expect(detail.getByRole("heading")).toHaveText("Processing");
  await expect(
    page.getByTestId(incidentImportTestId("progress")),
  ).toHaveAttribute("value", "2");
  fixture.failReads(true);
  await refresh();
  await expect(detail).toContainText("Observation unavailable");
  await expect(detail.getByRole("heading")).toHaveText("Processing");
  fixture.failReads(false);
  await refresh();
  fixture.cancellation("rejected");
  await page
    .getByRole("button", { name: "Cancel import", exact: true })
    .click();
  await expect(detail).toContainText("Cancellation was rejected");
  fixture.cancellation("lost");
  await expect(
    page.getByRole("button", { name: "Cancel import", exact: true }),
  ).toBeEnabled();
  await page
    .getByRole("button", { name: "Cancel import", exact: true })
    .click();
  await expect(detail).toContainText("Cancellation is unconfirmed");
  await expect(
    page.getByRole("button", { name: "Retry cancellation" }),
  ).toBeEnabled();
  fixture.cancellation("requested");
  fixture.setJob(
    fixture.importJob("running", { progress: { completed: 2, total: 8 } }),
  );
  await page.getByRole("button", { name: "Retry cancellation" }).click();
  await expect.poll(() => fixture.cancellationBodies.length).toBe(3);
  expect(fixture.cancellationBodies[2]).toBe(fixture.cancellationBodies[1]);
  await expect(detail.getByRole("heading")).toHaveText(
    "Cancellation requested",
  );
  await expect(
    page.getByRole("button", { name: "Open imported incident" }),
  ).toHaveCount(0);
  fixture.setJob(
    fixture.importJob("canceled", { progress: { completed: 2, total: 8 } }),
  );
  await refresh();
  await expect(detail.getByRole("heading")).toHaveText("Import canceled");
  fixture.setJob(
    fixture.importJob("running", { progress: { completed: 2, total: 8 } }),
  );
  await refresh();
  await expect(detail.getByRole("heading")).toHaveText("Import canceled");
  await expect(detail).toContainText("Observation unavailable");
  expect(fixture.admissionCount()).toBe(1);
});

test("checks import actions and recovers access explicitly before accepting current profile loss", async ({
  workerAdminPage: page,
  workerAdmin,
}) => {
  const fixture = await installImportObservationFixture(
    page,
    workerAdmin.user_id,
  );
  const form = await openImportPresentation(page);
  await form.getByLabel("Incident bundle file").setInputFiles({
    name: "Protected.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("fixture"),
  });
  await form.getByRole("button", { name: "Start import" }).click();
  const cancel = page.getByRole("button", {
    name: "Cancel import",
    exact: true,
  });
  await expect(cancel).toBeEnabled();
  await page.getByRole("button", { name: "Pause updates" }).click();
  const readGate = responseBarrier();
  fixture.gateReads(readGate.promise);
  const before = fixture.readCount();
  await cancel.focus();
  await page.keyboard.press("Enter");
  await expect.poll(() => fixture.readCount()).toBe(before + 1);
  await expect(cancel).toHaveAttribute("aria-busy", "true");
  await expect(cancel).toBeFocused();
  expect(fixture.cancellationBodies).toHaveLength(0);
  await page.keyboard.press("Enter");
  expect(fixture.readCount()).toBe(before + 1);
  fixture.failReads(true);
  readGate.release();
  await expect(
    page
      .getByText(
        "The action could not be confirmed. Review the current job status and retry.",
        { exact: true },
      )
      .first(),
  ).toBeVisible();
  expect(fixture.cancellationBodies).toHaveLength(0);
  await expect(cancel).toBeFocused();
  fixture.gateReads(null);
  fixture.failReads(false);

  let sessionUnavailable = true;
  await page.route("**/api/v1/auth/session", async (route) => {
    if (sessionUnavailable) await route.abort("failed");
    else await route.fallback();
  });
  fixture.readStatus(404);
  await page.getByRole("button", { name: "Retry observation" }).click();
  const retry = page.getByRole("button", { name: "Retry access", exact: true });
  await expect(retry).toBeVisible();
  await expect(
    page.getByText("Protected.tar", { exact: true }).first(),
  ).toBeVisible();
  await retry.focus();
  await page.keyboard.press("Enter");
  await expect(retry).toBeFocused();
  sessionUnavailable = false;
  await page.keyboard.press("Enter");
  await expect(retry).toHaveCount(0);
  await expect(form.getByLabel("Incident bundle file")).toBeFocused();
  await expect(form.getByLabel("Incident bundle file")).toBeEnabled();

  await page.route("**/api/v1/extensions", (route) =>
    route.fulfill({
      status: 200,
      json: { data: { extensions: [] }, meta: { request_id: "profile-loss" } },
    }),
  );
  await page.getByRole("button", { name: "Refresh job status" }).click();
  await expect(page.locator("[data-incident-import]")).toHaveCount(0);
  await expect(page.getByText("Protected.tar", { exact: true })).toHaveCount(0);
});
