import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  accountTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  genericCreateFieldTestId,
  gridShellTestId,
  incidentAdministrationTestId,
  incidentControlsActionMessageTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRoleSelectTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  rowCellTestId,
  savedViewFamilySelector,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { openIncidentControls } from "./pages/deploymentAdministration";
import {
  IncidentDirectory,
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
} from "./pages/incidentDirectory";
import {
  accountResponseGate,
  installAccountEditingFixture,
} from "./support/auth/accountEditingFixture";
import { csrfHeaders } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow } from "./support/workbook/query";
import {
  createSavedView,
  seedSystemSavedView,
  setCurrentSavedViewAsDefault,
  setCurrentSavedViewAsHome,
} from "./support/workbook/savedViews";

type AccountDensityMode = "compact" | "default" | "comfortable" | null;

test("operates account application menus with keyboard focus and viewport containment", async ({
  workerAdminPage: page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("menu"),
    "Account menu navigation",
  );
  await openIncidentFromLanding(page, incidentId);
  const trigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
  const incidents = page.getByRole("menuitem", {
    name: "Incidents",
    exact: true,
  });
  const settings = page.getByRole("menuitem", {
    name: "Account settings",
    exact: true,
  });
  const controls = page.getByTestId(incidentControlsTriggerTestId());
  const summary = page.getByTestId(incidentControlsMenuItemTestId("summary"));
  const menu = page.getByTestId(accountTestId("application-menu"));
  const inspector = page.getByTestId(workbookShellSlotTestId("inspector"));
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(inspector).toBeVisible();
  await trigger.focus();
  await page.keyboard.down("Enter");
  await page.keyboard.down("Enter");
  await page.keyboard.up("Enter");
  await expect(incidents).toBeFocused();
  await page.keyboard.press("End");
  await expect(settings).toBeFocused();
  await page.keyboard.press("ArrowUp");
  await expect(controls).toBeFocused();
  await page.keyboard.press("ArrowRight");
  await expect(summary).toBeFocused();
  await page.keyboard.press("End");
  await expect(
    page.getByTestId(incidentControlsMenuItemTestId("membership-audit")),
  ).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await expect(summary).toBeFocused();
  await page.keyboard.press("ArrowLeft");
  await expect(controls).toBeFocused();
  await expect(page.getByTestId(incidentControlsMenuTestId())).toHaveCount(0);
  await page.keyboard.press("Space");
  await expect(summary).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(controls).toBeFocused();
  await expect(inspector).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
  await expect(menu).toHaveCount(0);
  await expect(inspector).toBeVisible();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();

  await trigger.focus();
  await page.keyboard.press("ArrowUp");
  await expect(settings).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(menu).toHaveCount(0);
  await expect(trigger).not.toBeFocused();
  await trigger.focus();
  await page.keyboard.press("ArrowDown");
  await expect(incidents).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(trigger).toBeFocused();
  await expect(menu).toHaveCount(0);

  await trigger.click();
  await controls.click();
  await expect(summary).toBeFocused();
  await page.keyboard.press("Space");
  const closeControls = page.getByRole("button", {
    name: "Close incident controls",
  });
  await expect(closeControls).toBeFocused();
  await expect(menu).toHaveCount(0);
  await expect(page.getByTestId(incidentControlsPanelTestId())).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();

  await trigger.click();
  await page.keyboard.press("Home");
  await page.keyboard.press("Enter");
  await expect(page.getByTestId(landingAdminShellTestId("heading"))).toHaveText(
    "Incident directory",
  );
  await expect(
    page.getByTestId(landingAdminShellTestId("heading")),
  ).toBeFocused();
  await trigger.click();
  await expect(incidents).toHaveAttribute("aria-current", "page");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("Enter");
  await expect(page.getByTestId(landingAdminShellTestId("heading"))).toHaveText(
    "Deployment administration",
  );
  await expect(
    page.getByTestId(landingAdminShellTestId("heading")),
  ).toBeFocused();
  await trigger.click();
  await expect(
    page.getByRole("menuitem", {
      name: "Deployment administration",
      exact: true,
    }),
  ).toBeFocused();
  await page.keyboard.press("End");
  await page.keyboard.press("Space");
  const dialog = page.getByRole("dialog", { name: "Account settings" });
  await expect(
    dialog.getByRole("button", { name: "Close", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await trigger.click();
  await page.getByText("Cartulary", { exact: true }).click();
  await expect(menu).toHaveCount(0);
  await expect(trigger).not.toBeFocused();

  // Long display labels are ordinary profile input; authorization is unchanged.
  await trigger.click();
  await settings.click();
  const displayName = page.getByTestId(accountTestId("profile-display-name"));
  await expect(
    page.getByTestId(accountTestId("profile-display-name")),
  ).toBeEnabled();
  await displayName.fill("Analyst".repeat(36));
  await page.getByTestId(accountTestId("profile-save")).click();
  await expect(trigger).toHaveAttribute("title", "Analyst".repeat(36));
  await dialog.getByRole("button", { name: "Close", exact: true }).click();
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 1280, height: 400 },
    { width: 768, height: 400 },
    { width: 640, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await trigger.focus();
    await page.keyboard.press("Enter");
    await expectAccountMenuContained(page);
    await controls.click();
    await page.keyboard.press("End");
    await expectAccountMenuContained(page);
    await page.keyboard.press("Home");
    await expectAccountMenuContained(page);
    await page.keyboard.press("Escape");
    await page.keyboard.press("Escape");
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await trigger.click();
  await controls.click();
  await page.keyboard.press("End");
  await expectAccountMenuContained(page);
  await page.keyboard.press("Escape");
  await page.keyboard.press("Escape");
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  const spacing = await page.addStyleTag({
    content:
      "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await trigger.click();
  await controls.click();
  await page.keyboard.press("End");
  await expectAccountMenuContained(page);
  await testInfo.attach("account-menu-open", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await spacing.evaluate((element) => element.parentNode?.removeChild(element));
  await page.keyboard.press("Escape");
  await page.keyboard.press("Home");
  await page.keyboard.press("Enter");
  await page.setViewportSize({ width: 640, height: 480 });
  for (const destination of [
    "Incident directory",
    "Deployment administration",
  ]) {
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toHaveText(destination);
    await expect(
      page.getByTestId(landingAdminShellTestId("heading")),
    ).toBeFocused();
    await trigger.focus();
    await page.keyboard.press("Enter");
    await expectAccountMenuContained(page);
    await page.keyboard.press("End");
    await expectAccountMenuContained(page);
    if (destination === "Incident directory") {
      await page.keyboard.press("Home");
      await page.keyboard.press("ArrowDown");
      await page.keyboard.press("Enter");
    } else {
      await page.keyboard.press("Escape");
      await expect(trigger).toBeFocused();
    }
  }
});

async function expectAccountMenuContained(page: Page) {
  const menu = page.getByTestId(accountTestId("application-menu"));
  await expect
    .poll(() =>
      menu.evaluate((panel) => {
        const rect = panel.getBoundingClientRect();
        const active = document.activeElement;
        if (!(active instanceof HTMLElement) || !panel.contains(active))
          return false;
        const focus = active.getBoundingClientRect();
        const style = getComputedStyle(active);
        return (
          rect.left >= 0 &&
          rect.top >= 0 &&
          rect.right <= window.innerWidth + 1 &&
          rect.bottom <= window.innerHeight + 1 &&
          focus.top >= rect.top &&
          focus.bottom <= rect.bottom &&
          focus.left >= rect.left &&
          focus.right <= rect.right &&
          style.outlineStyle !== "none" &&
          Number.parseFloat(style.outlineWidth) > 0 &&
          window.scrollY === 0
        );
      }),
    )
    .toBe(true);
}

type AccountPreferencesResource = {
  density_mode: AccountDensityMode;
  preferences_version: number;
};

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByRole("button", {
    name: "Account and application navigation",
  });
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}

async function readWorkbookLayoutRects(page: Page) {
  return page.evaluate(
    ({ gridSelector, primaryGridSelector, statusStripSelector }) => {
      const roundRect = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          throw new Error(`Expected ${selector} to exist`);
        }
        const rect = element.getBoundingClientRect();
        return {
          bottom: Math.round(rect.bottom),
          height: Math.round(rect.height),
          left: Math.round(rect.left),
          right: Math.round(rect.right),
          top: Math.round(rect.top),
          width: Math.round(rect.width),
        };
      };
      return {
        grid: roundRect(gridSelector),
        primaryGrid: roundRect(primaryGridSelector),
        statusStrip: roundRect(statusStripSelector),
      };
    },
    {
      gridSelector: dataTestIdSelector(gridShellTestId(timelineViewSchemaId)),
      primaryGridSelector: dataTestIdSelector(
        workbookShellSlotTestId("primary-grid"),
      ),
      statusStripSelector: dataTestIdSelector(
        workbookShellSlotTestId("status-strip"),
      ),
    },
  );
}

async function closeIncidentControlsIfOpen(page: Page) {
  const closeButton = page.getByRole("button", {
    name: "Close incident controls",
  });
  if ((await closeButton.count()) > 0 && (await closeButton.isVisible())) {
    await closeButton.click();
    await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(
      0,
    );
  }
}

async function expectWorkbookShellComposition(
  page: Page,
  options: {
    expectIncidentPreferences?: boolean;
    incidentId: string;
    incidentKey: string;
    incidentTitle: string;
  },
) {
  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toHaveCount(1);
  await expect(shell).toBeVisible();

  const shellId = workbookShellReadyTestId();
  await expect(shell).toHaveAttribute("data-workbook-shell-id", shellId);
  for (const slot of workbookShellSlots) {
    const slotLocator = shell.locator(
      dataTestIdSelector(workbookShellSlotTestId(slot)),
    );
    if (slot === "inspector") {
      await expect(slotLocator).toHaveCount(0);
      continue;
    }
    await expect(slotLocator).toHaveCount(1);
    await expect(slotLocator).toHaveAttribute(
      "data-workbook-shell-id",
      shellId,
    );
  }

  await expect(
    shell
      .locator(dataTestIdSelector(workbookShellSlotTestId("top-bar")))
      .getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();

  const tabBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("top-bar")),
  );
  const builtInTabsByRegistryIndex = await tabBar
    .locator("[data-workbook-tab-index]")
    .evaluateAll((nodes) =>
      nodes
        .map((node) => ({
          index: Number(node.getAttribute("data-workbook-tab-index")),
          testId: node.getAttribute("data-testid"),
          viewSchemaId: node.getAttribute("data-view-schema-id"),
        }))
        .sort((left, right) => left.index - right.index),
    );
  expect(builtInTabsByRegistryIndex).toEqual(
    requiredBuiltInWorkbookSurfaceIds.map((viewSchemaId, index) => ({
      index,
      testId: surfaceTabTestId(viewSchemaId),
      viewSchemaId,
    })),
  );

  const currentWorkbookUrl = new URL(page.url());
  expect(currentWorkbookUrl.pathname).toBe("/");
  expect(currentWorkbookUrl.searchParams.get("incident_id")).toBe(
    options.incidentId,
  );

  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expectCurrentIncidentRole(page, "Current incident role: admin");

  const closedLayout = await readWorkbookLayoutRects(page);
  await openIncidentControls(page);
  const openLayout = await readWorkbookLayoutRects(page);
  expect(openLayout).toEqual(closedLayout);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-key")),
  ).toHaveText(options.incidentKey);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-title")),
  ).toHaveText(options.incidentTitle);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-role")),
  ).toHaveText("admin");
  if (options.expectIncidentPreferences === true) {
    await expect(
      page.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref")),
    ).toHaveText("View schema: Timeline (cartulary.view.timeline.v2)");
    await expect(
      page.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref")),
    ).toHaveText("View schema: Timeline (cartulary.view.timeline.v2)");
  }
  await closeIncidentControlsIfOpen(page);
}

async function readAccountPreferences(
  page: Page,
): Promise<AccountPreferencesResource> {
  const response = await page.request.get(
    `${apiBase}/api/v1/account/preferences`,
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: AccountPreferencesResource }).data;
}

async function putAccountDensityPreference(
  page: Page,
  densityMode: AccountDensityMode,
) {
  const latest = await readAccountPreferences(page);
  const response = await page.request.put(
    `${apiBase}/api/v1/account/preferences`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_preferences_version: latest.preferences_version,
        client_txn_id: uniqueTxn(
          "browser.workbook-shell.row-01-density-restore",
        ),
        density_mode: densityMode,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function readTimelineSummaryGeometry(page: Page, recordId: string) {
  return page
    .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
    .evaluate((control) => {
      const controlElement = control as HTMLElement;
      const cellElement =
        controlElement.closest<HTMLElement>('[role="gridcell"]');
      const contentElement = controlElement.closest<HTMLElement>(
        ".cartulary-grid-cell-content",
      );
      const gridElement = controlElement.closest<HTMLElement>('[role="grid"]');
      if (
        cellElement === null ||
        contentElement === null ||
        gridElement === null
      ) {
        throw new Error("Timeline summary editor is not inside a gridcell");
      }
      const rectFor = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        return {
          blockSize: rect.height,
          inlineSize: rect.width,
          insetBlockStart: rect.top,
          insetInlineStart: rect.left,
        };
      };
      const gridStyle = window.getComputedStyle(gridElement);
      const controlStyle = window.getComputedStyle(controlElement);
      const contentStyle = window.getComputedStyle(contentElement);
      const headerStyle = window.getComputedStyle(
        gridElement.querySelector<HTMLElement>(
          '[role="columnheader"] .cartulary-grid-header-content',
        ) ?? contentElement,
      );
      const gutterStyle = window.getComputedStyle(
        gridElement.querySelector<HTMLElement>(
          ".cartulary-grid-gutter-cell .cartulary-grid-gutter-content",
        ) ?? contentElement,
      );
      return {
        cell: rectFor(cellElement),
        content: rectFor(contentElement),
        control: rectFor(controlElement),
        density: gridStyle.getPropertyValue("--cartulary-grid-density").trim(),
        fontSize: controlStyle.fontSize,
        gutterFontSize: gutterStyle.fontSize,
        gutterLineHeight: gutterStyle.lineHeight,
        gutterPaddingBlockEnd: gutterStyle.paddingBlockEnd,
        gutterPaddingBlockStart: gutterStyle.paddingBlockStart,
        gutterPaddingInlineEnd: gutterStyle.paddingInlineEnd,
        gutterPaddingInlineStart: gutterStyle.paddingInlineStart,
        hasEditableControl:
          cellElement.querySelector(
            "input, textarea, select, [contenteditable='true']",
          ) !== null,
        headerFontSize: headerStyle.fontSize,
        headerLineHeight: headerStyle.lineHeight,
        headerPaddingBlockEnd: headerStyle.paddingBlockEnd,
        headerPaddingBlockStart: headerStyle.paddingBlockStart,
        headerPaddingInlineEnd: headerStyle.paddingInlineEnd,
        headerPaddingInlineStart: headerStyle.paddingInlineStart,
        lineHeight: controlStyle.lineHeight,
        padding: contentStyle.padding,
      };
    });
}

async function readTimelineEditorGeometry(page: Page, recordId: string) {
  return page
    .getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId,
        surface: "grid",
      }),
    )
    .evaluate((control) => {
      const controlElement = control as HTMLElement;
      const cellElement =
        controlElement.closest<HTMLElement>('[role="gridcell"]');
      if (cellElement === null) {
        throw new Error("Timeline editor is not inside a gridcell");
      }
      const controlRect = controlElement.getBoundingClientRect();
      const cellRect = cellElement.getBoundingClientRect();
      const style = window.getComputedStyle(controlElement);
      return {
        cellBlockSize: cellRect.height,
        cellInlineSize: cellRect.width,
        controlBlockSize: controlRect.height,
        controlInlineSize: controlRect.width,
        fontSize: style.fontSize,
        lineHeight: style.lineHeight,
        padding: style.padding,
      };
    });
}

function expectNear(actual: number, expected: number, tolerance = 1) {
  expect(Math.abs(actual - expected)).toBeLessThanOrEqual(tolerance);
}

test("creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  const pageErrors: string[] = [];
  page.on("pageerror", (error) => {
    pageErrors.push(error.message);
  });
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("INCIDENT-ADMINISTRATION");
  await new IncidentDirectory(page).createAndOpenIncident(
    incidentKey,
    "Incident administration incident-directory",
  );

  await expect(page).toHaveURL(/incident_id=/);
  const openedWorkbookUrl = new URL(page.url());
  const openedIncidentId = openedWorkbookUrl.searchParams.get("incident_id");
  expect(openedWorkbookUrl.pathname).toBe("/");
  if (openedIncidentId === null) {
    throw new Error("expected incident-directory to open an incident workbook");
  }
  await setCurrentSavedViewAsHome(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
  ).toHaveText("Home view updated.");
  await setCurrentSavedViewAsDefault(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
  ).toHaveText("Default view updated.");

  await expectWorkbookShellComposition(page, {
    expectIncidentPreferences: true,
    incidentId: openedIncidentId,
    incidentKey,
    incidentTitle: "Incident administration incident-directory",
  });
  expect(pageErrors).toEqual([]);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("updates workbook density from Account Settings while the workbook remains open", async ({
  page,
}) => {
  const originalPreferences = await readAccountPreferences(page);
  const incidentKey = uniqueIncidentKey("INCIDENTDELETE");
  const incidentTitle = "browser.workbook-shell.row-01 density";
  const incidentId = await createIncident(page, incidentKey, incidentTitle);
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("browser.workbook-shell.row-01-density-row"),
    "timeline.activity_synopsis_text":
      "browser.workbook-shell.row-01 density row",
    "timeline.raw_activity_text":
      "browser.workbook-shell.row-01 density details",
  });

  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: row.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("browser.workbook-shell.row-01 density row");
    await expectWorkbookShellComposition(page, {
      incidentId,
      incidentKey,
      incidentTitle,
    });

    await new AccountSettings(page).openAppearance();
    const settings = new AccountSettings(page);
    const saveButton = page.getByTestId(accountTestId("appearance-save"));

    await settings.selectDensity("Comfortable");
    await saveButton.click();
    await expect
      .poll(
        async () =>
          (await readTimelineSummaryGeometry(page, row.record_id)).density,
      )
      .toBe("comfortable");
    const comfortableGeometry = await readTimelineSummaryGeometry(
      page,
      row.record_id,
    );

    await settings.selectDensity("Compact");
    await saveButton.click();
    await expect
      .poll(
        async () =>
          (await readTimelineSummaryGeometry(page, row.record_id)).density,
      )
      .toBe("compact");
    await page.getByRole("button", { exact: true, name: "Close" }).click();

    const compactGeometry = await readTimelineSummaryGeometry(
      page,
      row.record_id,
    );
    expect(comfortableGeometry.cell.blockSize).toBeGreaterThan(
      compactGeometry.cell.blockSize + 12,
    );
    expect(compactGeometry.cell.blockSize).toBeGreaterThanOrEqual(23);
    expect(compactGeometry.cell.blockSize).toBeLessThanOrEqual(26);
    expect(comfortableGeometry.fontSize).toBe("14px");
    expectNear(Number.parseFloat(comfortableGeometry.lineHeight), 18.9);
    expect(comfortableGeometry.padding).toBe("5px 9px");
    expect(comfortableGeometry.headerPaddingBlockStart).toBe("5px");
    expect(comfortableGeometry.headerPaddingBlockEnd).toBe("5px");
    expect(comfortableGeometry.headerPaddingInlineStart).toBe("9px");
    expect(comfortableGeometry.headerPaddingInlineEnd).toBe("9px");
    expect(comfortableGeometry.gutterPaddingBlockStart).toBe("5px");
    expect(comfortableGeometry.gutterPaddingBlockEnd).toBe("5px");
    expect(comfortableGeometry.gutterPaddingInlineStart).toBe("9px");
    expect(comfortableGeometry.gutterPaddingInlineEnd).toBe("9px");
    expect(compactGeometry.fontSize).toBe("12px");
    expectNear(Number.parseFloat(compactGeometry.lineHeight), 14.4);
    expect(compactGeometry.padding).toBe("2px 5px");
    expect(compactGeometry.headerFontSize).toBe("12px");
    expectNear(Number.parseFloat(compactGeometry.headerLineHeight), 14.4);
    expect(compactGeometry.headerPaddingBlockStart).toBe("2px");
    expect(compactGeometry.headerPaddingBlockEnd).toBe("2px");
    expect(compactGeometry.headerPaddingInlineStart).toBe("5px");
    expect(compactGeometry.headerPaddingInlineEnd).toBe("5px");
    expect(compactGeometry.gutterFontSize).toBe("12px");
    expectNear(Number.parseFloat(compactGeometry.gutterLineHeight), 14.4);
    expect(compactGeometry.gutterPaddingBlockStart).toBe("2px");
    expect(compactGeometry.gutterPaddingBlockEnd).toBe("2px");
    expect(compactGeometry.gutterPaddingInlineStart).toBe("5px");
    expect(compactGeometry.gutterPaddingInlineEnd).toBe("5px");
    expect(comfortableGeometry.hasEditableControl).toBe(false);
    expect(compactGeometry.hasEditableControl).toBe(false);

    await page
      .getByTestId(
        rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      )
      .dblclick();
    const editorGeometry = await readTimelineEditorGeometry(
      page,
      row.record_id,
    );
    expectNear(editorGeometry.controlBlockSize, editorGeometry.cellBlockSize);
    expectNear(editorGeometry.controlInlineSize, editorGeometry.cellInlineSize);
    expect(editorGeometry.fontSize).toBe("12px");
    expectNear(Number.parseFloat(editorGeometry.lineHeight), 14.4);
    expect(editorGeometry.padding).toBe("2px 5px");
    await page.keyboard.press("Escape");
  } finally {
    await putAccountDensityPreference(page, originalPreferences.density_mode);
  }
});

test("Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("INCIDENTMEMBERSHIP"),
    "browser.workbook-shell.row-02 switcher focus",
  );
  await createViewRow(page, incidentId, indicatorsViewSchemaId, {
    client_txn_id: uniqueTxn("browser.workbook-shell.row-02-indicator"),
    "indicator.indicator_type": "ipv4_addr",
    "indicator.value_kind": "atomic",
    "indicator.display_value": "198.51.100.24",
  });
  const timelineUrlPattern = new RegExp(
    `view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page).toHaveURL(timelineUrlPattern);

  const lastBuiltInTabId =
    requiredBuiltInWorkbookSurfaceIds[
      requiredBuiltInWorkbookSurfaceIds.length - 1
    ];
  if (!lastBuiltInTabId) {
    throw new Error("Missing final built-in workbook surface");
  }
  await page.getByTestId(surfaceTabTestId(lastBuiltInTabId)).focus();
  await page.keyboard.press("Tab");

  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expect(trigger).toBeFocused();
  await expect(
    page.getByRole("button", { name: "System views" }),
  ).toHaveAttribute("data-testid", systemViewSwitcherTriggerTestId());

  await page.keyboard.press("Enter");
  const indicatorOption = page.getByTestId(
    systemViewSwitcherOptionTestId("scope-indicators", indicatorsViewSchemaId),
  );
  const assessmentOption = page.getByTestId(
    systemViewSwitcherOptionTestId(
      "scope-indicators",
      "cartulary.view.assessments.v1",
    ),
  );
  await expect(page.getByTestId(systemViewSwitcherMenuTestId())).toBeVisible();
  await expect(indicatorOption).toBeFocused();

  await page.keyboard.press("ArrowDown");
  await expect(assessmentOption).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(systemViewSwitcherMenuTestId())).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await expect(page).toHaveURL(timelineUrlPattern);

  await page.keyboard.press(" ");
  await expect(indicatorOption).toBeFocused();
  await page.keyboard.press("Enter");

  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`),
  );
  await expect(
    page.getByTestId(genericCreateFieldTestId("indicator.indicator_type")),
  ).toBeFocused();
});

test("Verify saved views appear only under the active surface's view selector and system views open inside the same workbook shell.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("INCIDENTADMIN"),
    "end-to-end.workbook-shell.row-01 saved-view placement",
  );
  const host = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("end-to-end.workbook-shell.row-01-host"),
    "host.display_name": "end-to-end.workbook-shell.row-01 host",
    "host.hostname": "end-to-end.workbook-shell.row-01-host.example.test",
  });
  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("end-to-end.workbook-shell.row-01-indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "203.0.113.45",
    },
  );
  const hostSavedView = await createSavedView(page, incidentId, {
    display_name: "end-to-end.workbook-shell.row-01 host saved view",
    scope: "shared",
    view_schema_id: hostsViewSchemaId,
  });
  const indicatorSavedView = await createSavedView(page, incidentId, {
    display_name: "end-to-end.workbook-shell.row-01 indicator saved view",
    scope: "shared",
    view_schema_id: indicatorsViewSchemaId,
  });
  const indicatorSystemSavedView = await seedSystemSavedView(page, incidentId, {
    display_name:
      "end-to-end.workbook-shell.row-01 indicator system saved view",
    view_schema_id: indicatorsViewSchemaId,
  });

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );

  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toBeVisible();
  await expect(shell).toHaveAttribute(
    "data-active-view-schema-id",
    hostsViewSchemaId,
  );
  const shellId = await shell.getAttribute("data-workbook-shell-id");
  expect(shellId).toBe(workbookShellReadyTestId());

  await expect(
    page.getByTestId(
      rowCellTestId(host.record_id as string, "host.display_name"),
    ),
  ).toBeVisible();

  const tabBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("top-bar")),
  );
  await expect(tabBar.locator(savedViewFamilySelector())).toHaveCount(0);

  const viewBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("view-bar")),
  );
  const hostSelector = viewBar.getByTestId(
    savedViewSelectorTestId(hostsViewSchemaId),
  );
  await expect(hostSelector).toHaveCount(1);
  await expect(
    hostSelector.getByTestId(
      savedViewOptionTestId(hostsViewSchemaId, hostSavedView.saved_view_id),
    ),
  ).toHaveAttribute("data-view-schema-id", hostsViewSchemaId);
  await expect(
    hostSelector.getByTestId(
      savedViewOptionTestId(
        hostsViewSchemaId,
        indicatorSavedView.saved_view_id,
      ),
    ),
  ).toHaveCount(0);

  const indicatorsQuery = waitForViewQuery(
    page,
    incidentId,
    indicatorsViewSchemaId,
  );
  await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
  const indicatorSystemViewOption = page.getByTestId(
    systemViewSwitcherOptionTestId("scope-indicators", indicatorsViewSchemaId),
  );
  await expect(indicatorSystemViewOption).toHaveAttribute(
    "data-view-schema-id",
    indicatorsViewSchemaId,
  );
  await expect(indicatorSystemViewOption).not.toHaveAttribute(
    "data-saved-view-id",
  );
  await indicatorSystemViewOption.click();
  await indicatorsQuery;

  await expect(shell).toHaveAttribute("data-workbook-shell-id", shellId ?? "");
  await expect(shell).toHaveAttribute(
    "data-active-view-schema-id",
    indicatorsViewSchemaId,
  );
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      rowCellTestId(indicator.record_id as string, "indicator.indicator_type"),
    ),
  ).toBeVisible();
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`),
  );
  await expect(tabBar.locator(savedViewFamilySelector())).toHaveCount(0);

  const indicatorSelector = viewBar.getByTestId(
    savedViewSelectorTestId(indicatorsViewSchemaId),
  );
  await expect(indicatorSelector).toHaveCount(1);
  await expect(
    indicatorSelector.getByTestId(
      savedViewOptionTestId(
        indicatorsViewSchemaId,
        indicatorSavedView.saved_view_id,
      ),
    ),
  ).toHaveAttribute("data-saved-view-id", indicatorSavedView.saved_view_id);
  await expect(
    indicatorSelector.getByTestId(
      savedViewOptionTestId(
        indicatorsViewSchemaId,
        indicatorSystemSavedView.saved_view_id,
      ),
    ),
  ).toHaveAttribute(
    "data-saved-view-id",
    indicatorSystemSavedView.saved_view_id,
  );
  await expect(
    indicatorSelector.getByTestId(
      savedViewOptionTestId(
        indicatorsViewSchemaId,
        hostSavedView.saved_view_id,
      ),
    ),
  ).toHaveCount(0);

  const selectedSavedViewQuery = waitForViewQuery(
    page,
    incidentId,
    indicatorsViewSchemaId,
  );
  await indicatorSelector.selectOption(indicatorSystemSavedView.saved_view_id);
  await selectedSavedViewQuery;
  await expect(indicatorSelector).toHaveAttribute(
    "data-selected-sheet-ref-kind",
    "saved_view",
  );
  await expect(indicatorSelector).toHaveAttribute(
    "data-selected-saved-view-id",
    indicatorSystemSavedView.saved_view_id,
  );
  const savedViewUrl = new URL(page.url());
  expect(savedViewUrl.searchParams.get("sheet_ref_kind")).toBe("saved_view");
  expect(savedViewUrl.searchParams.get("sheet_ref_id")).toBe(
    indicatorSystemSavedView.saved_view_id,
  );
  expect(savedViewUrl.searchParams.get("view_schema_id")).toBeNull();
  await expect(shell).toHaveAttribute(
    "data-active-view-schema-id",
    indicatorsViewSchemaId,
  );
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      rowCellTestId(indicator.record_id as string, "indicator.indicator_type"),
    ),
  ).toBeVisible();
});

test("shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("INCIDENT-MEMBERSHIP");
  const incidentId = await createIncident(
    page,
    incidentKey,
    "Incident administration incident-directory",
  );

  await new IncidentDirectory(page).goto();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText(incidentKey);
  await page.getByTestId(landingIncidentOpenButtonTestId(incidentId)).click();

  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-key")),
  ).toHaveText(incidentKey);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-title")),
  ).toHaveText("Incident administration incident-directory");
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-version")),
  ).toHaveText("Version 1");
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-tlp")),
  ).toHaveText("Unset");
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-current-phase")),
  ).toHaveText("Unset");
  await expect(
    page.getByTestId(
      incidentAdministrationTestId("summary-primary-external-case-ref"),
    ),
  ).toHaveText("Unset");

  await openIncidentControls(page, "incident-fields");
  await page
    .getByTestId(incidentAdministrationTestId("patch-tlp"))
    .selectOption("TLP:AMBER");
  await page
    .getByTestId(incidentAdministrationTestId("patch-current-phase"))
    .fill("containment");
  await page
    .getByTestId(incidentAdministrationTestId("patch-external-case"))
    .fill("CASE-E202-PRIMARY");
  await page.getByTestId(incidentAdministrationTestId("patch-button")).click();

  const actionMessage = page.getByTestId(incidentControlsActionMessageTestId());
  await expect(actionMessage).toHaveText("Saved promoted incident fields.");
  await actionMessage.hover();
  await page.waitForTimeout(5_100);
  await expect(actionMessage).toHaveText("Saved promoted incident fields.");
  await page.getByLabel("Close incident controls").hover();
  await expect(actionMessage).toHaveText("", { timeout: 15_000 });

  await openIncidentControls(page);
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-version")),
  ).toHaveText("Version 2");
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-tlp")),
  ).toHaveText("TLP:AMBER");
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-current-phase")),
  ).toHaveText("containment");
  await expect(
    page.getByTestId(
      incidentAdministrationTestId("summary-primary-external-case-ref"),
    ),
  ).toHaveText("CASE-E202-PRIMARY");

  await page.getByLabel("Account and application navigation").click();
  await page.getByTestId(incidentLandingTestId("return")).click();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("containment");
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("TLP:AMBER");
});

test("lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const memberEmail = uniqueEmail("incident_membership-e203-member");
  const memberPassword = "IncidentMembershipE203Pass!";
  const memberUser = await createDeploymentUser(page, {
    email: memberEmail,
    display_name: "Incident administration E203 Member",
    initial_password: memberPassword,
    is_deployment_admin: false,
    mfa_required: false,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("INCIDENT-ACCESS"),
    "Incident administration incident-directory",
  );

  await openIncidentFromLanding(page, incidentId);
  await openIncidentControls(page, "memberships");

  await page
    .getByRole("button", { name: "Add existing account", exact: true })
    .click();
  await page
    .getByTestId(incidentMembershipEmailInputTestId())
    .fill(memberEmail);
  await page
    .getByTestId(incidentMembershipRoleSelectTestId())
    .selectOption("viewer");
  await page.getByTestId(incidentMembershipCreateButtonTestId()).click();

  await expect(
    page.getByTestId(incidentMembershipRowTestId(memberUser.user_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(incidentMembershipRoleDisplayTestId(memberUser.user_id)),
  ).toHaveText("viewer");
  await page
    .getByRole("button", {
      name: `Change role for ${memberUser.display_name} (${memberUser.user_id})`,
      exact: true,
    })
    .click();

  await page
    .getByTestId(incidentMembershipRoleInputTestId(memberUser.user_id))
    .selectOption("reviewer");
  await page
    .getByTestId(incidentMembershipPatchButtonTestId(memberUser.user_id))
    .click();

  await expect(
    page.getByTestId(incidentMembershipRoleDisplayTestId(memberUser.user_id)),
  ).toHaveText("reviewer");
  await expect(
    page.getByTestId(incidentMembershipVersionTestId(memberUser.user_id)),
  ).toHaveText("Version 2");

  await closeIncidentControlsIfOpen(page);
  await openIncidentControls(page, "membership-audit");
  await expect(
    page.getByTestId(incidentAdministrationTestId("membership-audit-list")),
  ).toBeVisible();
  await expect(page.getByText("Membership role changed")).toBeVisible();
  await expect(
    page
      .getByTestId(incidentAdministrationTestId("membership-audit-list"))
      .getByText(memberUser.user_id)
      .first(),
  ).toBeVisible();
  await closeIncidentControlsIfOpen(page);

  const memberPage = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "incident_membership non-admin membership view",
    email: memberEmail,
    incidentId,
    password: memberPassword,
    purpose: "incident_membership e203 non-admin incident shell",
    userId: memberUser.user_id,
  });

  await openIncidentControls(memberPage, "memberships");
  await expect(
    memberPage.getByTestId(incidentMembershipAdminNoteTestId()),
  ).toBeVisible();
  await expect(
    memberPage.getByTestId(incidentMembershipRowTestId(memberUser.user_id)),
  ).toBeVisible();
  await expect(
    memberPage.getByTestId(
      incidentMembershipRoleDisplayTestId(memberUser.user_id),
    ),
  ).toHaveText("reviewer");
  await expect(
    memberPage.getByTestId(incidentMembershipCreateButtonTestId()),
  ).toHaveCount(0);
  await expect(
    memberPage.getByTestId(
      incidentMembershipPatchButtonTestId(memberUser.user_id),
    ),
  ).toHaveCount(0);
  await expect(
    memberPage.getByTestId(
      incidentMembershipDeleteButtonTestId(memberUser.user_id),
    ),
  ).toHaveCount(0);
  await closeIncidentControlsIfOpen(memberPage);
  await openIncidentControls(memberPage, "membership-audit");
  await expect(
    memberPage.getByTestId(
      incidentAdministrationTestId("membership-audit-note"),
    ),
  ).toContainText("Only incident admins");
  await memberPage.context().close();

  await openIncidentControls(page, "summary");
  await page
    .getByTestId(incidentAdministrationTestId("lifecycle-reason"))
    .fill("Membership access review complete");
  await page.getByTestId(incidentAdministrationTestId("close-button")).click();
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-status")),
  ).toHaveText("Closed, read-only");
  await page
    .getByTestId(incidentAdministrationTestId("lifecycle-reason"))
    .fill("Additional membership cleanup required");
  await page.getByTestId(incidentAdministrationTestId("reopen-button")).click();
  await expect(
    page.getByTestId(incidentAdministrationTestId("summary-status")),
  ).toHaveText("active");
  await closeIncidentControlsIfOpen(page);

  await openIncidentControls(page, "memberships");
  await page
    .getByTestId(incidentMembershipDeleteButtonTestId(memberUser.user_id))
    .click();
  await page
    .getByRole("button", { name: "Confirm removal", exact: true })
    .click();
  await expect(
    page.getByTestId(incidentMembershipRowTestId(memberUser.user_id)),
  ).toHaveCount(0);
});

function waitForViewQuery(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
) {
  return page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request
        .url()
        .endsWith(
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
  );
}

test("account settings retain edits through real version conflicts and exact committed replay", async ({
  workerAdminPage: page,
}) => {
  const originalResponse = await page.request.get(
    `${apiBase}/api/v1/account/profile`,
  );
  expect(originalResponse.ok()).toBeTruthy();
  const original = (await originalResponse.json()).data as {
    display_name: string;
    user_version: number;
  };
  const bodies: Record<string, unknown>[] = [];
  const committed = accountResponseGate();
  const release = accountResponseGate();
  let loseResponse = false;
  await page.route("**/api/v1/account/profile", async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    bodies.push(route.request().postDataJSON() as Record<string, unknown>);
    if (!loseResponse) return route.continue();
    loseResponse = false;
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    committed.release();
    await release.promise;
    await route.abort("failed");
  });
  try {
    await page.goto("/");
    const settings = new AccountSettings(page);
    await settings.openProfile();
    const field = page.getByRole("textbox", { name: "Display name" });
    await field.fill("Retained conflict draft");
    const external = await page.request.patch(
      `${apiBase}/api/v1/account/profile`,
      {
        headers: await csrfHeaders(page),
        data: {
          client_txn_id: uniqueTxn("account-external-profile"),
          base_user_version: original.user_version,
          display_name: "Another session name",
        },
      },
    );
    expect(external.ok()).toBeTruthy();
    await field.press("Enter");
    await expect(page.getByRole("alert")).toContainText(
      "Display name changed elsewhere",
    );
    await expect(field).toHaveValue("Retained conflict draft");
    await expect(
      page.getByText("Another session name", { exact: true }),
    ).toBeVisible();
    expect(bodies).toHaveLength(1);
    await page.getByRole("button", { name: "Review my edit" }).focus();
    await page.keyboard.press("Enter");
    await field.focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("button", { name: "Account and application navigation" }),
    ).toHaveAttribute("title", "Retained conflict draft");
    expect(bodies).toHaveLength(2);
    expect(bodies[1]?.client_txn_id).not.toBe(bodies[0]?.client_txn_id);
    loseResponse = true;
    await field.fill("Committed response lost");
    await field.press("Enter");
    await committed.promise;
    await field.fill("Newer unsaved edit");
    await page.getByRole("tab", { name: "Appearance" }).click();
    await page.getByRole("button", { name: "Close", exact: true }).click();
    release.release();
    await settings.openProfile();
    await expect(field).toHaveValue("Newer unsaved edit");
    await expect(
      page.getByRole("button", { name: "Retry save" }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Retry save" }).focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("button", { name: "Account and application navigation" }),
    ).toHaveAttribute("title", "Committed response lost");
    expect(bodies).toHaveLength(4);
    expect(bodies[3]).toEqual(bodies[2]);
    await expect(field).toHaveValue("Newer unsaved edit");
    const current = (
      await (await page.request.get(`${apiBase}/api/v1/account/profile`)).json()
    ).data;
    expect(current.display_name).toBe("Committed response lost");
    expect(current.user_version).toBe(Number(bodies[2]?.base_user_version) + 1);
  } finally {
    release.release();
    await page.unroute("**/api/v1/account/profile");
    const current = (
      await (await page.request.get(`${apiBase}/api/v1/account/profile`)).json()
    ).data;
    const restored = await page.request.patch(
      `${apiBase}/api/v1/account/profile`,
      {
        headers: await csrfHeaders(page),
        data: {
          client_txn_id: uniqueTxn("account-profile-restore"),
          base_user_version: current.user_version,
          display_name: original.display_name,
        },
      },
    );
    expect(restored.ok()).toBeTruthy();
  }
});

test("account settings keyboard density recovery and constrained presentation", async ({
  workerAdminPage: page,
}, testInfo) => {
  const fixture = await installAccountEditingFixture(page);
  await page.goto("/");
  const settings = new AccountSettings(page);
  await settings.openProfile();
  const field = page.getByRole("textbox", { name: "Display name" });
  await field.focus();
  await page.keyboard.press("ControlOrMeta+A");
  await page.keyboard.insertText("Analyst".repeat(36));
  await page.getByRole("tab", { name: "Appearance" }).focus();
  await page.keyboard.press("Enter");
  const group = page.getByRole("radiogroup", { name: "Density", exact: true });
  await expect(group.getByRole("radio")).toHaveCount(4);
  const surfaceDefault = group.getByRole("radio", {
    name: "Use surface default",
    exact: true,
  });
  await surfaceDefault.focus();
  await page.keyboard.press("ArrowRight");
  await expect(
    group.getByRole("radio", { name: "Compact", exact: true }),
  ).toBeChecked();
  await page.keyboard.press("ArrowRight");
  await expect(
    group.getByRole("radio", { name: "Default", exact: true }),
  ).toBeChecked();
  await page.keyboard.press("ArrowRight");
  await expect(
    group.getByRole("radio", { name: "Comfortable", exact: true }),
  ).toBeChecked();
  fixture.fault("appearance", "conflict");
  await page.getByRole("button", { name: "Save appearance" }).focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("alert")).toContainText(
    "Density changed elsewhere",
  );
  await expect(
    group.getByRole("radio", { name: "Comfortable", exact: true }),
  ).toBeChecked();
  expect(fixture.requests.appearance).toHaveLength(1);
  await page.getByRole("button", { name: "Review my edit" }).focus();
  await page.keyboard.press("Enter");
  fixture.fault("appearance", "lost");
  await page.getByRole("button", { name: "Save appearance" }).focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("button", { name: "Retry save" })).toBeVisible();
  await page.getByRole("button", { name: "Retry save" }).focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "Appearance saved." }),
  ).toBeVisible();
  expect(fixture.requests.appearance[2]).toEqual(
    fixture.requests.appearance[1],
  );
  await surfaceDefault.focus();
  await page.keyboard.press("Space");
  await page.getByRole("button", { name: "Save appearance" }).focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "Appearance saved." }),
  ).toBeVisible();
  expect(fixture.requests.appearance.at(-1)?.density_mode).toBeNull();
  for (const [width, height] of [
    [1280, 720],
    [1024, 720],
    [768, 640],
    [640, 480],
    [480, 320],
  ] as const) {
    await page.setViewportSize({ width, height });
    await group
      .getByRole("radio", { name: "Comfortable", exact: true })
      .focus();
    await expectAccountControlInViewport(
      group.getByRole("radio", { name: "Comfortable", exact: true }),
    );
    await page.getByRole("button", { name: "Refresh appearance" }).focus();
    await expectAccountControlInViewport(
      page.getByRole("button", { name: "Refresh appearance" }),
    );
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await expectAccountControlInViewport(
      page.getByRole("dialog", { name: "Account settings" }),
    );
    expect(
      await group.evaluate((node) => node.scrollWidth <= node.clientWidth),
    ).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await surfaceDefault.focus();
  await expectAccountControlInViewport(surfaceDefault);
  await expectAccountControlInViewport(
    page.getByRole("dialog", { name: "Account settings" }),
  );
  expect(
    await group.evaluate((node) => node.scrollWidth <= node.clientWidth),
  ).toBe(true);
  await page.keyboard.press("ArrowRight");
  await expect(
    page.getByRole("button", { name: "Save appearance" }),
  ).toBeEnabled();
  for (const control of [
    group.getByRole("radio", { name: "Comfortable", exact: true }),
    page.getByRole("button", { name: "Save appearance" }),
    page.getByRole("button", { name: "Refresh appearance" }),
    page.getByRole("button", { name: "Close", exact: true }),
  ]) {
    await control.focus();
    await expectAccountControlInViewport(control);
  }
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  const spacing = await page.addStyleTag({
    content:
      "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
  });
  await page.setViewportSize({ width: 640, height: 480 });
  await surfaceDefault.focus();
  await expectAccountControlInViewport(surfaceDefault);
  await testInfo.attach("account-density-keyboard-tree", {
    body: await group.ariaSnapshot(),
    contentType: "text/plain",
  });
  await spacing.evaluate((node) => node.parentNode?.removeChild(node));
  await page.getByRole("tab", { name: "Profile" }).click();
  await expect(field).toHaveValue("Analyst".repeat(36));
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toBeFocused();
});

async function expectAccountControlInViewport(
  control: import("@playwright/test").Locator,
) {
  await expect(control).toBeInViewport();
  const visible = await control.evaluate((element) => {
    const bounds = element.getBoundingClientRect();
    return (
      bounds.left >= 0 &&
      bounds.top >= 0 &&
      bounds.right <= window.innerWidth &&
      bounds.bottom <= window.innerHeight
    );
  });
  expect(visible).toBe(true);
}
