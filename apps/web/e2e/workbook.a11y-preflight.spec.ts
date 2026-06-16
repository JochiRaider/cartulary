import { assertActiveFilterChipVisible } from "@cartulary/test-utils";
import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  savedViewCreateButtonTestId,
  savedViewNameInputTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorSectionTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineScalarEditorTestId,
  workbookRowActionMenuButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";
import {
  blockedAccessibilityScenarioTitles,
  scenarioTitlesForAccessibilityRow,
} from "./a11yPhaseMap";
import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { clickTimelineRowAction, openTimelineInspector } from "./phase4Helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const p8AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P8-01",
) as [string];
const p11AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P11-01",
) as [string];

if (p8AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P8-01 must declare exactly 1 scenario; found ${p8AccessibilityScenarioTitles.length}`,
  );
}
if (p11AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P11-01 must declare exactly 1 scenario; found ${p11AccessibilityScenarioTitles.length}`,
  );
}

const focusableSelector =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
const blockedScenarioTitles = blockedAccessibilityScenarioTitles();
const indicatorsViewSchemaId = "cartulary.view.indicators.v1";

type Rgb = {
  b: number;
  g: number;
  r: number;
};

function parseRgb(value: string): Rgb {
  const match = value.match(
    /rgba?\(\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)\s*,\s*(\d+(?:\.\d+)?)/,
  );
  if (!match) {
    throw new Error(`unsupported color value ${value}`);
  }
  return {
    r: Number(match[1]),
    g: Number(match[2]),
    b: Number(match[3]),
  };
}

function channelLuminance(channel: number): number {
  const value = channel / 255;
  return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
}

function relativeLuminance(color: Rgb): number {
  return (
    0.2126 * channelLuminance(color.r) +
    0.7152 * channelLuminance(color.g) +
    0.0722 * channelLuminance(color.b)
  );
}

function contrastRatio(foreground: Rgb, background: Rgb): number {
  const lighter = Math.max(
    relativeLuminance(foreground),
    relativeLuminance(background),
  );
  const darker = Math.min(
    relativeLuminance(foreground),
    relativeLuminance(background),
  );
  return (lighter + 0.05) / (darker + 0.05);
}

async function expectTextContrast(locator: Locator) {
  const colors = await locator.evaluate((element) => {
    function isTransparent(value: string): boolean {
      return (
        value === "transparent" ||
        value === "rgba(0, 0, 0, 0)" ||
        /rgba\([^)]*,\s*0(?:\.0+)?\)$/.test(value)
      );
    }

    const style = window.getComputedStyle(element);
    let background = style.backgroundColor;
    let candidate = element.parentElement;
    while (isTransparent(background) && candidate) {
      background = window.getComputedStyle(candidate).backgroundColor;
      candidate = candidate.parentElement;
    }
    return {
      background,
      color: style.color,
    };
  });
  expect(
    contrastRatio(parseRgb(colors.color), parseRgb(colors.background)),
  ).toBeGreaterThanOrEqual(4.5);
}

async function expectLaterPhaseSmoke(page: Page) {
  const focusableCount = await page.locator(focusableSelector).count();
  expect(focusableCount).toBeGreaterThan(0);

  await page.locator(focusableSelector).first().focus();
  await page.keyboard.press("Tab");
  const activeElementIsFocusable = await page.evaluate(() => {
    const active = document.activeElement;
    return Boolean(active && active !== document.body);
  });
  expect(activeElementIsFocusable).toBeTruthy();

  const unnamedButtons = await page.locator("button").evaluateAll(
    (buttons) =>
      buttons.filter((button) => {
        const text = button.textContent?.trim() ?? "";
        const ariaLabel = button.getAttribute("aria-label")?.trim() ?? "";
        const titleAttr = button.getAttribute("title")?.trim() ?? "";
        const labelledBy = button.getAttribute("aria-labelledby")?.trim() ?? "";
        return !text && !ariaLabel && !titleAttr && !labelledBy;
      }).length,
  );
  expect(unnamedButtons).toBe(0);
}

async function expectKeyboardReachable(locator: Locator) {
  await locator.focus();
  await expect(locator).toBeFocused();
}

async function setFilterValue(valueControl: Locator, value: string) {
  try {
    await valueControl.selectOption(value);
  } catch {
    await valueControl.fill(value);
  }
}

test.describe("FE-P8 accessibility preflight", () => {
  test(p8AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP8"),
      "Accessibility FE-P8 query controls",
    );
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("A11YP8-REVIEWED"),
        "timeline.summary": "FE-A11Y-P8 reviewed row",
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("A11YP8-ROUGH"),
      "timeline.summary": "FE-A11Y-P8 rough row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.locator("body")).toBeVisible();
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    const summarySortHeader = page.getByTestId(
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
    );
    await expectKeyboardReachable(summarySortHeader);
    await page.keyboard.press("Enter");
    await expect(summarySortHeader).toContainText("Asc");

    await clickTimelineRowAction(
      page,
      reviewedRow.record_id,
      timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
    );
    await expect(
      page.getByTestId(
        rowCellTestId(reviewedRow.record_id, "timeline.capture_state"),
      ),
    ).toHaveText("reviewed");

    const filterField = page.getByTestId(
      gridFilterFieldTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(filterField);
    await filterField.selectOption("timeline.capture_state");
    const filterValue = page.getByTestId(
      gridFilterValueTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(filterValue);
    await setFilterValue(filterValue, "reviewed");
    const filterApply = page.getByTestId(
      gridFilterApplyTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(filterApply);
    await page.keyboard.press("Enter");
    await assertActiveFilterChipVisible(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
    );
    const filterChip = page.getByTestId(
      gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
    );
    await expectKeyboardReachable(filterChip);

    const groupingSelect = page.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(groupingSelect);
    await groupingSelect.selectOption("timeline.capture_state");
    const reviewedGroup = page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    );
    await expect(reviewedGroup).toBeVisible();
    await expectKeyboardReachable(reviewedGroup);
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
    await page.keyboard.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
    await page.keyboard.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");

    const savedViewSelector = page.getByTestId(
      savedViewSelectorTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(savedViewSelector);
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "view_schema",
    );
    const savedViewNameInput = page.getByTestId(
      savedViewNameInputTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(savedViewNameInput);
    await savedViewNameInput.fill("FE-A11Y-P8 keyboard saved view");
    const createSavedViewButton = page.getByTestId(
      savedViewCreateButtonTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(createSavedViewButton);
    await page.keyboard.press("Enter");
    const savedViewStatus = page.getByTestId(
      savedViewStatusTestId(timelineViewSchemaId),
    );
    await expect(savedViewStatus).toHaveAttribute("aria-live", "polite");
    await expect(savedViewStatus).toHaveText("Saved view created.");
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "saved_view",
    );

    const homeButton = page.getByTestId(
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(homeButton);
    await page.keyboard.press("Enter");
    await expect(savedViewStatus).toHaveText("Home view updated.");
    const defaultButton = page.getByTestId(
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(defaultButton);
    await page.keyboard.press("Enter");
    await expect(savedViewStatus).toHaveText("Default view updated.");
  });
});

test.describe("FE-P11 accessibility preflight", () => {
  test(p11AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP11"),
      "Accessibility FE-P11 readiness matrix",
    );
    const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("A11YP11-ROW"),
      "timeline.details": "FE-A11Y-P11 details",
      "timeline.summary": "FE-A11Y-P11 keyboard row",
    });
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("A11YP11-REVIEWED"),
        "timeline.summary": "FE-A11Y-P11 reviewed row",
      },
    );

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.locator("body")).toBeVisible();
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    const timelineTab = page.getByTestId(
      surfaceTabTestId(timelineViewSchemaId),
    );
    await expect(timelineTab).toHaveAttribute("aria-current", "page");
    await expectKeyboardReachable(timelineTab);

    const systemSwitcher = page.getByTestId(systemViewSwitcherTriggerTestId());
    await expectKeyboardReachable(systemSwitcher);
    await expect(systemSwitcher).toHaveAttribute("aria-label", "System views");
    await systemSwitcher.press("Enter");
    await expect(systemSwitcher).toHaveAttribute("aria-expanded", "true");
    const menu = page.getByTestId(systemViewSwitcherMenuTestId());
    await expect(menu).toBeVisible();
    await expect(menu).toHaveAttribute("role", "menu");
    const indicatorsOption = page.getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        indicatorsViewSchemaId,
      ),
    );
    await expect(indicatorsOption).toHaveAttribute("role", "menuitemradio");
    await expect(indicatorsOption).toHaveAttribute("aria-checked", "false");
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await expect(systemSwitcher).toBeFocused();

    const summarySortHeader = page.getByTestId(
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
    );
    await expectKeyboardReachable(summarySortHeader);
    await summarySortHeader.press("Enter");
    await expect(summarySortHeader).toContainText("Asc");

    const summaryCell = page.getByTestId(
      rowCellTestId(row.record_id, "timeline.summary"),
    );
    await expectKeyboardReachable(summaryCell);
    await page.keyboard.press("Enter");
    const summaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.summary",
        recordId: row.record_id,
        surface: "grid",
      }),
    );
    await expect(summaryEditor).toBeVisible();
    await expect(summaryEditor).toHaveAttribute("aria-label", /Summary/);
    await summaryEditor.focus();
    await expect(summaryEditor).toBeFocused();
    await summaryEditor.fill("FE-A11Y-P11 edited summary");
    await page.keyboard.press("Escape");
    await expect(summaryCell).toBeVisible();

    const inspectButton = page.getByTestId(
      workbookRowActionMenuButtonTestId(timelineViewSchemaId, row.record_id),
    );
    await expectKeyboardReachable(inspectButton);
    await openTimelineInspector(page, row.record_id);
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("details")),
    ).toBeVisible();

    await clickTimelineRowAction(
      page,
      reviewedRow.record_id,
      timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
    );
    const groupingSelect = page.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    );
    await expectKeyboardReachable(groupingSelect);
    await groupingSelect.selectOption("timeline.capture_state");
    const reviewedGroup = page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    );
    await expectKeyboardReachable(reviewedGroup);
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");

    const saveState = page.getByTestId(saveStateTestId());
    await expect(saveState).toHaveAttribute("aria-live", "polite");
    await expect(saveState).not.toHaveText("");
    await expectTextContrast(saveState);
    await expectLaterPhaseSmoke(page);
  });
});

test.describe("frontend accessibility preflight smoke for blocked later phases", () => {
  for (const title of blockedScenarioTitles) {
    test(title, async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11Y"),
        `Accessibility ${title.slice(0, 24)}`,
      );
      await page.goto(`/?incident_id=${incidentId}`);

      await expect(page.locator("body")).toBeVisible();
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expectLaterPhaseSmoke(page);
    });
  }
});
