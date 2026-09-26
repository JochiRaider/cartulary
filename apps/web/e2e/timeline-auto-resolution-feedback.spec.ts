import {
  applyFilterChip,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  autoResolutionNoticeFamilySelector,
  dataTestIdSelector,
  gridRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  incidentLandingTestId,
  mentionItemTestId,
  relationshipItemsTestId,
  rowCellTestId,
  surfaceTabTestId,
  timelineCollectionInputTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Route, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { revokeAllSessions } from "./support/auth/sessions";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import {
  addRelationshipTokenViaUI,
  aliasCollectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { openTimelineInspector } from "./support/workbook/rowMutations";

type DiagnosticWindow = Window & {
  arf: {
    counts: Record<string, number>;
    owner: { acceptVersion(id: string, version: number): void } | null;
  };
};

test("Timeline Review read cannot interrupt newer scalar editing", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ARF-REVIEW"),
    "Review read continuity",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-host"),
    "host.display_name": "Review canonical host",
    "host.hostname": "review.example.test",
    "host.aliases": aliasCollectionActionsPayload(["review alias"]),
  });
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-source"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Original review source",
  });
  const editing = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-editing"),
    "timeline.activity_utc_text": "2026-04-02T00:00:00Z",
    "timeline.activity_synopsis_text": "Editable destination",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [{ op: "add_tag", tag_name: "review-editing" }],
    },
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Hosts"]);
  const resolution = await addRelationshipTokenViaUI(
    page,
    source.record_id,
    "hostRefs",
    "review alias",
  );
  const itemRef = String(
    collectionItems(resolution.data.row, hostRefsFieldKey).find(
      (item) => item.raw_text === "review alias",
    )?.item_ref ?? "",
  );
  expect(itemRef).not.toBe("");
  const notices = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  await expect(notices).toHaveCount(1);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    "review-editing",
  );
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
  ).toHaveCount(0);
  await expect(notices).toHaveCount(1);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, editing.record_id)),
  ).toBeVisible();

  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/query`,
  });
  let reviewWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/records\/|\/entity-mentions\//u.test(request.url())
    )
      reviewWrites++;
  });
  try {
    const review = notices
      .first()
      .getByRole("button", { name: "Review", exact: true });
    await review.click();
    await held.waitForHit;
    await expect(review).toHaveAttribute("aria-busy", "true");
    await expect(review).toBeEnabled();
    await review.press("Enter");
    const cell = page.getByTestId(
      rowCellTestId(editing.record_id, "timeline.activity_synopsis_text"),
    );
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        editing.record_id,
        "timeline.activity_synopsis_text",
      ),
    });
    await cell.click();
    const input = page.getByTestId(
      timelineScalarEditorTestId({
        recordId: editing.record_id,
        fieldKey: "timeline.activity_synopsis_text",
        surface: "grid",
      }),
    );
    await input.fill("newer native draft Ω");
    await input.evaluate((element: HTMLInputElement) =>
      element.setSelectionRange(2, 7, "backward"),
    );
    const original = await input.elementHandle();
    held.release();
    await expect.poll(() => held.hitCount()).toBe(1);
    await expect(input).toBeFocused();
    await expect(input).toHaveValue("newer native draft Ω");
    expect(
      await original?.evaluate((element: HTMLInputElement) => ({
        connected: element.isConnected,
        focused: document.activeElement === element,
        start: element.selectionStart,
        end: element.selectionEnd,
        direction: element.selectionDirection,
      })),
    ).toEqual({
      connected: true,
      focused: true,
      start: 2,
      end: 7,
      direction: "backward",
    });
    await expect(notices).toHaveCount(1);
    expect(held.hitCount()).toBe(1);
    expect(reviewWrites).toBe(0);
    await input.press("Escape");
    await review.click();
    await expect(page.getByTestId(mentionItemTestId(itemRef))).toBeFocused();
    await expect(notices).toHaveCount(1);
    await expect(
      page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
    ).toHaveCount(0);
    await expect(
      page.getByTestId(gridRowTestId(timelineViewSchemaId, editing.record_id)),
    ).toBeVisible();
    expect(reviewWrites).toBe(0);
  } finally {
    await held.dispose();
  }
});

test("Timeline Review read preserves narrow collection editing", async ({
  page,
}) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 768, height: 800 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ARF-REVIEW-COLLECTION"),
    "Review collection continuity",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-collection-host"),
    "host.display_name": "Collection canonical host",
    "host.hostname": "collection-review.example.test",
    "host.aliases": aliasCollectionActionsPayload(["collection alias"]),
  });
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-collection-source"),
    "timeline.activity_synopsis_text": "Collection review source",
  });
  const editing = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-collection-editing"),
    "timeline.activity_synopsis_text": "Collection editing row",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [{ op: "add_tag", tag_name: "review-collection-editing" }],
    },
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Hosts", "Tags"]);
  await addRelationshipTokenViaUI(
    page,
    source.record_id,
    "hostRefs",
    "collection alias",
  );
  const notices = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  await expect(notices).toHaveCount(1);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    "review-collection-editing",
  );
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, editing.record_id)),
  ).toBeVisible();
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/query`,
  });
  let reviewWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/records\/|\/entity-mentions\//u.test(request.url())
    )
      reviewWrites++;
  });
  try {
    await notices
      .first()
      .getByRole("button", { name: "Review", exact: true })
      .click();
    await held.waitForHit;
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(
        editing.record_id,
        "timeline.tags",
        "grid",
      ),
    });
    const cell = page
      .getByRole("group", { name: "Tags collection cell", exact: true })
      .filter({
        has: page.getByTestId(
          relationshipItemsTestId(editing.record_id, "timeline.tags", "grid"),
        ),
      });
    await cell
      .getByRole("button", { name: "Add tags token", exact: true })
      .click();
    const input = page.getByTestId(
      timelineCollectionInputTestId(editing.record_id, "timeline.tags", "grid"),
    );
    await input.fill("newer collection draft Ω");
    await input.evaluate((element: HTMLInputElement) =>
      element.setSelectionRange(2, 8, "backward"),
    );
    const original = await input.elementHandle();
    held.release();
    await expect(input).toBeFocused();
    await expect(input).toHaveValue("newer collection draft Ω");
    expect(
      await original?.evaluate((element: HTMLInputElement) => ({
        connected: element.isConnected,
        focused: document.activeElement === element,
        start: element.selectionStart,
        end: element.selectionEnd,
        direction: element.selectionDirection,
      })),
    ).toEqual({
      connected: true,
      focused: true,
      start: 2,
      end: 8,
      direction: "backward",
    });
    await expect(notices).toHaveCount(1);
    expect(reviewWrites).toBe(0);
  } finally {
    await held.dispose();
  }
});

test("Timeline Review read failure retries at the original source mention", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ARF-REVIEW-RETRY"),
    "Review retry",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-retry-host"),
    "host.display_name": "Retry canonical host",
    "host.hostname": "retry-review.example.test",
    "host.aliases": aliasCollectionActionsPayload(["retry alias"]),
  });
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-retry-source"),
    "timeline.activity_synopsis_text": "Retry review source",
  });
  await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-review-retry-visible"),
    "timeline.activity_synopsis_text": "Visible retry row",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [{ op: "add_tag", tag_name: "review-retry-visible" }],
    },
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Hosts"]);
  const resolution = await addRelationshipTokenViaUI(
    page,
    source.record_id,
    "hostRefs",
    "retry alias",
  );
  const itemRef = String(
    collectionItems(resolution.data.row, hostRefsFieldKey).find(
      (item) => item.raw_text === "retry alias",
    )?.item_ref ?? "",
  );
  expect(itemRef).not.toBe("");
  const notices = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  await expect(notices).toHaveCount(1);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    "review-retry-visible",
  );
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
  ).toHaveCount(0);
  let readCount = 0;
  let reviewWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/records\/|\/entity-mentions\//u.test(request.url())
    )
      reviewWrites++;
  });
  const queryPath = `**/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/query`;
  await page.route(queryPath, async (route) => {
    if (route.request().method() !== "POST") return route.fallback();
    readCount++;
    if (readCount === 1) return route.abort("failed");
    return route.continue();
  });
  try {
    const review = notices
      .first()
      .getByRole("button", { name: "Review", exact: true });
    await review.click();
    await expect(notices.first().getByRole("alert")).toHaveText(
      /Could not open the source\. Review again to retry\./u,
    );
    await expect(review).toBeEnabled();
    await review.click();
    await expect(page.getByTestId(mentionItemTestId(itemRef))).toBeFocused();
    await expect(notices.first().getByRole("alert")).toHaveCount(0);
    await expect(notices).toHaveCount(1);
    expect(readCount).toBe(2);
    expect(reviewWrites).toBe(0);
  } finally {
    await page.unroute(queryPath);
  }
});

async function prepareUndoContinuity(page: Page) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ARF-UNDO-CONTINUITY"),
    "Undo continuity",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-undo-host"),
    "host.display_name": "Undo canonical host",
    "host.hostname": "undo-continuity.example.test",
    "host.aliases": aliasCollectionActionsPayload(["undo continuity alias"]),
  });
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-undo-source"),
    "timeline.activity_synopsis_text": "Undo continuity source",
  });
  await createTimelineFillers(page, incident, "undo continuity filler", 24);
  const editing = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-undo-editing"),
    "timeline.activity_synopsis_text": "Continue after Undo",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [{ op: "add_tag", tag_name: "undo-continuity-editing" }],
    },
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Hosts"]);
  const resolution = await addRelationshipTokenViaUI(
    page,
    source.record_id,
    "hostRefs",
    "undo continuity alias",
  );
  const itemRef = String(
    collectionItems(resolution.data.row, hostRefsFieldKey).find(
      (item) => item.raw_text === "undo continuity alias",
    )?.item_ref ?? "",
  );
  expect(itemRef).not.toBe("");
  const notice = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  await expect(notice).toHaveCount(1);
  return { incident, source, editing, itemRef, notice: notice.first() };
}

test("Timeline auto-resolution feedback retry Undo retains keyboard position through a held replay", async ({
  page,
}) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 768, height: 640 });
  const { notice } = await prepareUndoContinuity(page);
  const path = "**/api/v1/entity-mentions/*/resolve";
  const bodies: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      /\/entity-mentions\/.*\/resolve$/u.test(request.url())
    )
      bodies.push(request.postData() ?? "");
  });
  const failFirst = async (route: Route) => {
    await route.abort("failed");
  };
  await page.route(path, failFirst);
  const undo = notice.getByRole("button", { name: "Undo", exact: true });
  try {
    await undo.focus();
    await undo.press("Enter");
    await expect(
      notice.getByRole("button", { name: "Retry Undo", exact: true }),
    ).toBeVisible();
    await page.unroute(path, failFirst);
    const held = await holdBrowserRequest(page, {
      method: "POST",
      path: "/api/v1/entity-mentions/*/resolve",
    });
    try {
      const retry = notice.getByRole("button", {
        name: "Retry Undo",
        exact: true,
      });
      await retry.focus();
      await retry.press("Enter");
      await held.waitForHit;
      await expect(retry).toBeFocused();
      await expect(retry).toHaveAttribute("aria-busy", "true");
      await retry.press("Enter");
      expect(held.hitCount()).toBe(1);
      await page.keyboard.press("Shift+Tab");
      await expect(
        notice.getByRole("button", { name: "Review" }),
      ).toBeFocused();
      await page.keyboard.press("Tab");
      await expect(retry).toBeFocused();
      held.release();
      await expect(notice).toHaveCount(0);
      expect(bodies).toHaveLength(2);
      expect(bodies[1]).toBe(bodies[0]);
    } finally {
      await held.dispose();
    }
  } finally {
    await page.unroute(path, failFirst);
  }
});

test("Timeline auto-resolution feedback owned Undo restores a visible source destination", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice, source } = await prepareUndoContinuity(page);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    await expect(undo).toBeFocused();
    held.release();
    await expect(notice).toHaveCount(0);
    const sourceCell = page.getByRole("gridcell").filter({
      has: page.getByTestId(
        rowCellTestId(source.record_id, "timeline.activity_synopsis_text"),
      ),
    });
    await expect(sourceCell).toBeFocused();
    await expect(sourceCell).toBeInViewport();
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback owned Undo falls back when its source is off-query", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice, source, editing } = await prepareUndoContinuity(page);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    "undo-continuity-editing",
  );
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, editing.record_id)),
  ).toBeVisible();
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    await expect(undo).toBeFocused();
    held.release();
    await expect(notice).toHaveCount(0);
    await expect
      .poll(() =>
        page.evaluate(
          (selector) => !!document.activeElement?.closest(selector),
          dataTestIdSelector(gridShellTestId(timelineViewSchemaId)),
        ),
      )
      .toBe(true);
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback undo late acceptance cannot reclaim newer scroll or body focus", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice } = await prepareUndoContinuity(page);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    await expect(undo).toBeFocused();
    await expect(undo).toHaveAttribute("aria-busy", "true");
    const grid = page.locator(gridScrollportSelector());
    await grid.hover();
    await page.mouse.wheel(0, 36);
    await expect
      .poll(() => grid.evaluate((element) => element.scrollTop))
      .toBeGreaterThan(0);
    await notice.locator("strong").first().click();
    await page.mouse.click(1, 1);
    await expect
      .poll(() => page.evaluate(() => document.activeElement?.tagName))
      .toBe("BODY");
    const ownedScroll = await grid.evaluate((element) => element.scrollTop);
    held.release();
    await expect(notice).toHaveCount(0);
    await expect
      .poll(() => grid.evaluate((element) => element.scrollTop))
      .toBe(ownedScroll);
    expect(await page.evaluate(() => document.activeElement?.tagName)).toBe(
      "BODY",
    );
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback undo late acceptance preserves same-row Inspector focus", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice, itemRef } = await prepareUndoContinuity(page);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    await notice.getByRole("button", { name: "Review", exact: true }).click();
    const item = page.getByTestId(mentionItemTestId(itemRef));
    await expect(item).toBeFocused();
    const grid = page.locator(gridScrollportSelector());
    const newerScroll = await grid.evaluate((element) => element.scrollTop);
    held.release();
    await expect(notice).toHaveCount(0);
    await expect(item).toBeFocused();
    const settledScroll = await grid.evaluate((element) => ({
      top: element.scrollTop,
      maximum: element.scrollHeight - element.clientHeight,
    }));
    expect(settledScroll.top).toBe(
      Math.min(newerScroll, settledScroll.maximum),
    );
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback undo late acceptance preserves native scalar editing", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice, editing } = await prepareUndoContinuity(page);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    const cellId = rowCellTestId(
      editing.record_id,
      "timeline.activity_synopsis_text",
    );
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: cellId,
    });
    await page.getByTestId(cellId).click();
    const input = page.getByTestId(
      timelineScalarEditorTestId({
        recordId: editing.record_id,
        fieldKey: "timeline.activity_synopsis_text",
        surface: "grid",
      }),
    );
    await input.fill("newer Undo editing Ω");
    await input.evaluate((element: HTMLInputElement) =>
      element.setSelectionRange(2, 7, "backward"),
    );
    const original = await input.elementHandle();
    const grid = page.locator(gridScrollportSelector());
    const newerScroll = await grid.evaluate((element) => element.scrollTop);
    held.release();
    await expect(notice).toHaveCount(0);
    await expect(input).toBeFocused();
    expect(await grid.evaluate((element) => element.scrollTop)).toBe(
      newerScroll,
    );
    expect(
      await original?.evaluate((element: HTMLInputElement) => ({
        connected: element.isConnected,
        focused: document.activeElement === element,
        text: element.value,
        start: element.selectionStart,
        end: element.selectionEnd,
        direction: element.selectionDirection,
      })),
    ).toEqual({
      connected: true,
      focused: true,
      text: "newer Undo editing Ω",
      start: 2,
      end: 7,
      direction: "backward",
    });
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback undo late acceptance stays detached after surface navigation", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const { notice } = await prepareUndoContinuity(page);
  const held = await holdBrowserRequest(page, {
    method: "POST",
    path: "/api/v1/entity-mentions/*/resolve",
  });
  try {
    const undo = notice.getByRole("button", { name: "Undo", exact: true });
    await undo.focus();
    await undo.press("Enter");
    await held.waitForHit;
    const hostsTab = page.getByTestId(surfaceTabTestId(hostsViewSchemaId));
    await hostsTab.click();
    await expect(hostsTab).toHaveAttribute("aria-current", "page");
    held.release();
    await expect(hostsTab).toHaveAttribute("aria-current", "page");
    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    await expect(notice).toHaveCount(0);
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
});

test("Timeline auto-resolution feedback production characterization compact", async ({
  page,
  workerAdmin,
}, testInfo) => characterize(page, testInfo, workerAdmin.user_id, "compact"));
test("Timeline auto-resolution feedback production characterization default", async ({
  page,
  workerAdmin,
}, testInfo) => characterize(page, testInfo, workerAdmin.user_id, "default"));
test("Timeline auto-resolution feedback production characterization comfortable", async ({
  page,
  workerAdmin,
}, testInfo) =>
  characterize(page, testInfo, workerAdmin.user_id, "comfortable"));

async function characterize(
  page: Page,
  testInfo: TestInfo,
  userId: string,
  density: "compact" | "default" | "comfortable",
) {
  const preferences = await installVisualPreferences(page, userId);
  preferences.select(density);
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  // Diagnostic observations of production React work; never used by timing gates.
  await page.addInitScript(() => {
    type Fiber = {
      type?: unknown;
      flags: number;
      child?: Fiber;
      sibling?: Fiber;
      memoizedProps?: Record<string, unknown> & {
        binding?: { kind?: string };
        actions?: {
          owner?: {
            acceptVersion?: unknown;
            getSnapshot?: () => { mentions?: unknown };
          };
        };
        owner?: {
          acceptVersion?: unknown;
          getSnapshot?: () => { mentions?: unknown };
        };
      };
      alternate?: Fiber;
    };
    const observed = window as unknown as DiagnosticWindow & {
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    observed.arf = { counts: {}, owner: null };
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        const count = (key: string) => {
          observed.arf.counts[key] = (observed.arf.counts[key] ?? 0) + 1;
        };
        count("commits");
        const current = new WeakSet<object>();
        const visit = (fiber?: Fiber) => {
          if (!fiber) return;
          current.add(fiber);
          const props = fiber.memoizedProps;
          const owner = props?.owner ?? props?.actions?.owner;
          if (
            typeof owner?.acceptVersion === "function" &&
            owner.getSnapshot?.().mentions
          )
            observed.arf.owner = owner as NonNullable<
              DiagnosticWindow["arf"]["owner"]
            >;
          if (!previous.has(fiber)) {
            if (
              typeof fiber.type === "function" &&
              (fiber.flags & 1) !== 0 &&
              props?.binding?.kind === "collection"
            )
              count("collectionRenders");
            for (const key of ["columns", "rows"])
              if (
                Array.isArray(props?.[key]) &&
                fiber.alternate &&
                props?.[key] !== fiber.alternate.memoizedProps?.[key]
              )
                count(`${key}Replacements`);
          }
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
        previous = current;
      },
    };
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ARF"),
    "Auto-resolution feedback characterization",
  );
  const longAlias = `gateway-${"longsegment".repeat(14)}.example.test`;
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-host"),
    "host.display_name": "Canonical gateway node",
    "host.hostname": "arf-gateway.example.test",
    "host.aliases": aliasCollectionActionsPayload(["VPN Gateway", longAlias]),
  });
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-source"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Source-bound automatic matches",
  });
  await createTimelineFillers(page, incidentId, "arf-filler", 32, {
    occurredAtStart: "2026-04-02T00:00:00Z",
  });
  const writes: string[] = [];
  page.on("request", (request) => {
    const path = new URL(request.url()).pathname;
    if (
      request.method() !== "GET" &&
      /\/records\/|\/entity-mentions\//.test(path)
    )
      writes.push(path);
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Hosts", "Tags"]);
  const notices = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  const settle = () =>
    page.evaluate(
      () =>
        new Promise<void>((resolve) =>
          requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
        ),
    );
  const observations: unknown[] = [];
  const capture = async (name: string) => {
    await settle();
    const geometry = await page.evaluate(
      ({ noticeSelector, gridSelector, scrollSelector, inspectorSelector }) => {
        const rect = (element: Element | null) =>
          element?.getBoundingClientRect().toJSON() ?? null;
        const grid = document.querySelector(gridSelector);
        const scroll = document.querySelector<HTMLElement>(scrollSelector);
        const active = document.activeElement;
        return {
          viewport: {
            width: innerWidth,
            height: innerHeight,
            zoom: document.documentElement.style.zoom,
          },
          grid: rect(grid),
          notices: [...document.querySelectorAll(noticeSelector)]
            .filter((element) => element.querySelector("button"))
            .map((element) => ({
              rect: rect(element),
              text: element.textContent,
            })),
          inspector: rect(document.querySelector(inspectorSelector)),
          focused:
            active?.getAttribute("data-testid") ??
            active?.getAttribute("aria-label") ??
            active?.tagName,
          editor:
            active instanceof HTMLInputElement ||
            active instanceof HTMLTextAreaElement
              ? {
                  text: active.value,
                  start: active.selectionStart,
                  end: active.selectionEnd,
                  direction: active.selectionDirection,
                  rect: rect(active),
                }
              : null,
          scroll: { top: scroll?.scrollTop, left: scroll?.scrollLeft },
          work: { ...(window as unknown as DiagnosticWindow).arf.counts },
          documentOverflow:
            document.documentElement.scrollWidth > innerWidth ||
            document.documentElement.scrollHeight > innerHeight,
        };
      },
      {
        noticeSelector: autoResolutionNoticeFamilySelector(),
        gridSelector: dataTestIdSelector(gridShellTestId(timelineViewSchemaId)),
        scrollSelector: gridScrollportSelector(),
        inspectorSelector: dataTestIdSelector(
          workbookShellSlotTestId("inspector"),
        ),
      },
    );
    observations.push({ name, writes: writes.length, ...geometry });
    await testInfo.attach(name, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  };
  await addRelationshipTokenViaUI(
    page,
    row.record_id,
    "hostRefs",
    "VPN Gateway",
  );
  await expect(notices).toHaveCount(1);
  await capture("one-disclosure-inspector-open");
  for (const raw of ["VPN Gateway", "VPN Gateway", longAlias])
    await addRelationshipTokenViaUI(page, row.record_id, "hostRefs", raw);
  await expect(notices).toHaveCount(4);
  await expect(notices.first()).toContainText("Canonical gateway node");
  expect(writes).toHaveLength(4);
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  const bulkCheckbox = page.getByRole("checkbox", {
    name: `Select record ${row.record_id}`,
    exact: true,
  });
  await bulkCheckbox.check();
  const bulkTag = page.getByRole("textbox", {
    name: "Tag for selected Timeline records",
  });
  await bulkTag.fill("Retained tag beside auto-resolution " + longAlias);
  for (const profile of [
    {
      name: "base-closed",
      width: 1440,
      height: 900,
      inspector: false,
      zoom: "1",
    },
    { name: "base-open", width: 1440, height: 900, inspector: true, zoom: "1" },
    {
      name: "base-minimum",
      width: 1440,
      height: 900,
      inspector: true,
      zoom: "1",
      resize: "Home",
    },
    {
      name: "base-maximum",
      width: 1440,
      height: 900,
      inspector: true,
      zoom: "1",
      resize: "End",
    },
    {
      name: "narrow-open",
      width: 1024,
      height: 720,
      inspector: true,
      zoom: "1",
    },
    {
      name: "compact-open",
      width: 768,
      height: 640,
      inspector: true,
      zoom: "1",
    },
    {
      name: "compact-closed",
      width: 768,
      height: 640,
      inspector: false,
      zoom: "1",
    },
    {
      name: "text-spacing",
      width: 1024,
      height: 720,
      inspector: false,
      zoom: "1",
    },
    {
      name: "zoom-closed",
      width: 1440,
      height: 900,
      inspector: false,
      zoom: "2",
    },
  ]) {
    if (await close.count()) await close.click();
    await page.setViewportSize({
      width: profile.width,
      height: profile.height,
    });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = zoom;
    }, profile.zoom);
    await page.evaluate((spaced) => {
      document.body.style.letterSpacing = spaced ? "0.12em" : "";
      document.body.style.wordSpacing = spaced ? "0.16em" : "";
    }, profile.name === "text-spacing");
    if (profile.inspector) await openTimelineInspector(page, row.record_id);
    if (profile.resize)
      await page
        .getByRole("separator", { name: "Resize inspector" })
        .press(profile.resize);
    await expect(bulkCheckbox).toBeChecked();
    await expect(bulkTag).toHaveValue(
      "Retained tag beside auto-resolution " + longAlias,
    );
    await bulkTag.scrollIntoViewIfNeeded();
    await bulkTag.click({ trial: true });
    const disclosureRegion = page.getByRole("complementary", {
      name: "Auto-resolution disclosures",
    });
    const bounds = await disclosureRegion.boundingBox();
    const gridBounds = await page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .boundingBox();
    expect((bounds?.y ?? 0) + (bounds?.height ?? 0)).toBeLessThanOrEqual(
      gridBounds?.y ?? 0,
    );
    for (const notice of await notices.all()) {
      const review = notice.getByRole("button", {
        name: "Review",
        exact: true,
      });
      await review.scrollIntoViewIfNeeded();
      await review.click({ trial: true });
      await notice
        .getByRole("button", { name: "Undo", exact: true })
        .click({ trial: true });
    }
    if (profile.width === 768 || profile.zoom === "2") {
      await capture(`${profile.name}-long-detail`);
      await disclosureRegion.evaluate((element) => {
        element.scrollTop = element.scrollHeight;
      });
      await capture(`${profile.name}-long-detail-end`);
    }
    await disclosureRegion.evaluate((element) => {
      element.scrollTop = 0;
    });
    await capture(profile.name);
  }
  await page
    .getByRole("button", { name: "Clear tag draft", exact: true })
    .click();
  await bulkCheckbox.uncheck();
  await page.evaluate(() => {
    document.documentElement.style.zoom = "1";
  });
  await page.setViewportSize({ width: 1440, height: 900 });
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(
      row.record_id,
      "timeline.tags",
      "grid",
    ),
  });
  const cell = page
    .getByRole("group", { name: "Tags collection cell", exact: true })
    .filter({
      has: page.getByTestId(
        relationshipItemsTestId(row.record_id, "timeline.tags", "grid"),
      ),
    });
  await cell
    .getByRole("button", { name: "Add tags token", exact: true })
    .click();
  const input = page.getByTestId(
    timelineCollectionInputTestId(row.record_id, "timeline.tags", "grid"),
  );
  await input.fill("continued native authoring Ω");
  await input.evaluate((element: HTMLInputElement) =>
    element.setSelectionRange(3, 8, "backward"),
  );
  const original = await input.elementHandle();
  await capture("continued-typing");
  const before = await page.evaluate(() => ({
    ...(window as unknown as DiagnosticWindow).arf.counts,
  }));
  const ownerObserved = await page.evaluate(() => {
    const owner = (window as unknown as DiagnosticWindow).arf.owner;
    // Diagnostic version bookkeeping for an unrelated identity, without a write.
    owner?.acceptVersion("arf-diagnostic-unrelated-source", 1);
    return owner !== null;
  });
  await settle();
  const after = await page.evaluate(() => ({
    ...(window as unknown as DiagnosticWindow).arf.counts,
  }));
  observations.push({
    name: "unrelated-owner-publication",
    ownerObserved,
    delta: Object.fromEntries(
      Object.keys(after).map((key) => [
        key,
        (after[key] ?? 0) - (before[key] ?? 0),
      ]),
    ),
    native: await original?.evaluate((element: HTMLInputElement) => ({
      connected: element.isConnected,
      focused: document.activeElement === element,
      text: element.value,
      start: element.selectionStart,
      end: element.selectionEnd,
    })),
  });
  expect(writes).toHaveLength(4);
  await input.fill("");
  await input.press("Escape");
  await notices
    .first()
    .getByRole("button", { name: "Review", exact: true })
    .click();
  await expect(notices).toHaveCount(4);
  const correctedId = await notices.nth(1).getAttribute("data-testid");
  const mentionBodies: string[] = [];
  let failRefresh = density === "default";
  const queryPath = `**/views/${timelineViewSchemaId}/query`;
  const mentionPath = "**/entity-mentions/*/resolve";
  if (density === "default") {
    await page.route(queryPath, (route) =>
      failRefresh ? route.abort("failed") : route.continue(),
    );
    await page.route(mentionPath, async (route) => {
      mentionBodies.push(route.request().postData() ?? "");
      if (mentionBodies.length === 1) return route.abort("failed");
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      await route.fulfill({ response });
    });
  }
  await notices
    .nth(1)
    .getByRole("button", { name: "Undo", exact: true })
    .click();
  if (density === "default") {
    await expect(
      notices.nth(1).getByRole("button", { name: "Retry Undo", exact: true }),
    ).toBeVisible();
    await expect(notices).toHaveCount(4);
    await notices
      .nth(1)
      .getByRole("button", { name: "Retry Undo", exact: true })
      .click();
  }
  await expect(notices).toHaveCount(3);
  if (density === "default") {
    await openRecoveryItem(page, /^Mention resolution ·/);
    const recovery = page.getByRole("region", {
      name: "Retained mention operations",
    });
    await expect(recovery).toContainText(
      "Mention action completed; refresh is still required.",
    );
    failRefresh = false;
    await recovery
      .getByRole("button", { name: "Refresh mention result" })
      .click();
    await expect(recovery).toContainText("Mention action completed.");
    expect(mentionBodies).toHaveLength(2);
    expect(mentionBodies[1]).toBe(mentionBodies[0]);
    await recovery.press("Escape");
    await page.unroute(queryPath);
    await page.unroute(mentionPath);
  }
  expect(
    await notices
      .all()
      .then((items) =>
        Promise.all(items.map((item) => item.getAttribute("data-testid"))),
      ),
  ).not.toContain(correctedId);
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
  await expect(notices).toHaveCount(3);
  await expect(notices.first()).toContainText("Canonical gateway node");
  await testInfo.attach("auto-resolution-observations", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
}

test("Timeline disclosure arrival preserves scalar collection and native composition editing", async ({
  page,
}, testInfo) => {
  test.setTimeout(180_000);
  for (const mode of ["scalar", "collection", "composition"] as const) {
    await page.setViewportSize({ width: 1440, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("ARF-ARRIVAL"),
      "Disclosure arrival continuity",
    );
    await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("arf-arrival-host"),
      "host.display_name": "Arrival canonical",
      "host.hostname": "arrival.example.test",
      "host.aliases": aliasCollectionActionsPayload(["arrival alias"]),
    });
    const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("arf-arrival-source"),
      "timeline.activity_utc_text": "2026-01-01T00:00:00Z",
      "timeline.activity_synopsis_text": "Matching source",
    });
    await createTimelineFillers(page, incidentId, "arrival filler", 30, {
      occurredAtStart: "2026-02-01T00:00:00Z",
    });
    const editing = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("arf-arrival-editor"),
        "timeline.activity_utc_text": "2026-03-01T00:00:00Z",
        "timeline.activity_synopsis_text": "Continue editing",
      },
    );
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page, ["Hosts", "Tags"]);
    let release = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    let writes = 0;
    const path = `**/api/v1/records/${source.record_id}`;
    await page.route(path, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      writes++;
      await gate;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      await route.fulfill({ response });
    });
    try {
      await openTimelineInspector(page, source.record_id);
      const host = page.getByTestId(
        timelineCollectionInputTestId(
          source.record_id,
          "timeline.host_refs",
          "inspector",
        ),
      );
      await host.fill("arrival alias");
      await host.press("Enter");
      await expect.poll(() => writes).toBe(1);
      await host.fill("ARRIVAL ALIAS");
      await host.press("Enter");
      await page
        .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
        .click();
      const target =
        mode === "scalar"
          ? rowCellTestId(editing.record_id, "timeline.activity_synopsis_text")
          : relationshipItemsTestId(editing.record_id, "timeline.tags", "grid");
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: target,
      });
      if (mode === "scalar") await page.getByTestId(target).click();
      else
        await page
          .getByRole("group", { name: "Tags collection cell", exact: true })
          .filter({ has: page.getByTestId(target) })
          .getByRole("button", { name: "Add tags token", exact: true })
          .click();
      const input = page.getByTestId(
        mode === "scalar"
          ? timelineScalarEditorTestId({
              recordId: editing.record_id,
              fieldKey: "timeline.activity_synopsis_text",
              surface: "grid",
            })
          : timelineCollectionInputTestId(
              editing.record_id,
              "timeline.tags",
              "grid",
            ),
      );
      await input.fill("native raw Ω 東京 authoring");
      await input.evaluate((element: HTMLInputElement) =>
        element.setSelectionRange(3, 8, "backward"),
      );
      const cdp = await page.context().newCDPSession(page);
      if (mode === "composition")
        await cdp.send("Input.imeSetComposition", {
          text: "仮",
          selectionStart: 1,
          selectionEnd: 1,
        });
      const original = await input.elementHandle();
      const native = (element: HTMLInputElement | HTMLTextAreaElement) => ({
        text: element.value,
        start: element.selectionStart,
        end: element.selectionEnd,
        direction: element.selectionDirection,
        focused: document.activeElement === element,
        connected: element.isConnected,
      });
      const before = await input.evaluate(native);
      const scrollBefore = await page
        .locator(gridScrollportSelector())
        .evaluate((element) => ({
          top: element.scrollTop,
          left: element.scrollLeft,
        }));
      release();
      const region = page.getByRole("complementary", {
        name: "Auto-resolution disclosures",
      });
      await expect(region).toContainText("Arrival canonical");
      await expect(
        region.getByRole("button", { name: "Undo", exact: true }),
      ).toHaveCount(2);
      await expect(input).toBeFocused();
      expect(await original?.evaluate(native)).toEqual(before);
      const geometry = await input.evaluate((element) => {
        const grid = element.closest('[role="grid"], [role="treegrid"]');
        return {
          editor: element.getBoundingClientRect().toJSON(),
          grid: grid?.getBoundingClientRect().toJSON(),
        };
      });
      expect(geometry.editor.bottom).toBeLessThanOrEqual(
        (geometry.grid?.bottom ?? 0) + 1,
      );
      const scrollAfter = await page
        .locator(gridScrollportSelector())
        .evaluate((element) => ({
          top: element.scrollTop,
          left: element.scrollLeft,
        }));
      expect(scrollAfter.left).toBe(scrollBefore.left);
      expect(scrollAfter.top).toBeGreaterThanOrEqual(scrollBefore.top);
      await testInfo.attach(`${mode}-arrival`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });
      await testInfo.attach(`${mode}-continuity`, {
        body: JSON.stringify({
          before,
          after: await original?.evaluate(native),
          scrollBefore,
          scrollAfter,
          geometry,
          writes,
        }),
        contentType: "application/json",
      });
      if (mode === "composition") {
        await cdp.send("Input.imeSetComposition", {
          text: "仮名",
          selectionStart: 2,
          selectionEnd: 2,
        });
        await expect(input).toHaveValue(/仮名/);
        await cdp.send("Input.imeSetComposition", {
          text: "",
          selectionStart: 0,
          selectionEnd: 0,
        });
      }
      await cdp.detach();
      expect(writes).toBe(2);
    } finally {
      release();
      await page.unroute(path);
    }
  }
});

test("Timeline disclosures follow read-only session recovery account replacement and access loss", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}, testInfo) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ARF-AUTH"),
    "Disclosure authority",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("arf-auth-host"),
    "host.display_name": "Authority canonical host",
    "host.hostname": "authority.example.test",
    "host.aliases": aliasCollectionActionsPayload(["authority alias"]),
  });
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("arf-auth-row"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Disclosure authority source",
  });
  const members = [];
  for (const name of ["original", "replacement"])
    members.push(
      await createIncidentMemberUser(page, incident, {
        email: uniqueEmail("arf-auth-" + name),
        display_name: name,
        initial_password: "DisclosureLife1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    );
  const original = members[0],
    replacement = members[1];
  if (!original || !replacement)
    throw new Error("Missing disclosure authority members");
  const login = (member: typeof original, recovery = false) =>
    sessionTracker.loginTrackedUser(page, {
      recovery,
      createdBy: "timeline-auto-resolution-feedback",
      email: member.email,
      password: member.initial_password,
      purpose: "Disclosure authority lifetime",
      userId: member.user_id,
    });
  await login(original);
  await page.goto("/?incident_id=" + incident);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const notices = page
    .locator(autoResolutionNoticeFamilySelector())
    .filter({ has: page.getByRole("button", { name: "Review", exact: true }) });
  const add = async () => {
    await showTimelineCollectionColumns(page, ["Hosts"]);
    await addRelationshipTokenViaUI(
      page,
      row.record_id,
      "hostRefs",
      "authority alias",
    );
    await expect(notices).toHaveCount(1);
    await expect(notices.first()).toContainText("Authority canonical host");
  };
  await add();
  const noticeIdentity = await notices.first().getAttribute("data-testid");
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Disclosure same-account recovery",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await expect(notices).toHaveCount(0);
  await login(original, true);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(notices).toHaveCount(1);
  await expect(notices.first()).toHaveAttribute(
    "data-testid",
    noticeIdentity ?? "",
  );
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Disclosure account replacement",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await login(replacement, true);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(notices).toHaveCount(0);
  await add();
  expect(await notices.first().getAttribute("data-testid")).not.toBe(
    noticeIdentity,
  );
  const membership =
    "/api/v1/incidents/" + incident + "/memberships/" + replacement.user_id;
  expect(
    (
      await workerAdminRequest.patch(membership, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  // A real forbidden correction rechecks authority; it cannot retire disclosure.
  await notices
    .first()
    .getByRole("button", { name: "Undo", exact: true })
    .click();
  await expect(notices.first()).toContainText("Read only");
  await expect(
    notices.first().getByRole("button", { name: "Undo", exact: true }),
  ).toHaveCount(0);
  await notices
    .first()
    .getByRole("button", { name: "Review", exact: true })
    .click();
  await expect(notices).toHaveCount(1);
  await expect(
    page.getByTestId(workbookShellSlotTestId("inspector")),
  ).toBeVisible();
  await testInfo.attach("read-only-disclosure", {
    body: await page.screenshot({ fullPage: true }),
    contentType: "image/png",
  });
  expect(
    (
      await workerAdminRequest.delete(membership, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(notices).toHaveCount(0);
  await expect(
    page.getByText("Authority canonical host", { exact: true }),
  ).toHaveCount(0);
});
