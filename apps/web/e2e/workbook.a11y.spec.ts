import { Buffer } from "node:buffer";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import type {
  CollectionActionsV1,
  EvidenceCreateRequest,
  RecordHistoryData,
  RecordHistoryItem,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  pasteGridMatrix,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  accountTestId,
  appRouteTestId,
  authTestId,
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentAdministrationTestId,
  incidentControlsStatusTestId,
  incidentLandingTestId,
  landingIncidentCardTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  networkAnalysisTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
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
  savedViewModifiedTestId,
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
  workbookActiveSurfaceFocusTargetTestId,
  workbookAddRowButtonTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
  workbookQueryEntryTestId,
  workbookQueryOverflowEntryTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
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
} from "@cartulary/view-contracts";
import type { APIRequestContext, Locator, Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { AuthGateway } from "./pages/authGateway";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { revokeAllSessions } from "./support/auth/sessions";
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
  networkFlowMinimalCSV,
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
import { fetchRecordHistory } from "./support/workbook/history";
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

const p1AccessibilityScenarioTitles = [
  "a11y.incident-selection.row-01 deferred session loading exposes progress and keeps recovery controls keyboard reachable",
  "a11y.incident-selection.row-01 anonymous login after initial session_required reaches login controls and authenticated landing",
  "a11y.incident-selection.row-01 mfa_required challenge is keyboard reachable, visibly focused, named, and safely announced",
  "a11y.incident-selection.row-01 mfa_setup_required enrollment is keyboard reachable and public errors hide private setup diagnostics",
  "a11y.incident-selection.row-01 authenticated landing exposes account, admin, incident, retry, and visible incident controls",
  "a11y.incident-selection.row-01 incident empty, list, selected, stale-selection, and incident-error states expose keyboard recovery",
  "a11y.incident-selection.row-01 forbidden access-denied public envelope is announced and exposes recovery without private diagnostics",
  "a11y.incident-selection.row-01 revoked session after prior authentication announces session end and supports re-authentication",
  "a11y.incident-selection.row-01 generic public error envelope renders safe diagnostics and keyboard error recovery",
] as const;
const p2AccessibilityScenarioTitles = [
  "a11y.workbook-shell.row-01 Verify shell regions, tabs, switchers, menus, inspector controls, and status strip are keyboard reachable, visibly focused, and named.",
] as const;
const p3AccessibilityScenarioTitles = [
  "a11y.grid-interaction.row-01 Verify grid cells, editors, group rows, active cell, edit mode, disabled/read-only state, and blocked actions are keyboard accessible and announced without color-only signals.",
] as const;
const p4AccessibilityScenarioTitles = [
  "a11y.mutation-lifecycle.row-01 Verify grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and Esc priority are keyboard and screen-reader safe.",
] as const;
const p5AccessibilityScenarioTitles = [
  "a11y.entity-linking.row-01 Verify mention chip states and manual-resolution controls have accessible names, visible focus, and non-color-only distinction.",
] as const;
const p6AccessibilityScenarioTitles = [
  "a11y.evidence-workflow.row-01 Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have names, focus, contrast, and non-color-only distinctions.",
] as const;
const p7AccessibilityScenarioTitles = [
  "a11y.collaboration.row-01 Verify conflict state, resolver controls, presence hint, stale-row notice, and save-state conflict communicate state by accessible name/state, not color alone.",
] as const;
const p8AccessibilityScenarioTitles = [
  "a11y.saved-view-query.row-01 Verify sort, filter, group, saved-view menu, active chips, group expand-collapse, and default/startup controls are keyboard reachable and announced.",
] as const;
const p9AccessibilityScenarioTitles = [
  "a11y.inspector-history.row-01 Verify inspector tabs, relationship links, evidence controls, history controls, rollback, destructive actions, and errors are keyboard reachable and announced.",
] as const;
const p9ConfigAccessibilityScenarioTitles = [
  "a11y.inspector-history.row-02 Verify keyboard open/close, panel navigation, Esc, focus restoration, disabled/blocked states, no-row empty state, and destructive confirmation focus for config-driven inspector behavior.",
] as const;
const p10AccessibilityScenarioTitles = [
  "a11y.coordination-review.row-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.",
] as const;
const p11AccessibilityScenarioTitles = [
  "a11y.design-readiness.row-01 Verify global accessibility matrix for keyboard access, visible focus, System views, grid navigation/edit entry/exit, Esc, ARIA states, icon-only labels, contrast, and non-color-only empty/loading/error/blocked states.",
] as const;

if (p2AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.workbook-shell.row-01 must declare exactly 1 scenario; found ${p2AccessibilityScenarioTitles.length}`,
  );
}
if (p3AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.grid-interaction.row-01 must declare exactly 1 scenario; found ${p3AccessibilityScenarioTitles.length}`,
  );
}
if (p4AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.mutation-lifecycle.row-01 must declare exactly 1 scenario; found ${p4AccessibilityScenarioTitles.length}`,
  );
}
if (p5AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.entity-linking.row-01 must declare exactly 1 scenario; found ${p5AccessibilityScenarioTitles.length}`,
  );
}
if (p6AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.evidence-workflow.row-01 must declare exactly 1 scenario; found ${p6AccessibilityScenarioTitles.length}`,
  );
}
if (p7AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.collaboration.row-01 must declare exactly 1 scenario; found ${p7AccessibilityScenarioTitles.length}`,
  );
}
if (p8AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.saved-view-query.row-01 must declare exactly 1 scenario; found ${p8AccessibilityScenarioTitles.length}`,
  );
}
if (p9AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.inspector-history.row-01 must declare exactly 1 scenario; found ${p9AccessibilityScenarioTitles.length}`,
  );
}
if (p9ConfigAccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.inspector-history.row-02 must declare exactly 1 scenario; found ${p9ConfigAccessibilityScenarioTitles.length}`,
  );
}
if (p10AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.coordination-review.row-01 must declare exactly 1 scenario; found ${p10AccessibilityScenarioTitles.length}`,
  );
}
if (p11AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `a11y.design-readiness.row-01 must declare exactly 1 scenario; found ${p11AccessibilityScenarioTitles.length}`,
  );
}

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

async function expectTextSpacingViewBarResilience(
  page: Page,
  options: {
    readonly capacity: number;
    readonly height: number;
    readonly hiddenCount: number;
    readonly width: number;
  },
) {
  await page.setViewportSize({ height: options.height, width: options.width });
  const queryControls = page.getByTestId(
    workbookViewBarQueryControlsTestId(timelineViewSchemaId),
  );
  await expect(queryControls).toHaveAttribute(
    "data-query-chip-capacity",
    String(options.capacity),
  );
  await expect(queryControls).toHaveAttribute(
    "data-hidden-query-chip-count",
    String(options.hiddenCount),
  );
  const geometry = await page.evaluate(
    ({
      actionMenuSelector,
      addRowSelector,
      filterSelector,
      groupingSelector,
      inspectorSelector,
      queryControlsSelector,
      savedViewSelector,
      sortSelector,
    }) => {
      const select = (selector: string) =>
        document.querySelector<HTMLElement>(selector);
      const requireElement = (element: HTMLElement | null, label: string) => {
        if (element === null) throw new Error(`Expected ${label} to exist`);
        return element;
      };
      const controls = requireElement(
        select(queryControlsSelector),
        "query controls",
      );
      const viewBar = requireElement(
        controls.closest<HTMLElement>(
          'section[aria-label="Workbook query and action controls"]',
        ),
        "workbook view bar",
      );
      const columns = requireElement(
        Array.from(controls.querySelectorAll<HTMLButtonElement>("button")).find(
          (button) => button.textContent?.trim() === "Columns",
        ) ?? null,
        "Columns button",
      );
      const chipButtons = Array.from(
        controls.querySelectorAll<HTMLButtonElement>(
          '[role="toolbar"][aria-label="Active query chips"] button[data-query-entry-key]',
        ),
      ).filter((button) => button.getBoundingClientRect().width > 0);
      const orderedControls: ReadonlyArray<readonly [string, HTMLElement]> = [
        ["saved-view", requireElement(select(savedViewSelector), "saved view")],
        [
          "saved-view-actions",
          requireElement(select(actionMenuSelector), "saved view actions"),
        ],
        ["sort", requireElement(select(sortSelector), "Sort button")],
        ["group", requireElement(select(groupingSelector), "Group select")],
        ["filters", requireElement(select(filterSelector), "Filters button")],
        ["columns", columns],
        ...chipButtons.map(
          (button, index) => [`chip-${index}`, button] as const,
        ),
        [
          "inspector",
          requireElement(select(inspectorSelector), "Inspector button"),
        ],
        ["add-row", requireElement(select(addRowSelector), "Add row button")],
      ];
      const viewBarRect = viewBar.getBoundingClientRect();
      const filter = requireElement(select(filterSelector), "Filters button");
      return {
        controls: orderedControls.map(([name, element]) => {
          const rect = element.getBoundingClientRect();
          return {
            clientHeight: element.clientHeight,
            left: rect.left,
            name,
            right: rect.right,
            scrollHeight: element.scrollHeight,
          };
        }),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        filter: {
          clientWidth: filter.clientWidth,
          scrollWidth: filter.scrollWidth,
        },
        viewBar: { left: viewBarRect.left, right: viewBarRect.right },
      };
    },
    {
      actionMenuSelector: dataTestIdSelector(
        savedViewActionMenuTriggerTestId(timelineViewSchemaId),
      ),
      addRowSelector: dataTestIdSelector(
        workbookAddRowButtonTestId(timelineViewSchemaId),
      ),
      filterSelector: dataTestIdSelector(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
      groupingSelector: dataTestIdSelector(
        gridGroupingSelectTestId(timelineViewSchemaId),
      ),
      inspectorSelector: dataTestIdSelector(
        workbookInspectorToggleTestId(timelineViewSchemaId),
      ),
      queryControlsSelector: dataTestIdSelector(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
      savedViewSelector: dataTestIdSelector(
        savedViewSelectorTestId(timelineViewSchemaId),
      ),
      sortSelector: dataTestIdSelector(
        workbookSortMenuTriggerTestId(timelineViewSchemaId),
      ),
    },
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
    geometry.document.clientHeight + 1,
  );
  expect(geometry.filter.scrollWidth).toBeLessThanOrEqual(
    geometry.filter.clientWidth + 1,
  );
  for (const control of geometry.controls) {
    expect(
      control.left,
      `${control.name} left containment`,
    ).toBeGreaterThanOrEqual(geometry.viewBar.left - 1);
    expect(
      control.right,
      `${control.name} right containment`,
    ).toBeLessThanOrEqual(geometry.viewBar.right + 1);
    expect(
      control.scrollHeight,
      `${control.name} block clipping`,
    ).toBeLessThanOrEqual(control.clientHeight + 1);
  }
  geometry.controls.slice(1).forEach((control, index) => {
    const previous = geometry.controls[index];
    expect(
      control.left,
      `${previous?.name} before ${control.name}`,
    ).toBeGreaterThanOrEqual((previous?.right ?? 0) - 1);
  });
}

async function expectRecoverySurfaceGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  const geometry = await page.evaluate(
    ({ activeSurfaceSelector, panelSelector }) => {
      const panel = document.querySelector<HTMLElement>(panelSelector);
      const activeSurface = document.querySelector<HTMLElement>(
        activeSurfaceSelector,
      );
      if (panel === null || activeSurface === null) {
        throw new Error("Expected recovery panel and active surface");
      }
      const heading = panel.querySelector<HTMLElement>("h2");
      const message = panel.querySelector<HTMLElement>('[role="status"]');
      const buttons = Array.from(
        panel.querySelectorAll<HTMLButtonElement>("button"),
      );
      const actionRow = buttons[0]?.parentElement;
      if (heading === null || message === null || actionRow == null) {
        throw new Error("Expected recovery heading, message, and actions");
      }
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          left: value.left,
          right: value.right,
          top: value.top,
        };
      };
      const style = getComputedStyle(panel);
      return {
        activeSurface: rect(activeSurface),
        actions: rect(actionRow),
        backgroundColor: style.backgroundColor,
        buttons: buttons.map(rect),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        heading: rect(heading),
        message: rect(message),
        overflowX: style.overflowX,
        overflowY: style.overflowY,
        panel: rect(panel),
        panelClientWidth: panel.clientWidth,
        panelScrollWidth: panel.scrollWidth,
      };
    },
    {
      activeSurfaceSelector: dataTestIdSelector(
        workbookActiveSurfaceFocusTargetTestId(),
      ),
      panelSelector: dataTestIdSelector(workbookEditRecoveryTestId()),
    },
  );

  expect(geometry.backgroundColor).not.toBe("rgba(0, 0, 0, 0)");
  expect(geometry.backgroundColor).not.toBe("transparent");
  expect(geometry.overflowX).toBe("hidden");
  expect(["auto", "scroll"]).toContain(geometry.overflowY);
  expect(geometry.panel.left).toBeGreaterThanOrEqual(
    geometry.activeSurface.left - 1,
  );
  expect(geometry.panel.right).toBeLessThanOrEqual(
    geometry.activeSurface.right + 1,
  );
  expect(geometry.panel.top).toBeGreaterThanOrEqual(
    geometry.activeSurface.top - 1,
  );
  expect(geometry.panel.bottom).toBeLessThanOrEqual(
    geometry.activeSurface.bottom + 1,
  );
  expect(geometry.heading.bottom).toBeLessThanOrEqual(geometry.message.top + 1);
  expect(geometry.message.bottom).toBeLessThanOrEqual(geometry.actions.top + 1);
  expect(geometry.panelScrollWidth).toBeLessThanOrEqual(
    geometry.panelClientWidth + 1,
  );
  for (const button of geometry.buttons) {
    expect(button.left).toBeGreaterThanOrEqual(geometry.panel.left - 1);
    expect(button.right).toBeLessThanOrEqual(geometry.panel.right + 1);
  }
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
}

async function expectNetworkAnalysisChromeGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  await page.evaluate(() => window.scrollTo(0, 0));
  const geometry = await page.evaluate(
    ({ advancedSelector, workspaceSelector }) => {
      const workspace = document.querySelector<HTMLElement>(workspaceSelector);
      if (workspace === null) {
        throw new Error("Expected Network Analysis workspace");
      }
      const controls = Array.from(
        workspace.querySelectorAll<HTMLElement>(
          "[data-network-flow-control]:not([hidden])",
        ),
      ).filter((control) => control.getClientRects().length > 0);
      const popover = workspace.querySelector<HTMLElement>(
        `${advancedSelector} .network-flow-popover`,
      );
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          height: value.height,
          left: value.left,
          right: value.right,
          top: value.top,
          width: value.width,
        };
      };
      return {
        controls: controls.map((control) => {
          const style = getComputedStyle(control);
          return {
            backgroundColor: style.backgroundColor,
            className: control.className,
            color: style.color,
            label:
              control.getAttribute("aria-label") ??
              control.textContent?.trim() ??
              control.tagName,
            ownedHorizontalOverflow: (() => {
              let ancestor = control.parentElement;
              while (ancestor !== null && ancestor !== workspace) {
                const ancestorStyle = getComputedStyle(ancestor);
                if (
                  (ancestorStyle.overflowX === "auto" ||
                    ancestorStyle.overflowX === "scroll") &&
                  ancestor.scrollWidth > ancestor.clientWidth
                ) {
                  return true;
                }
                ancestor = ancestor.parentElement;
              }
              return false;
            })(),
            rect: rect(control),
          };
        }),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        popover: popover === null ? null : rect(popover),
        workspace: rect(workspace),
        workspaceClientWidth: workspace.clientWidth,
        workspaceScrollWidth: workspace.scrollWidth,
      };
    },
    {
      advancedSelector: dataTestIdSelector(
        networkAnalysisTestId("advanced-filters"),
      ),
      workspaceSelector: dataTestIdSelector(networkAnalysisTestId("workspace")),
    },
  );

  expect(geometry.controls.length).toBeGreaterThan(10);
  expect(geometry.workspaceScrollWidth).toBeLessThanOrEqual(
    geometry.workspaceClientWidth + 1,
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
  for (const control of geometry.controls) {
    expect(control.backgroundColor).not.toBe("rgb(255, 255, 255)");
    expect(control.backgroundColor).not.toBe("rgba(255, 255, 255, 1)");
    expect(control.color).not.toBe("rgb(0, 0, 0)");
    expect(control.rect.width).toBeGreaterThan(0);
    expect(control.rect.height).toBeGreaterThan(0);
    if (!control.ownedHorizontalOverflow) {
      expect(control.rect.left, JSON.stringify(control)).toBeGreaterThanOrEqual(
        geometry.workspace.left - 1,
      );
      expect(control.rect.right, JSON.stringify(control)).toBeLessThanOrEqual(
        geometry.workspace.right + 1,
      );
    }
  }
  if (geometry.popover !== null) {
    expect(geometry.popover.left).toBeGreaterThanOrEqual(0);
    expect(geometry.popover.right).toBeLessThanOrEqual(
      geometry.document.clientWidth + 1,
    );
    expect(geometry.popover.top).toBeGreaterThanOrEqual(0);
    expect(geometry.popover.bottom).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
}

async function expectResolverSurfaceGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  const resolver = page.getByTestId(workbookConflictResolverTestId());
  await resolver.evaluate((element) => {
    element.scrollTop = 0;
  });
  const geometry = await page.evaluate(
    ({ activeSurfaceSelector, resolverSelector }) => {
      const panel = document.querySelector<HTMLElement>(resolverSelector);
      const activeSurface = document.querySelector<HTMLElement>(
        activeSurfaceSelector,
      );
      if (panel === null || activeSurface === null) {
        throw new Error("Expected resolver and active surface");
      }
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          left: value.left,
          right: value.right,
          top: value.top,
        };
      };
      const style = getComputedStyle(panel);
      return {
        activeSurface: rect(activeSurface),
        backgroundColor: style.backgroundColor,
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        overflowX: style.overflowX,
        overflowY: style.overflowY,
        panel: rect(panel),
        panelClientWidth: panel.clientWidth,
        panelScrollWidth: panel.scrollWidth,
      };
    },
    {
      activeSurfaceSelector: dataTestIdSelector(
        workbookActiveSurfaceFocusTargetTestId(),
      ),
      resolverSelector: dataTestIdSelector(workbookConflictResolverTestId()),
    },
  );
  expect(geometry.backgroundColor).not.toBe("rgba(0, 0, 0, 0)");
  expect(geometry.backgroundColor).not.toBe("transparent");
  expect(geometry.overflowX).toBe("hidden");
  expect(["auto", "scroll"]).toContain(geometry.overflowY);
  expect(geometry.panel.left).toBeGreaterThanOrEqual(
    geometry.activeSurface.left - 1,
  );
  expect(geometry.panel.right).toBeLessThanOrEqual(
    geometry.activeSurface.right + 1,
  );
  expect(geometry.panel.top).toBeGreaterThanOrEqual(
    geometry.activeSurface.top - 1,
  );
  expect(geometry.panel.bottom).toBeLessThanOrEqual(
    geometry.activeSurface.bottom + 1,
  );
  expect(geometry.panelScrollWidth).toBeLessThanOrEqual(
    geometry.panelClientWidth + 1,
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }

  const finalAction = page.getByTestId(
    workbookConflictControlTestId("use-merged"),
  );
  await finalAction.scrollIntoViewIfNeeded();
  await expect(finalAction).toBeVisible();
  const actionGeometry = await resolver.evaluate(
    (panel, actionSelector) => {
      const action = panel.querySelector<HTMLElement>(actionSelector);
      if (action === null) throw new Error("Expected resolver final action");
      const panelRect = panel.getBoundingClientRect();
      const actionRect = action.getBoundingClientRect();
      return {
        actionBottom: actionRect.bottom,
        actionLeft: actionRect.left,
        actionRight: actionRect.right,
        actionTop: actionRect.top,
        panelBottom: panelRect.bottom,
        panelLeft: panelRect.left,
        panelRight: panelRect.right,
        panelTop: panelRect.top,
      };
    },
    dataTestIdSelector(workbookConflictControlTestId("use-merged")),
  );
  expect(actionGeometry.actionLeft).toBeGreaterThanOrEqual(
    actionGeometry.panelLeft - 1,
  );
  expect(actionGeometry.actionRight).toBeLessThanOrEqual(
    actionGeometry.panelRight + 1,
  );
  expect(actionGeometry.actionTop).toBeGreaterThanOrEqual(
    actionGeometry.panelTop - 1,
  );
  expect(actionGeometry.actionBottom).toBeLessThanOrEqual(
    actionGeometry.panelBottom + 1,
  );
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
  surface: string,
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
  surface: string,
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

function A11yAttachedEvidencePayload(recordId: string): CollectionActionsV1 {
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
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function A11yRollbackAnchor(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireA11yHistoryEntryAction(history: RecordHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error(
      "missing a11y.inspector-history history_entry rollback item",
    );
  }
  return item;
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
  | "quarantined"
  | "pending_receipt"
  | "requested";

async function createA11yEvidenceRow(
  page: Page,
  incidentId: string,
  options: {
    lifecycleState: NonNullable<
      EvidenceCreateRequest["evidence.lifecycle_state"]
    >;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow accessibility fixture",
    ...options,
  });
}

async function createUploadedA11yEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow accessibility fixture",
    ...options,
  });
}

async function expectEvidenceAccessState(
  page: Page,
  recordId: string,
  stateKey: EvidenceA11yRowStateKey,
  options: {
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
  if (options.messageText !== undefined) {
    const message = page.getByTestId(evidenceAccessMessageTestId(recordId));
    await expect(message).toBeVisible();
    if (options.messageText !== undefined) {
      await expect(message).toContainText(options.messageText);
    }
  }
}

function evidenceAccessStateContainer(page: Page, recordId: string): Locator {
  return page.getByTestId(evidenceAccessStateTestId(recordId));
}

async function expectEvidenceControlsPainted(page: Page, recordId: string) {
  await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidenceAccessStateTestId(recordId),
  );
  const shell = page.getByTestId(gridShellTestId(evidenceViewSchemaId));
  const shellBox = await shell.boundingBox();
  const scrollportBox = await shell
    .locator(gridScrollportSelector())
    .boundingBox();
  expect(scrollportBox?.width).toBeLessThanOrEqual((shellBox?.width ?? 0) + 1);
  const state = page.getByTestId(evidenceAccessStateTestId(recordId));
  const result = await state.evaluate((element) =>
    Array.from(element.querySelectorAll("button")).map((button) => {
      const box = button.getBoundingClientRect();
      const hit = document.elementFromPoint(
        box.x + box.width / 2,
        box.y + box.height / 2,
      );
      let unclipped =
        box.width > 0 &&
        box.height > 0 &&
        (hit === button || button.contains(hit));
      for (
        let parent = button.parentElement;
        parent;
        parent = parent.parentElement
      ) {
        const style = getComputedStyle(parent);
        const clip = parent.getBoundingClientRect();
        if (
          ["hidden", "clip", "auto", "scroll"].includes(style.overflowY) &&
          (box.top < clip.top - 1 || box.bottom > clip.bottom + 1)
        )
          unclipped = false;
        if (
          ["hidden", "clip", "auto", "scroll"].includes(style.overflowX) &&
          (box.left < clip.left - 1 || box.right > clip.right + 1)
        )
          unclipped = false;
      }
      return { label: button.textContent, unclipped };
    }),
  );
  expect(result).toHaveLength(4);
  expect(
    result.filter((button) => !button.unclipped),
    "Every production Evidence control must be painted without clipping",
  ).toEqual([]);
}

async function armA11yPublicErrorFault(
  page: Page,
  options: {
    path: string;
    reasonCode: "blob_failed" | "evidence_inconsistent" | "unsupported_preview";
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
      message:
        "Evidence access failed for browser.evidence-workflow accessibility fixture.",
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

test.describe("browser.workbook-shell accessibility readiness", () => {
  test(p2AccessibilityScenarioTitles[0], async ({ page }) => {
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
  test("a11y.network-analysis.row-01 Verify claimed Network Analysis tabs, query controls, semantic grids, inspector, graph, contributor drawer, mapping modal, focus return, names, and ARIA evidence.", async ({
    page,
  }) => {
    await openClaimedNetworkAnalysis(page, "NETWORKFLOWA11Y");
    await page
      .getByTestId(networkAnalysisTestId("import-input"))
      .setInputFiles(networkFlowMinimalCSV);
    const mappingDialog = page.getByTestId(
      networkAnalysisTestId("mapping-dialog"),
    );
    await expect(mappingDialog).toBeVisible({ timeout: 30_000 });
    const mappingProfile = page.getByTestId(
      networkAnalysisTestId("mapping-profile"),
    );
    await expect(mappingProfile).toBeFocused();
    await expectAllInteractiveControlsNamed(page);
    await expectAndRecordContrast(page, [
      networkAnalysisTestId("mapping-dialog"),
      networkAnalysisTestId("mapping-profile"),
      networkAnalysisTestId("mapping-preview"),
    ]);
    await page
      .getByTestId(networkAnalysisTestId("mapping-display-name"))
      .fill("accessible-flow");
    await page.getByTestId(networkAnalysisTestId("mapping-preview")).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
    ).toBeVisible();
    await page.getByTestId(networkAnalysisTestId("mapping-apply")).click();
    await expect(mappingDialog).toHaveCount(0);

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

    await page.getByRole("button", { name: "Saved graphs" }).click();
    const savedPanel = page.getByTestId(networkAnalysisTestId("saved-graphs"));
    const saveTrigger = page.getByTestId(
      networkAnalysisTestId("saved-graph-create"),
    );
    await saveTrigger.focus();
    await expectVisibleFocus(saveTrigger);
    await saveTrigger.press("Enter");
    const savedName = page.getByTestId(
      networkAnalysisTestId("saved-graph-name"),
    );
    await expect(savedName).toBeFocused();
    await savedName.press("Shift+Tab");
    await expect(page.getByRole("button", { name: "Cancel" })).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(savedName).toBeFocused();
    await savedName.fill("Accessible saved graph");
    await page.getByRole("button", { name: "Save graph" }).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("saved-graph-dialog")),
    ).toHaveCount(0);
    await expect(saveTrigger).toBeFocused();
    await expect(
      savedPanel.getByText("Materialization succeeded.", { exact: true }),
    ).toBeVisible({ timeout: 15_000 });

    const savedVertex = page
      .getByTestId(/^network-flow-saved-graph-vertex-/u)
      .first()
      .getByRole("button");
    await savedVertex.click();
    const savedContributors = page.getByTestId(
      networkAnalysisTestId("saved-graph-contributors"),
    );
    await expect(savedContributors).toBeVisible();
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      networkAnalysisTestId("saved-graphs"),
      networkAnalysisTestId("saved-graph-result"),
      networkAnalysisTestId("saved-graph-contributors"),
    ]);
    await savedContributors.getByRole("button", { name: "Close" }).click();
    await expect(savedVertex).toBeFocused();

    await page
      .getByTestId(networkAnalysisTestId("graph-surface-explore"))
      .click();
    await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
    await page
      .getByLabel("Endpoint IP value")
      .fill(`2001:db8:${"longsegment".repeat(12)}`);
    await page
      .getByTestId(networkAnalysisTestId("advanced-filters"))
      .locator("summary")
      .click();
    for (const viewport of [
      { width: 1440, height: 900 },
      { width: 1024, height: 720 },
      { width: 768, height: 640 },
    ]) {
      await page.setViewportSize(viewport);
      await expectNetworkAnalysisChromeGeometry(page);
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    await expectNetworkAnalysisChromeGeometry(page, {
      enforceDocumentBlockExtent: false,
    });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "100%";
    });
    await page.setViewportSize({ width: 768, height: 640 });
    await page.evaluate(() => {
      const style = document.createElement("style");
      style.id = "network-flow-text-spacing";
      style.textContent = `
        .network-flow-chrome * {
          letter-spacing: 0.12em !important;
          line-height: 1.5 !important;
          word-spacing: 0.16em !important;
        }
      `;
      document.head.append(style);
    });
    await expectNetworkAnalysisChromeGeometry(page);
    await page.evaluate(() => {
      document.getElementById("network-flow-text-spacing")?.remove();
    });
    await page.setViewportSize({ width: 1440, height: 900 });
  });
}

test.describe("browser.grid-interaction accessibility readiness", () => {
  test(p3AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YGRIDINTERACTION"),
      "a11y.grid-interaction.row-01 grid adapter",
    );
    const alphaRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.grid-interaction-01-alpha"),
        "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
        "timeline.activity_synopsis_text": "Alpha accessibility row",
        "timeline.raw_activity_text": "Keyboard grid coverage",
      },
    );
    const betaRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.grid-interaction-01-beta"),
        "timeline.activity_utc_text": "2026-05-31T10:05:00Z",
        "timeline.activity_synopsis_text": "Beta accessibility row",
        "timeline.raw_activity_text": "Grouped grid coverage",
      },
    );

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

test.describe("browser.mutation-lifecycle accessibility readiness", () => {
  test(p4AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YMUTATIONLIFECYCLE"),
      "a11y.mutation-lifecycle.row-01 Timeline accessibility",
    );
    const editRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-edit"),
        "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
        "timeline.activity_synopsis_text":
          "browser.mutation-lifecycle edit accessibility row",
        "timeline.raw_activity_text": "Escape priority details",
      },
    );
    const pasteRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-paste"),
        "timeline.activity_utc_text": "2026-06-03T10:05:00Z",
        "timeline.activity_synopsis_text":
          "browser.mutation-lifecycle paste accessibility row",
      },
    );
    const pendingRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-pending"),
        "timeline.activity_utc_text": "2026-06-03T10:10:00Z",
        "timeline.activity_synopsis_text":
          "browser.mutation-lifecycle pending accessibility row",
      },
    );
    const validationRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-validation"),
        "timeline.activity_utc_text": "2026-06-03T10:15:00Z",
        "timeline.activity_synopsis_text":
          "browser.mutation-lifecycle validation accessibility row",
      },
    );

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
    await editSummary.fill(
      "browser.mutation-lifecycle accessibility committed edit",
    );
    await editSummary.press("Enter");
    await expect(
      page.getByTestId(
        rowCellTestId(editRow.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("browser.mutation-lifecycle accessibility committed edit");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    await pasteGridMatrix({
      fieldKey: "timeline.activity_synopsis_text",
      matrix: [
        [
          "browser.mutation-lifecycle accessibility pasted summary",
          "a11y-host.example",
        ],
      ],
      page,
      recordId: pasteRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        rowCellTestId(pasteRow.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("browser.mutation-lifecycle accessibility pasted summary");

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      const pendingSummary = await activateTimelineGridEditor(
        page,
        pendingRow.record_id,
        "timeline.activity_synopsis_text",
      );
      await pendingSummary.fill(
        "browser.mutation-lifecycle accessibility pending replay",
      );
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
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
      await blockedSummary.fill(
        "browser.mutation-lifecycle accessibility blocked edit",
      );
      await blockedSummary.press("Enter");
      await expect.poll(() => recoveryController.calls.length).toBe(1);

      const recoveryPanel = page.getByTestId(workbookEditRecoveryTestId());
      const retryButton = page.getByTestId(
        workbookEditRecoveryRetryButtonTestId(),
      );
      const discardButton = page.getByTestId(
        workbookEditRecoveryDiscardButtonTestId(),
      );
      await expect(recoveryPanel).toHaveRole("complementary");
      await expect(recoveryPanel).toHaveAccessibleName(
        "Workbook edit recovery",
      );
      await expect(recoveryPanel).not.toBeFocused();
      await expect(retryButton).toHaveAccessibleName(
        "Retry with a new request ID",
      );
      await expect(discardButton).toHaveAccessibleName("Discard blocked edit");
      await expect(retryButton).toBeEnabled();
      await expect(discardButton).toBeEnabled();
      await expect(
        recoveryPanel.getByRole("status", {
          name: "Queued edit recovery message",
        }),
      ).toHaveAttribute("aria-live", "polite");
      await expect(
        recoveryPanel.getByRole("status", {
          name: "Queued edit recovery message",
        }),
      ).toHaveAttribute("aria-atomic", "true");
      await expect(recoveryPanel.getByRole("status")).toHaveCount(1);
      await expect(recoveryPanel).toContainText("Queued edits");
      await expect(recoveryPanel).toContainText(
        "A queued edit could not be replayed safely. Retry it with a new request ID, or discard the blocked edit to continue.",
      );
      await expect(recoveryPanel).not.toContainText("client_txn_conflict");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);

      for (const viewport of [
        { width: 1440, height: 900 },
        { width: 1024, height: 720 },
        { width: 768, height: 640 },
      ]) {
        await page.setViewportSize(viewport);
        await expectRecoverySurfaceGeometry(page);
      }
      await page.setViewportSize({ width: 1440, height: 900 });
      await page.evaluate(() => {
        document.documentElement.style.zoom = "200%";
      });
      await expectRecoverySurfaceGeometry(page, {
        enforceDocumentBlockExtent: false,
      });
      await page.evaluate(() => {
        document.documentElement.style.zoom = "100%";
      });
      await page.setViewportSize({ width: 768, height: 640 });
      await page.evaluate(() => {
        const style = document.createElement("style");
        style.id = "workbook-recovery-text-spacing";
        style.textContent = `
          #root * {
            letter-spacing: 0.12em !important;
            line-height: 1.5 !important;
            word-spacing: 0.16em !important;
          }
          #root p {
            margin-bottom: 2em !important;
          }
        `;
        document.head.append(style);
      });
      await expectRecoverySurfaceGeometry(page);
      await page.evaluate(() => {
        document.getElementById("workbook-recovery-text-spacing")?.remove();
      });
      await page.setViewportSize({ width: 1440, height: 900 });

      await page.getByTestId(saveStateActionButtonTestId()).click();
      await expect(recoveryPanel).toBeFocused();
      await page.keyboard.press("Tab");
      await expect(retryButton).toBeFocused();
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
        workbookEditRecoveryTestId(),
        workbookEditRecoveryRetryButtonTestId(),
        workbookEditRecoveryDiscardButtonTestId(),
        saveStateTestId(),
      ]);

      await discardButton.press("Space");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(recoveryPanel).toHaveCount(0);
      await expect(
        page.getByTestId(workbookActiveSurfaceFocusTargetTestId()),
      ).toBeFocused();
      expect(recoveryController.calls).toHaveLength(1);
    } finally {
      await recoveryController.dispose();
    }
  });
});

test.describe("browser.entity-linking accessibility readiness", () => {
  test(p5AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YENTITYLINKING"),
      "a11y.entity-linking.row-01 mention states",
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
      displayPrefix: "a11y.entity-linking",
      hostnamePrefix: "a11y.entity-linking",
      occurredAt: {
        auto: "2026-06-06T16:15:00Z",
        dismissed: "2026-06-06T16:20:00Z",
        manual: "2026-06-06T16:10:00Z",
        resolved: "2026-06-06T16:05:00Z",
        unresolved: "2026-06-06T16:00:00Z",
      },
      rawTextPrefix: "A11YENTITYLINKING",
      summary: {
        auto: "a11y.entity-linking auto chip",
        dismissed: "a11y.entity-linking dismissed chip",
        manual: "a11y.entity-linking manual chip",
        resolved: "a11y.entity-linking resolved chip",
        unresolved: "a11y.entity-linking unresolved chip",
      },
      txnPrefix: "a11y.entity-linking",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

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
      /^Resolved a11y.entity-linking Resolved Target$/u,
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
    await expect(autoNotice).toContainText("a11y.entity-linking Auto Target");
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

test.describe("browser.evidence-workflow accessibility readiness", () => {
  test(p6AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.emulateMedia({ reducedMotion: "reduce" });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YEVIDENCEWORKFLOW"),
      "a11y.evidence-workflow.row-01 evidence access",
    );
    const requested = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "requested",
      requestedAt: "2026-06-07T10:00:00Z",
      storageRef: "case://a11y.evidence-workflow/requested",
      title: "01 requested evidence",
      txnPrefix: "a11y.evidence-workflow-requested",
    });
    const pending = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "pending_receipt",
      requestedAt: "2026-06-07T10:05:00Z",
      storageRef: "case://a11y.evidence-workflow/pending",
      title: "02 pending evidence",
      txnPrefix: "a11y.evidence-workflow-pending",
    });
    const blocked = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "quarantined",
      requestedAt: "2026-06-07T10:10:00Z",
      storageRef: "case://a11y.evidence-workflow/quarantined",
      title: "03 quarantined evidence",
      txnPrefix: "a11y.evidence-workflow-blocked",
    });
    const availablePreview = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("a11y.evidence-workflow preview evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "a11y.evidence-workflow-preview.txt",
        requestedAt: "2026-06-07T10:15:00Z",
        title: "04 available preview evidence",
        txnPrefix: "a11y.evidence-workflow-preview",
      },
    );
    const downloadHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from("a11y.evidence-workflow download evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "a11y.evidence-workflow-download.txt",
      requestedAt: "2026-06-07T10:20:00Z",
      title: "05 download handle evidence",
      txnPrefix: "a11y.evidence-workflow-download",
    });
    const previewBlocked = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from(
        "<!doctype html><title>a11y.evidence-workflow unsupported preview</title>",
        "utf8",
      ),
      contentType: "text/html",
      filename: "a11y.evidence-workflow-preview-blocked.html",
      requestedAt: "2026-06-07T10:25:00Z",
      title: "06 preview blocked evidence",
      txnPrefix: "a11y.evidence-workflow-preview-blocked",
    });
    const failedHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from(
        "a11y.evidence-workflow failed handle evidence\n",
        "utf8",
      ),
      contentType: "text/plain",
      filename: "a11y.evidence-workflow-failed.txt",
      requestedAt: "2026-06-07T10:30:00Z",
      title: "07 failed handle evidence",
      txnPrefix: "a11y.evidence-workflow-failed",
    });
    const inconsistentHandle = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "a11y.evidence-workflow inconsistent handle evidence\n",
          "utf8",
        ),
        contentType: "text/plain",
        filename: "a11y.evidence-workflow-inconsistent.txt",
        requestedAt: "2026-06-07T10:35:00Z",
        title: "08 inconsistent handle evidence",
        txnPrefix: "a11y.evidence-workflow-inconsistent",
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
      messageText: "Requested",
    });
    await expectEvidenceAccessState(
      page,
      pending.record_id,
      "pending_receipt",
      {
        messageText: "Pending receipt",
      },
    );
    await expectEvidenceAccessState(page, blocked.record_id, "quarantined", {
      messageText: "Quarantined",
    });

    await expectEvidenceAccessState(
      page,
      availablePreview.record_id,
      "available",
    );
    const paintedAction = await page
      .getByTestId(evidencePreviewButtonTestId(availablePreview.record_id))
      .evaluate((button) => {
        const box = button.getBoundingClientRect();
        let parent = button.parentElement;
        while (parent) {
          const style = getComputedStyle(parent);
          const clip = parent.getBoundingClientRect();
          if (
            ["hidden", "clip", "auto", "scroll"].includes(style.overflowY) &&
            (box.top < clip.top - 1 || box.bottom > clip.bottom + 1)
          )
            return false;
          parent = parent.parentElement;
        }
        return box.height > 0;
      });
    expect(
      paintedAction,
      "Evidence Preview must fit its painted row bounds",
    ).toBe(true);
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
    ).not.toHaveAttribute("aria-live");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveText("Preview open");
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
    expect(download.suggestedFilename()).toBe(
      "a11y.evidence-workflow-download.txt",
    );
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).not.toHaveAttribute("aria-live");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveText("Download requested");

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
    await expect(previewBlockedMessage).not.toHaveAttribute("aria-live");
    await expect(
      page.getByRole("status").filter({ hasText: /evidence:/u }),
    ).toHaveCount(1);
    await expect(previewBlockedMessage).toContainText("No preview");
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
    await expect(failedMessage).not.toHaveAttribute("aria-live");
    await expect(
      page.getByRole("status").filter({ hasText: /evidence:/u }),
    ).toHaveCount(1);
    await expect(failedMessage).toContainText("Upload failed");
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
    await expect(inconsistentMessage).not.toHaveAttribute("aria-live");
    await expect(
      page.getByRole("status").filter({ hasText: /evidence:/u }),
    ).toHaveCount(1);
    await expect(inconsistentMessage).toContainText("Inconsistent");
    await expectNoPrivateDiagnostics(inconsistentMessage);

    // Exercise the production row and inspector under the existing user profiles.
    const preferences = await page.request.get(
      `${apiBase}/api/v1/account/preferences`,
    );
    const originalPreferences = (await preferences.json()).data;
    const setDensity = async (density: string) => {
      const response = await page.request.get(
        `${apiBase}/api/v1/account/preferences`,
      );
      const current = (await response.json()).data;
      const updated = await page.request.put(
        `${apiBase}/api/v1/account/preferences`,
        {
          headers: await csrfHeaders(page),
          data: {
            base_preferences_version: current.preferences_version,
            client_txn_id: uniqueTxn("evidence-access-density"),
            density_mode: density,
          },
        },
      );
      expect(updated.ok()).toBeTruthy();
      await page.reload();
    };
    try {
      for (const density of ["compact", "default", "comfortable"]) {
        await setDensity(density);
        await expectEvidenceControlsPainted(page, availablePreview.record_id);
        const spacing = await page.addStyleTag({
          content: `* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-block-end: 2em !important; }`,
        });
        await expectEvidenceControlsPainted(page, availablePreview.record_id);
        await spacing.evaluate((element) =>
          element.parentNode?.removeChild(element),
        );
      }
      await setDensity("default");
      for (const width of [1440, 1024, 768]) {
        await page.setViewportSize({
          width,
          height: width >= 1280 ? 900 : width >= 1024 ? 720 : 640,
        });
        await expectEvidenceControlsPainted(page, availablePreview.record_id);
        const state = page.getByTestId(
          evidenceAccessMessageTestId(availablePreview.record_id),
        );
        await state.focus();
        await state.press("Enter");
        const inspectorPreview = page.getByTestId(
          evidencePreviewButtonTestId(availablePreview.record_id, "inspector"),
        );
        await expect(inspectorPreview).toBeVisible();
        await inspectorPreview.focus();
        await inspectorPreview.press("Enter");
        const panel = page.getByTestId(evidencePreviewPanelTestId());
        const close = panel.getByRole("button", { name: "Close" });
        await expect(close).toBeVisible();
        await close.focus();
        await expect(close).toBeFocused();
        await close.press("Escape");
        await expect(panel).toHaveCount(0);
        await expect(inspectorPreview).toBeFocused();
        if (width === 768) {
          await armA11yPublicErrorFault(page, {
            path: `/api/v1/evidence-records/${availablePreview.record_id}/preview-handle`,
            reasonCode: "unsupported_preview",
          });
          await inspectorPreview.press("Enter");
          await expect(
            page.getByTestId(
              evidenceAccessMessageTestId(
                availablePreview.record_id,
                "inspector",
              ),
            ),
          ).toHaveText("This file type cannot be previewed.");
          await expect(
            page
              .getByRole("status")
              .filter({ hasText: "This file type cannot be previewed." }),
          ).toHaveCount(1);
          await expect.soft(inspectorPreview).toBeFocused();
        }
        await inspectorPreview.press("Escape");
        await expect(inspectorPreview).toHaveCount(0);
      }
      await page.setViewportSize({ width: 1440, height: 900 });
      await page.evaluate(() => {
        document.documentElement.style.zoom = "200%";
      });
      await expectEvidenceControlsPainted(page, availablePreview.record_id);
      await page.evaluate(() => {
        document.documentElement.style.zoom = "";
      });
      await page.setViewportSize({ width: 1440, height: 900 });

      const availableState = page.getByTestId(
        evidenceAccessStateTestId(availablePreview.record_id),
      );
      const attach = availableState.getByRole("button", {
        name: /Attach file/u,
      });
      await attach.focus();
      const chooserPromise = page.waitForEvent("filechooser");
      await attach.press("Enter");
      await (await chooserPromise).setFiles([]);

      // Hold a real issuance response while the user dismisses the pending panel.
      let release!: () => void;
      let finished!: () => void;
      const held = new Promise<void>((resolve) => {
        release = resolve;
      });
      const completed = new Promise<void>((resolve) => {
        finished = resolve;
      });
      const previewPath = `**/api/v1/evidence-records/${availablePreview.record_id}/preview-handle`;
      await page.route(previewPath, async (route) => {
        const response = await route.fetch();
        await held;
        await route.fulfill({ response });
        finished();
      });
      const preview = page.getByTestId(
        evidencePreviewButtonTestId(availablePreview.record_id),
      );
      await preview.focus();
      await preview.press("Enter");
      const pendingPanel = page.getByTestId(evidencePreviewPanelTestId());
      await expect(pendingPanel).toContainText("Opening preview…");
      await pendingPanel.getByRole("button", { name: "Close" }).click();
      release();
      await completed;
      await expect(pendingPanel).toHaveCount(0);
      await expect(preview).toBeFocused();
      await page.unroute(previewPath);

      const incidentResponse = await page.request.get(
        `${apiBase}/api/v1/incidents/${incidentId}`,
      );
      const incident = (await incidentResponse.json()).data;
      const closed = await page.request.post(
        `${apiBase}/api/v1/incidents/${incidentId}/close`,
        {
          headers: await csrfHeaders(page),
          data: {
            base_incident_version: incident.incident_version,
            client_txn_id: uniqueTxn("evidence-access-close"),
            reason: "Evidence read access verification",
          },
        },
      );
      expect(closed.ok()).toBeTruthy();
      await page.reload();
      await expectEvidenceControlsPainted(page, availablePreview.record_id);
      await expect(
        page.getByTestId(
          evidenceAttachFileInputTestId(availablePreview.record_id),
        ),
      ).toBeDisabled();
      await expect(
        page.getByTestId(
          evidencePreviewButtonTestId(availablePreview.record_id),
        ),
      ).toBeEnabled();
      await expect(
        page.getByTestId(
          evidenceDownloadButtonTestId(availablePreview.record_id),
        ),
      ).toBeEnabled();
      await page
        .getByTestId(evidencePreviewButtonTestId(availablePreview.record_id))
        .click();
      await expect(
        page.getByTestId(
          evidencePreviewFrameTestId(availablePreview.record_id),
        ),
      ).toBeVisible();
      await page
        .getByTestId(evidencePreviewPanelTestId())
        .getByRole("button", { name: "Close" })
        .click();
      const closedDownload = page.waitForEvent("download");
      await page
        .getByTestId(evidenceDownloadButtonTestId(availablePreview.record_id))
        .click();
      expect((await closedDownload).suggestedFilename()).toBe(
        "a11y.evidence-workflow-preview.txt",
      );
    } finally {
      await page.evaluate(() => {
        document.documentElement.style.zoom = "";
      });
      await setDensity(originalPreferences.density_mode);
    }

    await expectEvidenceControlsPainted(page, availablePreview.record_id);
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

test.describe("browser.collaboration accessibility readiness", () => {
  test(
    p7AccessibilityScenarioTitles[0],
    async ({ browser, page, sessionTracker }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YCOLLABORATION"),
        "a11y.collaboration conflict accessibility",
      );
      const remote = await createIncidentMemberUser(page, incidentId, {
        display_name: "Accessible Analyst",
        email: uniqueEmail("a11y.collaboration-remote"),
        initial_password: "A11yCollaborationRemotePass!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      });
      const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("a11y.collaboration-row"),
        "timeline.activity_synopsis_text": "a11y.collaboration conflict base",
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
            createdBy: "a11y.collaboration.row-01",
            email: remote.email,
            incidentId,
            password: remote.initial_password,
            purpose: "a11y.collaboration remote presence analyst",
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

        const localConflictValue = `a11y_collaboration_local_${"L".repeat(96)}`;
        const savedConflictValue = `a11y_collaboration_saved_${"S".repeat(96)}`;
        await driveRealTimelineSummaryConflict({
          baseRowVersion: 1,
          localValue: localConflictValue,
          page,
          patchController,
          recordId,
          remotePatchPage: remotePage,
          remoteValue: savedConflictValue,
          txnPrefix: "fea11yp7-conflict",
        });
        const resolver = page.getByTestId(workbookConflictResolverTestId());
        await expect(resolver).toHaveAttribute(
          "aria-label",
          "Workbook conflict recovery",
        );
        const summary = page.getByTestId(workbookConflictSummaryTestId());
        await expect(summary).toBeFocused();
        await expect(resolver).toHaveAttribute(
          "data-conflict-field-key",
          "timeline.activity_synopsis_text",
        );
        await expect(
          page.getByTestId(workbookConflictSavedValueTestId()),
        ).toHaveValue(savedConflictValue);
        await expect(
          page.getByTestId(workbookConflictLocalValueTestId()),
        ).toHaveValue(localConflictValue);
        await expect(page.getByRole("button", { name: "Close" })).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Discard local draft" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use my unsaved value" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use merged value" }),
        ).toBeVisible();

        for (const viewport of [
          { width: 1440, height: 900 },
          { width: 1024, height: 720 },
          { width: 768, height: 640 },
        ]) {
          await page.setViewportSize(viewport);
          await expectResolverSurfaceGeometry(page);
        }
        await page.setViewportSize({ width: 1440, height: 900 });
        await page.evaluate(() => {
          document.documentElement.style.zoom = "200%";
        });
        await expectResolverSurfaceGeometry(page, {
          enforceDocumentBlockExtent: false,
        });
        await page.evaluate(() => {
          document.documentElement.style.zoom = "100%";
        });
        await page.setViewportSize({ width: 768, height: 640 });
        await page.evaluate(() => {
          const style = document.createElement("style");
          style.id = "workbook-resolver-text-spacing";
          style.textContent = `
            #root * {
              letter-spacing: 0.12em !important;
              line-height: 1.5 !important;
              word-spacing: 0.16em !important;
            }
            #root p {
              margin-bottom: 2em !important;
            }
          `;
          document.head.append(style);
        });
        await expectResolverSurfaceGeometry(page);
        await page.evaluate(() => {
          document.getElementById("workbook-resolver-text-spacing")?.remove();
        });
        await page.setViewportSize({ width: 1440, height: 900 });
        await resolver.evaluate((element) => {
          element.scrollTop = 0;
        });

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
        const retainedConflictMarker = semanticGridCell(
          retainedConflictEditor,
        ).locator('[data-grid-state-marker="conflicted"]');
        await expect(retainedConflictMarker).toBeVisible();
        await expect(retainedConflictMarker).toHaveAttribute(
          "aria-label",
          "Conflict on Activity Synopsis",
        );
        await expectAllInteractiveControlsNamed(page);
        await expectNoFocusTrap(page);
        await expectAndRecordContrast(page, [
          saveStateTestId(),
          rowPresenceMarkerTestId(recordId),
          cellPresenceMarkerTestId(recordId, "timeline.activity_synopsis_text"),
          workbookConflictControlTestId("close"),
          workbookConflictControlTestId("keep-saved"),
          workbookConflictControlTestId("use-unsaved"),
          workbookConflictControlTestId("use-merged"),
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

test.describe("browser.saved-view-query accessibility readiness", () => {
  test(p8AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const longSavedViewName =
      "a11y.saved-view-query workbook layout resilience with a deliberately long selected view name";
    const longTagToken =
      "unbroken-tag-token-0123456789-abcdefghijklmnopqrstuvwxyz";
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YSAVEDVIEWQUERY"),
      "a11y.saved-view-query query controls",
    );
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.saved-view-query-reviewed"),
        "timeline.activity_synopsis_text": "a11y.saved-view-query reviewed row",
        "timeline.tags": {
          actions: [{ op: "add_tag", tag_name: longTagToken }],
          kind: "collection_actions_v1",
        },
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("a11y.saved-view-query-rough"),
      "timeline.activity_synopsis_text": "a11y.saved-view-query rough row",
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
      workbookQueryEntryTestId(
        timelineViewSchemaId,
        "filter",
        "timeline.capture_state",
      ),
    );
    await expectVisibleFocus(filterChip);

    const sortMenuTrigger = page.getByTestId(
      workbookSortMenuTriggerTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(sortMenuTrigger);
    await sortMenuTrigger.press("Enter");
    const sortMenu = page.getByTestId(
      workbookSortMenuTestId(timelineViewSchemaId),
    );
    await expect(sortMenu).toBeVisible();
    for (const fieldKey of [
      "timeline.activity_sort_ts",
      "timeline.date_entered_sort_day",
      "timeline.analyst_text",
      "timeline.mitre_stage_text",
      "timeline.device_object_text",
      "timeline.has_evidence",
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
    await expect(sortMenuTrigger).toBeFocused();
    await expect(sortMenuTrigger).toHaveAttribute(
      "aria-label",
      "Sort, 8 applied",
    );
    await expect(sortMenuTrigger).not.toHaveAttribute("title");

    await openFilterPopover(page, timelineViewSchemaId);
    await filterField.selectOption("timeline.tags");
    await filterValue.fill(longTagToken);
    await filterApply.press("Enter");

    const groupingSelect = page.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(groupingSelect);
    await groupingSelect.selectOption("timeline.capture_state");
    await expect(groupingSelect).toHaveAttribute("title", "Capture State");
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
    await savedViewNameInput.fill(longSavedViewName);
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
    await expect(savedViewSelector).toHaveAttribute("title", longSavedViewName);

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
    await expect(savedViewStatus).toHaveAttribute(
      "title",
      "Default view updated.",
    );

    await groupingSelect.selectOption("timeline.has_evidence");
    await expect(
      page.getByTestId(savedViewModifiedTestId(timelineViewSchemaId)),
    ).toHaveText("Modified");
    await expect(
      page.locator('[data-grid-data-state="refreshing"]'),
    ).toHaveCount(0);
    await expect(
      page.locator(
        '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
      ),
    ).toHaveCount(0);

    const queryControls = page.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    await expect(queryControls).toHaveAttribute(
      "data-hidden-query-chip-count",
      "8",
    );
    const filterTrigger = page.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
    );
    await expect(filterTrigger).toHaveAttribute(
      "aria-label",
      "Filters, 2 active filters, 8 hidden query entries",
    );
    await expectVisibleFocus(filterTrigger);
    await filterTrigger.press("Enter");
    const filterPopover = page.getByTestId(
      workbookFilterPopoverTestId(timelineViewSchemaId),
    );
    await expect(filterPopover).toBeVisible();
    await expect(filterTrigger).toHaveAttribute(
      "aria-controls",
      workbookFilterPopoverTestId(timelineViewSchemaId),
    );
    await expect(
      filterPopover.getByTestId(
        workbookQueryOverflowEntryTestId(
          timelineViewSchemaId,
          "filter",
          "timeline.tags",
        ),
      ),
    ).toContainText(longTagToken);
    await filterPopover.press("Escape");
    await expect(filterPopover).toHaveCount(0);
    await expect(filterTrigger).toBeFocused();

    const textSpacingStyle = await page.addStyleTag({
      content: `
        * {
          line-height: 1.5 !important;
          letter-spacing: 0.12em !important;
          word-spacing: 0.16em !important;
        }
        p {
          margin-block-end: 2em !important;
        }
      `,
    });
    await expectTextSpacingViewBarResilience(page, {
      capacity: 3,
      height: 900,
      hiddenCount: 8,
      width: 1440,
    });
    await expectTextSpacingViewBarResilience(page, {
      capacity: 2,
      height: 720,
      hiddenCount: 9,
      width: 1024,
    });
    await expectTextSpacingViewBarResilience(page, {
      capacity: 0,
      height: 640,
      hiddenCount: 11,
      width: 768,
    });
    await expectVisibleFocus(
      page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    );
    await expectVisibleFocus(
      page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
    );
    await textSpacingStyle.evaluate((element) => {
      element.parentNode?.removeChild(element);
    });
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
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
    const zoomedDocumentGeometry = await page.evaluate(() => ({
      clientWidth: document.documentElement.clientWidth,
      scrollWidth: document.documentElement.scrollWidth,
    }));
    expect(zoomedDocumentGeometry.scrollWidth).toBeLessThanOrEqual(
      zoomedDocumentGeometry.clientWidth + 1,
    );
    await page.evaluate(() => {
      document.documentElement.style.zoom = "";
    });
    await groupingSelect.selectOption("timeline.capture_state");
    await expect(reviewedGroup).toBeVisible();

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
      workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
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

test.describe("browser.inspector-history accessibility readiness", () => {
  test(
    p9ConfigAccessibilityScenarioTitles[0],
    async ({ browser, page, sessionTracker }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YINSPECTORCONFIG"),
        "a11y.inspector-history.row-02 config-driven inspector",
      );
      const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("a11y.inspector-history-02-row"),
        "timeline.raw_activity_text":
          "a11y.inspector-history.row-02 inspector details",
        "timeline.activity_synopsis_text":
          "a11y.inspector-history.row-02 selected row",
      });
      const viewerPassword = "A11yInspectorViewer1!";
      const viewer = await createIncidentMemberUser(page, incidentId, {
        display_name: "A11y Inspector Viewer",
        email: uniqueEmail("a11y-inspector-viewer"),
        initial_password: viewerPassword,
        role: "viewer",
        is_deployment_admin: false,
        mfa_required: false,
      });

      await page.goto(`/?incident_id=${incidentId}`);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

      const toggle = page.getByTestId(
        workbookInspectorToggleTestId(timelineViewSchemaId),
      );
      await toggle.focus();
      await expectVisibleFocus(toggle);
      await toggle.press("Enter");
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveAttribute(
        "data-inspector-state",
        "no_row_selected",
      );
      await expect(
        page.getByText("Select a saved row to inspect its details."),
      ).toBeVisible();
      await expect(
        page.getByTestId(timelineInspectorTestId()),
      ).not.toContainText("no_row_selected");
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
      await expect(detailsEditor).toHaveValue(
        "a11y.inspector-history.row-02 inspector details",
      );

      await detailsEditor.press("Escape");
      const semanticSummaryCell = semanticGridCell(summaryCell);
      await expect(semanticSummaryCell).toBeFocused();
      await semanticSummaryCell.press("Escape");
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
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
      await expect(deletePanel).not.toHaveAttribute("aria-modal");
      await expect(deletePanel).toContainText(row.record_id);
      const deleteConfirm = page.getByTestId(
        rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      );
      const deleteCancel = page.getByTestId(
        rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
      );
      await expect(deleteCancel).toBeFocused();
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
      await deleteCancel.press("Escape");
      await expect(deletePanel).toHaveCount(0);
      await expect(deleteButton).toBeFocused();

      const viewerSession = await openIncidentAsTrackedUserReady(
        browser,
        sessionTracker,
        {
          createdBy: "a11y.inspector-history.row-02",
          email: viewer.email,
          incidentId,
          password: viewerPassword,
          purpose: "a11y inspector disabled reason viewer",
          readyRecordId: row.record_id,
          userId: viewer.user_id,
        },
      );
      try {
        const viewerPage = viewerSession.page;
        await viewerPage.setViewportSize({ width: 1024, height: 720 });
        await openTimelineInspector(viewerPage, row.record_id);
        const disabledAction = viewerPage.getByTestId(
          workbookInspectorFeatureActionTestId(
            timelineViewSchemaId,
            "indicator.observations.manage",
          ),
        );
        await expect(disabledAction).toBeDisabled();
        const descriptionId =
          await disabledAction.getAttribute("aria-describedby");
        expect(descriptionId).not.toBeNull();
        await expect(viewerPage.locator(`[id="${descriptionId}"]`)).toHaveText(
          "Requires the editor incident role.",
        );

        await viewerPage.evaluate((inspectorSelector) => {
          const style = document.createElement("style");
          style.id = "workbook-inspector-text-spacing";
          style.textContent = `
              ${inspectorSelector} * {
                letter-spacing: 0.12em !important;
                line-height: 1.5 !important;
                word-spacing: 0.16em !important;
              }
              ${inspectorSelector} p {
                margin-block-end: 2em !important;
              }
            `;
          document.head.append(style);
        }, dataTestIdSelector(timelineInspectorTestId()));
        const textSpacingGeometry = await viewerPage
          .getByTestId(timelineInspectorTestId())
          .evaluate((element) => ({
            clientWidth: element.clientWidth,
            scrollWidth: element.scrollWidth,
          }));
        expect(textSpacingGeometry.scrollWidth).toBeLessThanOrEqual(
          textSpacingGeometry.clientWidth + 1,
        );
        await viewerPage.evaluate(() => {
          document.getElementById("workbook-inspector-text-spacing")?.remove();
          document.documentElement.style.zoom = "200%";
        });
        await expect(
          viewerPage.getByTestId(
            workbookInspectorCloseButtonTestId(timelineViewSchemaId),
          ),
        ).toBeVisible();
        const zoomGeometry = await viewerPage
          .getByTestId(timelineInspectorTestId())
          .evaluate((element) => ({
            clientWidth: element.clientWidth,
            scrollWidth: element.scrollWidth,
          }));
        expect(zoomGeometry.scrollWidth).toBeLessThanOrEqual(
          zoomGeometry.clientWidth + 1,
        );
        await viewerPage.evaluate(() => {
          document.documentElement.style.zoom = "100%";
        });
      } finally {
        await viewerSession.page.context().close();
      }
    },
  );

  test(p9AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YINSPECTORHISTORY"),
      "a11y.inspector-history inspector actions",
    );
    const evidence = await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.inspector-history-evidence"),
        "evidence.collector_party_text": "a11y.inspector-history collector",
        "evidence.title": "a11y.inspector-history evidence",
      },
    );
    const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
      [hostRefsFieldKey]: collectionActionsPayload([
        "a11y.inspector-history host",
      ]),
      client_txn_id: uniqueTxn("a11y.inspector-history-row"),
      "timeline.raw_activity_text": "a11y.inspector-history inspector details",
      "timeline.activity_synopsis_text": "a11y.inspector-history selected row",
    });
    const linkedRow = await patchRecord(page, row.record_id, {
      base_row_version: row.row_version,
      changes: [
        {
          action_payload: A11yAttachedEvidencePayload(evidence.record_id),
          field_key: "timeline.attached_evidence_ids",
        },
      ],
      client_txn_id: uniqueTxn("a11y.inspector-history-link"),
      view_schema_id: timelineViewSchemaId,
    });
    const hostItem = requireItemByRawText(
      collectionItems(linkedRow, hostRefsFieldKey),
      "a11y.inspector-history host",
    );
    const history = await fetchRecordHistory(page, row.record_id);
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
    await expect(detailsEditor).toHaveValue(
      "a11y.inspector-history inspector details",
    );

    const relationshipChip = page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(hostItem.item_ref)));
    await expect(relationshipChip).toContainText("Unresolved");
    await expectVisibleFocus(relationshipChip);

    await expectVisibleFocus(detailsEditor);
    await detailsEditor.press("Escape");
    const semanticSummaryCell = semanticGridCell(summaryCell);
    await expect(semanticSummaryCell).toBeFocused();
    await semanticSummaryCell.press("Escape");
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
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
    await expect(rollbackPreview).toHaveAttribute("role", "alertdialog");
    await expect(rollbackPreview).not.toHaveAttribute("aria-modal");
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
            status: 409,
            code: "row_version_conflict",
            message: "row_version_conflict",
            request_id: "a11y-inspector-history-conflict",
            retryable: false,
            details: {
              reason_code: "stale_row_version",
            },
          },
        }),
      }),
    );
    await rollbackConfirm.press("Enter");
    const historyMessage = page.getByTestId(rowHistoryMessageTestId());
    await expectAlertRole(historyMessage);
    await expect(historyMessage).toHaveAttribute("aria-live", "assertive");
    await expect(historyMessage).toContainText(
      "This row changed; refresh it before retrying.",
    );
    await expect(historyMessage).toContainText("row_version_conflict");
    await expectNoPrivateDiagnostics(historyMessage);

    const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
    await expectVisibleFocus(deleteButton);
    await deleteButton.press("Enter");
    const deletePanel = page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    );
    await expect(deletePanel).toHaveAttribute("role", "alertdialog");
    await expect(deletePanel).not.toHaveAttribute("aria-modal");
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

test.describe("browser.coordination-review accessibility readiness", () => {
  test(p10AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YCOORDINATIONREVIEW"),
      "a11y.coordination-review coordination accessibility",
    );
    const owner = await createIncidentMemberUser(page, incidentId, {
      display_name: "browser.coordination-review accessibility owner",
      email: uniqueEmail("coordination-review-a11y-owner"),
      initial_password: "BackupRestoreA11y1!",
      role: "editor",
      is_deployment_admin: false,
      mfa_required: false,
    });
    const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("a11y.coordination-review-party"),
      "party.display_name": "a11y.coordination-review response party",
      "party.party_kind": "team",
    });
    const task = await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.coordination-review-task"),
        "task.priority": "normal",
        "task.requester_party_id": party.record_id,
        "task.task_kind": "collection",
        "task.title": "a11y.coordination-review task alpha",
      },
    );
    const urgentTask = await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.coordination-review-urgent-task"),
        "task.priority": "urgent",
        "task.task_kind": "follow_up",
        "task.title": "a11y.coordination-review task urgent",
      },
    );
    const clipboardRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.coordination-review-clipboard"),
        "timeline.activity_utc_text": "2026-06-12T10:00:00Z",
        "timeline.activity_synopsis_text":
          "a11y.coordination-review clipboard row",
      },
    );
    const decision = await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.coordination-review-decision"),
        "decision.decision_type": "containment",
        "decision.rationale": "a11y.coordination-review coordination rationale",
        "decision.summary": "a11y.coordination-review decision summary",
      },
    );
    const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("a11y.coordination-review-comm"),
      "comm_log.audience": "a11y.coordination-review responders",
      "comm_log.channel_or_meeting": "a11y.coordination-review bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.decision_ids": {
        actions: [
          { linked_record_id: decision.record_id, op: "add_record_ref" },
        ],
        kind: "collection_actions_v1",
      },
      "comm_log.summary": "a11y.coordination-review communications log",
    });
    const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
      client_txn_id: uniqueTxn("a11y.coordination-review-handoff"),
      "handoff.current_state_summary": "a11y.coordination-review handoff state",
      "handoff.incoming_owner_user_id": owner.user_id,
    });
    const status = await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("a11y.coordination-review-status"),
        "status_review.current_state_summary":
          "a11y.coordination-review status review state",
      },
    );
    const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("a11y.coordination-review-lesson"),
      "lesson.summary": "a11y.coordination-review lesson summary",
    });

    const surfaces = [
      {
        expected: "a11y.coordination-review task alpha",
        fieldKey: "task.title",
        groupToken: "coordination",
        label: "Task Requests",
        row: task,
        viewSchemaId: taskRequestsViewSchemaId,
      },
      {
        expected: "a11y.coordination-review decision summary",
        fieldKey: "decision.summary",
        groupToken: "coordination",
        label: "Decisions",
        row: decision,
        viewSchemaId: decisionsViewSchemaId,
      },
      {
        expected: "a11y.coordination-review response party",
        fieldKey: "party.display_name",
        groupToken: "coordination",
        label: "Parties",
        row: party,
        viewSchemaId: partiesViewSchemaId,
      },
      {
        expected: "a11y.coordination-review communications log",
        fieldKey: "comm_log.summary",
        groupToken: "coordination",
        label: "Communications Log",
        row: comm,
        viewSchemaId: commLogViewSchemaId,
      },
      {
        expected: "a11y.coordination-review handoff state",
        fieldKey: "handoff.current_state_summary",
        groupToken: "coordination",
        label: "Handoff",
        row: handoff,
        viewSchemaId: handoffViewSchemaId,
      },
      {
        expected: "a11y.coordination-review status review state",
        fieldKey: "status_review.current_state_summary",
        groupToken: "review-learning",
        label: "Status Review",
        row: status,
        viewSchemaId: statusReviewViewSchemaId,
      },
      {
        expected: "a11y.coordination-review lesson summary",
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
        await page.keyboard.press("Escape");
        await expect(
          page.getByTestId(workbookFilterPopoverTestId(surface.viewSchemaId)),
        ).toHaveCount(0);

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
        await expect(
          page.getByTestId(workbookFocusAnchorTestId()),
        ).toContainText(surface.viewSchemaId);
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
    expect(copiedTaskTitle).toBe("a11y.coordination-review task alpha");

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
      matrix: [
        [
          "a11y.coordination-review pasted timeline",
          "a11y.coordination-review-host",
        ],
      ],
      page,
      recordId: clipboardRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(clipboardSummary).toHaveText(
      "a11y.coordination-review pasted timeline",
    );

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
      workbookQueryEntryTestId(
        taskRequestsViewSchemaId,
        "filter",
        "task.priority",
      ),
    );
    await expect(priorityChip).toContainText("urgent");
    await expectVisibleFocus(priorityChip);

    const urgentTitle = await mountedGridCell(
      page,
      taskRequestsViewSchemaId,
      urgentTask.record_id,
      "task.title",
    );
    await expectCellTextOrValue(
      urgentTitle,
      "a11y.coordination-review task urgent",
    );

    await openSavedViewActionMenu(page, taskRequestsViewSchemaId);
    const savedViewName = page.getByTestId(
      savedViewNameInputTestId(taskRequestsViewSchemaId),
    );
    await expectVisibleFocus(savedViewName);
    await savedViewName.fill("a11y.coordination-review keyboard saved view");
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

    await test
      .info()
      .attach("a11y.coordination-review-01-readiness-matrix.json", {
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
      workbookQueryEntryTestId(
        taskRequestsViewSchemaId,
        "filter",
        "task.priority",
      ),
      gridGroupingSelectTestId(taskRequestsViewSchemaId),
      gridSortHeaderTestId(taskRequestsViewSchemaId, "task.title"),
      rowCellTestId(urgentTask.record_id, "task.title"),
      saveStateTestId(),
    ]);
  });
});

test.describe("browser.design-readiness accessibility readiness", () => {
  test(p11AccessibilityScenarioTitles[0], async ({ page }) => {
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
    await semanticGridCell(inspectorSummaryCell).focus();
    await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
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
    await expectCellTextOrValue(taskTitle, "a11y.design-readiness task row");
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

test.describe("browser.incident-selection accessibility readiness", () => {
  test(p1AccessibilityScenarioTitles[0], async ({ page }) => {
    await clearBrowserSession(page);
    const heldSession = await holdSinglePublicAPIResponse(page, {
      method: "GET",
      path: "/api/v1/auth/session",
    });
    await new AuthGateway(page).goto();
    await heldSession.waitForHit;

    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "loading",
    );
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await expectStatusRole(page.getByTestId(authTestId("status")));
    await expect(page.getByTestId(authTestId("status"))).toContainText(
      "Checking current session",
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: authTestId("login-submit"),
      tabStops: [
        authTestId("login-username"),
        authTestId("login-password"),
        authTestId("login-submit"),
      ],
    });

    try {
      heldSession.release();
      await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
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
      const user = await createDeploymentUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Login",
        initial_password: password,
        mfa_required: false,
        is_deployment_admin: false,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "anonymous",
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: authTestId("login-username"),
        tabStops: [
          authTestId("login-username"),
          authTestId("login-password"),
          authTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(email, password);
      await expect(
        page.getByTestId(incidentLandingTestId("shell")),
      ).toBeVisible();
      await expectVisibleFocus(
        page.getByTestId(incidentLandingTestId("refresh")),
      );
      await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
      await expectP1SurfaceA11y(page, {
        focusTestId: incidentLandingTestId("refresh"),
        tabStops: [
          incidentLandingTestId("search"),
          incidentLandingTestId("status-filter"),
          incidentLandingTestId("refresh"),
          incidentLandingTestId("create-open-button"),
        ],
      });
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "authentication accessibility",
        email,
        purpose: "a11y.incident-selection.row-01 anonymous login",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[2],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-mfa");
      const password = "A11yP1MfaPass!";
      const user = await createDeploymentUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA",
        initial_password: password,
        mfa_required: true,
        is_deployment_admin: false,
      });
      const secretBase32 = await enrollTotpViaBootstrap(email, password);

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_required",
      );
      await expectStatusRole(page.getByTestId(authTestId("status")));
      await expect(page.getByTestId(authTestId("status"))).toContainText(
        "Authenticator code",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId(publicErrorSummaryTestIds("auth").container),
      );
      expect(await hasSessionCookie(page)).toBeFalsy();
      await expectP1SurfaceA11y(page, {
        focusTestId: authTestId("login-totp-code"),
        tabStops: [
          authTestId("login-username"),
          authTestId("login-password"),
          authTestId("login-totp-code"),
          authTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(
        email,
        password,
        generateTotpCode(secretBase32),
      );
      await expect(
        page.getByTestId(incidentLandingTestId("current-user")),
      ).toContainText("A11Y P1 MFA");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "authentication accessibility",
        email,
        purpose: "a11y.incident-selection.row-01 mfa_required retry",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[3],
    async ({ page, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-mfa-setup");
      const password = "A11yP1SetupPass!";
      await createDeploymentUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA Setup",
        initial_password: password,
        mfa_required: true,
        is_deployment_admin: false,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_setup_required",
      );
      await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
        "Authenticator setup is required before sign-in.",
      );
      await expect(page.getByTestId(authTestId("bootstrap-token"))).toHaveText(
        "Stored for TOTP setup requests.",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId(publicErrorSummaryTestIds("auth").container),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: authTestId("bootstrap-begin"),
        tabStops: [
          authTestId("bootstrap-begin"),
          authTestId("bootstrap-complete-code"),
        ],
      });

      await new AuthGateway(page).beginBootstrapEnrollment();
      await expectStatusRole(page.getByTestId(authTestId("status")));
      const secretBase32 = await new AuthGateway(page).requireText(
        authTestId("bootstrap-secret-base32"),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: authTestId("bootstrap-complete-code"),
        tabStops: [
          authTestId("bootstrap-complete-code"),
          authTestId("bootstrap-complete"),
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
    await expect(
      page.getByTestId(incidentLandingTestId("shell")),
    ).toBeVisible();
    await expect(
      page.getByTestId(landingIncidentCardTestId(incidentId)),
    ).toBeVisible();
    await expectVisibleFocus(
      page.getByTestId(incidentLandingTestId("create-open-button")),
    );
    await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
    await expectP1SurfaceA11y(page, {
      focusTestId: incidentLandingTestId("create-open-button"),
      tabStops: [
        incidentLandingTestId("search"),
        incidentLandingTestId("status-filter"),
        incidentLandingTestId("refresh"),
        incidentLandingTestId("create-open-button"),
      ],
    });
  });

  test(
    p1AccessibilityScenarioTitles[5],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const email = uniqueEmail("a11y-p1-incident");
      const password = "A11yP1IncidentPass!";
      const user = await createDeploymentUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Incident",
        initial_password: password,
        mfa_required: false,
        is_deployment_admin: false,
      });

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(
        page.getByTestId(incidentLandingTestId("empty-state")),
      ).toContainText("No incidents are visible");
      await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));

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
        page.getByTestId(incidentAdministrationTestId("patch-button")),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: appRouteTestId("workbook-current-user"),
        tabStops: [appRouteTestId("workbook-current-user")],
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
      await expect(
        page.getByTestId(incidentAdministrationTestId("patch-tlp")),
      ).toBeVisible();
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
      await expectAlertRole(
        page.getByTestId(incidentAdministrationTestId("admin-error-code")),
      );
      await expect(
        page.getByTestId(incidentAdministrationTestId("admin-error-code")),
      ).toHaveText("authorization_denied");
      await expectNoPrivateDiagnostics(
        page.getByTestId(incidentAdministrationTestId("admin-error-code")),
      );
      await expectVisibleFocus(
        page.getByTestId(appRouteTestId("workbook-current-user")),
      );
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "authentication accessibility",
        email,
        purpose: "a11y.incident-selection.row-01 incident states",
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
    await expect(
      page.getByTestId(incidentLandingTestId("shell")),
    ).toBeVisible();
    await expect(
      page.getByTestId(incidentLandingTestId("shell")),
    ).toHaveAttribute("data-bootstrap-state", "forbidden");
    await expectStatusRole(page.getByTestId(incidentLandingTestId("status")));
    await expectAlertRole(page.getByTestId(publicErrorCodeTestId("landing")));
    await expect(page.getByTestId(publicErrorCodeTestId("landing"))).toHaveText(
      "authorization_denied",
    );
    await expectNoPrivateDiagnostics(
      page.getByTestId(publicErrorSummaryTestIds("landing").container),
    );
    await expectVisibleFocus(
      page.getByTestId(incidentLandingTestId("refresh")),
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: incidentLandingTestId("refresh"),
      tabStops: [incidentLandingTestId("refresh")],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(incidentLandingTestId("refresh")).click();
    await expect(page.getByTestId(publicErrorCodeTestId("landing"))).toHaveText(
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
      const user = await createDeploymentUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Revoked",
        initial_password: password,
        mfa_required: false,
        is_deployment_admin: false,
      });
      await createIncidentMembership(page, incidentId, email, "viewer");

      await clearBrowserSession(page);
      await new AuthGateway(page).goto();
      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(
        page.getByTestId(appRouteTestId("workbook-current-user")),
      ).toContainText("A11Y P1 Revoked");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "authentication accessibility",
        email,
        purpose:
          "a11y.incident-selection.row-01 revoked session before revoke-all",
        userId: user.user_id,
      });

      await revokeAllSessions(
        workerAdminRequest,
        user.user_id,
        "a11y.incident-selection.row-01 revoked-session",
      );
      await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "revoked",
      );
      await expect(page.getByTestId(authTestId("shell-message"))).toContainText(
        "Sign in again",
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: authTestId("login-submit"),
        tabStops: [
          authTestId("login-username"),
          authTestId("login-password"),
          authTestId("login-submit"),
        ],
      });

      await new AuthGateway(page).login(email, password);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(
        page.getByTestId(appRouteTestId("workbook-current-user")),
      ).toContainText("A11Y P1 Revoked");
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "authentication accessibility",
        email,
        purpose: "a11y.incident-selection.row-01 revoked session re-auth",
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
    await new AccountSettings(page).openSecurity();
    await page.getByTestId(accountTestId("refresh-state")).focus();
    await expectVisibleFocus(page.getByTestId(accountTestId("refresh-state")));
    await expect(page.getByTestId(publicErrorCodeTestId("account"))).toHaveText(
      "credential_state_unavailable",
    );
    await expectAlertRole(page.getByTestId(publicErrorCodeTestId("account")));
    await expectAlertRole(
      page.getByTestId(publicErrorSummaryTestIds("account").container),
    );
    await expect(
      page.getByTestId(publicErrorSummaryTestIds("account").message),
    ).toHaveText("Request failed.");
    await expect(
      page.getByTestId(publicErrorSummaryTestIds("account").details),
    ).toContainText("Reason: temporary_failure");
    await expect(
      page.getByTestId(publicErrorSummaryTestIds("account").details),
    ).toContainText("Field: credential_state");
    await expectNoPrivateDiagnostics(
      page.getByTestId(publicErrorSummaryTestIds("account").container),
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: accountTestId("refresh-state"),
      tabStops: [accountTestId("refresh-state"), accountTestId("logout")],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(accountTestId("refresh-state")).click();
    await expect(page.getByTestId(accountTestId("status"))).toHaveText(
      "Refreshed account security.",
    );
  });
});
