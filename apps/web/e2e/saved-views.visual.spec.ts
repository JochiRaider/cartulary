import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  changeGrouping,
} from "@cartulary/test-utils/grid";
import {
  cartularyDesignTokenVars,
  gridGroupRowTestId,
  gridShellTestId,
  incidentAdministrationTestId,
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewModifiedTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
  timelineCaptureActionTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookAddRowButtonTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorToggleTestId,
  workbookPreferenceTestId,
  workbookQueryEntryTestId,
  workbookQueryOverflowEntryTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { gridSavedRows } from "./pages/workbookInspector";
import { withPresentation } from "./support/accessibility/presentation";
import { selectVisualDensity } from "./support/auth/visualPreferences";
import { tagActionsPayload } from "./support/entities/tagActions";
import { createIncident } from "./support/incidents/fixtures";
import { closeIncidentForVisual } from "./support/incidents/visualState";
import {
  navigateVisualApplication,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { clickTimelineRowAction } from "./support/visual/rowActions";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import { assertWorkbookGridVisualRegression } from "./support/visual/workbookCaptures";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import {
  createSavedViewFromCurrentSurface,
  setCurrentSavedViewAsDefault,
  setCurrentSavedViewAsHome,
  setSavedViewDraftName,
} from "./support/workbook/savedViews";

test("Capture saved-view selector, active chips, grouped result, group row, default/startup state indicator, and empty successful query fixtures.", async ({
  page,
}) => {
  test.setTimeout(120_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  const longSavedViewName =
    "browser.saved-view-query visual layout resilience with a deliberately long selected saved-view name";
  const longTagToken =
    "visual-unbroken-tag-0123456789-abcdefghijklmnopqrstuvwxyz";
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALSAVEDVIEW"),
    "browser.saved-view-query visual saved view query controls",
  );
  const reviewedRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualsavedview-reviewed",
    {
      client_txn_id: uniqueTxn("VISUALSAVEDVIEW-REVIEWED"),
      "timeline.activity_utc_text": "2026-06-08T12:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.saved-view-query reviewed saved-view visual row",
      "timeline.tags": tagActionsPayload([longTagToken]),
    },
  );
  await createVisualTimelineRow(page, incidentId, "visualsavedview-rough", {
    client_txn_id: uniqueTxn("VISUALSAVEDVIEW-ROUGH"),
    "timeline.activity_utc_text": "2026-06-08T12:05:00Z",
    "timeline.activity_synopsis_text":
      "browser.saved-view-query rough grouped visual row",
  });

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
  ).toBeVisible();

  await clickTimelineRowAction(
    page,
    reviewedRow.record_id,
    timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
  );
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      reviewedRow.record_id,
      "timeline.capture_state",
    ),
  ).toHaveText("reviewed");

  await openRecoveryItem(page, /^Timeline action ·/);
  await expect(
    page.getByTestId(
      timelineCaptureActionTestId("result", reviewedRow.record_id),
    ),
  ).toHaveText("Timeline review completed.");
  await page.keyboard.press("Escape");
  await expect(recoveryEntry(page)).toHaveText("Recovery (0)");

  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(
        timelineViewSchemaId,
        "filter",
        "timeline.capture_state",
      ),
    ),
  ).toBeVisible();
  await assertActiveFilterChipVisible(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  await expect(
    page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    ),
  ).toBeVisible();

  const sortTrigger = page.getByTestId(
    workbookSortMenuTriggerTestId(timelineViewSchemaId),
  );
  await sortTrigger.click();
  const sortMenu = page.getByTestId(
    workbookSortMenuTestId(timelineViewSchemaId),
  );
  await expect(sortMenu).toBeVisible();
  for (const fieldKey of [
    "timeline.activity_sort_ts",
    "timeline.date_entered_sort_day",
    "timeline.activity_synopsis_text",
    "timeline.analyst_text",
    "timeline.mitre_stage_text",
    "timeline.device_object_text",
    "timeline.ip_address_text",
    "timeline.capture_state",
  ]) {
    const option = page.getByTestId(
      workbookSortOptionTestId(timelineViewSchemaId, fieldKey),
    );
    await option.click();
    await expect(option).toHaveAttribute("aria-checked", "true");
  }
  await sortMenu.press("Escape");
  await expect(sortMenu).toHaveCount(0);
  await expect(sortTrigger).toBeFocused();
  await expect(page.locator('[data-grid-data-state="refreshing"]')).toHaveCount(
    0,
  );
  await expect(
    page.locator(
      '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
    ),
  ).toHaveCount(0);

  await setSavedViewDraftName(page, timelineViewSchemaId, longSavedViewName);
  await createSavedViewFromCurrentSurface(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
  ).toHaveText("Saved view created.");
  await setCurrentSavedViewAsHome(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "shortcut-outcome")),
  ).toHaveText("Home update confirmed.");
  await setCurrentSavedViewAsDefault(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(workbookPreferenceTestId("default", "shortcut-outcome")),
  ).toHaveText("Incident default update confirmed.");
  await page.keyboard.press("Escape");

  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-saved-view-clean",
  );

  const savedViewActionTrigger = page.getByTestId(
    savedViewActionMenuTriggerTestId(timelineViewSchemaId),
  );
  await savedViewActionTrigger.click();
  await expect(
    page.getByTestId(savedViewActionMenuTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-saved-view-actions",
  );
  await page.keyboard.press("Escape");

  await sortTrigger.click();
  await expect(sortMenu).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-ordered-maximum-sort",
  );
  await page.keyboard.press("Escape");

  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    longTagToken,
  );
  await expect(
    page.getByTestId(savedViewModifiedTestId(timelineViewSchemaId)),
  ).toHaveText("Modified");
  await expect(
    page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
  ).toHaveAttribute("title", longSavedViewName);
  const queryControls = page.getByTestId(
    workbookViewBarQueryControlsTestId(timelineViewSchemaId),
  );
  await expect(queryControls).toHaveAttribute("data-query-chip-capacity", "3");
  await expect(queryControls).toHaveAttribute(
    "data-hidden-query-chip-count",
    "8",
  );
  await expect(
    page
      .getByRole("toolbar", { name: "Active query chips" })
      .locator("button[data-query-entry-key]"),
  ).toHaveCount(3);
  await expect(
    page.getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId)),
  ).toHaveAttribute(
    "aria-label",
    "Filters, 2 active filters, 8 hidden query entries",
  );
  await expect(
    page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
  ).toBeVisible();

  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  const filterTrigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
  );
  await filterTrigger.evaluate((element) => (element as HTMLElement).blur());
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-saved-view-modified",
  );
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-maximum-pressure-base",
  );
  await assertViewportVisualRegression(
    page,
    "workbook-query-saved-view-query-controls",
  );

  await filterTrigger.click();
  const filterPanel = page.getByTestId(
    workbookFilterPopoverTestId(timelineViewSchemaId),
  );
  const overflowEntry = filterPanel.getByTestId(
    workbookQueryOverflowEntryTestId(
      timelineViewSchemaId,
      "filter",
      "timeline.tags",
    ),
  );
  await overflowEntry.click();
  await expect(filterPanel).toHaveAttribute("aria-label", "Edit filter");
  await overflowEntry.focus();
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-filter-editing-overflow",
  );
  await page.keyboard.press("Escape");

  await page
    .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
    .click();
  await expect(
    page.getByTestId(workbookColumnsMenuTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "workbook-view-bar-long-columns");
  await page.keyboard.press("Escape");

  await page.setViewportSize({ width: 1024, height: 720 });
  await expect(queryControls).toHaveAttribute("data-query-chip-capacity", "2");
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-maximum-pressure-narrow",
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await expect(queryControls).toHaveAttribute("data-query-chip-capacity", "0");
  await assertViewportVisualRegression(
    page,
    "workbook-view-bar-maximum-pressure-compact",
  );

  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expect(
      page.getByTestId(workbookResponsiveBandTestId()),
    ).toHaveAttribute(
      "data-workbook-responsive-band",
      "below_supported_minimum",
    );
    await expect(
      page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await assertViewportVisualRegression(page, "workbook-view-bar-zoom-200");
  });

  await withPresentation(
    page,
    {
      textSpacing: `
        * {
          line-height: 1.5 !important;
          letter-spacing: 0.12em !important;
          word-spacing: 0.16em !important;
        }
        p { margin-block-end: 2em !important; }
      `,
    },
    async () => {
      await assertViewportVisualRegression(
        page,
        "workbook-view-bar-text-spacing",
      );
    },
  );

  const emptyIncidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALSAVEDVIEWEMPTY"),
    "browser.saved-view-query empty successful Timeline query",
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await navigateVisualApplication(page, `/?incident_id=${emptyIncidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, emptyIncidentId, timelineViewSchemaId))
          .length,
    )
    .toBe(0);
  await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
  await expect(
    page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .evaluate((element) => {
      element.setAttribute("data-design-fixture", "empty-state");
    });
  const emptyState = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[data-grid-data-state="empty"]');
  await expect(emptyState).toContainText(
    "No Timeline records have been added.",
  );
  await expect(
    emptyState.getByRole("button", { name: "Add row" }),
  ).toBeVisible();
  await assertWorkbookGridVisualRegression(
    page,
    "workbook-query-empty-successful-query",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );

  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(
        timelineViewSchemaId,
        "filter",
        "timeline.capture_state",
      ),
    ),
  ).toBeVisible();
  await assertActiveFilterChipVisible(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
  );
  const filteredEmptyState = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[data-grid-data-state="filtered_empty"]');
  await expect(filteredEmptyState).toContainText(
    "No rows match the current filters.",
  );
  await expect(
    filteredEmptyState.getByRole("button", { name: "Clear filters" }),
  ).toBeVisible();
  await assertWorkbookGridVisualRegression(
    page,
    "workbook-query-filtered-empty",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
  await filteredEmptyState
    .getByRole("button", { name: "Clear filters" })
    .click();
  await expect(emptyState).toBeVisible();
  await expect
    .poll(() =>
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .locator('[role="grid"]')
        .evaluate((grid) => grid === document.activeElement),
    )
    .toBe(true);

  await page.setViewportSize({ width: 1024, height: 720 });
  await assertWorkbookGridVisualRegression(
    page,
    "workbook-query-empty-narrow",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await assertWorkbookGridVisualRegression(
    page,
    "workbook-query-empty-compact",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await assertViewportVisualRegression(page, "workbook-query-empty-zoom-200");
  });
  await withPresentation(
    page,
    {
      textSpacing: `
        * {
          line-height: 1.5 !important;
          letter-spacing: 0.12em !important;
          word-spacing: 0.16em !important;
        }
        p { margin-block-end: 2em !important; }
      `,
    },
    async () => {
      await assertViewportVisualRegression(
        page,
        "workbook-query-empty-text-spacing",
      );
    },
  );

  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));
    const densityGrid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
    await expect(densityGrid).toContainText(
      "No Timeline records have been added.",
    );
    await densityGrid.evaluate((element) => {
      element.setAttribute("data-design-fixture", "empty-state");
    });
    await assertWorkbookGridVisualRegression(
      page,
      `workbook-query-empty-density-${density}`,
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  }
  await selectVisualDensity(page, null);
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));

  await closeIncidentForVisual(page, emptyIncidentId);
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await expect(
    page.getByRole("status", { name: "Incident lifecycle", exact: true }),
  ).toHaveText("Closed, read-only");
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toContainText("No Timeline records have been added.");
  await expect(
    page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
  ).toHaveCount(0);
  const closedEmptyState = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator('[data-grid-data-state="empty"]');
  await expect(closedEmptyState.getByRole("button")).toHaveCount(0);
  await expect(
    page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
  ).toBeDisabled();
  await assertWorkbookGridVisualRegression(
    page,
    "workbook-query-empty-closed-read-only",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
});

test("Capture workbook preferences inspection uncertainty confirmation responsive and density.", async ({
  testActorPage: page,
}) => {
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("WP-VISUAL"),
    "Workbook startup preferences",
  );
  await openIncidentFromLanding(page, incidentId);
  await openIncidentControls(page);
  const panel = page.getByRole("region", {
    name: "Workbook startup preferences",
    exact: true,
  });
  await expect(
    page.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref")),
  ).toHaveText("Unset");
  const captureCurrent = async (name: string, anchor = panel) => {
    await assertViewportVisualRegression(page, name, {
      anchor: { locator: anchor, align: anchor === panel ? "start" : "center" },
    });
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({
        animations: "disabled",
        caret: "hide",
        fullPage: false,
      }),
      contentType: "image/png",
    });
  };
  const capture = async (name: string) => {
    await captureCurrent(name);
  };
  await page.setViewportSize({ width: 1280, height: 720 });
  await capture("workbook-preferences-unset");
  let failRead = false;
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/workbook-preferences/me`,
    async (route) => {
      if (route.request().method() === "PUT") {
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        await route.abort("failed");
      } else if (failRead)
        await route.fulfill({
          status: 503,
          json: {
            error: { code: "internal_error" },
            meta: { request_id: "preference-read" },
          },
        });
      else await route.continue();
    },
  );
  await page.getByTestId(workbookPreferenceTestId("home", "set")).click();
  const keep = page.getByTestId(
    workbookPreferenceTestId("home", "keep-observed"),
  );
  await expect(keep).toHaveAttribute("aria-disabled", "false");
  await keep.scrollIntoViewIfNeeded();
  await captureCurrent("workbook-preferences-uncertain", keep);
  await page.setViewportSize({ width: 390, height: 480 });
  await keep.scrollIntoViewIfNeeded();
  await captureCurrent("workbook-preferences-uncertain-narrow", keep);
  await keep.click();
  await removeOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/workbook-preferences/me`,
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/workbook-preferences/me`,
    async (route) => {
      if (route.request().method() === "PUT") {
        const response = await route.fetch();
        failRead = true;
        await route.fulfill({ response });
      } else if (failRead)
        await route.fulfill({
          status: 503,
          json: {
            error: { code: "internal_error" },
            meta: { request_id: "preference-read" },
          },
        });
      else await route.continue();
    },
  );
  await page.getByTestId(workbookPreferenceTestId("home", "clear")).click();
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "read")),
  ).toContainText("may be stale");
  await page
    .getByTestId(workbookPreferenceTestId("home", "outcome"))
    .scrollIntoViewIfNeeded();
  await captureCurrent("workbook-preferences-confirmed-stale");
  await removeOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/workbook-preferences/me`,
  );
  for (const density of ["compact", "comfortable"] as const) {
    await selectVisualDensity(page, density);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));
    await openIncidentControls(page);
    await page.setViewportSize({ width: 768, height: 640 });
    await expect(panel).toHaveCSS(
      "font-size",
      cartularyDesignTokenVars[`--ct-density-${density}-fontSize`],
    );
    await expect(panel).toHaveCSS(
      "padding",
      cartularyDesignTokenVars[`--ct-density-${density}-cellPadding`],
    );
    await expect(
      page.getByTestId(workbookPreferenceTestId("home", "read")),
    ).toHaveText("");
    await capture(`workbook-preferences-${density}`);
  }
});
