import {
  dataTestIdSelector,
  timelineCaptureActionTestId,
  workbookInspectorToggleTestId,
  workbookRowContextMenuTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleSemanticGridCellFocus } from "./support/accessibility/grid";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { openTimelineSupersessionFixture } from "./support/workbook/timelineCaptureActions";

test("a11y.timeline-capture review and recovery retain keyboard focus across responsive layouts", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const { target } = await openTimelineSupersessionFixture(page);
  const editor = page.getByTestId(
    timelineCaptureActionTestId("editor", target.record_id),
  );
  const confirm = page.getByTestId(
    timelineCaptureActionTestId("confirm", target.record_id),
  );
  const back = editor.getByRole("button", { name: "Back", exact: true });
  await expectVisibleFocus(confirm);
  await back.focus();
  await back.press("Enter");
  await expectVisibleFocus(
    page.getByTestId(timelineCaptureActionTestId("reason", target.record_id)),
  );
  await page
    .getByTestId(timelineCaptureActionTestId("review", target.record_id))
    .focus();
  await page.keyboard.press("Enter");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [confirm, back]) {
      await control.scrollIntoViewIfNeeded();
      await control.focus();
      await expectVisibleFocus(control);
      const box = await control.boundingBox();
      if (!box) throw new Error("Timeline control bounds unavailable");
      expect(box.x).toBeGreaterThanOrEqual(-1);
      expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
      expect(box.y).toBeGreaterThanOrEqual(-1);
      expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
    }
    expect(
      await page.evaluate(
        () =>
          document.documentElement.scrollWidth <=
          document.documentElement.clientWidth + 1,
      ),
    ).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await confirm.scrollIntoViewIfNeeded();
    await confirm.focus();
    await expectVisibleFocus(confirm);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await back.scrollIntoViewIfNeeded();
      await back.focus();
      await expectVisibleFocus(back);
    },
  );
  await expectAllInteractiveControlsNamed(page);
  await confirm.focus();
  await confirm.press("Enter");
  const trigger = recoveryEntry(page);
  await trigger.focus();
  await openRecoveryItem(page, /^Timeline action ·/);
  await expect(
    page.getByRole("region", { name: "Timeline action recovery", exact: true }),
  ).toBeVisible();
  await expectVisibleFocus(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  );
  await expect(
    page.getByTestId(timelineCaptureActionTestId("result", target.record_id)),
  ).toHaveText("Timeline supersession completed.");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    const close = page.getByRole("button", {
      name: "Close recovery",
      exact: true,
    });
    await close.scrollIntoViewIfNeeded();
    await close.focus();
    await expectVisibleFocus(close);
    const box = await close.boundingBox();
    if (!box) throw new Error("Timeline recovery bounds unavailable");
    expect(box.x).toBeGreaterThanOrEqual(-1);
    expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
    expect(box.y).toBeGreaterThanOrEqual(-1);
    expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.keyboard.press("Escape");
  await expectVisibleFocus(trigger);
});

test("a11y.timeline-clear selected-cell action retains keyboard access across supported layouts", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YCLEAR"),
    "Clear accessibility",
  );
  const field = "timeline.activity_synopsis_text";
  const created = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("clear-a11y-seed"),
    [field]: "Clear this source",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  const cell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    created.record_id,
    field,
  );
  await cell.click();
  await page.keyboard.press("Escape");
  const clear = page.getByRole("button", {
    name: "Clear contents",
    exact: true,
  });
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 720 },
  ]) {
    await page.setViewportSize(viewport);
    await page
      .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
      .focus();
    for (let index = 0; index < 4; index++) {
      await page.keyboard.press("Shift+Tab");
      if (await clear.evaluate((element) => element === document.activeElement))
        break;
    }
    await expect(clear).toBeFocused();
    await expectDecisionControlReachable(page, clear);
    await expectVisibleFocus(clear);
    const response = page.waitForResponse(
      (response) =>
        response.url().endsWith("/bulk-mutations") &&
        response.request().method() === "POST",
    );
    await page.keyboard.press("Enter");
    expect((await response).ok()).toBe(true);
    await expect(clear).toBeFocused();
    await testInfo.attach(`timeline-clear-${viewport.width}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 2560, height: 1440 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await withPresentation(
      page,
      {
        textSpacing:
          "#root * { letter-spacing:0.12em !important; word-spacing:0.16em !important; line-height:1.5 !important; }",
      },
      async () => {
        await expectDecisionControlReachable(page, clear);
        await expectVisibleFocus(clear);
        await testInfo.attach("timeline-clear-zoom-spacing", {
          body: await page.screenshot(),
          contentType: "image/png",
        });
      },
    );
  });
});

test("a11y.timeline-row-actions focus and menu scrolling remain reachable across zoom and viewport changes", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ROW-MENU"),
    "Timeline row menu geometry",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("row-menu"),
    "timeline.activity_synopsis_text": "Row action focus and overflow",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await withPresentation(
    page,
    {
      textSpacing: `${dataTestIdSelector(workbookRowContextMenuTestId(timelineViewSchemaId, row.record_id))} button { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; }`,
    },
    async () => {
      for (const [width, height, zoom] of [
        [1440, 900, 1],
        [1024, 720, 1],
        [768, 640, 1],
        [390, 720, 1],
        [1280, 720, 2],
      ]) {
        await page.setViewportSize({
          width: width ?? 1440,
          height: height ?? 900,
        });
        await page.locator("html").evaluate((node, scale) => {
          node.style.zoom = String(scale);
        }, zoom ?? 1);
        const cell = await mountedGridCell(
          page,
          timelineViewSchemaId,
          row.record_id,
          "timeline.activity_synopsis_text",
        );
        await expectVisibleSemanticGridCellFocus(cell);
        await page.keyboard.press("Shift+F10");
        const menu = page.getByTestId(
          workbookRowContextMenuTestId(timelineViewSchemaId, row.record_id),
        );
        await expect(menu).toBeVisible();
        for (let index = 0; index < 4; index++) {
          const focused = menu.locator(":focus");
          await expect(focused).toBeVisible();
          const box = await focused.boundingBox();
          expect(box).not.toBeNull();
          expect(box?.x).toBeGreaterThanOrEqual(0);
          expect(box?.y).toBeGreaterThanOrEqual(0);
          expect((box?.x ?? 0) + (box?.width ?? 0)).toBeLessThanOrEqual(
            (width ?? 1440) + 1,
          );
          expect((box?.y ?? 0) + (box?.height ?? 0)).toBeLessThanOrEqual(
            (height ?? 900) + 1,
          );
          await page.keyboard.press("ArrowDown");
        }
        await testInfo.attach(`row-menu-${width}-${height}-zoom-${zoom}`, {
          body: await page.screenshot(),
          contentType: "image/png",
        });
        await page.keyboard.press("Escape");
        await expect(menu).toHaveCount(0);
        await expectVisibleSemanticGridCellFocus(cell);
        await page.keyboard.press("Shift+F10");
        await expect(menu).toBeVisible();
        await page.setViewportSize({
          width: width ?? 1440,
          height: (height ?? 900) - 20,
        });
        await expect(menu).toHaveCount(0);
      }
    },
  );
});
