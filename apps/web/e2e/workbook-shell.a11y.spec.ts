import {
  dataTestIdSelector,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  rowCellTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorTestId,
  workbookFilterPopoverTestId,
  workbookFocusAnchorTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  notesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleSemanticGridCellFocus } from "./support/accessibility/grid";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectWorkbookSavePresentation } from "./support/accessibility/status";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  mountedGridCell,
  semanticGridCell,
} from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { observeQuery, seed } from "./support/workbook/queryBrowsing";
import { openFilterPopover } from "./support/workbook/queryControls";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import { observeSaveEvents, saveEvents } from "./support/workbook/saveStatus";
import { expectIncidentHeadingAndHeaders } from "./support/workbook/shellAccessibility";

test("a11y.workbook-shell.row-01 Verify incident heading, content-backed headers, shell regions, tabs, switchers, menus, inspector controls, and status strip retain names and keyboard access.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YWORKBOOKSHELL"),
    "a11y.workbook-shell.row-01 workbook shell",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.workbook-shell-01-row"),
      "timeline.activity_utc_text": "2026-05-31T09:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.workbook-shell accessibility shell row",
      "timeline.raw_activity_text": "Inspector control coverage",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Workbook shell" }),
  ).toHaveCount(1);

  const statusStrip = shell.getByRole("region", {
    name: workbookShellSlotLabel("status-strip"),
    exact: true,
  });
  const diagnostic = statusStrip.getByTestId(workbookFocusAnchorTestId());
  await expect(diagnostic).toHaveText("cleared");
  await expect(statusStrip.getByTestId(saveStateTestId())).toHaveText("Saved");
  const initialStatus = await statusStrip.ariaSnapshot();
  expect(initialStatus).toContain("Saved");
  expect.soft(initialStatus).not.toContain("cleared");
  await observeSaveEvents(page);

  await page.setViewportSize({ width: 1024, height: 720 });
  const activityCell = await expectVisibleSemanticGridCellFocus(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_utc_text",
    ),
  );
  const activityAnchor = `${timelineViewSchemaId}:${timelineRow.record_id}:timeline.activity_utc_text`;
  await expect(diagnostic).toHaveText(activityAnchor);
  const activeStatus = await statusStrip.ariaSnapshot();
  expect(activeStatus).toContain("Saved");
  expect.soft(activeStatus).not.toContain(activityAnchor);

  await activityCell.press("ArrowRight");
  const localDateCell = semanticGridCell(
    page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.activity_local_text"),
    ),
  );
  await expectVisibleFocus(localDateCell);
  const movedAnchor = `${timelineViewSchemaId}:${timelineRow.record_id}:timeline.activity_local_text`;
  await expect(diagnostic).toHaveText(movedAnchor);
  const movedStatus = await statusStrip.ariaSnapshot();
  expect(movedStatus).toContain("Saved");
  expect.soft(movedStatus).not.toContain(movedAnchor);

  await page.setViewportSize({ width: 1440, height: 900 });
  const hostsTab = page.getByTestId(surfaceTabTestId(hostsViewSchemaId));
  await hostsTab.focus();
  await hostsTab.press("Enter");
  await expect(hostsTab).toHaveAttribute("aria-selected", "true");
  await expect(diagnostic).toHaveText("cleared");
  const hostsStatus = await statusStrip.ariaSnapshot();
  expect(hostsStatus).toContain("Saved");
  expect.soft(hostsStatus).not.toContain("cleared");
  const timelineTab = page.getByTestId(surfaceTabTestId(timelineViewSchemaId));
  await timelineTab.focus();
  await timelineTab.press("Enter");
  await expect(timelineTab).toHaveAttribute("aria-selected", "true");
  expect(await saveEvents(page)).toEqual([]);

  await expectIncidentHeadingAndHeaders(page, "Row");
  const identity = page.getByRole("button", { name: /^Incident details:/ });
  await expect(identity).toHaveAccessibleName(
    /a11y.workbook-shell.row-01 workbook shell/,
  );
  await expectVisibleFocus(identity);
  await identity.press("Enter");
  const identityDetails = page.getByRole("region", {
    name: "Incident identity",
    exact: true,
  });
  await expect(identityDetails).toContainText(
    "a11y.workbook-shell.row-01 workbook shell",
  );
  await expect(identityDetails.locator("strong")).toHaveText(
    await identity.locator("strong").innerText(),
  );
  await identity.press("Escape");
  await expect(identityDetails).toHaveCount(0);
  await expect(identity).toBeFocused();
  await identity.click();
  await expect(identityDetails).toBeVisible();
  await page.getByTestId(systemViewSwitcherTriggerTestId()).focus();
  await expect(identityDetails).toHaveCount(0);

  for (const viewport of [
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await expectIncidentHeadingAndHeaders(page, "Row");
    await expectVisibleFocus(identity);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollHeight <= window.innerHeight + 1,
      ),
    ).toBe(true);
  }
  await page.setViewportSize({ width: 2048, height: 1440 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await withPresentation(
      page,
      {
        textSpacing:
          "h1, h1 * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; }",
      },
      async () => {
        await expectIncidentHeadingAndHeaders(page, "Row");
        await expectVisibleFocus(identity);
        await identity.press("Enter");
        await expect(identityDetails).toBeVisible();
        await identity.press("Escape");
        await expect(identity).toBeFocused();
      },
    );
  });
  await page.setViewportSize({ width: 1440, height: 900 });

  for (const slot of workbookShellSlots.filter(
    (slot) => slot !== "inspector",
  )) {
    const label = workbookShellSlotLabel(slot);
    const slotByTestId = shell.locator(
      dataTestIdSelector(workbookShellSlotTestId(slot)),
    );
    await expect(slotByTestId).toBeVisible();
    await expect(slotByTestId).toHaveAttribute("aria-label", label);
    await expect(shell.getByRole("region", { name: label })).toHaveCount(1);
  }
  await expect(
    shell.locator(dataTestIdSelector(workbookShellSlotTestId("inspector"))),
  ).toHaveCount(0);
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  const inspectorSlot = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("inspector")),
  );
  await expect(inspectorSlot).toBeVisible();
  await expect(inspectorSlot).toHaveAttribute(
    "aria-label",
    workbookShellSlotLabel("inspector"),
  );
  await expect(
    shell.getByRole("region", { name: workbookShellSlotLabel("inspector") }),
  ).toHaveCount(1);

  for (const surface of requiredBuiltInWorkbookSurfaceIds) {
    const tab = page.getByTestId(surfaceTabTestId(surface));
    await expect(tab).toBeVisible();
    await expect(tab).not.toHaveText("");
  }
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");

  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expect(trigger).toBeVisible();
  await expect(trigger).toHaveAttribute("aria-label", "More views");
  await expectVisibleFocus(trigger);
  await page.keyboard.press("Enter");

  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(menu).toHaveAttribute("role", "menu");
  const indicatorOption = page.getByTestId(
    systemViewSwitcherOptionTestId("scope-indicators", indicatorsViewSchemaId),
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
  await openFilterPopover(page, timelineViewSchemaId);
  await expectAndRecordContrast(
    page,
    [gridFilterApplyTestId(timelineViewSchemaId)],
    "filter-popover",
  );
  await expect(
    page.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridFilterApplyTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await page
    .getByTestId(workbookFilterPopoverTestId(timelineViewSchemaId))
    .getByRole("button", { name: "Cancel" })
    .click();
  await expect(
    page.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
  ).toHaveCount(0);

  const summaryCell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(summaryCell).toBeVisible();
  await expectVisibleSemanticGridCellFocus(summaryCell);
  await openTimelineInspector(page, timelineRow.record_id);

  const inspector = page.getByTestId(timelineInspectorTestId());
  await expect(inspector).toBeVisible();
  await expect(inspector).toHaveAttribute("aria-label", "Timeline inspector");
  await expect(
    inspector.locator(
      '[data-inspector-saved-field="timeline.raw_activity_text"]',
    ),
  ).toBeVisible();

  const saveState = page.getByTestId(saveStateTestId());
  await expectWorkbookSavePresentation(page);
  await expect(saveState).toHaveText("Saved");

  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    workbookShellSlotTestId("top-bar"),
    workbookShellSlotTestId("status-strip"),
    systemViewSwitcherTriggerTestId(),
    savedViewSelectorTestId(timelineViewSchemaId),
    rowCellTestId(timelineRow.record_id, "timeline.activity_synopsis_text"),
    saveStateTestId(),
  ]);
});

test("a11y.workbook query controls preserve keyboard refresh at narrow width and 200 percent enlargement", async ({
  page,
  testActor,
}) => {
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-PRESENTATION"),
    "Query control layout",
  );
  await seed(page, incident, view, 101, testActor.user_id);
  const reads = await observeQuery(page, incident, view);
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  for (const profile of [
    { viewport: { width: 360, height: 720 } },
    { viewport: { width: 720, height: 1280 }, zoom: "200%" },
  ]) {
    await withPresentation(page, profile, async () => {
      const refresh = controls.getByRole("button", {
        name: "Refresh",
        exact: true,
      });
      await refresh.scrollIntoViewIfNeeded();
      await refresh.focus();
      await refresh.press("Enter");
      await expect(controls).toContainText(
        "100 records loaded; more available.",
      );
      expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
      await expect(refresh).toBeFocused();
      const rect = await controls.boundingBox();
      if (!rect) throw new Error("Browsing controls are missing");
      expect(rect.x).toBeGreaterThanOrEqual(0);
      expect(rect.x + rect.width).toBeLessThanOrEqual(profile.viewport.width);
    });
  }
});
