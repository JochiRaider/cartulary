import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  currentIncidentRoleTestId,
  dataTestIdSelector,
  genericCreateFieldTestId,
  gridShellTestId,
  incidentControlsPanelTestId,
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
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1AccountTestId,
  phase1LandingTestId,
  rowCellTestId,
  savedViewFamilySelector,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { openIncidentControls } from "./pages/deploymentAdministration";
import {
  IncidentDirectory,
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
} from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import { createIncident } from "./support/incidents/fixtures";
import { createLocalUser } from "./support/incidents/memberships";
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
type AccountPreferencesResource = {
  density_mode: AccountDensityMode;
  preferences_version: number;
};

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
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
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    options.incidentKey,
  );
  await expect(page.getByTestId("incident-summary-title")).toHaveText(
    options.incidentTitle,
  );
  await expect(page.getByTestId("incident-summary-role")).toHaveText("admin");
  if (options.expectIncidentPreferences === true) {
    await expect(
      page.getByTestId("incident-pref-default-sheet-ref"),
    ).toHaveText("View schema: Timeline (cartulary.view.timeline.v2)");
    await expect(page.getByTestId("incident-pref-home-sheet-ref")).toHaveText(
      "View schema: Timeline (cartulary.view.timeline.v2)",
    );
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
        client_txn_id: uniqueTxn("fe-b-p2-01-density-restore"),
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
      const gridElement = controlElement.closest<HTMLElement>('[role="grid"]');
      if (cellElement === null || gridElement === null) {
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
      return {
        cell: rectFor(cellElement),
        control: rectFor(controlElement),
        density: gridStyle.getPropertyValue("--cartulary-grid-density").trim(),
        fontSize: controlStyle.fontSize,
        hasEditableControl:
          cellElement.querySelector(
            "input, textarea, select, [contenteditable='true']",
          ) !== null,
        lineHeight: controlStyle.lineHeight,
      };
    });
}

function expectNear(actual: number, expected: number, tolerance = 1) {
  expect(Math.abs(actual - expected)).toBeLessThanOrEqual(tolerance);
}

test("E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  const pageErrors: string[] = [];
  page.on("pageerror", (error) => {
    pageErrors.push(error.message);
  });
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("E201");
  await new IncidentDirectory(page).createAndOpenIncident(
    incidentKey,
    "Phase 2 E-2-01",
  );

  await expect(page).toHaveURL(/incident_id=/);
  const openedWorkbookUrl = new URL(page.url());
  const openedIncidentId = openedWorkbookUrl.searchParams.get("incident_id");
  expect(openedWorkbookUrl.pathname).toBe("/");
  if (openedIncidentId === null) {
    throw new Error("expected E-2-01 to open an incident workbook");
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
    incidentTitle: "Phase 2 E-2-01",
  });
  expect(pageErrors).toEqual([]);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("FE-B-P2-01 updates workbook density from Account Settings while the workbook remains open", async ({
  page,
}) => {
  const originalPreferences = await readAccountPreferences(page);
  const incidentKey = uniqueIncidentKey("FEBP201D");
  const incidentTitle = "FE-B-P2-01 density";
  const incidentId = await createIncident(page, incidentKey, incidentTitle);
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("fe-b-p2-01-density-row"),
    "timeline.activity_synopsis_text": "FE-B-P2-01 density row",
    "timeline.raw_activity_text": "FE-B-P2-01 density details",
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
    ).toHaveText("FE-B-P2-01 density row");
    await expectWorkbookShellComposition(page, {
      incidentId,
      incidentKey,
      incidentTitle,
    });

    await new AccountSettings(page).open("account-appearance");
    const densitySelect = page.getByTestId(
      phase1AccountTestId("appearance-density-mode"),
    );
    const saveButton = page.getByTestId(phase1AccountTestId("appearance-save"));

    await densitySelect.selectOption("comfortable");
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

    await densitySelect.selectOption("compact");
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
    expect(compactGeometry.fontSize).toBe("13px");
    expectNear(Number.parseFloat(compactGeometry.lineHeight), 19.5);
    expect(comfortableGeometry.hasEditableControl).toBe(false);
    expect(compactGeometry.hasEditableControl).toBe(false);
  } finally {
    await putAccountDensityPreference(page, originalPreferences.density_mode);
  }
});

test("FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEBP202"),
    "FE-B-P2-02 switcher focus",
  );
  await createViewRow(page, incidentId, indicatorsViewSchemaId, {
    client_txn_id: uniqueTxn("fe-b-p2-02-indicator"),
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

test("FE-E-P2-01 Verify saved views appear only under the active surface's view selector and system views open inside the same workbook shell.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP201"),
    "FE-E-P2-01 saved-view placement",
  );
  const host = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("fe-e-p2-01-host"),
    "host.display_name": "FE-E-P2-01 host",
    "host.hostname": "fe-e-p2-01-host.example.test",
  });
  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("fe-e-p2-01-indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "203.0.113.45",
    },
  );
  const hostSavedView = await createSavedView(page, incidentId, {
    display_name: "FE-E-P2-01 host saved view",
    scope: "shared",
    view_schema_id: hostsViewSchemaId,
  });
  const indicatorSavedView = await createSavedView(page, incidentId, {
    display_name: "FE-E-P2-01 indicator saved view",
    scope: "shared",
    view_schema_id: indicatorsViewSchemaId,
  });
  const indicatorSystemSavedView = await seedSystemSavedView(page, incidentId, {
    display_name: "FE-E-P2-01 indicator system saved view",
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

test("E-2-02 shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("E202");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-02");

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
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );
  await expect(page.getByTestId("incident-summary-title")).toHaveText(
    "Phase 2 E-2-02",
  );
  await expect(page.getByTestId("incident-summary-version")).toHaveText(
    "Version 1",
  );
  await expect(page.getByTestId("incident-summary-tlp")).toHaveText("Unset");
  await expect(page.getByTestId("incident-summary-current-phase")).toHaveText(
    "Unset",
  );
  await expect(
    page.getByTestId("incident-summary-primary-external-case-ref"),
  ).toHaveText("Unset");

  await openIncidentControls(page, "incident-fields");
  await page.getByTestId("incident-patch-tlp").selectOption("TLP:AMBER");
  await page.getByTestId("incident-patch-current-phase").fill("containment");
  await page
    .getByTestId("incident-patch-external-case")
    .fill("CASE-E202-PRIMARY");
  await page.getByTestId("incident-patch-button").click();

  await openIncidentControls(page);
  await expect(page.getByTestId("incident-summary-version")).toHaveText(
    "Version 2",
  );
  await expect(page.getByTestId("incident-summary-tlp")).toHaveText(
    "TLP:AMBER",
  );
  await expect(page.getByTestId("incident-summary-current-phase")).toHaveText(
    "containment",
  );
  await expect(
    page.getByTestId("incident-summary-primary-external-case-ref"),
  ).toHaveText("CASE-E202-PRIMARY");

  await page.getByLabel("Account and application navigation").click();
  await page.getByTestId(phase1LandingTestId("return")).click();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("containment");
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("TLP:AMBER");
});

test("E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const memberEmail = uniqueEmail("phase2-e203-member");
  const memberPassword = "Phase2E203Pass!";
  const memberUser = await createLocalUser(page, {
    email: memberEmail,
    display_name: "Phase 2 E203 Member",
    initial_password: memberPassword,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E203"),
    "Phase 2 E-2-03",
  );

  await openIncidentFromLanding(page, incidentId);
  await openIncidentControls(page, "memberships");

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
    page.getByTestId(incidentMembershipRoleInputTestId(memberUser.user_id)),
  ).toHaveValue("viewer");

  await page
    .getByTestId(incidentMembershipRoleInputTestId(memberUser.user_id))
    .selectOption("reviewer");
  await page
    .getByTestId(incidentMembershipPatchButtonTestId(memberUser.user_id))
    .click();

  await expect(
    page.getByTestId(incidentMembershipRoleInputTestId(memberUser.user_id)),
  ).toHaveValue("reviewer");
  await expect(
    page.getByTestId(incidentMembershipVersionTestId(memberUser.user_id)),
  ).toHaveText("Version 2");

  const memberPage = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "phase2 non-admin membership view",
    email: memberEmail,
    incidentId,
    password: memberPassword,
    purpose: "phase2 e203 non-admin incident shell",
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
  await memberPage.context().close();

  await page
    .getByTestId(incidentMembershipDeleteButtonTestId(memberUser.user_id))
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
