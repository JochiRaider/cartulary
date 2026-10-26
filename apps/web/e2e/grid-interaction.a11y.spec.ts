import { Buffer } from "node:buffer";
// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import {
  gridFillHandleSelector,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridSortHeaderTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Request } from "@playwright/test";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  activateTimelineGridEditor,
  expectVisibleSemanticGridCellFocus,
} from "./support/accessibility/grid";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectWorkbookSavePresentation } from "./support/accessibility/status";
import { openIncidentAsTrackedUserReady } from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  mountedGridCell,
  mountedGridTarget,
} from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { expectIncidentHeadingAndHeaders } from "./support/workbook/shellAccessibility";

test("a11y.grid editing and grouped navigation preserve keyboard semantics", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YGRIDINTERACTION"),
    "a11y.grid-interaction.row-01 grid adapter",
  );
  const alphaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.grid-interaction-01-alpha"),
    "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
    "timeline.activity_synopsis_text": "Alpha accessibility row",
    "timeline.raw_activity_text": "Keyboard grid coverage",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.grid-interaction-01-beta"),
    "timeline.activity_utc_text": "2026-05-31T10:05:00Z",
    "timeline.activity_synopsis_text": "Beta accessibility row",
    "timeline.raw_activity_text": "Grouped grid coverage",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      alphaRow.record_id,
      "timeline.activity_synopsis_text",
    ),
  ).toHaveText("Alpha accessibility row");
  const grid = page.locator(gridScrollportSelector());
  await expectIncidentHeadingAndHeaders(page, "Row");
  const fillLabel = "Drag to fill this value";
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Accessibility.enable");
  const inspectFillAccessibility = async (state: string) => {
    const result = await new AxeBuilder({ page })
      .include(gridScrollportSelector())
      .withRules(["aria-required-children"])
      .analyze();
    const handle = page.locator(gridFillHandleSelector());
    const accessibleCells = (
      await cdp.send("Accessibility.getFullAXTree")
    ).nodes
      .filter(
        (node) =>
          node.role?.value === "gridcell" &&
          node.description?.value?.includes(fillLabel),
      )
      .map((node) => ({
        backendDOMNodeId: node.backendDOMNodeId,
        description: node.description?.value,
        name: node.name?.value,
      }));
    const observation = {
      state,
      accessibilitySnapshot: await grid.ariaSnapshot(),
      accessibleCells,
      handle:
        (await handle.count()) === 0
          ? null
          : await handle.first().evaluate((element) => ({
              html: element.outerHTML,
              ownerRole: element.parentElement?.getAttribute("role"),
            })),
      violations: result.violations.filter(
        (violation) => violation.id === "aria-required-children",
      ),
    };
    await testInfo.attach(`timeline-fill-${state}`, {
      body: Buffer.from(JSON.stringify(observation, null, 2)),
      contentType: "application/json",
    });
    return observation;
  };
  const alphaSummaryControl = await mountedGridCell(
    page,
    timelineViewSchemaId,
    alphaRow.record_id,
    "timeline.activity_synopsis_text",
  );
  const alphaCell =
    await expectVisibleSemanticGridCellFocus(alphaSummaryControl);
  const navigation = await inspectFillAccessibility("navigation");
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(1);
  await expect(alphaCell).toHaveAttribute(
    "aria-description",
    /Drag to fill this value/,
  );
  const alphaEditor = await activateTimelineGridEditor(
    page,
    alphaRow.record_id,
    "timeline.activity_synopsis_text",
  );
  const editing = await inspectFillAccessibility("editing");
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(0);
  await alphaEditor.press("Escape");
  await expect(alphaCell).toBeFocused();
  const restored = await inspectFillAccessibility("restored");
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(1);
  const betaCell = await expectVisibleSemanticGridCellFocus(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      betaRow.record_id,
      "timeline.activity_synopsis_text",
    ),
  );
  const moved = await inspectFillAccessibility("moved");
  await expect(betaCell).toHaveAttribute(
    "aria-description",
    /Drag to fill this value/,
  );
  await expect(alphaCell).not.toHaveAttribute(
    "aria-description",
    /Drag to fill this value/,
  );
  const betaMarkReviewed = page.getByTestId(
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  const betaSummaryControl = await mountedGridCell(
    page,
    timelineViewSchemaId,
    betaRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(betaSummaryControl);
  await page.keyboard.press("Shift+F10");
  await expectVisibleFocus(betaMarkReviewed);
  await betaMarkReviewed.click();
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      betaRow.record_id,
      "timeline.capture_state",
    ),
  ).toHaveText("reviewed");
  await mountedGridCell(
    page,
    timelineViewSchemaId,
    betaRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(betaSummaryControl);
  await page.keyboard.press("Shift+F10");
  await expect(
    page.getByTestId(timelineRowMarkReviewedButtonTestId(betaRow.record_id)),
  ).toBeDisabled();
  await page.keyboard.press("Escape");
  await page
    .getByTestId(gridGroupingSelectTestId(timelineViewSchemaId))
    .selectOption("timeline.capture_state");
  const reviewedGroup = page.getByTestId(
    gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    ),
  );
  await expect(reviewedGroup).toBeVisible();
  await expectAndRecordContrast(
    page,
    [
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    ],
    "reviewed-group",
  );
  await expect(reviewedGroup).toContainText("reviewed");
  const grouped = await inspectFillAccessibility("grouped");
  await expect(page.locator(gridFillHandleSelector())).toHaveCount(0);
  const groupWrites: string[] = [];
  const recordGroupWrite = (request: Request) => {
    if (
      ["POST", "PUT", "PATCH", "DELETE"].includes(request.method()) &&
      new URL(request.url()).pathname.startsWith("/api/")
    )
      groupWrites.push(
        `${request.method()} ${new URL(request.url()).pathname}`,
      );
  };
  page.on("request", recordGroupWrite);
  const roughGroup = page.getByTestId(
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  );
  await expect(roughGroup).toBeVisible();
  const inspectGroupFocus = async (state: string) => {
    const axe = await new AxeBuilder({ page })
      .include(gridScrollportSelector())
      .withRules(["aria-required-children", "aria-hidden-focus"])
      .analyze();
    const active = await cdp.send("Runtime.evaluate", {
      expression: "document.activeElement",
      returnByValue: false,
    });
    const activeAx = active.result.objectId
      ? await cdp.send("Accessibility.getPartialAXTree", {
          objectId: active.result.objectId,
          fetchRelatives: true,
        })
      : null;
    const activeNode = activeAx?.nodes[0];
    const observation = {
      state,
      activeAccessible: activeNode
        ? {
            role: activeNode.role?.value,
            name: activeNode.name?.value,
            properties: activeNode.properties?.map((property) => ({
              name: property.name,
              value: property.value?.value,
            })),
          }
        : null,
      dom: await grid.evaluate((element) => {
        const focus = document.activeElement;
        const focusedGroupRow = element.contains(focus)
          ? focus?.closest('[role="row"][aria-level="1"]')
          : null;
        const sink = element.querySelector(
          ":scope > div:not([role])[tabindex]",
        );
        return {
          focusHtml: focus?.outerHTML ?? null,
          focusInsideGrid: focus === element || element.contains(focus),
          focusRole: focus?.getAttribute("role") ?? null,
          focusedGroupRowHtml: focusedGroupRow?.outerHTML ?? null,
          sinkHtml: sink?.outerHTML ?? null,
          toggleCount:
            focusedGroupRow?.querySelectorAll(
              "button[data-cartulary-grid-group-id]",
            ).length ?? 0,
        };
      }),
      accessibilitySnapshot: await grid.ariaSnapshot(),
      violations: axe.violations.filter((violation) =>
        ["aria-required-children", "aria-hidden-focus"].includes(violation.id),
      ),
    };
    await testInfo.attach(`timeline-group-focus-${state}`, {
      body: Buffer.from(JSON.stringify(observation, null, 2)),
      contentType: "application/json",
    });
    return observation;
  };
  const initialGroupFocus = await inspectGroupFocus("initial");
  expect(initialGroupFocus.dom.toggleCount).toBe(0);
  await reviewedGroup.click();
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
  const pointerGroupFocus = await inspectGroupFocus("pointer-collapse");
  expect(pointerGroupFocus.dom.toggleCount).toBe(1);
  expect(pointerGroupFocus.dom.focusHtml).toContain("<button");
  expect(pointerGroupFocus.activeAccessible?.name).toBe("reviewed");
  await page.keyboard.press("Enter");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
  const enterGroupFocus = await inspectGroupFocus("button-enter-expand");
  await page.keyboard.press("Space");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
  const spaceGroupFocus = await inspectGroupFocus("button-space-collapse");
  await page.keyboard.press("ArrowRight");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
  const rightGroupFocus = await inspectGroupFocus("right-expand");
  await page.keyboard.press("ArrowLeft");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
  const leftGroupFocus = await inspectGroupFocus("left-collapse");
  await page.keyboard.press("Shift+Tab");
  const regionExit = await inspectGroupFocus("group-shift-tab-exit");
  expect(regionExit.dom.focusInsideGrid).toBe(false);
  await reviewedGroup.click();
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
  const returnGroupFocus = await inspectGroupFocus("pointer-after-region-exit");
  await expect(reviewedGroup).toBeFocused();
  await page.keyboard.press("Tab");
  const forwardTab = await inspectGroupFocus("tab");
  expect(forwardTab.dom.focusInsideGrid).toBe(false);
  await page.keyboard.press("Shift+Tab");
  const backwardTab = await inspectGroupFocus("shift-tab");
  expect(backwardTab.dom.focusInsideGrid).toBe(true);
  for (const observation of [
    initialGroupFocus,
    pointerGroupFocus,
    enterGroupFocus,
    spaceGroupFocus,
    rightGroupFocus,
    leftGroupFocus,
    returnGroupFocus,
    forwardTab,
    backwardTab,
    regionExit,
  ]) {
    expect(observation.violations, observation.state).toEqual([]);
  }
  page.off("request", recordGroupWrite);
  expect(groupWrites).toEqual([]);
  for (const observation of [
    pointerGroupFocus,
    enterGroupFocus,
    spaceGroupFocus,
    rightGroupFocus,
    leftGroupFocus,
    returnGroupFocus,
  ]) {
    expect(observation.dom.focusHtml, observation.state).toContain("<button");
    expect(observation.activeAccessible?.role, observation.state).toBe(
      "button",
    );
    expect(observation.activeAccessible?.name, observation.state).toBe(
      "reviewed",
    );
  }
  for (const observation of [navigation, editing, restored, moved]) {
    expect(observation.violations, observation.state).toEqual([]);
  }
  expect(navigation.accessibleCells).toHaveLength(1);
  expect(editing.accessibleCells).toHaveLength(0);
  expect(restored.accessibleCells).toHaveLength(1);
  expect(moved.accessibleCells).toHaveLength(1);
  expect(grouped.accessibleCells).toHaveLength(0);
  await expect(alphaCell).not.toHaveAttribute(
    "aria-description",
    new RegExp(fillLabel),
  );
  await expect(betaCell).not.toHaveAttribute(
    "aria-description",
    new RegExp(fillLabel),
  );
  const betaSummary = await activateTimelineGridEditor(
    page,
    betaRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(betaSummary).toHaveAttribute(
    "aria-label",
    `Activity Synopsis ${betaRow.record_id}`,
  );
  await betaSummary.fill("Beta accessibility active edit");
  await expect(betaSummary).toHaveValue("Beta accessibility active edit");
  await expectWorkbookSavePresentation(page);
  await expect(
    await mountedGridTarget(
      page,
      timelineViewSchemaId,
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
    ),
  ).toContainText("Activity Synopsis");
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    gridGroupingSelectTestId(timelineViewSchemaId),
    gridSortHeaderTestId(
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    ),
    rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
    saveStateTestId(),
  ]);
});

test("a11y.viewer grid grouping remains read only and keyboard operable", async ({
  browserScope,
  page,
  sessionTracker,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YGRIDINTERACTION"),
    "a11y.grid-interaction.row-01 grid adapter",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.grid-interaction-01-alpha"),
    "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
    "timeline.activity_synopsis_text": "Alpha accessibility row",
    "timeline.raw_activity_text": "Keyboard grid coverage",
  });
  const betaRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.grid-interaction-01-beta"),
    "timeline.activity_utc_text": "2026-05-31T10:05:00Z",
    "timeline.activity_synopsis_text": "Beta accessibility row",
    "timeline.raw_activity_text": "Grouped grid coverage",
  });
  const viewerPassword = "A11yGroupedViewer1!";
  const viewer = await createIncidentMemberUser(page, incidentId, {
    display_name: "A11y Grouped Viewer",
    email: uniqueEmail("a11y-grouped-viewer"),
    initial_password: viewerPassword,
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await clickTimelineRowAction(
    page,
    betaRow.record_id,
    timelineRowMarkReviewedButtonTestId(betaRow.record_id),
  );
  const current = (
    await queryViewRows(page, incidentId, timelineViewSchemaId)
  ).find((row) => row.record_id === betaRow.record_id);
  if (!current) throw new Error("Missing viewer grid fixture row");
  await patchRecord(page, betaRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: current.row_version,
    client_txn_id: uniqueTxn("viewer-enriched-row"),
    changes: [
      {
        field_key: "timeline.activity_synopsis_text",
        value: "Beta accessibility active edit",
      },
    ],
  });
  const viewerSession = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      createdBy: "a11y.grid-interaction.row-01",
      email: viewer.email,
      incidentId,
      password: viewerPassword,
      purpose: "grouped Timeline viewer accessibility",
      readyRecordId: betaRow.record_id,
      userId: viewer.user_id,
    },
  );
  {
    const viewerPage = viewerSession.page;
    await viewerPage
      .getByTestId(gridGroupingSelectTestId(timelineViewSchemaId))
      .selectOption("timeline.capture_state");
    const viewerGrid = viewerPage.locator(gridScrollportSelector());
    await expect(viewerGrid).toHaveAttribute("role", "treegrid");
    await expect(viewerGrid).toHaveAttribute("aria-readonly", "true");
    await expectIncidentHeadingAndHeaders(viewerPage, "Capture State");
    const viewerInitialAxe = await new AxeBuilder({ page: viewerPage })
      .withRules(["aria-required-children", "aria-hidden-focus"])
      .analyze();
    await testInfo.attach("timeline-group-focus-viewer-initial", {
      body: Buffer.from(
        JSON.stringify(
          {
            focused: await viewerPage.evaluate(() => ({
              tag: document.activeElement?.tagName ?? null,
              role: document.activeElement?.getAttribute("role") ?? null,
              name: document.activeElement?.getAttribute("aria-label") ?? null,
            })),
            rootChildren: await viewerGrid.evaluate((element) =>
              Array.from(element.children).map((child) => ({
                tag: child.tagName,
                role: child.getAttribute("role"),
                ariaHidden: child.getAttribute("aria-hidden"),
                tabIndex: child.getAttribute("tabindex"),
                className: child.className,
                style: child.getAttribute("style"),
              })),
            ),
            violations: viewerInitialAxe.violations,
          },
          null,
          2,
        ),
      ),
      contentType: "application/json",
    });
    expect(viewerInitialAxe.violations).toEqual([]);
    const viewerGroup = viewerPage.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "enriched",
      ),
    );
    await expect(viewerGroup).toBeVisible();
    await viewerGroup.click();
    await expect(viewerGroup).toBeFocused();
    await expect(viewerGroup).toHaveAttribute("aria-expanded", "false");
    await viewerPage.keyboard.press("Enter");
    await expect(viewerGroup).toHaveAttribute("aria-expanded", "true");
    const viewerAxe = await new AxeBuilder({ page: viewerPage })
      .include(gridScrollportSelector())
      .withRules(["aria-required-children", "aria-hidden-focus"])
      .analyze();
    expect(viewerAxe.violations).toEqual([]);
    await testInfo.attach("timeline-group-focus-viewer", {
      body: Buffer.from(
        JSON.stringify(
          {
            accessibilitySnapshot: await viewerGrid.ariaSnapshot(),
            focused: await viewerPage.evaluate(
              () => document.activeElement?.outerHTML ?? null,
            ),
            violations: viewerAxe.violations,
          },
          null,
          2,
        ),
      ),
      contentType: "application/json",
    });

    await browserScope.closeContext(viewerSession.page.context());
  }
});

import { patchRecord, queryViewRows } from "./support/workbook/query";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";
