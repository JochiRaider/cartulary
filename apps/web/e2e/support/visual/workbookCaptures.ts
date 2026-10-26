import {
  dataTestIdSelector,
  gridShellTestId,
  timelineInspectorSectionTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookConflictResolverTestId,
  workbookFocusAnchorTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "../../fixtures";
import { assertVisualRegression } from "./capture";
import {
  attachWorkbookGridVisualDiagnostics,
  type GridVisualRegressionOptions,
  normalizeWorkbookGridVisualState,
  readWorkbookGridScroll,
  setWorkbookGridScroll,
  type WorkbookGridVisualScrollSnapshot,
  waitForVisualLayoutFrame,
} from "./workbookGeometry";

export async function blurActiveElement(page: Page) {
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  });
}

export async function hideVisualTechnicalColumns(
  page: Page,
  view: string,
  labels: string[],
) {
  const trigger = page.getByTestId(workbookColumnsMenuTriggerTestId(view));
  await trigger.click();
  const menu = page.getByTestId(workbookColumnsMenuTestId(view));
  for (const label of labels)
    await menu.getByRole("checkbox", { name: label, exact: true }).uncheck();
  await trigger.click();
}

export async function assertStatusStripVisualRegression(
  page: Page,
  name: string,
) {
  const statusStrip = page.getByTestId(workbookShellSlotTestId("status-strip"));
  await expectStatusStripFocusAnchorVisuallyHidden(statusStrip);
  await statusStrip.scrollIntoViewIfNeeded();
  await assertVisualRegression(page, name, statusStrip);
}

export async function assertStatusStripVisualFixture(page: Page, name: string) {
  const cloneTestId = `visual-status-strip-${name}`;
  await page.evaluate(
    ({ cloneSelector, cloneTestId, sourceSelector, sourceTestId }) => {
      document.querySelector(cloneSelector)?.remove();
      const source = document.querySelector(sourceSelector);
      if (!(source instanceof HTMLElement)) {
        throw new Error(`missing status strip fixture source ${sourceTestId}`);
      }
      const sourceRect = source.getBoundingClientRect();
      const clone = source.cloneNode(true) as HTMLElement;
      clone.setAttribute("aria-hidden", "true");
      clone.setAttribute("data-design-fixture", "status-strip");
      clone.setAttribute("data-testid", cloneTestId);
      clone.style.position = "fixed";
      clone.style.insetInlineStart = "0";
      clone.style.insetBlockStart = "0";
      clone.style.inlineSize = `${Math.round(sourceRect.width)}px`;
      clone.style.margin = "0";
      clone.style.zIndex = "2147483647";
      document.body.append(clone);
    },
    {
      cloneSelector: dataTestIdSelector(cloneTestId),
      cloneTestId,
      sourceSelector: dataTestIdSelector(
        workbookShellSlotTestId("status-strip"),
      ),
      sourceTestId: workbookShellSlotTestId("status-strip"),
    },
  );
  try {
    await assertVisualRegression(page, name, page.getByTestId(cloneTestId));
  } finally {
    await page.evaluate((selector) => {
      document.querySelector(selector)?.remove();
    }, dataTestIdSelector(cloneTestId));
  }
}

async function expectStatusStripFocusAnchorVisuallyHidden(
  statusStrip: Locator,
) {
  const focusAnchor = statusStrip.getByTestId(workbookFocusAnchorTestId());
  await expect(focusAnchor).toHaveCount(1);
  await expect
    .poll(
      async () =>
        focusAnchor.evaluate((node) => {
          const rect = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            blockSize: Math.round(rect.height),
            clipPath: style.clipPath,
            inlineSize: Math.round(rect.width),
            overflow: style.overflow,
            position: style.position,
          };
        }),
      {
        message:
          "Expected status-strip focus anchor to remain present but visually hidden",
      },
    )
    .toEqual({
      blockSize: 1,
      clipPath: "inset(50%)",
      inlineSize: 1,
      overflow: "hidden",
      position: "absolute",
    });
}

export async function assertTimelineEditorVisualRegression(
  page: Page,
  name: string,
  editor: Locator,
) {
  const scroll = await readWorkbookGridScroll(page, timelineViewSchemaId);
  await assertVisualRegression(
    page,
    name,
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    {
      renderSurface: timelineViewSchemaId,
      verifyFraming: async () => {
        await expect(editor).toBeFocused();
        await expect(editor).toBeInViewport({ ratio: 1 });
        await expect
          .poll(() => readWorkbookGridScroll(page, timelineViewSchemaId))
          .toEqual(scroll);
      },
    },
  );
}

export async function assertWorkbookGridVisualRegression(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  let expectedScroll: WorkbookGridVisualScrollSnapshot;
  try {
    await assertVisualRegression(
      page,
      name,
      page.getByTestId(gridShellTestId(surface)),
      {
        ...(options.maxDiffPixels === undefined
          ? {}
          : { maxDiffPixels: options.maxDiffPixels }),
        renderSurface: surface,
        verifyFraming: async () => {
          await expect
            .poll(() => readWorkbookGridScroll(page, surface))
            .toEqual(expectedScroll);
        },
        prepareState: async () => {
          await normalizeWorkbookGridVisualState(page, surface, options);
          await normalizeWorkbookInspectorVisualState(page, options);
          expectedScroll =
            "scroll" in options
              ? await setWorkbookGridScroll(page, surface, options.scroll)
              : await readWorkbookGridScroll(page, surface);
        },
      },
    );
  } catch (error) {
    try {
      await attachWorkbookGridVisualDiagnostics(page, name, surface, options);
    } catch {
      // Preserve the assertion failure when the page is already torn down.
    }
    throw error;
  }
}

async function normalizeWorkbookInspectorVisualState(
  page: Page,
  options: GridVisualRegressionOptions,
) {
  if (!("anchor" in options)) {
    return;
  }
  switch (options.anchor.kind) {
    case "timelineEvidenceActions": {
      const evidenceSection = page.getByTestId(
        timelineInspectorSectionTestId("evidence"),
      );
      await evidenceSection.scrollIntoViewIfNeeded();
      await expect(evidenceSection).toContainText("Attached evidence count: 1");
      await waitForVisualLayoutFrame(page);
      break;
    }
  }
}

async function parkVisualPointer(page: Page) {
  await page.mouse.move(720, 360);
  await waitForVisualLayoutFrame(page);
}

export async function stabilizeConflictResolverVisual(page: Page) {
  if (!(await page.getByTestId(workbookConflictResolverTestId()).isVisible()))
    await page
      .getByRole("button", { name: "Open conflict recovery", exact: true })
      .click();
  const gridEditorIsActive = await page.evaluate(
    () => document.activeElement?.closest('[role="gridcell"]') !== null,
  );
  if (gridEditorIsActive) {
    await page.keyboard.press("Escape");
  }
  await expect(
    page.getByTestId(workbookConflictResolverTestId()),
  ).toBeVisible();
  await expect(
    page.getByText("Resolve the existing field conflict before editing.", {
      exact: true,
    }),
  ).toHaveCount(0);
  await blurActiveElement(page);
  await parkVisualPointer(page);
}
