import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyReportingProgressContract as contract,
  dataTestIdSelector,
  gridShellTestId,
  incidentControlsMenuItemTestId,
  incidentControlsTriggerTestId,
  rowCellTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, Request, TestInfo } from "@playwright/test";
import type { WorkerAdminEntry } from "../support/auth/workerAdmin";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { expect, test } from "./fixtures";
import type { JobProgressCompletion } from "./jobProgressPaint";
import {
  attachReportingProgressObservation,
  measureReportingProgress,
  type ReportingProgressPurpose,
} from "./reportingProgressTimingSupport";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1_000 });

async function runReportingProgress(
  purpose: ReportingProgressPurpose,
  browser: Browser,
  page: Page,
  workerAdmin: WorkerAdminEntry,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  info: TestInfo,
) {
  const predicate = contract.predicates.find((value) => value.kind === purpose);
  if (!predicate) throw new Error("Reporting predicate is not admitted");
  let traffic: Ac043TrafficDriver | null = null,
    completion: JobProgressCompletion | undefined;
  let target: Parameters<typeof attachReportingProgressObservation>[2];
  let attached = false,
    primary: unknown;
  const panel = page.getByRole("region", {
    name: "Snapshots and reports",
    exact: true,
  });
  const requests: unknown[] = [];
  const cleanupErrors: unknown[] = [];
  const capture = (request: Request) => {
    const path = new URL(request.url()).pathname;
    if (
      request.method() === "POST" &&
      path ===
        (purpose === "snapshot" ? "/api/v1/snapshots" : "/api/v1/releases")
    )
      requests.push(request.postDataJSON());
  };
  try {
    expect(info.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, workerAdmin, {
      fixtureProfileId: contract.fixture_profile_id,
    });
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const record = firstPage[0];
    if (!record) throw new Error("Reporting measurement target is missing");
    const excluded = new Set(firstPage.map((row) => row.record_id)),
      fieldKey = contract.workload.field_key;
    expect(record.cells[fieldKey]?.value).toBeNull();
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await page.getByLabel("Account and application navigation").click();
    await page.getByTestId(incidentControlsTriggerTestId()).click();
    await page
      .getByTestId(incidentControlsMenuItemTestId("snapshot-reports"))
      .click();
    await expect(panel).toBeVisible();
    let snapshotId: string | null = null;
    if (purpose === "report") {
      await panel
        .getByRole("button", { name: "Create snapshot", exact: true })
        .click();
      const ready = panel.locator("p").filter({ hasText: /^Snapshot ready:/ });
      const snapshotJob = panel.locator('[data-job-purpose="snapshot"]');
      await expect(snapshotJob).toHaveAttribute(
        "data-job-status",
        /^(succeeded|failed|canceled)$/,
        { timeout: 120_000 },
      );
      await expect(snapshotJob).toHaveAttribute("data-job-status", "succeeded");
      await expect(ready).toBeVisible();
      snapshotId = await ready.locator("code").textContent();
      expect(snapshotId).toMatch(/^[a-f0-9-]{36}$/);
    }
    traffic = await startAc043SnapshotTraffic(
      browser,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows.filter((row) => !excluded.has(row.record_id)),
      snapshot.runtime.backgroundAccounts,
      predicate.predicate_id,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    const presence = page.getByTestId(workbookPresenceSummaryTestId());
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
    );
    await scrollGridCellIntoView({
      page,
      surface: contract.view_schema_id,
      recordId: record.record_id,
      cellKey: fieldKey,
    });
    const cell = page.getByTestId(rowCellTestId(record.record_id, fieldKey));
    await expect(cell).toHaveText("—");
    const button = panel.getByRole("button", {
      name:
        purpose === "snapshot" ? "Create snapshot" : "Render internal report",
      exact: true,
    });
    await button.scrollIntoViewIfNeeded();
    await button.focus();
    await expect(button).toBeFocused();
    const clickPoint = await cell.evaluate((node) => {
      const field = node.closest<HTMLElement>("[data-grid-field-key]");
      if (!field) throw new Error("Reporting target field is missing");
      const box = field.getBoundingClientRect();
      if (
        box.left < 0 ||
        box.right > innerWidth ||
        box.top < 0 ||
        box.bottom > innerHeight ||
        ![
          [box.left + 4, box.top + 4],
          [box.right - 4, box.bottom - 4],
          [(box.left + box.right) / 2, (box.top + box.bottom) / 2],
        ].every(([x, y]) => {
          if (x === undefined || y === undefined) return false;
          const hit = document.elementFromPoint(x, y);
          return hit !== null && field.contains(hit);
        })
      )
        throw new Error(
          "Reporting target is offscreen or covered before measurement",
        );
      return { x: box.left + box.width / 2, y: box.top + box.height / 2 };
    });
    target = {
      recordId: record.record_id,
      rowVersion: record.row_version,
      snapshotId,
    };
    const editorId = timelineScalarEditorTestId({
      recordId: record.record_id,
      fieldKey,
      surface: "grid",
    });
    page.on("request", capture);
    completion = await measureReportingProgress(
      purpose,
      page,
      {
        recordId: record.record_id,
        rowVersion: record.row_version,
        fieldKey,
        draftValue: contract.workload.draft_suffix,
        gridSelector: dataTestIdSelector(
          gridShellTestId(contract.view_schema_id),
        ),
        editorSelector: dataTestIdSelector(editorId),
        progressSurfaceSelector: '[data-reporting-workflow="true"]',
      },
      async () => {
        await page.keyboard.press("Enter");
        await page.mouse.click(clickPoint.x, clickPoint.y);
        const editor = page.getByTestId(editorId);
        await editor.press("End");
        await editor.press(contract.workload.draft_suffix);
      },
    );
    page.off("request", capture);
    expect(requests).toHaveLength(1);
    expect(requests[0]).toEqual(
      purpose === "snapshot"
        ? {
            incident_id: snapshot.incidentId,
            client_txn_id: expect.any(String),
          }
        : {
            snapshot_id: snapshotId,
            client_txn_id: expect.any(String),
            template_id: contract.workload.template_id,
            template_version: contract.workload.template_version,
            output_kind: contract.workload.output_kind,
            release_scope: contract.workload.release_scope,
          },
    );
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
    );
    await traffic.stop();
    traffic = null;
    await attachReportingProgressObservation(
      info,
      predicate.predicate_id,
      target,
      completion,
    );
    attached = true;
    await info.attach(`Reporting ${purpose} progress surface`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    expect(completion.elapsed).toBeLessThanOrEqual(contract.threshold_ms);
  } catch (error) {
    primary = error;
    if (!attached)
      await attachReportingProgressObservation(
        info,
        predicate.predicate_id,
        target,
        completion,
        "Reporting progress measurement did not complete",
      );
  } finally {
    page.off("request", capture);
    try {
      await traffic?.stop();
    } catch (error) {
      cleanupErrors.push(error);
    }
    try {
      await page.keyboard.press("Escape");
      const cancel = panel.getByRole("button", {
        name: "Cancel job",
        exact: true,
      });
      let cancellationRequested = false;
      if ((await cancel.isVisible()) && (await cancel.isEnabled())) {
        await cancel.click();
        cancellationRequested = true;
      }
      const job = panel.locator("[data-job-id]");
      if (await job.count()) {
        await expect(job).toHaveAttribute(
          "data-job-status",
          /^(succeeded|failed|canceled)$/,
          { timeout: 120_000 },
        );
        await info.attach(`Reporting ${purpose} terminal cleanup`, {
          body: JSON.stringify({
            cancellation_requested: cancellationRequested,
            terminal_status: await job.getAttribute("data-job-status"),
          }),
          contentType: "application/json",
        });
      }
    } catch (error) {
      cleanupErrors.push(error);
    }
  }
  if (cleanupErrors.length)
    throw new AggregateError(
      [...(primary ? [primary] : []), ...cleanupErrors],
      "Reporting measurement or cleanup failed",
    );
  if (primary) throw primary;
}

test("measures snapshot job progress and cancel while Timeline accepts editing", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runReportingProgress(
    "snapshot",
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
  );
});
test("measures report job progress and cancel while Timeline accepts editing", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runReportingProgress(
    "report",
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
  );
});
