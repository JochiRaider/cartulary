import {
  changeGrouping,
  collapseGridGroup,
  expandGridGroup,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  conflictMarkerTestId,
  draftCellTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  rowCellTestId,
  saveStateTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineScalarEditorTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { installPatchController } from "./support/collaboration/replay";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";

const synopsis = "timeline.activity_synopsis_text";
const source = "timeline.data_source_text";
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
const grid = (page: Page) => page.locator(gridScrollportSelector());
const selected = (page: Page) =>
  grid(page).locator('[role="gridcell"][aria-selected="true"]');
const preview = (page: Page) =>
  grid(page).locator(".cartulary-grid-cell-is-range-preview");
function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Missing fixture target");
  return value;
}
async function point(target: Locator) {
  return target.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const port = element
      .closest('[role="grid"], [role="treegrid"]')
      ?.getBoundingClientRect();
    const left = Math.max(0, rect.left, port?.left ?? 0),
      right = Math.min(
        window.innerWidth,
        rect.right,
        port?.right ?? window.innerWidth,
      );
    const top = Math.max(0, rect.top, port?.top ?? 0),
      bottom = Math.min(
        window.innerHeight,
        rect.bottom,
        port?.bottom ?? window.innerHeight,
      );
    if (right <= left || bottom <= top)
      throw new Error(
        "Pointer target must be visible in the current grid viewport",
      );
    return { x: (left + right) / 2, y: (top + bottom) / 2 };
  });
}
async function reveal(page: Page, id: string, field = synopsis) {
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId: id,
    cellKey: field,
  });
}
async function seed(page: Page, count = 4) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("RANGE"),
    "Timeline range selection regression",
  );
  for (let i = 0; i < count; i++)
    await createViewRow(page, incident, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("range-seed"),
      [synopsis]: `Range fact ${i}`,
      [source]: `Range source ${i}`,
    });
  const rows = await queryViewRows(page, incident, timelineViewSchemaId);
  await page.goto(`/?incident_id=${incident}`);
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
  return { incident, rows, ids: rows.map((row) => row.record_id) };
}
async function drag(
  page: Page,
  from: Locator,
  to: Locator,
  release = true,
  expectPreview = true,
) {
  const a = await point(from),
    b = await point(to);
  await page.mouse.move(a.x, a.y);
  await page.mouse.down();
  await page.mouse.move(b.x, b.y, { steps: 6 });
  if (expectPreview) await expect(preview(page).first()).toBeVisible();
  if (release) await page.mouse.up();
}
async function dimensions(page: Page, rows: number, columns: number) {
  await expect(
    page.getByRole("status").filter({ hasText: /^Selected / }),
  ).toHaveText(`Selected ${rows} rows by ${columns} columns.`);
  await expect(selected(page)).toHaveCount(rows * columns);
}
async function copy(page: Page) {
  await page.keyboard.press("Control+c");
  return page.evaluate(async () => {
    const item = (await navigator.clipboard.read())[0];
    return item
      ? {
          plain: await (await item.getType("text/plain")).text(),
          html: await (await item.getType("text/html")).text(),
        }
      : null;
  });
}

test("Timeline pointer rectangles retain anchors and match keyboard copy and fill", async ({
  page,
}) => {
  const f = await seed(page);
  const [a, b, c] = f.ids.map(required);
  const first = required(a),
    second = required(b),
    third = required(c);
  await reveal(page, first, source);
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      request.url().includes("/bulk-mutations") ||
      request.url().includes("/clipboard-paste")
    )
      mutations.push(request.url());
  });
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  // Start with no anchor, then extend the pointer endpoint with the keyboard.
  await cell(page, second).click({ modifiers: ["Shift"] });
  await dimensions(page, 1, 1);
  await cell(page, third, source).click({ modifiers: ["Shift"] });
  await dimensions(page, 2, 2);
  await page.keyboard.press("Shift+ArrowUp");
  await dimensions(page, 1, 2);
  await page.keyboard.press("Shift+ArrowDown");
  await dimensions(page, 2, 2);
  // Forward, reverse, horizontal, vertical and rectangular pointer ranges.
  for (const [from, to, r, c] of [
    [cell(page, first), cell(page, third, source), 3, 2],
    [cell(page, third, source), cell(page, first), 3, 2],
    [cell(page, first), cell(page, first, source), 1, 2],
    [cell(page, first), cell(page, third), 3, 1],
  ] as const) {
    await drag(page, from, to);
    await dimensions(page, r, c);
    await expect(to).toBeFocused();
    await expect(preview(page)).toHaveCount(0);
  }
  const pointerCopy = await copy(page);
  await cell(page, first).click();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+ArrowDown");
  await page.keyboard.press("Shift+ArrowDown");
  expect(await copy(page)).toEqual(pointerCopy);
  expect(mutations).toHaveLength(0);
  // Shift intent is sampled at down; releasing it does not retire a completed range.
  await page.keyboard.down("Shift");
  await drag(page, cell(page, second), cell(page, third, source), false);
  await page.keyboard.up("Shift");
  await page.mouse.up();
  await dimensions(page, 3, 2);
  // A modifier added later cancels the candidate and preserves that completed range.
  await drag(page, cell(page, second), cell(page, third), false);
  await page.keyboard.down("Control");
  await page.keyboard.up("Control");
  await page.mouse.up();
  await dimensions(page, 3, 2);
  await drag(page, cell(page, second), cell(page, third), false);
  await page.keyboard.down("Shift");
  await page.mouse.up();
  await page.keyboard.up("Shift");
  await dimensions(page, 2, 1);
  // Pointer membership reaches the existing one-column fill planner.
  await drag(page, cell(page, first), cell(page, third));
  await page.keyboard.press("Control+d");
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
          (row) => row.record_id === third,
        )?.cells[synopsis]?.value,
    )
    .toBe(required(f.rows[0]).cells[synopsis]?.value);
  expect(
    mutations.filter((url) => url.includes("/bulk-mutations")),
  ).toHaveLength(1);
  expect(
    await queryViewRows(page, f.incident, timelineViewSchemaId),
  ).toHaveLength(4);
  await page.getByTestId(draftCellTestId(synopsis)).click();
  await expect(selected(page)).toHaveCount(0);
  expect(
    await queryViewRows(page, f.incident, timelineViewSchemaId),
  ).toHaveLength(4);
});

test("Timeline range departures retain pending drafts and supersede obsolete destinations", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    second = required(f.ids[1]),
    third = required(f.ids[2]);
  await reveal(page, first, source);
  await cell(page, first).click();
  await expect(editor(page, first)).toBeFocused();
  await editor(page, first).fill("pending range draft");
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${first}`,
  });
  try {
    await drag(page, cell(page, second), cell(page, third, source));
    await held.waitForHit;
    await expect(editor(page, first)).toBeFocused();
    await expect(editor(page, first)).toHaveValue("pending range draft");
    await expect(selected(page)).toHaveCount(0);
    await expect(preview(page).first()).toBeVisible();
    // A later range uses the same commit and becomes the only destination.
    await cell(page, third).click({ modifiers: ["Shift"] });
    expect(held.hitCount()).toBe(1);
    held.release();
    await dimensions(page, 3, 1);
    await expect(cell(page, third)).toBeFocused();
    await expect(editor(page, first)).toHaveCount(0);
    await expect(preview(page)).toHaveCount(0);
    expect(held.hitCount()).toBe(1);
  } finally {
    await held.dispose();
  }
  // Acceptance before release keeps the live gesture and admits its final endpoint.
  await cell(page, first).click();
  await editor(page, first).fill("accepted during drag");
  const early = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${first}`,
  });
  try {
    await drag(page, cell(page, second), cell(page, third), false);
    await early.waitForHit;
    early.release();
    await expect(editor(page, first)).toHaveCount(0);
    await expect(preview(page).first()).toBeVisible();
    const last = await point(cell(page, required(f.ids[3])));
    await page.mouse.move(last.x, last.y);
    await page.mouse.up();
    await dimensions(page, 3, 1);
    await expect(cell(page, required(f.ids[3]))).toBeFocused();
  } finally {
    await early.dispose();
  }
  // Native editor selection remains owned by the textarea/input.
  await cell(page, first).click();
  const native = required(await editor(page, first).boundingBox());
  await page.mouse.move(native.x + 10, native.y + native.height / 2);
  await page.mouse.down();
  await page.mouse.move(native.x + 120, native.y + native.height / 2, {
    steps: 5,
  });
  await page.mouse.up();
  expect(
    await editor(page, first).evaluate(
      (el: HTMLInputElement) =>
        (el.selectionEnd ?? 0) - (el.selectionStart ?? 0),
    ),
  ).toBeGreaterThan(0);
  await expect(preview(page)).toHaveCount(0);
  await editor(page, first).press("Control+a");
  expect(
    await editor(page, first).evaluate(
      (el: HTMLInputElement) =>
        (el.selectionEnd ?? 0) - (el.selectionStart ?? 0),
    ),
  ).toBe("accepted during drag".length);
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline pointer cancellation preserves completed membership and stationary editing", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    second = required(f.ids[1]),
    third = required(f.ids[2]);
  await reveal(page, first, source);
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 3, 1);
  for (const cancel of ["escape", "lost", "pointercancel", "blur"] as const) {
    await drag(page, cell(page, second), cell(page, third, source), false);
    if (cancel === "escape") await page.keyboard.press("Escape");
    else if (cancel === "blur")
      await page.evaluate(() => window.dispatchEvent(new Event("blur")));
    else {
      await grid(page).evaluate(
        (root, kind) =>
          root.addEventListener(
            "pointermove",
            (event) => {
              const id = (event as PointerEvent).pointerId;
              if (kind === "lost") root.releasePointerCapture(id);
              else
                root.dispatchEvent(
                  new PointerEvent("pointercancel", {
                    bubbles: true,
                    pointerId: id,
                  }),
                );
            },
            { once: true },
          ),
        cancel,
      );
      const end = await point(cell(page, third, source));
      await page.mouse.move(end.x + 1, end.y);
      await page.mouse.move(end.x + 2, end.y);
    }
    await page.mouse.up();
    await expect(preview(page)).toHaveCount(0);
    await dimensions(page, 3, 1);
    await expect(grid(page)).not.toHaveAttribute(
      "data-grid-pointer-selecting",
      "true",
    );
  }
  // A crossing stays a drag after reversal; exactly four pixels is still a click.
  const a = await point(cell(page, first));
  await page.mouse.move(a.x, a.y);
  await page.mouse.down();
  await page.mouse.move(a.x + 5, a.y);
  await page.mouse.move(a.x, a.y);
  await page.mouse.up();
  await expect(editor(page, first)).toHaveCount(0);
  await dimensions(page, 1, 1);
  await page.mouse.move(a.x, a.y);
  await page.mouse.down();
  await page.mouse.move(a.x + 4, a.y);
  await page.mouse.up();
  await expect(editor(page, first)).toBeFocused();
  expect(
    await editor(page, first).evaluate((el: HTMLInputElement) => ({
      start: el.selectionStart,
      end: el.selectionEnd,
      length: el.value.length,
    })),
  ).toEqual({ start: 12, end: 12, length: 12 });
  await page.keyboard.press("Escape");
  // Complete outside the scrollport at the last bounded loaded endpoint.
  await drag(page, cell(page, first), cell(page, third), false);
  const bounds = required(await grid(page).boundingBox());
  await page.mouse.move(
    bounds.x + bounds.width + 10,
    bounds.y + bounds.height + 10,
  );
  await page.mouse.up();
  await expect(preview(page)).toHaveCount(0);
  await expect(editor(page, first)).toHaveCount(0);
  await expect(grid(page)).not.toHaveAttribute(
    "data-grid-pointer-selecting",
    "true",
  );
  expect(
    await queryViewRows(page, f.incident, timelineViewSchemaId),
  ).toHaveLength(4);
});

test("Timeline ranges scroll virtualized loaded cells without querying or scrolling the document", async ({
  page,
}) => {
  const f = await seed(page, 405);
  const first = required(f.ids[0]);
  await reveal(page, first);
  const queries: string[] = [];
  page.on("request", (request) => {
    if (request.url().includes(`/views/${timelineViewSchemaId}/query`))
      queries.push(request.url());
  });
  const port = grid(page);
  const root = required(await port.boundingBox());
  const start = await point(cell(page, first));
  const documentScroll = await page.evaluate(() => ({
    x: window.scrollX,
    y: window.scrollY,
  }));
  await page.mouse.move(start.x, start.y);
  await page.mouse.down();
  await page.mouse.move(start.x, root.y + root.height - 2, { steps: 6 });
  await expect
    .poll(() => port.evaluate((el) => el.scrollTop))
    .toBeGreaterThan(900);
  await expect(cell(page, first)).toHaveCount(0);
  await expect
    .poll(() =>
      port.evaluate((el) => el.scrollHeight - el.clientHeight - el.scrollTop),
    )
    .toBeLessThan(2);
  await page.mouse.up();
  await expect(
    page.getByRole("status").filter({ hasText: /^Selected / }),
  ).toHaveText("Selected 100 rows by 1 columns.");
  expect(queries).toHaveLength(0);
  expect(
    await page.evaluate(() => ({ x: window.scrollX, y: window.scrollY })),
  ).toEqual(documentScroll);
  const end = await port.evaluate((el) => el.scrollTop);
  await page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  );
  expect(await port.evaluate((el) => el.scrollTop)).toBe(end);
  // Completed vertical cycles reveal the unmounted first member without fetching.
  await page.keyboard.press("Enter");
  await expect(cell(page, first)).toBeFocused();
  await page.keyboard.press("Shift+Enter");
  await expect(cell(page, required(f.ids[99]))).toBeFocused();
  expect(queries).toHaveLength(0);
  // Horizontal virtualization uses the same mounted semantic-cell registry.
  await reveal(page, first);
  const origin = await point(cell(page, first));
  const bounds = required(await port.boundingBox());
  await page.mouse.move(origin.x, origin.y);
  await page.mouse.down();
  await page.mouse.move(bounds.x + bounds.width - 2, origin.y, { steps: 6 });
  await expect
    .poll(() =>
      port.evaluate((el) => el.scrollWidth - el.clientWidth - el.scrollLeft),
    )
    .toBeLessThan(2);
  await page.mouse.up();
  await expect(preview(page)).toHaveCount(0);
  await expect(
    page.getByRole("status").filter({ hasText: /^Selected / }),
  ).toContainText("Selected 1 rows by");
  expect(queries).toHaveLength(0);
  expect(
    await page.evaluate(() => ({ x: window.scrollX, y: window.scrollY })),
  ).toEqual(documentScroll);
  await page.keyboard.press("Tab");
  await expect(cell(page, first)).toBeFocused();
  expect(queries).toHaveLength(0);
  // Explicit paging may append compatible members, then evict the captured
  // source. The gesture itself never requests either transition.
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  const rangeStatus = page
    .getByRole("status")
    .filter({ hasText: /^Selected / });
  const completed = await rangeStatus.textContent();
  for (const count of [200, 300]) {
    await more.click();
    await expect(controls).toContainText(`${count} records loaded`);
    await expect(rangeStatus).toHaveText(required(completed));
  }
  await more.click();
  const earlier = controls.getByRole("button", { name: "Earlier rows" });
  await expect(earlier).toHaveAttribute("aria-disabled", "false");
  await expect(selected(page)).toHaveCount(0);
  await expect(rangeStatus).toHaveCount(0);
  await earlier.click();
  await expect(controls).toContainText("100 records loaded; more available.");
  // Unmount during a live gesture disposes its capture and frame work.
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, required(f.ids[1])), false);
  await page.goto("about:blank");
  await page.mouse.up();
  expect(
    await page.locator('[data-grid-pointer-selecting="true"]').count(),
  ).toBe(0);
});

test("Timeline ranges respect columns groups inspector context and bulk checkboxes", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    third = required(f.ids[2]);
  await clickTimelineRowAction(
    page,
    third,
    timelineRowMarkReviewedButtonTestId(third),
  );
  await reveal(page, first, source);
  await cell(page, first).click();
  await page.keyboard.press("Escape");
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  const inspector = page.getByTestId(timelineInspectorTestId());
  await expect(inspector).toBeVisible();
  const subject = await inspector.innerText();
  await reveal(page, first, source);
  await drag(page, cell(page, first), cell(page, third, source));
  await dimensions(page, 3, 2);
  expect(await inspector.innerText()).toBe(subject);
  for (let step = 0; step < 3; step++) await page.keyboard.press("Tab");
  await page.keyboard.press("F2");
  await expect(editor(page, required(f.ids[1]))).toBeFocused();
  expect(await inspector.innerText()).toBe(subject);
  await page.keyboard.press("Escape");
  await dimensions(page, 3, 2);
  await page.keyboard.press("Escape");
  await expect(selected(page)).toHaveCount(1);
  await expect(inspector).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(inspector).toHaveCount(0);
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(inspector).toBeVisible();
  await reveal(page, first, source);
  await drag(page, cell(page, first), cell(page, third, source));
  await dimensions(page, 3, 2);
  await page
    .getByRole("checkbox", { name: `Select record ${first}`, exact: true })
    .check();
  await dimensions(page, 3, 2);
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  const menu = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  const toggle = page.getByTestId(
    workbookColumnsMenuTriggerTestId(timelineViewSchemaId),
  );
  const label = required(
    requireViewContract(timelineViewSchemaId).fieldMap[source],
  ).label;
  await toggle.click();
  await menu.getByRole("checkbox", { name: label, exact: true }).uncheck();
  await menu
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await expect(selected(page)).toHaveCount(0);
  await toggle.click();
  await menu.getByRole("checkbox", { name: label, exact: true }).check();
  await menu
    .getByRole("button", { name: `Move ${label} earlier`, exact: true })
    .click();
  await menu
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await reveal(page, first);
  await drag(page, cell(page, first, source), cell(page, third));
  await dimensions(page, 3, 2);
  await page.keyboard.press("Tab");
  await expect(cell(page, first, source)).toBeFocused();
  await dimensions(page, 3, 2);
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  await expect(selected(page)).toHaveCount(0);
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 4, 1);
  const bulk: string[] = [];
  page.on("request", (request) => {
    if (request.url().includes("/bulk-mutations")) bulk.push(request.url());
  });
  await page.keyboard.press("Control+d");
  expect(bulk).toHaveLength(0);
  const groupId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await collapseGridGroup({
    page,
    surface: timelineViewSchemaId,
    groupTestId: groupId,
  });
  await expect(selected(page)).toHaveCount(0);
  await expect(cell(page, third)).toHaveCount(0);
  await expandGridGroup({
    page,
    surface: timelineViewSchemaId,
    groupTestId: groupId,
  });
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 4, 1);
  await expect(
    page.getByRole("checkbox", { name: `Select record ${first}`, exact: true }),
  ).toBeChecked();
});

test("Timeline range rejection and accepted query replacement preserve authoritative editing", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    second = required(f.ids[1]),
    third = required(f.ids[2]);
  await reveal(page, first);
  const patches = await installPatchController(page);
  try {
    // Escape cancels the tentative destination without discarding the active draft.
    await cell(page, first).click();
    await editor(page, first).fill("accepted after range cancellation");
    const held = patches.holdNextPatch({ recordId: first });
    await drag(page, cell(page, second), cell(page, third), false);
    await held.waitForHit;
    await page.keyboard.press("Escape");
    await page.mouse.up();
    await expect(editor(page, first)).toHaveValue(
      "accepted after range cancellation",
    );
    held.release();
    await held.waitForCompletion;
    await expect(preview(page)).toHaveCount(0);
    await expect(selected(page)).toHaveCount(0);
  } finally {
    await patches.dispose();
  }
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 3, 1);
  const pending = await holdBrowserRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${f.incident}/views/${timelineViewSchemaId}/query`,
  });
  try {
    await sortByHeader(page, timelineViewSchemaId, synopsis);
    await pending.waitForHit;
    await dimensions(page, 3, 1);
    pending.release();
    await expect(selected(page)).toHaveCount(0);
  } finally {
    await pending.dispose();
  }
  const rejected = await installPatchController(page);
  try {
    await reveal(page, first);
    await cell(page, first).click();
    await editor(page, first).fill("  exact rejected Ω draft  ");
    rejected.failNextPatch(422, "invalid_request", { recordId: first });
    await drag(page, cell(page, second), cell(page, third), true, false);
    await expect(editor(page, first)).toBeFocused();
    await expect(editor(page, first)).toHaveValue("  exact rejected Ω draft  ");
    await expect(preview(page)).toHaveCount(0);
    await expect(selected(page)).toHaveCount(0);
    expect(rejected.calls).toHaveLength(1);
  } finally {
    await rejected.dispose();
  }
});

test("Timeline range feedback retains focus noncolor cues zoom and text spacing", async ({
  page,
}, info) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    third = required(f.ids[2]);
  for (const zoom of [1, 2]) {
    await page.setViewportSize({ width: 1440 * zoom, height: 900 * zoom });
    await page.evaluate((value) => {
      document.documentElement.style.zoom = String(value);
    }, zoom);
    const spacing = await page.addStyleTag({
      content:
        "* { letter-spacing: 0.12em !important; word-spacing: 0.16em !important; line-height: 1.5 !important; }",
    });
    await reveal(page, first);
    await drag(page, cell(page, first), cell(page, third), false);
    expect(
      await preview(page)
        .first()
        .evaluate((el) => getComputedStyle(el).outlineStyle),
    ).toBe("dashed");
    await page.mouse.up();
    await dimensions(page, 3, 1);
    await expect(cell(page, third)).toBeFocused();
    const cue = await cell(page, first).evaluate((el) => ({
      image: getComputedStyle(el).backgroundImage,
      shadow: getComputedStyle(el).boxShadow,
    }));
    expect(cue.image).toContain("repeating-linear-gradient");
    expect(cue.shadow).not.toBe("none");
    await expect(
      page.getByRole("status").filter({ hasText: /^Selected / }),
    ).toHaveCount(1);
    await page.keyboard.press("Enter");
    await expect(cell(page, first)).toBeFocused();
    await page.keyboard.press("F2");
    await expect(editor(page, first)).toBeFocused();
    await dimensions(page, 3, 1);
    await page.keyboard.press("Escape");
    await dimensions(page, 3, 1);
    await page.keyboard.press("Escape");
    await expect(selected(page)).toHaveCount(1);
    await page.keyboard.press("Shift+ArrowDown");
    await page.keyboard.press("Shift+ArrowDown");
    await dimensions(page, 3, 1);
    await info.attach(`timeline-range-zoom-${zoom}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    await spacing.evaluate((el) => el.parentNode?.removeChild(el));
  }
});

test("Timeline ranges cancel at authority transitions and remain readable without mutation rights", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    third = required(f.ids[2]);
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, third), false);
  const before = await currentLifecycle(page, f.incident);
  expect(
    (
      await lifecycleAction(page, f.incident, "closeIncident", {
        client_txn_id: uniqueTxn("range-close"),
        base_incident_version: before.incident_version,
        reason: "Range authority regression",
      })
    ).ok,
  ).toBe(true);
  await expect(grid(page)).toHaveAttribute("aria-readonly", "true");
  await expect(preview(page)).toHaveCount(0);
  await page.mouse.up();
  await expect(selected(page)).toHaveCount(0);
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      request.url().includes("/bulk-mutations")
    )
      mutations.push(request.url());
  });
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 3, 1);
  await page.keyboard.press("Enter");
  await expect(cell(page, first)).toBeFocused();
  await page.keyboard.press("F2");
  await expect(editor(page, first)).toHaveCount(0);
  await dimensions(page, 3, 1);
  await page.keyboard.press("Control+d");
  await cell(page, first).click();
  await expect(editor(page, first)).toHaveCount(0);
  expect(mutations).toHaveLength(0);
});

test("Timeline completed ranges retain value refreshes and invalidate deleted membership", async ({
  page,
}) => {
  const f = await seed(page);
  const first = required(f.ids[0]),
    second = required(f.ids[1]),
    third = required(f.ids[2]);
  await reveal(page, first);
  await drag(page, cell(page, first), cell(page, third));
  await dimensions(page, 3, 1);
  await patchRecord(page, second, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: required(f.rows[1]).row_version,
    client_txn_id: uniqueTxn("range-peer-update"),
    changes: [{ field_key: synopsis, value: "Peer value update" }],
  });
  await expect(cell(page, second)).toHaveText("Peer value update");
  await dimensions(page, 3, 1);
  const refresh = page
    .getByRole("group", { name: "Workbook browsing" })
    .getByRole("button", { name: "Refresh", exact: true });
  const query = `**/api/v1/incidents/${f.incident}/views/${timelineViewSchemaId}/query`;
  await page.route(query, (route) => route.abort("failed"), { times: 1 });
  await refresh.click();
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).toContainText("Retry");
  await dimensions(page, 3, 1);
  await page
    .getByRole("group", { name: "Workbook browsing" })
    .getByRole("button", { name: "Retry", exact: true })
    .click();
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).not.toContainText("Retry");
  await dimensions(page, 3, 1);
  const current = required(
    (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
      (row) => row.record_id === second,
    ),
  );
  const removed = await publicHttpOperation({
    operationID: "deleteRecord",
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    pathParameters: { record_id: second },
    body: {
      base_row_version: current.row_version,
      client_txn_id: uniqueTxn("range-peer-delete"),
      reason: "Selection membership regression",
    },
  });
  expect(removed.ok).toBe(true);
  await expect(cell(page, second)).toHaveCount(0);
  await expect(selected(page)).toHaveCount(0);
  expect(
    await queryViewRows(page, f.incident, timelineViewSchemaId),
  ).toHaveLength(3);
});

test("Timeline range entry cycles both orders preserves geometry and exits accessibly", async ({
  page,
}) => {
  const f = await seed(page, 3);
  const a = required(f.ids[0]),
    b = required(f.ids[1]);
  await reveal(page, a, source);
  const requests: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      request.url().includes(`/views/${timelineViewSchemaId}/query`)
    )
      requests.push(request.url());
  });
  for (const reverse of [false, true]) {
    for (const shape of ["rectangle", "row", "column"] as const) {
      const endRow = shape === "row" ? a : b;
      const endField = shape === "column" ? synopsis : source;
      await drag(
        page,
        reverse ? cell(page, endRow, endField) : cell(page, a),
        reverse ? cell(page, a) : cell(page, endRow, endField),
      );
      await expect(
        reverse ? cell(page, a) : cell(page, endRow, endField),
      ).toBeFocused();
      const rows = shape === "row" ? [a] : [a, b];
      const fields = shape === "column" ? [synopsis] : [synopsis, source];
      for (const key of ["Tab", "Shift+Tab", "Enter", "Shift+Enter"]) {
        const order = key.includes("Tab")
          ? rows.flatMap((id) => fields.map((field) => [id, field] as const))
          : fields.flatMap((field) => rows.map((id) => [id, field] as const));
        let index = reverse ? 0 : order.length - 1;
        for (let step = 0; step < order.length; step++) {
          index =
            (index + (key.startsWith("Shift") ? -1 : 1) + order.length) %
            order.length;
          await page.keyboard.press(key);
          const next = required(order[index]);
          await expect(cell(page, next[0], next[1])).toBeFocused();
          await dimensions(page, rows.length, fields.length);
        }
      }
    }
  }
  // Build with keyboard and extend from the traversed member, keeping A1 anchor.
  await cell(page, a).click();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+ArrowRight");
  await page.keyboard.press("Shift+ArrowDown");
  await dimensions(page, 2, 2);
  await page.keyboard.press("Tab");
  await expect(cell(page, a)).toBeFocused();
  await page.keyboard.press("Shift+ArrowRight");
  await dimensions(page, 1, 2);
  await expect(cell(page, a, source)).toBeFocused();
  await expect(grid(page)).toHaveAttribute(
    "aria-description",
    /loaded window.*Escape returns to a single active cell/,
  );
  await page.keyboard.press("Escape");
  await expect(selected(page)).toHaveCount(1);
  await expect(cell(page, a, source)).toBeFocused();
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "Selection collapsed to the active cell." }),
  ).toHaveCount(1);
  await page.keyboard.press("Shift+ArrowLeft");
  await dimensions(page, 1, 2);
  expect(requests).toHaveLength(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
});

test("Timeline retained range editing gates rejection latest keys and newer focus with one write", async ({
  page,
}) => {
  const f = await seed(page, 3);
  const a = required(f.ids[0]),
    b = required(f.ids[1]);
  await reveal(page, a, source);
  await drag(page, cell(page, a), cell(page, b, source));
  await page.keyboard.press("Tab");
  await expect(cell(page, a)).toBeFocused();
  await page.keyboard.press("F2");
  await expect(editor(page, a)).toHaveValue(
    String(required(f.rows[0]).cells[synopsis]?.value),
  );
  expect(
    await editor(page, a).evaluate((el: HTMLInputElement) => [
      el.selectionStart,
      el.selectionEnd,
    ]),
  ).toEqual([12, 12]);
  await page.keyboard.press("Home");
  await page.keyboard.press("Shift+ArrowRight");
  expect(
    await editor(page, a).evaluate((el: HTMLInputElement) => [
      el.selectionStart,
      el.selectionEnd,
    ]),
  ).toEqual([0, 1]);
  await page.keyboard.press("Escape");
  await dimensions(page, 2, 2);
  await page.keyboard.type("replacement Ω");
  await expect(editor(page, a)).toHaveValue("replacement Ω");
  await dimensions(page, 2, 2);
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${a}`,
  });
  const submitted: unknown[] = [];
  page.on("request", (request) => {
    if (request.method() === "PATCH" && request.url().endsWith(`/records/${a}`))
      submitted.push(request.postDataJSON());
  });
  try {
    await page.keyboard.press("Enter");
    await held.waitForHit;
    await page.keyboard.press("Tab");
    await page.keyboard.press("Tab");
    await expect(editor(page, a)).toBeFocused();
    expect(held.hitCount()).toBe(1);
    held.release();
    await expect(cell(page, a, source)).toBeFocused();
    await dimensions(page, 2, 2);
    expect(held.hitCount()).toBe(1);
    expect(submitted).toEqual([
      expect.objectContaining({
        changes: [
          expect.objectContaining({
            field_key: synopsis,
            value: "replacement Ω",
          }),
        ],
      }),
    ]);
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  } finally {
    await held.dispose();
  }
  // A later external control owns focus even if the pending write succeeds.
  await page.keyboard.press("F2");
  await editor(page, a, source).fill("accepted without late focus");
  const later = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${a}`,
  });
  try {
    await page.keyboard.press("Tab");
    await later.waitForHit;
    const toggle = page.getByTestId(
      workbookColumnsMenuTriggerTestId(timelineViewSchemaId),
    );
    await toggle.click();
    later.release();
    await expect(cell(page, a, source)).toContainText(
      "accepted without late focus",
    );
    await expect(
      page.getByTestId(workbookColumnsMenuTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(cell(page, b)).not.toBeFocused();
    expect(later.hitCount()).toBe(1);
  } finally {
    await later.dispose();
  }
  await page
    .getByTestId(workbookColumnsMenuTestId(timelineViewSchemaId))
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await reveal(page, a, source);
  await drag(page, cell(page, a), cell(page, b, source));
  await page.keyboard.press("Tab");
  await page.keyboard.press("F2");
  await editor(page, a).fill(" rejected exact Ω ");
  const patches = await installPatchController(page);
  try {
    patches.failNextPatch(422, "invalid_request", { recordId: a });
    await page.keyboard.press("Tab");
    await expect(editor(page, a)).toBeFocused();
    await expect(editor(page, a)).toHaveValue(" rejected exact Ω ");
    await expect.poll(() => patches.calls.length).toBe(1);
    await dimensions(page, 2, 2);
  } finally {
    await patches.dispose();
  }
});

test("Timeline range entry preserves multiline composition Find restoration and single cell clear", async ({
  page,
}) => {
  const f = await seed(page, 3);
  const a = required(f.ids[0]),
    b = required(f.ids[1]);
  const raw = "timeline.raw_activity_text";
  await reveal(page, a, raw);
  await drag(page, cell(page, a, raw), cell(page, b, raw));
  await page.keyboard.press("Enter");
  await expect(cell(page, a, raw)).toBeFocused();
  await page.keyboard.press("F2");
  const input = editor(page, a, raw);
  await input.fill("first");
  await page.keyboard.press("Shift+Enter");
  await page.keyboard.type("second");
  await expect(input).toHaveValue("first\nsecond");
  await dimensions(page, 2, 1);
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "PATCH") mutations.push(request.url());
  });
  await input.dispatchEvent("compositionstart");
  for (const key of ["Enter", "Tab", "Escape"])
    await input.dispatchEvent("keydown", {
      key,
      isComposing: true,
      bubbles: true,
      cancelable: true,
    });
  await expect(input).toBeFocused();
  await expect(input).toHaveValue("first\nsecond");
  expect(mutations).toHaveLength(0);
  await input.dispatchEvent("compositionend");
  await page.keyboard.press("Tab");
  await expect(cell(page, b, raw)).toBeFocused();
  await dimensions(page, 2, 1);
  expect(mutations).toHaveLength(1);
  await page.keyboard.press("Enter");
  await expect(cell(page, a, raw)).toBeFocused();
  await page.keyboard.press("Control+f");
  await expect(
    page.getByRole("textbox", { name: "Find in loaded rows", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(cell(page, a, raw)).toBeFocused();
  await dimensions(page, 2, 1);
  await page.keyboard.press("Backspace");
  await expect(editor(page, a, raw)).toBeFocused();
  await expect(selected(page)).toHaveCount(0);
  expect(mutations).toHaveLength(1);
  await page.keyboard.press("Escape");
  await expect(cell(page, a, raw)).toContainText("first");
});

test("Timeline pending range entry retains conflicts and cancels deleted destinations and lost authority", async ({
  page,
}) => {
  for (const transition of ["conflict", "deletion", "closure"] as const) {
    const f = await seed(page, 3),
      a = required(f.ids[0]),
      b = required(f.ids[1]);
    await reveal(page, a);
    await drag(page, cell(page, a), cell(page, b));
    await page.keyboard.press("Enter");
    await expect(cell(page, a)).toBeFocused();
    await page.keyboard.press("F2");
    await editor(page, a).fill(` exact ${transition} draft `);
    const held = await holdBrowserRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${a}`,
    });
    try {
      await page.keyboard.press("Enter");
      await held.waitForHit;
      await expect(editor(page, a)).toBeFocused();
      await dimensions(page, 2, 1);
      if (transition === "conflict") {
        await patchRecord(page, a, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: required(f.rows[0]).row_version,
          client_txn_id: uniqueTxn("range-entry-conflict"),
          changes: [{ field_key: synopsis, value: "Peer accepted value" }],
        });
      } else if (transition === "deletion") {
        const removed = await publicHttpOperation({
          operationID: "deleteRecord",
          request: atJsonOrigin(page.request, apiBase),
          headers: await csrfHeaders(page),
          pathParameters: { record_id: b },
          body: {
            base_row_version: required(f.rows[1]).row_version,
            client_txn_id: uniqueTxn("range-entry-delete"),
            reason: "Pending semantic destination regression",
          },
        });
        expect(removed.ok).toBe(true);
        await expect(cell(page, b)).toHaveCount(0);
        await expect(selected(page)).toHaveCount(0);
      } else {
        const before = await currentLifecycle(page, f.incident);
        expect(
          (
            await lifecycleAction(page, f.incident, "closeIncident", {
              client_txn_id: uniqueTxn("range-entry-close"),
              base_incident_version: before.incident_version,
              reason: "Pending range authority regression",
            })
          ).ok,
        ).toBe(true);
        await expect(grid(page)).toHaveAttribute("aria-readonly", "true");
        await expect(editor(page, a)).toHaveCount(0);
        await expect(selected(page)).toHaveCount(0);
      }
      held.release();
      if (transition === "conflict") {
        await expect(editor(page, a)).toHaveValue(" exact conflict draft ");
        await expect(editor(page, a)).toBeFocused();
        await dimensions(page, 2, 1);
        await expect(
          page.getByTestId(conflictMarkerTestId(a, synopsis)),
        ).toBeVisible();
      } else if (transition === "deletion") {
        await expect(cell(page, a)).toContainText("exact deletion draft");
        await expect(editor(page, a)).toHaveCount(0);
        await expect(cell(page, required(f.ids[2]))).not.toBeFocused();
        await expect(selected(page)).toHaveCount(0);
      } else {
        await expect(grid(page)).toHaveAttribute("aria-readonly", "true");
        await expect(editor(page, a)).toHaveCount(0);
        await expect(cell(page, b)).not.toBeFocused();
      }
      expect(held.hitCount()).toBe(1);
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    } finally {
      await held.dispose();
    }
  }
});

test("Timeline clear preserves reversed rectangles and commits one nullable batch", async ({
  page,
}) => {
  const f = await seed(page, 3),
    first = required(f.ids[0]),
    second = required(f.ids[1]);
  const attempts: string[] = [],
    patches: string[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations"))
      attempts.push(required(request.postData()));
    if (request.method() === "PATCH") patches.push(request.url());
  });
  await reveal(page, first, source);
  await page
    .getByRole("checkbox", { name: `Select record ${first}`, exact: true })
    .check();
  await drag(page, cell(page, second, source), cell(page, first));
  await dimensions(page, 2, 2);
  const response = page.waitForResponse(
    (response) =>
      response.url().endsWith("/bulk-mutations") &&
      response.request().method() === "POST",
  );
  await page.keyboard.down("Delete");
  await page.keyboard.down("Delete");
  await page.keyboard.up("Delete");
  const result = await (await response).json();
  expect(result.data.rows).toHaveLength(2);
  expect(result.data.conflicts).toEqual([]);
  expect(result.data.change_set_id).toBeTruthy();
  expect(attempts).toHaveLength(1);
  expect(JSON.parse(required(attempts[0]))).toMatchObject({
    kind: "clear_cells_v1",
    view_schema_id: timelineViewSchemaId,
    field_keys: [synopsis, source],
    targets: [
      { record_id: first, base_row_version: 1 },
      { record_id: second, base_row_version: 1 },
    ],
  });
  expect(JSON.parse(required(attempts[0]))).not.toHaveProperty("value");
  await expect
    .poll(async () =>
      (await queryViewRows(page, f.incident, timelineViewSchemaId))
        .filter((row) => [first, second].includes(row.record_id))
        .map((row) => [
          row.cells[synopsis]?.value,
          row.cells[source]?.value,
          row.row_version,
        ]),
    )
    .toEqual([
      [null, null, 2],
      [null, null, 2],
    ]);
  await dimensions(page, 2, 2);
  await expect(cell(page, first)).toBeFocused();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(
    page.getByRole("checkbox", { name: `Select record ${first}`, exact: true }),
  ).toBeChecked();
  const noOpResponse = page.waitForResponse(
    (response) =>
      response.url().endsWith("/bulk-mutations") &&
      response.request().method() === "POST",
  );
  const action = page.getByRole("button", {
    name: "Clear contents",
    exact: true,
  });
  await action.focus();
  await page.keyboard.press("Enter");
  const noOp = await (await noOpResponse).json();
  expect(noOp.data.rows).toEqual([]);
  expect(noOp.data.change_set_id).toBeUndefined();
  await dimensions(page, 2, 2);
  await expect(action).toBeFocused();
  expect(attempts).toHaveLength(2);
  expect(JSON.parse(required(attempts[0])).client_txn_id).not.toBe(
    JSON.parse(required(attempts[1])).client_txn_id,
  );
  const duplicateResponse = page.waitForResponse((response) =>
    response.url().endsWith("/bulk-mutations"),
  );
  await action.evaluate((button) => {
    const delivery = new MouseEvent("click", { bubbles: true });
    button.dispatchEvent(delivery);
    button.dispatchEvent(delivery);
  });
  await duplicateResponse;
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  expect(attempts).toHaveLength(3);
  expect(patches).toEqual([]);
});

test("Timeline clear rejects unsubmitted authoring and preserves native Delete", async ({
  page,
}) => {
  const f = await seed(page, 2),
    first = required(f.ids[0]);
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      request.url().endsWith("/bulk-mutations")
    )
      writes.push(request.url());
  });
  await reveal(page, first);
  await cell(page, first).click();
  await editor(page, first).fill("Unsubmitted source");
  await page
    .getByRole("button", { name: "Clear contents", exact: true })
    .click();
  await expect(
    page.getByRole("alert").filter({ hasText: "blocked by unsaved work" }),
  ).toBeVisible();
  await expect(editor(page, first)).toHaveValue("Unsubmitted source");
  expect(writes).toEqual([]);
  await editor(page, first).focus();
  await page.keyboard.press("Control+a");
  await page.keyboard.press("Delete");
  await expect(editor(page, first)).toHaveValue("");
  expect(writes).toEqual([]);
  expect(
    (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
      (row) => row.record_id === first,
    )?.cells[synopsis]?.value,
  ).toBe(required(f.rows[0]).cells[synopsis]?.value);
});

test("Timeline clear rejects mixed derived membership and respects closed access", async ({
  page,
}) => {
  const f = await seed(page, 2),
    first = required(f.ids[0]);
  const writes: string[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations")) writes.push(request.url());
  });
  const field = "timeline.evidence_count";
  const menu = page.getByTestId(
    workbookColumnsMenuTestId(timelineViewSchemaId),
  );
  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  await menu
    .getByRole("checkbox", {
      name: required(requireViewContract(timelineViewSchemaId).fieldMap[field])
        .label,
      exact: true,
    })
    .check();
  await menu
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await reveal(page, first);
  await cell(page, first).click();
  await page.keyboard.press("Escape");
  await reveal(page, first, field);
  await cell(page, first, field).click({ modifiers: ["Shift"] });
  await page.keyboard.press("Delete");
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: "available, editable Timeline cells" }),
  ).toBeVisible();
  expect(writes).toEqual([]);
  const lifecycle = await currentLifecycle(page, f.incident);
  expect(
    (
      await lifecycleAction(page, f.incident, "closeIncident", {
        client_txn_id: uniqueTxn("clear-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Clear closed access",
      })
    ).ok,
  ).toBe(true);
  await expect(grid(page)).toHaveAttribute("aria-readonly", "true");
  await page
    .getByRole("button", { name: "Clear contents", exact: true })
    .click();
  expect(writes).toEqual([]);
});

test("Timeline clear captures offscreen loaded membership and orders later overlapping edits", async ({
  page,
}) => {
  const f = await seed(page, 120),
    first = required(f.ids[0]),
    last = required(f.ids[99]);
  await reveal(page, first);
  await cell(page, first).click();
  await page.keyboard.press("Escape");
  for (let i = 0; i < 99; i++) await page.keyboard.press("Shift+ArrowDown");
  await expect(cell(page, first)).toHaveCount(0);
  await expect(cell(page, last)).toBeFocused();
  await expect(
    page.getByRole("status").filter({ hasText: /^Selected / }),
  ).toHaveText("Selected 100 rows by 1 columns.");
  const path = `/api/v1/incidents/${f.incident}/views/${timelineViewSchemaId}/bulk-mutations`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  const queries: string[] = [],
    writes: { path: string; body: Record<string, unknown> }[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith(`/views/${timelineViewSchemaId}/query`))
      queries.push(request.url());
    if (request.url().endsWith(path) || request.method() === "PATCH")
      writes.push({ path: request.url(), body: request.postDataJSON() });
  });
  try {
    await page.keyboard.press("Delete");
    await held.waitForHit;
    expect(
      (writes[0]?.body.targets as { record_id: string }[]).map(
        (target) => target.record_id,
      ),
    ).toEqual(f.ids.slice(0, 100));
    expect(queries).toEqual([]);
    await expect(cell(page, last)).toContainText(
      String(required(f.rows[99]).cells[synopsis]?.value),
    );
    await page.keyboard.type("Later value");
    await page.keyboard.press("Tab");
    expect(writes).toHaveLength(1);
    held.release();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
            (row) => row.record_id === last,
          )?.cells[synopsis]?.value,
      )
      .toBe("Later value");
    expect(writes).toHaveLength(2);
    expect(writes[1]?.body.base_row_version).toBe(2);
    const saved = await queryViewRows(page, f.incident, timelineViewSchemaId);
    expect(
      saved
        .filter((row) => f.ids.slice(0, 99).includes(row.record_id))
        .every((row) => row.cells[synopsis]?.value === null),
    ).toBe(true);
    await reveal(page, last);
    await cell(page, last).click();
    await page.keyboard.press("Escape");
    const next = await holdBrowserRequest(page, { method: "POST", path });
    try {
      await page.keyboard.press("Delete");
      await next.waitForHit;
      await page.keyboard.type("Still authoring");
      await expect(editor(page, last)).toHaveValue("Still authoring");
      next.release();
      await expect
        .poll(
          async () =>
            (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
              (row) => row.record_id === last,
            )?.cells[synopsis]?.value,
        )
        .toBeNull();
      await expect(editor(page, last)).toHaveValue("Still authoring");
      await expect(editor(page, last)).toBeFocused();
      await page.keyboard.press("Tab");
      await expect
        .poll(
          async () =>
            (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
              (row) => row.record_id === last,
            )?.cells[synopsis]?.value,
        )
        .toBe("Still authoring");
    } finally {
      await next.dispose();
    }
  } finally {
    await held.dispose();
  }
});

test("Timeline clear includes expanded group records and excludes collapsed membership", async ({
  page,
}) => {
  const f = await seed(page, 4),
    first = required(f.ids[0]),
    second = required(f.ids[1]),
    third = required(f.ids[2]);
  await clickTimelineRowAction(
    page,
    first,
    timelineRowMarkReviewedButtonTestId(first),
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const groupId = gridGroupRowTestId(
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await collapseGridGroup({
    page,
    surface: timelineViewSchemaId,
    groupTestId: groupId,
  });
  await reveal(page, second, source);
  await drag(page, cell(page, second), cell(page, third, source));
  const response = page.waitForResponse(
    (response) =>
      response.url().endsWith("/bulk-mutations") &&
      response.request().method() === "POST",
  );
  await page.keyboard.press("Delete");
  const received = await response;
  expect(received.ok()).toBe(true);
  const request = received.request().postDataJSON();
  expect(
    request.targets.map((target: { record_id: string }) => target.record_id),
  ).toEqual([second, third]);
  expect(request.field_keys).toEqual([synopsis, source]);
  const saved = await queryViewRows(page, f.incident, timelineViewSchemaId);
  expect(
    saved.find((row) => row.record_id === first)?.cells[synopsis]?.value,
  ).toBe(required(f.rows[0]).cells[synopsis]?.value);
});

test("Timeline clear waits for captured autosave predecessors", async ({
  page,
}) => {
  const f = await seed(page, 2),
    first = required(f.ids[0]);
  await reveal(page, first);
  const held = await holdBrowserRequest(page, {
    method: "PATCH",
    path: `/api/v1/records/${first}`,
  });
  const writes: { method: string; body: Record<string, unknown> }[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      request.url().endsWith("/bulk-mutations")
    )
      writes.push({ method: request.method(), body: request.postDataJSON() });
  });
  try {
    await cell(page, first).click();
    await editor(page, first).fill("Prior autosave");
    await page.keyboard.press("Tab");
    await held.waitForHit;
    await page
      .getByRole("button", { name: "Clear contents", exact: true })
      .click();
    expect(writes).toHaveLength(1);
    held.release();
    await expect.poll(() => writes.length).toBe(2);
    expect(writes[1]).toMatchObject({
      method: "POST",
      body: {
        kind: "clear_cells_v1",
        targets: [{ record_id: first, base_row_version: 2 }],
      },
    });
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
            (row) => row.record_id === first,
          )?.cells[synopsis]?.value,
      )
      .toBeNull();
  } finally {
    await held.dispose();
  }
});
