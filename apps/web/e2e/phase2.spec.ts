import {
  currentIncidentRoleTestId,
  dataTestIdSelector,
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
  phase1LandingTestId,
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  indicatorsViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
} from "../src/workbookSurfaceRegistry";
import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  createLocalUser,
  openIncidentAsTrackedUser,
  openIncidentFromLanding,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

test("E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("E201");
  await page.getByTestId(phase1LandingTestId("incident-key")).fill(incidentKey);
  await page
    .getByTestId(phase1LandingTestId("incident-title"))
    .fill("Phase 2 E-2-01");
  await page.getByTestId(phase1LandingTestId("create-button")).click();

  await expect(page).toHaveURL(/incident_id=/);
  const openedWorkbookUrl = new URL(page.url());
  const openedIncidentId = openedWorkbookUrl.searchParams.get("incident_id");
  expect(openedWorkbookUrl.pathname).toBe("/");
  expect(openedIncidentId).not.toBeNull();

  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toHaveCount(1);
  await expect(shell).toBeVisible();

  const shellId = workbookShellReadyTestId();
  await expect(shell).toHaveAttribute("data-workbook-shell-id", shellId);
  for (const slot of workbookShellSlots) {
    const slotLocator = shell.locator(
      dataTestIdSelector(workbookShellSlotTestId(slot)),
    );
    await expect(slotLocator).toHaveCount(1);
    await expect(slotLocator).toHaveAttribute(
      "data-workbook-shell-id",
      shellId,
    );
  }

  await expect(
    shell
      .locator(dataTestIdSelector(workbookShellSlotTestId("system-views")))
      .getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();

  const tabBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("tab-bar")),
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
    openedIncidentId,
  );

  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    "Current incident role: admin",
  );
  await expect(page.getByTestId("incident-summary-key")).toHaveText(
    incidentKey,
  );
  await expect(page.getByTestId("incident-summary-title")).toHaveText(
    "Phase 2 E-2-01",
  );
  await expect(page.getByTestId("incident-summary-role")).toHaveText("admin");
  await expect(page.getByTestId("incident-pref-default-sheet-ref")).toHaveText(
    "Unset",
  );
  await expect(page.getByTestId("incident-pref-home-sheet-ref")).toHaveText(
    "Unset",
  );
});

test("FE-B-P2-02 Verify System views switcher keyboard entry, roving focus, selection, dismissal, and focus restoration.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEBP202"),
    "FE-B-P2-02 switcher focus",
  );
  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("fe-b-p2-02-indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "198.51.100.24",
    },
  );
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
    systemViewSwitcherOptionTestId("scope-assessment", indicatorsViewSchemaId),
  );
  const assessmentOption = page.getByTestId(
    systemViewSwitcherOptionTestId(
      "scope-assessment",
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
  const firstGridFocusTarget = page.getByTestId(
    rowCellTestId(indicator.record_id as string, "indicator.indicator_type"),
  );
  await expect(firstGridFocusTarget).toBeFocused();
});

test("E-2-02 shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("E202");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-02");

  await page.goto("/");
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

  await page.getByTestId("incident-patch-tlp").fill("amber");
  await page.getByTestId("incident-patch-current-phase").fill("containment");
  await page
    .getByTestId("incident-patch-external-case")
    .fill("CASE-E202-PRIMARY");
  await page.getByTestId("incident-patch-button").click();

  await expect(page.getByTestId("incident-summary-version")).toHaveText(
    "Version 2",
  );
  await expect(page.getByTestId("incident-summary-tlp")).toHaveText("amber");
  await expect(page.getByTestId("incident-summary-current-phase")).toHaveText(
    "containment",
  );
  await expect(
    page.getByTestId("incident-summary-primary-external-case-ref"),
  ).toHaveText("CASE-E202-PRIMARY");

  await page.getByTestId(phase1LandingTestId("return")).click();
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("containment");
  await expect(
    page.getByTestId(landingIncidentCardTestId(incidentId)),
  ).toContainText("amber");
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
