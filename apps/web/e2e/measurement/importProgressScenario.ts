import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyImportApplyPerformanceContract as applyContract,
  dataTestIdSelector,
  cartularyImportDiscoveryPerformanceContract as discoveryContract,
  gridShellTestId,
  incidentControlsMenuItemTestId,
  incidentControlsTriggerTestId,
  rowCellTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookImportAssistantTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, TestInfo } from "@playwright/test";
import type { FixtureActorCredentials } from "../support/auth/sessions";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { expect } from "./fixtures";
import { importDiscoveryWorkbook } from "./importDiscoveryWorkbook";
import {
  attachImportProgressObservation,
  type ImportProgressCompletion,
  type ImportSourceFormat,
  measureImportProgress,
} from "./importProgressTimingSupport";

export async function runImportProgress(
  purpose: "discovery" | "apply",
  browser: Browser,
  page: Page,
  testActor: FixtureActorCredentials,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  info: TestInfo,
  format: ImportSourceFormat,
) {
  const contract = purpose === "discovery" ? discoveryContract : applyContract;
  const predicate = contract.predicates.find((value) => value.kind === format);
  if (!predicate) throw new Error("Import progress predicate is not admitted");
  let traffic: Ac043TrafficDriver | null = null;
  let completion: ImportProgressCompletion | undefined;
  let target: Parameters<typeof attachImportProgressObservation>[3];
  let attached = false;
  const assistant = page.getByTestId(workbookImportAssistantTestId());
  try {
    expect(info.retry).toBe(0);
    const source = importDiscoveryWorkbook(format);
    const snapshot = await prepareAc043Snapshot(page, testActor, {
      fixtureProfileId: contract.fixture_profile_id,
    });
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const record = firstPage[0];
    if (!record) throw new Error("Import measurement has no target row");
    const excluded = new Set(firstPage.map((row) => row.record_id));
    const fieldKey = contract.workload.field_key;
    const initialValue = record.cells[fieldKey]?.value;
    if (initialValue !== null)
      throw new Error("Import target Analyst cell must be initially null");
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
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
    await page.getByLabel("Account and application navigation").click();
    await page.getByTestId(incidentControlsTriggerTestId()).click();
    await page
      .getByTestId(incidentControlsMenuItemTestId("import-assistant"))
      .click();
    await expect(assistant).toBeVisible();
    await assistant.getByLabel("Source workbook").setInputFiles(source);
    if (purpose === "apply") {
      await assistant
        .getByRole("button", { name: "Upload and discover" })
        .click();
      await expect(assistant.getByRole("status")).toContainText(
        "Discovered 1 import unit",
        { timeout: 120_000 },
      );
      const unit = assistant.getByRole("region", {
        name: "Import unit 1",
        exact: true,
      });
      await expect(unit).toContainText("20000 rows × 2 columns");
      await expect(
        unit.getByRole("combobox", { name: "Target view", exact: true }),
      ).toHaveValue(applyContract.view_schema_id);
      await expect(
        unit.getByRole("combobox", { name: "Activity Synopsis", exact: true }),
      ).toHaveValue(applyContract.workload.mapped_field_key);
      await expect(
        unit.getByRole("combobox", {
          name: "Unmapped source note",
          exact: true,
        }),
      ).toHaveValue("");
      await unit
        .getByRole("button", {
          name: "Approve mapping and select",
          exact: true,
        })
        .click();
      await expect(assistant.getByRole("status")).toContainText(
        "ready to apply",
      );
      await expect(
        assistant.getByRole("region", {
          name: "Import session review",
          exact: true,
        }),
      ).toContainText("Persisted selection: 1 unit.");
      await expect(
        unit.getByRole("button", {
          name: "Mapping approved and selected",
          exact: true,
        }),
      ).toBeDisabled();
    }
    await scrollGridCellIntoView({
      page,
      surface: contract.view_schema_id,
      recordId: record.record_id,
      cellKey: fieldKey,
    });
    const cell = page.getByTestId(rowCellTestId(record.record_id, fieldKey));
    await expect(cell).toHaveText("—");
    // The default Analyst field remains beside the non-modal support drawer.
    // Verify its complete hit-testable geometry before accepting the measured action.
    const actionButton = assistant.getByRole("button", {
      name:
        purpose === "discovery"
          ? "Upload and discover"
          : "Apply 1 selected unit",
      exact: true,
    });
    await actionButton.scrollIntoViewIfNeeded();
    await actionButton.focus();
    await expect(actionButton).toBeFocused();
    await expect
      .poll(() =>
        cell.evaluate((node) => {
          const field = node.closest<HTMLElement>("[data-grid-field-key]");
          if (!field) return false;
          const box = field.getBoundingClientRect();
          return (
            box.left >= 0 &&
            box.right <= innerWidth &&
            box.top >= 0 &&
            box.bottom <= innerHeight &&
            [
              [box.left + 4, box.top + 4],
              [box.right - 4, box.bottom - 4],
              [(box.left + box.right) / 2, (box.top + box.bottom) / 2],
            ].every(([x, y]) => {
              if (x === undefined || y === undefined) return false;
              const hit = document.elementFromPoint(x, y);
              return hit !== null && field.contains(hit);
            })
          );
        }),
      )
      .toBe(true);
    const clickPoint = await cell.evaluate((node) => {
      const field = node.closest<HTMLElement>("[data-grid-field-key]");
      if (!field) throw new Error("Import target field disappeared");
      const box = field.getBoundingClientRect();
      return { x: box.left + box.width / 2, y: box.top + box.height / 2 };
    });
    const editorId = timelineScalarEditorTestId({
      recordId: record.record_id,
      fieldKey,
      surface: "grid",
    });
    target = {
      format,
      recordId: record.record_id,
      rowVersion: record.row_version,
    };
    completion = await measureImportProgress(
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
        progressSurfaceSelector: dataTestIdSelector(
          workbookImportAssistantTestId(),
        ),
      },
      async () => {
        await page.keyboard.press("Enter");
        // A single pointer click selects and opens the scalar editor.
        await page.mouse.click(clickPoint.x, clickPoint.y);
        const editor = page.getByTestId(editorId);
        await editor.press("End");
        await editor.press(contract.workload.draft_suffix);
      },
    );
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
    );
    await traffic.stop();
    traffic = null;
    await attachImportProgressObservation(
      purpose,
      info,
      predicate.predicate_id,
      target,
      completion,
    );
    attached = true;
    expect(
      completion.elapsed,
      `${format} ${purpose} progress, cancel and grid editing`,
    ).toBeLessThanOrEqual(contract.threshold_ms);
  } catch (error) {
    if (!attached)
      await attachImportProgressObservation(
        purpose,
        info,
        predicate.predicate_id,
        target,
        completion,
        "Import progress measurement did not complete",
      );
    throw error;
  } finally {
    await traffic?.stop();
    // Draft discard and ordinary cancellation are outside the timed interval.
    await page.keyboard.press("Escape");
    const cancel = assistant.getByRole("button", {
      name: "Cancel import",
      exact: true,
    });
    let cancellationRequested = false;
    if ((await cancel.isVisible()) && (await cancel.isEnabled())) {
      await cancel.click();
      cancellationRequested = true;
    }
    const job = assistant.getByRole("region", {
      name: "Import job",
      exact: true,
    });
    if (await job.count()) {
      await expect(job).toHaveAttribute(
        "data-job-status",
        /^(succeeded|failed|canceled)$/,
        { timeout: purpose === "apply" ? 120_000 : 30_000 },
      );
      const refresh = assistant.getByRole("button", {
        name: "Refresh session and outcomes",
        exact: true,
      });
      await expect(refresh).toBeEnabled();
      await refresh.click();
      await expect(refresh).toBeEnabled();
      const unitOutcomes = await assistant
        .locator("[data-import-unit-status]")
        .evaluateAll((nodes) =>
          nodes.map((node) => node.getAttribute("data-import-unit-status")),
        );
      for (const outcome of unitOutcomes)
        expect([
          "discovered",
          "mapped",
          "ready",
          "skipped",
          "applied",
          "failed",
          "rejected",
        ]).toContain(outcome);
      await info.attach(`Import ${purpose} terminal cleanup`, {
        body: JSON.stringify({
          cancellation_requested: cancellationRequested,
          unit_outcomes: unitOutcomes,
          terminal_status: await job.getAttribute("data-job-status"),
        }),
        contentType: "application/json",
      });
    }
  }
}
