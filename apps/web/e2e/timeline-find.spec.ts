import {
  applyFilterChip,
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  removeFilterChip,
  scrollGridTargetIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  draftCellTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  relationshipItemsTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { revokeAllSessions } from "./support/auth/sessions";
import { installPatchController } from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

const synopsis = "timeline.activity_synopsis_text";
const raw = "timeline.date_entered_text";
const timestamp = "timeline.activity_utc_text";
const grid = (page: Page) => page.locator(gridScrollportSelector());
const cell = (page: Page, id: string, field = synopsis) =>
  page
    .getByTestId(rowCellTestId(id, field))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
const editor = (page: Page, id: string, field = synopsis) =>
  page.getByTestId(
    timelineScalarEditorTestId({
      recordId: id,
      fieldKey: field,
      surface: "grid",
    }),
  );
const entry = (page: Page) =>
  page.getByRole("button", { name: "Find in loaded rows", exact: true });
const input = (page: Page) =>
  page.getByRole("textbox", { name: "Find in loaded rows", exact: true });
const status = (page: Page) =>
  page.getByRole("status").filter({
    hasText:
      /(?:matching cell|No matches in loaded rows|Enter text to find|Finding in loaded rows|Waiting for the edit)/,
  });
async function find(page: Page, term: string, count: number) {
  await entry(page).click();
  await input(page).fill(term);
  await expect(status(page)).toContainText(
    count
      ? `${count} matching ${count === 1 ? "cell" : "cells"}`
      : "No matches in loaded rows.",
  );
}
async function showField(page: Page, label: string) {
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  await page
    .getByTestId(workbookColumnsMenuTestId(timelineViewSchemaId))
    .getByRole("checkbox", { name: label, exact: true })
    .check();
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
}
async function seed(page: Page, count = 3) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TFN"),
    "Timeline local Find evidence",
  );
  const rows: { record_id: string; row_version: number }[] = [];
  let next = 0;
  await Promise.all(
    Array.from({ length: 8 }, async () => {
      for (;;) {
        const index = next++;
        if (index >= count) return;
        rows[index] = await createViewRow(
          page,
          incident,
          timelineViewSchemaId,
          {
            client_txn_id: uniqueTxn("find-seed"),
            [synopsis]: `${String(index).padStart(4, "0")} needle needle Café\nsecond line ${"long content ".repeat(index % 5 === 0 ? 800 : 0)}`,
            [raw]:
              index === count - 1
                ? "Far column target"
                : "Overflow metadata absent",
            "timeline.data_source_text": index === 0 ? "NEEDLE" : "source",
            [timestamp]: new Date(
              Date.UTC(2026, 0, 1, 0, 0, index),
            ).toISOString(),
          },
        );
      }
    }),
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  const frozenColumns = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const fields = requireViewContract(timelineViewSchemaId).fields;
  const boundary = fields.findIndex((field) => field.fieldKey === synopsis);
  for (const field of fields.slice(0, boundary + 1)) {
    if (field.fieldKey === "record_id" || field.fieldKey === "row_version")
      continue;
    await frozenColumns
      .getByRole("button", { name: `Width for ${field.label}`, exact: true })
      .click();
    await frozenColumns
      .getByRole("textbox", { name: "Width in CSS pixels" })
      .fill(field.fieldKey === synopsis ? "220" : "40");
    await frozenColumns
      .getByRole("button", { name: "Apply width", exact: true })
      .click();
    await frozenColumns
      .getByRole("button", { name: "Cancel", exact: true })
      .click();
  }
  await frozenColumns
    .getByRole("button", {
      name: `Freeze through ${requireViewContract(timelineViewSchemaId).fieldMap[synopsis]?.label}`,
      exact: true,
    })
    .click();
  await frozenColumns
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await expect(page.locator(gridScrollportSelector())).toHaveAttribute(
    "data-grid-freeze-state",
    "active",
  );
  await sortByHeader(page, timelineViewSchemaId, synopsis);
  await expect(cell(page, rows[0]?.record_id ?? "missing")).toBeVisible();
  return {
    incident,
    rows,
    first: rows[0]?.record_id ?? "missing",
    last: rows[count - 1]?.record_id ?? "missing",
  };
}
function observe(page: Page, incident: string) {
  const requests: string[] = [];
  page.on("request", (request) => {
    if (
      (request.url().includes(`/incidents/${incident}/views/`) &&
        request.url().endsWith("/query")) ||
      request.url().includes("/saved-views") ||
      request.method() === "PATCH" ||
      request.url().endsWith("/rows")
    )
      requests.push(`${request.method()} ${request.url()}`);
  });
  return requests;
}
async function navigate(
  page: Page,
  direction: "Enter" | "Shift+Enter" = "Enter",
) {
  await input(page).press(direction);
  await expect(input(page)).toHaveCount(0);
}
async function noNativeFindCapture(target: Locator) {
  return target.evaluate((element) =>
    element.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "f",
        ctrlKey: true,
        bubbles: true,
        cancelable: true,
      }),
    ),
  );
}

test("Timeline Find searches committed loaded cells and reveals both virtualized axes without query or layout writes", async ({
  page,
}) => {
  const f = await seed(page, 85);
  await showField(page, "Date Entered");
  await showField(page, "Tags");
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  const columns = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const later = columns.getByRole("button", {
    name: "Move Date Entered later",
    exact: true,
  });
  for (let i = 0; i < 40 && (await later.isEnabled()); i++) await later.click();
  await columns
    .getByRole("button", {
      name: "Width for Data Source",
      exact: true,
    })
    .click();
  await columns
    .getByRole("textbox", { name: "Width in CSS pixels" })
    .fill("4096");
  await columns
    .getByRole("button", { name: "Apply width", exact: true })
    .click();
  await columns.getByRole("button", { name: "Cancel", exact: true }).click();
  await columns
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  const requests = observe(page, f.incident);
  await find(page, "needle", 86);
  await expect(grid(page).locator("[data-grid-record-id]")).not.toHaveCount(85);
  await expect(cell(page, f.last)).toHaveCount(0);
  await input(page).fill("0084 needle");
  await expect(status(page)).toContainText("1 matching cell");
  await navigate(page);
  await expect(cell(page, f.last)).toBeFocused();
  await expect(editor(page, f.last)).toHaveCount(0);
  await cell(page, f.last).press("Control+f");
  await expect(input(page)).toBeFocused();
  await expect(cell(page, f.last, raw)).toHaveCount(0);
  await input(page).fill("Far column target");
  await expect(status(page)).toContainText("1 matching cell");
  await navigate(page);
  await expect(cell(page, f.last, raw)).toBeFocused();
  await expect(cell(page, f.last, raw)).toContainText("Far column target");
  const frozenRight = await grid(page)
    .locator('.cartulary-grid-frozen-data[role="columnheader"]')
    .last()
    .evaluate((node) => node.getBoundingClientRect().right);
  const revealedBounds = await cell(page, f.last, raw).boundingBox();
  if (!revealedBounds) throw new Error("Missing revealed Find cell");
  expect(revealedBounds.x).toBeGreaterThanOrEqual(frozenRight);
  await expect(cell(page, f.last, raw)).toHaveAttribute(
    "aria-description",
    /Current Find match/,
  );
  await find(page, "needle", 86);
  await navigate(page, "Shift+Enter");
  // Date Entered is now after Synopsis in the full semantic order. Previous
  // therefore reaches this row's frozen Synopsis before wrapping to row one.
  await expect(cell(page, f.last)).toBeFocused();
  await entry(page).click();
  await navigate(page);
  await expect(cell(page, f.first)).toBeFocused();
  await expect(status(page)).toContainText("Wrapped to beginning.");
  await entry(page).click();
  await navigate(page);
  await expect(cell(page, f.first, "timeline.data_source_text")).toBeFocused();
  expect(requests).toEqual([]);
});

test("Timeline Find literal case Unicode multiline and empty feedback preserve editor shortcut ownership", async ({
  page,
}) => {
  const f = await seed(page);
  const bulk = page.getByRole("checkbox", {
    name: `Select record ${f.first}`,
    exact: true,
  });
  await bulk.check();
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowCellTestId(f.first, synopsis),
  });
  await cell(page, f.first).click();
  await editor(page, f.first).press("Escape");
  await cell(page, f.first).press("Shift+ArrowDown");
  const selected = grid(page).locator(
    '[role="gridcell"][aria-selected="true"]',
  );
  await expect(selected).toHaveCount(2);
  const requests = observe(page, f.incident);
  await find(page, "needle", 4);
  await page.getByRole("checkbox", { name: "Match case" }).check();
  await expect(status(page)).toContainText("3 matching cells");
  await input(page).fill("Cafe\u0301");
  await expect(status(page)).toContainText("3 matching cells");
  await input(page).fill("Café\nsecond line");
  await expect(status(page)).toContainText("3 matching cells");
  await input(page).fill("second line");
  await expect(status(page)).toContainText("3 matching cells");
  await input(page).fill("needle.*");
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await expect(
    page.getByRole("region", { name: "Find in loaded rows" }),
  ).toContainText("Other incident rows were not searched.");
  await input(page).fill("");
  await expect(status(page)).toHaveText("Enter text to find in loaded rows.");
  await expect(
    page.getByRole("button", { name: "Next", exact: true }),
  ).toBeDisabled();
  await expect(selected).toHaveCount(2);
  await expect(bulk).toBeChecked();
  await input(page).press("Escape");
  await expect(selected).toHaveCount(2);
  await find(page, "0001 needle", 1);
  await navigate(page);
  await expect(selected).toHaveCount(1);
  await expect(bulk).toBeChecked();
  await entry(page).click();
  await input(page).press("Escape");
  await cell(page, f.first).click();
  await expect(editor(page, f.first)).toBeFocused();
  expect(await noNativeFindCapture(editor(page, f.first))).toBe(true);
  await editor(page, f.first).press("Escape");
  await expect(cell(page, f.first)).toBeFocused();
  await cell(page, f.first).press("Control+f");
  await expect(input(page)).toBeFocused();
  await expect(input(page)).toHaveValue("");
  expect(requests).toEqual([]);
});

test("Timeline Find opening preserves exact scalar drafts and explicit navigation gates acceptance rejection and supersession", async ({
  page,
}) => {
  const f = await seed(page);
  await cell(page, f.first).click();
  await editor(page, f.first).fill("Exact unfinished draft");
  const hold = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${f.first}`,
  });
  try {
    await find(page, "0002 needle", 1);
    expect(hold.hitCount()).toBe(0);
    await expect(editor(page, f.first)).toHaveValue("Exact unfinished draft");
    await input(page).press("Enter");
    await hold.waitForHit;
    await expect(status(page)).toContainText("Waiting for the edit to save");
    await expect(editor(page, f.first)).toHaveValue("Exact unfinished draft");
    await page.getByRole("button", { name: "Close Find" }).click();
    await expect(editor(page, f.first)).toBeFocused();
    hold.release();
    await expect.poll(() => hold.hitCount()).toBe(1);
    await find(page, "0002 needle", 1);
    await navigate(page);
    await expect(cell(page, f.last)).toBeFocused();
  } finally {
    await hold.dispose();
  }
  await find(page, "Exact unfinished draft", 1);
  await navigate(page);
  await cell(page, f.first).click();
  await editor(page, f.first).fill("rejected exact draft");
  const rejected = await installPatchController(page);
  try {
    rejected.failNextPatch(422, "invalid_request", { recordId: f.first });
    await find(page, "0002 needle", 1);
    await input(page).press("Enter");
    await expect(editor(page, f.first)).toHaveValue("rejected exact draft");
    await expect(editor(page, f.first)).toBeFocused();
    await expect(status(page)).toContainText("Correct the original edit");
  } finally {
    await rejected.dispose();
  }
});

test("Timeline Find follows column group saved-view and live membership while retaining only eligible current matches", async ({
  page,
}) => {
  const f = await seed(page);
  await find(page, "needle", 4);
  await navigate(page);
  await expect(cell(page, f.first)).toBeFocused();
  const columns = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const toggle = page.getByTestId(
    workbookColumnsMenuTriggerTestId(timelineViewSchemaId),
  );
  await toggle.click();
  await columns
    .getByRole("checkbox", { name: "Data Source", exact: true })
    .uncheck();
  await columns.getByRole("button", { name: "Close columns" }).click();
  await entry(page).click();
  await expect(input(page)).toHaveValue("needle");
  await expect(status(page)).toContainText("1 of 3 matching cells");
  await toggle.click();
  await columns
    .getByRole("checkbox", { name: "Data Source", exact: true })
    .check();
  await columns
    .getByRole("button", { name: "Move Data Source earlier", exact: true })
    .click();
  await columns.getByRole("button", { name: "Close columns" }).click();
  await entry(page).click();
  await expect(status(page)).toContainText("2 of 4 matching cells");
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const groupId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "rough",
  );
  await collapseGridGroup({
    page,
    surface: timelineViewSchemaId,
    groupTestId: groupId,
  });
  await entry(page).click();
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await expandGridGroup({
    page,
    surface: timelineViewSchemaId,
    groupTestId: groupId,
  });
  await entry(page).click();
  await expect(status(page)).toContainText("4 matching cells");
  const saved = await createSavedView(page, f.incident, {
    display_name: "Find hidden summary",
    view_schema_id: timelineViewSchemaId,
    layout_json: {
      layout_schema_id: "cartulary.layout.v2",
      frozen_through_field_key: null,
      column_order: requireViewContract(timelineViewSchemaId).fields.map(
        (field) => field.fieldKey,
      ),
      column_widths: [],
      hidden_field_keys: [synopsis],
    },
  });
  // Refresh discovery through ordinary surface activation; Find retires on departure.
  await switchOrdinarySheet(page, hostsViewSchemaId);
  await expect(entry(page)).toBeVisible();
  await entry(page).click();
  await expect(input(page)).toHaveValue("");
  await switchOrdinarySheet(page, timelineViewSchemaId);
  await find(page, "needle", 4);
  await selectSavedView(page, timelineViewSchemaId, saved.saved_view_id);
  await entry(page).click();
  await expect(input(page)).toHaveValue("needle");
  await expect(status(page)).toContainText("1 matching cell");
  await selectSavedView(page, timelineViewSchemaId, "");
  await showField(page, "Activity Synopsis");
  await find(page, "0002 needle", 1);
  await navigate(page);
  await patchRecord(page, f.last, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: f.rows[2]?.row_version ?? 1,
    client_txn_id: uniqueTxn("find-live"),
    changes: [{ field_key: synopsis, value: "Changed committed value" }],
  });
  await expect(cell(page, f.last)).toContainText("Changed committed value");
  await expect(cell(page, f.last)).toBeFocused();
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await find(page, "Changed committed value", 1);
  await navigate(page);
  const current = (
    await queryViewRows(page, f.incident, timelineViewSchemaId)
  ).find((row) => row.record_id === f.last);
  if (!current) throw new Error("Missing current fixture row");
  const removed = await publicHttpOperation({
    operationID: "deleteRecord",
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    pathParameters: { record_id: f.last },
    body: {
      base_row_version: current.row_version,
      client_txn_id: uniqueTxn("find-delete"),
      reason: "Find fixture deletion",
    },
  });
  expect(removed.ok).toBe(true);
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await expect(cell(page, f.last)).toHaveCount(0);
});

test("Timeline Find remains responsive across the full retained window append eviction and authorized query failure", async ({
  page,
}, info) => {
  const f = await seed(page, 405);
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  const requests = observe(page, f.incident);
  await find(page, "needle", 101);
  expect(requests).toEqual([]);
  const more = browsing.getByRole("button", { name: "Load more", exact: true });
  for (const expected of [201, 301]) {
    await more.click();
    await entry(page).click();
    await expect(status(page)).toContainText(`${expected} matching cells`);
  }
  const baseline = [...requests];
  await input(page).fill("obsolete term");
  await input(page).fill("long content");
  await input(page).fill("needle");
  await expect(status(page)).toContainText("301 matching cells");
  await input(page).fill("cancel immediately");
  await input(page).press("Escape");
  await expect(input(page)).toHaveCount(0);
  await find(page, "needle", 301);
  expect(requests).toEqual(baseline);
  expect(
    await grid(page).locator("[data-grid-record-id]").count(),
  ).toBeLessThan(60);
  await info.attach("find-full-window", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await more.click();
  await entry(page).click();
  await expect(status(page)).toContainText("300 matching cells");
  await input(page).fill("0000 needle");
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await input(page).fill("needle");
  await expect(status(page)).toContainText("300 matching cells");
  const route = `**/api/v1/incidents/${f.incident}/views/${timelineViewSchemaId}/query`;
  await page.route(route, (request) => request.abort("failed"), { times: 1 });
  await browsing.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(browsing).toContainText("Retry");
  await entry(page).click();
  await expect(status(page)).toContainText("300 matching cells");
  await expect(
    page.getByRole("region", { name: "Find in loaded rows" }),
  ).toContainText("Showing retained rows");
});

test("Timeline Find collection summaries read-only values and supported geometry retain accessible focus and native controls", async ({
  page,
}, info) => {
  const f = await seed(page);
  await patchRecord(page, f.first, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: f.rows[0]?.row_version ?? 1,
    client_txn_id: uniqueTxn("find-tags"),
    changes: [
      {
        field_key: "timeline.tags",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_tag", tag_name: "Alpha visible summary Ω" },
            { op: "add_tag", tag_name: "Zulu overflow exclusive" },
          ],
        },
      },
    ],
  });
  await showField(page, "Tags");
  await find(page, "Alpha visible summary Ω", 1);
  await input(page).fill("Zulu overflow exclusive");
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await input(page).fill("No items");
  await expect(status(page)).toHaveText("No matches in loaded rows.");
  await showField(page, "Recorded");
  const committed = (
    await queryViewRows(page, f.incident, timelineViewSchemaId)
  )[0]?.cells["timeline.recorded_at"]?.value;
  if (typeof committed !== "string")
    throw new Error("Missing read-only timestamp");
  await find(page, committed, 1);
  await navigate(page);
  await expect(grid(page).locator('[role="gridcell"]:focus')).toHaveAttribute(
    "aria-readonly",
    "true",
  );
  for (const [width, zoom] of [
    [1440, 1],
    [1280, 1],
    [1024, 1],
    [768, 1],
    [1440, 2],
  ] as const) {
    await page.setViewportSize({ width: width * zoom, height: 900 * zoom });
    await page.evaluate((value) => {
      document.documentElement.style.zoom = String(value);
    }, zoom);
    const spacing = await page.addStyleTag({
      content:
        "* { letter-spacing: 0.12em !important; word-spacing: 0.16em !important; line-height: 1.5 !important; }",
    });
    await find(page, "0000 needle", 1);
    await expect(input(page)).toBeFocused();
    const panel = page.getByRole("region", { name: "Find in loaded rows" });
    const rect = await panel.boundingBox();
    expect(rect).not.toBeNull();
    if (!rect) throw new Error("Missing Find geometry");
    expect(rect.x).toBeGreaterThanOrEqual(0);
    expect(rect.x + rect.width).toBeLessThanOrEqual(width * zoom);
    await expect(
      page.getByRole("button", { name: "Previous", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Close Find" }),
    ).toBeVisible();
    await info.attach(`find-${width}-${zoom}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    await navigate(page);
    await expect(cell(page, f.first)).toBeFocused();
    await expect(
      cell(page, f.first).getByRole("img", {
        name: "Current Find match",
        exact: true,
      }),
    ).toBeVisible();
    await expect(panel).toHaveCount(0);
    await spacing.evaluate((element) =>
      element.parentNode?.removeChild(element),
    );
  }
});

test("Timeline Find borrows collection and inspector editors without writes and settles each departure once", async ({
  page,
}) => {
  const f = await seed(page, 85);
  await showField(page, "Tags");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(f.first, "timeline.tags", "grid"),
  });
  const tags = page
    .getByTestId(relationshipItemsTestId(f.first, "timeline.tags", "grid"))
    .locator("xpath=ancestor::fieldset[1]");
  await tags.getByRole("button", { name: "Add tags token" }).click();
  const tagInput = page.getByTestId(
    timelineCollectionInputTestId(f.first, "timeline.tags", "grid"),
  );
  await tagInput.fill("Find borrowed tag");
  const gate = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${f.first}`,
  });
  try {
    await find(page, "Find borrowed tag", 0);
    expect(gate.hitCount()).toBe(0);
    await expect(tagInput).toHaveValue("Find borrowed tag");
    await input(page).fill("0084 needle");
    await expect(status(page)).toContainText("1 matching cell");
    await input(page).press("Enter");
    await gate.waitForHit;
    expect(gate.hitCount()).toBe(1);
    gate.release();
    await expect(input(page)).toHaveCount(0);
    await expect(cell(page, f.last)).toBeFocused();
    await expect(cell(page, f.last)).toBeInViewport({ ratio: 1 });
  } finally {
    await gate.dispose();
  }
  await openTimelineInspector(page, f.first);
  await page.locator(`[data-inspector-edit-field="${synopsis}"]`).click();
  const inspectorInput = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: f.first,
      fieldKey: synopsis,
      surface: "inspector",
    }),
  );
  await inspectorInput.fill("Exact inspector draft");
  const inspectorGate = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${f.first}`,
  });
  try {
    await find(page, "0084 needle", 1);
    expect(inspectorGate.hitCount()).toBe(0);
    await expect(inspectorInput).toHaveValue("Exact inspector draft");
    await input(page).press("Enter");
    await expect(input(page)).toHaveCount(0);
    await expect(cell(page, f.last)).toBeFocused();
    await expect(cell(page, f.last)).toBeInViewport({ ratio: 1 });
    await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
    await expect(inspectorInput).toHaveValue("Exact inspector draft");
    expect(inspectorGate.hitCount()).toBe(0);
  } finally {
    await inspectorGate.dispose();
  }
});

test("Timeline Find retires protected terms and matches on session suspension and does not return them after recovery", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const f = await seed(page);
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("find-authority"),
    display_name: "Find authority fixture",
    initial_password: "TimelineFind1!",
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const login = (recovery = false) =>
    sessionTracker.loginTrackedUser(page, {
      recovery,
      createdBy: "timeline-find",
      email: member.email,
      password: member.initial_password,
      purpose: "Find authority lifetime",
      userId: member.user_id,
    });
  await login();
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}`);
  await sockets.waitForAcceptedSocket();
  await find(page, "Protected absent term", 0);
  await input(page).fill("needle");
  await expect(status(page)).toContainText("4 matching cells");
  await navigate(page);
  await expect(grid(page).locator('[role="gridcell"]:focus')).toHaveAttribute(
    "aria-readonly",
    "true",
  );
  await entry(page).click();
  await revokeAllSessions(
    workerAdminRequest,
    member.user_id,
    "Find suspension evidence",
  );
  const revoked = await sockets.waitForMessage("session_revoked");
  await sockets.waitForClose(revoked.socketIndex);
  await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
    "data-bootstrap-state",
    "revoked",
  );
  await expect(input(page)).toHaveCount(0);
  await expect(
    page.getByRole("img", { name: "Current Find match", exact: true }),
  ).toHaveCount(0);
  const messageStart = sockets.messageCount();
  await login(true);
  await sockets.waitForAcceptedSocket({ startAt: messageStart });
  await expect(entry(page)).toBeVisible();
  await entry(page).click();
  await expect(input(page)).toHaveValue("");
  await expect(status(page)).toHaveText("Enter text to find in loaded rows.");
});

test("Timeline Find preserves blank authoring and excludes creation pins through explicit query replacement", async ({
  page,
}) => {
  const f = await seed(page);
  const requests = observe(page, f.incident);
  const draft = page.getByTestId(draftCellTestId(synopsis));
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftCellTestId(synopsis),
  });
  await draft.focus();
  await find(page, "needle", 4);
  await navigate(page);
  expect(requests).toEqual([]);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await find(page, "needle", 0);
  await input(page).press("Escape");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftCellTestId(synopsis),
  });
  const createPath = `/api/v1/incidents/${f.incident}/views/${timelineViewSchemaId}/rows`;
  const accepted = page.waitForResponse(
    (response) =>
      response.url().endsWith(createPath) &&
      response.request().method() === "POST",
  );
  await draft.fill("Pinned creation needle");
  const creation = await accepted;
  expect(creation.ok()).toBe(true);
  const pinnedId = (await creation.json()).data.row.record_id as string;
  await find(page, "Pinned creation needle", 0);
  // A visible creation pin is not an accepted query member or bulk target.
  await expect(
    page.getByRole("checkbox", {
      name: `Select record ${pinnedId}`,
      exact: true,
    }),
  ).toHaveCount(0);
  expect(
    requests.filter((request) => request.endsWith(createPath)),
  ).toHaveLength(1);
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  await entry(page).click();
  await expect(input(page)).toHaveValue("Pinned creation needle");
  await expect(status(page)).toContainText("1 matching cell");
  await navigate(page);
  await expect(
    page.getByRole("checkbox", {
      name: `Select record ${pinnedId}`,
      exact: true,
    }),
  ).not.toBeChecked();
  expect(
    requests.filter((request) => request.endsWith(createPath)),
  ).toHaveLength(1);
});
