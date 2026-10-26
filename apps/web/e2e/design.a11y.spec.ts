import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
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
  expectNoFocusTrap,
  expectTabOrderIncludes,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { openA11ySystemSurface } from "./support/accessibility/queryControls";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectWorkbookSavePresentation } from "./support/accessibility/status";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
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

  await expectTabOrderIncludes(page, [
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
  await trigger.press("Enter");
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(menu).toHaveAttribute("role", "menu");
  await expect(
    page.getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        indicatorsViewSchemaId,
      ),
    ),
  ).toHaveAttribute("role", "menuitemradio");
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();

  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    systemViewSwitcherTriggerTestId(),
    savedViewSelectorTestId(taskRequestsViewSchemaId),
    gridGroupingSelectTestId(taskRequestsViewSchemaId),
    rowCellTestId(taskRow.record_id, "task.title"),
    saveStateTestId(),
  ]);
});
