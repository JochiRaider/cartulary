import {
  changeGrouping,
  collapseGridGroup,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  savedViewModifiedTestId,
  savedViewResetButtonTestId,
  savedViewSelectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookAddRowButtonTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  createSavedViewFromCurrentSurface,
  duplicateSavedViewFromCurrentSurface,
  openSavedViewActionMenu,
  readSavedView,
  selectSavedView,
  setSavedViewDraftName,
  updateSavedViewFromCurrentSurface,
} from "./support/workbook/savedViews";

const summary = "timeline.activity_synopsis_text";
const surface = timelineViewSchemaId;
const columns = (page: Page, view: string = surface) =>
  page.getByTestId(workbookColumnsMenuTestId(view));
const trigger = (page: Page, view: string = surface) =>
  page.getByTestId(workbookColumnsMenuTriggerTestId(view));
const header = (page: Page, field = summary, view: string = surface) =>
  page.getByTestId(gridSortHeaderTestId(view, field));
const width = (node: Locator) =>
  node.evaluate((element) =>
    Number.parseFloat(getComputedStyle(element).width),
  );
async function showColumns(page: Page, view: string = surface) {
  if (!(await columns(page, view).isVisible()))
    await trigger(page, view).click();
}
async function widthPanel(page: Page, field = summary, view: string = surface) {
  await showColumns(page, view);
  await columns(page, view)
    .getByRole("button", {
      name: `Width for ${requireViewContract(view).fieldMap[field]?.label}`,
      exact: true,
    })
    .click();
  return columns(page, view).getByRole("textbox", {
    name: "Width in CSS pixels",
  });
}
async function setWidth(
  page: Page,
  value: string,
  field = summary,
  view: string = surface,
) {
  const input = await widthPanel(page, field, view);
  await input.fill(value);
  await columns(page, view)
    .getByRole("button", { name: "Apply width", exact: true })
    .click();
  await columns(page, view)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await columns(page, view)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
}
async function fitVisible(page: Page, field = summary, view: string = surface) {
  await widthPanel(page, field, view);
  await columns(page, view)
    .getByRole("button", { name: "Fit visible content", exact: true })
    .click();
  await expect(
    columns(page, view).getByRole("button", {
      name: "Fit visible content",
      exact: true,
    }),
  ).toBeEnabled();
  await expect(
    page.getByRole("status").filter({ hasText: /fitted to/ }),
  ).toBeVisible();
  await columns(page, view)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await columns(page, view)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  return width(header(page, field, view));
}
async function firstColumn(
  page: Page,
  field = summary,
  view: string = surface,
) {
  await showColumns(page, view);
  const label = requireViewContract(view).fieldMap[field]?.label ?? field;
  const checkbox = columns(page, view).getByRole("checkbox", {
    name: label,
    exact: true,
  });
  await checkbox.check();
  const earlier = columns(page, view).getByRole("button", {
    name: `Move ${label} earlier`,
    exact: true,
  });
  for (let left = 40; left > 0 && (await earlier.isEnabled()); left -= 1)
    await earlier.click();
  await expect(earlier).toBeDisabled();
  await columns(page, view)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await scrollGridTargetIntoView({
    page,
    surface: view,
    targetTestId: gridSortHeaderTestId(view, field),
  });
}
async function seed(page: Page, count = 2) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CSL"),
    "Column sizing evidence",
  );
  for (let i = 0; i < count; i += 1)
    await createViewRow(page, incident, surface, {
      client_txn_id: uniqueTxn("csl"),
      [summary]: `Committed summary ${i}`,
    });
  await page.goto(`/?incident_id=${incident}`);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await page.evaluate(() => document.fonts.ready.then(() => undefined));
  await firstColumn(page);
  return { incident, rows: await queryViewRows(page, incident, surface) };
}
async function dragBoundary(page: Page, delta: number) {
  const box = await header(page).boundingBox();
  if (!box) throw new Error("Visible header geometry is required");
  await page.mouse.move(box.x + box.width - 2, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(box.x + box.width - 2 + delta, box.y + box.height / 2, {
    steps: 4,
  });
  await page.mouse.up();
}

test("Workbook sizing preserves production geometry drafts and saved configuration across every entry point", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page);
  const defaultWidth = await width(header(page));
  const initialSort = await header(page).getAttribute("aria-sort");
  let mutations = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/(records|rows)(\/|$)/.test(new URL(request.url()).pathname)
    )
      mutations += 1;
  });
  const input = await widthPanel(page);
  for (const value of ["39", "4097", "40.5", "NaN"]) {
    await input.fill(value);
    await columns(page)
      .getByRole("button", { name: "Apply width", exact: true })
      .click();
    await expect(columns(page).getByRole("alert")).toHaveText(
      "Enter a whole number from 40 to 4096.",
    );
    await expect(input).toHaveValue(value);
    expect(await width(header(page))).toBe(defaultWidth);
  }
  await input.press("Escape");
  await expect(
    columns(page).getByRole("button", {
      name: `Width for ${requireViewContract(surface).fieldMap[summary]?.label}`,
      exact: true,
    }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(trigger(page)).toBeFocused();
  for (const value of [40, 4096, 240]) {
    await setWidth(page, String(value));
    await expect.poll(() => width(header(page))).toBe(value);
  }
  await header(page).focus();
  await page.keyboard.press("Control+ArrowRight");
  await expect.poll(() => width(header(page))).toBe(250);
  await page.keyboard.press("Meta+ArrowLeft");
  await expect.poll(() => width(header(page))).toBe(240);
  await dragBoundary(page, 37.5);
  await expect.poll(() => width(header(page))).toBe(278);
  await page.setViewportSize({ width: 2560, height: 1600 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await dragBoundary(page, 40);
  await expect.poll(() => width(header(page))).toBe(298);
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  await page.setViewportSize({ width: 1440, height: 900 });
  const record = f.rows[0]?.record_id;
  if (!record) throw new Error("Committed fixture row is required");
  await page.getByTestId(rowCellTestId(record, summary)).click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: record,
      fieldKey: summary,
      surface: "grid",
    }),
  );
  await editor.fill("  unfinished Ω " + "draft ".repeat(300));
  await editor.evaluate((element) =>
    (element as HTMLTextAreaElement).setSelectionRange(3, 7),
  );
  await dragBoundary(page, 20);
  await expect(editor).toBeFocused();
  expect(
    await editor.evaluate((element) => [
      (element as HTMLTextAreaElement).selectionStart,
      (element as HTMLTextAreaElement).selectionEnd,
    ]),
  ).toEqual([3, 7]);
  const fitted = await fitVisible(page);
  expect(fitted).toBeLessThan(1000);
  await expect(editor).toHaveValue("  unfinished Ω " + "draft ".repeat(300));
  expect(mutations).toBe(0);
  await editor.focus();
  await editor.press("Escape");
  await fitVisible(page);
  const fittedCommitted = await width(header(page));
  await setWidth(page, "600");
  const box = await header(page).boundingBox();
  if (!box) throw new Error("Header is missing");
  await page.mouse.dblclick(box.x + box.width - 2, box.y + box.height / 2);
  await expect.poll(() => width(header(page))).toBe(fittedCommitted);
  expect(await header(page).getAttribute("aria-sort")).toBe(initialSort);
  await setWidth(page, "4096");
  await setSavedViewDraftName(page, surface, "Sizing saved configuration");
  const createdResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/incidents/${f.incident}/saved-views`),
  );
  await createSavedViewFromCurrentSurface(page, surface);
  const saved = (await (await createdResponse).json()).data;
  expect(saved.layout_json.column_widths).toContainEqual({
    field_key: summary,
    width_px: 4096,
  });
  await setWidth(page, "40");
  await expect(
    page.getByTestId(savedViewModifiedTestId(surface)),
  ).toBeVisible();
  await setSavedViewDraftName(page, surface, "Duplicate selected sizing");
  const duplicateResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/incidents/${f.incident}/saved-views`),
  );
  await duplicateSavedViewFromCurrentSurface(
    page,
    surface,
    saved.saved_view_id,
  );
  expect(
    (await (await duplicateResponse).json()).data.layout_json.column_widths,
  ).toContainEqual({ field_key: summary, width_px: 4096 });
  await setWidth(page, "240");
  await widthPanel(page);
  await holdAnimationFrames(page);
  await columns(page)
    .getByRole("button", { name: "Fit visible content", exact: true })
    .click();
  await selectSavedView(page, surface, saved.saved_view_id);
  await expect(
    page.getByTestId(savedViewSelectorTestId(surface)),
  ).toHaveAttribute("title", saved.display_name);
  await expect.poll(() => width(header(page))).toBe(4096);
  await page.getByTestId(savedViewSelectorTestId(surface)).focus();
  await releaseAnimationFrames(page);
  await expect(
    page.getByTestId(savedViewSelectorTestId(surface)),
  ).toBeFocused();
  await expect.poll(() => width(header(page))).toBe(4096);
  await page.goto(
    `/?incident_id=${f.incident}&sheet_ref_kind=saved_view&sheet_ref_id=${saved.saved_view_id}`,
  );
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(savedViewSelectorTestId(surface)),
  ).toHaveAttribute("title", saved.display_name);
  await expect.poll(() => width(header(page))).toBe(4096);
  await widthPanel(page);
  await columns(page)
    .getByRole("button", { name: "Restore default", exact: true })
    .click();
  await columns(page)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await columns(page)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  expect(await width(header(page))).toBe(defaultWidth);
  await updateSavedViewFromCurrentSurface(page, surface, saved.saved_view_id);
  await expect
    .poll(
      async () =>
        (await readSavedView(page, f.incident, saved.saved_view_id)).layout_json
          .column_widths,
    )
    .toEqual([]);
  await setWidth(page, "40");
  await showColumns(page);
  await columns(page)
    .getByRole("button", { name: "Reset columns", exact: true })
    .click();
  await expect(
    columns(page).getByRole("checkbox").first(),
  ).toHaveAccessibleName("Date Entered");
  await expect(
    page.getByTestId(savedViewSelectorTestId(surface)),
  ).toHaveAttribute("title", saved.display_name);
  await columns(page)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await scrollGridTargetIntoView({
    page,
    surface,
    targetTestId: gridSortHeaderTestId(surface, summary),
  });
  await expect.poll(() => width(header(page))).toBe(defaultWidth);
  await expect(
    page.getByTestId(savedViewModifiedTestId(surface)),
  ).toBeVisible();
  await setWidth(page, "40");
  await openSavedViewActionMenu(page, surface);
  await page
    .getByTestId(savedViewResetButtonTestId(surface, saved.saved_view_id))
    .click();
  await scrollGridTargetIntoView({
    page,
    surface,
    targetTestId: gridSortHeaderTestId(surface, summary),
  });
  await expect.poll(() => width(header(page))).toBe(defaultWidth);
  await page.reload();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect.poll(() => width(header(page))).toBe(defaultWidth);
  expect(mutations).toBe(0);
  await widthPanel(page);
  await testInfo.attach("production-column-sizing-panel", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
});

test("Workbook sizing binds empty and hidden columns across all surface families and densities", async ({
  page,
  workerAdmin,
}, testInfo) => {
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CSL-SURFACES"),
    "Shared sizing boundary",
  );
  const surfaces = [
    surface,
    hostsViewSchemaId,
    assessmentsViewSchemaId,
    evidenceViewSchemaId,
  ];
  for (const view of surfaces) {
    const field = requireViewContract(view).fields.find(
      (entry) => !entry.defaultHidden,
    );
    if (!field) throw new Error("Expected default field");
    for (const density of ["compact", "default", "comfortable"] as const) {
      preferences.select(density);
      await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
      await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
      await page.evaluate(() => document.fonts.ready.then(() => undefined));
      await setWidth(page, "40", field.fieldKey, view);
      await expect
        .poll(() => width(header(page, field.fieldKey, view)))
        .toBe(40);
      await fitVisible(page, field.fieldKey, view);
      await expect(
        page.getByRole("status").filter({ hasText: /using the header/ }),
      ).toBeVisible();
      await showColumns(page, view);
      for (const checkbox of await columns(page, view)
        .getByRole("checkbox")
        .all())
        await checkbox.uncheck();
      const input = await widthPanel(page, field.fieldKey, view);
      await expect(
        columns(page, view).getByRole("button", {
          name: "Fit visible content",
          exact: true,
        }),
      ).toBeDisabled();
      await input.fill("4096");
      await columns(page, view)
        .getByRole("button", { name: "Apply width", exact: true })
        .click();
      await columns(page, view)
        .getByRole("button", { name: "Cancel", exact: true })
        .click();
      await columns(page, view)
        .getByRole("checkbox", { name: field.label, exact: true })
        .check();
      await columns(page, view)
        .getByRole("button", { name: "Close columns", exact: true })
        .click();
      await expect
        .poll(() => width(header(page, field.fieldKey, view)))
        .toBe(4096);
    }
  }
  await setWidth(page, "200", "evidence.title", evidenceViewSchemaId);
  await widthPanel(page, "evidence.title", evidenceViewSchemaId);
  await holdAnimationFrames(page);
  await columns(page, evidenceViewSchemaId)
    .getByRole("button", { name: "Fit visible content", exact: true })
    .click();
  const beforeClose = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("csl-close"),
        base_incident_version: beforeClose.incident_version,
        reason: "Verify readable local layout",
      })
    ).ok,
  ).toBe(true);
  await expect(
    page.getByTestId(workbookAddRowButtonTestId(evidenceViewSchemaId)),
  ).toBeDisabled();
  await releaseAnimationFrames(page);
  await expect
    .poll(() => width(header(page, "evidence.title", evidenceViewSchemaId)))
    .toBe(200);
  await columns(page, evidenceViewSchemaId)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await columns(page, evidenceViewSchemaId)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  await setWidth(page, "40", "evidence.title", evidenceViewSchemaId);
  await expect
    .poll(() => width(header(page, "evidence.title", evidenceViewSchemaId)))
    .toBe(40);
  await fitVisible(page, "evidence.title", evidenceViewSchemaId);
  await page.setViewportSize({ width: 2560, height: 1600 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await page.addStyleTag({
    content:
      "* { letter-spacing: 0.12em !important; word-spacing: 0.16em !important; line-height: 1.5 !important; }",
  });
  await showColumns(page, evidenceViewSchemaId);
  await expect(columns(page, evidenceViewSchemaId)).toBeInViewport();
  await testInfo.attach("production-sizing-columns-zoom-spacing", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  expect(await page.locator(gridScrollportSelector()).count()).toBeGreaterThan(
    0,
  );
});

test("Workbook fitting measures only the visible committed viewport and rejects cancelled production measurements", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page, 60);
  const distant = f.rows.at(-1);
  if (!distant) throw new Error("Expected off-screen record");
  await patchRecord(page, distant.record_id, {
    view_schema_id: surface,
    client_txn_id: uniqueTxn("csl-long"),
    base_row_version: distant.row_version,
    changes: [{ field_key: summary, value: "X".repeat(2000) }],
  });
  await page.reload();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await firstColumn(page);
  let queries = 0;
  page.on("request", (request) => {
    if (request.url().endsWith("/query")) queries += 1;
  });
  const small = await fitVisible(page);
  expect(small).toBeLessThan(1000);
  expect(queries).toBe(0);
  await expect(
    page.getByTestId(rowCellTestId(distant.record_id, summary)),
  ).toHaveCount(0);
  const mounted = await page
    .getByTestId(gridShellTestId(surface))
    .getByRole("row")
    .count();
  expect(mounted).toBeLessThan(60);
  await scrollGridCellIntoView({
    page,
    surface,
    recordId: distant.record_id,
    cellKey: summary,
  });
  expect(await width(header(page))).toBe(small);
  expect(await fitVisible(page)).toBe(4096);
  await expect(
    page.getByRole("status").filter({ hasText: /Maximum width reached/ }),
  ).toBeVisible();
  expect(queries).toBe(0);
  await setWidth(page, "300");
  await widthPanel(page);
  await holdAnimationFrames(page);
  await columns(page)
    .getByRole("button", { name: "Fit visible content", exact: true })
    .click();
  await expect(
    columns(page).getByText("Measuring visible content…"),
  ).toBeVisible();
  await columns(page)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await releaseAnimationFrames(page);
  await expect.poll(() => width(header(page))).toBe(300);
  await columns(page)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  const input = await widthPanel(page);
  await holdAnimationFrames(page);
  await columns(page)
    .getByRole("button", { name: "Fit visible content", exact: true })
    .click();
  await input.fill("420");
  await columns(page)
    .getByRole("button", { name: "Apply width", exact: true })
    .click();
  await releaseAnimationFrames(page);
  await expect.poll(() => width(header(page))).toBe(420);
  await columns(page)
    .getByRole("button", { name: "Cancel", exact: true })
    .click();
  await columns(page)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  const collection = await createViewRow(page, f.incident, surface, {
    client_txn_id: uniqueTxn("csl-collection"),
    [summary]: "Collection sizing",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "Visible tag Ω 東京" },
        { op: "add_tag", tag_name: "A longer hidden collection member" },
        { op: "add_tag", tag_name: "Third member" },
      ],
    },
  });
  await page.reload();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await firstColumn(page, "timeline.tags");
  await scrollGridTargetIntoView({
    page,
    surface,
    targetTestId: relationshipItemsTestId(
      collection.record_id,
      "timeline.tags",
      "grid",
    ),
  });
  const overflow = page.getByTestId(
    relationshipOverflowButtonTestId(collection.record_id, "timeline.tags"),
  );
  await expect(overflow).toBeVisible();
  const collectionCell = page.getByTestId(
    relationshipItemsTestId(collection.record_id, "timeline.tags", "grid"),
  );
  const summaryBefore = await collectionCell.innerText();
  const fittedCollection = await fitVisible(page, "timeline.tags");
  expect(fittedCollection).toBeGreaterThan(40);
  expect(fittedCollection).toBeLessThan(1000);
  await expect(overflow).toBeVisible();
  expect(await collectionCell.innerText()).toBe(summaryBefore);
  await expect(page.getByRole("dialog", { name: /Tags/ })).toHaveCount(0);
  expect(
    (await queryViewRows(page, f.incident, surface)).find(
      (row) => row.record_id === collection.record_id,
    )?.row_version,
  ).toBe(collection.row_version);
  await firstColumn(page);
  await changeGrouping(page, surface, "timeline.capture_state");
  const groups = new Set(
    (await queryViewRows(page, f.incident, surface)).map((row) =>
      String(row.cells["timeline.capture_state"]?.value),
    ),
  );
  for (const group of groups)
    await collapseGridGroup({
      page,
      surface,
      groupTestId: gridGroupRowTestId(surface, "timeline.capture_state", group),
    });
  expect(await fitVisible(page)).toBeLessThan(small);
  await expect(
    page.getByRole("status").filter({ hasText: /using the header/ }),
  ).toBeVisible();
  await widthPanel(page, "timeline.data_source_text");
  await expect(
    columns(page).getByRole("button", {
      name: "Fit visible content",
      exact: true,
    }),
  ).toBeDisabled();
  await expect(
    columns(page).getByText(/Scroll this column into view|Show this column/),
  ).toBeVisible();
});

async function holdAnimationFrames(page: Page) {
  await page.evaluate(() => {
    const request = window.requestAnimationFrame;
    const frames: FrameRequestCallback[] = [];
    window.requestAnimationFrame = (callback) => {
      frames.push(callback);
      return frames.length;
    };
    Object.assign(window, {
      releaseSizingFrames: () => {
        window.requestAnimationFrame = request;
        for (const frame of frames) frame(performance.now());
      },
    });
  });
}
async function releaseAnimationFrames(page: Page) {
  await page.evaluate(() => {
    (
      window as unknown as { releaseSizingFrames: () => void }
    ).releaseSizingFrames();
  });
}

test("a11y.column-sizing native controls retain keyboard focus at narrow width zoom and text spacing", async ({
  page,
}, info) => {
  const f = await seed(page);
  for (const [viewportWidth, zoom] of [
    [1024, 1],
    [1536, 2],
  ] as const) {
    await page.setViewportSize({ width: viewportWidth, height: 1200 });
    await page.evaluate((value) => {
      document.documentElement.style.zoom = String(value);
    }, zoom);
    const spacing = await page.addStyleTag({
      content:
        "* { letter-spacing: 0.12em !important; word-spacing: 0.16em !important; line-height: 1.5 !important; }",
    });
    await trigger(page).focus();
    await page.keyboard.press("Enter");
    const checkbox = columns(page).getByRole("checkbox").first();
    await expect(checkbox).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(
      columns(page).getByRole("button", {
        name: "Move Activity Synopsis later",
        exact: true,
      }),
    ).toBeFocused();
    await page.keyboard.press("Tab");
    const widthButton = columns(page).getByRole("button", {
      name: "Width for Activity Synopsis",
      exact: true,
    });
    await expect(widthButton).toBeFocused();
    await page.keyboard.press("Enter");
    const input = columns(page).getByRole("textbox", {
      name: "Width in CSS pixels",
    });
    await expect(input).toBeFocused();
    await input.fill("39");
    await input.press("Enter");
    await expect(input).toHaveAttribute("aria-invalid", "true");
    await expect(input).toBeInViewport({ ratio: 1 });
    await input.fill("248");
    await page.keyboard.press("Tab");
    await expect(
      columns(page).getByRole("button", { name: "Apply width", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Enter");
    await expect.poll(() => width(header(page))).toBe(248);
    await page.keyboard.press("Tab");
    await expect(
      columns(page).getByRole("button", {
        name: "Fit visible content",
        exact: true,
      }),
    ).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(
      columns(page).getByRole("button", {
        name: "Restore default",
        exact: true,
      }),
    ).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(
      columns(page).getByRole("button", { name: "Cancel", exact: true }),
    ).toBeFocused();
    await info.attach(`column-sizing-accessibility-${viewportWidth}-${zoom}`, {
      body: await columns(page).ariaSnapshot(),
      contentType: "text/plain",
    });
    await info.attach(
      `column-sizing-accessibility-image-${viewportWidth}-${zoom}`,
      { body: await page.screenshot(), contentType: "image/png" },
    );
    await page.keyboard.press("Escape");
    await expect(widthButton).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(trigger(page)).toBeFocused();
    await spacing.evaluate((element) =>
      element.parentNode?.removeChild(element),
    );
  }
  expect(
    (await queryViewRows(page, f.incident, surface)).every(
      (row) => row.row_version === 1,
    ),
  ).toBe(true);
});
