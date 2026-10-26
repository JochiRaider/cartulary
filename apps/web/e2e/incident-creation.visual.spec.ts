import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectCreationControlReachable,
  holdCreationFailure,
  openCreationPresentation,
} from "./support/incidents/creation";
import {
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";

test("Capture incident creation form validation pending recovery and confirmed handoff states.", async ({
  testActorPage: page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const form = await openCreationPresentation(page);
  const key = form.getByLabel("Incident key", { exact: true });
  const title = form.getByLabel("Title", { exact: true });
  await assertViewportVisualRegression(page, "incident-create-initial");
  await key.press("Enter");
  await expect(key).toHaveAttribute("aria-invalid", "true");
  await assertViewportVisualRegression(page, "incident-create-required-errors");
  await key.fill("IR-CREATION-PRESENTATION");
  await title.fill(
    "Incident investigation with a long title and retained operational context",
  );
  await form.locator("summary").click();
  await form
    .getByLabel("Description", { exact: true })
    .fill(
      "Initial report\nPreserve the investigation context while the operation is unresolved.",
    );
  await form.getByLabel("Severity", { exact: true }).fill("high");
  await form
    .getByLabel("TLP", { exact: true })
    .selectOption("TLP:AMBER+STRICT");
  await form.getByLabel("Current phase", { exact: true }).fill("triage");
  await form
    .getByLabel("External case", { exact: true })
    .fill("CASE-" + "REFERENCE".repeat(10));
  await page.setViewportSize({ width: 1024, height: 720 });
  await expectCreationControlReachable(page, title);
  await assertViewportVisualRegression(
    page,
    "incident-create-expanded-details",
  );
  const { release, requested } = await holdCreationFailure(page);
  await form.locator("summary").click();
  await title.press("Enter");
  await requested.promise;
  await page.setViewportSize({ width: 768, height: 640 });
  await expectCreationControlReachable(page, title);
  await assertViewportVisualRegression(page, "incident-create-pending");
  release.release();
  const retry = form.getByRole("button", { name: "Retry creation" });
  await expect(retry).toBeVisible();
  await page.setViewportSize({ width: 640, height: 480 });
  await expectCreationControlReachable(page, retry);
  await assertViewportVisualRegression(page, "incident-create-recovery-short");
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectCreationControlReachable(page, retry);
    await assertViewportVisualRegression(page, "incident-create-recovery-zoom");
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectCreationControlReachable(page, retry);
      await assertViewportVisualRegression(
        page,
        "incident-create-recovery-text-spacing",
      );
    },
  );
  await removeOwnedRoute(page, "**/api/v1/incidents");
  await installOwnedRoute(page, "**/api/v1/auth/session", async (route) =>
    route.fulfill({
      status: 502,
      contentType: "application/json",
      body: JSON.stringify({
        error: { code: "service_unavailable", status: 502, retryable: true },
      }),
    }),
  );
  // This request was blocked before reaching the server; replay is still exact.
  await retry.click();
  await expect(
    page.getByTestId(incidentLandingTestId("create-status")),
  ).toContainText("Incident created, but the workbook could not be opened");
  await page.setViewportSize({ width: 1280, height: 720 });
  await expectCreationControlReachable(
    page,
    form.getByRole("button", { name: "Open created incident" }),
  );
  await assertViewportVisualRegression(
    page,
    "incident-create-confirmed-handoff-failure",
  );
});
