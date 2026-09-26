import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  entityInspectButtonTestId,
  gridShellTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryItemTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryReadControlTestId,
  rowHistoryRollbackConfirmButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  fetchFullRecordHistory,
  fetchRecordHistoryPage,
  openHistoryEventDetails,
} from "./support/workbook/history";
import { createViewRow, patchRecord } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import {
  clickTimelineRowAction,
  openGenericInspectorForRecord,
} from "./support/workbook/rowMutations";

type Surface = "Timeline" | "Generic" | "Entity" | "Assessment";
async function prepare(page: Page, surface: Surface) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("HISTORY-RECOVERY"),
    `History recovery ${surface}`,
  );
  const host =
    surface === "Assessment" || surface === "Entity"
      ? await createViewRow(page, incidentId, hostsViewSchemaId, {
          client_txn_id: uniqueTxn("history-host"),
          "host.display_name": "History recovery host",
          "host.hostname": "history-recovery.example.test",
        })
      : null;
  const viewSchemaId =
    surface === "Timeline"
      ? timelineViewSchemaId
      : surface === "Generic"
        ? evidenceViewSchemaId
        : surface === "Entity"
          ? hostsViewSchemaId
          : assessmentsViewSchemaId;
  const row =
    surface === "Entity" && host
      ? host
      : await createViewRow(page, incidentId, viewSchemaId, {
          client_txn_id: uniqueTxn("history-row"),
          ...(surface === "Timeline"
            ? { "timeline.activity_synopsis_text": "History recovery row" }
            : surface === "Generic"
              ? {
                  "evidence.collector_party_text": "History recovery collector",
                  "evidence.title": "History recovery row",
                }
              : {
                  "assessment.subject_ref": host?.record_id,
                  "assessment.subject_type": "host",
                  "assessment.assessment_state": "confirmed",
                  "assessment.confidence_score": 85,
                  "assessment.rationale": "History recovery rationale",
                  "assessment.assessed_at": "2026-09-10T00:00:00Z",
                }),
        });
  const field =
    surface === "Timeline"
      ? "timeline.activity_synopsis_text"
      : surface === "Generic"
        ? "evidence.title"
        : surface === "Entity"
          ? "host.display_name"
          : "assessment.rationale";
  let version = row.row_version;
  const production = await fetchRecordHistoryPage(page, row.record_id);
  expect(production.meta.paging.limit).toBe(100);
  for (let index = 0; index < production.meta.paging.limit + 2; index++) {
    if (surface === "Assessment") {
      const response = await publicHttpOperation({
        operationID: index % 2 === 0 ? "deleteRecord" : "restoreRecord",
        pathParameters: { record_id: row.record_id },
        body: {
          base_row_version: version,
          client_txn_id: uniqueTxn("assessment-history-lifecycle"),
          reason: "Retained lifecycle browsing fixture",
        },
        headers: await csrfHeaders(page),
        request: atJsonOrigin(page.request, apiBase),
      });
      expect(response.ok).toBe(true);
      if (!response.ok) throw new Error("Assessment lifecycle fixture failed");
      version = response.payload.data.row_version;
      continue;
    }
    const updated = await patchRecord(page, row.record_id, {
      view_schema_id: viewSchemaId,
      base_row_version: version,
      client_txn_id: uniqueTxn("overflow-edit"),
      changes: [{ field_key: field, value: `Retained entry ${index}` }],
    });
    version = updated.row_version;
  }
  const first = await fetchRecordHistoryPage(page, row.record_id);
  expect(first.meta.paging.has_more).toBe(true);
  expect(first.data.items).toHaveLength(production.meta.paging.limit);
  const complete = await fetchFullRecordHistory(page, row.record_id);
  expect(complete.items.length).toBeGreaterThan(production.meta.paging.limit);
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(viewSchemaId)}`,
  );
  await expect(page.getByTestId(gridShellTestId(viewSchemaId))).toBeVisible();
  if (surface === "Timeline")
    await clickTimelineRowAction(
      page,
      row.record_id,
      rowHistoryOpenButtonTestId(row.record_id),
    );
  else {
    if (surface === "Generic")
      await openGenericInspectorForRecord(page, viewSchemaId, row.record_id);
    else if (surface === "Entity") {
      const targetTestId = entityInspectButtonTestId("host", row.record_id);
      await scrollGridTargetIntoView({
        page,
        surface: viewSchemaId,
        targetTestId,
      });
      await page.getByTestId(targetTestId).click();
    } else {
      await page
        .getByTestId(
          rowCellTestId(row.record_id, "assessment.assessment_state"),
        )
        .click();
      await page
        .getByTestId(workbookInspectorToggleTestId(viewSchemaId))
        .click();
    }
    await page
      .getByRole("button", { name: "Open history", exact: true })
      .click();
  }
  await expect(page.getByTestId(rowHistoryDeleteButtonTestId())).toBeVisible();
  return { incidentId, row, viewSchemaId, first, complete };
}

async function browse(page: Page, surface: Surface) {
  test.setTimeout(180_000);
  const { row, first, complete } = await prepare(page, surface);
  const panel = page.getByTestId(rowHistoryPanelTestId());
  const firstItem = panel.getByTestId(
    rowHistoryItemTestId({
      historyItemRef: required(first.data.items[0]).history_item_ref,
    }),
  );
  const older = panel.getByTestId(rowHistoryReadControlTestId("load-older"));
  await expect(firstItem).toBeVisible();
  const observed: string[] = [];
  let fail = true;
  let invalid = false;
  let releaseRetry = () => {};
  let retryHeld = false;
  const retryGate = new Promise<void>((resolve) => {
    releaseRetry = resolve;
  });
  await page.route(
    `**/api/v1/records/${row.record_id}/history?*`,
    async (route) => {
      observed.push(
        new URL(route.request().url()).searchParams.get("cursor_token") ?? "",
      );
      if (invalid) {
        invalid = false;
        await route.fulfill({
          status: 400,
          contentType: "application/json",
          body: JSON.stringify({
            error: {
              code: "invalid_pagination_request",
              status: 400,
              request_id: "cursor-recovery",
              retryable: false,
              message: "Invalid cursor",
              details: { reason_code: "invalid_cursor_token" },
            },
          }),
        });
      } else if (fail) {
        fail = false;
        await route.abort("failed");
      } else if (surface === "Timeline" && observed.length === 2) {
        const response = await route.fetch();
        retryHeld = true;
        await retryGate;
        await route.fulfill({ response });
      } else await route.continue();
    },
  );
  await older.focus();
  const anchor = await older.boundingBox();
  await older.press("Enter");
  await expect(panel).toContainText("History could not be read");
  await expect(
    panel.getByRole("status").filter({ hasText: "History could not be read" }),
  ).toHaveCount(1);
  await expect(firstItem).toBeVisible();
  await expect(older).toBeFocused();
  const retry = panel.getByTestId(rowHistoryReadControlTestId("retry"));
  await retry.focus();
  await retry.press("Enter");
  if (surface === "Timeline") {
    try {
      await expect.poll(() => retryHeld).toBe(true);
      await expect(retry).toBeAttached();
      await expect(retry).toBeFocused();
      await expect(retry).toHaveAttribute("aria-busy", "true");
      await expect(retry).toHaveAttribute("aria-disabled", "true");
      await page.keyboard.press("Tab");
      await expect(retry).not.toBeFocused();
      await page.keyboard.press("Shift+Tab");
      await expect(retry).toBeFocused();
    } finally {
      releaseRetry();
    }
  }
  await expect(panel).toContainText("No older entries.");
  expect(observed).toHaveLength(2);
  expect(observed[0]).not.toBe("");
  expect(observed[1]).toBe(observed[0]);
  await expect(older).toBeFocused();
  expect(
    Math.abs(required(await older.boundingBox()).y - required(anchor).y),
  ).toBeLessThan(2);
  for (const item of complete.items)
    await expect(
      panel.getByTestId(
        rowHistoryItemTestId({ historyItemRef: item.history_item_ref }),
      ),
    ).toBeAttached();
  const expectedOrder = complete.items.map((item) =>
    rowHistoryItemTestId({ historyItemRef: item.history_item_ref }),
  );
  const order = await panel
    .locator("[data-testid]")
    .evaluateAll(
      (nodes, expected) =>
        nodes
          .map((node) => node.getAttribute("data-testid"))
          .filter((id) => id !== null && expected.includes(id)),
      expectedOrder,
    );
  expect(order).toEqual(expectedOrder);
  await test.info().attach(`overflow-${surface}`, {
    body: JSON.stringify({
      first: first.meta.paging,
      retainedCount: complete.items.length,
      identities: complete.items.map((item) => item.history_item_ref),
      observed,
    }),
    contentType: "application/json",
  });

  let failRefresh = true;
  let refreshRetryHeld = false;
  let releaseRefreshRetry = () => {};
  const refreshRetryGate = new Promise<void>((resolve) => {
    releaseRefreshRetry = resolve;
  });
  let refreshReads = 0;
  await page.route(
    `**/api/v1/records/${row.record_id}/history`,
    async (route) => {
      refreshReads += 1;
      if (failRefresh) {
        failRefresh = false;
        await route.abort("failed");
      } else if (refreshReads === 2) {
        const response = await route.fetch();
        refreshRetryHeld = true;
        await refreshRetryGate;
        await route.fulfill({ response });
      } else await route.continue();
    },
  );
  await panel
    .getByRole("button", { name: "Refresh history", exact: true })
    .click();
  await expect(panel).toContainText("History could not be read");
  const last = required(complete.items.at(-1));
  await expect(
    panel.getByTestId(
      rowHistoryItemTestId({ historyItemRef: last.history_item_ref }),
    ),
  ).toBeAttached();
  const refreshRetry = panel.getByTestId(rowHistoryReadControlTestId("retry"));
  await refreshRetry.focus();
  await refreshRetry.press("Enter");
  try {
    await expect.poll(() => refreshRetryHeld).toBe(true);
    await expect(refreshRetry).toBeAttached();
    await expect(refreshRetry).toBeFocused();
    await expect(refreshRetry).toHaveAttribute("aria-busy", "true");
    await expect(refreshRetry).toHaveAttribute("aria-disabled", "true");
    await refreshRetry.press("Enter");
    expect(refreshReads).toBe(2);
  } finally {
    releaseRefreshRetry();
  }
  await expect(panel).not.toContainText("History could not be read");
  await expect(
    panel.getByRole("button", { name: "Refresh history", exact: true }),
  ).toBeFocused();
  await expect(
    panel.getByTestId(
      rowHistoryItemTestId({ historyItemRef: last.history_item_ref }),
    ),
  ).toHaveCount(0);
  // Replacing an in-flight continuation cancels its presentation anchor.
  // Its eventual response must neither append the old page nor move focus/viewport.
  let releaseOlder = () => {};
  let heldOlder = false;
  let releasedOlder = false;
  const olderGate = new Promise<void>((resolve) => {
    releaseOlder = resolve;
  });
  const continuationPattern = `**/api/v1/records/${row.record_id}/history?*`;
  const holdOlder = async (route: Route) => {
    const response = await route.fetch();
    heldOlder = true;
    await olderGate;
    await route.fulfill({ response });
    releasedOlder = true;
  };
  await page.route(continuationPattern, holdOlder);
  try {
    await older.focus();
    await older.press("Enter");
    await expect.poll(() => heldOlder).toBe(true);
    const refresh = panel.getByRole("button", {
      name: "Refresh history",
      exact: true,
    });
    await refresh.focus();
    await refresh.press("Enter");
    await expect(refresh).toHaveAttribute("aria-disabled", "false");
    await expect(firstItem).toBeAttached();
    const refreshedAnchor = required(await refresh.boundingBox());
    releaseOlder();
    await expect.poll(() => releasedOlder).toBe(true);
    await expect(refresh).toBeFocused();
    expect(
      Math.abs(required(await refresh.boundingBox()).y - refreshedAnchor.y),
    ).toBeLessThan(2);
    await expect(
      panel.getByTestId(
        rowHistoryItemTestId({ historyItemRef: last.history_item_ref }),
      ),
    ).toHaveCount(0);
    await expect(older).toHaveAttribute("aria-disabled", "false");
  } finally {
    releaseOlder();
    await page.unroute(continuationPattern, holdOlder);
  }
  invalid = true;
  await older.click();
  await expect(
    panel.getByRole("button", { name: "Start fresh history", exact: true }),
  ).toBeVisible();
  await expect(firstItem).toBeAttached();
  let freshHeld = false;
  let releaseFresh = () => {};
  const freshGate = new Promise<void>((resolve) => {
    releaseFresh = resolve;
  });
  let freshCursor: string | null = null;
  const holdFresh = async (route: Route) => {
    freshCursor = new URL(route.request().url()).searchParams.get(
      "cursor_token",
    );
    const response = await route.fetch();
    freshHeld = true;
    await freshGate;
    await route.fulfill({ response });
  };
  const freshPattern = `**/api/v1/records/${row.record_id}/history`;
  await page.route(freshPattern, holdFresh);
  const startFresh = panel.getByTestId(
    rowHistoryReadControlTestId("start-fresh"),
  );
  await startFresh.focus();
  await startFresh.press("Enter");
  try {
    await expect.poll(() => freshHeld).toBe(true);
    expect(freshCursor).toBeNull();
    await expect(startFresh).toBeAttached();
    await expect(startFresh).toBeFocused();
    await expect(startFresh).toHaveAttribute("aria-busy", "true");
    await expect(startFresh).toHaveAttribute("aria-disabled", "true");
    await expect(firstItem).toBeAttached();
  } finally {
    releaseFresh();
  }
  await expect(panel).not.toContainText("History could not be read");
  await page.unroute(freshPattern, holdFresh);
  await older.click();
  await expect(panel).toContainText("No older entries.");
  const target = required(
    complete.items.find(
      (item) =>
        !first.data.items.some(
          (head) => head.history_item_ref === item.history_item_ref,
        ) && item.available_rollback_actions.includes("row_restore"),
    ),
  );
  expect(target).toBeDefined();
  await openHistoryEventDetails(panel, target.history_item_ref);
  await panel
    .getByTestId(
      rowHistoryActionTestId({
        action: "row_restore",
        historyItemRef: target.history_item_ref,
      }),
    )
    .click();

  const bodies: string[] = [];
  let observeCommit = () => {};
  const firstCommit = new Promise<void>((resolve) => {
    observeCommit = resolve;
  });
  await page.route(
    `**/api/v1/records/${row.record_id}/rollback`,
    async (route) => {
      bodies.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      if (surface === "Timeline" && bodies.length === 1)
        await route.abort("failed");
      else await route.fulfill({ response });
      observeCommit();
    },
  );
  await panel
    .getByTestId(
      rowHistoryRollbackConfirmButtonTestId({
        action: "row_restore",
        historyItemRef: target.history_item_ref,
      }),
    )
    .click();
  await firstCommit;
  expect(JSON.parse(required(bodies[0]))).toMatchObject({
    base_row_version: complete.row_version,
    target: { kind: "row_restore", restore_to_revision_no: target.revision_no },
  });
  if (surface === "Timeline") {
    const committed = await fetchFullRecordHistory(page, row.record_id);
    expect(committed.items.length).toBeGreaterThan(100);
    await openRecoveryItem(
      page,
      /^(Soft-delete row|Restore[^·]*|Reverse[^·]*) ·/,
    );
    const recovery = page.getByRole("region", {
      name: "History action recovery",
      exact: true,
    });
    await expect(recovery).toContainText("Outcome unknown");
    await recovery.getByRole("button", { name: "Replay exact action" }).click();
    await expect(recovery).toContainText("Action completed.");
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    expect(await fetchFullRecordHistory(page, row.record_id)).toEqual(
      committed,
    );
  }
}

test("Timeline browses production history overflow with keyboard recovery and later rollback", async ({
  page,
}) => browse(page, "Timeline"));
test("Generic browses production history overflow with keyboard recovery and later rollback", async ({
  page,
}) => browse(page, "Generic"));
test("Entity browses production history overflow with keyboard recovery and later rollback", async ({
  page,
}) => browse(page, "Entity"));
test("Assessment browses production history overflow with keyboard recovery and later rollback", async ({
  page,
}) => browse(page, "Assessment"));

test("Timeline retains initial History retry focus in a narrow Inspector", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("HISTORY-INITIAL-RETRY"),
    "Initial History retry",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("initial-retry-row"),
    "timeline.activity_synopsis_text": "Initial History retry row",
  });
  const original = await fetchFullRecordHistory(page, row.record_id);
  let reads = 0;
  let held = false;
  let release = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (
      request.url().includes(`/api/v1/records/${row.record_id}`) &&
      request.method() !== "GET"
    )
      mutations.push(request.method());
  });
  const initialPattern = `**/api/v1/records/${row.record_id}/history`;
  await page.route(initialPattern, async (route) => {
    reads += 1;
    if (reads === 1) {
      await route.abort("failed");
      return;
    }
    const response = await route.fetch();
    held = true;
    await gate;
    await route.fulfill({ response });
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${timelineViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await clickTimelineRowAction(
    page,
    row.record_id,
    rowHistoryOpenButtonTestId(row.record_id),
  );
  const panel = page.getByTestId(rowHistoryPanelTestId());
  await expect(panel).toContainText("History could not be read");
  const retry = panel.getByRole("button", {
    name: "Retry history",
    exact: true,
  });
  await retry.focus();
  await retry.press("Enter");
  try {
    await expect.poll(() => held).toBe(true);
    await expect(retry).toBeAttached();
    await expect(retry).toBeFocused();
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await expect(retry).toBeInViewport();
    await page.keyboard.press("Tab");
    await expect(retry).not.toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(retry).toBeFocused();
    expect(reads).toBe(2);
    expect(mutations).toEqual([]);
  } finally {
    release();
  }
  await expect(
    panel.getByRole("button", { name: "Refresh history", exact: true }),
  ).toBeFocused();
  await expect(retry).toHaveCount(0);
  await page.unroute(initialPattern);
  let refreshHeld = false;
  let releaseRefresh = () => {};
  const refreshGate = new Promise<void>((resolve) => {
    releaseRefresh = resolve;
  });
  const holdRefresh = async (route: Route) => {
    const response = await route.fetch();
    refreshHeld = true;
    await refreshGate;
    await route.fulfill({ response });
  };
  await page.route(initialPattern, holdRefresh);
  const refresh = panel.getByRole("button", {
    name: "Refresh history",
    exact: true,
  });
  await refresh.focus();
  await refresh.press("Enter");
  try {
    await expect.poll(() => refreshHeld).toBe(true);
    const summary = panel
      .getByTestId(
        rowHistoryItemTestId({
          historyItemRef: required(original.items[0]).history_item_ref,
        }),
      )
      .getByText("Event details", { exact: true });
    await summary.click();
    await expect(summary).toBeFocused();
    const scrollPosition = () =>
      summary.evaluate((element) => {
        let parent = element.parentElement;
        while (
          parent &&
          !/(auto|scroll)/.test(getComputedStyle(parent).overflowY)
        )
          parent = parent.parentElement;
        return parent?.scrollTop ?? 0;
      });
    const beforeWheel = await scrollPosition();
    await page.mouse.wheel(0, 120);
    await expect.poll(scrollPosition).toBeGreaterThan(beforeWheel);
    const position = await scrollPosition();
    releaseRefresh();
    await expect(panel).not.toContainText("Refreshing…");
    await expect(summary).toBeFocused();
    expect(await scrollPosition()).toBe(position);
  } finally {
    releaseRefresh();
    await page.unroute(initialPattern, holdRefresh);
  }
  expect(await fetchFullRecordHistory(page, row.record_id)).toEqual(original);
  expect(mutations).toEqual([]);
});

function required<T>(value: T | undefined | null): T {
  if (value === undefined || value === null)
    throw new Error("Missing history fixture value");
  return value;
}
