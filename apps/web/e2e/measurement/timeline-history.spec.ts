import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyHistoryPerformanceContract as contract,
  dataTestIdSelector,
  gridShellTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackConfirmButtonTestId,
  timelineMutationSubstrateReadyTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, TestInfo } from "@playwright/test";
import { readCurrentSession } from "../support/auth/sessions";
import type { WorkerAdminEntry } from "../support/auth/workerAdmin";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { uniqueTxn } from "../support/runtime/fixtureIdentity";
import {
  fetchFullRecordHistory,
  openHistoryEventDetails,
} from "../support/workbook/history";
import { patchRecord, queryViewRows } from "../support/workbook/query";
import { clickTimelineRowAction } from "../support/workbook/rowMutations";
import { expect, test } from "./fixtures";
import {
  attachHistoryObservation,
  type HistoryKind,
  measureHistory,
} from "./historyTimingSupport";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1_000 });

async function runHistoryMeasurement(
  browser: Browser,
  page: Page,
  workerAdmin: WorkerAdminEntry,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  info: TestInfo,
  kind: HistoryKind,
) {
  const predicate = contract.predicates.find(
    (predicate) => predicate.kind === kind,
  );
  if (predicate === undefined)
    throw new Error("History predicate is not admitted");
  const predicateId = predicate.predicate_id;
  let traffic: Ac043TrafficDriver | null = null;
  let completion: Awaited<ReturnType<typeof measureHistory>> | undefined;
  let target: Parameters<typeof attachHistoryObservation>[2];
  let attached = false;
  try {
    expect(info.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, workerAdmin, {
      foregroundRole: contract.workload.foreground_role,
    });
    const actor = await readCurrentSession(page);
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const record = firstPage[0];
    if (record === undefined)
      throw new Error("History fixture has no target row");
    const excluded = new Set(firstPage.map((row) => row.record_id));
    const [analyst, device] = contract.workload.fields;
    const prior = await patchRecord(page, record.record_id, {
      view_schema_id: contract.view_schema_id,
      base_row_version: record.row_version,
      client_txn_id: uniqueTxn("history-before"),
      changes: [
        { field_key: analyst.field_key, value: analyst.before },
        { field_key: device.field_key, value: device.before },
      ],
    });
    const current = await patchRecord(page, record.record_id, {
      view_schema_id: contract.view_schema_id,
      base_row_version: prior.row_version,
      client_txn_id: uniqueTxn("history-mistake"),
      changes: [
        { field_key: analyst.field_key, value: analyst.mistake },
        ...(kind === "history_entry"
          ? []
          : [{ field_key: device.field_key, value: device.mistake }]),
      ],
    });
    const before = await fetchFullRecordHistory(page, record.record_id);
    const selected = before.items.find(
      (item) =>
        item.available_rollback_actions.includes(kind) &&
        item.revision_no ===
          (kind === "row_restore" ? prior.row_version : current.row_version),
    );
    expect(
      selected,
      "prepared history must advertise the requested action",
    ).toBeDefined();
    if (!selected) throw new Error("Prepared history action unavailable");
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
      predicateId,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    const presence = page.getByTestId(workbookPresenceSummaryTestId());
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
    );
    await clickTimelineRowAction(
      page,
      record.record_id,
      rowHistoryOpenButtonTestId(record.record_id),
    );
    for (const field of contract.workload.fields)
      await scrollGridCellIntoView({
        page,
        surface: contract.view_schema_id,
        recordId: record.record_id,
        cellKey: field.field_key,
      });
    await openHistoryEventDetails(page, selected.history_item_ref);
    await page
      .getByTestId(
        rowHistoryActionTestId({
          action: kind,
          historyItemRef: selected.history_item_ref,
        }),
      )
      .click();
    const confirm = page.getByTestId(
      rowHistoryRollbackConfirmButtonTestId({
        action: kind,
        historyItemRef: selected.history_item_ref,
      }),
    );
    await confirm.scrollIntoViewIfNeeded();
    await expect(confirm).toBeVisible();
    await expect(confirm).toBeEnabled();
    for (const [index, field] of contract.workload.fields.entries()) {
      const cell = page.getByTestId(
        rowCellTestId(record.record_id, field.field_key),
      );
      await expect(cell).toHaveText(
        kind === "history_entry" && index === 1 ? field.before : field.mistake,
      );
      expect(
        await cell.evaluate((node) => {
          const b = node.getBoundingClientRect();
          const grid = node.closest('[role="grid"], [role="treegrid"]');
          if (grid === null) return false;
          const g = grid.getBoundingClientRect();
          return (
            b.left >= Math.max(0, g.left) &&
            b.right <= Math.min(innerWidth, g.right) &&
            b.top >= Math.max(0, g.top) &&
            b.bottom <= Math.min(innerHeight, g.bottom)
          );
        }),
      ).toBe(true);
    }
    await confirm.focus();
    await expect(confirm).toBeFocused();
    target = {
      kind,
      recordId: record.record_id,
      actorId: actor.user_id,
      priorVersion: current.row_version,
    };
    completion = await measureHistory(
      page,
      {
        ...target,
        priorHistoryRefs: before.items.map((item) => item.history_item_ref),
        priorChangeSetIds: before.items.map((item) => item.change_set_id),
        gridSelector: dataTestIdSelector(
          gridShellTestId(contract.view_schema_id),
        ),
        historySelector: dataTestIdSelector(rowHistoryPanelTestId()),
        cells: contract.workload.fields.map((field) => ({
          selector: dataTestIdSelector(
            rowCellTestId(record.record_id, field.field_key),
          ),
          fieldKey: field.field_key,
          value: field.before,
        })),
      },
      () => page.keyboard.press("Enter"),
    );
    await expect(presence).toHaveAttribute(
      "aria-label",
      /^24 collaborators present on this sheet/,
    );
    const after = await fetchFullRecordHistory(page, record.record_id);
    const appended = after.items.filter(
      (item) =>
        !before.items.some(
          (old) => old.history_item_ref === item.history_item_ref,
        ),
    );
    expect(appended).toHaveLength(1);
    expect(appended[0]).toMatchObject({
      history_item_ref: completion.historyRef,
      change_set_id: completion.changeSetId,
      actor_user_id: actor.user_id,
      revision_no: completion.rowVersion,
      committed_at: completion.committedAt,
      operation: kind === "row_restore" ? "row_restore" : "rollback",
    });
    const restored = (
      await queryViewRows(page, snapshot.incidentId, contract.view_schema_id)
    ).find((row) => row.record_id === record.record_id);
    expect(restored?.row_version).toBe(current.row_version + 1);
    for (const field of contract.workload.fields)
      expect(restored?.cells[field.field_key]?.value).toBe(field.before);
    await traffic.stop();
    traffic = null;
    await attachHistoryObservation(info, predicateId, target, completion);
    attached = true;
    expect(completion.elapsed).toBeLessThan(contract.threshold_ms);
  } catch (error) {
    if (!attached)
      await attachHistoryObservation(
        info,
        predicateId,
        target,
        completion,
        "History measurement did not complete",
      );
    throw error;
  } finally {
    await traffic?.stop();
  }
}

test("measures Timeline single-entry rollback through visible row and history within AC-011", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runHistoryMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "history_entry",
  );
});
test("measures Timeline change-set rollback through visible row and history within AC-011", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runHistoryMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "change_set",
  );
});
test("measures Timeline whole-row restore through visible row and history within AC-011", async ({
  browser,
  page,
  workerAdmin,
  sessionTracker,
}, info) => {
  await runHistoryMeasurement(
    browser,
    page,
    workerAdmin,
    sessionTracker,
    info,
    "row_restore",
  );
});
