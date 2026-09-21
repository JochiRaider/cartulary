import { Buffer } from "node:buffer";
import type { AppendIndicatorStateIntervalResponse } from "@cartulary/protocol-ts/http";
import {
  indicatorLifecycleTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryRollbackConfirmButtonTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  indicatorsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { uniqueTxn } from "./support/runtime/fixtureIdentity";
import {
  fetchFullRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import {
  appendLifecycleInterval,
  createLifecycleFixture,
  listLifecycleIntervals,
  openLifecycleEditor,
} from "./support/workbook/indicatorLifecycle";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import {
  activateCommittedGridCell,
  openGenericInspectorForRecord,
} from "./support/workbook/rowMutations";

test("Indicator intervals recover a response lost after commit with the original receipt and History rollback", async ({
  page,
}, testInfo) => {
  const { incidentId, indicator } = await createLifecycleFixture(page);
  const requests: string[] = [],
    receipts: AppendIndicatorStateIntervalResponse["data"][] = [];
  await page.route(
    `**/api/v1/indicators/${indicator.record_id}/state-intervals`,
    async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.status()).toBe(requests.length === 1 ? 201 : 200);
      receipts.push((await response.json()).data);
      if (requests.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  const editor = await openLifecycleEditor(
    page,
    incidentId,
    indicator.record_id,
  );
  await editor
    .getByLabel("Effective from (UTC)", { exact: true })
    .fill("2026-09-11T12:00");
  await editor.getByLabel("Confidence (optional)", { exact: true }).fill("0");
  await editor
    .getByLabel("Rationale (optional)", { exact: true })
    .fill("Reviewed supporting evidence");
  await editor
    .getByRole("button", { name: "Append lifecycle interval", exact: true })
    .click();
  await expect(
    page.getByText(
      "The interval outcome is unknown. The server may have saved it. Your original request is retained.",
      { exact: true },
    ),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(indicatorsViewSchemaId))
    .click();
  const before = {
    intervals: (await listLifecycleIntervals(page, indicator.record_id)).data
      .intervals,
    history: await fetchFullRecordHistory(page, indicator.record_id),
  };
  expect(before.intervals).toHaveLength(1);
  let failRefresh = true;
  await page.route(
    `**/views/${indicatorsViewSchemaId}/query`,
    async (route) => {
      if (!failRefresh) return route.continue();
      failRefresh = false;
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "service_unavailable",
            status: 503,
            retryable: true,
            message: "Refresh unavailable",
            request_id: "refresh-failure",
          },
        }),
      });
    },
  );
  await openRecoveryItem(page, /^Indicator interval ·/);
  const recovery = page.getByTestId(indicatorLifecycleTestId("recovery"));
  await recovery
    .getByRole("button", { name: "Replay original interval request" })
    .click();
  await expect(
    recovery.getByText("Interval saved. Refresh is still required.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  expect(receipts[1]).toEqual({ ...receipts[0], replayed: true });
  await recovery
    .getByRole("button", { name: "Retry interval refresh" })
    .click();
  await expect(
    recovery.getByText("Interval saved. Indicator and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  const after = {
    intervals: (await listLifecycleIntervals(page, indicator.record_id)).data
      .intervals,
    history: await fetchFullRecordHistory(page, indicator.record_id),
  };
  expect(after).toEqual(before);
  const receipt = receipts[0];
  if (!receipt) throw new Error("receipt");
  expect(
    new Set(
      after.history.items
        .filter((item) => item.change_set_id === receipt.change_set_id)
        .map((item) => item.change_set_id),
    ).size,
  ).toBe(1);
  await page.getByRole("button", { name: "Close recovery" }).click();
  // Use the existing History preview and rollback; the interval API has no delete.
  await openLifecycleEditor(page, incidentId, indicator.record_id);
  await expect(
    page.getByRole("article", {
      name: "active interval from 2026-09-11T12:00:00Z",
    }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Open history", exact: true }).click();
  const historyItem = after.history.items.find(
    (item) =>
      item.change_set_id === receipt.change_set_id &&
      item.available_rollback_actions.includes("change_set"),
  );
  if (!historyItem) throw new Error("rollback item");
  const anchor = {
    action: "change_set" as const,
    historyItemRef: historyItem.history_item_ref,
  };
  await openHistoryEventDetails(page, historyItem.history_item_ref);
  await page.getByTestId(rowHistoryActionTestId(anchor)).click();
  await page.getByTestId(rowHistoryRollbackConfirmButtonTestId(anchor)).click();
  await expect
    .poll(
      async () =>
        (await listLifecycleIntervals(page, indicator.record_id)).data.intervals
          .length,
    )
    .toBe(0);
  const rolledBack = await fetchFullRecordHistory(page, indicator.record_id);
  expect(
    rolledBack.items.some(
      (item) => item.change_set_id === receipt.change_set_id,
    ),
  ).toBe(true);
  await testInfo.attach("indicator-interval-recovery-evidence", {
    body: Buffer.from(
      JSON.stringify(
        { requests, receipts, before, after, rolledBack },
        null,
        2,
      ),
    ),
    contentType: "application/json",
  });
});

test("Indicator interval paging recovers failures and backdated appends preserve ordered support and collections", async ({
  page,
}) => {
  const { incidentId, indicator } = await createLifecycleFixture(page);
  let version = indicator.row_version;
  const expectedIds: string[] = [];
  for (let n = 0; n < 101; n++) {
    const receipt = await appendLifecycleInterval(page, indicator.record_id, {
      client_txn_id: uniqueTxn("seed-interval"),
      base_row_version: version,
      lifecycle_state: "benign",
      valid_from: "2026-09-11T12:00:00Z",
      valid_to: "2026-09-11T12:00:00Z",
      confidence: 100,
      rationale: null,
      assessor: null,
      support_refs: [],
    });
    version = receipt.affected_records[0]?.row_version ?? 0;
    expectedIds.push(receipt.interval.interval_id);
  }
  for (let n = 0; n < 101; n += 5)
    await Promise.all(
      Array.from({ length: Math.min(5, 101 - n) }, (_, offset) =>
        createViewRow(page, incidentId, timelineViewSchemaId, {
          client_txn_id: uniqueTxn("support"),
          "timeline.activity_synopsis_text": `Supporting event ${n + offset}`,
        }),
      ),
    );
  let failPage = true;
  await page.route(
    `**/api/v1/indicators/${indicator.record_id}/state-intervals?*`,
    async (route) => {
      if (
        failPage &&
        new URL(route.request().url()).searchParams.has("cursor_token")
      ) {
        failPage = false;
        return route.abort("failed");
      }
      return route.continue();
    },
  );
  const editor = await openLifecycleEditor(
    page,
    incidentId,
    indicator.record_id,
  );
  const collection = page.getByTestId(indicatorLifecycleTestId("intervals"));
  await expect(collection.getByRole("article")).toHaveCount(100);
  await collection
    .getByRole("button", { name: "Load more lifecycle intervals" })
    .click();
  await expect(
    collection.getByRole("button", { name: "Retry lifecycle intervals" }),
  ).toBeVisible();
  await collection
    .getByRole("button", { name: "Retry lifecycle intervals" })
    .click();
  await expect(collection.getByRole("article")).toHaveCount(101);
  const support = page.getByTestId(indicatorLifecycleTestId("support"));
  await expect(support.getByRole("checkbox")).toHaveCount(100);
  await support
    .getByRole("button", { name: "Load more supporting records" })
    .click();
  await expect(support.getByRole("checkbox")).toHaveCount(101);
  const selected = [
    await support.getByRole("checkbox").last().inputValue(),
    await support.getByRole("checkbox").first().inputValue(),
  ];
  await support.getByRole("checkbox").last().check();
  await support.getByRole("checkbox").first().check();
  await editor
    .getByLabel("Effective from (UTC)", { exact: true })
    .fill("2020-01-01T00:00");
  await editor.getByLabel("Confidence (optional)", { exact: true }).fill("101");
  await editor
    .getByRole("button", { name: "Append lifecycle interval", exact: true })
    .click();
  await expect(
    editor.getByLabel("Confidence (optional)", { exact: true }),
  ).toHaveAttribute("aria-invalid", "true");
  await editor.getByLabel("Confidence (optional)", { exact: true }).fill("0");
  const sent: string[] = [];
  await page.route(
    `**/api/v1/indicators/${indicator.record_id}/state-intervals`,
    async (route) => {
      if (route.request().method() === "POST")
        sent.push(route.request().postData() ?? "");
      return route.continue();
    },
  );
  await editor
    .getByRole("button", { name: "Append lifecycle interval", exact: true })
    .click();
  await expect(
    collection.getByText("Interval saved. Indicator and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(sent).toHaveLength(1);
  expect(JSON.parse(sent[0] ?? "{}").support_refs).toEqual(selected);
  await expect(collection.getByRole("article")).toHaveCount(102);
  await expect(collection.getByRole("article").last()).toHaveAccessibleName(
    "active interval from 2020-01-01T00:00:00Z",
  );
  const live = (await listLifecycleIntervals(page, indicator.record_id)).data
    .intervals;
  expect(live.map((item) => item.interval_id)).toEqual(
    expectedIds.sort().reverse().slice(0, 100),
  );
});

test("Indicator recovery fences late malformed receipts and replays the original request after incident closure", async ({
  page,
}) => {
  const { incidentId, indicator } = await createLifecycleFixture(page);
  const other = await createViewRow(page, incidentId, indicatorsViewSchemaId, {
    client_txn_id: uniqueTxn("second-indicator"),
    "indicator.indicator_type": "ipv4_addr",
    "indicator.value_kind": "atomic",
    "indicator.display_value": "203.0.113.9",
  });
  let release = () => {};
  const delivery = new Promise<void>((resolve) => {
    release = resolve;
  });
  const requests: string[] = [];
  let committed: AppendIndicatorStateIntervalResponse["data"] | null = null;
  await page.route(
    `**/api/v1/indicators/${indicator.record_id}/state-intervals`,
    async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      requests.push(route.request().postData() ?? "");
      const response = await route.fetch();
      const payload = await response.json();
      committed = payload.data;
      if (requests.length === 1) {
        await delivery;
        return route.fulfill({
          response,
          json: {
            ...payload,
            data: {
              ...payload.data,
              interval: {
                ...payload.data.interval,
                assessor: "Unexpected assessor",
              },
            },
          },
        });
      }
      return route.fulfill({ response });
    },
  );
  let editor = await openLifecycleEditor(page, incidentId, indicator.record_id);
  await editor
    .getByLabel("Effective from (UTC)", { exact: true })
    .fill("2026-09-11T12:00");
  await editor
    .getByRole("button", { name: "Append lifecycle interval", exact: true })
    .click();
  await expect.poll(() => committed !== null).toBe(true);
  const cell = page
    .getByTestId(rowCellTestId(other.record_id, "indicator.indicator_type"))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await activateCommittedGridCell(cell);
  editor = page.getByTestId(indicatorLifecycleTestId("editor"));
  const otherFrom = editor.getByLabel("Effective from (UTC)", { exact: true });
  await otherFrom.fill("2030-01-01T00:00");
  await otherFrom.focus();
  const received = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/indicators/${indicator.record_id}/state-intervals`),
  );
  release();
  await received;
  await expect(otherFrom).toBeFocused();
  await openRecoveryItem(page, /^Indicator interval ·/);
  const recovery = page.getByTestId(indicatorLifecycleTestId("recovery"));
  await expect(
    recovery.getByText(
      "The interval outcome is unknown. The server may have saved it. Your original request is retained.",
      { exact: true },
    ),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close recovery" }).click();
  await openGenericInspectorForRecord(
    page,
    indicatorsViewSchemaId,
    other.record_id,
  );
  await expect(otherFrom).toHaveValue("2030-01-01T00:00");
  const lifecycle = await currentLifecycle(page, incidentId);
  expect(
    (
      await lifecycleAction(page, incidentId, "closeIncident", {
        base_incident_version: lifecycle.incident_version,
        client_txn_id: uniqueTxn("close-after-append"),
        reason: "Investigation paused",
      })
    ).ok,
  ).toBe(true);
  await expect(editor).toHaveCount(0);
  await openRecoveryItem(page, /^Indicator interval ·/);
  await recovery
    .getByRole("button", { name: "Replay original interval request" })
    .click();
  await expect(
    recovery.getByText("Interval saved. Indicator and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  expect(
    (await listLifecycleIntervals(page, indicator.record_id)).data.intervals,
  ).toHaveLength(1);
  expect(
    (await listLifecycleIntervals(page, other.record_id)).data.intervals,
  ).toHaveLength(0);
});
