import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import {
  auditBrowserEvent,
  installAuditPresentation,
  openAdministrativeAudit,
} from "./support/administrativeAudit";
import { responseBarrier } from "./support/runtime/cleanup";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";

test("Capture Administrative audit loading empty unavailable paging and inspected values.", async ({
  testActorPage: page,
}) => {
  const fixture = await installAuditPresentation(page);
  const loading = responseBarrier(page);
  fixture.gateRead(loading.promise);
  await page.setViewportSize({ width: 1280, height: 900 });
  const panel = await openAdministrativeAudit(page);
  await expect(panel.getByRole("status")).toContainText("Loading");
  await assertViewportVisualRegression(page, "administrative-audit-loading");
  loading.release();
  await expect(panel.getByRole("status")).toContainText("Page 1:");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  await panel
    .getByRole("region", { name: /^Details for / })
    .evaluate((element) =>
      element.scrollIntoView({ block: "start", behavior: "instant" }),
    );
  await assertViewportVisualRegression(page, "administrative-audit-inspected");
  fixture.fail();
  await panel.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    panel.getByRole("button", { name: "Retry refresh" }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "administrative-audit-stale");
  fixture.setPage([]);
  await panel.getByRole("button", { name: "Retry refresh" }).click();
  await expect(
    panel.getByText("No deployment audit events are available."),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "administrative-audit-empty");
  fixture.fail();
  await panel.getByLabel("Target kind", { exact: true }).selectOption("user");
  await panel
    .getByRole("button", { name: "Apply filters", exact: true })
    .click();
  await expect(panel.getByRole("button", { name: "Retry read" })).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "administrative-audit-unavailable",
  );
  fixture.setPage([auditBrowserEvent()], "page-two");
  await panel
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  await expect(
    panel.getByRole("button", { name: "Next page", exact: true }),
  ).toBeEnabled();
  fixture.setPage([
    auditBrowserEvent({
      occurred_at: "2026-05-23T12:00:00Z",
      audit_event_id: "00000000-0000-4000-8000-000000002000",
    }),
  ]);
  await panel.getByRole("button", { name: "Next page", exact: true }).click();
  await expect(panel.getByRole("status")).toContainText("Page 2:");
  await assertViewportVisualRegression(page, "administrative-audit-page-two");
  await panel.getByRole("button", { name: /^Inspect / }).click();
  await page.setViewportSize({ width: 390, height: 480 });
  await panel
    .getByRole("table", { name: "Published field changes", exact: true })
    .evaluate((element) =>
      element.scrollIntoView({ block: "start", behavior: "instant" }),
    );
  await assertViewportVisualRegression(
    page,
    "administrative-audit-inspected-narrow",
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await panel
      .getByRole("table", { name: "Published field changes", exact: true })
      .evaluate((element) =>
        element.scrollIntoView({ block: "start", behavior: "instant" }),
      );
    await assertViewportVisualRegression(
      page,
      "administrative-audit-inspected-zoom",
    );
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await panel
        .getByRole("table", { name: "Published field changes", exact: true })
        .evaluate((element) =>
          element.scrollIntoView({ block: "start", behavior: "instant" }),
        );
      await assertViewportVisualRegression(
        page,
        "administrative-audit-inspected-spacing",
      );
    },
  );
});
