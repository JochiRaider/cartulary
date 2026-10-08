import {
  changeGrouping,
  collapseGridGroup,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridRowTestId,
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
import { openTimelineInspector } from "./support/workbook/rowMutations";
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
  if (!(await columns(page, view).isVisible())) {
    if (!(await trigger(page, view).isVisible()))
      await page
        .getByRole("button", { name: "View options controls", exact: true })
        .click();
    await trigger(page, view).click();
  }
}
async function closeColumns(page: Page, view: string = surface) {
  await columns(page, view)
    .getByRole("button", { name: "Close columns", exact: true })
    .click();
  const query = page.getByRole("button", {
    name: "Query controls",
    exact: true,
  });
  if (
    (await query.isVisible()) &&
    (await query.getAttribute("aria-expanded")) === "true"
  )
    await query.click();
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
  await closeColumns(page, view);
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
  await closeColumns(page, view);
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
  await closeColumns(page, view);
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
async function tabTo(page: Page, control: Locator, limit = 500) {
  for (let step = 0; step < limit; step += 1) {
    if (await control.evaluate((node) => document.activeElement === node))
      return;
    await page.keyboard.press("Tab");
  }
  await expect(control).toBeFocused();
}

// Observation only: never reveal or focus the destination under assertion.
async function columnsFocusGeometry(target: Locator) {
  return target.evaluate((element) => {
    const node = element as HTMLElement;
    const panel = node.closest<HTMLElement>('[role="dialog"]');
    if (!panel) throw new Error("Missing Columns panel");
    const box = node.getBoundingClientRect();
    const outer = panel.getBoundingClientRect();
    const sx = outer.width / panel.offsetWidth;
    const sy = outer.height / panel.offsetHeight;
    const style = getComputedStyle(node);
    const ring = Math.max(
      0,
      (Number.parseFloat(style.outlineWidth) || 0) +
        (Number.parseFloat(style.outlineOffset) || 0),
    );
    const bounds = {
      top: box.top - ring * sy,
      bottom: box.bottom + ring * sy,
      left: box.left - ring * sx,
      right: box.right + ring * sx,
    };
    const clip = {
      top: Math.max(0, outer.top + panel.clientTop * sy),
      bottom: Math.min(
        innerHeight,
        outer.top + (panel.clientTop + panel.clientHeight) * sy,
      ),
      left: Math.max(0, outer.left + panel.clientLeft * sx),
      right: Math.min(
        innerWidth,
        outer.left + (panel.clientLeft + panel.clientWidth) * sx,
      ),
    };
    return {
      focused: document.activeElement === node,
      contained:
        bounds.top >= clip.top - 1 &&
        bounds.bottom <= clip.bottom + 1 &&
        bounds.left >= clip.left - 1 &&
        bounds.right <= clip.right + 1,
      bounds,
      clip,
      scrollTop: panel.scrollTop,
      scrollLeft: panel.scrollLeft,
    };
  });
}

async function expectColumnsFocus(target: Locator) {
  try {
    await expect
      .poll(() => columnsFocusGeometry(target))
      .toMatchObject({ focused: true, contained: true });
  } catch (error) {
    throw new Error(
      `Columns focus geometry: ${JSON.stringify(await columnsFocusGeometry(target))}`,
      { cause: error },
    );
  }
}

async function outsideColumnsScroll(page: Page) {
  return page.evaluate(
    (selector) => ({
      page: [scrollX, scrollY],
      regions: [
        ...document.querySelectorAll<HTMLElement>(
          `${selector}, [data-inspector-scroll-body]`,
        ),
      ].map((node) => [node.scrollLeft, node.scrollTop]),
    }),
    gridScrollportSelector(),
  );
}

test("Columns reveals semantic Width returns and relocated actions inside its clipped panel", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page);
  const record = f.rows[0]?.record_id;
  if (!record) throw new Error("Missing Timeline row");
  await openTimelineInspector(page, record);
  await page.locator("[data-inspector-scroll-body]").evaluate((node) => {
    node.scrollTop = node.scrollHeight;
  });
  await page.getByTestId(rowCellTestId(record, summary)).click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: record,
      fieldKey: summary,
      surface: "grid",
    }),
  );
  const raw = "  Retained Columns authoring Ω  ";
  await editor.fill(raw);
  await editor.evaluate((node) =>
    (node as HTMLTextAreaElement).setSelectionRange(3, 9),
  );
  let recordWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/(records|rows)(\/|$)/.test(new URL(request.url()).pathname)
    )
      recordWrites += 1;
  });
  const label = (field: string) =>
    requireViewContract(surface).fieldMap[field]?.label;
  const late = "timeline.has_unresolved_mentions";
  const early = "timeline.date_entered_text";
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
  ]) {
    await page.setViewportSize(viewport);
    for (const dismiss of ["Escape", "Cancel"]) {
      await showColumns(page);
      await columns(page).evaluate((node) => {
        node.scrollTop = node.scrollHeight;
      });
      const before = await columns(page).evaluate((node) => node.scrollTop);
      expect(before).toBeGreaterThan(0);
      const otherScroll = await outsideColumnsScroll(page);
      const action = columns(page).getByRole("button", {
        name: `Width for ${label(late)}`,
        exact: true,
      });
      await action.click();
      if (dismiss === "Cancel") {
        await tabTo(
          page,
          columns(page).getByRole("button", { name: "Cancel", exact: true }),
        );
        await page.keyboard.press("Enter");
      } else await page.keyboard.press("Escape");
      await expectColumnsFocus(action);
      expect(
        await columns(page).evaluate((node) => node.scrollTop),
      ).toBeGreaterThan(0);
      expect(await outsideColumnsScroll(page)).toEqual(otherScroll);
      await expect(editor).toHaveValue(raw);
      expect(
        await editor.evaluate((node) => [
          (node as HTMLTextAreaElement).selectionStart,
          (node as HTMLTextAreaElement).selectionEnd,
        ]),
      ).toEqual([3, 9]);
      await info.attach(`columns-return-${viewport.width}-${dismiss}`, {
        body: JSON.stringify(await columnsFocusGeometry(action)),
        contentType: "application/json",
      });
      await page.keyboard.press("Escape");
      await expect(trigger(page)).toBeFocused();
    }
    await showColumns(page);
    const freeze = columns(page).getByRole("button", {
      name: `Freeze through ${label(early)}`,
      exact: true,
    });
    await freeze.focus();
    await page.keyboard.press("Enter");
    await expect(freeze).toBeDisabled();
    await expectColumnsFocus(
      columns(page).getByRole("button", {
        name: `Width for ${label(early)}`,
        exact: true,
      }),
    );
    await columns(page).evaluate((node) => {
      node.scrollTop = node.scrollHeight;
    });
    await columns(page)
      .getByRole("button", { name: "Unfreeze columns", exact: true })
      .click();
    await expectColumnsFocus(freeze);
    // Activation uses the current keyboard destination; locator actions would
    // conceal a missing reveal between moves by scrolling before the next click.
    const earlier = columns(page).getByRole("button", {
      name: `Move ${label(late)} earlier`,
      exact: true,
    });
    const later = columns(page).getByRole("button", {
      name: `Move ${label(late)} later`,
      exact: true,
    });
    await earlier.focus();
    for (const [invoker, fallback] of [
      [earlier, later],
      [later, earlier],
    ] as const) {
      while (await invoker.isEnabled()) {
        await page.keyboard.press("Enter");
        await expectColumnsFocus(
          (await invoker.isEnabled()) ? invoker : fallback,
        );
      }
    }
    await info.attach(`columns-move-${viewport.width}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    await page.keyboard.press("Escape");
  }
  expect(recordWrites).toBe(0);
  expect(
    (await queryViewRows(page, f.incident, surface)).every(
      (row) => row.row_version === 1,
    ),
  ).toBe(true);
});

test("Columns reveals Width return on the shared Hosts surface", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 1024, height: 720 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CSL-FOCUS-HOSTS"),
    "Shared Columns focus",
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${hostsViewSchemaId}`,
  );
  const field = requireViewContract(hostsViewSchemaId).fields.at(-1);
  if (!field) throw new Error("Expected Hosts field");
  await showColumns(page, hostsViewSchemaId);
  await columns(page, hostsViewSchemaId).evaluate((node) => {
    node.scrollTop = node.scrollHeight;
  });
  expect(
    await columns(page, hostsViewSchemaId).evaluate((node) => node.scrollTop),
  ).toBeGreaterThan(0);
  const action = columns(page, hostsViewSchemaId).getByRole("button", {
    name: `Width for ${field.label}`,
    exact: true,
  });
  await action.click();
  await page.keyboard.press("Escape");
  await expectColumnsFocus(action);
  await info.attach("columns-hosts-return", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await page.keyboard.press("Escape");
  await expect(trigger(page, hostsViewSchemaId)).toBeFocused();
});
async function setFreeze(
  page: Page,
  field: string | null,
  view: string = surface,
) {
  await showColumns(page, view);
  await columns(page, view)
    .getByRole("button", {
      name:
        field === null
          ? "Unfreeze columns"
          : `Freeze through ${requireViewContract(view).fieldMap[field]?.label}`,
      exact: true,
    })
    .click();
  await closeColumns(page, view);
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

test("Columns keeps one production keyboard session through Fit movement and freezing", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await seed(page);
  await page.reload();
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await page.evaluate(() => document.fonts.ready.then(() => undefined));
  await firstColumn(page);
  let recordWrites = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/(records|rows)(\/|$)/.test(new URL(request.url()).pathname)
    )
      recordWrites += 1;
  });
  await tabTo(page, trigger(page));
  await page.keyboard.press("Enter");
  await expect(columns(page).getByRole("checkbox").first()).toBeFocused();
  const firstLabel =
    requireViewContract(surface).fieldMap[summary]?.label ?? "";
  const lastLabel =
    (await columns(page)
      .getByRole("checkbox")
      .last()
      .evaluate((node) => node.parentElement?.textContent?.trim())) ?? "";
  const firstEarlier = columns(page).getByRole("button", {
    name: `Move ${firstLabel} earlier`,
  });
  const firstLater = columns(page).getByRole("button", {
    name: `Move ${firstLabel} later`,
  });
  await tabTo(page, firstLater);
  await page.keyboard.press("Enter");
  await expect(firstLater).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(firstEarlier).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(firstEarlier).toBeDisabled();
  await expect(firstLater).toBeFocused();
  await expect(firstLater).toHaveCSS("outline-style", "solid");
  const lastEarlier = columns(page).getByRole("button", {
    name: `Move ${lastLabel} earlier`,
  });
  const lastLater = columns(page).getByRole("button", {
    name: `Move ${lastLabel} later`,
  });
  await tabTo(page, lastEarlier);
  await page.keyboard.press("Enter");
  await expect(lastEarlier).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(lastLater).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(lastLater).toBeDisabled();
  await expect(lastEarlier).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(trigger(page)).toBeFocused();
  await page.keyboard.press("Enter");
  const freeze = columns(page).getByRole("button", {
    name: `Freeze through ${firstLabel}`,
  });
  const widthAction = columns(page).getByRole("button", {
    name: `Width for ${firstLabel}`,
  });
  await tabTo(page, freeze);
  await page.keyboard.press("Enter");
  await expect(freeze).toBeDisabled();
  await expect(widthAction).toBeFocused();
  const unfreeze = columns(page).getByRole("button", {
    name: "Unfreeze columns",
  });
  await tabTo(page, unfreeze);
  await page.keyboard.press("Enter");
  await expect(unfreeze).toBeDisabled();
  await expect(freeze).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(widthAction).toBeFocused();
  await page.keyboard.press("Enter");
  const input = columns(page).getByRole("textbox", {
    name: "Width in CSS pixels",
  });
  await expect(input).toBeFocused();
  const fit = columns(page).getByRole("button", {
    name: "Fit visible content",
    exact: true,
  });
  await tabTo(page, fit);
  await holdAnimationFrames(page);
  await page.keyboard.press("Enter");
  await expect(
    columns(page).getByText("Measuring visible content…"),
  ).toBeVisible();
  await expect(fit).toBeFocused();
  await expect(fit).toHaveAttribute("aria-busy", "true");
  await expect(fit).toHaveCSS("outline-style", "solid");
  await page.keyboard.press("Enter");
  await expect(fit).toBeFocused();
  await page.keyboard.press("Tab");
  const restore = columns(page).getByRole("button", {
    name: "Restore default",
  });
  await expect(restore).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(fit).toBeFocused();
  await releaseAnimationFrames(page);
  await expect(
    page.getByRole("status").filter({ hasText: /fitted to/ }),
  ).toBeVisible();
  await expect(fit).toBeFocused();
  await expect(fit).not.toHaveAttribute("aria-busy", "true");
  await page.keyboard.press("Tab");
  await expect(restore).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(fit).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(widthAction).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(trigger(page)).toBeFocused();
  expect(recordWrites).toBe(0);
});

test("Workbook sizing preserves production geometry drafts and saved configuration across every entry point", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page);
  await setFreeze(page, summary);
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
  await input.fill("700");
  await input.press("Enter");
  await expect.poll(() => width(header(page))).toBe(700);
  await input.press("Tab");
  await expect(
    columns(page).getByRole("button", { name: "Apply width", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  const fitButton = columns(page).getByRole("button", {
    name: "Fit visible content",
    exact: true,
  });
  await expect(fitButton).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: /fitted to/ }),
  ).toBeVisible();
  await expect.poll(() => width(header(page))).not.toBe(700);
  const samePanelFitWidth = await width(header(page));
  expect(samePanelFitWidth).toBeLessThan(1000);
  await expect(input).toHaveValue(String(samePanelFitWidth));
  await expect(fitButton).toBeEnabled();
  await expect(fitButton).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(
    columns(page).getByRole("button", { name: "Apply width", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(input).toBeFocused();
  await input.press("Enter");
  await expect.poll(() => width(header(page))).toBe(samePanelFitWidth);
  await expect(input).toHaveValue(String(samePanelFitWidth));
  await expect(input).toBeFocused();
  await expect(
    page.getByRole("status").filter({ hasText: /width set to/ }),
  ).toBeVisible();
  const refinedWidth = samePanelFitWidth + 37;
  await input.fill(String(refinedWidth));
  await input.press("Enter");
  await expect.poll(() => width(header(page))).toBe(refinedWidth);
  await expect(input).toHaveValue(String(refinedWidth));
  await expect(input).toBeFocused();
  await input.press("Tab");
  await page.keyboard.press("Tab");
  await page.keyboard.press("Tab");
  const restoreButton = columns(page).getByRole("button", {
    name: "Restore default",
    exact: true,
  });
  await expect(restoreButton).toBeFocused();
  await page.keyboard.press("Enter");
  await expect.poll(() => width(header(page))).toBe(defaultWidth);
  await expect(input).toHaveValue(String(defaultWidth));
  await expect(
    page.getByRole("status").filter({ hasText: /default width restored/ }),
  ).toBeVisible();
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
  expect(saved.layout_json.frozen_through_field_key).toBe(summary);
  expect(saved.layout_json.column_widths).toContainEqual({
    field_key: summary,
    width_px: 4096,
  });
  await setWidth(page, "40");
  await setFreeze(page, null);
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
  const duplicatedLayout = (await (await duplicateResponse).json()).data
    .layout_json;
  expect(duplicatedLayout.column_widths).toContainEqual({
    field_key: summary,
    width_px: 4096,
  });
  expect(duplicatedLayout.frozen_through_field_key).toBe(summary);
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
  await closeColumns(page);
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
  await closeColumns(page);
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
      await setFreeze(page, field.fieldKey, view);
      await expect(page.locator(gridScrollportSelector())).toHaveAttribute(
        "data-grid-freeze-state",
        "active",
      );
      await showColumns(page, view);
      for (const checkbox of await columns(page, view)
        .getByRole("checkbox")
        .all())
        await checkbox.uncheck();
      await expect(columns(page, view)).toContainText("0 visible data columns");
      await expect(columns(page, view)).toContainText(
        `${field.label} (hidden)`,
      );
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
      await closeColumns(page, view);
      await expect
        .poll(() => width(header(page, field.fieldKey, view)))
        .toBe(4096);
      await expect(page.locator(gridScrollportSelector())).toHaveAttribute(
        "data-grid-freeze-state",
        "suspended",
      );
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
  await closeColumns(page, evidenceViewSchemaId);
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
  await closeColumns(page);
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
  await closeColumns(page);
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
    if (!(await trigger(page).isVisible())) {
      await tabTo(
        page,
        page.getByRole("button", {
          name: "View options controls",
          exact: true,
        }),
      );
      await page.keyboard.press("Enter");
    }
    await tabTo(page, trigger(page));
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
    const fit = columns(page).getByRole("button", {
      name: "Fit visible content",
      exact: true,
    });
    await expect(fit).toBeFocused();
    await holdAnimationFrames(page);
    await page.keyboard.press("Enter");
    await expect(fit).toBeFocused();
    await expect(fit).toHaveAttribute("aria-busy", "true");
    await expect(fit).toBeInViewport({ ratio: 1 });
    await expect(fit).toHaveCSS("outline-style", "solid");
    await page.keyboard.press("Tab");
    const restore = columns(page).getByRole("button", {
      name: "Restore default",
      exact: true,
    });
    await expect(restore).toBeFocused();
    await releaseAnimationFrames(page);
    await expect(
      page.getByRole("status").filter({ hasText: /fitted to/ }),
    ).toBeVisible();
    await expect(restore).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(fit).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(restore).toBeFocused();
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
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Shift+Tab");
    await expect(fit).toBeFocused();
    await holdAnimationFrames(page);
    await page.keyboard.press("Enter");
    // Keep the current request and focus while its production font capability
    // becomes unavailable. No grid or Columns geometry is repaired here.
    await page.evaluate(() =>
      Object.defineProperty(document.fonts, "status", {
        configurable: true,
        value: "loading",
      }),
    );
    try {
      await releaseAnimationFrames(page);
      await expect(fit).toBeDisabled();
      await expectColumnsFocus(restore);
      await info.attach(`columns-fit-fallback-${viewportWidth}-${zoom}`, {
        body: JSON.stringify(await columnsFocusGeometry(restore)),
        contentType: "application/json",
      });
    } finally {
      await page.evaluate(() =>
        Reflect.deleteProperty(document.fonts, "status"),
      );
    }
    await page.keyboard.press("Escape");
    await expect(widthButton).toBeFocused();
    await expectColumnsFocus(widthButton);
    const lateLabel =
      requireViewContract(surface).fieldMap["timeline.has_unresolved_mentions"]
        ?.label;
    const lateWidth = columns(page).getByRole("button", {
      name: `Width for ${lateLabel}`,
      exact: true,
    });
    for (const dismiss of ["Escape", "Cancel"]) {
      await columns(page).evaluate((node) => {
        node.scrollTop = node.scrollHeight;
      });
      expect(
        await columns(page).evaluate((node) => node.scrollTop),
      ).toBeGreaterThan(0);
      await lateWidth.click();
      if (dismiss === "Cancel") {
        await tabTo(
          page,
          columns(page).getByRole("button", { name: "Cancel", exact: true }),
        );
        await page.keyboard.press("Enter");
      } else await page.keyboard.press("Escape");
      await expectColumnsFocus(lateWidth);
    }
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

test("Workbook frozen gutter occludes crossing cell paint and owns its pointer target", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 1024, height: 720 });
  const fixture = await seed(page);
  const recordId = fixture.rows[0]?.record_id;
  if (!recordId) throw new Error("Missing committed Timeline row");
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/(records|rows)(\/|$)/.test(new URL(request.url()).pathname)
    )
      writes += 1;
  });

  const grid = page.locator(gridScrollportSelector());
  const scrollingCell = page.getByTestId(rowCellTestId(recordId, summary));
  const gutter = page.getByTestId(gridRowGutterTestId(surface, recordId));
  await scrollingCell.evaluate(
    (content, gutterSelector) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(gutterSelector);
      if (!root || !cell || !cover)
        throw new Error("Expected mounted source, gutter, and grid");
      const gutterBounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        cell.getBoundingClientRect().left -
        (gutterBounds.left + gutterBounds.width / 4);
    },
    dataTestIdSelector(gridRowGutterTestId(surface, recordId)),
  );
  await expect
    .poll(() => grid.evaluate((node) => node.scrollLeft))
    .toBeGreaterThan(0);
  const sourceBounds = await scrollingCell.boundingBox();
  const gutterBounds = await gutter.boundingBox();
  if (!sourceBounds || !gutterBounds)
    throw new Error("Expected visible crossing geometry");
  expect(sourceBounds.x).toBeLessThan(gutterBounds.x + gutterBounds.width);
  expect(sourceBounds.x + sourceBounds.width).toBeGreaterThan(gutterBounds.x);

  const painted = await gutter.screenshot();
  await scrollingCell.evaluate((content) => {
    const cell = content.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Expected mounted source cell");
    cell.style.visibility = "hidden";
  });
  let occluded: Buffer;
  try {
    occluded = await gutter.screenshot();
  } finally {
    await scrollingCell.evaluate((content) => {
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      if (cell) cell.style.visibility = "";
    });
  }
  await info.attach("frozen-gutter-crossing-before", {
    body: painted,
    contentType: "image/png",
  });
  await info.attach("frozen-gutter-crossing-source-hidden", {
    body: occluded,
    contentType: "image/png",
  });
  const ordinaryHit = await gutter.evaluate((node) => {
    const bounds = node.getBoundingClientRect();
    const target = document.elementFromPoint(
      bounds.left + bounds.width / 2,
      bounds.top + bounds.height / 2,
    );
    return (
      target?.closest('[role="gridcell"]') === node.closest('[role="gridcell"]')
    );
  });
  const ordinaryStyles = await grid.evaluate((root) => {
    const firstRow = root.querySelector(
      '[data-cartulary-grid-row-kind="data"]',
    );
    const cells = firstRow?.querySelectorAll<HTMLElement>('[role="gridcell"]');
    const header = root.querySelector<HTMLElement>('[role="columnheader"]');
    const draft = root.querySelector<HTMLElement>(
      '[data-cartulary-grid-draft-row="true"] [role="gridcell"]',
    );
    const style = (node: HTMLElement | null | undefined) => {
      if (!node) return null;
      const value = getComputedStyle(node);
      return {
        backgroundColor: value.backgroundColor,
        backgroundImage: value.backgroundImage,
        overflow: value.overflow,
        position: value.position,
        zIndex: value.zIndex,
      };
    };
    return {
      row: style(firstRow as HTMLElement | null),
      selection: style(cells?.[0]),
      gutter: style(cells?.[1]),
      scrolling: style(cells?.[2]),
      header: style(header),
      draft: style(draft),
    };
  });
  await scrollingCell.click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      recordId,
      fieldKey: summary,
      surface: "grid",
    }),
  );
  await expect(editor).toBeVisible();
  const editorStyles = await editor.evaluate((node) => {
    const cell = node.closest<HTMLElement>('[role="gridcell"]');
    const value = cell && getComputedStyle(cell);
    return {
      backgroundColor: value?.backgroundColor,
      overflow: value?.overflow,
      zIndex: value?.zIndex,
    };
  });
  await editor.fill("Retained scrolling draft");
  await editor.press("Home");
  for (let i = 0; i < 4; i += 1) await editor.press("ArrowRight");
  await editor.evaluate(
    (node, gutterSelector) => {
      const root = node.closest<HTMLElement>('[role="grid"]');
      const cell = node.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(gutterSelector);
      if (!root || !cell || !cover)
        throw new Error("Expected mounted scrolling editor and gutter");
      const bounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        cell.getBoundingClientRect().left - (bounds.left + bounds.width / 4);
    },
    dataTestIdSelector(gridRowGutterTestId(surface, recordId)),
  );
  const editingPaint = await gutter.screenshot();
  const editingHit = await gutter.evaluate((node) => {
    const bounds = node.getBoundingClientRect();
    return (
      document
        .elementFromPoint(
          bounds.left + bounds.width / 2,
          bounds.top + bounds.height / 2,
        )
        ?.closest('[role="gridcell"]') === node.closest('[role="gridcell"]')
    );
  });
  await expect(editor).toBeFocused();
  await expect(editor).toHaveValue("Retained scrolling draft");
  expect(
    await editor.evaluate(
      (node) => (node as HTMLInputElement | HTMLTextAreaElement).selectionStart,
    ),
  ).toBe(4);
  expect(editingPaint.equals(painted)).toBe(true);
  expect(editingHit).toBe(true);
  await info.attach("frozen-gutter-crossing-editor-active", {
    body: editingPaint,
    contentType: "image/png",
  });
  await page.keyboard.press("Escape");
  await scrollingCell.evaluate(
    (content, gutterSelector) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(gutterSelector);
      if (!root || !cell || !cover)
        throw new Error("Expected mounted active source and gutter");
      const gutterBounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        cell.getBoundingClientRect().left -
        (gutterBounds.left + gutterBounds.width / 4);
    },
    dataTestIdSelector(gridRowGutterTestId(surface, recordId)),
  );
  const activeStyles = await scrollingCell.evaluate((content) => {
    const cell = content.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Expected mounted active source cell");
    const value = getComputedStyle(cell);
    return {
      className: cell.className,
      backgroundColor: value.backgroundColor,
      zIndex: value.zIndex,
    };
  });
  const hit = await gutter.evaluate((node) => {
    const bounds = node.getBoundingClientRect();
    const target = document.elementFromPoint(
      bounds.left + bounds.width / 2,
      bounds.top + bounds.height / 2,
    );
    return (
      target?.closest('[role="gridcell"]') === node.closest('[role="gridcell"]')
    );
  });
  await info.attach("frozen-gutter-render-characterization", {
    body: JSON.stringify({
      ordinaryHit,
      ordinaryStyles,
      editorStyles,
      activeStyles,
      activeHit: hit,
    }),
    contentType: "application/json",
  });
  expect(painted.equals(occluded)).toBe(true);
  expect(ordinaryStyles.gutter?.backgroundColor).not.toBe("rgba(0, 0, 0, 0)");
  expect(ordinaryStyles.selection?.backgroundColor).not.toBe(
    "rgba(0, 0, 0, 0)",
  );
  expect(ordinaryStyles.draft?.backgroundImage).not.toBe("none");
  expect(ordinaryHit).toBe(true);
  expect(hit).toBe(true);
  const selection = page
    .getByTestId(gridRowTestId(timelineViewSchemaId, recordId))
    .getByRole("checkbox");
  await scrollingCell.evaluate(
    (content, checkbox) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      if (!root || !cell || !checkbox)
        throw new Error("Expected scrolling cell and frozen selection control");
      const box = checkbox.getBoundingClientRect();
      root.scrollLeft +=
        cell.getBoundingClientRect().left - (box.left + box.width / 2);
    },
    await selection.elementHandle(),
  );
  const selectionHit = await selection.evaluate((node) => {
    const box = node.getBoundingClientRect();
    return (
      document.elementFromPoint(
        box.left + box.width / 2,
        box.top + box.height / 2,
      ) === node
    );
  });
  expect(selectionHit).toBe(true);
  await selection.click();
  await expect(selection).toBeChecked();
  const selectedCell = selection.locator(
    'xpath=ancestor::*[@role="gridcell"][1]',
  );
  await expect
    .poll(() =>
      selectedCell.evaluate((node) => getComputedStyle(node).boxShadow),
    )
    .not.toBe("none");
  await info.attach("frozen-selected-row-control", {
    body: await selectedCell.screenshot(),
    contentType: "image/png",
  });
  expect(writes).toBe(0);

  const host = await createViewRow(page, fixture.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("frozen-host"),
    "host.display_name": "Host cell crossing the frozen gutter",
    "host.hostname": "frozen-gutter.example.test",
  });
  await page.goto(
    `/?incident_id=${fixture.incident}&view_schema_id=${hostsViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(hostsViewSchemaId)),
  ).toBeVisible();
  const hostSource = page.getByTestId(
    rowCellTestId(host.record_id, "host.display_name"),
  );
  const hostGutter = page.getByTestId(
    gridRowGutterTestId(hostsViewSchemaId, host.record_id),
  );
  await hostSource.evaluate(
    (content, gutterSelector) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const source = content.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(gutterSelector);
      if (!root || !source || !cover)
        throw new Error("Expected mounted Hosts source and gutter");
      const bounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        source.getBoundingClientRect().left - (bounds.left + bounds.width / 4);
    },
    dataTestIdSelector(gridRowGutterTestId(hostsViewSchemaId, host.record_id)),
  );
  const hostWithSource = await hostGutter.screenshot();
  await hostSource.evaluate((content) => {
    const cell = content.closest<HTMLElement>('[role="gridcell"]');
    if (cell) cell.style.visibility = "hidden";
  });
  let hostWithoutSource: Buffer;
  try {
    hostWithoutSource = await hostGutter.screenshot();
  } finally {
    await hostSource.evaluate((content) => {
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      if (cell) cell.style.visibility = "";
    });
  }
  expect(hostWithSource.equals(hostWithoutSource)).toBe(true);
  await info.attach("hosts-frozen-gutter-crossing", {
    body: hostWithSource,
    contentType: "image/png",
  });
});

test("Workbook frozen data cell occludes scrolling paint at the effective boundary", async ({
  page,
  workerAdmin,
}, info) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  preferences.select("comfortable");
  const fixture = await seed(page);
  const recordId = fixture.rows[0]?.record_id;
  if (!recordId) throw new Error("Missing committed Timeline row");
  const frozenField = "timeline.date_entered_text";
  await firstColumn(page, frozenField);
  await setFreeze(page, frozenField);

  const grid = page.locator(gridScrollportSelector());
  await expect(grid).toHaveAttribute("data-grid-freeze-state", "active");
  const frozen = page.getByTestId(rowCellTestId(recordId, frozenField));
  const scrolling = page.getByTestId(rowCellTestId(recordId, summary));
  const frozenCell = frozen.locator('xpath=ancestor::*[@role="gridcell"][1]');
  const scrollingGridCell = scrolling.locator(
    'xpath=ancestor::*[@role="gridcell"][1]',
  );
  await scrolling.evaluate(
    (content, frozenSelector) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const source = content.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(frozenSelector);
      if (!root || !source || !cover)
        throw new Error("Expected mounted frozen and scrolling cells");
      const coverBounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        source.getBoundingClientRect().left -
        (coverBounds.left + coverBounds.width / 2);
    },
    dataTestIdSelector(rowCellTestId(recordId, frozenField)),
  );
  const sourceBounds = await scrollingGridCell.boundingBox();
  const coverBounds = await frozenCell.boundingBox();
  if (!sourceBounds || !coverBounds)
    throw new Error("Expected crossing frozen-data geometry");
  expect(sourceBounds.x).toBeLessThan(coverBounds.x + coverBounds.width);
  expect(sourceBounds.x + sourceBounds.width).toBeGreaterThan(coverBounds.x);

  const withSource = await frozenCell.screenshot();
  await scrolling.evaluate((content) => {
    const cell = content.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Expected mounted scrolling cell");
    cell.style.visibility = "hidden";
  });
  let withoutSource: Buffer;
  try {
    withoutSource = await frozenCell.screenshot();
  } finally {
    await scrolling.evaluate((content) => {
      const cell = content.closest<HTMLElement>('[role="gridcell"]');
      if (cell) cell.style.visibility = "";
    });
  }
  const styles = await frozen.evaluate((content) => {
    const cell = content.closest<HTMLElement>('[role="gridcell"]');
    if (!cell) throw new Error("Expected mounted frozen data cell");
    const value = getComputedStyle(cell);
    const rect = cell.getBoundingClientRect();
    return {
      backgroundColor: value.backgroundColor,
      backgroundImage: value.backgroundImage,
      zIndex: value.zIndex,
      hit:
        document
          .elementFromPoint(
            rect.left + rect.width / 2,
            rect.top + rect.height / 2,
          )
          ?.closest('[role="gridcell"]') === cell,
    };
  });
  await info.attach("frozen-data-crossing-before", {
    body: withSource,
    contentType: "image/png",
  });
  await info.attach("frozen-data-crossing-source-hidden", {
    body: withoutSource,
    contentType: "image/png",
  });
  await info.attach("frozen-data-render-characterization", {
    body: JSON.stringify(styles),
    contentType: "application/json",
  });
  expect(withSource.equals(withoutSource)).toBe(true);
  expect(styles.hit).toBe(true);

  await page.setViewportSize({ width: 2048, height: 1440 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "2";
  });
  await scrolling.evaluate(
    (content, frozenSelector) => {
      const root = content.closest<HTMLElement>('[role="grid"]');
      const source = content.closest<HTMLElement>('[role="gridcell"]');
      const cover = root?.querySelector<HTMLElement>(frozenSelector);
      if (!root || !source || !cover)
        throw new Error("Expected zoomed frozen and scrolling cells");
      const bounds = cover.getBoundingClientRect();
      root.scrollLeft +=
        (source.getBoundingClientRect().left -
          (bounds.left + bounds.width / 2)) /
        2;
    },
    dataTestIdSelector(rowCellTestId(recordId, frozenField)),
  );
  const zoomedSourceBounds = await scrollingGridCell.boundingBox();
  const zoomedCoverBounds = await frozenCell.boundingBox();
  if (!zoomedSourceBounds || !zoomedCoverBounds)
    throw new Error("Expected zoomed crossing geometry");
  expect(zoomedSourceBounds.x).toBeLessThan(
    zoomedCoverBounds.x + zoomedCoverBounds.width,
  );
  expect(zoomedSourceBounds.x + zoomedSourceBounds.width).toBeGreaterThan(
    zoomedCoverBounds.x,
  );
  const zoomedWithSource = await frozenCell.screenshot();
  await scrollingGridCell.evaluate((node) => {
    (node as HTMLElement).style.visibility = "hidden";
  });
  let zoomedWithoutSource: Buffer;
  try {
    zoomedWithoutSource = await frozenCell.screenshot();
  } finally {
    await scrollingGridCell.evaluate((node) => {
      (node as HTMLElement).style.visibility = "";
    });
  }
  expect(zoomedWithSource.equals(zoomedWithoutSource)).toBe(true);
  await info.attach("frozen-data-zoom-two-crossing", {
    body: zoomedWithSource,
    contentType: "image/png",
  });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  await changeGrouping(page, surface, "timeline.capture_state");
  const groupValue = String(
    fixture.rows[0]?.cells["timeline.capture_state"]?.value,
  );
  const group = page.getByTestId(
    gridGroupRowTestId(surface, "timeline.capture_state", groupValue),
  );
  await expect(group).toBeVisible();
  const groupSurfaces = await group.evaluate((button) => {
    const row = button.closest<HTMLElement>('[role="row"]');
    const frozen = button.closest<HTMLElement>('[role="gridcell"]');
    if (!row || !frozen) throw new Error("Expected grouped frozen gutter");
    return {
      row: getComputedStyle(row).backgroundColor,
      frozen: getComputedStyle(frozen).backgroundColor,
      position: getComputedStyle(frozen).position,
    };
  });
  expect(groupSurfaces.frozen).toBe(groupSurfaces.row);
  expect(groupSurfaces.position).toBe("sticky");
});

test("Workbook frozen columns retain semantic placement saved bytes and drafts through suspension", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page, 30);
  await setWidth(page, "240");
  const grid = page.locator(gridScrollportSelector());
  const freeze = async () => {
    await showColumns(page);
    await columns(page)
      .getByRole("button", {
        name: `Freeze through ${requireViewContract(surface).fieldMap[summary]?.label}`,
        exact: true,
      })
      .click();
    await closeColumns(page);
  };
  await freeze();
  await expect(grid).toHaveAttribute("data-grid-freeze-state", "active");
  const initial = await header(page).boundingBox();
  await grid.evaluate((node) => {
    node.scrollLeft = 600;
  });
  await expect
    .poll(async () => Math.round((await header(page).boundingBox())?.x ?? -1))
    .toBe(Math.round(initial?.x ?? -2));
  await expect(header(page)).toHaveClass(/cartulary-grid-frozen-boundary/);
  await expect
    .poll(() =>
      header(page).evaluate(
        (node) => getComputedStyle(node, "::before").borderInlineEndStyle,
      ),
    )
    .toBe("double");
  await setSavedViewDraftName(page, surface, "Frozen identifying column");
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/incidents/${f.incident}/saved-views`),
  );
  await createSavedViewFromCurrentSurface(page, surface);
  const saved = (await (await response).json()).data;
  expect(saved.layout_json.frozen_through_field_key).toBe(summary);
  expect(saved.layout_json.layout_schema_id).toBe("cartulary.layout.v2");
  const savedBytes = JSON.stringify(saved.layout_json);
  await expect(page.getByTestId(savedViewModifiedTestId(surface))).toBeHidden();
  const record = f.rows[0]?.record_id;
  if (!record) throw new Error("Missing frozen row fixture");
  await page.getByTestId(rowCellTestId(record, summary)).click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      recordId: record,
      fieldKey: summary,
      surface: "grid",
    }),
  );
  await editor.fill("Retained frozen draft");
  await editor.press("Home");
  for (let i = 0; i < 4; i += 1) await editor.press("ArrowRight");
  await editor.evaluate((node) => {
    (window as unknown as { frozenEditor: Element }).frozenEditor = node;
  });
  const caretSamples: { stage: string; start: number | null }[] = [];
  const sampleCaret = async (stage: string) => {
    caretSamples.push({
      stage,
      start: await editor.evaluate(
        (node) =>
          (node as HTMLInputElement | HTMLTextAreaElement).selectionStart,
      ),
    });
  };
  await sampleCaret("after-set");
  let writes = 0;
  page.on("request", (r) => {
    if (
      r.method() !== "GET" &&
      /\/(records|rows)(\/|$)/.test(new URL(r.url()).pathname)
    )
      writes += 1;
  });
  await showColumns(page);
  await sampleCaret("after-open-columns");
  // Layout commands borrow focus without submitting the unrelated draft.
  await columns(page)
    .getByRole("button", { name: "Unfreeze columns", exact: true })
    .click();
  await sampleCaret("after-unfreeze");
  await expect(editor).toHaveValue("Retained frozen draft");
  await columns(page)
    .getByRole("button", {
      name: "Freeze through Activity Synopsis",
      exact: true,
    })
    .click();
  await sampleCaret("after-refreeze");
  expect(
    await editor.evaluate(
      (node) =>
        node === (window as unknown as { frozenEditor: Element }).frozenEditor,
    ),
  ).toBe(true);
  for (const direction of ["later", "earlier"]) {
    await columns(page)
      .getByRole("button", {
        name: `Move Activity Synopsis ${direction}`,
        exact: true,
      })
      .click();
    await expect(columns(page)).toBeVisible();
    await expect(editor).toHaveCount(1);
    await expect(editor).toHaveValue("Retained frozen draft");
    await expect(editor).not.toBeFocused();
    await expect
      .poll(() =>
        editor.evaluate((node) => ({
          recordId: node
            .closest("[data-grid-record-id]")
            ?.getAttribute("data-grid-record-id"),
          fieldKey: node
            .closest("[data-grid-field-key]")
            ?.getAttribute("data-grid-field-key"),
        })),
      )
      .toEqual({ recordId: record, fieldKey: summary });
    if (direction === "later")
      await expect(
        page.getByTestId(rowCellTestId(record, "timeline.date_entered_text")),
      ).toHaveCount(1);
    await expect(
      page.getByTestId(
        timelineScalarEditorTestId({
          recordId: record,
          fieldKey: "timeline.date_entered_text",
          surface: "grid",
        }),
      ),
    ).toHaveCount(0);
  }
  await closeColumns(page);
  // The retained editor follows its semantic field through both moves without
  // opening the intervening field or stealing command focus.
  await expect(trigger(page)).toBeFocused();
  await expect(editor).toHaveValue("Retained frozen draft");
  await editor.evaluate((node) => {
    (window as unknown as { frozenEditor: Element }).frozenEditor = node;
  });
  await showColumns(page);
  await closeColumns(page);
  await page.setViewportSize({ width: 390, height: 900 });
  await expect(grid).toHaveAttribute("data-grid-freeze-state", "suspended");
  await sampleCaret("after-suspend");
  await expect(grid.locator(".cartulary-grid-frozen-data")).toHaveCount(0);
  await info.attach("frozen-layout-suspended", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await expect(page.getByTestId(savedViewModifiedTestId(surface))).toBeHidden();
  expect(
    JSON.stringify(
      (await readSavedView(page, f.incident, saved.saved_view_id)).layout_json,
    ),
  ).toBe(savedBytes);
  await page.setViewportSize({ width: 1440, height: 900 });
  await expect(grid).toHaveAttribute("data-grid-freeze-state", "active");
  await expect(editor).toHaveValue("Retained frozen draft");
  await sampleCaret("after-resume");
  await info.attach("frozen-layout-caret-samples", {
    body: JSON.stringify(caretSamples),
    contentType: "application/json",
  });
  expect(
    await editor.evaluate(
      (node) => (node as HTMLInputElement | HTMLTextAreaElement).selectionStart,
    ),
  ).toBe(4);
  expect(
    await editor.evaluate(
      (node) =>
        node === (window as unknown as { frozenEditor: Element }).frozenEditor,
    ),
  ).toBe(true);
  expect(writes).toBe(0);
  await showColumns(page);
  await info.attach("frozen-layout-active-columns", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await closeColumns(page);
  await info.attach("frozen-column-production-geometry", {
    contentType: "application/json",
    body: JSON.stringify(
      await grid.evaluate((node) => ({
        tracks: getComputedStyle(node).gridTemplateColumns,
        width: node.getBoundingClientRect().width,
        scrollLeft: node.scrollLeft,
        frozen: [
          ...node.querySelectorAll(
            '.cartulary-grid-frozen-data[role="columnheader"]',
          ),
        ].map((header) => ({
          key: (header as HTMLElement).dataset.gridFieldKey,
          x: header.getBoundingClientRect().x,
          width: header.getBoundingClientRect().width,
        })),
      })),
    ),
  });
  await setFreeze(page, null);
  await updateSavedViewFromCurrentSurface(page, surface, saved.saved_view_id);
  expect(
    (await readSavedView(page, f.incident, saved.saved_view_id)).layout_json
      .frozen_through_field_key,
  ).toBeNull();
  await expect(page.getByTestId(savedViewModifiedTestId(surface))).toBeHidden();
  await setFreeze(page, summary);
  await updateSavedViewFromCurrentSurface(page, surface, saved.saved_view_id);
  expect(
    (await readSavedView(page, f.incident, saved.saved_view_id)).layout_json
      .frozen_through_field_key,
  ).toBe(summary);
  await expect(editor).toHaveValue("Retained frozen draft");
  expect(writes).toBe(0);
});

test("Workbook frozen prefixes reconcile hidden order all visible fields and exact viewport thresholds", async ({
  page,
}, info) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const f = await seed(page);
  const raw = "timeline.raw_activity_text";
  await firstColumn(page, raw);
  await firstColumn(page, summary);
  await setWidth(page, "120", summary);
  await setWidth(page, "140", raw);
  await showColumns(page);
  const label = (field: string) => {
    const entry = requireViewContract(surface).fieldMap[field];
    if (!entry) throw new Error(`Missing field ${field}`);
    return entry.label;
  };
  for (const checkbox of await columns(page).getByRole("checkbox").all()) {
    if (
      ![label(summary), label(raw)].includes(
        await checkbox.evaluate(
          (node) => node.parentElement?.textContent?.trim() ?? "",
        ),
      )
    )
      await checkbox.uncheck();
  }
  await columns(page)
    .getByRole("button", { name: `Freeze through ${label(raw)}`, exact: true })
    .click();
  await expect(columns(page)).toContainText("2 visible data columns");
  await closeColumns(page);
  const grid = page.locator(gridScrollportSelector());
  const frozenKeys = () =>
    grid
      .locator('.cartulary-grid-frozen-data[role="columnheader"]')
      .evaluateAll((nodes) =>
        nodes.map((node) => (node as HTMLElement).dataset.gridFieldKey),
      );
  await expect.poll(frozenKeys).toEqual([summary, raw]);
  await setSavedViewDraftName(page, surface, "All visible frozen prefix");
  const created = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/incidents/${f.incident}/saved-views`),
  );
  await createSavedViewFromCurrentSurface(page, surface);
  const saved = (await (await created).json()).data;
  const savedBytes = JSON.stringify(saved.layout_json);
  const samples: unknown[] = [];
  await trigger(page).focus();
  for (const budget of [239, 240, 241, 240, 239, 241]) {
    const measured = await grid.evaluate((node, budget) => {
      const root = node as HTMLElement;
      const style = getComputedStyle(root);
      const tracks = style.gridTemplateColumns
        .split(/\s+/)
        .map(Number.parseFloat);
      const headers = [
        ...root.querySelectorAll<HTMLElement>('[role="columnheader"]'),
      ];
      const structural = headers.filter(
        (header) =>
          header.classList.contains("cartulary-grid-selection-header-cell") ||
          header.classList.contains("cartulary-grid-gutter-header-cell"),
      ).length;
      const prefix = tracks.slice(0, structural + 2).reduce((a, b) => a + b, 0);
      const borderAndScrollbar = root.offsetWidth - root.clientWidth;
      root.style.minInlineSize = "0";
      root.style.inlineSize = `${prefix + budget + borderAndScrollbar}px`;
      root.style.maxInlineSize = `${prefix + budget + borderAndScrollbar}px`;
      window.dispatchEvent(new Event("resize"));
      return { budget, prefix, structural, tracks };
    }, budget);
    await expect(grid).toHaveAttribute(
      "data-grid-freeze-state",
      budget < 240 ? "suspended" : "active",
    );
    for (let frame = 0; frame < 3; frame++) {
      await page.evaluate(
        () =>
          new Promise<void>((resolve) =>
            requestAnimationFrame(() => {
              window.dispatchEvent(new Event("resize"));
              resolve();
            }),
          ),
      );
      await expect(grid).toHaveAttribute(
        "data-grid-freeze-state",
        budget < 240 ? "suspended" : "active",
      );
    }
    await expect.poll(frozenKeys).toEqual(budget < 240 ? [] : [summary, raw]);
    await expect(trigger(page)).toBeFocused();
    await expect(
      page.getByTestId(savedViewModifiedTestId(surface)),
    ).toBeHidden();
    samples.push({
      ...measured,
      clientWidth: await grid.evaluate((node) => node.clientWidth),
    });
  }
  expect(
    JSON.stringify(
      (await readSavedView(page, f.incident, saved.saved_view_id)).layout_json,
    ),
  ).toBe(savedBytes);
  await grid.evaluate((node) => {
    const style = (node as HTMLElement).style;
    style.removeProperty("inline-size");
    style.removeProperty("max-inline-size");
    style.removeProperty("min-inline-size");
  });
  await showColumns(page);
  await columns(page)
    .getByRole("checkbox", { name: label(raw), exact: true })
    .uncheck();
  await expect(columns(page)).toContainText(`${label(raw)} (hidden)`);
  await expect(columns(page)).toContainText("1 visible data column.");
  await expect.poll(frozenKeys).toEqual([summary]);
  await columns(page)
    .getByRole("button", { name: `Move ${label(raw)} earlier`, exact: true })
    .click();
  await expect(columns(page)).toContainText("0 visible data columns.");
  await expect.poll(frozenKeys).toEqual([]);
  await columns(page)
    .getByRole("checkbox", { name: label(raw), exact: true })
    .check();
  await expect.poll(frozenKeys).toEqual([raw]);
  await columns(page)
    .getByRole("checkbox", { name: label(summary), exact: true })
    .uncheck();
  await columns(page)
    .getByRole("checkbox", { name: label(raw), exact: true })
    .uncheck();
  await expect(columns(page)).toContainText(`${label(raw)} (hidden)`);
  await columns(page)
    .getByRole("button", { name: "Unfreeze columns", exact: true })
    .click();
  await expect(columns(page)).toContainText("No frozen data columns.");
  await columns(page)
    .getByRole("button", { name: "Reset columns", exact: true })
    .click();
  await closeColumns(page);
  await expect(grid).toHaveAttribute("data-grid-freeze-state", "none");
  await openSavedViewActionMenu(page, surface);
  await page
    .getByTestId(savedViewResetButtonTestId(surface, saved.saved_view_id))
    .click();
  await expect.poll(frozenKeys).toEqual([summary, raw]);
  await expect(page.getByTestId(savedViewModifiedTestId(surface))).toBeHidden();
  await info.attach("frozen-prefix-threshold-measurements", {
    body: JSON.stringify(samples),
    contentType: "application/json",
  });
});
