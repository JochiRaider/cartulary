import {
  dataTestIdSelector,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  savedViewSelectorTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import {
  blockedAccessibilityScenarioTitles,
  scenarioTitlesForAccessibilityRow,
} from "./a11yPhaseMap";
import { expect, test } from "./fixtures";
import { createIncident, createViewRow, uniqueIncidentKey, uniqueTxn } from "./helpers";
import {
  indicatorsViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "../src/workbookSurfaceRegistry";

type ViewRow = {
  record_id: string;
};

const focusableSelector =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
const p2AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P2-01",
) as [string];

if (p2AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P2-01 must declare exactly 1 scenario; found ${p2AccessibilityScenarioTitles.length}`,
  );
}

const p2AccessibilityScenarioTitle = p2AccessibilityScenarioTitles[0];
const laterPhaseAccessibilityScenarioTitles =
  blockedAccessibilityScenarioTitles().filter(
    (title) => title !== p2AccessibilityScenarioTitle,
  );

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

async function expectInteractiveControlsNamed(page: Page) {
  const unnamedControls = await page.locator("body").evaluate((body) => {
    function isVisible(element: Element) {
      const htmlElement = element as HTMLElement;
      const style = window.getComputedStyle(htmlElement);
      return (
        style.visibility !== "hidden" &&
        style.display !== "none" &&
        htmlElement.getClientRects().length > 0
      );
    }

    function labelledByText(element: Element) {
      const labelledBy = element.getAttribute("aria-labelledby") ?? "";
      return labelledBy
        .split(/\s+/u)
        .filter(Boolean)
        .map((id) => document.getElementById(id)?.textContent?.trim() ?? "")
        .join(" ")
        .trim();
    }

    function labelText(element: Element) {
      if (
        element instanceof HTMLInputElement ||
        element instanceof HTMLSelectElement ||
        element instanceof HTMLTextAreaElement
      ) {
        return Array.from(element.labels ?? [])
          .map((label) => label.textContent?.trim() ?? "")
          .join(" ")
          .trim();
      }
      return "";
    }

    function accessibleName(element: Element) {
      return (
        element.getAttribute("aria-label")?.trim() ||
        labelledByText(element) ||
        labelText(element) ||
        element.textContent?.trim() ||
        element.getAttribute("title")?.trim() ||
        element.getAttribute("placeholder")?.trim() ||
        ""
      );
    }

    return Array.from(
      body.querySelectorAll(
        'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [role="button"]:not([disabled]), [tabindex]:not([tabindex="-1"])',
      ),
    )
      .filter(isVisible)
      .filter((element) => accessibleName(element) === "")
      .map((element) => {
        const tag = element.tagName.toLowerCase();
        const testId = element.getAttribute("data-testid") ?? "";
        const role = element.getAttribute("role") ?? "";
        return `${tag}${testId ? ` data-testid:${testId}` : ""}${role ? ` role:${role}` : ""}`;
      });
  });
  expect(unnamedControls).toEqual([]);
}

async function expectVisibleFocus(locator: Locator) {
  await locator.focus();
  await expect(locator).toBeFocused();
  const hasVisibleFocus = await locator.evaluate((element) => {
    const style = window.getComputedStyle(element);
    const outlineVisible =
      style.outlineStyle !== "none" &&
      style.outlineWidth !== "0px" &&
      style.outlineColor !== "transparent";
    const shadowVisible = style.boxShadow !== "none";
    return outlineVisible || shadowVisible;
  });
  expect(hasVisibleFocus).toBeTruthy();
}

async function expectNoSimpleFocusTrap(page: Page) {
  await page.evaluate(() => {
    const sentinel = document.createElement("button");
    sentinel.id = "a11y-preflight-sentinel";
    sentinel.setAttribute("aria-label", "Accessibility preflight sentinel");
    sentinel.setAttribute("type", "button");
    sentinel.style.position = "fixed";
    sentinel.style.inlineSize = "1px";
    sentinel.style.blockSize = "1px";
    sentinel.style.opacity = "0";
    sentinel.style.pointerEvents = "none";
    document.body.prepend(sentinel);
    sentinel.focus();
  });
  try {
    await page.keyboard.press("Tab");
    const first = await page.evaluate(() => {
      const active = document.activeElement;
      return active instanceof HTMLElement && active !== document.body
        ? active.outerHTML
        : "";
    });
    await page.keyboard.press("Tab");
    const second = await page.evaluate(() => {
      const active = document.activeElement;
      return active instanceof HTMLElement && active !== document.body
        ? active.outerHTML
        : "";
    });
    expect(first).not.toBe("");
    expect(second).not.toBe("");
    expect(second).not.toBe(first);
  } finally {
    await page.evaluate(() => {
      document.getElementById("a11y-preflight-sentinel")?.remove();
    });
  }
}

test.describe("frontend accessibility preflight smoke for blocked later phases", () => {
  test(p2AccessibilityScenarioTitle, async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP2"),
      "FE-A11Y-P2-01 workbook shell",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p2-01-row"),
        "timeline.occurred_at": "2026-05-31T09:00:00Z",
        "timeline.summary": "FE-P2 accessibility shell row",
        "timeline.details": "Inspector control coverage",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    await expect(shell).toHaveAttribute("aria-label", "Workbook shell");

    for (const slot of workbookShellSlots) {
      const slotRegion = shell.locator(
        dataTestIdSelector(workbookShellSlotTestId(slot)),
      );
      await expect(slotRegion).toBeVisible();
      await expect(slotRegion).toHaveAttribute("role", "region");
      await expect(slotRegion).toHaveAttribute("aria-label", /.+/);
    }

    for (const surface of requiredBuiltInWorkbookSurfaceIds) {
      const tab = page.getByTestId(surfaceTabTestId(surface));
      await expect(tab).toBeVisible();
      await expect(tab).not.toHaveText("");
    }
    await expect(page.getByTestId(surfaceTabTestId(timelineViewSchemaId))).toHaveAttribute(
      "aria-current",
      "page",
    );

    const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
    await expect(trigger).toBeVisible();
    await expect(trigger).toHaveAttribute("aria-label", "System views");
    await expectVisibleFocus(trigger);
    await page.keyboard.press("Enter");

    const menu = page.getByTestId(systemViewSwitcherMenuTestId());
    await expect(menu).toBeVisible();
    await expect(menu).toHaveAttribute("role", "menu");
    const indicatorOption = page.getByTestId(
      systemViewSwitcherOptionTestId("scope-assessment", indicatorsViewSchemaId),
    );
    await expect(indicatorOption).toBeFocused();
    await expect(indicatorOption).toHaveAttribute("role", "menuitemradio");
    await expect(indicatorOption).toHaveAttribute("aria-checked", "false");
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await expect(trigger).toBeFocused();

    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridFilterApplyTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
    ).toBeVisible();

    const inspectButton = page.getByTestId(
      rowInspectButtonTestId(timelineRow.record_id),
    );
    await expect(inspectButton).toBeVisible();
    await expectVisibleFocus(inspectButton);
    await inspectButton.click();

    const inspector = page.getByTestId("timeline-inspector");
    await expect(inspector).toBeVisible();
    await expect(inspector).toHaveAttribute("aria-label", "Timeline inspector");
    await expect(
      page.getByTestId(
        rowInspectorFieldTestId(timelineRow.record_id, "timeline.details"),
      ),
    ).toBeVisible();

    const saveState = page.getByTestId(saveStateTestId());
    await expect(saveState).toBeVisible();
    await expect(saveState).toHaveAttribute("role", "status");
    await expect(saveState).toHaveText("Saved");

    await expectInteractiveControlsNamed(page);
    await expectNoSimpleFocusTrap(page);
  });

  for (const title of laterPhaseAccessibilityScenarioTitles) {
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
