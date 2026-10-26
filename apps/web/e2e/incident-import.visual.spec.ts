import { Buffer } from "node:buffer";
import { incidentImportTestId } from "@cartulary/ui-contracts";
import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectImportControlReachable,
  installImportObservationFixture,
  openImportPresentation,
} from "./support/incidents/import";
import { responseBarrier } from "./support/runtime/cleanup";
import { installOwnedRoute } from "./support/transport/requestInterception";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";

test("Capture incident import admission observation cancellation and result recovery states.", async ({
  testActorPage: page,
  testActor,
}) => {
  const fixture = await installImportObservationFixture(
    page,
    testActor.user_id,
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  const form = await openImportPresentation(page);
  await assertViewportVisualRegression(page, "incident-import-empty");
  await form.getByRole("button", { name: "Start import" }).click();
  await expect(form.getByLabel("Incident bundle file")).toHaveAttribute(
    "aria-invalid",
    "true",
  );
  await assertViewportVisualRegression(page, "incident-import-required");
  await form.getByLabel("Incident bundle file").setInputFiles({
    name: "Investigation.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("presentation fixture"),
  });
  const pending = responseBarrier(page);
  fixture.gateAdmission(pending.promise);
  fixture.failAdmission(true);
  await form.getByRole("button", { name: "Start import" }).click();
  await expect(
    page.getByRole("progressbar", { name: "Upload admission" }),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "incident-import-admission-pending",
  );
  pending.release();
  const retry = page.getByRole("button", { name: "Retry admission" });
  await expect(retry).toBeVisible();
  await page.setViewportSize({ width: 640, height: 480 });
  await expectImportControlReachable(page, retry);
  await assertViewportVisualRegression(page, "incident-import-recovery-short");
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectImportControlReachable(page, retry);
    await assertViewportVisualRegression(page, "incident-import-recovery-zoom");
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectImportControlReachable(page, retry);
      await assertViewportVisualRegression(
        page,
        "incident-import-recovery-text-spacing",
      );
    },
  );
  await page.setViewportSize({ width: 1280, height: 900 });
  fixture.failAdmission(false);
  await retry.click();
  const detail = page.getByTestId(incidentImportTestId("detail"));
  await expect(detail.getByRole("heading")).toHaveText("Queued");
  await page.getByRole("button", { name: "Pause updates" }).click();
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
  const focusObservationRecovery = async () => {
    const control = page.getByRole("button", {
      name: /^(Refresh job status|Retry observation)$/,
    });
    await expect(control).toBeEnabled();
    // Observation can replace a previously disabled control. Declare keyboard
    // focus so captures do not depend on whether that DOM node was retained.
    await control.focus();
    await expectImportControlReachable(page, control);
    return control;
  };
  await assertViewportVisualRegression(
    page,
    "incident-import-queued-indeterminate",
  );
  fixture.setJob(
    fixture.importJob("running", { progress: { completed: 2, total: 8 } }),
  );
  await refresh();
  await expect(detail.getByRole("heading")).toHaveText("Processing");
  await focusObservationRecovery();
  await assertViewportVisualRegression(
    page,
    "incident-import-running-determinate",
  );
  fixture.failReads(true);
  await refresh();
  await expect(detail).toContainText("Observation unavailable");
  const retryObservation = await focusObservationRecovery();
  await assertViewportVisualRegression(
    page,
    "incident-import-observation-unavailable",
  );
  await expect(retryObservation).toBeFocused();
  fixture.failReads(false);
  await refresh();
  await page
    .getByRole("button", { name: "Cancel import", exact: true })
    .click();
  await expect(detail.getByRole("heading")).toHaveText(
    "Cancellation requested",
  );
  await assertViewportVisualRegression(
    page,
    "incident-import-cancel-requested",
  );
  fixture.setJob(
    fixture.importJob("canceled", { progress: { completed: 2, total: 8 } }),
  );
  await refresh();
  await expect(detail.getByRole("heading")).toHaveText("Import canceled");
  const canceledRefresh = await focusObservationRecovery();
  await assertViewportVisualRegression(page, "incident-import-canceled");
  await expect(canceledRefresh).toBeFocused();
  await expect(canceledRefresh).toBeEnabled();
  for (const status of ["failed", "succeeded"] as const) {
    // A new application lifetime isolates terminal outcomes without regressing a job.
    await openImportPresentation(page);
    fixture.setJob(fixture.importJob(status));
    await form.getByLabel("Incident bundle file").setInputFiles({
      name: "Investigation.tar",
      mimeType: "application/x-tar",
      buffer: Buffer.from("presentation fixture"),
    });
    await form.getByRole("button", { name: "Start import" }).click();
    await expect(detail.getByRole("heading")).toHaveText(
      status === "failed" ? "Import failed" : "Import succeeded",
    );
    await expect(
      page.getByRole("button", { name: "Refresh job status" }),
    ).toBeEnabled();
    await assertViewportVisualRegression(page, `incident-import-${status}`);
  }
  await installOwnedRoute(page, "**/api/v1/auth/session", (route) =>
    route.fulfill({
      status: 503,
      json: {
        error: { code: "service_unavailable", status: 503, retryable: true },
      },
    }),
  );
  await page
    .getByRole("button", { name: "Open imported incident", exact: true })
    .click();
  await expect(detail).toContainText(
    "The import succeeded, but the workbook could not be opened",
  );
  await assertViewportVisualRegression(
    page,
    "incident-import-handoff-unavailable",
  );
  const actionRead = responseBarrier(page);
  fixture.gateReads(actionRead.promise);
  const open = page.getByRole("button", {
    name: "Open imported incident",
    exact: true,
  });
  await open.click();
  await expect(open).toHaveAttribute("aria-busy", "true");
  await assertViewportVisualRegression(page, "incident-import-action-checking");
  fixture.failReads(true);
  actionRead.release();
  await expect(
    page
      .getByText(
        "The action could not be confirmed. Review the current job status and retry.",
        { exact: true },
      )
      .first(),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "incident-import-action-unconfirmed",
  );
  fixture.gateReads(null);
  fixture.failReads(false);
  fixture.readStatus(404);
  await page.getByRole("button", { name: "Retry observation" }).click();
  const retryAccess = page.getByRole("button", {
    name: "Retry access",
    exact: true,
  });
  await expect(retryAccess).toBeVisible();
  await expectImportControlReachable(page, retryAccess);
  await assertViewportVisualRegression(page, "incident-import-access-retry");
});
