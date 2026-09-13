import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridFillHandleSelector,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookEditRecoveryDiscardButtonTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import {
  createViewRow,
  queryViewRows,
  waitForViewRowByCell,
} from "./support/workbook/query";

function required<T>(value: T | null | undefined): T {
  if (value === null || value === undefined)
    throw new Error("Expected fixture value is missing.");
  return value;
}

const synopsis = "timeline.activity_synopsis_text";
const source = "timeline.data_source_text";
const editor = (page: Page, recordId: string, fieldKey = synopsis) =>
  page.getByTestId(
    timelineScalarEditorTestId({ recordId, fieldKey, surface: "grid" }),
  );
const cell = (page: Page, recordId: string, fieldKey = synopsis) =>
  page
    .getByTestId(rowCellTestId(recordId, fieldKey))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");

async function openTimeline(page: Page, incidentId: string) {
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
}

async function seedTimeline(page: Page, count: number) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-ENTRY"),
    "Timeline grid interaction regression",
  );
  for (let index = 0; index < count; index += 1)
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("grid-seed"),
      [synopsis]: `Fact ${index}`,
      [source]: `Source ${index}`,
    });
  await openTimeline(page, incidentId);
  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  return { incidentId, rows };
}

async function selectCell(page: Page, recordId: string, fieldKey = synopsis) {
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId,
    cellKey: fieldKey,
  });
  await page.getByTestId(rowCellTestId(recordId, fieldKey)).click();
  if (fieldKey !== "timeline.capture_state")
    await expect(editor(page, recordId, fieldKey)).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(cell(page, recordId, fieldKey)).toBeFocused();
}

async function clipboard(page: Page, text: string) {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.evaluate((value) => navigator.clipboard.writeText(value), text);
  await page.keyboard.press("Control+v");
}

async function tabTo(page: Page, target: Locator) {
  for (let step = 0; step < 128; step += 1) {
    if (
      await target.evaluateAll((elements) =>
        elements.some((element) => element === document.activeElement),
      )
    )
      return;
    await page.keyboard.press("Tab");
    await page.evaluate(
      () =>
        new Promise<void>((resolve) => requestAnimationFrame(() => resolve())),
    );
  }
  await expect(target).toBeFocused();
}

async function externalPatch(
  page: Page,
  incidentId: string,
  recordId: string,
  fieldKey: string,
  value: string,
) {
  const current = required(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === recordId,
    ),
  );
  const response = await page.request.patch(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        view_schema_id: timelineViewSchemaId,
        client_txn_id: uniqueTxn("concurrent-edit"),
        base_row_version: current.row_version,
        changes: [{ field_key: fieldKey, value }],
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

test("Timeline pointer transitions preserve one edit and wait for acceptance", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  const requests: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${id}`)
    )
      requests.push(request.postData() ?? "");
  });
  await page.getByTestId(rowCellTestId(id, synopsis)).click();
  await expect(editor(page, id)).toBeFocused();
  expect(
    await editor(page, id).evaluate(
      (element: HTMLInputElement | HTMLTextAreaElement) => ({
        start: element.selectionStart,
        end: element.selectionEnd,
        length: element.value.length,
      }),
    ),
  ).toEqual({
    start: String(required(rows[0]).cells[synopsis]?.value).length,
    end: String(required(rows[0]).cells[synopsis]?.value).length,
    length: String(required(rows[0]).cells[synopsis]?.value).length,
  });
  await page.keyboard.press("Escape");
  await page.getByTestId(rowCellTestId(id, synopsis)).dblclick();
  await expect(editor(page, id)).toBeFocused();
  // A second click may reposition the caret, but never starts another session.
  await page.keyboard.press("End");
  await page.keyboard.type(" pointer");
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.getByTestId(rowCellTestId(id, source)).click();
    await held.waitForHit;
    await expect(editor(page, id)).toHaveValue(
      `${required(rows[0]).cells[synopsis]?.value} pointer`,
    );
    await expect(editor(page, id, source)).toHaveCount(0);
    held.release();
    await expect(editor(page, id, source)).toBeFocused();
    expect(requests).toHaveLength(1);
    await page.keyboard.press("Escape");
    await expect(cell(page, id, source)).toBeFocused();
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      `${required(rows[0]).cells[synopsis]?.value} pointer`,
    );
    await page
      .getByRole("checkbox", { name: `Select record ${id}`, exact: true })
      .click();
    await expect(editor(page, id)).toHaveCount(0);
    expect(requests).toHaveLength(1);
  } finally {
    await held.dispose();
  }
  await selectCell(page, id);
  await page.keyboard.type("Accepted after Escape");
  const late = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.keyboard.press("Tab");
    await late.waitForHit;
    await page.keyboard.press("Escape");
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.press("ArrowDown");
    await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
    late.release();
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Accepted after Escape",
    );
    await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
    expect(requests).toHaveLength(2);
    expect(await fetchRecordHistoryCount(page, id)).toBe(3);
  } finally {
    await late.dispose();
  }
});

test("Timeline rectangle paste keyboard fill and pointer fill preserve targets", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 3);
  const [first, second, third] = rows.map((row) => row.record_id) as [
    string,
    string,
    string,
  ];
  await selectCell(page, first);
  await clipboard(page, "Pasted one\tShared source\nPasted two\tSecond source");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Pasted two",
  );
  await selectCell(page, first, source);
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+d");
  await expect(page.getByTestId(rowCellTestId(second, source))).toHaveText(
    "Shared source",
  );
  await expect(cell(page, first, source)).toBeFocused();
  await selectCell(page, first, source);
  const handle = page.locator(gridFillHandleSelector());
  await expect(handle).toHaveAttribute("aria-label", "Drag to fill this value");
  await handle.scrollIntoViewIfNeeded();
  const start = required(await handle.boundingBox());
  const end = required(await cell(page, third, source).boundingBox());
  expect(
    await handle.evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return (
        document
          .elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2)
          ?.closest('[data-cartulary-fill-handle="true"]') === element
      );
    }),
  ).toBe(true);
  await page.mouse.move(start.x + start.width / 2, start.y + start.height / 2);
  await page.mouse.down();
  await page.mouse.move(end.x + end.width / 2, end.y + end.height / 2, {
    steps: 8,
  });
  await page.mouse.up();
  await expect(page.getByTestId(rowCellTestId(third, source))).toHaveText(
    "Shared source",
  );
  const persisted = await queryViewRows(page, incidentId, timelineViewSchemaId);
  for (const id of [first, second, third])
    expect(
      persisted.find((row) => row.record_id === id)?.cells[source]?.value,
    ).toBe("Shared source");
  expect(
    persisted.find((row) => row.record_id === first)?.cells[synopsis]?.value,
  ).toBe("Pasted one");
  expect(
    persisted.find((row) => row.record_id === second)?.cells[synopsis]?.value,
  ).toBe("Pasted two");
  await selectCell(page, first, source);
  let fills = 0;
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations")) fills += 1;
  });
  await page.keyboard.press("Shift+ArrowLeft");
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Control+d");
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: "Select a writable one-column range" }),
  ).toHaveText("Select a writable one-column range before using fill down.");
  await page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  );
  expect(fills).toBe(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline rejected edits keep correction local and preserve rough date text", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 2);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Unsaved");
  await editor(page, id).fill("Invalid\u0001text");
  await page.keyboard.press("Tab");
  await expect(editor(page, id)).toBeFocused();
  await expect(editor(page, id)).toHaveValue("Invalid\u0001text");
  expect(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === id,
    )?.cells[synopsis]?.value,
  ).toBe(required(rows[0]).cells[synopsis]?.value);
  await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
  await selectCell(page, id);
  await page.keyboard.type("Corrected fact");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Corrected fact",
  );
  await selectCell(page, id, "timeline.activity_utc_text");
  await page.keyboard.type("not-a-timestamp");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    "timeline.activity_utc_text",
    "not-a-timestamp",
  );
  await selectCell(page, id);
  await page.keyboard.type("My conflict draft");
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${id}`,
  });
  try {
    await page.keyboard.press("Tab");
    await held.waitForHit;
    await externalPatch(
      page,
      incidentId,
      id,
      synopsis,
      "Concurrent saved fact",
    );
    held.release();
    await expect(editor(page, id)).toHaveValue("My conflict draft");
    await expect(editor(page, id)).toBeFocused();
    await expect(
      page.getByTestId(conflictMarkerTestId(id, synopsis)),
    ).toBeVisible();
    expect(
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
        (row) => row.record_id === id,
      )?.cells[synopsis]?.value,
    ).toBe("Concurrent saved fact");
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await page
      .getByRole("button", { name: "Close conflict recovery", exact: true })
      .click();
    await page.getByTestId(conflictMarkerTestId(id, synopsis)).click();
    await page
      .getByRole("button", { name: "Discard local draft", exact: true })
      .click();
    await expect(page.getByTestId(rowCellTestId(id, synopsis))).toHaveText(
      "Concurrent saved fact",
    );
    await selectCell(page, id);
    await page.keyboard.type("Continued after conflict");
    await page.keyboard.press("Enter");
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "Continued after conflict",
    );
  } finally {
    await held.dispose();
  }
});

test("Timeline active drafts survive in-app refresh and stale refresh failure", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 3);
  const id = required(rows.at(-1)).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Draft kept through refresh");
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  const before = await scrollport.evaluate((element) => ({
    top: element.scrollTop,
    left: element.scrollLeft,
  }));
  const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  let releaseRefresh!: () => void;
  let queryCaptured!: () => void;
  const refreshGate = new Promise<void>((resolve) => {
    releaseRefresh = resolve;
  });
  const captured = new Promise<void>((resolve) => {
    queryCaptured = resolve;
  });
  await page.route(
    `**${queryPath}`,
    async (route) => {
      const response = await route.fetch();
      queryCaptured();
      await refreshGate;
      await route.fulfill({ response });
    },
    { times: 1 },
  );
  try {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("live-create"),
      [synopsis]: "Concurrent new row",
    });
    await captured;
    await externalPatch(
      page,
      incidentId,
      id,
      source,
      "Newer source during refresh",
    );
    await expect(page.getByTestId(rowCellTestId(id, source))).toHaveText(
      "Newer source during refresh",
    );
    await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
    await expect(editor(page, id)).toBeFocused();
    releaseRefresh();
    await expect(
      page.locator('[data-grid-data-state="refreshing"]'),
    ).toHaveCount(0);
    await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
    await expect(editor(page, id)).toBeFocused();
    expect(
      await scrollport.evaluate((element) => ({
        top: element.scrollTop,
        left: element.scrollLeft,
      })),
    ).toEqual(before);
    await expect(page.getByTestId(rowCellTestId(id, source))).toHaveText(
      "Newer source during refresh",
    );
  } finally {
    releaseRefresh();
  }
  await page.keyboard.press("Home");
  await page.keyboard.press("ArrowRight");
  // An authoritative sort-key update moves the active record from the end
  // to the beginning of the query while its synopsis remains uncommitted.
  await externalPatch(
    page,
    incidentId,
    id,
    "timeline.activity_utc_text",
    "2026-04-10T12:00:00Z",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("reorder-refresh"),
    [synopsis]: "Concurrent sort refresh",
  });
  await expect(
    page.locator('[role="row"][data-grid-record-id]').first(),
  ).toHaveAttribute("data-grid-record-id", id);
  await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
  await expect(editor(page, id)).toBeFocused();
  expect(
    await editor(page, id).evaluate(
      (element: HTMLInputElement | HTMLTextAreaElement) =>
        element.selectionStart,
    ),
  ).toBe(1);
  const reordered = await queryViewRows(page, incidentId, timelineViewSchemaId);
  expect(required(reordered[0]).record_id).toBe(id);
  expect(required(reordered[0]).cells[synopsis]?.value).toBe(
    required(rows.at(-1)).cells[synopsis]?.value,
  );
  await page.route(
    `**${queryPath}`,
    async (route) =>
      route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "internal_error",
            message: "Temporary refresh failure",
          },
        }),
      }),
    { times: 1 },
  );
  const failedRefresh = page.waitForResponse(
    (response) =>
      response.url().endsWith(queryPath) && response.status() === 503,
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("live-create-failure"),
    [synopsis]: "Another concurrent row",
  });
  await failedRefresh;
  await expect(editor(page, id)).toHaveValue("Draft kept through refresh");
  await expect(editor(page, id)).toBeFocused();
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Draft kept through refresh",
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline virtualization retains the draft and semantic cell without mounting every row", async ({
  page,
}) => {
  const { incidentId, rows } = await seedTimeline(page, 70);
  const id = required(rows[0]).record_id;
  await selectCell(page, id);
  await page.keyboard.type("Virtualized retained draft");
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  await scrollport.hover();
  await page.mouse.wheel(0, 100_000);
  // RDG pins one active row to preserve its editor, while ordinary rows unmount.
  await expect(
    page.locator(
      `[role="row"][data-grid-record-id="${required(rows[1]).record_id}"]`,
    ),
  ).toHaveCount(0);
  await expect
    .poll(() => page.locator('[role="row"][data-grid-record-id]').count())
    .toBeLessThan(40);
  expect(
    await editor(page, id).evaluate((element) => {
      const viewport = element
        .closest('[role="grid"]')
        ?.getBoundingClientRect();
      const rect = element.getBoundingClientRect();
      return viewport !== undefined && rect.bottom <= viewport.top;
    }),
  ).toBe(true);
  await page.mouse.wheel(0, -100_000);
  await expect
    .poll(() => scrollport.evaluate((element) => element.scrollTop))
    .toBe(0);
  await expect(editor(page, id)).toHaveValue("Virtualized retained draft");
  await expect(editor(page, id)).toBeFocused();
  await page.keyboard.type(" continued");
  await page.keyboard.press("Enter");
  await waitForViewRowByCell(
    page,
    incidentId,
    timelineViewSchemaId,
    synopsis,
    "Virtualized retained draft continued",
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline spreadsheet keys commit and navigate across multiple rows", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-KEYS"),
    "Timeline spreadsheet keyboard capture",
  );
  for (let index = 0; index < 3; index += 1) {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("grid-key-seed"),
      [synopsis]: `Seed ${index}`,
    });
  }
  await openTimeline(page, incidentId);
  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const first = required(rows[0]).record_id;
  await tabTo(page, cell(page, first));
  for (let index = 0; index < rows.length; index += 1) {
    const id = required(rows[index]).record_id;
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.type(`Keyboard row ${index}`);
    await expect(editor(page, id)).toHaveValue(`Keyboard row ${index}`);
    await page.keyboard.press("Tab");
    await expect(cell(page, id, source)).toBeFocused();
    await page.keyboard.type(`Source ${index}`);
    await page.keyboard.press("Shift+Tab");
    await expect(cell(page, id)).toBeFocused();
    await page.keyboard.press("Enter");
  }
  await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
  await page.keyboard.press("Shift+Enter");
  await expect(cell(page, required(rows[2]).record_id)).toBeFocused();
  await page.keyboard.press("ArrowUp");
  await expect(cell(page, required(rows[1]).record_id)).toBeFocused();
  const fields = requireViewContract(timelineViewSchemaId).defaultVisibleFields;
  const firstField = required(fields[0]);
  const lastField = required(fields.at(-1));
  await page.keyboard.press("End");
  await expect(
    cell(page, required(rows[1]).record_id, lastField),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    cell(page, required(rows[2]).record_id, firstField),
  ).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(
    cell(page, required(rows[1]).record_id, lastField),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await page.keyboard.press("Tab");
  const draftFields = fields.filter(
    (field) =>
      requireViewContract(timelineViewSchemaId).fieldMap[field]?.writeKind !==
      "read_only",
  );
  for (const field of draftFields) {
    await expect(page.getByTestId(draftCellTestId(field))).toBeFocused();
    await page.keyboard.press("Tab");
  }
  await expect
    .poll(() =>
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .evaluate((element) => element.contains(document.activeElement)),
    )
    .toBe(false);
  expect(
    await queryViewRows(page, incidentId, timelineViewSchemaId),
  ).toHaveLength(3);
  const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
  for (let index = 0; index < rows.length; index += 1) {
    const row = required(
      saved.find(
        (candidate) => candidate.record_id === required(rows[index]).record_id,
      ),
    );
    expect(row.cells[synopsis]?.value).toBe(`Keyboard row ${index}`);
    expect(row.cells[source]?.value).toBe(`Source ${index}`);
  }
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline first input creates once and retains typing through acknowledgement", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-CREATE"),
    "Timeline continuous rough capture",
  );
  await openTimeline(page, incidentId);
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  const submissions: unknown[] = [];
  page.on("request", (request) => {
    if (
      ["POST", "PATCH"].includes(request.method()) &&
      (request.url().endsWith(path) ||
        request.url().includes("/api/v1/records/"))
    )
      submissions.push({
        method: request.method(),
        payload: request.postDataJSON(),
      });
  });
  try {
    const draft = page.getByTestId(draftCellTestId(synopsis));
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftCellTestId(synopsis),
    });
    await tabTo(page, draft);
    const originalDraftId = await draft.getAttribute("id");
    await page.keyboard.type("First");
    await held.waitForHit;
    await page.keyboard.press("Enter");
    await page.keyboard.type(" continuous fact");
    await expect(draft).toHaveValue("First continuous fact");
    expect(held.hitCount()).toBe(1);
    await page.keyboard.press("Home");
    await page.keyboard.press("ArrowRight");
    expect(
      await draft.evaluate(
        (element: HTMLTextAreaElement) => element.selectionStart,
      ),
    ).toBe(1);
    held.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(rows).toHaveLength(1);
    const id = required(rows[0]).record_id;
    await expect(editor(page, id)).toBeFocused();
    await expect(editor(page, id)).toHaveValue("First continuous fact");
    expect(
      await editor(page, id).evaluate(
        (element: HTMLTextAreaElement) => element.selectionStart,
      ),
    ).toBe(1);
    await page.keyboard.press("End");
    await page.keyboard.type(" continued");
    await page.keyboard.press("Enter");
    await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
    await expect(editor(page, id)).toHaveCount(0);
    await expect(
      page.getByTestId(draftCellTestId(synopsis)),
    ).not.toHaveAttribute("id", required(originalDraftId));
    await waitForViewRowByCell(
      page,
      incidentId,
      timelineViewSchemaId,
      synopsis,
      "First continuous fact continued",
    );
    expect(
      await queryViewRows(page, incidentId, timelineViewSchemaId),
    ).toHaveLength(1);
    for (const label of ["Second keyboard capture", "Third keyboard capture"]) {
      const holdDuplicate = label.startsWith("Second");
      const createGate = holdDuplicate
        ? await holdBrowserRequest(page, { method: "POST", path })
        : null;
      const patchGate = holdDuplicate
        ? await holdBrowserRequest(page, {
            method: "PATCH",
            path: "/api/v1/records/*",
          })
        : null;
      try {
        await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
        await page.keyboard.type(label);
        if (createGate !== null && patchGate !== null) {
          await createGate.waitForHit;
          createGate.release();
          await patchGate.waitForHit;
          const capturing = required(
            (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
              (row) => row.record_id !== id,
            ),
          );
          await expect(editor(page, capturing.record_id)).toHaveValue(label);
          await expect(editor(page, capturing.record_id)).toBeFocused();
        }
        // Tab must join the identical follow-on patch already in flight.
        await page.keyboard.press("Tab");
        patchGate?.release();
        const created = await waitForViewRowByCell(
          page,
          incidentId,
          timelineViewSchemaId,
          synopsis,
          label,
        );
        expect(created.record_id, JSON.stringify(submissions)).not.toBe(id);
        await expect(cell(page, created.record_id, source)).toBeFocused();
        await expect(editor(page, created.record_id)).toHaveCount(0);
        if (patchGate !== null) expect(patchGate.hitCount()).toBe(1);
        await page.keyboard.type(`Source for ${label}`);
        await page.keyboard.press("Shift+Tab");
        await expect(cell(page, created.record_id)).toBeFocused();
        await page.keyboard.press("Enter");
        await waitForViewRowByCell(
          page,
          incidentId,
          timelineViewSchemaId,
          source,
          `Source for ${label}`,
        );
      } finally {
        await createGate?.dispose();
        await patchGate?.dispose();
      }
    }
    expect(
      await queryViewRows(page, incidentId, timelineViewSchemaId),
    ).toHaveLength(3);
    await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
    const navigationGate = await holdBrowserRequest(page, {
      method: "POST",
      path,
    });
    try {
      await page
        .getByTestId(draftCellTestId(synopsis))
        .fill("Accepted draft navigation");
      await navigationGate.waitForHit;
      await page.keyboard.press("Enter");
      navigationGate.release();
      const created = await waitForViewRowByCell(
        page,
        incidentId,
        timelineViewSchemaId,
        synopsis,
        "Accepted draft navigation",
      );
      await expect(editor(page, created.record_id)).toHaveCount(0);
      await expect(page.getByTestId(draftCellTestId(synopsis))).toBeFocused();
      expect(await fetchRecordHistoryCount(page, created.record_id)).toBe(1);
      expect(
        await queryViewRows(page, incidentId, timelineViewSchemaId),
      ).toHaveLength(4);
    } finally {
      await navigationGate.dispose();
    }
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  } finally {
    await held.dispose();
  }
});

test("Timeline uncertain creation replays exactly through refresh without losing input", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("GRID-RECOVER"),
    "Timeline uncertain capture recovery",
  );
  await openTimeline(page, incidentId);
  const path = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`;
  const attempts: string[] = [];
  const changeSets: string[] = [];
  let createdRecordId = "";
  let releaseReplay!: () => void;
  const replayGate = new Promise<void>((resolve) => {
    releaseReplay = resolve;
  });
  let reportLoss!: () => void;
  const responseLost = new Promise<void>((resolve) => {
    reportLoss = resolve;
  });
  await page.route(`**${path}`, async (route) => {
    attempts.push(required(route.request().postData()));
    if (attempts.length > 1) await replayGate;
    const response = await route.fetch();
    expect(response.status()).toBe(attempts.length === 1 ? 201 : 200);
    const body = await response.json();
    changeSets.push(body.data.change_set_id);
    createdRecordId = body.data.row.record_id;
    if (attempts.length === 1) {
      await route.abort("connectionfailed");
      reportLoss();
    } else await route.fulfill({ response });
  });
  try {
    const draft = page.getByTestId(draftCellTestId(synopsis));
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftCellTestId(synopsis),
    });
    await draft.focus();
    await page.keyboard.insertText("Uncertain fact");
    await responseLost;
    await expect(draft).toHaveValue("Uncertain fact");
    await expect(draft).toBeFocused();
    await page.keyboard.type(" still typing");
    const queryPath = `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
    const refreshed = page.waitForResponse(
      (response) => response.url().endsWith(queryPath) && response.ok(),
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("refresh-uncertain"),
      [synopsis]: "Other analyst fact",
    });
    await refreshed;
    await expect(draft).toHaveValue("Uncertain fact still typing");
    await expect(draft).toBeFocused();
    releaseReplay();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(attempts.length).toBeGreaterThanOrEqual(2);
    expect(new Set(attempts).size).toBe(1);
    expect(new Set(changeSets).size).toBe(1);
    await expect(editor(page, createdRecordId)).toHaveValue(
      "Uncertain fact still typing",
    );
    await expect(editor(page, createdRecordId)).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(cell(page, createdRecordId)).toBeFocused();
    await expect(
      page.getByTestId(rowCellTestId(createdRecordId, synopsis)),
    ).toHaveText("Uncertain fact still typing");
    const saved = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(saved).toHaveLength(2);
    expect(
      saved.find((row) => row.record_id === createdRecordId)?.cells[synopsis]
        ?.value,
    ).toBe("Uncertain fact still typing");
    expect(await fetchRecordHistoryCount(page, createdRecordId)).toBe(2);
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  } finally {
    releaseReplay();
    if (!page.isClosed()) await page.unroute(`**${path}`);
  }
});
