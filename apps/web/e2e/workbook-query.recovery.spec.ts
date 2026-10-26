import type {
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import { sortByHeader } from "@cartulary/test-utils/grid";
import {
  gridRowTestId,
  rowCellTestId,
  workbookSortMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";

import { createIncident } from "./support/incidents/fixtures";

import {
  type BarrierOutcome,
  responseBarrier,
} from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { observePageRequests } from "./support/runtime/observations";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import { createViewRow, patchRecord } from "./support/workbook/query";

import { seed } from "./support/workbook/queryBrowsing";

async function reachButtonByTab(page: Page, button: Locator) {
  if (
    await page.evaluate(() =>
      Boolean(document.activeElement?.closest('[role="grid"]')),
    )
  )
    await page.keyboard.press("Control+End");
  for (let step = 0; step < 160; step++) {
    if (await button.evaluate((element) => document.activeElement === element))
      return;
    await page.keyboard.press("Tab");
  }
  throw new Error("Keyboard Tab did not reach the workbook browsing action");
}

async function expectVisibleKeyboardFocus(button: Locator) {
  await expect(button).toBeFocused();
  expect(
    await button.evaluate(
      (element) =>
        element.matches(":focus-visible") &&
        getComputedStyle(element).outlineStyle !== "none" &&
        Number.parseFloat(getComputedStyle(element).outlineWidth) > 0,
    ),
  ).toBe(true);
}

test("Workbook real-route recovery preserves accepted labels and reconciles live placement without draining pages", async ({
  page,
  testActor,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-LIVE"),
    "Workbook live continuation recovery",
  );
  await seed(page, incident, view, 405, testActor.user_id);
  const reads: {
    request: QueryWorkbookViewRequest;
    response?: QueryWorkbookViewResponse;
    status: number;
    delivered?: boolean;
  }[] = [];
  let mode: "normal" | "abort" | "malformed" | "invalid_cursor" | "hold" =
    "normal";
  let held = false;
  let release: () => void = () => {};
  let gate: Promise<BarrierOutcome> = Promise.resolve("released");
  const path = `**/incidents/${incident}/views/${view}/query`;
  await installOwnedRoute(page, path, async (route) => {
    const request = route.request().postDataJSON() as QueryWorkbookViewRequest;
    const behavior = mode;
    mode = "normal";
    if (behavior === "abort") {
      reads.push({ request, status: 0 });
      await route.abort("failed");
      return;
    }
    const response = await route.fetch(
      behavior === "invalid_cursor"
        ? { postData: { ...request, cursor_token: "not-a-core-cursor" } }
        : {},
    );
    const payload = await response.json();
    const read = {
      request,
      status: response.status(),
      ...(response.ok() ? { response: payload } : {}),
      delivered: false,
    };
    reads.push(read);
    if (behavior === "hold") {
      held = true;
      if (!(await awaitResponseBarrier(route, gate))) return;
    }
    if (behavior === "malformed") {
      delete payload.meta.paging;
      await route.fulfill({ response, json: payload });
    } else await route.fulfill({ response });
    // route.fetch also completes for reads cancelled during startup. Only a
    // response delivered to the page can own the next continuation token.
    const delivered = await route.request().response();
    read.delivered =
      delivered !== null && (await delivered.finished()) === null;
  });
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await sockets.waitForAcceptedSocket();
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(more).toBeEnabled();
  await expect.poll(() => reads.some((read) => read.delivered)).toBe(true);
  const first = reads.filter((read) => read.delivered).at(-1)?.response;
  if (!first) throw new Error("Missing delivered first page");
  mode = "abort";
  const failedIndex = reads.length;
  await more.evaluate((button) => {
    (button as HTMLButtonElement).click();
    (button as HTMLButtonElement).click();
  });
  await expect(
    controls.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  expect(reads.length).toBe(failedIndex + 1);
  expect(reads.at(-1)?.request.cursor_token).toBe(
    first.meta.paging.next_cursor,
  );
  await expect(controls).toContainText("100 records loaded");
  const failed = reads.at(-1)?.request;
  await controls.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(controls).toContainText("200 records loaded; more available.");
  expect(reads.at(-1)?.request).toEqual(failed);
  mode = "invalid_cursor";
  const invalidIndex = reads.length;
  await more.click();
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.slice(invalidIndex).map((read) => read.status)).toEqual([
    400, 200,
  ]);
  expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
  mode = "malformed";
  await more.click();
  await expect(
    controls.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await expect(controls).toContainText("100 records loaded");
  await controls.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(controls).toContainText("200 records loaded; more available.");
  mode = "abort";
  await sortByHeader(page, view, "note.title");
  await expect(controls).toContainText("Query changes are unapplied.");
  await expect(
    page.getByTestId(workbookSortMenuTriggerTestId(view)),
  ).toHaveAttribute("aria-label", "Sort, no user sorts");
  await expect(controls).toContainText("200 records loaded");
  await controls.getByRole("button", { name: "Revert", exact: true }).click();
  await expect(controls).not.toContainText("Query changes are unapplied.");
  await expect(controls).toContainText("200 records loaded; more available.");
  const liveStart = reads.length;
  const inserted = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("wqc-live-insert"),
    "note.title": "Live inserted",
    "note.body": "Live source",
  });
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Live inserted");
  await expect(controls).toContainText("200 records loaded; more available.");
  expect(reads.length - liveStart).toBeLessThanOrEqual(4);
  const mutate = async (
    operationID: "deleteRecord" | "restoreRecord",
    version: number,
  ) =>
    publicHttpOperation({
      operationID,
      pathParameters: { record_id: inserted.record_id },
      headers: await csrfHeaders(page),
      request: atJsonOrigin(page.request, apiBase),
      body: {
        client_txn_id: uniqueTxn(operationID),
        base_row_version: version,
        reason: "Workbook query reconciliation",
      },
    });
  const deleted = await mutate("deleteRecord", 1);
  expect(deleted.ok).toBe(true);
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveCount(0);
  expect((await mutate("restoreRecord", 2)).ok).toBe(true);
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Live inserted");
  ({ promise: gate, release: release } = responseBarrier(page));
  mode = "hold";
  await controls.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect.poll(() => held).toBe(true);
  await patchRecord(page, inserted.record_id, {
    client_txn_id: uniqueTxn("wqc-newer"),
    base_row_version: 3,
    view_schema_id: view,
    changes: [{ field_key: "note.title", value: "Newer accepted live source" }],
  });
  release();
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Newer accepted live source");
  await expect(
    page.getByTestId(gridRowTestId(view, inserted.record_id)),
  ).toHaveAttribute("data-grid-row-version", "4");
  await expect(controls).toContainText("100 records loaded; more available.");
});

test("Workbook query recovery retains keyboard focus through Timeline replacement and continuation", async ({
  page,
  testActor,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-FOCUS"),
    "Workbook query recovery focus",
  );
  await seed(page, incident, timelineViewSchemaId, 105, testActor.user_id);
  const writes: string[] = [];
  observePageRequests(page, (request) => {
    if (
      ["PATCH", "PUT", "DELETE"].includes(request.method()) &&
      request.url().includes("/records/")
    )
      writes.push(`${request.method()} ${request.url()}`);
    if (
      request.method() === "POST" &&
      request.url().includes(`/incidents/${incident}/views/`) &&
      request.url().endsWith("/rows")
    )
      writes.push(`POST ${request.url()}`);
  });
  const reads: QueryWorkbookViewRequest[] = [];
  let next: "ordinary" | "fail" | "hold_failure" | "hold_success" = "ordinary";
  let release = () => {};
  let cancel = () => {};
  let gate: Promise<BarrierOutcome> = Promise.resolve("released");
  const hold = (outcome: "hold_failure" | "hold_success") => {
    ({
      promise: gate,
      release: release,
      cancel: cancel,
    } = responseBarrier(page));
    next = outcome;
  };
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const behavior = next;
      next = "ordinary";
      reads.push(request);
      if (behavior === "fail") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      if (behavior === "hold_failure" || behavior === "hold_success") {
        if (!(await awaitResponseBarrier(route, gate))) return;
        if (behavior === "hold_failure") {
          await route.abort("failed");
          return;
        }
      }
      await route.fulfill({ response });
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const retry = controls.getByRole("button", { name: "Retry", exact: true });
  const revert = controls.getByRole("button", { name: "Revert", exact: true });
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  try {
    next = "fail";
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await expect(retry).toBeVisible();
    await expect(revert).toBeVisible();
    const failedReplacement = reads.at(-1);
    expect(failedReplacement?.cursor_token).toBeUndefined();
    expect(failedReplacement?.limit).toBe(100);
    expect(failedReplacement?.sort?.[0]?.field_key).toBe(
      "timeline.activity_synopsis_text",
    );
    await expect(controls).toContainText("100 records loaded");
    await reachButtonByTab(page, retry);
    hold("hold_failure");
    const firstRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(firstRetry + 1);
    expect(reads.at(-1)).toEqual(failedReplacement);
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await expect(controls).toContainText("Retrying records…");
    await expect(controls).not.toContainText("Query changes are unapplied.");
    await page.keyboard.press("Enter");
    expect(reads.length).toBe(firstRetry + 1);
    release();
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "false");
    await expect(controls).toContainText("Query changes are unapplied.");

    hold("hold_success");
    const secondRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(secondRetry + 1);
    expect(reads.at(-1)).toEqual(failedReplacement);
    await expect(retry).toBeFocused();
    release();
    await expect(controls).toContainText("100 records loaded; more available.");
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Tab");
    await expect(retry).toHaveCount(0);

    await reachButtonByTab(page, more);
    next = "fail";
    await page.keyboard.press("Enter");
    await expect(retry).toBeVisible();
    const failedContinuation = reads.at(-1);
    expect(failedContinuation?.cursor_token).toBeTruthy();
    expect(failedContinuation?.limit).toBe(100);
    expect(failedContinuation?.sort).toEqual(failedReplacement?.sort);
    await expect(controls).toContainText("100 records loaded");
    await reachButtonByTab(page, retry);
    hold("hold_success");
    const continuationRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(continuationRetry + 1);
    expect(reads.at(-1)).toEqual(failedContinuation);
    await expectVisibleKeyboardFocus(retry);
    await expect(controls).toContainText("Retrying records…");
    await page.keyboard.press("Shift+Tab");
    await expect(refresh).toBeFocused();
    release();
    await expect(controls).toContainText(
      "105 records loaded; end of current results.",
    );
    await expect(refresh).toBeFocused();
    await expect(retry).toHaveCount(0);

    next = "fail";
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await expect(revert).toBeVisible();
    await reachButtonByTab(page, revert);
    hold("hold_success");
    const reverted = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(reverted + 1);
    await expectVisibleKeyboardFocus(revert);
    await expect(revert).toHaveAttribute("aria-disabled", "true");
    release();
    await expect(controls).toContainText(
      "105 records loaded; end of current results.",
    );
    await expectVisibleKeyboardFocus(revert);
    await page.keyboard.press("Tab");
    await expect(revert).toHaveCount(0);
    expect(writes).toEqual([]);
  } finally {
    cancel();
  }
});

test("Workbook query recovery keeps keyboard focus with accepted-empty Notes", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-EMPTY-FOCUS"),
    "Workbook empty recovery focus",
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(notesViewSchemaId)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText(
    "0 records loaded; end of current results.",
  );
  let next: "ordinary" | "fail" | "hold" = "ordinary";
  let release = () => {};
  let cancel = () => {};
  let gate: Promise<BarrierOutcome> = Promise.resolve("released");
  const reads: QueryWorkbookViewRequest[] = [];
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${notesViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const behavior = next;
      next = "ordinary";
      reads.push(request);
      if (behavior === "fail") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      if (behavior === "hold")
        if (!(await awaitResponseBarrier(route, gate))) return;
      await route.fulfill({ response });
    },
  );
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  const retry = controls.getByRole("button", { name: "Retry", exact: true });
  try {
    await reachButtonByTab(page, refresh);
    next = "fail";
    await page.keyboard.press("Enter");
    await expect(retry).toBeVisible();
    const failed = reads.at(-1);
    expect(failed?.cursor_token).toBeUndefined();
    await expect(controls).toContainText("0 records loaded");
    await reachButtonByTab(page, retry);
    ({
      promise: gate,
      release: release,
      cancel: cancel,
    } = responseBarrier(page));
    next = "hold";
    const count = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(count + 1);
    expect(reads.at(-1)).toEqual(failed);
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(controls).toContainText("0 records loaded. Retrying records…");
    release();
    await expect(controls).toContainText(
      "0 records loaded; end of current results.",
    );
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Tab");
    await expect(retry).toHaveCount(0);
  } finally {
    cancel();
  }
});
