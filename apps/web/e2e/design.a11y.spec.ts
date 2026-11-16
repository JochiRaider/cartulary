import {
  applyFilterChip,
  scrollGridCellIntoView,
} from "@cartulary/test-utils/grid";
import {
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  rowCellTestId,
  rowInspectorFieldTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  decisionsViewSchemaId,
  indicatorsViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  activateTimelineGridEditor,
  expectVisibleSemanticGridCellFocus,
} from "./support/accessibility/grid";
import {
  blurActiveElement,
  expectInitialTabRoundTrip,
  expectTabReachability,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { openA11ySystemSurface } from "./support/accessibility/queryControls";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectWorkbookSavePresentation } from "./support/accessibility/status";
import { createIncident } from "./support/incidents/fixtures";
import { responseBarrier } from "./support/runtime/cleanup";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import {
  expectCellTextOrValue,
  mountedGridCell,
  semanticGridCell,
} from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { openFilterPopover } from "./support/workbook/queryControls";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("a11y.design-readiness.row-01 Verify global accessibility matrix for keyboard access, visible focus, System views, grid navigation/edit entry/exit, Esc, ARIA states, icon-only labels, contrast, and non-color-only empty/loading/error/blocked states.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YDESIGNREADINESS"),
    "a11y.design-readiness global accessibility matrix",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.design-readiness-timeline"),
      "timeline.activity_utc_text": "2026-06-13T09:30:00Z",
      "timeline.activity_synopsis_text": "a11y.design-readiness timeline row",
      "timeline.raw_activity_text": "Global accessibility matrix details",
    },
  );
  const taskRow = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.design-readiness-task"),
      "task.priority": "normal",
      "task.task_kind": "collection",
      "task.title": "a11y.design-readiness task row",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectWorkbookSavePresentation(page);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );

  await expectTabReachability(page, [
    systemViewSwitcherTriggerTestId(),
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
    gridGroupingSelectTestId(timelineViewSchemaId),
    /^(?:draft-row-|row-)/u,
  ]);

  await page.keyboard.press("Escape");
  await blurActiveElement(page);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: timelineRow.record_id,
    surface: timelineViewSchemaId,
  });
  const summaryCell = await activateTimelineGridEditor(
    page,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(summaryCell).toHaveAttribute(
    "aria-label",
    `Activity Synopsis ${timelineRow.record_id}`,
  );
  await summaryCell.fill("a11y.design-readiness edited via keyboard");
  await summaryCell.press("Enter");
  await expect(
    page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("a11y.design-readiness edited via keyboard");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await openTimelineInspector(page, timelineRow.record_id);
  const inspectorSummaryCell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await semanticGridCell(inspectorSummaryCell).click();
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${timelineRow.record_id}:timeline.activity_synopsis_text`,
  );
  const semanticSummaryCell = semanticGridCell(inspectorSummaryCell);
  const editRawActivity = page
    .getByTestId(timelineInspectorTestId())
    .getByRole("button", { name: "Edit RAW Activity", exact: true });
  await expectVisibleFocus(editRawActivity);
  await editRawActivity.click();
  const inspectorDetails = page.getByTestId(
    rowInspectorFieldTestId(
      timelineRow.record_id,
      "timeline.raw_activity_text",
    ),
  );
  await expectVisibleFocus(inspectorDetails);
  await page.keyboard.press("Escape");
  await expect(editRawActivity).toBeFocused();
  await editRawActivity.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(semanticSummaryCell).toBeFocused();

  await openA11ySystemSurface(page, {
    groupToken: "coordination",
    viewSchemaId: taskRequestsViewSchemaId,
  });
  const taskTitle = await mountedGridCell(
    page,
    taskRequestsViewSchemaId,
    taskRow.record_id,
    "task.title",
  );
  await expectCellTextOrValue(taskTitle, "a11y.design-readiness task row");
  await expectVisibleSemanticGridCellFocus(taskTitle);
  await openFilterPopover(page, taskRequestsViewSchemaId);
  await expectAndRecordContrast(
    page,
    [gridFilterFieldTestId(taskRequestsViewSchemaId)],
    "filter-popover",
  );
  await expect(
    page.getByTestId(gridFilterFieldTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridGroupingSelectTestId(taskRequestsViewSchemaId)),
  ).toBeVisible();

  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expectVisibleFocus(trigger);
  await expect(trigger).toHaveAttribute("aria-expanded", "false");
  await trigger.press("Enter");
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(trigger).toHaveAttribute("aria-expanded", "true");
  await expect(menu).toHaveAttribute("role", "menu");
  const indicatorsOption = page.getByTestId(
    systemViewSwitcherOptionTestId("scope-indicators", indicatorsViewSchemaId),
  );
  await expect(indicatorsOption).toHaveAttribute("role", "menuitemradio");
  await expect(indicatorsOption).toHaveAttribute("aria-checked", "false");
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveCount(
    1,
  );
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
  await expect(trigger).toHaveAttribute("aria-expanded", "false");

  await expectAllInteractiveControlsNamed(page);
  await expectInitialTabRoundTrip(page);
  await expectAndRecordContrast(page, [
    systemViewSwitcherTriggerTestId(),
    savedViewSelectorTestId(taskRequestsViewSchemaId),
    gridGroupingSelectTestId(taskRequestsViewSchemaId),
    rowCellTestId(taskRow.record_id, "task.title"),
    saveStateTestId(),
  ]);

  // Non-color-only states: each grid state carries a role, text and action.
  await page.keyboard.press("Escape");
  // The fixture task is "normal" priority, so an "urgent" filter matches nothing.
  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.priority",
    "urgent",
  );
  const emptyState = page.locator('[data-grid-data-state="filtered_empty"]');
  await expect(emptyState.getByRole("status")).toHaveText(
    "No rows match the current filters.",
  );
  const failQuery = async (
    route: Parameters<Parameters<typeof installOwnedRoute>[2]>[0],
  ) => {
    await route.fulfill({
      status: 503,
      json: {
        error: {
          code: "service_unavailable",
          details: {},
          message: "Service unavailable.",
          request_id: "a11y-design-readiness",
          retryable: true,
        },
      },
    });
  };
  // Error: a failed refresh keeps the loaded rows and says so in text.
  const taskQuery = `**/api/v1/incidents/${incidentId}/views/${taskRequestsViewSchemaId}/query`;
  await installOwnedRoute(page, taskQuery, failQuery);
  await emptyState
    .getByRole("button", { name: "Clear filters", exact: true })
    .click();
  const errorState = page.locator('[data-grid-data-state="stale_error"]');
  await expect(errorState).toHaveAttribute("data-grid-blocking", "false");
  await expect(errorState.getByRole("alert")).toHaveText(
    "Workbook view load failed. Previously loaded rows may be stale.",
  );
  await removeOwnedRoute(page, taskQuery);
  await errorState.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(errorState).toHaveCount(0);
  await expect(
    page.getByTestId(rowCellTestId(taskRow.record_id, "task.title")),
  ).toBeVisible();

  // Blocked: a surface whose first read fails blocks its body in text.
  const decisionQuery = `**/api/v1/incidents/${incidentId}/views/${decisionsViewSchemaId}/query`;
  await installOwnedRoute(page, decisionQuery, failQuery);
  await openA11ySystemSurface(page, {
    groupToken: "coordination",
    viewSchemaId: decisionsViewSchemaId,
  });
  const blockedState = page.locator('[data-grid-data-state="unavailable"]');
  await expect(blockedState).toHaveAttribute("data-grid-blocking", "true");
  await expect(blockedState.getByRole("alert")).toHaveText(/\S/);
  const retry = blockedState.getByRole("button", {
    name: "Retry",
    exact: true,
  });
  await expect(retry).toBeVisible();
  await removeOwnedRoute(page, decisionQuery);

  // Loading: a held read is announced in text until it settles.
  const held = responseBarrier(page);
  await installOwnedRoute(page, decisionQuery, async (route) => {
    if (!(await awaitResponseBarrier(route, held.promise))) return;
    await route.continue();
  });
  await retry.click();
  const loadingState = page.locator(
    '[data-grid-data-state="initial_loading"], [data-grid-data-state="refreshing"]',
  );
  await expect(loadingState.getByRole("status")).toHaveText(/\S/);
  held.release();
  await expect(loadingState).toHaveCount(0);
  await removeOwnedRoute(page, decisionQuery);
});
