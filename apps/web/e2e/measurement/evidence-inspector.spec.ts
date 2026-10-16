import {
  cartularyEvidenceInspectorContract as contract,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { request } from "@playwright/test";
import {
  loginTrackedUserViaPage,
  logoutAndVerify,
  readCurrentSession,
} from "../support/auth/sessions";
import { authHeadersForStorageState } from "../support/auth/storageState";
import {
  type EvidenceInspectorTarget,
  evidenceInspectorAccount,
  prepareEvidenceInspectorSnapshot,
} from "../support/performance/evidenceInspectorSnapshot";
import { apiBase, webBase } from "../support/runtime/configuration";
import { openTimelineInspector } from "../support/workbook/rowMutations";
import {
  attachEvidenceInspectorObservation,
  type EvidenceInspectorSample,
  measureEvidenceInspectorPaint,
} from "./evidenceInspectorTimingSupport";
import { expect, test } from "./fixtures";

// Access capabilities and complete object bodies stay private during preflight.
test.use({ trace: "off", screenshot: "off", video: "off" });
test.describe.configure({ mode: "serial", timeout: 5 * 60 * 1000 });
test("measures the fresh Evidence inspector metadata shell within AC-045", async ({
  browser,
}, info) => {
  expect(info.retry).toBe(0);
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
  });
  const page = await context.newPage();
  const samples: EvidenceInspectorSample[] = [];
  let target: EvidenceInspectorTarget | undefined,
    attached = false,
    authenticated = false;
  try {
    await page.goto(webBase);
    await loginTrackedUserViaPage(page, evidenceInspectorAccount());
    authenticated = true;
    const snapshot = await prepareEvidenceInspectorSnapshot(page);
    target = snapshot.target;
    const session = await readCurrentSession(page);
    expect(session.is_deployment_admin).toBe(false);
    expect(session.memberships).toEqual([
      { incident_id: snapshot.incidentId, role: "editor" },
    ]);
    await page.goto(`${webBase}/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    const view = contract.view_schema_id;
    await page.getByTestId(workbookFilterPopoverTriggerTestId(view)).click();
    await page
      .getByTestId(gridFilterFieldTestId(view))
      .selectOption(contract.workload.filter_field_key);
    await page
      .getByTestId(workbookFilterOperatorTestId(view))
      .selectOption(contract.workload.filter_operator);
    await page.getByTestId(gridFilterValueTestId(view)).selectOption("true");
    await page.getByTestId(gridFilterApplyTestId(view)).click();
    const filter = page.getByTestId(workbookFilterPopoverTriggerTestId(view));
    if ((await filter.getAttribute("aria-expanded")) === "true")
      await filter.click();
    const frame = page.locator(`[data-workbook-view-schema-id="${view}"]`);
    await expect(frame).toHaveAttribute("data-workbook-query-pending", "false");
    await expect(frame).toHaveAttribute(
      "data-workbook-query-filter-count",
      "1",
    );
    await expect(frame).toHaveAttribute("data-workbook-query-sort-count", "0");
    await expect(frame).toHaveAttribute("data-workbook-query-group-field", "");
    await expect(
      frame.locator('[role="row"][data-grid-record-id]'),
    ).toHaveCount(1);
    // Preparation selects the Evidence section. Its presentation selection may
    // survive closing; metadata itself must be discarded by the product.
    const preparationRead = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname ===
          `/api/v1/incidents/${snapshot.incidentId}/evidence-metadata` &&
        response.request().method() === "POST",
    );
    await openTimelineInspector(page, target.record_id);
    await (await preparationRead).finished();
    const section = page.locator(
      '[data-inspector-navigation-panel="evidence"]',
    );
    if (await section.isVisible()) await section.click();
    else {
      await page.getByRole("button", { name: /^Sections:/u }).click();
      await page
        .getByRole("navigation", { name: "Inspector sections", exact: true })
        .getByRole("button", { name: "Evidence", exact: true })
        .click();
    }
    const inspector = page.getByTestId(timelineInspectorTestId());
    await expect(
      page
        .getByTestId(timelineInspectorSectionTestId("evidence"))
        .locator('[data-evidence-metadata-state="ready"]'),
    ).toHaveCount(5);
    for (
      let index = 0;
      index < contract.warmup_passes + contract.measured_samples;
      index++
    ) {
      await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
      await expect(inspector).toHaveCount(0);
      const paint = await measureEvidenceInspectorPaint(
        page,
        snapshot.incidentId,
        target,
        () => page.getByTestId(workbookInspectorToggleTestId(view)).click(),
      );
      samples.push({
        ...paint,
        sample_index: index,
        warmup: index === 0,
        total_ms: paint.second_frame_at - paint.accepted_at,
      });
    }
    const observation = await attachEvidenceInspectorObservation(
      info,
      samples,
      target,
    );
    attached = true;
    expect(observation.outcome).toBe("passed");
    expect(observation.statistics.p95_ms).toBeLessThanOrEqual(
      contract.threshold_ms,
    );
  } catch (error) {
    if (!attached)
      await attachEvidenceInspectorObservation(
        info,
        samples,
        target,
        error instanceof Error ? error.message : String(error),
      );
    throw error;
  } finally {
    try {
      if (authenticated) {
        const state = await context.storageState();
        const auth = await request.newContext({
          baseURL: apiBase,
          extraHTTPHeaders: authHeadersForStorageState(state),
        });
        try {
          await logoutAndVerify(auth, state, "Evidence inspector analyst");
        } finally {
          await auth.dispose();
        }
      }
    } finally {
      await context.close();
    }
  }
});
