import { Buffer } from "node:buffer";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  pasteGridMatrix,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsStatusTestId,
  landingIncidentCardTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  networkAnalysisTestId,
  pendingQueueCountTestId,
  pendingQueueDiscardButtonTestId,
  pendingQueueNoticeTestId,
  pendingQueueRecoveryPanelTestId,
  pendingQueueRetryButtonTestId,
  phase1AccountTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  type SystemViewSwitcherGroupToken,
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewNameInputTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineScalarEditorTestId,
  type WorkbookSurface,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Locator, Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { AuthGateway } from "./pages/authGateway";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  createLocalUser as createAuthLocalUser,
  revokeAllSessions,
} from "./support/auth/sessions";
import { sessionCookieName } from "./support/auth/storageState";
import {
  enrollTotpViaBootstrap,
  generateTotpCode,
} from "./support/auth/suiteAdmin";
import {
  driveRealTimelineSummaryConflict,
  focusRemoteTimelineCellAndWaitForPresence,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  requireRecordId,
  successfulPatchCalls,
} from "./support/collaboration/replay";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  requireItemByRawText,
  seedHostMentionStateFixture,
} from "./support/entities/mentions";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./support/evidence/fixtures";
import {
  importNetworkFlowCSV,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";
import { createIncident } from "./support/incidents/fixtures";
import {
  createIncidentMembership,
  createIncidentMemberUser,
} from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { safelyRemoveRoute as safeUnroute } from "./support/transport/requestInterception";
import { createEnvironmentTestControlClient } from "./support/transport/testControlEnvironment";
import { createViewRow, patchRecord } from "./support/workbook/query";
import {
  clickTimelineRowAction,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

type A11yHistoryItem = {
  available_rollback_actions: Array<
    "history_entry" | "change_set" | "row_restore"
  >;
  history_entry_ref?: string;
  history_item_ref: string;
};

type A11yHistoryData = {
  items: A11yHistoryItem[];
};

type ViewRow = {
  cells: Record<string, { value: unknown }>;
  record_id: string;
  row_version: number;
};

declare const A11yAppLocalTestIdBrand: unique symbol;

type AuthA11yAppLocalTestId = string & {
  readonly [A11yAppLocalTestIdBrand]: "AuthA11yAppLocalTestId";
};

const p1AccessibilityScenarioTitles = [
  "FE-A11Y-P1-01 deferred session loading exposes progress and keeps recovery controls keyboard reachable",
  "FE-A11Y-P1-01 anonymous login after initial session_required reaches login controls and authenticated landing",
  "FE-A11Y-P1-01 mfa_required challenge is keyboard reachable, visibly focused, named, and safely announced",
  "FE-A11Y-P1-01 mfa_setup_required enrollment is keyboard reachable and public errors hide private setup diagnostics",
  "FE-A11Y-P1-01 authenticated landing exposes account, admin, incident, retry, and visible incident controls",
  "FE-A11Y-P1-01 incident empty, list, selected, stale-selection, and incident-error states expose keyboard recovery",
  "FE-A11Y-P1-01 forbidden access-denied public envelope is announced and exposes recovery without private diagnostics",
  "FE-A11Y-P1-01 revoked session after prior authentication announces session end and supports re-authentication",
  "FE-A11Y-P1-01 generic public error envelope renders safe diagnostics and keyboard error recovery",
] as const;
const p2AccessibilityScenarioTitles = [
  "FE-A11Y-P2-01 Verify shell regions, tabs, switchers, menus, inspector controls, and status strip are keyboard reachable, visibly focused, and named.",
] as const;
const p3AccessibilityScenarioTitles = [
  "FE-A11Y-P3-01 Verify grid cells, editors, group rows, active cell, edit mode, disabled/read-only state, and blocked actions are keyboard accessible and announced without color-only signals.",
] as const;
const p4AccessibilityScenarioTitles = [
  "FE-A11Y-P4-01 Verify grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and Esc priority are keyboard and screen-reader safe.",
] as const;
const p5AccessibilityScenarioTitles = [
  "FE-A11Y-P5-01 Verify mention chip states and manual-resolution controls have accessible names, visible focus, and non-color-only distinction.",
] as const;
const p6AccessibilityScenarioTitles = [
  "FE-A11Y-P6-01 Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have names, focus, contrast, and non-color-only distinctions.",
] as const;
const p7AccessibilityScenarioTitles = [
  "FE-A11Y-P7-01 Verify conflict state, resolver controls, presence hint, stale-row notice, and save-state conflict communicate state by accessible name/state, not color alone.",
] as const;
const p8AccessibilityScenarioTitles = [
  "FE-A11Y-P8-01 Verify sort, filter, group, saved-view menu, active chips, group expand-collapse, and default/startup controls are keyboard reachable and announced.",
] as const;
const p9AccessibilityScenarioTitles = [
  "FE-A11Y-P9-01 Verify inspector tabs, relationship links, evidence controls, history controls, rollback, destructive actions, and errors are keyboard reachable and announced.",
] as const;
const p9ConfigAccessibilityScenarioTitles = [
  "FE-A11Y-P9-02 Verify keyboard open/close, panel navigation, Esc, focus restoration, disabled/blocked states, no-row empty state, and destructive confirmation focus for config-driven inspector behavior.",
] as const;
const p10AccessibilityScenarioTitles = [
  "FE-A11Y-P10-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.",
] as const;
const p11AccessibilityScenarioTitles = [
  "FE-A11Y-P11-01 Verify global accessibility matrix for keyboard access, visible focus, System views, grid navigation/edit entry/exit, Esc, ARIA states, icon-only labels, contrast, and non-color-only empty/loading/error/blocked states.",
] as const;

if (p2AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P2-01 must declare exactly 1 scenario; found ${p2AccessibilityScenarioTitles.length}`,
  );
}
if (p3AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P3-01 must declare exactly 1 scenario; found ${p3AccessibilityScenarioTitles.length}`,
  );
}
if (p4AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P4-01 must declare exactly 1 scenario; found ${p4AccessibilityScenarioTitles.length}`,
  );
}
if (p5AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P5-01 must declare exactly 1 scenario; found ${p5AccessibilityScenarioTitles.length}`,
  );
}
if (p6AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P6-01 must declare exactly 1 scenario; found ${p6AccessibilityScenarioTitles.length}`,
  );
}
if (p7AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P7-01 must declare exactly 1 scenario; found ${p7AccessibilityScenarioTitles.length}`,
  );
}
if (p8AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P8-01 must declare exactly 1 scenario; found ${p8AccessibilityScenarioTitles.length}`,
  );
}
if (p9AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P9-01 must declare exactly 1 scenario; found ${p9AccessibilityScenarioTitles.length}`,
  );
}
if (p9ConfigAccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P9-02 must declare exactly 1 scenario; found ${p9ConfigAccessibilityScenarioTitles.length}`,
  );
}
if (p10AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P10-01 must declare exactly 1 scenario; found ${p10AccessibilityScenarioTitles.length}`,
  );
}
if (p11AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P11-01 must declare exactly 1 scenario; found ${p11AccessibilityScenarioTitles.length}`,
  );
}

const A11yAppLocalSelectors = Object.freeze({
  incidentPatchButton: {
    owner: "apps/web incident administration",
    reason:
      "Incident patch controls are app-local to the incident admin panel until later incident-surface selector promotion.",
    scope: "FE-P1 selected-incident accessibility recovery path",
    testId: "incident-patch-button" as AuthA11yAppLocalTestId,
  },
});

const keyboardSentinelId = "a11y-keyboard-sentinel";
const contrastThreshold = 4.5;

const privateDiagnosticPatterns = [
  /bootstrap[_ -]?token/i,
  /secret[_ -]?base32/i,
  /PRIVATESECRET|ERRORSECRET|CREATESECRET|MFAREQUIREDSECRET/i,
  /otpauth/i,
  /request[_ -]?id/i,
  /\bstack\b/i,
  /\btraceback\b/i,
  /\/(?:home|var|tmp|usr|app|workspace)\//i,
  /\bselect\b[\s\S]{0,80}\bfrom\b/i,
  /\binsert\b[\s\S]{0,80}\binto\b/i,
  /\bupdate\b[\s\S]{0,80}\bset\b/i,
  /internal_path/i,
  /raw_request/i,
];

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}

async function expectAllInteractiveControlsNamed(page: Page) {
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

async function activeTestId(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement)) {
      return "";
    }
    return (
      active.getAttribute("data-testid") ??
      active.querySelector("[data-testid]")?.getAttribute("data-testid") ??
      active.closest("[data-testid]")?.getAttribute("data-testid") ??
      ""
    );
  });
}

async function activeElementOwnTestId(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement)) {
      return "";
    }
    return active.getAttribute("data-testid") ?? "";
  });
}

async function activeElementSignature(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement) || active === document.body) {
      return "";
    }
    const testId =
      active.closest("[data-testid]")?.getAttribute("data-testid") ?? "";
    const id = active.id === "" ? "" : `#${active.id}`;
    const tag = active.tagName.toLowerCase();
    const role = active.getAttribute("role") ?? "";
    const nameParts = [
      active.getAttribute("aria-label")?.trim() ?? "",
      active.getAttribute("title")?.trim() ?? "",
      active.textContent?.trim() ?? "",
    ].filter(Boolean);
    return [
      tag,
      id,
      testId === "" ? "" : ` data-testid:${testId}`,
      role === "" ? "" : ` role:${role}`,
      nameParts.length === 0 ? "" : `:${nameParts.join(" ")}`,
    ].join("");
  });
}

async function blurActiveElement(page: Page) {
  await page.evaluate(() => {
    const active = document.activeElement;
    if (active instanceof HTMLElement) {
      active.blur();
    }
  });
}

async function focusKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    let sentinel = document.getElementById(sentinelId);
    if (sentinel === null) {
      sentinel = document.createElement("button");
      sentinel.id = sentinelId;
      sentinel.setAttribute("aria-label", "Keyboard traversal sentinel");
      sentinel.setAttribute("type", "button");
      sentinel.style.position = "fixed";
      sentinel.style.inlineSize = "1px";
      sentinel.style.blockSize = "1px";
      sentinel.style.opacity = "0";
      sentinel.style.pointerEvents = "none";
      document.body.prepend(sentinel);
    }
    sentinel.focus();
  }, keyboardSentinelId);
}

async function removeKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    document.getElementById(sentinelId)?.remove();
  }, keyboardSentinelId);
}

async function expectTabOrderIncludes(
  page: Page,
  testIds: readonly (string | RegExp)[],
  maxTabs = 180,
) {
  await focusKeyboardSentinel(page);
  try {
    const remaining = [...testIds];
    for (let index = 0; index < maxTabs && remaining.length > 0; index += 1) {
      await page.keyboard.press("Tab");
      const active = await activeTestId(page);
      const reachedIndex = remaining.findIndex((target) =>
        typeof target === "string" ? target === active : target.test(active),
      );
      if (reachedIndex >= 0) {
        remaining.splice(reachedIndex, 1);
      }
    }
    expect(remaining.map(String)).toEqual([]);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

async function expectNoFocusTrap(page: Page) {
  await focusKeyboardSentinel(page);
  try {
    await page.keyboard.press("Tab");
    const first = await activeElementSignature(page);
    await page.keyboard.press("Tab");
    const second = await activeElementSignature(page);
    expect(first).not.toBe("");
    expect(second).not.toBe("");
    expect(second).not.toBe(first);
    await page.keyboard.press("Shift+Tab");
    expect(await activeElementSignature(page)).toBe(first);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

async function openFilterPopover(page: Page, viewSchemaId: string) {
  const trigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(viewSchemaId),
  );
  await expect(trigger).toBeVisible();
  const field = page.getByTestId(gridFilterFieldTestId(viewSchemaId));
  if (!(await field.isVisible().catch(() => false))) {
    await trigger.click();
  }
  await expect(field).toBeVisible();
}

async function openSavedViewActionMenu(page: Page, viewSchemaId: string) {
  const trigger = page.getByTestId(
    savedViewActionMenuTriggerTestId(viewSchemaId),
  );
  await expect(trigger).toBeVisible();
  const menu = page.getByTestId(savedViewActionMenuTestId(viewSchemaId));
  if (!(await menu.isVisible().catch(() => false))) {
    await trigger.click();
  }
  await expect(menu).toBeVisible();
}

async function expectTabTraversalAdvancesFrom(
  page: Page,
  origin: Locator,
  tabCount = 4,
) {
  await expectVisibleFocus(origin);
  const visited: string[] = [];
  for (let index = 0; index < tabCount; index += 1) {
    await page.keyboard.press("Tab");
    visited.push(await activeElementSignature(page));
  }
  expect(visited.every((signature) => signature !== "")).toBeTruthy();
  expect(new Set(visited).size).toBeGreaterThan(1);
}

async function expectKeyboardFocusReachesTestId(
  page: Page,
  testId: string,
  maxTabs = 20,
) {
  const visited: string[] = [];
  for (let index = 0; index < maxTabs; index += 1) {
    const current = await activeElementOwnTestId(page);
    if (current === testId) {
      return;
    }
    visited.push(current);
    await page.keyboard.press("Tab");
  }
  const current = await activeElementOwnTestId(page);
  if (current === testId) {
    return;
  }
  visited.push(current);
  expect(visited).toContain(testId);
}

async function expectVisibleFocus(locator: Locator) {
  await expect
    .poll(async () => {
      await locator.focus();
      return locator.evaluate((element) => element === document.activeElement);
    })
    .toBeTruthy();
  await expect
    .poll(() =>
      locator.evaluate((element) => {
        const style = window.getComputedStyle(element);
        const outlineVisible =
          style.outlineStyle !== "none" &&
          style.outlineWidth !== "0px" &&
          style.outlineColor !== "transparent";
        const shadowVisible = style.boxShadow !== "none";
        return outlineVisible || shadowVisible;
      }),
    )
    .toBeTruthy();
}

async function mountedGridCell(
  page: Page,
  surface: WorkbookSurface,
  recordId: string,
  fieldKey: string,
): Promise<Locator> {
  await scrollGridCellIntoView({
    cellKey: fieldKey,
    page,
    recordId,
    surface,
  });
  return page.getByTestId(rowCellTestId(recordId, fieldKey));
}

async function mountedGridTarget(
  page: Page,
  surface: WorkbookSurface,
  targetTestId: string,
): Promise<Locator> {
  await scrollGridTargetIntoView({ page, surface, targetTestId });
  return page.getByTestId(targetTestId);
}

function semanticGridCell(content: Locator): Locator {
  return content.locator("xpath=ancestor::*[@role='gridcell'][1]");
}

async function expectVisibleSemanticGridCellFocus(
  content: Locator,
): Promise<Locator> {
  const identity = await content.evaluate((element) => ({
    fieldKey: element
      .closest<HTMLElement>("[data-grid-field-key]")
      ?.getAttribute("data-grid-field-key"),
    rowTestId: element
      .closest<HTMLElement>('[role="row"][data-testid]')
      ?.getAttribute("data-testid"),
  }));
  if (
    typeof identity.fieldKey !== "string" ||
    typeof identity.rowTestId !== "string"
  ) {
    throw new Error("Semantic grid cell is missing its stable test identity");
  }
  const cell = content
    .page()
    .getByTestId(identity.rowTestId)
    .getByRole("gridcell")
    .filter({
      has: content
        .page()
        .locator(`[data-grid-field-key="${identity.fieldKey}"]`),
    });
  await content.click();
  const editor = cell.locator(
    "input:not([type='checkbox']), textarea, select, [contenteditable='true']",
  );
  await expect
    .poll(async () => {
      if ((await editor.count()) > 0) return "editor";
      return cell.evaluate((element) =>
        element === document.activeElement ? "cell" : "pending",
      );
    })
    .not.toBe("pending");
  if ((await editor.count()) > 0) {
    await editor.first().press("Escape");
    await expect(editor).toHaveCount(0);
  }
  await expectVisibleFocus(cell);
  return cell;
}

async function activateTimelineGridEditor(
  page: Page,
  recordId: string,
  fieldKey: string,
): Promise<Locator> {
  const display = await mountedGridCell(
    page,
    timelineViewSchemaId,
    recordId,
    fieldKey,
  );
  await display.click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey,
      recordId,
      surface: "grid",
    }),
  );
  await expectVisibleFocus(editor);
  return editor;
}

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toContainText(
    roleText,
  );
  await accountMenuTrigger.click();
}

async function expectStatusRole(
  locator: Locator,
  options: { readonly visible?: boolean } = {},
) {
  if (options.visible === false) {
    await expect(locator).toHaveCount(1);
  } else {
    await expect(locator).toBeVisible();
  }
  await expect(locator).toHaveAttribute("role", "status");
  await expect(locator).not.toHaveText("");
}

async function expectAlertRole(locator: Locator) {
  await expect(locator).toBeVisible();
  await expect(locator).toHaveAttribute("role", "alert");
  await expect(locator).not.toHaveText("");
}

async function expectNoPrivateDiagnostics(locator: Locator) {
  const text = (await locator.textContent()) ?? "";
  for (const pattern of privateDiagnosticPatterns) {
    expect(text).not.toMatch(pattern);
  }
}

function contrastRecordPath(title: string) {
  const dir = process.env.CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR;
  if (!dir) {
    return "";
  }
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 120);
  mkdirSync(dir, { recursive: true });
  return path.join(dir, `${slug}.json`);
}

async function collectContrastChecks(page: Page, testIds: readonly string[]) {
  const targets = [...new Set(testIds)].map((id) => ({
    id,
    selector: dataTestIdSelector(id),
  }));
  return page.evaluate(
    ({ targets, threshold }) => {
      type Rgba = { a: number; b: number; g: number; r: number };

      function parseRgba(value: string): Rgba | null {
        const match =
          /^rgba?\((\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?)(?:,\s*(\d+(?:\.\d+)?))?\)$/i.exec(
            value.trim(),
          );
        if (!match) {
          return null;
        }
        return {
          r: Number(match[1]),
          g: Number(match[2]),
          b: Number(match[3]),
          a: match[4] === undefined ? 1 : Number(match[4]),
        };
      }

      function rgbText(color: Rgba) {
        return `rgb(${Math.round(color.r)}, ${Math.round(color.g)}, ${Math.round(color.b)})`;
      }

      function channel(value: number) {
        const normalized = value / 255;
        if (normalized <= 0.03928) {
          return normalized / 12.92;
        }
        return ((normalized + 0.055) / 1.055) ** 2.4;
      }

      function luminance(color: Rgba) {
        return (
          0.2126 * channel(color.r) +
          0.7152 * channel(color.g) +
          0.0722 * channel(color.b)
        );
      }

      function contrastRatio(foreground: Rgba, background: Rgba) {
        const lighter = Math.max(luminance(foreground), luminance(background));
        const darker = Math.min(luminance(foreground), luminance(background));
        return (lighter + 0.05) / (darker + 0.05);
      }

      function backgroundFor(element: Element) {
        let candidate: Element | null = element;
        while (candidate) {
          const color = parseRgba(
            window.getComputedStyle(candidate).backgroundColor,
          );
          if (color && color.a > 0) {
            return color;
          }
          candidate = candidate.parentElement;
        }
        return { a: 1, b: 255, g: 255, r: 255 };
      }

      return targets
        .map(({ id, selector }) => {
          const element = document.querySelector(selector);
          if (!(element instanceof HTMLElement)) {
            return null;
          }
          const style = window.getComputedStyle(element);
          if (
            style.display === "none" ||
            style.visibility === "hidden" ||
            element.getClientRects().length === 0
          ) {
            return null;
          }
          const foreground = parseRgba(style.color);
          const background = backgroundFor(element);
          if (!foreground) {
            return null;
          }
          const ratio = contrastRatio(foreground, background);
          return {
            background: rgbText(background),
            foreground: rgbText(foreground),
            ratio: Math.round(ratio * 100) / 100,
            result: ratio >= threshold ? "pass" : "fail",
            target: id,
            threshold,
          };
        })
        .filter(Boolean);
    },
    { targets, threshold: contrastThreshold },
  );
}

async function expectAndRecordContrast(page: Page, testIds: readonly string[]) {
  const title = test.info().title;
  const checks = await collectContrastChecks(page, testIds);
  expect(checks.length).toBeGreaterThan(0);
  expect(checks.filter((check) => check?.result !== "pass")).toEqual([]);

  const recordPath = contrastRecordPath(title);
  if (recordPath) {
    writeFileSync(
      recordPath,
      `${JSON.stringify(
        {
          scenario_title: title,
          checks,
        },
        null,
        2,
      )}\n`,
      "utf8",
    );
  }
}

async function expectCellTextOrValue(locator: Locator, value: string) {
  const mode = await locator.evaluate((element) => {
    if (
      element instanceof HTMLInputElement ||
      element instanceof HTMLTextAreaElement ||
      element instanceof HTMLSelectElement
    ) {
      return "value";
    }
    return "text";
  });
  if (mode === "value") {
    await expect(locator).toHaveValue(value);
    return;
  }
  await expect(locator).toContainText(value);
}

async function openA11ySystemSurface(
  page: Page,
  options: {
    groupToken: SystemViewSwitcherGroupToken;
    viewSchemaId: string;
  },
) {
  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expectVisibleFocus(trigger);
  await trigger.press("Enter");
  await expect(trigger).toHaveAttribute("aria-expanded", "true");
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(menu).toHaveAttribute("role", "menu");

  const option = page.getByTestId(
    systemViewSwitcherOptionTestId(options.groupToken, options.viewSchemaId),
  );
  await expect(option).toHaveAttribute("role", "menuitemradio");
  await expect(option).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(option).not.toHaveText("");
  const menuOptions = menu.getByRole("menuitemradio");
  await expect
    .poll(() =>
      menuOptions.evaluateAll((entries) => {
        return entries.some((entry) => entry === document.activeElement);
      }),
    )
    .toBeTruthy();
  const optionCount = await menuOptions.count();
  const activeIndex = await menuOptions.evaluateAll((entries) => {
    const activeElement = document.activeElement;
    if (
      !(activeElement instanceof HTMLElement) &&
      !(activeElement instanceof SVGElement)
    ) {
      return -1;
    }
    return entries.indexOf(activeElement);
  });
  const targetIndex = await menuOptions.evaluateAll(
    (entries, viewSchemaId) =>
      entries.findIndex(
        (entry) => entry.getAttribute("data-view-schema-id") === viewSchemaId,
      ),
    options.viewSchemaId,
  );
  expect(activeIndex).toBeGreaterThanOrEqual(0);
  expect(targetIndex).toBeGreaterThanOrEqual(0);
  const arrowCount = (targetIndex - activeIndex + optionCount) % optionCount;
  for (let index = 1; index <= arrowCount; index += 1) {
    await page.keyboard.press("ArrowDown");
    await expect(
      menuOptions.nth((activeIndex + index) % optionCount),
    ).toBeFocused();
  }
  await expect(option).toBeFocused();
  await expectVisibleFocus(option);
  await option.press("Enter");
  await expect(menu).toHaveCount(0);
  await expect(trigger).toHaveAttribute("aria-expanded", "false");
  await expect(trigger).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(
    page.getByTestId(gridShellTestId(options.viewSchemaId)),
  ).toBeVisible();
}

function A11yAttachedEvidencePayload(recordId: string) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_record_ref",
        linked_record_id: recordId,
      },
    ],
  };
}

function A11yHistoryActionTestId(
  item: A11yHistoryItem,
  action: A11yHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function A11yRollbackAnchor(
  item: A11yHistoryItem,
  action: A11yHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireA11yHistoryEntryAction(history: A11yHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error("missing FE-A11Y-P9 history_entry rollback item");
  }
  return item;
}

async function fetchA11yRecordHistory(page: Page, recordId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/records/${recordId}/history`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: A11yHistoryData }).data;
}

async function expectP1SurfaceA11y(
  page: Page,
  options: {
    focusTestId?: string;
    tabStops?: readonly string[];
  } = {},
) {
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  if (options.focusTestId) {
    await expectVisibleFocus(page.getByTestId(options.focusTestId));
  }
  if (options.tabStops && options.tabStops.length > 0) {
    await expectTabOrderIncludes(page, options.tabStops);
  }
  await expectAndRecordContrast(page, [
    ...(options.focusTestId ? [options.focusTestId] : []),
    ...(options.tabStops ?? []),
  ]);
}

function authA11yAppLocalTestId(
  key: keyof typeof A11yAppLocalSelectors,
): AuthA11yAppLocalTestId {
  const entry = A11yAppLocalSelectors[key];
  if (
    entry.owner.trim() === "" ||
    entry.reason.trim() === "" ||
    entry.scope.trim() === ""
  ) {
    throw new Error(`missing app-local selector ownership for ${key}`);
  }
  return entry.testId;
}

async function loadIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
) {
  const response = await authRequests.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function createIncidentWithRequest(
  authRequests: APIRequestContext,
  incidentKey: string,
  title: string,
) {
  const response = await authRequests.post(`${apiBase}/api/v1/incidents`, {
    data: {
      client_txn_id: uniqueTxn("a11y-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function createIncidentMembershipWithRequest(
  authRequests: APIRequestContext,
  incidentId: string,
  email: string,
  role: string,
) {
  const response = await authRequests.post(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    {
      data: {
        client_txn_id: uniqueTxn("a11y-membership"),
        email,
        role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function patchIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
  headers: Record<string, string> = {},
) {
  const response = await authRequests.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      headers,
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: IncidentMembershipRecord }).data;
}

async function deleteIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
  baseMembershipVersion: number,
  headers: Record<string, string> = {},
) {
  const response = await authRequests.delete(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${userId}`,
    {
      headers,
      data: {
        base_membership_version: baseMembershipVersion,
      },
    },
  );
  expect(response.status()).toBe(204);
}

type EvidenceA11yRowStateKey =
  | "available"
  | "blocked"
  | "pending_upload"
  | "requested";

async function createA11yEvidenceRow(
  page: Page,
  incidentId: string,
  options: {
    lifecycleState: string;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "FE-P6 accessibility fixture",
    ...options,
  });
}

async function createUploadedA11yEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "FE-P6 accessibility fixture",
    ...options,
  });
}

async function expectEvidenceAccessState(
  page: Page,
  recordId: string,
  stateKey: EvidenceA11yRowStateKey,
  options: {
    liveRole?: "alert" | "status";
    messageText?: RegExp | string;
  } = {},
) {
  await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidencePreviewButtonTestId(recordId),
  );
  await expect(evidenceAccessStateContainer(page, recordId)).toHaveAttribute(
    "data-evidence-state-key",
    stateKey,
  );
  const previewButton = page.getByTestId(evidencePreviewButtonTestId(recordId));
  const downloadButton = page.getByTestId(
    evidenceDownloadButtonTestId(recordId),
  );
  await expect(previewButton).toHaveText("Preview");
  await expect(downloadButton).toHaveText("Download");
  if (stateKey === "available") {
    await expect(previewButton).toBeEnabled();
    await expect(downloadButton).toBeEnabled();
  } else {
    await expect(previewButton).toBeDisabled();
    await expect(downloadButton).toBeDisabled();
  }
  if (options.messageText !== undefined || options.liveRole !== undefined) {
    const message = page.getByTestId(evidenceAccessMessageTestId(recordId));
    await expect(message).toBeVisible();
    if (options.messageText !== undefined) {
      await expect(message).toContainText(options.messageText);
    }
    if (options.liveRole !== undefined) {
      await expect(message).toHaveAttribute("role", options.liveRole);
      await expect(message).toHaveAttribute(
        "aria-live",
        options.liveRole === "alert" ? "assertive" : "polite",
      );
    }
  }
}

function evidenceAccessStateContainer(page: Page, recordId: string): Locator {
  return page
    .getByTestId(evidencePreviewButtonTestId(recordId))
    .locator("xpath=ancestor::*[@data-evidence-state-key][1]");
}

async function armA11yPublicErrorFault(
  page: Page,
  options: {
    path: string;
    reasonCode: "blob_failed" | "evidence_inconsistent";
  },
) {
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    body: {
      code: "evidence_access_unavailable",
      consume_once: true,
      details: {
        reason_code: options.reasonCode,
      },
      message: "Evidence access failed for FE-P6 accessibility fixture.",
      method: "POST",
      path: options.path,
      retryable: false,
      status: 409,
    },
    method: "POST",
    path: "/api/v1/test/runtime/public-error-faults",
  });
  expect(response.status).toBe(201);
}

async function holdSinglePublicAPIResponse(
  page: Page,
  options: {
    method: string;
    path: string;
  },
) {
  const routePattern = `**${options.path}`;
  const expectedMethod = options.method.toUpperCase();
  let releaseHold: (() => void) | null = null;
  let resolveHit: (() => void) | null = null;
  const waitForHit = new Promise<void>((resolve) => {
    resolveHit = resolve;
  });
  const hold = new Promise<void>((resolve) => {
    releaseHold = resolve;
  });
  let released = false;

  const routeHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== expectedMethod || released) {
      await route.fallback();
      return;
    }
    released = true;
    resolveHit?.();
    await hold;
    const response = await route.fetch();
    await route.fulfill({ response });
  };

  await page.route(routePattern, routeHandler);
  return {
    dispose: async () => {
      releaseHold?.();
      await safeUnroute(page, routePattern, routeHandler);
    },
    release: () => {
      releaseHold?.();
    },
    waitForHit,
  };
}

async function fulfillPublicError(
  route: Route,
  options: {
    code: string;
    details?: Record<string, unknown>;
    message?: string;
    status: number;
  },
) {
  await route.fulfill({
    contentType: "application/json",
    status: options.status,
    body: JSON.stringify({
      error: {
        code: options.code,
        message: options.message,
        status: options.status,
        request_id: "raw-request-id-must-not-render",
        details: {
          reason_code: "safe_reason",
          ...(options.details ?? {}),
          bootstrap_token: "private-bootstrap-token",
          internal_path: "/home/cartulary/private/file.go",
          raw_request_id: "raw-request-id-must-not-render",
          secret_base32: "PRIVATESECRETBASE32",
          sql: "select * from private_table",
          stack: "stack frame at /home/cartulary/private/file.go:1",
        },
      },
    }),
  });
}

test.describe("FE-P2 accessibility readiness", () => {
  test(p2AccessibilityScenarioTitles[0], async ({ page }) => {
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
        "timeline.activity_utc_text": "2026-05-31T09:00:00Z",
        "timeline.activity_synopsis_text": "FE-P2 accessibility shell row",
        "timeline.raw_activity_text": "Inspector control coverage",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    await expect(
      page.getByRole("region", { name: "Workbook shell" }),
    ).toHaveCount(1);

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
    ).toHaveAttribute("aria-current", "page");

    const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
    await expect(trigger).toBeVisible();
    await expect(trigger).toHaveAttribute("aria-label", "System views");
    await expectVisibleFocus(trigger);
    await page.keyboard.press("Enter");

    const menu = page.getByTestId(systemViewSwitcherMenuTestId());
    await expect(menu).toBeVisible();
    await expect(menu).toHaveAttribute("role", "menu");
    const indicatorOption = page.getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        indicatorsViewSchemaId,
      ),
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
      .getByRole("dialog", { name: "Draft filters" })
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
      page.getByTestId(
        rowInspectorFieldTestId(
          timelineRow.record_id,
          "timeline.raw_activity_text",
        ),
      ),
    ).toBeVisible();

    const saveState = page.getByTestId(saveStateTestId());
    await expectStatusRole(saveState);
    await expect(saveState).toHaveText("Saved");

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      workbookShellSlotTestId("top-bar"),
      workbookShellSlotTestId("status-strip"),
      systemViewSwitcherTriggerTestId(),
      savedViewSelectorTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
      rowCellTestId(timelineRow.record_id, "timeline.activity_synopsis_text"),
      saveStateTestId(),
    ]);
  });
});

if (
  (process.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID ?? "default") ===
  "network_flow_claimed"
) {
  test("FE-A11Y-P12-01 Verify claimed Network Analysis tabs, query controls, semantic grids, inspector, graph, contributor drawer, mapping modal, focus return, names, and ARIA evidence.", async ({
    page,
  }) => {
    await openClaimedNetworkAnalysis(page, "FEP12A11Y");
    await importNetworkFlowCSV(page, { displayName: "accessible-flow" });

    const workspace = page.getByTestId(networkAnalysisTestId("workspace"));
    await expect(workspace).toHaveAttribute(
      "data-extension-profile-id",
      "network_flow_activity",
    );
    await expect(
      page.getByRole("tab", { name: /accessible-flow/ }),
    ).toHaveAttribute("aria-selected", "true");
    await expect(
      page.getByRole("region", { name: "Network Flow filters" }),
    ).toBeVisible();
    await expect(
      page.getByRole("grid", { name: "Accepted Network Flow rows" }),
    ).toBeVisible();
    const semanticCell = page
      .getByRole("grid", { name: "Accepted Network Flow rows" })
      .locator(
        '[role="gridcell"] [data-grid-field-key]:not([data-grid-field-key^="__"])',
      )
      .first();
    const focusedGridCell = semanticGridCell(semanticCell);
    await semanticCell.click();
    await expect(
      page.getByRole("complementary", { name: "Network Flow cell inspector" }),
    ).toBeVisible();
    await page.getByTestId(networkAnalysisTestId("inspector-close")).click();
    await expect(focusedGridCell).toBeFocused();

    const graphMode = page.getByTestId(networkAnalysisTestId("mode-graph"));
    await graphMode.focus();
    await expectVisibleFocus(graphMode);
    await graphMode.press("Enter");
    const edgeSelect = page
      .getByTestId(/^network-flow-edge-/)
      .first()
      .getByRole("button", { name: "Select edge" });
    await expect(edgeSelect).toBeVisible();
    await edgeSelect.click();
    const drawer = page.getByTestId(
      networkAnalysisTestId("contributor-drawer"),
    );
    await expect(drawer).toBeVisible();
    await expect(drawer.getByRole("button", { name: /close/i })).toBeVisible();
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      networkAnalysisTestId("workspace"),
      networkAnalysisTestId("mode-graph"),
      networkAnalysisTestId("contributor-close"),
    ]);
    await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
    await expect(edgeSelect).toBeFocused();
  });
}

test.describe("FE-P3 accessibility readiness", () => {
  test(p3AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP3"),
      "FE-A11Y-P3-01 grid adapter",
    );
    const alphaRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p3-01-alpha"),
        "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
        "timeline.activity_synopsis_text": "Alpha accessibility row",
        "timeline.raw_activity_text": "Keyboard grid coverage",
      },
    )) as ViewRow;
    const betaRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p3-01-beta"),
        "timeline.activity_utc_text": "2026-05-31T10:05:00Z",
        "timeline.activity_synopsis_text": "Beta accessibility row",
        "timeline.raw_activity_text": "Grouped grid coverage",
      },
    )) as ViewRow;

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
    await expect(reviewedGroup).toContainText("reviewed");

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
    await expectStatusRole(page.getByTestId(saveStateTestId()));

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
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
      rowCellTestId(betaRow.record_id, "timeline.activity_synopsis_text"),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P4 accessibility readiness", () => {
  test(p4AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP4"),
      "FE-A11Y-P4-01 Timeline accessibility",
    );
    const editRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-edit"),
        "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
        "timeline.activity_synopsis_text": "FE-P4 edit accessibility row",
        "timeline.raw_activity_text": "Escape priority details",
      },
    )) as ViewRow;
    const pasteRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-paste"),
        "timeline.activity_utc_text": "2026-06-03T10:05:00Z",
        "timeline.activity_synopsis_text": "FE-P4 paste accessibility row",
      },
    )) as ViewRow;
    const pendingRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-pending"),
        "timeline.activity_utc_text": "2026-06-03T10:10:00Z",
        "timeline.activity_synopsis_text": "FE-P4 pending accessibility row",
      },
    )) as ViewRow;
    const validationRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-validation"),
        "timeline.activity_utc_text": "2026-06-03T10:15:00Z",
        "timeline.activity_synopsis_text": "FE-P4 validation accessibility row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expectStatusRole(page.getByTestId(saveStateTestId()));
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      editRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expectTabOrderIncludes(page, [
      workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridShellTestId(timelineViewSchemaId),
    ]);

    const editSummary = await activateTimelineGridEditor(
      page,
      editRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expect(editSummary).toHaveAttribute(
      "aria-label",
      `Activity Synopsis ${editRow.record_id}`,
    );
    await editSummary.fill("FE-P4 accessibility committed edit");
    await editSummary.press("Enter");
    await expect(
      page.getByTestId(
        rowCellTestId(editRow.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("FE-P4 accessibility committed edit");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    await pasteGridMatrix({
      fieldKey: "timeline.activity_synopsis_text",
      matrix: [["FE-P4 accessibility pasted summary", "a11y-host.example"]],
      page,
      recordId: pasteRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        rowCellTestId(pasteRow.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("FE-P4 accessibility pasted summary");

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      const pendingSummary = await activateTimelineGridEditor(
        page,
        pendingRow.record_id,
        "timeline.activity_synopsis_text",
      );
      await pendingSummary.fill("FE-P4 accessibility pending replay");
      await pendingSummary.press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await expectStatusRole(page.getByTestId(pendingQueueNoticeTestId()));
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );

      patchController.connect();
      await expect
        .poll(() => successfulPatchCalls(patchController.calls).length)
        .toBe(1);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    } finally {
      patchController.connect();
      await patchController.dispose();
    }

    const originSummary = await mountedGridCell(
      page,
      timelineViewSchemaId,
      editRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expectVisibleSemanticGridCellFocus(originSummary);
    await openTimelineInspector(page, editRow.record_id);
    const inspectorOriginSummary = await mountedGridCell(
      page,
      timelineViewSchemaId,
      editRow.record_id,
      "timeline.activity_synopsis_text",
    );
    const originSummaryCell = semanticGridCell(inspectorOriginSummary);
    await semanticGridCell(inspectorOriginSummary).focus();
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${timelineViewSchemaId}:${editRow.record_id}:timeline.activity_synopsis_text`,
    );
    const inspectorDetails = page.getByTestId(
      rowInspectorFieldTestId(editRow.record_id, "timeline.raw_activity_text"),
    );
    await expectVisibleFocus(inspectorDetails);
    await page.keyboard.press("Escape");
    await expect(originSummaryCell).toBeFocused();

    const validationCell = await activateTimelineGridEditor(
      page,
      validationRow.record_id,
      "timeline.activity_utc_text",
    );
    await validationCell.fill("not-a-timestamp");
    await validationCell.press("Enter");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await expect(
      page.getByTestId(
        rowCellTestId(validationRow.record_id, "timeline.activity_utc_text"),
      ),
    ).toHaveText("not-a-timestamp");

    const recoveryController = await installPatchController(page);
    try {
      const closeInspector = page.getByTestId(
        workbookInspectorCloseButtonTestId(timelineViewSchemaId),
      );
      if (await closeInspector.isVisible()) {
        await closeInspector.click();
        await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(
          0,
        );
      }
      recoveryController.failNextPatch(409, "client_txn_conflict", {
        recordId: pendingRow.record_id,
      });
      const blockedSummary = await activateTimelineGridEditor(
        page,
        pendingRow.record_id,
        "timeline.activity_synopsis_text",
      );
      await blockedSummary.fill("FE-P4 accessibility blocked edit");
      await blockedSummary.press("Enter");
      await expect.poll(() => recoveryController.calls.length).toBe(1);

      const recoveryPanel = page.getByTestId(pendingQueueRecoveryPanelTestId());
      const retryButton = page.getByTestId(pendingQueueRetryButtonTestId());
      const discardButton = page.getByTestId(pendingQueueDiscardButtonTestId());
      await expect(recoveryPanel).toHaveRole("region");
      await expect(recoveryPanel).toHaveAccessibleName("Queued edits");
      await expect(recoveryPanel).not.toBeFocused();
      await expect(retryButton).toHaveAccessibleName(
        "Retry with a new request ID",
      );
      await expect(discardButton).toHaveAccessibleName("Discard blocked edit");
      await expect(retryButton).toBeEnabled();
      await expect(discardButton).toBeEnabled();

      await page.getByTestId(saveStateActionButtonTestId()).click();
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeFocused();
      await retryButton.focus();
      await page.keyboard.press("Tab");
      await expect(discardButton).toBeFocused();

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
        workbookShellSlotTestId("status-strip"),
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
        gridGroupingSelectTestId(timelineViewSchemaId),
        gridSortHeaderTestId(
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        ),
        rowCellTestId(editRow.record_id, "timeline.activity_synopsis_text"),
        rowCellTestId(validationRow.record_id, "timeline.activity_utc_text"),
        pendingQueueNoticeTestId(),
        pendingQueueRetryButtonTestId(),
        pendingQueueDiscardButtonTestId(),
        saveStateTestId(),
      ]);

      await discardButton.press("Space");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(recoveryPanel).toHaveCount(0);
      expect(recoveryController.calls).toHaveLength(1);
    } finally {
      await recoveryController.dispose();
    }
  });
});

test.describe("FE-P5 accessibility readiness", () => {
  test(p5AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP5"),
      "FE-A11Y-P5-01 mention states",
    );
    const {
      autoRawText,
      autoRow,
      dismissedMention,
      dismissedRawText,
      dismissedRow,
      manualMention,
      manualRow,
      manualTarget,
      resolvedMention,
      resolvedRow,
      unresolvedMention,
      unresolvedRawText,
      unresolvedRow,
    } = await seedHostMentionStateFixture(page, incidentId, {
      displayPrefix: "FE-A11Y-P5",
      hostnamePrefix: "fe-a11y-p5",
      occurredAt: {
        auto: "2026-06-06T16:15:00Z",
        dismissed: "2026-06-06T16:20:00Z",
        manual: "2026-06-06T16:10:00Z",
        resolved: "2026-06-06T16:05:00Z",
        unresolved: "2026-06-06T16:00:00Z",
      },
      rawTextPrefix: "FEA11YP5",
      summary: {
        auto: "FE-A11Y-P5 auto chip",
        dismissed: "FE-A11Y-P5 dismissed chip",
        manual: "FE-A11Y-P5 manual chip",
        resolved: "FE-A11Y-P5 resolved chip",
        unresolved: "FE-A11Y-P5 unresolved chip",
      },
      txnPrefix: "fe-a11y-p5",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    await openTimelineInspector(page, unresolvedRow.record_id);
    const unresolvedChip = page
      .getByTestId(
        relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(unresolvedMention.item_ref)));
    await expect(unresolvedChip).toHaveAttribute(
      "aria-label",
      `Unresolved ${unresolvedRawText}`,
    );
    await expect(unresolvedChip).toContainText("Unresolved");
    await expectVisibleFocus(unresolvedChip);

    await openTimelineInspector(page, resolvedRow.record_id);
    const resolvedChip = page
      .getByTestId(
        relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(resolvedMention.item_ref)));
    await expect(resolvedChip).toHaveAttribute(
      "aria-label",
      /^Resolved FE-A11Y-P5 Resolved Target$/u,
    );
    await expect(resolvedChip).toContainText("Resolved");
    await expectVisibleFocus(resolvedChip);

    await openTimelineInspector(page, manualRow.record_id);
    const manualMentionItem = page.getByTestId(
      mentionItemTestId(String(manualMention.item_ref)),
    );
    await expectVisibleFocus(manualMentionItem);
    await manualMentionItem.click();
    await manualMentionItem.focus();
    const resolveSelect = page.getByTestId(mentionResolveTargetSelectTestId());
    await expect(resolveSelect).toHaveAccessibleName("Resolve to existing");
    await expectKeyboardFocusReachesTestId(
      page,
      mentionResolveTargetSelectTestId(),
    );
    await expectVisibleFocus(resolveSelect);
    await expect(resolveSelect).toBeEnabled();
    await resolveSelect.selectOption(manualTarget.record_id);
    const resolveButton = page.getByTestId(
      mentionResolveExistingButtonTestId(),
    );
    await expectVisibleFocus(resolveButton);
    await resolveButton.click();

    const manualChip = page
      .getByTestId(
        relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(manualMention.item_ref)));
    await expect(manualChip).toHaveAttribute(
      "aria-label",
      /^Resolved .*manual resolution/u,
    );
    await expect(manualChip).toContainText("Manual");
    await expectVisibleFocus(manualChip);
    await expectVisibleFocus(page.getByTestId(mentionDismissButtonTestId()));
    await expectVisibleFocus(
      page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    );

    const autoEnvelope = await addRelationshipTokenViaUI(
      page,
      autoRow.record_id,
      "hostRefs",
      autoRawText,
    );
    const autoItem = requireItemByRawText(
      collectionItems(autoEnvelope.data.row, hostRefsFieldKey),
      autoRawText,
    );
    const autoChip = page
      .getByTestId(relationshipItemsTestId(autoRow.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(autoItem.item_ref)));
    await expect(autoChip).toHaveAttribute(
      "aria-label",
      /^Auto-resolved .*matched/u,
    );
    await expect(autoChip).toContainText("Auto");
    await expectVisibleFocus(autoChip);
    const autoNotice = page.getByTestId(
      autoResolutionNoticeTestId(String(autoItem.item_ref)),
    );
    await expect(autoNotice).toContainText("FE-A11Y-P5 Auto Target");
    await expectVisibleFocus(
      autoNotice.getByTestId(
        autoResolutionUndoButtonTestId(String(autoItem.item_ref)),
      ),
    );

    await openTimelineInspector(page, dismissedRow.record_id);
    await page
      .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
      .click();
    await page
      .getByTestId(mentionResolveTargetSelectTestId())
      .selectOption(manualTarget.record_id);
    await page.getByTestId(mentionResolveExistingButtonTestId()).click();
    await page.getByTestId(mentionDismissButtonTestId()).click();
    const dismissedMentionItem = page
      .getByTestId(timelineInspectorTestId())
      .getByRole("button", { name: `Dismissed ${dismissedRawText}` });
    await expect(dismissedMentionItem).toBeVisible();
    await expect(dismissedMentionItem).toContainText("Dismissed");
    await expectVisibleFocus(dismissedMentionItem);
    await expectVisibleFocus(
      page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    );

    for (const recordId of [
      unresolvedRow.record_id,
      resolvedRow.record_id,
      manualRow.record_id,
      autoRow.record_id,
      dismissedRow.record_id,
    ]) {
      await expectVisibleSemanticGridCellFocus(
        await mountedGridCell(
          page,
          timelineViewSchemaId,
          recordId,
          "timeline.activity_synopsis_text",
        ),
      );
    }
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      relationshipChipTestId(String(unresolvedMention.item_ref)),
      relationshipChipTestId(String(resolvedMention.item_ref)),
      relationshipChipTestId(String(manualMention.item_ref)),
      relationshipChipTestId(String(autoItem.item_ref)),
      mentionItemTestId(String(dismissedMention.item_ref)),
      mentionRestoreUnresolvedButtonTestId(),
    ]);
  });
});

test.describe("FE-P6 accessibility readiness", () => {
  test(p6AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP6"),
      "FE-A11Y-P6-01 evidence access",
    );
    const requested = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "requested",
      requestedAt: "2026-06-07T10:00:00Z",
      storageRef: "case://fe-a11y-p6/requested",
      title: "01 requested evidence",
      txnPrefix: "fe-a11y-p6-requested",
    });
    const pending = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "pending_receipt",
      requestedAt: "2026-06-07T10:05:00Z",
      storageRef: "case://fe-a11y-p6/pending",
      title: "02 pending evidence",
      txnPrefix: "fe-a11y-p6-pending",
    });
    const blocked = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "quarantined",
      requestedAt: "2026-06-07T10:10:00Z",
      storageRef: "case://fe-a11y-p6/quarantined",
      title: "03 quarantined evidence",
      txnPrefix: "fe-a11y-p6-blocked",
    });
    const availablePreview = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-A11Y-P6 preview evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-a11y-p6-preview.txt",
        requestedAt: "2026-06-07T10:15:00Z",
        title: "04 available preview evidence",
        txnPrefix: "fe-a11y-p6-preview",
      },
    );
    const downloadHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from("FE-A11Y-P6 download evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "fe-a11y-p6-download.txt",
      requestedAt: "2026-06-07T10:20:00Z",
      title: "05 download handle evidence",
      txnPrefix: "fe-a11y-p6-download",
    });
    const previewBlocked = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from(
        "<!doctype html><title>FE-A11Y-P6 unsupported preview</title>",
        "utf8",
      ),
      contentType: "text/html",
      filename: "fe-a11y-p6-preview-blocked.html",
      requestedAt: "2026-06-07T10:25:00Z",
      title: "06 preview blocked evidence",
      txnPrefix: "fe-a11y-p6-preview-blocked",
    });
    const failedHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from("FE-A11Y-P6 failed handle evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "fe-a11y-p6-failed.txt",
      requestedAt: "2026-06-07T10:30:00Z",
      title: "07 failed handle evidence",
      txnPrefix: "fe-a11y-p6-failed",
    });
    const inconsistentHandle = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-A11Y-P6 inconsistent handle evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-a11y-p6-inconsistent.txt",
        requestedAt: "2026-06-07T10:35:00Z",
        title: "08 inconsistent handle evidence",
        txnPrefix: "fe-a11y-p6-inconsistent",
      },
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();

    await expectEvidenceAccessState(page, requested.record_id, "requested", {
      liveRole: "status",
      messageText: /^Requested:/u,
    });
    await expectEvidenceAccessState(page, pending.record_id, "pending_upload", {
      liveRole: "status",
      messageText: /pending/u,
    });
    await expectEvidenceAccessState(page, blocked.record_id, "blocked", {
      liveRole: "alert",
      messageText: /^Blocked:/u,
    });

    await expectEvidenceAccessState(
      page,
      availablePreview.record_id,
      "available",
    );
    const previewButton = page.getByTestId(
      evidencePreviewButtonTestId(availablePreview.record_id),
    );
    await expectVisibleFocus(
      page.getByTestId(
        evidenceDownloadButtonTestId(availablePreview.record_id),
      ),
    );
    await expectVisibleFocus(previewButton);
    await previewButton.click();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
    ).toHaveAttribute("title", /Evidence preview/u);
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveAttribute("role", "status");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveText("Preview loaded inline.");
    await page
      .getByTestId(evidencePreviewPanelTestId())
      .getByRole("button", { name: "Close" })
      .click();

    await expectEvidenceAccessState(
      page,
      downloadHandle.record_id,
      "available",
    );
    const downloadButton = page.getByTestId(
      evidenceDownloadButtonTestId(downloadHandle.record_id),
    );
    await expectVisibleFocus(
      page.getByTestId(evidencePreviewButtonTestId(downloadHandle.record_id)),
    );
    await expectVisibleFocus(downloadButton);
    const downloadPromise = page.waitForEvent("download");
    await downloadButton.click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe("fe-a11y-p6-download.txt");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveAttribute("role", "status");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveText("Download handle issued.");

    await expectEvidenceAccessState(
      page,
      previewBlocked.record_id,
      "available",
    );
    const unsupportedPreviewButton = page.getByTestId(
      evidencePreviewButtonTestId(previewBlocked.record_id),
    );
    await expectVisibleFocus(unsupportedPreviewButton);
    await expectVisibleFocus(
      page.getByTestId(evidenceDownloadButtonTestId(previewBlocked.record_id)),
    );
    await unsupportedPreviewButton.click();
    const previewBlockedMessage = page.getByTestId(
      evidenceAccessMessageTestId(previewBlocked.record_id),
    );
    await expect(previewBlockedMessage).toHaveAttribute("role", "alert");
    await expect(previewBlockedMessage).toContainText(
      "evidence_access_unavailable: unsupported_preview",
    );
    await expectNoPrivateDiagnostics(previewBlockedMessage);

    await armA11yPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${failedHandle.record_id}/preview-handle`,
      reasonCode: "blob_failed",
    });
    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(failedHandle.record_id),
      )
    ).click();
    const failedMessage = page.getByTestId(
      evidenceAccessMessageTestId(failedHandle.record_id),
    );
    await expect(failedMessage).toHaveAttribute("role", "alert");
    await expect(failedMessage).toContainText(
      "evidence_access_unavailable: blob_failed",
    );
    await expectNoPrivateDiagnostics(failedMessage);

    await armA11yPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${inconsistentHandle.record_id}/preview-handle`,
      reasonCode: "evidence_inconsistent",
    });
    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(inconsistentHandle.record_id),
      )
    ).click();
    const inconsistentMessage = page.getByTestId(
      evidenceAccessMessageTestId(inconsistentHandle.record_id),
    );
    await expect(inconsistentMessage).toHaveAttribute("role", "alert");
    await expect(inconsistentMessage).toContainText(
      "evidence_access_unavailable: evidence_inconsistent",
    );
    await expectNoPrivateDiagnostics(inconsistentMessage);

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      evidenceAccessMessageTestId(requested.record_id),
      evidenceAccessMessageTestId(pending.record_id),
      evidenceAccessMessageTestId(blocked.record_id),
      evidencePreviewButtonTestId(availablePreview.record_id),
      evidenceDownloadButtonTestId(downloadHandle.record_id),
      evidenceAccessMessageTestId(availablePreview.record_id),
      evidenceAccessMessageTestId(downloadHandle.record_id),
      evidenceAccessMessageTestId(previewBlocked.record_id),
      evidenceAccessMessageTestId(failedHandle.record_id),
      evidenceAccessMessageTestId(inconsistentHandle.record_id),
    ]);
  });
});

test.describe("FE-P7 accessibility readiness", () => {
  test(
    p7AccessibilityScenarioTitles[0],
    async ({ browser, page, sessionTracker }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YP7"),
        "FE-A11Y-P7 conflict accessibility",
      );
      const remote = await createIncidentMemberUser(page, incidentId, {
        display_name: "Accessible Analyst",
        email: uniqueEmail("fe-a11y-p7-remote"),
        initial_password: "FeA11yP7RemotePass!",
        role: "editor",
      });
      const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("fe-a11y-p7-row"),
        "timeline.activity_synopsis_text": "FE-A11Y-P7 conflict base",
      });
      const recordId = requireRecordId(row);
      const patchController = await installPatchController(page);
      const primarySocket = installIncidentSocketMonitor(page, incidentId);

      let remotePage: Page | null = null;
      try {
        await page.goto(`/?incident_id=${incidentId}`);
        await expect(
          page.getByTestId(workbookShellReadyTestId()),
        ).toBeVisible();
        await primarySocket.waitForAcceptedSocket();

        const remoteSession = await openIncidentAsTrackedUserReady(
          browser,
          sessionTracker,
          {
            createdBy: "FE-A11Y-P7-01",
            email: remote.email,
            incidentId,
            password: remote.initial_password,
            purpose: "FE-A11Y-P7 remote presence analyst",
            readyRecordId: recordId,
            userId: remote.user_id,
          },
        );
        remotePage = remoteSession.page;
        await focusRemoteTimelineCellAndWaitForPresence({
          actorText: "AA",
          fieldKey: "timeline.activity_synopsis_text",
          primaryPage: page,
          recordId,
          remotePage,
          socketMonitor: primarySocket,
        });

        await driveRealTimelineSummaryConflict({
          baseRowVersion: 1,
          localValue: "FE-A11Y-P7 local draft",
          page,
          patchController,
          recordId,
          remotePatchPage: remotePage,
          remoteValue: "FE-A11Y-P7 saved value",
          txnPrefix: "fea11yp7-conflict",
        });
        const resolver = page.getByTestId("conflict-resolver");
        await expect(resolver).toHaveAttribute(
          "aria-label",
          "Same-field conflict resolver",
        );
        const summary = page.getByTestId("conflict-resolver-summary");
        await expect(summary).toBeFocused();
        await expect(page.getByTestId("conflict-field-key")).toHaveValue(
          "timeline.activity_synopsis_text",
        );
        await expect(page.getByTestId("conflict-server-value")).toHaveValue(
          "FE-A11Y-P7 saved value",
        );
        await expect(page.getByTestId("conflict-local-value")).toHaveValue(
          "FE-A11Y-P7 local draft",
        );
        await expect(page.getByRole("button", { name: "Close" })).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Keep saved value" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use my unsaved value" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use merged value" }),
        ).toBeVisible();

        await summary.press("Enter");
        await expect(resolver).toBeVisible();
        await expect(page.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        const retainedConflictEditor = page.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId,
            surface: "grid",
          }),
        );
        await expect(
          semanticGridCell(retainedConflictEditor).getByRole("img", {
            name: "Conflict on Activity Synopsis",
          }),
        ).toBeVisible();
        await expectAllInteractiveControlsNamed(page);
        await expectNoFocusTrap(page);
        await expectAndRecordContrast(page, [
          saveStateTestId(),
          rowPresenceMarkerTestId(recordId),
          cellPresenceMarkerTestId(recordId, "timeline.activity_synopsis_text"),
          "conflict-close",
          "conflict-keep-saved",
          "conflict-use-unsaved",
          "conflict-use-merged",
        ]);

        await page.keyboard.press("Escape");
        await expect(resolver).toHaveCount(0);
        await expect(page.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        await expect(
          semanticGridCell(retainedConflictEditor).getByRole("img", {
            name: "Conflict on Activity Synopsis",
          }),
        ).toBeVisible();
      } finally {
        await patchController.dispose();
        await remotePage?.context().close();
      }
    },
  );
});

test.describe("FE-P8 accessibility readiness", () => {
  test(p8AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP8"),
      "FE-A11Y-P8 query controls",
    );
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p8-reviewed"),
        "timeline.activity_synopsis_text": "FE-A11Y-P8 reviewed row",
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p8-rough"),
      "timeline.activity_synopsis_text": "FE-A11Y-P8 rough row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    const summarySortHeader = await mountedGridTarget(
      page,
      timelineViewSchemaId,
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
    );
    await expectVisibleFocus(summarySortHeader);
    await summarySortHeader.press("Enter");
    await expect(summarySortHeader).toContainText("Asc");

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

    await openFilterPopover(page, timelineViewSchemaId);
    const filterField = page.getByTestId(
      gridFilterFieldTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterField);
    await filterField.selectOption("timeline.capture_state");
    const filterValue = page.getByTestId(
      gridFilterValueTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterValue);
    try {
      await filterValue.selectOption("reviewed");
    } catch {
      await filterValue.fill("reviewed");
    }
    const filterApply = page.getByTestId(
      gridFilterApplyTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterApply);
    await filterApply.press("Enter");
    await assertActiveFilterChipVisible(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
    );
    const filterChip = page.getByTestId(
      gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
    );
    await expectVisibleFocus(filterChip);

    const groupingSelect = page.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(groupingSelect);
    await groupingSelect.selectOption("timeline.capture_state");
    const reviewedGroup = page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    );
    await expect(reviewedGroup).toBeVisible();
    await expectVisibleFocus(reviewedGroup);
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");

    const savedViewSelector = page.getByTestId(
      savedViewSelectorTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(savedViewSelector);
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "view_schema",
    );
    await openSavedViewActionMenu(page, timelineViewSchemaId);
    const savedViewNameInput = page.getByTestId(
      savedViewNameInputTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(savedViewNameInput);
    await savedViewNameInput.fill("FE-A11Y-P8 keyboard saved view");
    const createSavedViewButton = page.getByTestId(
      savedViewCreateButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(createSavedViewButton);
    await createSavedViewButton.press("Enter");
    const savedViewStatus = page.getByTestId(
      savedViewStatusTestId(timelineViewSchemaId),
    );
    await expect(savedViewStatus).toHaveAttribute("aria-live", "polite");
    await expect(savedViewStatus).toHaveText("Saved view created.");
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "saved_view",
    );

    await openSavedViewActionMenu(page, timelineViewSchemaId);
    const homeButton = page.getByTestId(
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(homeButton);
    await homeButton.press("Enter");
    await expect(savedViewStatus).toHaveText("Home view updated.");
    await openSavedViewActionMenu(page, timelineViewSchemaId);
    const defaultButton = page.getByTestId(
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(defaultButton);
    await defaultButton.press("Enter");
    await expect(savedViewStatus).toHaveText("Default view updated.");

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await openFilterPopover(page, timelineViewSchemaId);
    await openSavedViewActionMenu(page, timelineViewSchemaId);
    await expectAndRecordContrast(page, [
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
      gridFilterFieldTestId(timelineViewSchemaId),
      gridFilterValueTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
      gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
      savedViewSelectorTestId(timelineViewSchemaId),
      savedViewNameInputTestId(timelineViewSchemaId),
      savedViewCreateButtonTestId(timelineViewSchemaId),
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
      savedViewStatusTestId(timelineViewSchemaId),
    ]);
  });
});

test.describe("FE-P9 accessibility readiness", () => {
  test(p9ConfigAccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP902"),
      "FE-A11Y-P9-02 config-driven inspector",
    );
    const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p9-02-row"),
      "timeline.raw_activity_text": "FE-A11Y-P9-02 inspector details",
      "timeline.activity_synopsis_text": "FE-A11Y-P9-02 selected row",
    })) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

    const toggle = page.getByTestId(
      workbookInspectorToggleTestId(timelineViewSchemaId),
    );
    await toggle.focus();
    await expectVisibleFocus(toggle);
    await toggle.press("Enter");
    await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
      "no_row_selected",
    );
    await page.keyboard.press("Escape");
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await expectVisibleFocus(toggle);

    const summaryCell = await mountedGridCell(
      page,
      timelineViewSchemaId,
      row.record_id,
      "timeline.activity_synopsis_text",
    );
    await expectVisibleSemanticGridCellFocus(summaryCell);
    await openTimelineInspector(page, row.record_id);
    const detailsEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: row.record_id,
        surface: "inspector",
      }),
    );
    await expectVisibleFocus(detailsEditor);
    await expect(detailsEditor).toHaveValue("FE-A11Y-P9-02 inspector details");

    await expectVisibleSemanticGridCellFocus(summaryCell);
    await page.keyboard.press("Shift+F10");
    const openHistory = page.getByTestId(
      rowHistoryOpenButtonTestId(row.record_id),
    );
    await expectVisibleFocus(openHistory);
    await openHistory.press("Enter");
    await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
    const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
    await expectVisibleFocus(deleteButton);
    await deleteButton.press("Enter");
    const deletePanel = page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    );
    await expect(deletePanel).toHaveAttribute("role", "alertdialog");
    await expect(deletePanel).toHaveAttribute("aria-modal", "true");
    await expect(deletePanel).toContainText(row.record_id);
    const deleteConfirm = page.getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    );
    const deleteCancel = page.getByTestId(
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    );
    await expectVisibleFocus(deleteConfirm);
    await expectVisibleFocus(deleteCancel);
    await expectAndRecordContrast(page, [
      workbookInspectorToggleTestId(timelineViewSchemaId),
      rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: row.record_id,
        surface: "inspector",
      }),
      rowHistoryOpenButtonTestId(row.record_id),
      rowHistoryDeleteButtonTestId(),
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    ]);
    await deleteCancel.press("Enter");
    await expect(deletePanel).toHaveCount(0);
  });

  test(p9AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP9"),
      "FE-A11Y-P9 inspector actions",
    );
    const evidence = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p9-evidence"),
        "evidence.collector_party_text": "FE-A11Y-P9 collector",
        "evidence.title": "FE-A11Y-P9 evidence",
      },
    )) as ViewRow;
    const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
      [hostRefsFieldKey]: collectionActionsPayload(["FE-A11Y-P9 host"]),
      client_txn_id: uniqueTxn("fe-a11y-p9-row"),
      "timeline.raw_activity_text": "FE-A11Y-P9 inspector details",
      "timeline.activity_synopsis_text": "FE-A11Y-P9 selected row",
    })) as ViewRow;
    const linkedRow = (await patchRecord(page, row.record_id, {
      base_row_version: row.row_version,
      changes: [
        {
          action_payload: A11yAttachedEvidencePayload(evidence.record_id),
          field_key: "timeline.attached_evidence_ids",
        },
      ],
      client_txn_id: uniqueTxn("fe-a11y-p9-link"),
      view_schema_id: timelineViewSchemaId,
    })) as ViewRow;
    const hostItem = requireItemByRawText(
      collectionItems(linkedRow, hostRefsFieldKey),
      "FE-A11Y-P9 host",
    );
    const history = await fetchA11yRecordHistory(page, row.record_id);
    const rollbackItem = requireA11yHistoryEntryAction(history);
    const rollbackAnchor = A11yRollbackAnchor(rollbackItem, "history_entry");

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    const summaryCell = await mountedGridCell(
      page,
      timelineViewSchemaId,
      row.record_id,
      "timeline.activity_synopsis_text",
    );
    await expectVisibleSemanticGridCellFocus(summaryCell);
    await openTimelineInspector(page, row.record_id);

    for (const section of [
      "operational-text",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      await expect(
        page.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeVisible();
    }
    const detailsEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: row.record_id,
        surface: "inspector",
      }),
    );
    await expectVisibleFocus(detailsEditor);
    await expect(detailsEditor).toHaveValue("FE-A11Y-P9 inspector details");

    const relationshipChip = page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(hostItem.item_ref)));
    await expect(relationshipChip).toContainText("Unresolved");
    await expectVisibleFocus(relationshipChip);

    await expectVisibleSemanticGridCellFocus(summaryCell);
    await page.keyboard.press("Shift+F10");
    const openHistory = page.getByTestId(
      rowHistoryOpenButtonTestId(row.record_id),
    );
    await expectVisibleFocus(openHistory);
    await openHistory.press("Enter");
    await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
    const rollbackAction = page.getByTestId(
      A11yHistoryActionTestId(rollbackItem, "history_entry"),
    );
    await expectVisibleFocus(rollbackAction);
    await rollbackAction.press("Enter");

    const rollbackPreview = page.getByTestId(
      rowHistoryRollbackPreviewTestId(rollbackAnchor),
    );
    await expect(rollbackPreview).toHaveAttribute("role", "dialog");
    await expect(rollbackPreview).toHaveAttribute("aria-modal", "true");
    await expect(rollbackPreview).toContainText(/rollback/i);
    const rollbackCancel = page.getByTestId(
      rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
    );
    const rollbackConfirm = page.getByTestId(
      rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
    );
    await expectVisibleFocus(rollbackCancel);
    await expectVisibleFocus(rollbackConfirm);

    await page.route(`**/api/v1/records/${row.record_id}/rollback`, (route) =>
      route.fulfill({
        contentType: "application/json",
        status: 409,
        body: JSON.stringify({
          error: {
            code: "row_version_conflict",
            message: "Rollback target is stale for FE-A11Y-P9.",
            retryable: false,
          },
        }),
      }),
    );
    await rollbackConfirm.press("Enter");
    const historyMessage = page.getByTestId(rowHistoryMessageTestId());
    await expectAlertRole(historyMessage);
    await expect(historyMessage).toHaveAttribute("aria-live", "assertive");
    await expect(historyMessage).toContainText("row_version_conflict");
    await expectNoPrivateDiagnostics(historyMessage);

    const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
    await expectVisibleFocus(deleteButton);
    await deleteButton.press("Enter");
    const deletePanel = page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    );
    await expect(deletePanel).toHaveAttribute("role", "alertdialog");
    await expect(deletePanel).toHaveAttribute("aria-modal", "true");
    await expect(deletePanel).toContainText(row.record_id);
    const deleteConfirm = page.getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    );
    const deleteCancel = page.getByTestId(
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    );
    await expectVisibleFocus(deleteConfirm);
    await expectVisibleFocus(deleteCancel);
    await expectAndRecordContrast(page, [
      rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: row.record_id,
        surface: "inspector",
      }),
      relationshipChipTestId(String(hostItem.item_ref)),
      rowHistoryOpenButtonTestId(row.record_id),
      A11yHistoryActionTestId(rollbackItem, "history_entry"),
      rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
      rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
      rowHistoryMessageTestId(),
      rowHistoryDeleteButtonTestId(),
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    ]);
    await deleteCancel.press("Enter");
    await expect(deletePanel).toHaveCount(0);

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
  });
});

test.describe("FE-P10 accessibility readiness", () => {
  test(p10AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP10"),
      "FE-A11Y-P10 coordination accessibility",
    );
    const owner = await createIncidentMemberUser(page, incidentId, {
      display_name: "FE-P10 accessibility owner",
      email: uniqueEmail("fe-p10-a11y-owner"),
      initial_password: "Phase10A11y1!",
      role: "editor",
    });
    const party = (await createViewRow(page, incidentId, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-party"),
      "party.display_name": "FE-A11Y-P10 response party",
      "party.party_kind": "team",
    })) as ViewRow;
    const task = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-task"),
        "task.priority": "normal",
        "task.requester_party_id": party.record_id,
        "task.task_kind": "collection",
        "task.title": "FE-A11Y-P10 task alpha",
      },
    )) as ViewRow;
    const urgentTask = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-urgent-task"),
        "task.priority": "urgent",
        "task.task_kind": "follow_up",
        "task.title": "FE-A11Y-P10 task urgent",
      },
    )) as ViewRow;
    const clipboardRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-clipboard"),
        "timeline.activity_utc_text": "2026-06-12T10:00:00Z",
        "timeline.activity_synopsis_text": "FE-A11Y-P10 clipboard row",
      },
    )) as ViewRow;
    const decision = (await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-decision"),
        "decision.decision_type": "containment",
        "decision.rationale": "FE-A11Y-P10 coordination rationale",
        "decision.summary": "FE-A11Y-P10 decision summary",
      },
    )) as ViewRow;
    const comm = (await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-comm"),
      "comm_log.audience": "FE-A11Y-P10 responders",
      "comm_log.channel_or_meeting": "FE-A11Y-P10 bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.decision_ids": {
        actions: [
          { linked_record_id: decision.record_id, op: "add_record_ref" },
        ],
        kind: "collection_actions_v1",
      },
      "comm_log.summary": "FE-A11Y-P10 communications log",
    })) as ViewRow;
    const handoff = (await createViewRow(
      page,
      incidentId,
      handoffViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-handoff"),
        "handoff.current_state_summary": "FE-A11Y-P10 handoff state",
        "handoff.incoming_owner_user_id": owner.user_id,
      },
    )) as ViewRow;
    const status = (await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-status"),
        "status_review.current_state_summary":
          "FE-A11Y-P10 status review state",
      },
    )) as ViewRow;
    const lesson = (await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-lesson"),
      "lesson.summary": "FE-A11Y-P10 lesson summary",
    })) as ViewRow;

    const surfaces = [
      {
        expected: "FE-A11Y-P10 task alpha",
        fieldKey: "task.title",
        groupToken: "coordination",
        label: "Task Requests",
        row: task,
        viewSchemaId: taskRequestsViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 decision summary",
        fieldKey: "decision.summary",
        groupToken: "coordination",
        label: "Decisions",
        row: decision,
        viewSchemaId: decisionsViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 response party",
        fieldKey: "party.display_name",
        groupToken: "coordination",
        label: "Parties",
        row: party,
        viewSchemaId: partiesViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 communications log",
        fieldKey: "comm_log.summary",
        groupToken: "coordination",
        label: "Communications Log",
        row: comm,
        viewSchemaId: commLogViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 handoff state",
        fieldKey: "handoff.current_state_summary",
        groupToken: "coordination",
        label: "Handoff",
        row: handoff,
        viewSchemaId: handoffViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 status review state",
        fieldKey: "status_review.current_state_summary",
        groupToken: "review-learning",
        label: "Status Review",
        row: status,
        viewSchemaId: statusReviewViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 lesson summary",
        fieldKey: "lesson.summary",
        groupToken: "review-learning",
        label: "Lesson",
        row: lesson,
        viewSchemaId: lessonViewSchemaId,
      },
    ] as const;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        taskRequestsViewSchemaId,
      )}`,
    );
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expectCurrentIncidentRole(page, "admin");
    await expectStatusRole(page.getByTestId(saveStateTestId()));
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectTabTraversalAdvancesFrom(
      page,
      page.getByTestId(systemViewSwitcherTriggerTestId()),
    );

    for (const surface of surfaces) {
      await test.step(`${surface.label} accessibility surface`, async () => {
        await openA11ySystemSurface(page, surface);
        await expect(page).toHaveURL(
          new RegExp(
            `view_schema_id=${encodeURIComponent(surface.viewSchemaId)}`,
          ),
        );
        await expect(
          page.getByTestId(savedViewSelectorTestId(surface.viewSchemaId)),
        ).toHaveAttribute("data-selected-sheet-ref-kind", "view_schema");
        await openFilterPopover(page, surface.viewSchemaId);
        await expect(
          page.getByTestId(gridFilterFieldTestId(surface.viewSchemaId)),
        ).toBeVisible();
        await expect(
          page.getByTestId(gridFilterApplyTestId(surface.viewSchemaId)),
        ).toBeVisible();
        await expect(
          page.getByTestId(gridGroupingSelectTestId(surface.viewSchemaId)),
        ).toBeVisible();

        const sortHeader = await mountedGridTarget(
          page,
          surface.viewSchemaId,
          gridSortHeaderTestId(surface.viewSchemaId, surface.fieldKey),
        );
        await expectVisibleFocus(sortHeader);
        await expect(sortHeader).not.toHaveText("");

        const cell = await mountedGridCell(
          page,
          surface.viewSchemaId,
          surface.row.record_id,
          surface.fieldKey,
        );
        await expectVisibleSemanticGridCellFocus(cell);
        await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
          surface.viewSchemaId,
        );
        await expectCellTextOrValue(cell, surface.expected);
      });
    }

    await openA11ySystemSurface(page, {
      groupToken: "coordination",
      viewSchemaId: taskRequestsViewSchemaId,
    });

    const taskTitle = await mountedGridCell(
      page,
      taskRequestsViewSchemaId,
      task.record_id,
      "task.title",
    );
    const taskTitleCell = await expectVisibleSemanticGridCellFocus(taskTitle);
    const copiedTaskTitle = await taskTitleCell.evaluate((element) => {
      const data = new DataTransfer();
      element.dispatchEvent(
        new ClipboardEvent("copy", {
          bubbles: true,
          cancelable: true,
          clipboardData: data,
        }),
      );
      return data.getData("text/plain");
    });
    expect(copiedTaskTitle).toBe("FE-A11Y-P10 task alpha");

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    const clipboardSummary = await mountedGridCell(
      page,
      timelineViewSchemaId,
      clipboardRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expectVisibleSemanticGridCellFocus(clipboardSummary);
    await pasteGridMatrix({
      fieldKey: "timeline.activity_synopsis_text",
      matrix: [["FE-A11Y-P10 pasted timeline", "fe-a11y-p10-host"]],
      page,
      recordId: clipboardRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(clipboardSummary).toHaveText("FE-A11Y-P10 pasted timeline");

    await openA11ySystemSurface(page, {
      groupToken: "coordination",
      viewSchemaId: taskRequestsViewSchemaId,
    });

    await applyFilterChip(
      page,
      taskRequestsViewSchemaId,
      "task.priority",
      "urgent",
    );
    await assertActiveFilterChipVisible(
      page,
      taskRequestsViewSchemaId,
      "task.priority",
    );
    const priorityChip = page.getByTestId(
      gridFilterChipTestId(taskRequestsViewSchemaId, "task.priority"),
    );
    await expect(priorityChip).toContainText("urgent");
    await expectVisibleFocus(priorityChip);

    const urgentTitle = await mountedGridCell(
      page,
      taskRequestsViewSchemaId,
      urgentTask.record_id,
      "task.title",
    );
    await expectCellTextOrValue(urgentTitle, "FE-A11Y-P10 task urgent");

    await openSavedViewActionMenu(page, taskRequestsViewSchemaId);
    const savedViewName = page.getByTestId(
      savedViewNameInputTestId(taskRequestsViewSchemaId),
    );
    await expectVisibleFocus(savedViewName);
    await savedViewName.fill("FE-A11Y-P10 keyboard saved view");
    const createSavedView = page.getByTestId(
      savedViewCreateButtonTestId(taskRequestsViewSchemaId),
    );
    await expectVisibleFocus(createSavedView);
    await createSavedView.press("Enter");
    const savedStatus = page.getByTestId(
      savedViewStatusTestId(taskRequestsViewSchemaId),
    );
    await expect(savedStatus).toHaveAttribute("aria-live", "polite");
    await expect(savedStatus).toHaveText("Saved view created.");

    await test.info().attach("fe-a11y-p10-01-readiness-matrix.json", {
      contentType: "application/json",
      body: Buffer.from(
        `${JSON.stringify(
          {
            checks: [
              "keyboard reachability",
              "visible focus",
              "accessible names",
              "system-view menu ARIA state",
              "status and saved-view live regions",
              "clipboard copy and paste",
              "non-color-only filter chip state",
            ],
            scenario_title: test.info().title,
            stable_identity_scope: "view_schema_id + record_id + field_key",
            surfaces: surfaces.map((surface) => ({
              field_key: surface.fieldKey,
              record_id: surface.row.record_id,
              surface: surface.label,
              view_schema_id: surface.viewSchemaId,
            })),
            viewport: "1440x900",
            zoom: "100%",
          },
          null,
          2,
        )}\n`,
      ),
    });

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await openFilterPopover(page, taskRequestsViewSchemaId);
    await openSavedViewActionMenu(page, taskRequestsViewSchemaId);
    await expectAndRecordContrast(page, [
      systemViewSwitcherTriggerTestId(),
      savedViewSelectorTestId(taskRequestsViewSchemaId),
      savedViewNameInputTestId(taskRequestsViewSchemaId),
      savedViewCreateButtonTestId(taskRequestsViewSchemaId),
      savedViewStatusTestId(taskRequestsViewSchemaId),
      gridFilterFieldTestId(taskRequestsViewSchemaId),
      gridFilterApplyTestId(taskRequestsViewSchemaId),
      gridFilterChipTestId(taskRequestsViewSchemaId, "task.priority"),
      gridGroupingSelectTestId(taskRequestsViewSchemaId),
      gridSortHeaderTestId(taskRequestsViewSchemaId, "task.title"),
      rowCellTestId(urgentTask.record_id, "task.title"),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P11 accessibility readiness", () => {
  test(p11AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP11"),
      "FE-A11Y-P11 global accessibility matrix",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p11-timeline"),
        "timeline.activity_utc_text": "2026-06-13T09:30:00Z",
        "timeline.activity_synopsis_text": "FE-A11Y-P11 timeline row",
        "timeline.raw_activity_text": "Global accessibility matrix details",
      },
    )) as ViewRow;
    const taskRow = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p11-task"),
        "task.priority": "normal",
        "task.task_kind": "collection",
        "task.title": "FE-A11Y-P11 task row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expectStatusRole(page.getByTestId(saveStateTestId()));
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
    await summaryCell.fill("FE-A11Y-P11 edited via keyboard");
    await summaryCell.press("Enter");
    await expect(
      page.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("FE-A11Y-P11 edited via keyboard");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    await openTimelineInspector(page, timelineRow.record_id);
    const inspectorSummaryCell = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await semanticGridCell(inspectorSummaryCell).focus();
    await expect(page.getByTestId("workbook-focus-anchor")).toHaveText(
      `${timelineViewSchemaId}:${timelineRow.record_id}:timeline.activity_synopsis_text`,
    );
    const semanticSummaryCell = semanticGridCell(inspectorSummaryCell);
    const inspectorDetails = page.getByTestId(
      rowInspectorFieldTestId(
        timelineRow.record_id,
        "timeline.raw_activity_text",
      ),
    );
    await expectVisibleFocus(inspectorDetails);
    await page.keyboard.press("Escape");
    await expect(semanticSummaryCell).toBeFocused();
    await semanticSummaryCell.press("Escape");
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
    await expectCellTextOrValue(taskTitle, "FE-A11Y-P11 task row");
    await expectVisibleSemanticGridCellFocus(taskTitle);
    await openFilterPopover(page, taskRequestsViewSchemaId);
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
      gridFilterFieldTestId(taskRequestsViewSchemaId),
      gridGroupingSelectTestId(taskRequestsViewSchemaId),
      rowCellTestId(taskRow.record_id, "task.title"),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P1 accessibility readiness", () => {
  test(p1AccessibilityScenarioTitles[0], async ({ page }) => {
    await clearBrowserSession(page);
    const heldSession = await holdSinglePublicAPIResponse(page, {
      method: "GET",
      path: "/api/v1/auth/session",
    });
    await new AuthGateway(page).goto();
    await heldSession.waitForHit;

    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "loading",
    );
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
    await expect(page.getByTestId(phase1AuthTestId("status"))).toContainText(
      "Checking current session",
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1AuthTestId("login-submit"),
      tabStops: [
        phase1AuthTestId("login-username"),
        phase1AuthTestId("login-password"),
        phase1AuthTestId("login-submit"),
      ],
    });

    try {
      heldSession.release();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "anonymous",
      );
    } finally {
      await heldSession.dispose();
    }
  });

  test(
    p1AccessibilityScenarioTitles[1],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-login");
      const password = "A11yP1LoginPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Login",
        initial_password: password,
        mfa_required: false,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "anonymous",
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-username"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(email, password);
      await expect(
        page.getByTestId(phase1LandingTestId("shell")),
      ).toBeVisible();
      await expectVisibleFocus(
        page.getByTestId(phase1LandingTestId("refresh")),
      );
      await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1LandingTestId("refresh"),
        tabStops: [
          phase1LandingTestId("search"),
          phase1LandingTestId("status-filter"),
          phase1LandingTestId("refresh"),
          phase1LandingTestId("create-open-button"),
        ],
      });
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 anonymous login",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[2],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-mfa");
      const password = "A11yP1MfaPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA",
        initial_password: password,
        mfa_required: true,
      });
      const secretBase32 = await enrollTotpViaBootstrap(email, password);

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_required",
      );
      await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
      await expect(page.getByTestId(phase1AuthTestId("status"))).toContainText(
        "Authenticator code",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId(phase1ErrorSummaryTestIds("auth").container),
      );
      expect(await hasSessionCookie(page)).toBeFalsy();
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-totp-code"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-totp-code"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(
        email,
        password,
        generateTotpCode(secretBase32),
      );
      await expect(
        page.getByTestId(phase1LandingTestId("current-user")),
      ).toContainText("A11Y P1 MFA");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 mfa_required retry",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[3],
    async ({ page, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-mfa-setup");
      const password = "A11yP1SetupPass!";
      await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA Setup",
        initial_password: password,
        mfa_required: true,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_setup_required",
      );
      await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
        "Authenticator setup is required before sign-in.",
      );
      await expect(
        page.getByTestId(phase1AuthTestId("bootstrap-token")),
      ).toHaveText("Stored for TOTP setup requests.");
      await expectNoPrivateDiagnostics(
        page.getByTestId(phase1ErrorSummaryTestIds("auth").container),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("bootstrap-begin"),
        tabStops: [
          phase1AuthTestId("bootstrap-begin"),
          phase1AuthTestId("bootstrap-complete-code"),
        ],
      });

      await new AuthGateway(page).beginBootstrapEnrollment();
      await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
      const secretBase32 = await new AuthGateway(page).requireText(
        phase1AuthTestId("bootstrap-secret-base32"),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("bootstrap-complete-code"),
        tabStops: [
          phase1AuthTestId("bootstrap-complete-code"),
          phase1AuthTestId("bootstrap-complete"),
        ],
      });
      await new AuthGateway(page).completeBootstrapEnrollment(
        generateTotpCode(secretBase32),
      );
      await expect(
        page
          .getByText("Authenticator setup is complete. Sign in again.")
          .first(),
      ).toBeVisible();
      expect(await hasSessionCookie(page)).toBeFalsy();
    },
  );

  test(p1AccessibilityScenarioTitles[4], async ({ page }) => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YLAND"),
      "A11Y authenticated landing",
    );

    await new IncidentDirectory(page).goto();
    await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
    await expect(
      page.getByTestId(landingIncidentCardTestId(incidentId)),
    ).toBeVisible();
    await expectVisibleFocus(
      page.getByTestId(phase1LandingTestId("create-open-button")),
    );
    await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1LandingTestId("create-open-button"),
      tabStops: [
        phase1LandingTestId("search"),
        phase1LandingTestId("status-filter"),
        phase1LandingTestId("refresh"),
        phase1LandingTestId("create-open-button"),
      ],
    });
  });

  test(
    p1AccessibilityScenarioTitles[5],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-incident");
      const password = "A11yP1IncidentPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Incident",
        initial_password: password,
        mfa_required: false,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(
        page.getByTestId(phase1LandingTestId("empty-state")),
      ).toContainText("No incidents are visible");
      await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));

      const selectedIncidentId = await createIncidentWithRequest(
        workerAdminRequest,
        uniqueIncidentKey("A11YEMPTYA"),
        "A11Y selected incident",
      );
      await createIncidentMembershipWithRequest(
        workerAdminRequest,
        selectedIncidentId,
        email,
        "admin",
      );
      const alternateIncidentId = await createIncidentWithRequest(
        workerAdminRequest,
        uniqueIncidentKey("A11YEMPTYB"),
        "A11Y alternate incident",
      );
      await createIncidentMembershipWithRequest(
        workerAdminRequest,
        alternateIncidentId,
        email,
        "admin",
      );

      await new IncidentDirectory(page).refresh();
      await expect(
        page.getByTestId(landingIncidentCardTestId(selectedIncidentId)),
      ).toBeVisible();
      await expect(
        page.getByTestId(landingIncidentCardTestId(alternateIncidentId)),
      ).toBeVisible();

      await new IncidentDirectory(page).openIncident(selectedIncidentId);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expectCurrentIncidentRole(page, "admin");
      await openIncidentControls(page, "incident-fields");
      await expectStatusRole(page.getByTestId(incidentControlsStatusTestId()));
      await expectVisibleFocus(
        page.getByTestId(authA11yAppLocalTestId("incidentPatchButton")),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1RouteTestId("workbook-current-user"),
        tabStops: [phase1RouteTestId("workbook-current-user")],
      });

      await new IncidentDirectory(page).open();
      await new IncidentDirectory(page).openIncident(selectedIncidentId);

      const selectedMembership = await loadIncidentMembership(
        workerAdminRequest,
        selectedIncidentId,
        user.user_id,
      );
      await deleteIncidentMembership(
        workerAdminRequest,
        selectedIncidentId,
        user.user_id,
        selectedMembership.membership_version,
      );
      await page.reload();
      await expect(page).not.toHaveURL(
        new RegExp(`incident_id=${selectedIncidentId}`),
      );
      await expect(page).toHaveURL(
        new RegExp(`incident_id=${alternateIncidentId}`),
      );
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expectCurrentIncidentRole(page, "admin");
      await openIncidentControls(page, "incident-fields");
      await expect(page.getByTestId("incident-patch-tlp")).toBeVisible();
      const alternateMembership = await loadIncidentMembership(
        workerAdminRequest,
        alternateIncidentId,
        user.user_id,
      );
      await patchIncidentMembership(workerAdminRequest, alternateIncidentId, {
        baseMembershipVersion: alternateMembership.membership_version,
        role: "editor",
        userId: user.user_id,
      });
      await new IncidentDirectory(page).patchIncidentFields({
        currentPhase: "containment",
        externalCase: "CASE-A11Y",
        tlp: "TLP:AMBER",
      });
      await expectAlertRole(page.getByTestId("incident-admin-error-code"));
      await expect(page.getByTestId("incident-admin-error-code")).toHaveText(
        "authorization_denied",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId("incident-admin-error-code"),
      );
      await expectVisibleFocus(
        page.getByTestId(phase1RouteTestId("workbook-current-user")),
      );
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 incident states",
        userId: user.user_id,
      });
    },
  );

  test(p1AccessibilityScenarioTitles[6], async ({ page }) => {
    const routePattern = "**/api/v1/incidents**";
    const routeHandler = async (route: Route) => {
      if (route.request().method().toUpperCase() !== "GET") {
        await route.fallback();
        return;
      }
      await fulfillPublicError(route, {
        code: "authorization_denied",
        details: {
          required_role: "incident_admin",
        },
        message: "Access denied.",
        status: 403,
      });
    };

    await page.route(routePattern, routeHandler);
    await page.goto("/");
    await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
    await expect(
      page.getByTestId(phase1LandingTestId("shell")),
    ).toHaveAttribute("data-bootstrap-state", "forbidden");
    await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
    await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("landing")));
    await expect(page.getByTestId(phase1ErrorCodeTestId("landing"))).toHaveText(
      "authorization_denied",
    );
    await expectNoPrivateDiagnostics(
      page.getByTestId(phase1ErrorSummaryTestIds("landing").container),
    );
    await expectVisibleFocus(page.getByTestId(phase1LandingTestId("refresh")));
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1LandingTestId("refresh"),
      tabStops: [phase1LandingTestId("refresh")],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(phase1LandingTestId("refresh")).click();
    await expect(page.getByTestId(phase1ErrorCodeTestId("landing"))).toHaveText(
      "",
    );
  });

  test(
    p1AccessibilityScenarioTitles[7],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YREVOKE"),
        "A11Y revoked incident",
      );
      const email = uniqueEmail("a11y-p1-revoked");
      const password = "A11yP1RevokedPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Revoked",
        initial_password: password,
        mfa_required: false,
      });
      await createIncidentMembership(page, incidentId, email, "viewer");

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(
        page.getByTestId(phase1RouteTestId("workbook-current-user")),
      ).toContainText("A11Y P1 Revoked");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 revoked session before revoke-all",
        userId: user.user_id,
      });

      await revokeAllSessions(
        workerAdminRequest,
        user.user_id,
        "FE-A11Y-P1-01 revoked-session",
      );
      await page.getByLabel("Account and application navigation").click();
      await page.getByRole("menuitem", { name: "Incidents" }).click();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "revoked",
      );
      await expect(
        page.getByTestId(phase1AuthTestId("shell-message")),
      ).toContainText("Sign in again");
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-submit"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(
        page.getByTestId(phase1RouteTestId("workbook-current-user")),
      ).toContainText("A11Y P1 Revoked");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 revoked session re-auth",
        userId: user.user_id,
      });
    },
  );

  test(p1AccessibilityScenarioTitles[8], async ({ page }) => {
    const routePattern = "**/api/v1/auth/credential-state";
    const routeHandler = async (route: Route) => {
      await fulfillPublicError(route, {
        code: "credential_state_unavailable",
        details: {
          field: "credential_state",
          reason_code: "temporary_failure",
        },
        message: "select private_column from local_credentials",
        status: 500,
      });
    };

    await page.route(routePattern, routeHandler);
    await page.goto("/");
    await new AccountSettings(page).open("account-security");
    await page.getByTestId(phase1AccountTestId("refresh-state")).focus();
    await expectVisibleFocus(
      page.getByTestId(phase1AccountTestId("refresh-state")),
    );
    await expect(page.getByTestId(phase1ErrorCodeTestId("account"))).toHaveText(
      "credential_state_unavailable",
    );
    await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("account")));
    await expectAlertRole(
      page.getByTestId(phase1ErrorSummaryTestIds("account").container),
    );
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").message),
    ).toHaveText("Request failed.");
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").details),
    ).toContainText("Reason: temporary_failure");
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").details),
    ).toContainText("Field: credential_state");
    await expectNoPrivateDiagnostics(
      page.getByTestId(phase1ErrorSummaryTestIds("account").container),
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1AccountTestId("refresh-state"),
      tabStops: [
        phase1AccountTestId("refresh-state"),
        phase1AccountTestId("logout"),
      ],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(phase1AccountTestId("refresh-state")).click();
    await expect(page.getByTestId(phase1AccountTestId("status"))).toHaveText(
      "Refreshed account security.",
    );
  });
});
